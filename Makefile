# ===========================================
# RCNbuild Makefile
# ===========================================
# Run `make help` to see all available commands

# Load .env file if it exists
ifneq (,$(wildcard ./.env))
    include .env
    export
endif

# Default Go binary output directory
BIN_DIR := ./bin

# Colors for pretty output
GREEN  := \033[0;32m
YELLOW := \033[0;33m
CYAN   := \033[0;36m
RESET  := \033[0m

# Docker configuration (use standard socket if Docker Desktop socket doesn't exist)
DOCKER_COMPOSE := DOCKER_HOST=unix:///var/run/docker.sock docker compose

.PHONY: help dev down logs ps ngrok-url api worker migrate-up migrate-down migrate-create migrate-status test lint build clean deps

# ===========================================
# Help
# ===========================================
help:
	@echo ""
	@echo "$(CYAN)RCNbuild Development Commands$(RESET)"
	@echo ""
	@echo "$(YELLOW)Infrastructure:$(RESET)"
	@echo "  $(GREEN)make dev$(RESET)              - Start all infrastructure services (PostgreSQL, Redis, Traefik, Registry, ngrok)"
	@echo "  $(GREEN)make down$(RESET)             - Stop all infrastructure services"
	@echo "  $(GREEN)make logs$(RESET)             - Tail logs from all services"
	@echo "  $(GREEN)make ps$(RESET)               - Show running containers"
	@echo "  $(GREEN)make ngrok-url$(RESET)        - Get the ngrok HTTPS URL for GitHub App callback"
	@echo ""
	@echo "$(YELLOW)Application:$(RESET)"
	@echo "  $(GREEN)make api$(RESET)              - Run the API server (with hot reload if air is installed)"
	@echo "  $(GREEN)make worker$(RESET)           - Run the build worker (with hot reload if air is installed)"
	@echo "  $(GREEN)make build$(RESET)            - Build all binaries to ./bin/"
	@echo ""
	@echo "$(YELLOW)Database:$(RESET)"
	@echo "  $(GREEN)make migrate-up$(RESET)       - Apply all pending migrations"
	@echo "  $(GREEN)make migrate-down$(RESET)     - Rollback the last migration"
	@echo "  $(GREEN)make migrate-create$(RESET)   - Create new migration (usage: make migrate-create name=create_users)"
	@echo "  $(GREEN)make migrate-status$(RESET)   - Show migration status"
	@echo "  $(GREEN)make db-shell$(RESET)         - Open psql shell to database"
	@echo ""
	@echo "$(YELLOW)Development:$(RESET)"
	@echo "  $(GREEN)make deps$(RESET)             - Download Go dependencies"
	@echo "  $(GREEN)make test$(RESET)             - Run all tests"
	@echo "  $(GREEN)make lint$(RESET)             - Run linter (golangci-lint)"
	@echo "  $(GREEN)make clean$(RESET)            - Remove build artifacts"
	@echo ""
	@echo "$(YELLOW)Ports:$(RESET)"
	@echo "  PostgreSQL:     localhost:5437"
	@echo "  Redis:          localhost:6379"
	@echo "  Traefik UI:     http://localhost:8080"
	@echo "  Registry:       localhost:5000"
	@echo "  ngrok UI:       http://localhost:4040"
	@echo "  API Server:     http://localhost:$(API_PORT)"
	@echo "  Dashboard:      http://localhost:3000"
	@echo ""

# ===========================================
# Infrastructure Commands
# ===========================================

# Start all infrastructure services
dev:
	@echo "$(CYAN)Starting infrastructure...$(RESET)"
	@$(DOCKER_COMPOSE) up -d
	@echo ""
	@echo "$(GREEN)✓ Infrastructure started successfully!$(RESET)"
	@echo ""
	@echo "  PostgreSQL:     localhost:5437"
	@echo "  Redis:          localhost:6379"
	@echo "  Traefik UI:     http://localhost:8080"
	@echo "  Registry:       localhost:5000"
	@echo "  ngrok UI:       http://localhost:4040"
	@echo ""
	@echo "Run $(YELLOW)make ngrok-url$(RESET) to get your GitHub App callback URL"
	@echo "Run $(YELLOW)make logs$(RESET) to see container logs"
	@echo "Run $(YELLOW)make ps$(RESET) to see container status"

# Stop all infrastructure services
down:
	@echo "$(CYAN)Stopping infrastructure...$(RESET)"
	@$(DOCKER_COMPOSE) down
	@echo "$(GREEN)✓ Infrastructure stopped$(RESET)"

# Stop and remove all data (volumes)
down-clean:
	@echo "$(YELLOW)WARNING: This will delete all data!$(RESET)"
	@read -p "Are you sure? [y/N] " confirm && [ "$$confirm" = "y" ] || exit 1
	@$(DOCKER_COMPOSE) down -v
	@echo "$(GREEN)✓ Infrastructure stopped and data removed$(RESET)"

# Tail logs from all services
logs:
	@$(DOCKER_COMPOSE) logs -f

# Show running containers
ps:
	@$(DOCKER_COMPOSE) ps

# Get ngrok HTTPS URL for GitHub App callback
ngrok-url:
	@echo "$(CYAN)Fetching ngrok tunnel URL...$(RESET)"
	@curl -s http://localhost:4040/api/tunnels | grep -o '"public_url":"https://[^"]*' | cut -d'"' -f4 | head -1 | while read url; do \
		if [ -n "$$url" ]; then \
			echo ""; \
			echo "$(GREEN)ngrok HTTPS URL:$(RESET) $$url"; \
			echo ""; \
			echo "$(YELLOW)GitHub App Callback URL:$(RESET)"; \
			echo "  $$url/api/auth/github/callback"; \
			echo ""; \
			echo "$(YELLOW)GitHub Webhook URL:$(RESET)"; \
			echo "  $$url/api/webhooks/github"; \
			echo ""; \
		else \
			echo "$(YELLOW)ngrok is not running or no tunnel found$(RESET)"; \
			echo "Run $(GREEN)make dev$(RESET) first to start infrastructure"; \
		fi \
	done

# ===========================================
# Application Commands
# ===========================================

# Run API server with hot reload (requires air: go install github.com/air-verse/air@latest)
api:
	@if command -v air > /dev/null 2>&1; then \
		echo "$(CYAN)Starting API server with hot reload (air)...$(RESET)"; \
		air -c .air.api.toml 2>/dev/null || air; \
	else \
		echo "$(YELLOW)air not found, running without hot reload$(RESET)"; \
		echo "$(YELLOW)Install with: go install github.com/air-verse/air@latest$(RESET)"; \
		go run ./cmd/api; \
	fi

# Run worker with hot reload
worker:
	@if command -v air > /dev/null 2>&1; then \
		echo "$(CYAN)Starting worker with hot reload (air)...$(RESET)"; \
		air -c .air.worker.toml 2>/dev/null || air; \
	else \
		echo "$(YELLOW)air not found, running without hot reload$(RESET)"; \
		go run ./cmd/worker; \
	fi

# Build all binaries
build:
	@echo "$(CYAN)Building binaries...$(RESET)"
	@mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 go build -ldflags="-s -w" -o $(BIN_DIR)/api ./cmd/api
	CGO_ENABLED=0 go build -ldflags="-s -w" -o $(BIN_DIR)/worker ./cmd/worker
	@echo "$(GREEN)✓ Binaries built to $(BIN_DIR)/$(RESET)"

# ===========================================
# Database Commands
# ===========================================

# Apply all pending migrations
migrate-up:
	@echo "$(CYAN)Applying migrations...$(RESET)"
	@migrate -path ./migrations -database "$(DATABASE_URL)" up
	@echo "$(GREEN)✓ Migrations applied$(RESET)"

# Rollback the last migration
migrate-down:
	@echo "$(YELLOW)Rolling back last migration...$(RESET)"
	@migrate -path ./migrations -database "$(DATABASE_URL)" down 1
	@echo "$(GREEN)✓ Migration rolled back$(RESET)"

# Create a new migration file
migrate-create:
ifndef name
	@echo "$(YELLOW)Usage: make migrate-create name=create_users_table$(RESET)"
	@exit 1
endif
	@echo "$(CYAN)Creating migration: $(name)$(RESET)"
	@migrate create -ext sql -dir ./migrations -seq $(name)
	@echo "$(GREEN)✓ Migration files created in ./migrations/$(RESET)"

# Show migration status
migrate-status:
	@migrate -path ./migrations -database "$(DATABASE_URL)" version

# Open psql shell
db-shell:
	@docker exec -it rcnbuild-postgres psql -U $(POSTGRES_USER) -d $(POSTGRES_DB)

# ===========================================
# Development Commands
# ===========================================

# Download dependencies
deps:
	@echo "$(CYAN)Downloading dependencies...$(RESET)"
	@go mod download
	@go mod tidy
	@echo "$(GREEN)✓ Dependencies downloaded$(RESET)"

# Run tests
test:
	@echo "$(CYAN)Running tests...$(RESET)"
	@go test -v -race -cover ./...

# Run tests with coverage report
test-coverage:
	@echo "$(CYAN)Running tests with coverage...$(RESET)"
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)✓ Coverage report: coverage.html$(RESET)"

# Run linter (requires golangci-lint: https://golangci-lint.run/usage/install/)
lint:
	@if command -v golangci-lint > /dev/null 2>&1; then \
		echo "$(CYAN)Running linter...$(RESET)"; \
		golangci-lint run ./...; \
	else \
		echo "$(YELLOW)golangci-lint not found$(RESET)"; \
		echo "Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
		exit 1; \
	fi

# Clean build artifacts
clean:
	@echo "$(CYAN)Cleaning...$(RESET)"
	@rm -rf $(BIN_DIR)
	@rm -f coverage.out coverage.html
	@go clean -cache
	@echo "$(GREEN)✓ Cleaned$(RESET)"

# ===========================================
# Convenience Targets
# ===========================================

# Start everything for development
start: dev
	@sleep 2
	@make api

# Full reset: stop, clean, rebuild
reset: down-clean dev migrate-up
	@echo "$(GREEN)✓ Environment reset complete$(RESET)"
