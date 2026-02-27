# InkForge Makefile
# Defines build, test, and release processes

# Variables
BINARY_NAME=inkforge
DOCKER_REGISTRY=registry.yygu.cn/skills
GIT_COMMIT ?= $(shell git rev-parse --short HEAD)
IMAGE_TAG ?= $(GIT_COMMIT)
BUILD_DIR ?= build

# Go variables
GO_CMD ?= go
GO_BUILD ?= $(GO_CMD) build
GO_TEST ?= $(GO_CMD) test
GO_MOD ?= $(GO_CMD) mod
GO_CLEAN ?= $(GO_CMD) clean
GO_FMT ?= $(GO_CMD) fmt

# Default target
.DEFAULT_GOAL := help

.PHONY: help build build-linux test clean fmt cover docker-build-base docker-push-base docker-build-release docker-push-release release deps fast-release

# Display help
help: ## Show this help message
	@echo "InkForge Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make <target>"
	@echo ""
	@echo "Targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Install dependencies
deps: ## Download dependencies
	$(GO_MOD) tidy
	$(GO_MOD) download

# Build for current platform
build: ## Build binary for current platform
	@mkdir -p $(BUILD_DIR)
	GOOS=$$(go env GOOS) GOARCH=$$(go env GOARCH) $(GO_BUILD) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/inkforge

# Build for Linux (production target)
build-linux: ## Build binary for Linux (for Docker)
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 $(GO_BUILD) -a -ldflags="-w -s -extldflags '-static'" -o $(BUILD_DIR)/$(BINARY_NAME)-linux ./cmd/inkforge

# Run tests
test: ## Run tests
	$(GO_TEST) -race -v ./...

# Run coverage
cover: ## Run tests with coverage
	$(GO_TEST) -race -coverprofile=coverage.out -covermode=atomic ./...
	$(GO_CMD) tool cover -html=coverage.out -o coverage.html

# Format code
fmt: ## Format code with gofmt
	$(GO_FMT) ./...

# Clean build artifacts
clean: ## Remove build artifacts
	rm -rf $(BUILD_DIR)

# Build base Docker image
docker-build-base: ## Build the base Docker image
	docker build -f deployments/base.Dockerfile -t $(DOCKER_REGISTRY)/$(BINARY_NAME):base .

# Push base Docker image to registry
docker-push-base: ## Push the base Docker image to registry
	docker push $(DOCKER_REGISTRY)/$(BINARY_NAME):base

# Build release Docker image (requires base image to be built first)
docker-build-release: ## Build the release Docker image
	docker build -f deployments/Dockerfile -t $(DOCKER_REGISTRY)/$(BINARY_NAME):$(IMAGE_TAG) .

# Alternative release that only rebuilds and pushes the release image (assumes base is already built)
fast-release: docker-build-release docker-push-release ## Release using pre-built base image
	@echo "Fast release complete. Release image pushed to $(DOCKER_REGISTRY)"