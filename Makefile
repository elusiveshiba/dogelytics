.PHONY: build run test test-race test-integration vet fmt fmt-check check docker-setup docker-build docker-up docker-down help

GOCACHE ?= /tmp/dogelytics-go-cache
GOMODCACHE ?= /tmp/dogelytics-go-mod-cache
GOENV = GOCACHE=$(GOCACHE) GOMODCACHE=$(GOMODCACHE)

build:
	$(GOENV) go build -trimpath -o dogelytics ./cmd/dogelytics

run:
	$(GOENV) go run ./cmd/dogelytics serve

test:
	$(GOENV) go test ./...

test-race:
	$(GOENV) go test -race ./...

test-integration:
	@test -n "$(DOGELYTICS_TEST_DBURL)" || (echo "DOGELYTICS_TEST_DBURL is required" >&2; exit 1)
	$(GOENV) go test ./internal/store ./cmd/dogelytics -run Integration -v

vet:
	$(GOENV) go vet ./...

fmt:
	gofmt -w $$(find cmd internal img -name '*.go' -type f)

fmt-check:
	@test -z "$$(gofmt -l cmd internal img)" || (gofmt -l cmd internal img; exit 1)

check: fmt-check test test-race vet
	git diff --check

docker-setup:
	./scripts/setup-compose.sh

docker-build:
	docker compose build

docker-up:
	docker compose up -d --build --wait

docker-down:
	docker compose down

help:
	@echo "Dogelytics targets:"
	@echo "  build             Build the native binary"
	@echo "  run               Start the native server"
	@echo "  test              Run all Go tests"
	@echo "  test-race         Run all Go tests with race detection"
	@echo "  test-integration  Run PostgreSQL tests (DOGELYTICS_TEST_DBURL required)"
	@echo "  vet               Run go vet"
	@echo "  fmt / fmt-check   Format or verify Go source"
	@echo "  check             Run the local release checks"
	@echo "  docker-setup      Generate non-overwriting Compose secrets"
	@echo "  docker-build      Build the repository-local image"
	@echo "  docker-up/down    Start or stop the Compose stack"

.DEFAULT_GOAL := help
