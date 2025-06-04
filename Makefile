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

start: ## Start Argus container
	@echo "Starting Argus container..."
	@docker stop argus 2>/dev/null || true
	@docker rm argus 2>/dev/null || true
	@docker run -d -p 3001:3001 --name argus --network traefik_network test-argus-v2
	@echo "✅ Argus started at http://localhost:3001"

stop: ## Stop Argus container
	@echo "Stopping Argus container..."
	@docker stop argus 2>/dev/null || true
	@docker rm argus 2>/dev/null || true
	@echo "✅ Argus stopped"

test: ## Run tests
	go test -v ./...

lint: ## Run linter
	golangci-lint run

clean: ## Clean build artifacts
	rm -rf $(BUILD_DIR)
	@echo "Cleaned build artifacts" 