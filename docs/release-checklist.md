# Dogelytics v0.1.0 release checklist

Run this checklist from a clean release-candidate checkout. It intentionally contains no container-registry publication step.

## Code and database

- [ ] `make check` passes.
- [ ] Staticcheck 2026.1 passes: `staticcheck ./...`.
- [ ] `govulncheck ./...` passes using the current patched Go toolchain.
- [ ] PostgreSQL integration tests pass against a disposable PostgreSQL 16 database: `DOGELYTICS_TEST_DBURL=... make test-integration`.
- [ ] A copy of production data successfully migrates and contains no `client_ip` or `wallet_address` columns after start.
- [ ] A legacy bcrypt API key authenticates once and is rewritten with a `sha256:` hash.

## Container and packaging

- [ ] `./scripts/setup-compose.sh` creates secrets on first run and leaves them unchanged on a second run.
- [ ] `docker compose config --quiet` passes.
- [ ] Host or remote indexer mode reaches `/readyz`.
- [ ] Existing-network mode validates and starts with `compose.indexer-network.yaml`.
- [ ] `docker compose up -d --build --wait` passes on amd64 and arm64.
- [ ] The Dogelytics container runs as `65532:65532`, has a read-only root filesystem and drops all capabilities.
- [ ] Trivy reports no fixed HIGH or CRITICAL vulnerability in `dogelytics:local`.
- [ ] `nix-build pup.nix -A dogelytics` passes in OrbStack Linux.
- [ ] A fresh Dogebox install creates unique database, session and analytics secrets and restarts without changing them.

## Behaviour and operations

- [ ] `/livez`, `/readyz`, `/health`, `/balance`, `/conversion` and `/openapi.yaml` match the documented contract.
- [ ] Large decimal balance values survive indexer decoding without rounding or overflow.
- [ ] An indexer failure produces the documented 502 or 503 response and never leaks its response body.
- [ ] A conversion-provider failure returns a stale cached quote when one exists.
- [ ] Admin user/key create, list and revoke commands work through `docker compose run --rm dogelytics ...`.
- [ ] Reverse-proxy client IP and HTTPS detection work only for configured trusted proxies.
- [ ] Cross-origin admin form posts fail and configured CORS origins succeed.
- [ ] A PostgreSQL backup can be restored into an empty volume and the restored stack becomes ready.

## Documentation and release state

- [ ] Every README command has been run against the release candidate.
- [ ] `.env.example`, `compose.env.example`, Compose and the configuration table agree.
- [ ] No configurable confirmation-depth setting, direct indexer-database setting, standalone admin binary or optional-PostgreSQL claim remains.
- [ ] No Docker Hub login, registry tag, image push or Dogelytics registry-image reference exists.
- [ ] `dogelytics version` reports `v0.1.0` and the intended commit when built with release linker values.
- [ ] `git status --short` is empty and the chronological commit list has been reviewed.
