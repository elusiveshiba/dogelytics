# Dogelytics

<p align="center">
  <img src="img/dogelytics%20icon.png" alt="Dogelytics logo" width="140">
</p>

Dogelytics is a small HTTP service for Dogecoin wallet balances, currency conversion, API-key access and privacy-preserving usage analytics. It reads balances and chain progress from the [Dogecoin Indexer](https://github.com/dogeorg/indexer) HTTP API and stores its own application data in PostgreSQL.

This release builds its container from this repository. There is no published Dogelytics registry image yet.

## Requirements

- Docker Engine with Docker Compose v2 for the supported deployment path.
- A reachable Dogecoin Indexer HTTP API. Dogecoin Core and the indexer are installed and operated separately by the Foundation or machine operator.
- Git and OpenSSL for the quick start.

PostgreSQL is required. The supplied Compose stack owns a dedicated PostgreSQL 16 instance for Dogelytics; it does not own Dogecoin Core, the indexer or the indexer's database.

Balance availability follows the indexer's fixed six-confirmation behaviour. Dogelytics does not configure that confirmation depth.

## Quick start

Clone the repository, generate local secrets, review the non-secret settings, then build and start the stack:

```bash
git clone https://github.com/elusiveshiba/dogelytics.git
cd dogelytics
./scripts/setup-compose.sh
${EDITOR:-vi} .env
docker compose up -d --build --wait
```

The setup script creates ignored files under `secrets/` and creates `.env` from `compose.env.example`. Running it again keeps every existing secret and configuration file unchanged.

The default listeners are bound to localhost:

| Service | URL |
|---|---|
| Public API | `http://127.0.0.1:4420` |
| Admin UI | `http://127.0.0.1:4421` |
| Public dashboard and docs | `http://127.0.0.1:4422` |

Check readiness and logs with:

```bash
docker compose exec dogelytics /dogelytics healthcheck
docker compose logs -f dogelytics
```

Create the first operator. The password prompt does not echo:

```bash
docker compose run --rm dogelytics admin user create --email operator@example.com
docker compose run --rm dogelytics admin key create --email operator@example.com
```

The API-key secret is printed once. Store it securely.

## Connecting an indexer

Dogelytics only needs the indexer's HTTP base URL. It never connects to the indexer's PostgreSQL database.

### Indexer on the Docker host

The Compose default is:

```dotenv
INDEXER_API_URL=http://host.docker.internal:8000
```

`compose.yaml` adds the Linux `host-gateway` mapping. The indexer must listen on an address reachable from Docker; a service bound exclusively to host loopback may not be reachable on Linux.

### Indexer on another machine

Set an HTTP or HTTPS URL that the Dogelytics container can reach:

```dotenv
INDEXER_API_URL=http://192.0.2.20:8000
```

Restrict the indexer port at the network or host firewall. Dogelytics requires `GET /height` and `GET /balance?address=...`.

### Indexer on an existing Docker network

Set the existing network and the indexer's service URL:

```dotenv
INDEXER_DOCKER_NETWORK=dogecoin-network
INDEXER_API_URL=http://dogecoin-indexer:8000
```

Start Compose with the network override:

```bash
docker compose -f compose.yaml -f compose.indexer-network.yaml up -d --build --wait
```

The network must already exist. Dogelytics still keeps its PostgreSQL service on its own private network.

## API

The machine-readable OpenAPI 3.1 contract is available at `GET /openapi.yaml` on the API and dashboard listeners. Human-readable API documentation is available at `/docs` on the dashboard listener.

Common requests:

```bash
curl http://127.0.0.1:4420/livez
curl http://127.0.0.1:4420/readyz
curl http://127.0.0.1:4420/health
curl 'http://127.0.0.1:4420/balance?address=DLAznsPDLDRgsVcTFWRMYMG5uH6GddDtv8'
curl 'http://127.0.0.1:4420/conversion?currency=aud'
```

Balance values are validated decimal strings rather than machine-sized integers, so large address totals remain exact. An API key is optional and selects the higher per-key request limit:

```bash
curl -H 'Authorization: Bearer dglk_KID.SECRET' \
  'http://127.0.0.1:4420/balance?address=DLAznsPDLDRgsVcTFWRMYMG5uH6GddDtv8'
```

Conversion quotes are cached for one hour. Concurrent refreshes for the same currency are combined, and an older cached quote is returned with `source: stale-cache` if the conversion provider is temporarily unavailable.

## Administration commands

The single `dogelytics` binary provides all runtime and administration commands. Running it without a command remains equivalent to `serve`.

```text
dogelytics serve [server flags]
dogelytics version
dogelytics healthcheck [--url URL]
dogelytics admin user create --email EMAIL [--password-stdin]
dogelytics admin user list
dogelytics admin key create --email EMAIL [--expires YYYY-MM-DD]
dogelytics admin key list [--email EMAIL]
dogelytics admin key revoke --kid KID
```

Use `--password-stdin` for non-interactive provisioning. Do not put a password on the command line. Compose administration commands use the same secret-mounted database configuration and migrations as the server:

```bash
docker compose run --rm dogelytics admin user list
docker compose run --rm dogelytics admin key list --email operator@example.com
docker compose run --rm dogelytics admin key revoke --kid KEY_ID
```

## Configuration

Server configuration is loaded from defaults, `.env`, environment variables and finally command flags. Existing environment variables are never overwritten by `.env`.

| Environment | Flag | Default | Purpose |
|---|---|---|---|
| `INDEXER_API_URL` | `-indexer-api-url` | `http://localhost:8000` | Required indexer HTTP API |
| `DOGELYTICS_DBURL` | `-dogelytics-dburl` | none | Required PostgreSQL URL |
| `BIND` | `-bind` | `localhost:4420` | Public API listener |
| `CORS` | `-cors` | `*` | Comma-separated allowed origins, or `*` |
| `PUBLIC_URL` | `-public-url` | none | External HTTP(S) origin, without a path |
| `TRUSTED_PROXIES` | `-trusted-proxies` | none | Comma-separated proxy IPs/CIDRs allowed to supply forwarded headers |
| `RATELIMIT` | `-ratelimit` | `10` | Requests per IP per minute; `0` disables |
| `API_KEY_RATELIMIT` | `-apikey-ratelimit` | `120` | Requests per API key per minute; `0` disables |
| `SESSION_SECRET` | `-session-secret` | none | Admin session HMAC secret, minimum 32 characters |
| `ANALYTICS_SECRET` | `-analytics-secret` | none | Analytics fingerprint HMAC secret, minimum 32 characters |
| `ENABLE_ANALYTICS` | `-enable-analytics` | `true` | Store private analytics and rollups |
| `ANALYTICS_RETENTION_DAYS` | `-analytics-retention-days` | `30` | Fingerprinted raw-event retention |
| `MAX_KEYS_PER_USER` | `-max-keys-per-user` | `1` | Maximum active keys per user |
| `ENABLE_ADMIN_UI` | `-enable-admin-ui` | `false` | Start the admin listener |
| `ADMIN_UI_PORT` | `-admin-ui-port` | `4421` | Admin listener port |
| `ENABLE_DASHBOARD_UI` | `-enable-dashboard-ui` | `false` | Start the dashboard listener |
| `DASHBOARD_UI_PORT` | `-dashboard-ui-port` | `4422` | Dashboard listener port |
| `ENABLE_DASHBOARD_TIPS` | `-enable-dashboard-tips` | `true` | Show the dashboard tips widget |
| `DASHBOARD_TIPS_ADDRESS` | `-dashboard-tips-address` | project address | Tips wallet address |
| `ENABLE_SIGNUPS` | `-enable-signups` | `false` | Allow admin-UI self-registration |

`DOGELYTICS_DBURL_FILE`, `SESSION_SECRET_FILE` and `ANALYTICS_SECRET_FILE` are supported for mounted secrets. A direct environment value takes precedence over its `_FILE` variant. The Compose stack uses files under `/run/secrets` and does not place secret values in container environment variables.

Configuration is validated before listeners start. PostgreSQL and the indexer are required dependencies; `/readyz` returns unavailable when either cannot be reached.

## Reverse proxy and public access

The default localhost port publishing is deliberate. Put TLS and public routing in a reverse proxy, then set:

```dotenv
PUBLIC_URL=https://api.example.org
TRUSTED_PROXIES=172.18.0.0/16
CORS=https://app.example.org
```

Only add addresses or CIDRs that belong to the proxy. Forward `Host`, `X-Forwarded-For` and `X-Forwarded-Proto`; Dogelytics ignores forwarded client/protocol data from untrusted peers. `PUBLIC_URL` controls secure cookies and same-origin CSRF checks. Route the API, admin and dashboard hosts or ports separately if all three are enabled.

Keep `ENABLE_SIGNUPS=false` unless public registration is intentional. API-key authentication attempts and password login attempts are rate limited before expensive verification.

## Privacy and data retention

Dogelytics stores HMAC fingerprints, never raw client IP addresses or wallet addresses. Successful-request totals, unique-wallet state and hourly rollups are retained for all-time statistics, including history belonging to revoked API keys. Fingerprinted request events are removed after `ANALYTICS_RETENTION_DAYS` (30 by default).

Keep `ANALYTICS_SECRET` stable if unique counts must remain comparable across restarts. Rotating it prevents future events from correlating with existing fingerprints; plan a data reset if rotation is required.

API-key secrets use constant-time SHA-256 verification. Existing bcrypt API-key records remain valid and are automatically rewritten after their next successful use. User passwords remain bcrypt hashes.

## Backup, upgrade and rollback

Back up PostgreSQL before every upgrade:

```bash
mkdir -p backups
docker compose exec -T postgres \
  pg_dump -U dogelytics -d dogelytics -Fc > "backups/dogelytics-$(date +%Y%m%d-%H%M%S).dump"
```

Also back up `.env` and `secrets/` through a secure secret-management channel. They are intentionally excluded from Git.

To upgrade a checked-out deployment:

```bash
git fetch --tags
git switch --detach v0.1.0
docker compose build --pull
docker compose up -d --wait
```

Database migrations are embedded, ordered and transactional. To roll back application code, restore the previous Git ref and rebuild. If that version cannot read the migrated schema, stop the stack, restore the pre-upgrade database dump into a new empty PostgreSQL volume, then start the previous version. Test backup restoration before relying on it in production.

`docker compose down` keeps the database volume. Do not use `docker compose down --volumes` on a real deployment unless permanent database deletion is intended.

## Native development

Native development requires Go 1.25.12 or a newer patched release, PostgreSQL and a reachable indexer:

```bash
cp .env.example .env
${EDITOR:-vi} .env
make check
make run
```

Useful targets are shown by `make help`. PostgreSQL integration tests require a disposable database:

```bash
DOGELYTICS_TEST_DBURL='postgres://dogelytics:password@127.0.0.1:5432/dogelytics?sslmode=disable' \
  make test-integration
```

The CI workflow runs formatting, unit and PostgreSQL integration tests, race detection, vet, Staticcheck, Go vulnerability analysis, Compose smoke tests, Trivy image scanning and non-published amd64/arm64 builds.

## Dogebox

`pup.nix` packages v0.1.0 with Go 1.25 and PostgreSQL 16. The pup obtains the indexer URL from Dogebox's `indexer-api` interface, generates separate database/session/analytics secrets on first start, persists them under `/storage`, runs migrations automatically and exposes the API, admin and dashboard listeners.

The web assets are embedded in the binary, so the package does not need a separate runtime asset directory.

## Release process

See [docs/release-checklist.md](docs/release-checklist.md). The v0.1.0 release candidate is repository-local: CI validates the Dockerfile and Compose stack, but no container is logged in, tagged for a registry or published.

## Licence

MIT — see [LICENSE](LICENSE).
