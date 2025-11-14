# Dogelytics

A lightweight REST API service that provides Dogecoin wallet balance information by querying the indexer's PostgreSQL database.

## Prerequisites

- Go 1.21 or later
- **Two PostgreSQL databases**:
  - **Indexer Database**: Shared with [Dogecoin Indexer](https://github.com/dogeorg/indexer) for blockchain data (read-only)
  - **Dogelytics Database**: Separate database for users, API keys, and sessions
  - **PostgreSQL is required** - SQLite is not supported due to concurrent access issues
  - indexer must be configured to use PostgreSQL (see [docker/indexer](../docker/indexer/))

## Features

- **Balance Queries**: Get incoming, available, outgoing, and current balance for any Dogecoin address
- **Health Checks**: Monitor indexer's status and current block height
- **Rate Limiting**: Per-IP and per-API key rate limiting to prevent abuse
- **API Key Management**: Secure user accounts with API key generation and management
- **Usage Statistics**: Real-time usage tracking with interactive charts and filtering
- **CORS Support**: Configurable CORS headers for browser-based applications

## Installation

```bash
cd dogelytics
go mod download
```

## Database Setup

Dogelytics requires a separate dogelytics database for users, API keys, and sessions. 

### Automated Setup (Recommended)

Use the provided setup script for easy and secure database initialization:

```bash
./scripts/setup-auth-db.sh
```

This interactive script will:
- Connect to your PostgreSQL instance
- **Auto-generate a secure random password** (or let you provide your own)
- Create the `dogelytics` database
- Create the `dogelytics` user with proper permissions
- Output the `DOGELYTICS_DBURL` connection string for your `.env` file

### Manual Setup

If you prefer manual setup, **use a strong password**:

```sql
-- Connect to PostgreSQL as admin
psql -U postgres

-- Create database and user (replace YOUR_SECURE_PASSWORD with a strong password!)
CREATE DATABASE dogelytics;
CREATE USER dogelytics WITH PASSWORD 'YOUR_SECURE_PASSWORD';
GRANT ALL PRIVILEGES ON DATABASE dogelytics TO dogelytics;

-- Connect to dogelytics database
\c dogelytics

-- Grant schema permissions (PostgreSQL 15+)
GRANT ALL ON SCHEMA public TO dogelytics;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO dogelytics;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO dogelytics;
```

Then add to your `.env` (use the same password you set above):
```bash
DOGELYTICS_DBURL=postgres://dogelytics:YOUR_SECURE_PASSWORD@localhost:5432/dogelytics?sslmode=disable
```

## Quick Start

**Important**: 
1. Ensure indexer is running with PostgreSQL before starting dogelytics
2. Create the dogelytics database using the setup script or manual steps above
3. Configure your `.env` file with the database URLs from setup

### Run with `.env` configuration (recommended)

First, ensure your `.env` file is configured:

```bash
# .env file
INDEXER_DBURL=postgres://indexer:yourpassword@localhost:5432/indexer?sslmode=disable
DOGELYTICS_DBURL=postgres://dogelytics:yourpassword@localhost:5432/dogelytics?sslmode=disable
BIND=localhost:4420
CONFIRMATIONS=6
RATELIMIT=10
API_KEY_RATELIMIT=120
SESSION_SECRET=your-secret-key-here
ENABLE_UI=true
ENABLE_SIGNUPS=false
```

Then run:

```bash
go run ./cmd/dogelytics
```

### Build and run

```bash
# Compile
go build -o dogelytics ./cmd/dogelytics

# Run (reads from .env automatically)
./dogelytics
```

### Override with command-line flags (optional)

You can override `.env` settings with flags if needed:

```bash
./dogelytics \
  -indexer-dburl="postgres://indexer:mypassword@localhost:5432/indexer?sslmode=disable" \
  -dogelytics-dburl="postgres://dogelytics:mypassword@localhost:5432/dogelytics?sslmode=disable" \
  -bind="0.0.0.0:4420" \
  -confirmations=6 \
  -ratelimit=60
```

### Configuration Options

| Environment Variable | Flag | Description | Default |
|---------------------|------|-------------|---------|
| `INDEXER_DBURL` | `-indexer-dburl` | PostgreSQL database URL for indexer (blockchain data) | `postgres://indexer:changeme@localhost:5432/indexer?sslmode=disable` |
| `DOGELYTICS_DBURL` | `-dogelytics-dburl` | PostgreSQL database URL for dogelytics (users, API keys, sessions, logs) | `postgres://dogelytics:changeme@localhost:5432/dogelytics?sslmode=disable` |
| `BIND` | `-bind` | HTTP server bind address | `localhost:4420` |
| `CORS` | `-cors` | CORS allowed origin (`*` for all) | `*` |
| `CONFIRMATIONS` | `-confirmations` | Number of confirmations for available balance | `6` |
| `RATELIMIT` | `-ratelimit` | Maximum requests per IP per minute (0 = disabled) | `10` |
| `API_KEY_RATELIMIT` | `-apikey-ratelimit` | Maximum requests per API key per minute (0 = disabled) | `120` |
| `SESSION_SECRET` | `-session-secret` | Session HMAC secret (required for UI/auth) | `""` (empty) |
| `MAX_KEYS_PER_USER` | `-max-keys-per-user` | Maximum API keys per user | `1` |
| `ENABLE_UI` | `-enable-ui` | Enable web UI endpoints | `true` |
| `ENABLE_SIGNUPS` | `-enable-signups` | Enable user registration through UI | `true` |

Environment variables take precedence over defaults, and command-line flags take precedence over environment variables.

## API Endpoints

### GET /balance

Get the balance for a Dogecoin address.

**Parameters:**
- `address` (required): Dogecoin address (e.g., `D7Y55r1psU6xRfUyXRr59kAjdikKb8njf3`)

**Example Request:**
```bash
curl "http://localhost:4420/balance?address=DLAznsPDLDRgsVcTFWRMYMG5uH6GddDtv8"
```

**Example Response:**
```json
{
  "incoming": "100000000.00000000",
  "available": "500000000.00000000",
  "outgoing": "50000000.00000000",
  "current": "600000000.00000000"
}
```

All values are in Koinu (1 DOGE = 100,000,000 Koinu) formatted as decimal strings.

**Error Response:**
```json
{
  "error": "invalid-address",
  "message": "Invalid Dogecoin address format"
}
```

**Rate Limit Error (HTTP 429):**
```json
{
  "error": "rate-limit-exceeded",
  "message": "Rate limit exceeded. Maximum 60 requests per minute."
}
```

### GET /health

Health check endpoint to verify the service is running and can connect to the database.

**Example Request:**
```bash
curl "http://localhost:4420/health"
```

**Example Response:**
```json
{
  "ok": true,
  "height": 5900000
}
```

## Rate Limiting

Dogelytics includes built-in rate limiting.

**Default Rate Limits:**
- **Without API Key**: 10 requests per IP per minute
- **With API Key**: 120 requests per API key per minute

When a client exceeds the rate limit, they will receive an HTTP 429 (Too Many Requests) response with details about the limit.

## Web UI

Dogelytics includes an optional web interface accessible at `http://localhost:4420` with a classic Windows 95 theme.

**Note:** The web UI can be completely disabled by setting `ENABLE_UI=false`, which is useful for API-only deployments.

### Controlling Access

- **`ENABLE_UI`**: Set to `false` to disable all web UI endpoints. When disabled, only API endpoints are available.
- **`ENABLE_SIGNUPS`**: Set to `false` to disable user registration through the web UI. When disabled, new users must be created using the admin CLI tool (see [Admin CLI](#admin-cli) below).

### User Authentication

- Login/registration at `/login` and `/register` (when UI is enabled)
- Session-based authentication with secure password hashing
- API key management interface at `/keys`
- Registration can be disabled for security while keeping the UI available for existing users

### API Key Management

Navigate to `/keys` after logging in to:
- **Generate API Keys**: Create new keys with optional expiry dates
- **View Active Keys**: See your currently active keys (revoked/expired keys hidden by default)
- **Revoke Keys**: Instantly revoke compromised keys

Key limits are configurable per user (default: 5 keys maximum).

### Usage Statistics

The usage statistics widget shows real-time analytics with:

- **Wallets Checked**: Total number of balance queries over time
- **Interactive Chart**: Line chart with dynamic Y-axis scaling
- **Auto-Refresh**: Updates every 3 seconds for live monitoring

#### Filter Options

- **Overall**: All server usage (everyone's wallet checks)
- **My Keys**: Usage from only your API keys

#### Timeframes

- **Hour**: Last 60 minutes (1-minute intervals)
- **Day**: Last 24 hours (1-hour intervals) - *default*
- **Week**: Last 7 days (6-hour intervals)
- **Month**: Last 30 days (1-day intervals)
- **Year**: Last 365 days (1-week intervals)

## Admin CLI

When running with `ENABLE_SIGNUPS=false` or `ENABLE_UI=false`, administrators need a way to create users and API keys. The admin CLI tool provides this functionality.

### Building the Admin Tool

```bash
go build -o admin ./cmd/admin
```

### Admin Commands

#### Create a User

```bash
./admin create-user --email user@example.com --password securepassword123
```

Requirements:
- Email must be unique
- Password must be at least 12 characters

#### Create an API Key

```bash
# Create a key with no expiry
./admin create-key --email user@example.com

# Create a key with expiry date
./admin create-key --email user@example.com --expiry 2025-12-31
```

The secret will be displayed only once. Save it securely.

#### List Users

```bash
./admin list-users
```

#### List API Keys for a User

```bash
./admin list-keys --email user@example.com
```

**Note:** This will not print the actual API key secrets (which are only shown once at creation), just the key metadata (Key ID, creation date, expiration date, and status). 

### Admin CLI Configuration

The admin tool only needs access to the dogelytics database (for users and API keys). Set the database URL via environment variable or flag:

```bash
# Using environment variable
export DOGELYTICS_DBURL="postgres://dogelytics:password@localhost:5432/dogelytics?sslmode=disable"
./admin create-user --email admin@example.com --password securepass123

# Using flag
./admin create-user --email admin@example.com --password securepass123 \
  -dogelytics-dburl="postgres://dogelytics:password@localhost:5432/dogelytics?sslmode=disable"
```

## Docker Deployment

See [dogecoinfoundation/docker](https://github.com/dogecoinfoundation/docker) for containerized deployment with Docker Compose.

The Docker setup automatically:
- Connects to the indexer's PostgreSQL database
- Configures proper networking
- Sets up health checks and resource limits

When deploying with Docker, you can:
1. Mount a `.env` file into the container
2. Pass environment variables via `docker-compose.yml`
3. Use Docker secrets for sensitive values like `SESSION_SECRET`

## Security Best Practices

For production deployments:

1. **Consider Disabling Public Signups**: Unless you want to allow others to sign up for your servlce, set `ENABLE_SIGNUPS=false` to prevent unauthorized user registration
2. **Generate Strong Session Secret**: Use `openssl rand -base64 32` to generate a secure `SESSION_SECRET`
4. **Limit API Access**: Set appropriate rate limits (`RATELIMIT` and `API_KEY_RATELIMIT`)
5. **Restrict CORS**: Set `CORS` to specific domains instead of `*` in production
6. **Secure Database**: Use strong passwords and `sslmode=require` for PostgreSQL connections

## License

MIT License - see LICENSE file for details
