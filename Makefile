# Argus - LGTM Stack Validator Makefile

# Version management
VERSION_FILE := internal/config/VERSION
VERSION := $(shell cat $(VERSION_FILE) 2>/dev/null || echo "0.0.1")
VERSION_TAG := v$(VERSION)

# Build variables
BINARY_NAME := argus
CMD_PATH := ./cmd/argus
BUILD_DIR := build
LDFLAGS := -X 'github.com/nahuelsantos/argus/internal/config.Version=$(VERSION_TAG)' \
           -X 'github.com/nahuelsantos/argus/internal/config.BuildTime=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)' \
           -X 'github.com/nahuelsantos/argus/internal/config.GitCommit=$(shell git rev-parse HEAD 2>/dev/null || echo "unknown")'

# Docker variables
IMAGE_NAME := ghcr.io/nahuelsantos/argus
IMAGE_TAG := $(VERSION_TAG)

.PHONY: help version build run test clean docker-build docker-push release

# Default target
.DEFAULT_GOAL := help

help: ## Show this help message
	@echo "Argus - LGTM Stack Validator"
	@echo ""
	@echo "Version: $(VERSION_TAG)"
	@echo ""
	@awk 'BEGIN {FS = ":.*##"; printf "Usage:\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

version: ## Show current version
	@echo "Current version: $(VERSION_TAG)"
	@echo "From file: $(VERSION_FILE)"

build: ## Build binary
	@echo "Building $(BINARY_NAME) $(VERSION_TAG)..."
	@mkdir -p $(BUILD_DIR)
	go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_PATH)
	@echo "Built: $(BUILD_DIR)/$(BINARY_NAME)"

run: build ## Build and run locally
	@echo "Starting $(BINARY_NAME) $(VERSION_TAG)..."
	./$(BUILD_DIR)/$(BINARY_NAME)

test: ## Run tests
	go test -v ./...

clean: ## Clean build artifacts
	rm -rf $(BUILD_DIR)
	docker image rm $(IMAGE_NAME):$(IMAGE_TAG) 2>/dev/null || true

docker-build: ## Build Docker image
	@echo "Building Docker image $(IMAGE_NAME):$(IMAGE_TAG)..."
	docker build -t $(IMAGE_NAME):$(IMAGE_TAG) .
	docker tag $(IMAGE_NAME):$(IMAGE_TAG) $(IMAGE_NAME):latest

docker-push: docker-build ## Build and push Docker image
	@echo "Pushing Docker image $(IMAGE_NAME):$(IMAGE_TAG)..."
	docker push $(IMAGE_NAME):$(IMAGE_TAG)
	docker push $(IMAGE_NAME):latest

release: ## Create a new release (usage: make release VERSION=1.0.0)
ifndef VERSION
	$(error VERSION is required. Usage: make release VERSION=1.0.0)
endif
	@echo "Creating release $(VERSION)..."
	@echo "$(VERSION)" > $(VERSION_FILE)
	@git add $(VERSION_FILE)
	@git commit -m "Bump version to v$(VERSION)" || true
	@git tag -a v$(VERSION) -m "Release v$(VERSION)" || true
	@echo "Release v$(VERSION) created. Push with: git push origin v$(VERSION)"

version-patch: ## Bump patch version (0.0.1 -> 0.0.2)
	@echo "Bumping patch version..."
	@$(call bump_version,patch)

version-minor: ## Bump minor version (0.1.0 -> 0.2.0)
	@echo "Bumping minor version..."
	@$(call bump_version,minor)

version-major: ## Bump major version (1.0.0 -> 2.0.0)
	@echo "Bumping major version..."
	@$(call bump_version,major)

# Helper function to bump version
define bump_version
	$(eval CURRENT := $(shell cat $(VERSION_FILE)))
	$(eval PARTS := $(subst ., ,$(CURRENT)))
	$(eval MAJOR := $(word 1,$(PARTS)))
	$(eval MINOR := $(word 2,$(PARTS)))
	$(eval PATCH := $(word 3,$(PARTS)))
	$(if $(filter patch,$1),\
		$(eval NEW_VERSION := $(MAJOR).$(MINOR).$(shell echo $$(($(PATCH) + 1)))),\
		$(if $(filter minor,$1),\
			$(eval NEW_VERSION := $(MAJOR).$(shell echo $$(($(MINOR) + 1))).0),\
			$(if $(filter major,$1),\
				$(eval NEW_VERSION := $(shell echo $$(($(MAJOR) + 1))).0.0),\
				$(error Invalid bump type: $1))))
	@echo "$(CURRENT) -> $(NEW_VERSION)"
	@echo "$(NEW_VERSION)" > $(VERSION_FILE)
	@echo "Version updated to $(NEW_VERSION)"
endef 