# Argus - LGTM Stack Validator Makefile

# Build variables
BINARY_NAME := argus
CMD_PATH := ./cmd/argus
BUILD_DIR := build

.PHONY: help build run start stop test lint clean

# Default target
.DEFAULT_GOAL := help

help: ## Show this help message
	@echo "Argus - LGTM Stack Validator"
	@echo ""
	@awk 'BEGIN {FS = ":.*##"; printf "Usage:\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-10s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

build: ## Build binary
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_PATH)
	@echo "Built: $(BUILD_DIR)/$(BINARY_NAME)"

run: build ## Build and run locally
	@echo "Starting $(BINARY_NAME)..."
	./$(BUILD_DIR)/$(BINARY_NAME)

start: ## Start with Docker Compose
	docker compose up -d
	@echo "Started Argus stack with Docker Compose"

stop: ## Stop Docker Compose
	docker compose down
	@echo "Stopped Argus stack"

test: ## Run tests
	go test -v ./...

lint: ## Run linter
	golangci-lint run

clean: ## Clean build artifacts
	rm -rf $(BUILD_DIR)
	@echo "Cleaned build artifacts" 