# Dogelytics

A lightweight REST API service that provides Dogecoin wallet balance information by querying the indexer's PostgreSQL database.

## Prerequisites

- Go 1.21 or later
- **PostgreSQL database** (shared with [Dogecoin Indexer](https://github.com/dogeorg/indexer))
  - **PostgreSQL is required** - SQLite is not supported due to concurrent access issues
  - The indexer must be configured to use PostgreSQL (see [docker/indexer](../docker/indexer/))
  - Dogelytics connects to the same Postgres database as the indexer for read-only queries

## Features

- **Balance Queries**: Get incoming, available, outgoing, and current balance for any Dogecoin address
- **Health Checks**: Monitor the indexer's status and current block height
- **Rate Limiting**: Per-IP rate limiting to prevent abuse (configurable, per-minute basis)
- **CORS Support**: Configurable CORS headers for browser-based applications
- **Read-Only Database Access**: Direct Postgres queries with concurrent read support
- **Address Validation**: Validates Dogecoin addresses before querying
- **Lightweight**: Minimal overhead, optimized for high-traffic scenarios

## Installation

```bash
cd dogelytics
go mod download
```

## Quick Start

**Important**: Ensure the indexer is running with PostgreSQL before starting dogelytics.

### Run with default settings

```bash
go run ./cmd/dogelytics
```

This will:
- Connect to PostgreSQL at `postgres://indexer:changeme@localhost:5432/indexer`
- Start the HTTP server on `localhost:4420`
- Enable rate limiting at 60 requests per IP per minute

### Run with custom settings

```bash
go run ./cmd/dogelytics \
  -dburl="postgres://indexer:mypassword@localhost:5432/indexer?sslmode=disable" \
  -bind="localhost:4420" \
  -confirmations=6 \
  -cors="*" \
  -ratelimit=60
```

### Build and run

```bash
# Compile
go build -o dogelytics ./cmd/dogelytics

# Run
./dogelytics -dburl="postgres://indexer:mypassword@localhost:5432/indexer?sslmode=disable"
```

## Configuration

Dogelytics is configured via command-line flags:

| Flag | Description | Default |
|------|-------------|---------|
| `-dburl` | PostgreSQL database URL (required) | `postgres://indexer:changeme@localhost:5432/indexer?sslmode=disable` |
| `-bind` | HTTP server bind address | `localhost:4420` |
| `-cors` | CORS allowed origin (`*` for all) | `*` |
| `-confirmations` | Number of confirmations for available balance | `6` |
| `-ratelimit` | Maximum requests per IP per minute (0 = disabled) | `60` |

### Example Configurations

**Development (local Postgres)**:
```bash
go run ./cmd/dogelytics \
  -dburl="postgres://indexer:changeme@localhost:5432/indexer?sslmode=disable" \
  -ratelimit=0
```

**Production (rate limited, specific CORS)**:
```bash
./dogelytics \
  -dburl="postgres://indexer:securepassword@postgres-host:5432/indexer?sslmode=require" \
  -bind="0.0.0.0:4420" \
  -cors="https://mydogeapp.com" \
  -ratelimit=100 \
  -confirmations=6
```

**Docker network (containerized setup)**:
```bash
./dogelytics \
  -dburl="postgres://indexer:password@indexer-postgres:5432/indexer?sslmode=disable" \
  -bind="0.0.0.0:4420" \
  -cors="*" \
  -ratelimit=60
```

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

Dogelytics includes built-in rate limiting to protect against abuse and ensure fair usage. The rate limiter tracks requests per IP address within a rolling 1-minute window.

**Features:**
- Per-IP rate limiting with configurable maximum requests per minute
- Automatic cleanup of expired rate limit entries
- Support for proxy headers (`X-Forwarded-For`, `X-Real-IP`) to identify real client IPs
- Rate limiting can be disabled by setting `-ratelimit=0`

**Default:** 60 requests per IP per minute

When a client exceeds the rate limit, they will receive an HTTP 429 (Too Many Requests) response with details about the limit.

## Architecture

**Why PostgreSQL is Required:**
- **Concurrent Access**: PostgreSQL's MVCC (Multi-Version Concurrency Control) allows dogelytics to query the database while the indexer writes to it without conflicts
- **Performance**: Optimized for high-concurrency read/write workloads
- **Reliability**: Production-grade database with proper transaction isolation
- **No Locking**: Unlike SQLite, Postgres won't throw database-locked errors under concurrent access

**Architecture Diagram:**
```
Dogecoin Core → Indexer (writes) → PostgreSQL
                                        ↓
                          Dogelytics (reads) → Users/Apps
```

Both indexer and dogelytics connect to the same PostgreSQL database. The indexer writes new blocks and updates UTXOs, while dogelytics performs read-only queries for balance information.

## Docker Deployment

See [docker/dogelytics](../docker/dogelytics/) for containerized deployment with Docker Compose.

The Docker setup automatically:
- Connects to the indexer's PostgreSQL database
- Configures proper networking
- Sets up health checks and resource limits

## License

MIT License - see LICENSE file for details
