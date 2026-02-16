# Dogelytics

A lightweight REST API service that provides Dogecoin wallet balance information by querying the [Dogecoin Indexer](https://github.com/dogeorg/indexer) PostgreSQL database.

## Features

- **Balance Queries**: Incoming, available, outgoing, and current balance for any Dogecoin address
- **Health Checks**: Monitor indexer status and current block height
- **Rate Limiting**: Per-IP and per-API key rate limiting
- **API Key Management**: User accounts with API key generation, revocation, and expiry
- **Usage Statistics**: Real-time analytics with interactive charts and filtering
- **Web UI**: Optional management interface with a Windows 95 theme
- **Admin CLI**: Command-line tool for managing users and API keys
- **CORS Support**: Configurable CORS headers for browser-based applications

## Prerequisites

- Go 1.21 or later
- PostgreSQL (two databases are required):
  - **Indexer database** — shared with [Dogecoin Indexer](https://github.com/dogeorg/indexer) for blockchain data (read-only)
  - **Dogelytics database** — stores users, API keys, and sessions

## Getting Started

### 1. Install Dependencies

```bash
make deps
```

Or directly:

```bash
go mod download
```

### 2. Set Up PostgreSQL

If you don't already have PostgreSQL running, pick one:

**Docker (quick start):**

```bash
docker run -d --name postgres-dev \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  -p 5432:5432 \
  postgres:16-alpine
```

**macOS (Homebrew):**

```bash
brew install postgresql@16 && brew services start postgresql@16
```

**Linux (Debian/Ubuntu):**

```bash
sudo apt install postgresql postgresql-contrib && sudo systemctl start postgresql
```

### 3. Create the Dogelytics Database

The indexer database is managed by the [Dogecoin Indexer](https://github.com/dogeorg/indexer). The dogelytics database needs to be created separately.

**Automated setup (recommended):**

```bash
./scripts/setup-auth-db.sh
```

This interactive script will:
- Connect to your PostgreSQL instance
- Auto-generate a secure random password (or accept a custom one)
- Create the `dogelytics` database and user with proper permissions
- Output the `DOGELYTICS_DBURL` connection string for your `.env` file

**Manual setup:**

```sql
-- Connect as admin
psql -U postgres

-- Create database and user
CREATE DATABASE dogelytics;
CREATE USER dogelytics WITH PASSWORD 'YOUR_SECURE_PASSWORD';
GRANT ALL PRIVILEGES ON DATABASE dogelytics TO dogelytics;

-- Connect to dogelytics database and grant schema permissions (PostgreSQL 15+)
\c dogelytics
GRANT ALL ON SCHEMA public TO dogelytics;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO dogelytics;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO dogelytics;
```

### 4. Configure

Copy the example environment file and edit it:

```bash
cp .env.example .env
```

At minimum, set the two database URLs and a session secret:

```bash
INDEXER_DBURL=postgres://indexer:yourpassword@localhost:5432/indexer?sslmode=disable
DOGELYTICS_DBURL=postgres://dogelytics:yourpassword@localhost:5432/dogelytics?sslmode=disable
SESSION_SECRET=your-secret-key-here
```

Generate a strong session secret with:

```bash
openssl rand -base64 32
```

`SESSION_SECRET` is required when `ENABLE_UI=true` (the default).

See [Configuration Reference](#configuration-reference) for all options.

### 5. Run

```bash
make run
```

Or directly:

```bash
go run ./cmd/dogelytics
```

To build a binary:

```bash
make build
./dogelytics
```

The [Dogecoin Indexer](https://github.com/dogeorg/indexer) must be running with a reachable PostgreSQL database for balance queries to work.

## Configuration Reference

Configuration is loaded in this order (later sources override earlier ones):

1. Defaults
2. `.env` file (if present in the working directory)
3. Environment variables
4. Command-line flags

| Environment Variable | Flag | Description | Default |
|---------------------|------|-------------|---------|
| `INDEXER_DBURL` | `-indexer-dburl` | PostgreSQL URL for indexer database (blockchain data) | `postgres://indexer:changeme@localhost:5432/indexer?sslmode=disable` |
| `DOGELYTICS_DBURL` | `-dogelytics-dburl` | PostgreSQL URL for dogelytics database (users, keys, sessions) | `postgres://dogelytics:changeme@localhost:5432/dogelytics?sslmode=disable` |
| `BIND` | `-bind` | HTTP server bind address | `localhost:4420` |
| `CORS` | `-cors` | CORS allowed origin (`*` for all) | `*` |
| `CONFIRMATIONS` | `-confirmations` | Confirmations required for available balance | `6` |
| `RATELIMIT` | `-ratelimit` | Max requests per IP per minute (0 = disabled) | `10` |
| `API_KEY_RATELIMIT` | `-apikey-ratelimit` | Max requests per API key per minute (0 = disabled) | `120` |
| `SESSION_SECRET` | `-session-secret` | Session HMAC secret (required when UI is enabled) | _(empty)_ |
| `MAX_KEYS_PER_USER` | `-max-keys-per-user` | Maximum API keys per user | `1` |
| `ENABLE_UI` | `-enable-ui` | Enable web UI endpoints | `true` |
| `ENABLE_SIGNUPS` | `-enable-signups` | Enable user registration through the UI | `true` |

## API Reference

### GET /balance

Returns the balance for a Dogecoin address. All values are in Koinu (1 DOGE = 100,000,000 Koinu) formatted as decimal strings.

**Parameters:**
- `address` (required): Dogecoin address

**Authentication (optional):** Include an API key via `Authorization: Bearer <key>` or `X-Api-Key: <key>` header for a higher rate limit.

**Example:**

```bash
curl "http://localhost:4420/balance?address=DLAznsPDLDRgsVcTFWRMYMG5uH6GddDtv8"
```

```json
{
  "incoming": "100000000.00000000",
  "available": "500000000.00000000",
  "outgoing": "50000000.00000000",
  "current": "600000000.00000000"
}
```

**Errors:**

| Status | Error | Description |
|--------|-------|-------------|
| 400 | `invalid-address` | Invalid Dogecoin address format |
| 401 | `invalid-api-key` | Invalid, expired, or revoked API key |
| 429 | `rate-limit-exceeded` | Rate limit exceeded |

### GET /health

Returns service status and the current indexed block height.

```bash
curl "http://localhost:4420/health"
```

```json
{
  "ok": true,
  "height": 5900000
}
```

### Rate Limiting

All requests are rate-limited per IP. Requests with a valid API key use a separate, higher limit.

| | Limit |
|---|---|
| Without API key | 10 requests/minute per IP |
| With API key | 120 requests/minute per key |

These defaults are configurable via `RATELIMIT` and `API_KEY_RATELIMIT`.

## User & Key Management

Users and API keys can be managed through the web UI or the admin CLI.

Each user can hold up to `MAX_KEYS_PER_USER` active keys (default: 1). Key secrets are shown only once at creation. Passwords must be at least 12 characters.

### Web UI

Dogelytics includes an optional web interface at `http://localhost:4420` with a Windows 95 theme. By default, both the UI and self-registration are enabled.

| Operation | How |
|---|---|
| Register | `/register` (when `ENABLE_SIGNUPS=true`) |
| Log in / out | `/login`, logout button on `/keys` |
| Create API key | `/keys` → "Create New API Key" (optional expiry, max 1 year) |
| View keys | `/keys` — active keys shown by default; toggle to include revoked/expired |
| Revoke a key | `/keys` → "Revoke" button next to the key |
| Usage statistics | `/keys` — real-time chart with filter (Overall / My Keys) and timeframe (Hour, Day, Week, Month, Year) |

**Controlling access:**

- `ENABLE_UI=false` disables all web UI endpoints, leaving only the API.
- `ENABLE_SIGNUPS=false` disables self-registration while keeping the UI available for existing users. When disabled, users can only be created via the admin CLI.

### Admin CLI

A command-line tool for managing users and API keys directly. Useful when the UI is disabled, self-registration is off, or for scripted provisioning.

**Build:**

```bash
go build -o admin ./cmd/admin
```

**Commands:**

| Operation | Command |
|---|---|
| Create user | `./admin create-user --email user@example.com --password securepass123` |
| Create API key | `./admin create-key --email user@example.com [--expiry 2026-12-31]` |
| Revoke a key | `./admin revoke-key --email user@example.com --kid <key-id>` |
| List users | `./admin list-users` |
| List keys | `./admin list-keys --email user@example.com` |

`list-keys` shows key metadata (ID, dates, status) — not the secret.

**Configuration:**

The admin tool only needs access to the dogelytics database. It loads `.env` automatically if present. The database URL can also be set via environment variable or flag:

```bash
export DOGELYTICS_DBURL="postgres://dogelytics:password@localhost:5432/dogelytics?sslmode=disable"
./admin list-users

# Or with a flag
./admin list-users -dogelytics-dburl="postgres://dogelytics:password@localhost:5432/dogelytics?sslmode=disable"
```

## Docker Deployment

See [dogecoinfoundation/docker](https://github.com/dogecoinfoundation/docker) for containerized deployment with Docker Compose.

## Security

For production deployments:

1. **Disable self-registration** if you don't want open sign-ups — set `ENABLE_SIGNUPS=false`
2. **Use a strong session secret** — generate with `openssl rand -base64 32`
3. **Set appropriate rate limits** via `RATELIMIT` and `API_KEY_RATELIMIT`
4. **Restrict CORS** — set `CORS` to specific domains instead of `*`
5. **Secure database connections** — use strong passwords and `sslmode=require` for PostgreSQL

## License

MIT License — see LICENSE file for details.
