# MCP Octo Enigma - Advanced MCP Server with Genkit
# Makefile for development and deployment

.PHONY: help build run test clean deps migrate-up migrate-down docker-build docker-up docker-down lint format vet check test-unit test-integration test-e2e bench coverage docs swagger install-tools dev-setup

# Default target
.DEFAULT_GOAL := help

# Variables
APP_NAME := mcp-octo-enigma
VERSION := 1.0.0
BUILD_TIME := $(shell date +%Y-%m-%dT%H:%M:%S%z)
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Go parameters
GOCMD := go
GOBUILD := $(GOCMD) build
GOCLEAN := $(GOCMD) clean
GOTEST := $(GOCMD) test
GOGET := $(GOCMD) get
GOMOD := $(GOCMD) mod
GOFMT := $(GOCMD) fmt
GOVET := $(GOCMD) vet

# Build flags
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)"

# Directories
SRC_DIR := .
CMD_DIR := cmd
INTERNAL_DIR := internal
MIGRATIONS_DIR := migrations
TESTS_DIR := tests
DOCS_DIR := docs

# Binary names
SERVER_BINARY := bin/$(APP_NAME)-server
CLIENT_BINARY := bin/$(APP_NAME)-client

# Database
DB_HOST := localhost
DB_PORT := 5432
DB_USER := mcp_user
DB_PASSWORD := mcp_password
DB_NAME := mcp_octo_enigma
DB_URL := postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable

# Docker
DOCKER_IMAGE_SERVER := $(APP_NAME)-server
DOCKER_IMAGE_CLIENT := $(APP_NAME)-client
DOCKER_TAG := $(VERSION)

# Colors for output
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[0;33m
BLUE := \033[0;34m
PURPLE := \033[0;35m
CYAN := \033[0;36m
WHITE := \033[0;37m
RESET := \033[0m

## Help
help: ## Show this help message
	@echo "$(CYAN)MCP Octo Enigma - Advanced MCP Server with Genkit$(RESET)"
	@echo "$(YELLOW)Version: $(VERSION)$(RESET)"
	@echo ""
	@echo "$(WHITE)Available commands:$(RESET)"
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z_-]+:.*##/ { printf "  $(GREEN)%-20s$(RESET) %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

## Development Setup
dev-setup: ## Setup development environment
	@echo "$(BLUE)Setting up development environment...$(RESET)"
	@mkdir -p bin
	@mkdir -p logs
	@mkdir -p config
	@cp .env.example .env 2>/dev/null || echo "$(YELLOW)Warning: .env.example not found$(RESET)"
	@echo "$(GREEN)Development environment setup complete!$(RESET)"

install-tools: ## Install development tools
	@echo "$(BLUE)Installing development tools...$(RESET)"
	$(GOGET) -u github.com/swaggo/swag/cmd/swag
	$(GOGET) -u github.com/golangci/golangci-lint/cmd/golangci-lint
	$(GOGET) -u github.com/air-verse/air
	$(GOGET) -u github.com/golang-migrate/migrate/v4/cmd/migrate
	@echo "$(GREEN)Development tools installed!$(RESET)"

## Dependencies
deps: ## Download dependencies
	@echo "$(BLUE)Downloading dependencies...$(RESET)"
	$(GOMOD) download
	$(GOMOD) tidy
	@echo "$(GREEN)Dependencies downloaded!$(RESET)"

deps-update: ## Update dependencies
	@echo "$(BLUE)Updating dependencies...$(RESET)"
	$(GOGET) -u ./...
	$(GOMOD) tidy
	@echo "$(GREEN)Dependencies updated!$(RESET)"

## Build
build: deps ## Build the application
	@echo "$(BLUE)Building $(APP_NAME)...$(RESET)"
	@mkdir -p bin
	$(GOBUILD) $(LDFLAGS) -o $(SERVER_BINARY) ./$(CMD_DIR)/server
	$(GOBUILD) $(LDFLAGS) -o $(CLIENT_BINARY) ./$(CMD_DIR)/client
	@echo "$(GREEN)Build complete!$(RESET)"

build-server: deps ## Build server only
	@echo "$(BLUE)Building server...$(RESET)"
	@mkdir -p bin
	$(GOBUILD) $(LDFLAGS) -o $(SERVER_BINARY) ./$(CMD_DIR)/server
	@echo "$(GREEN)Server build complete!$(RESET)"

build-client: deps ## Build client only
	@echo "$(BLUE)Building client...$(RESET)"
	@mkdir -p bin
	$(GOBUILD) $(LDFLAGS) -o $(CLIENT_BINARY) ./$(CMD_DIR)/client
	@echo "$(GREEN)Client build complete!$(RESET)"

build-linux: deps ## Build for Linux
	@echo "$(BLUE)Building for Linux...$(RESET)"
	@mkdir -p bin
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(SERVER_BINARY)-linux ./$(CMD_DIR)/server
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(CLIENT_BINARY)-linux ./$(CMD_DIR)/client
	@echo "$(GREEN)Linux build complete!$(RESET)"

## Run
run: build ## Run the server
	@echo "$(BLUE)Starting MCP server...$(RESET)"
	./$(SERVER_BINARY)

run-server: build-server ## Run server only
	@echo "$(BLUE)Starting MCP server...$(RESET)"
	./$(SERVER_BINARY)

run-client: build-client ## Run client only
	@echo "$(BLUE)Starting MCP client...$(RESET)"
	./$(CLIENT_BINARY)

dev: ## Run with hot reload
	@echo "$(BLUE)Starting development server with hot reload...$(RESET)"
	air -c .air.toml

dev-server: ## Run server with hot reload
	@echo "$(BLUE)Starting development server with hot reload...$(RESET)"
	air -c .air.toml

## Database
db-setup: ## Setup database
	@echo "$(BLUE)Setting up database...$(RESET)"
	@echo "Creating database $(DB_NAME)..."
	createdb -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) $(DB_NAME) 2>/dev/null || echo "$(YELLOW)Database may already exist$(RESET)"
	@echo "$(GREEN)Database setup complete!$(RESET)"

db-drop: ## Drop database
	@echo "$(BLUE)Dropping database...$(RESET)"
	@echo "Are you sure you want to drop database $(DB_NAME)? [y/N]" && read ans && [ $${ans:-N} = y ]
	dropdb -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) $(DB_NAME)
	@echo "$(GREEN)Database dropped!$(RESET)"

migrate-up: ## Run database migrations up
	@echo "$(BLUE)Running database migrations...$(RESET)"
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" up
	@echo "$(GREEN)Migrations completed!$(RESET)"

migrate-down: ## Run database migrations down
	@echo "$(BLUE)Rolling back database migrations...$(RESET)"
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" down
	@echo "$(GREEN)Migrations rolled back!$(RESET)"

migrate-force: ## Force migration version
	@echo "$(BLUE)Forcing migration version...$(RESET)"
	@echo "Enter version number:" && read version
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" force $$version
	@echo "$(GREEN)Migration version forced!$(RESET)"

migrate-create: ## Create new migration
	@echo "$(BLUE)Creating new migration...$(RESET)"
	@echo "Enter migration name:" && read name
	migrate create -ext sql -dir $(MIGRATIONS_DIR) $$name
	@echo "$(GREEN)Migration created!$(RESET)"

## Testing
test: test-unit test-integration ## Run all tests

test-unit: ## Run unit tests
	@echo "$(BLUE)Running unit tests...$(RESET)"
	$(GOTEST) -v -race -coverprofile=coverage.out ./$(INTERNAL_DIR)/...
	@echo "$(GREEN)Unit tests completed!$(RESET)"

test-integration: ## Run integration tests
	@echo "$(BLUE)Running integration tests...$(RESET)"
	$(GOTEST) -v -tags=integration ./$(TESTS_DIR)/integration/...
	@echo "$(GREEN)Integration tests completed!$(RESET)"

test-e2e: ## Run end-to-end tests
	@echo "$(BLUE)Running end-to-end tests...$(RESET)"
	$(GOTEST) -v -tags=e2e ./$(TESTS_DIR)/e2e/...
	@echo "$(GREEN)E2E tests completed!$(RESET)"

test-race: ## Run tests with race detection
	@echo "$(BLUE)Running tests with race detection...$(RESET)"
	$(GOTEST) -race ./$(INTERNAL_DIR)/...
	@echo "$(GREEN)Race detection tests completed!$(RESET)"

bench: ## Run benchmarks
	@echo "$(BLUE)Running benchmarks...$(RESET)"
	$(GOTEST) -bench=. -benchmem ./$(INTERNAL_DIR)/...
	@echo "$(GREEN)Benchmarks completed!$(RESET)"

coverage: test-unit ## Generate coverage report
	@echo "$(BLUE)Generating coverage report...$(RESET)"
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)Coverage report generated: coverage.html$(RESET)"

## Code Quality
lint: ## Run linter
	@echo "$(BLUE)Running linter...$(RESET)"
	golangci-lint run
	@echo "$(GREEN)Linting completed!$(RESET)"

lint-fix: ## Run linter with auto-fix
	@echo "$(BLUE)Running linter with auto-fix...$(RESET)"
	golangci-lint run --fix
	@echo "$(GREEN)Linting with auto-fix completed!$(RESET)"

format: ## Format code
	@echo "$(BLUE)Formatting code...$(RESET)"
	go fmt ./...
	@echo "$(GREEN)Code formatted!$(RESET)"

vet: ## Run go vet
	@echo "$(BLUE)Running go vet...$(RESET)"
	$(GOVET) ./$(INTERNAL_DIR)/...
	@echo "$(GREEN)Go vet completed!$(RESET)"

check: lint vet test-unit ## Run all code quality checks

## Documentation
docs: swagger ## Generate documentation

swagger: ## Generate Swagger documentation
	@echo "$(BLUE)Generating Swagger documentation...$(RESET)"
	swag init -g $(CMD_DIR)/server/main.go -o $(DOCS_DIR)/swagger
	@echo "$(GREEN)Swagger documentation generated!$(RESET)"

docs-serve: swagger ## Serve documentation
	@echo "$(BLUE)Serving documentation at http://localhost:8081...$(RESET)"
	@echo "$(YELLOW)Press Ctrl+C to stop$(RESET)"
	python3 -m http.server 8081 -d $(DOCS_DIR)/swagger

## Docker
docker-build: ## Build Docker images
	@echo "$(BLUE)Building Docker images...$(RESET)"
	docker build -t $(DOCKER_IMAGE_SERVER):$(DOCKER_TAG) -f Dockerfile.server .
	docker build -t $(DOCKER_IMAGE_CLIENT):$(DOCKER_TAG) -f Dockerfile.client .
	@echo "$(GREEN)Docker images built!$(RESET)"

docker-build-server: ## Build server Docker image
	@echo "$(BLUE)Building server Docker image...$(RESET)"
	docker build -t $(DOCKER_IMAGE_SERVER):$(DOCKER_TAG) -f Dockerfile.server .
	@echo "$(GREEN)Server Docker image built!$(RESET)"

docker-build-client: ## Build client Docker image
	@echo "$(BLUE)Building client Docker image...$(RESET)"
	docker build -t $(DOCKER_IMAGE_CLIENT):$(DOCKER_TAG) -f Dockerfile.client .
	@echo "$(GREEN)Client Docker image built!$(RESET)"

docker-up: ## Start Docker containers
	@echo "$(BLUE)Starting Docker containers...$(RESET)"
	docker-compose up -d
	@echo "$(GREEN)Docker containers started!$(RESET)"

docker-down: ## Stop Docker containers
	@echo "$(BLUE)Stopping Docker containers...$(RESET)"
	docker-compose down
	@echo "$(GREEN)Docker containers stopped!$(RESET)"

docker-logs: ## Show Docker logs
	@echo "$(BLUE)Showing Docker logs...$(RESET)"
	docker-compose logs -f

docker-shell: ## Open shell in server container
	@echo "$(BLUE)Opening shell in server container...$(RESET)"
	docker-compose exec server sh

docker-test: ## Run tests in Docker
	@echo "$(BLUE)Running tests in Docker...$(RESET)"
	docker-compose exec server make test
	@echo "$(GREEN)Docker tests completed!$(RESET)"

## Cleanup
clean: ## Clean build artifacts
	@echo "$(BLUE)Cleaning build artifacts...$(RESET)"
	$(GOCLEAN)
	rm -rf bin/
	rm -rf logs/
	rm -f coverage.out coverage.html
	@echo "$(GREEN)Cleanup completed!$(RESET)"

clean-deps: ## Clean dependencies
	@echo "$(BLUE)Cleaning dependencies...$(RESET)"
	$(GOMOD) clean
	@echo "$(GREEN)Dependencies cleaned!$(RESET)"

clean-all: clean clean-deps ## Clean everything
	@echo "$(BLUE)Cleaning everything...$(RESET)"
	rm -rf vendor/
	@echo "$(GREEN)Complete cleanup done!$(RESET)"

## Security
security-scan: ## Run security scan
	@echo "$(BLUE)Running security scan...$(RESET)"
	gosec ./$(INTERNAL_DIR)/...
	@echo "$(GREEN)Security scan completed!$(RESET)"

audit: ## Audit dependencies
	@echo "$(BLUE)Auditing dependencies...$(RESET)"
	$(GOCMD) list -json -deps ./... | nancy sleuth
	@echo "$(GREEN)Dependency audit completed!$(RESET)"

## Performance
profile: ## Run performance profiling
	@echo "$(BLUE)Running performance profiling...$(RESET)"
	$(GOTEST) -cpuprofile=cpu.prof -memprofile=mem.prof -bench=. ./$(INTERNAL_DIR)/...
	@echo "$(GREEN)Performance profiling completed!$(RESET)"

profile-cpu: ## Show CPU profile
	@echo "$(BLUE)Showing CPU profile...$(RESET)"
	$(GOCMD) tool pprof cpu.prof

profile-mem: ## Show memory profile
	@echo "$(BLUE)Showing memory profile...$(RESET)"
	$(GOCMD) tool pprof mem.prof

## Release
release-build: ## Build release binaries
	@echo "$(BLUE)Building release binaries...$(RESET)"
	@mkdir -p dist
	# Linux
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o dist/$(APP_NAME)-server-linux-amd64 $(CMD_DIR)/server
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o dist/$(APP_NAME)-client-linux-amd64 $(CMD_DIR)/client
	# macOS
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o dist/$(APP_NAME)-server-darwin-amd64 $(CMD_DIR)/server
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o dist/$(APP_NAME)-client-darwin-amd64 $(CMD_DIR)/client
	# Windows
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o dist/$(APP_NAME)-server-windows-amd64.exe $(CMD_DIR)/server
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o dist/$(APP_NAME)-client-windows-amd64.exe $(CMD_DIR)/client
	@echo "$(GREEN)Release binaries built!$(RESET)"

release-checksum: ## Generate checksums for release
	@echo "$(BLUE)Generating checksums...$(RESET)"
	cd dist && sha256sum * > checksums.txt
	@echo "$(GREEN)Checksums generated!$(RESET)"

## Monitoring
monitor: ## Start monitoring
	@echo "$(BLUE)Starting monitoring...$(RESET)"
	@echo "$(YELLOW)Prometheus: http://localhost:9090$(RESET)"
	@echo "$(YELLOW)Grafana: http://localhost:3000$(RESET)"
	docker-compose -f docker-compose.monitoring.yml up -d

monitor-stop: ## Stop monitoring
	@echo "$(BLUE)Stopping monitoring...$(RESET)"
	docker-compose -f docker-compose.monitoring.yml down
	@echo "$(GREEN)Monitoring stopped!$(RESET)"

## Utilities
show-logs: ## Show application logs
	@echo "$(BLUE)Showing application logs...$(RESET)"
	tail -f logs/*.log

show-logs-server: ## Show server logs
	@echo "$(BLUE)Showing server logs...$(RESET)"
	tail -f logs/server.log

show-logs-client: ## Show client logs
	@echo "$(BLUE)Showing client logs...$(RESET)"
	tail -f logs/client.log

status: ## Show application status
	@echo "$(BLUE)Application Status:$(RESET)"
	@echo "$(YELLOW)Version: $(VERSION)$(RESET)"
	@echo "$(YELLOW)Git Commit: $(GIT_COMMIT)$(RESET)"
	@echo "$(YELLOW)Build Time: $(BUILD_TIME)$(RESET)"
	@echo ""
	@echo "$(WHITE)Docker Containers:$(RESET)"
	@docker-compose ps 2>/dev/null || echo "$(RED)No Docker containers running$(RESET)"
	@echo ""
	@echo "$(WHITE)Database:$(RESET)"
	@pg_isready -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) 2>/dev/null && echo "$(GREEN)Database is ready$(RESET)" || echo "$(RED)Database is not ready$(RESET)"

health: ## Check application health
	@echo "$(BLUE)Checking application health...$(RESET)"
	@curl -s http://localhost:8080/health | jq . 2>/dev/null || echo "$(RED)Health check failed$(RESET)"

## Database Management
db-backup: ## Backup database
	@echo "$(BLUE)Backing up database...$(RESET)"
	@mkdir -p backups
	pg_dump -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) $(DB_NAME) > backups/$(DB_NAME)_$(shell date +%Y%m%d_%H%M%S).sql
	@echo "$(GREEN)Database backup completed!$(RESET)"

db-restore: ## Restore database
	@echo "$(BLUE)Restoring database...$(RESET)"
	@echo "Available backups:"
	@ls -la backups/*.sql 2>/dev/null || echo "$(YELLOW)No backups found$(RESET)"
	@echo "Enter backup filename:" && read backup
	psql -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) $(DB_NAME) < backups/$$backup
	@echo "$(GREEN)Database restore completed!$(RESET)"

## Quick Commands
quick-start: dev-setup install-tools deps db-setup migrate-up build run ## Quick start for new developers
	@echo "$(GREEN)Quick start completed! Application should be running at http://localhost:8080$(RESET)"

quick-test: lint vet test-unit ## Quick test suite
	@echo "$(GREEN)Quick test suite completed!$(RESET)"

quick-build: deps build ## Quick build
	@echo "$(GREEN)Quick build completed!$(RESET)"

# Default targets that should always run
.PHONY: all
all: clean deps build test

# Ensure directories exist
bin:
	@mkdir -p bin

logs:
	@mkdir -p logs

config:
	@mkdir -p config

backups:
	@mkdir -p backups