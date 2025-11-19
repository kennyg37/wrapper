# Makefile for Mock Data Generator Backend

# Variables
BINARY_NAME=mockdata-api
GO=go
GOFLAGS=-v
MAIN_PATH=./cmd/api

# Colors for output
COLOR_RESET=\033[0m
COLOR_BOLD=\033[1m
COLOR_GREEN=\033[32m
COLOR_YELLOW=\033[33m
COLOR_BLUE=\033[34m

.PHONY: help
help: ## Show this help message
	@echo "$(COLOR_BOLD)Available commands:$(COLOR_RESET)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(COLOR_BLUE)%-15s$(COLOR_RESET) %s\n", $$1, $$2}'

.PHONY: install
install: ## Install dependencies
	@echo "$(COLOR_GREEN)Installing dependencies...$(COLOR_RESET)"
	$(GO) mod download
	$(GO) mod verify

.PHONY: run
run: ## Run the application
	@echo "$(COLOR_GREEN)Starting server...$(COLOR_RESET)"
	$(GO) run $(GOFLAGS) $(MAIN_PATH)/main.go

.PHONY: build
build: ## Build the application
	@echo "$(COLOR_GREEN)Building $(BINARY_NAME)...$(COLOR_RESET)"
	$(GO) build $(GOFLAGS) -o bin/$(BINARY_NAME) $(MAIN_PATH)/main.go
	@echo "$(COLOR_GREEN)✓ Binary created at bin/$(BINARY_NAME)$(COLOR_RESET)"

.PHONY: test
test: ## Run tests
	@echo "$(COLOR_GREEN)Running tests...$(COLOR_RESET)"
	$(GO) test $(GOFLAGS) ./...

.PHONY: test-coverage
test-coverage: ## Run tests with coverage
	@echo "$(COLOR_GREEN)Running tests with coverage...$(COLOR_RESET)"
	$(GO) test -cover ./...
	@echo ""
	@echo "$(COLOR_YELLOW)For detailed coverage report, run: make coverage-html$(COLOR_RESET)"

.PHONY: coverage-html
coverage-html: ## Generate HTML coverage report
	@echo "$(COLOR_GREEN)Generating coverage report...$(COLOR_RESET)"
	$(GO) test -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "$(COLOR_GREEN)✓ Coverage report generated: coverage.html$(COLOR_RESET)"

.PHONY: lint
lint: ## Run linter (requires golangci-lint)
	@echo "$(COLOR_GREEN)Running linter...$(COLOR_RESET)"
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run ./...; \
	else \
		echo "$(COLOR_YELLOW)golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest$(COLOR_RESET)"; \
	fi

.PHONY: fmt
fmt: ## Format code
	@echo "$(COLOR_GREEN)Formatting code...$(COLOR_RESET)"
	$(GO) fmt ./...
	@echo "$(COLOR_GREEN)✓ Code formatted$(COLOR_RESET)"

.PHONY: vet
vet: ## Run go vet
	@echo "$(COLOR_GREEN)Running go vet...$(COLOR_RESET)"
	$(GO) vet ./...

.PHONY: clean
clean: ## Clean build artifacts
	@echo "$(COLOR_GREEN)Cleaning...$(COLOR_RESET)"
	rm -rf bin/
	rm -f coverage.out coverage.html
	@echo "$(COLOR_GREEN)✓ Cleaned$(COLOR_RESET)"

.PHONY: dev
dev: ## Run with auto-reload (requires air)
	@echo "$(COLOR_GREEN)Starting development server with auto-reload...$(COLOR_RESET)"
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "$(COLOR_YELLOW)air not installed. Install with: go install github.com/cosmtrek/air@latest$(COLOR_RESET)"; \
		echo "$(COLOR_YELLOW)Falling back to normal run...$(COLOR_RESET)"; \
		make run; \
	fi

.PHONY: db-create
db-create: ## Create database
	@echo "$(COLOR_GREEN)Creating database...$(COLOR_RESET)"
	createdb mockdata_generator || echo "$(COLOR_YELLOW)Database may already exist$(COLOR_RESET)"

.PHONY: db-drop
db-drop: ## Drop database (WARNING: destructive)
	@echo "$(COLOR_YELLOW)⚠️  Dropping database...$(COLOR_RESET)"
	dropdb mockdata_generator || echo "$(COLOR_YELLOW)Database may not exist$(COLOR_RESET)"

.PHONY: tidy
tidy: ## Tidy dependencies
	@echo "$(COLOR_GREEN)Tidying dependencies...$(COLOR_RESET)"
	$(GO) mod tidy

.PHONY: check
check: fmt vet test ## Run all checks (format, vet, test)
	@echo "$(COLOR_GREEN)✓ All checks passed$(COLOR_RESET)"

.PHONY: docker-build
docker-build: ## Build Docker image
	@echo "$(COLOR_GREEN)Building Docker image...$(COLOR_RESET)"
	docker build -t $(BINARY_NAME):latest .

.PHONY: docker-run
docker-run: ## Run Docker container
	@echo "$(COLOR_GREEN)Running Docker container...$(COLOR_RESET)"
	docker run -p 3000:3000 --env-file .env $(BINARY_NAME):latest

# Default target
.DEFAULT_GOAL := help
