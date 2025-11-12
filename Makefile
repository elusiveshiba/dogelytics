.PHONY: build run test clean help

# Build the dogelytics binary
build:
	@echo "Building dogelytics..."
	@go build -o dogelytics ./cmd/dogelytics
	@echo "Build complete: ./dogelytics"

# Run the service with default settings
run:
	@echo "Starting dogelytics..."
	@go run ./cmd/dogelytics

# Run with custom database path
run-custom:
	@echo "Starting dogelytics with custom settings..."
	@go run ./cmd/dogelytics -dbpath="$(DBPATH)" -bind="$(BIND)" -confirmations=$(CONFIRMATIONS)

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -f dogelytics
	@echo "Clean complete"

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy
	@echo "Dependencies updated"

# Run go vet
vet:
	@echo "Running go vet..."
	@go vet ./...

# Run go fmt
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Test the API (requires service to be running)
test-api:
	@echo "Testing health endpoint..."
	@curl -s http://localhost:4420/health | jq
	@echo "\nTesting balance endpoint (replace with valid address)..."
	@echo "curl http://localhost:4420/balance?address=YOUR_DOGE_ADDRESS"

# Show help
help:
	@echo "Dogelytics Makefile"
	@echo ""
	@echo "Available targets:"
	@echo "  build       - Build the dogelytics binary"
	@echo "  run         - Run the service with default settings"
	@echo "  run-custom  - Run with custom settings (requires DBPATH, BIND, CONFIRMATIONS env vars)"
	@echo "  clean       - Remove build artifacts"
	@echo "  deps        - Download and tidy dependencies"
	@echo "  vet         - Run go vet"
	@echo "  fmt         - Format code with go fmt"
	@echo "  test        - Run unit tests"
	@echo "  test-api    - Test API endpoints (requires service to be running)"
	@echo "  help        - Show this help message"
	@echo ""
	@echo "Examples:"
	@echo "  make build"
	@echo "  make run"
	@echo "  DBPATH=../indexer/index.db BIND=localhost:9090 CONFIRMATIONS=6 make run-custom"

# Default target
.DEFAULT_GOAL := help

