# Argus - LGTM Stack Validator Makefile

# Build variables
BINARY_NAME := argus
CMD_PATH := ./cmd/argus
BUILD_DIR := build

# Container variables
CONTAINER_NAME := argus
CONTAINER_IMAGE := argus:latest
TEST_CONTAINER_NAME := argus-test
TEST_CONTAINER_IMAGE := argus:test

.PHONY: help build run docker-build docker-run docker-test docker-stop test test-short lint clean

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

# Internal helper: stop and remove a container
define stop_container
	@echo "Stopping $(1) container..."
	@docker stop $(1) 2>/dev/null || true
	@docker rm $(1) 2>/dev/null || true
endef

test: ## Run tests
	go test -v ./...

test-short: ## Run short tests (for CI)
	go test -short -race -parallel 8 ./...

lint: ## Run linter
	golangci-lint run

clean: ## Clean build artifacts
	rm -rf $(BUILD_DIR)
	@echo "Cleaned build artifacts"

docker-build: ## Build Docker image with ARGUS_ env support
	@echo "Building Argus Docker image..."
	@docker build -t $(TEST_CONTAINER_IMAGE) \
		--build-arg VERSION=v0.0.1-test \
		--build-arg BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S') \
		--build-arg GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown") \
		.
	@echo "✅ Docker image built: $(TEST_CONTAINER_IMAGE)"

docker-run: docker-build ## Build and run Argus in Docker with ARGUS_ env vars
	$(call stop_container,$(TEST_CONTAINER_NAME))
	@echo "Starting Argus container with ARGUS_ environment variables..."
	@docker run -d \
		--name $(TEST_CONTAINER_NAME) \
		-p 3001:3001 \
		--env-file .env \
		$(TEST_CONTAINER_IMAGE)
	@echo "✅ Argus container started!"
	@echo "Dashboard: http://localhost:3001 (or http://your-server-ip:3001)"
	@echo "API docs: http://localhost:3001/api"
	@echo "Health: http://localhost:3001/health"
	@echo ""
	@echo "📋 To check logs: docker logs -f $(TEST_CONTAINER_NAME)"
	@echo "🛑 To stop: make docker-stop"

docker-test: docker-run ## Build, run, and test the container
	@echo "🧪 Testing Argus container..."
	@sleep 3
	@echo "Testing health endpoint..."
	@curl -s http://localhost:3001/health | grep -q '"status":"healthy"' && echo "✅ Health check passed" || echo "❌ Health check failed"
	@echo "Testing settings endpoint (ARGUS_ env vars)..."
	@curl -s http://localhost:3001/api/settings | grep -q '"grafana"' && echo "✅ Settings endpoint passed" || echo "❌ Settings endpoint failed"
	@echo "Testing dashboard..."
	@curl -s http://localhost:3001/ | grep -q 'Argus' && echo "✅ Dashboard accessible" || echo "❌ Dashboard failed"
	@echo ""
	@echo "🔍 Container logs (last 10 lines):"
	@docker logs --tail 10 $(TEST_CONTAINER_NAME)

docker-stop: ## Stop and remove Argus test container
	$(call stop_container,$(TEST_CONTAINER_NAME))
	@echo "✅ Argus test container stopped" 