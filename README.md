# Dogelytics

A lightweight REST API service that provides Dogecoin wallet balance information by querying the [Dogecoin Indexer](https://github.com/dogeorg/indexer) PostgreSQL database.

## Features

- **Balance Queries**: Incoming, available, outgoing, and current balance for any Dogecoin address
- **Health Checks**: Monitor indexer, blocks, and headers sync heights from the indexer API
- **Rate Limiting**: Per-IP and per-API key rate limiting
- **API Key Management**: User accounts with API key generation, revocation, and expiry
- **Usage Statistics**: Real-time analytics with interactive charts and filtering
- **Admin UI**: Optional management interface with login, registration, key management, and usage statistics
- **Dashboard UI**: Optional public dashboard with a wallet checker and server snapshot cards
- **Meme Number Easter Egg**: Public pages highlight `69` and `420` variants and fire colour-matched confetti when clicked
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

At minimum, set the indexer database URL:

```bash
INDEXER_DBURL=postgres://indexer:yourpassword@localhost:5432/indexer?sslmode=disable
```

If the indexer API is not running on `http://localhost:8000`, also set:

```bash
INDEXER_API_URL=http://localhost:8000
```

If you enable the admin UI, also set the Dogelytics database URL and a session secret:

```bash
DOGELYTICS_DBURL=postgres://dogelytics:yourpassword@localhost:5432/dogelytics?sslmode=disable
SESSION_SECRET=your-secret-key-here
```

Generate a strong session secret with:

```bash
openssl rand -base64 32
```

`SESSION_SECRET` is required only when `ENABLE_ADMIN_UI=true`.

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

The [Dogecoin Indexer](https://github.com/dogeorg/indexer) must be running with a reachable PostgreSQL database for balance queries and a reachable HTTP API for sync heights.

## Configuration Reference

Configuration is loaded in this order (later sources override earlier ones):

1. Defaults
2. `.env` file (if present in the working directory)
3. Environment variables
4. Command-line flags

| Environment Variable | Flag | Description | Default |
|---------------------|------|-------------|---------|
| `INDEXER_DBURL` | `-indexer-dburl` | PostgreSQL URL for indexer database (blockchain data) | `postgres://indexer:changeme@localhost:5432/indexer?sslmode=disable` |
| `INDEXER_API_URL` | `-indexer-api-url` | Indexer API base URL for sync heights | `http://localhost:8000` |
| `DOGELYTICS_DBURL` | `-dogelytics-dburl` | PostgreSQL URL for dogelytics database (users, keys, sessions) | `postgres://dogelytics:changeme@localhost:5432/dogelytics?sslmode=disable` |
| `BIND` | `-bind` | HTTP server bind address | `localhost:4420` |
| `CORS` | `-cors` | CORS allowed origin (`*` for all) | `*` |
| `CONFIRMATIONS` | `-confirmations` | Confirmations required for available balance | `6` |
| `RATELIMIT` | `-ratelimit` | Max requests per IP per minute (0 = disabled) | `10` |
| `API_KEY_RATELIMIT` | `-apikey-ratelimit` | Max requests per API key per minute (0 = disabled) | `120` |
| `SESSION_SECRET` | `-session-secret` | Session HMAC secret (required when admin UI is enabled) | _(empty)_ |
| `MAX_KEYS_PER_USER` | `-max-keys-per-user` | Maximum API keys per user | `1` |
| `ENABLE_ADMIN_UI` | `-enable-admin-ui` | Enable the admin UI listener | `false` |
| `ADMIN_UI_PORT` | `-admin-ui-port` | Port for the admin UI listener | `4421` |
| `ENABLE_DASHBOARD_UI` | `-enable-dashboard-ui` | Enable the dashboard UI listener | `false` |
| `DASHBOARD_UI_PORT` | `-dashboard-ui-port` | Port for the dashboard UI listener | `4422` |
| `ENABLE_DASHBOARD_TIPS` | `-enable-dashboard-tips` | Enable the dashboard "Such coffee?" tips widget | `true` |
| `DASHBOARD_TIPS_ADDRESS` | `-dashboard-tips-address` | Dogecoin address shown in the dashboard tips widget | `DChPB3HbQgNYgWRrpeRKqNT6939rRLceNz` |
| `ENABLE_SIGNUPS` | `-enable-signups` | Enable user registration through the admin UI | `false` |

## API Reference

### API Hosts

Production deployments normally expose the public API at `https://api.dogelytics.com`. Local development uses the API listener, which defaults to `http://localhost:4420`.

When the dashboard UI is enabled, it also exposes same-origin helper routes under its own host:

- `GET /api/balance` mirrors `GET /balance` for dashboard wallet checks.
- `GET /api/conversion` mirrors `GET /conversion` for dashboard currency conversions.
- `GET /api/dashboard-stats` returns public dashboard usage and sync metrics.

### GET /balance

Returns the balance for a Dogecoin address. All values are decimal DOGE strings.

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
| 400 | `missing-parameter` | Missing `address` query parameter |
| 400 | `invalid-address` | Invalid Dogecoin address format |
| 400 | `unsupported-address` | Unsupported Dogecoin address type |
| 401 | `invalid-api-key` | Invalid, expired, or revoked API key |
| 429 | `rate-limit-exceeded` | Rate limit exceeded |
| 500 | `database-error` | Failed to read balance data |

### GET /conversion

Returns the current DOGE conversion rate for a single target currency code.

**Parameters:**
- `currency` (required): Currency code such as `usd`, `aud`, or `eur`

**Authentication (optional):** Include an API key via `Authorization: Bearer <key>` or `X-Api-Key: <key>` header for a higher rate limit.

**Example:**

```bash
curl "http://localhost:4420/conversion?currency=usd"
```

```json
{
  "currency": "usd",
  "rate": "0.23543210",
  "source": "coingecko",
  "cached": false,
  "fetched_at": "2026-05-28T01:44:00Z",
  "coingecko_updated_at": "2026-05-28T01:43:20Z"
}
```

Conversion rates are sourced from CoinGecko and cached locally in the Dogelytics database for one hour per currency. Dogelytics only refreshes a currency when it is missing from the cache or the cached row is older than one hour.

**Errors:**

| Status | Error | Description |
|--------|-------|-------------|
| 400 | `missing-parameter` | Missing `currency` query parameter |
| 400 | `invalid-currency` | Invalid currency code format |
| 400 | `unsupported-currency` | Unsupported currency code |
| 401 | `invalid-api-key` | Invalid, expired, or revoked API key |
| 429 | `rate-limit-exceeded` | Rate limit exceeded |
| 502 | `conversion-source-error` | Failed to refresh conversion rate from CoinGecko |
| 503 | `conversion-cache-unavailable` | Local conversion cache unavailable |

### GET /health

Returns service status and the raw sync heights Dogelytics reads from the indexer API.

```bash
curl "http://localhost:4420/health"
```

```json
{
  "ok": true,
  "indexer_height": 5900000,
  "core_blocks_height": 6224976,
  "core_headers_height": 6224976,
  "core_sync_updated_at": "2026-06-01T04:00:00Z"
}
```

- `indexer_height`: last block indexed by the indexer.
- `core_blocks_height`: local Core blocks height as reported by the indexer.
- `core_headers_height`: best known chain tip from headers as reported by the indexer.
- `core_sync_updated_at`: when the indexer last refreshed Core sync heights.

If the indexer sync heights are unavailable, `/health` still returns `ok` and `indexer_height`, but omits `core_blocks_height`, `core_headers_height`, and `core_sync_updated_at`.

**Errors:**

| Status | Error | Description |
|--------|-------|-------------|
| 500 | `database-error` | Failed to read indexer height |

### Rate Limiting

All requests are rate-limited per IP. Requests with a valid API key use a separate, higher limit.

| | Limit |
|---|---|
| Without API key | 10 requests/minute per IP |
| With API key | 120 requests/minute per key |

These defaults are configurable via `RATELIMIT` and `API_KEY_RATELIMIT`.

Dogelytics uses CoinGecko's public `simple/price` API for conversion data and keeps a local one-hour cache per requested currency to avoid unnecessary upstream calls.

### Dashboard Helper APIs

These routes are served from the dashboard UI host when `ENABLE_DASHBOARD_UI=true`. They exist so the browser dashboard can call Dogelytics without cross-origin setup:

| Route | Description |
|-------|-------------|
| `GET /api/balance?address=<address>` | Same response and errors as `GET /balance` |
| `GET /api/conversion?currency=<currency>` | Same response and errors as `GET /conversion` |
| `GET /api/dashboard-stats` | Public dashboard stats, including wallet lookup counters and indexer sync metrics |

Example dashboard stats response:

```json
{
  "available": true,
  "indexer_height": 5900000,
  "core_blocks_height": 6224976,
  "core_headers_height": 6224976,
  "core_sync_updated_at": "2026-06-01T04:00:00Z",
  "indexed_percent": 94.7792,
  "total_wallets_checked": 77,
  "wallets_checked_last_24h": 12,
  "unique_wallets_checked": 5,
  "unique_wallets_last_24h": 3
}
```

## User & Key Management

Users and API keys can be managed through the web UI or the admin CLI.

Each user can hold up to `MAX_KEYS_PER_USER` active keys (default: 1). Key secrets are shown only once at creation. Passwords must be at least 12 characters.

### Admin UI

Dogelytics includes an optional admin interface with a Windows 95 theme. It runs on its own port and is disabled by default.

| Operation | How |
|---|---|
| Register | `/register` (when `ENABLE_SIGNUPS=true`) |
| Log in / out | `/login`, logout button on `/keys` |
| Create API key | `/keys` → "Create New API Key" (optional expiry, max 1 year) |
| View keys | `/keys` — active keys shown by default; toggle to include revoked/expired |
| Revoke a key | `/keys` → "Revoke" button next to the key |
| Usage statistics | `/keys` — real-time chart with filter (Overall / My Keys) and timeframe (Hour, Day, Week, Month, Year) |

**Controlling access:**

- `ENABLE_ADMIN_UI=true` starts the admin UI listener on `ADMIN_UI_PORT`.
- `ENABLE_SIGNUPS=false` disables self-registration while keeping the admin UI available for existing users. When disabled, users can only be created via the admin CLI.

### Dashboard UI

Dogelytics also includes an optional public dashboard on `DASHBOARD_UI_PORT`, separate from the API and admin UI. The first dashboard focuses on:

- Looking up a wallet balance from a pasted Dogecoin address
- Showing total successful wallet checks
- Showing wallet checks in the last 24 hours
- Showing unique wallets checked in the last 24 hours
- Showing the current indexed height
- Showing a small minimisable "Such coffee?" tips widget when `ENABLE_DASHBOARD_TIPS=true`

Enable it with `ENABLE_DASHBOARD_UI=true`.

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
