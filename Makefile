# Variables
APP_NAME=mcp-octo-enigma
VERSION=$(shell git describe --tags --always --dirty)
BUILD_TIME=$(shell date +%Y-%m-%dT%H:%M:%S%z)
LDFLAGS=-ldflags "-X main.version=${VERSION} -X main.buildTime=${BUILD_TIME}"

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt

# Directories
BUILD_DIR=build
CMD_DIR=cmd
INTERNAL_DIR=internal

# Binaries
SERVER_BINARY=$(BUILD_DIR)/mcp-server
CLIENT_BINARY=$(BUILD_DIR)/mcp-client

.PHONY: all build clean test test-coverage test-unit test-integration test-db deps fmt lint docker-build docker-up docker-down help

all: clean deps fmt test build

# Build the applications
build: build-server build-client

build-server:
	@echo "Building server..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(SERVER_BINARY) ./$(CMD_DIR)/server

build-client:
	@echo "Building client..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(CLIENT_BINARY) ./$(CMD_DIR)/client

# Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	@rm -rf $(BUILD_DIR)

# Install dependencies
deps:
	@echo "Installing dependencies..."
	$(GOMOD) download
	$(GOMOD) verify

# Format code
fmt:
	@echo "Formatting code..."
	$(GOFMT) ./...

# Lint code
lint:
	@echo "Linting code..."
	golangci-lint run ./...

# Run all tests
test: test-unit test-integration

# Run unit tests
test-unit:
	@echo "Running unit tests..."
	$(GOTEST) -v -race -coverprofile=coverage.out ./tests/unit/...

# Run integration tests
test-integration:
	@echo "Running integration tests..."
	$(GOTEST) -v -race ./tests/integration/...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -race -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

# Start test database
test-db:
	@echo "Starting test database..."
	docker-compose -f docker-compose.test.yml up -d
	@echo "Waiting for database to be ready..."
	@sleep 10

# Stop test database
test-db-down:
	@echo "Stopping test database..."
	docker-compose -f docker-compose.test.yml down

# Run database migrations
migrate-up:
	@echo "Running database migrations..."
	migrate -path migrations -database "$(DATABASE_URL)" up

migrate-down:
	@echo "Rolling back database migrations..."
	migrate -path migrations -database "$(DATABASE_URL)" down

# Docker commands
docker-build:
	@echo "Building Docker image..."
	docker build -t $(APP_NAME):$(VERSION) .
	docker build -t $(APP_NAME):latest .

docker-up:
	@echo "Starting Docker services..."
	docker-compose up -d

docker-down:
	@echo "Stopping Docker services..."
	docker-compose down

docker-logs:
	@echo "Showing Docker logs..."
	docker-compose logs -f

# Development commands
dev: test-db
	@echo "Starting development environment..."
	$(GOMOD) download
	air -c .air.toml

dev-down: test-db-down

# Install development tools
install-tools:
	@echo "Installing development tools..."
	go install github.com/cosmtrek/air@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Generate code
generate:
	@echo "Generating code..."
	$(GOCMD) generate ./...

# Run benchmarks
bench:
	@echo "Running benchmarks..."
	$(GOTEST) -bench=. -benchmem ./...

# Security scan
security:
	@echo "Running security scan..."
	gosec ./...

# Dependency check
deps-check:
	@echo "Checking dependencies..."
	$(GOMOD) verify
	govulncheck ./...

# Release build
release: clean deps test
	@echo "Building release..."
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-linux-amd64 ./$(CMD_DIR)/server
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-darwin-amd64 ./$(CMD_DIR)/server
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-windows-amd64.exe ./$(CMD_DIR)/server

# Help
help:
	@echo "Available commands:"
	@echo "  build         - Build the application"
	@echo "  clean         - Clean build artifacts"
	@echo "  deps          - Install dependencies"
	@echo "  fmt           - Format code"
	@echo "  lint          - Lint code"
	@echo "  test          - Run all tests"
	@echo "  test-unit     - Run unit tests"
	@echo "  test-integration - Run integration tests"
	@echo "  test-coverage - Run tests with coverage"
	@echo "  test-db       - Start test database"
	@echo "  migrate-up    - Run database migrations"
	@echo "  migrate-down  - Rollback database migrations"
	@echo "  docker-build  - Build Docker image"
	@echo "  docker-up     - Start Docker services"
	@echo "  docker-down   - Stop Docker services"
	@echo "  dev           - Start development environment"
	@echo "  install-tools - Install development tools"
	@echo "  generate      - Generate code"
	@echo "  bench         - Run benchmarks"
	@echo "  security      - Run security scan"
	@echo "  release       - Build release binaries"
	@echo "  help          - Show this help"