.PHONY: help build test lint lint-fix clean install dev-deps fmt fmt-go-s fmt-examples tidy check pre-commit dev release tag tag-push tag-delete tag-repush changelog changelog-release

# Variables
BINARY_NAME=terraform-provider-kkp
VERSION?=0.0.1
BUILD_DIR=./bin
GOOS?=$(shell go env GOOS)
GOARCH?=$(shell go env GOARCH)

help: ## Display this help message
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the Terraform provider binary
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build \
		-ldflags="-X main.version=$(VERSION)" \
		-o $(BUILD_DIR)/$(BINARY_NAME)_v$(VERSION) .
	@echo "Binary built: $(BUILD_DIR)/$(BINARY_NAME)_v$(VERSION)"

test: ## Run all tests
	@echo "Running tests..."
	@go test -v ./...

test-coverage: ## Run tests with coverage report
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

lint: ## Run linting with golangci-lint
	@echo "Running golangci-lint..."
	@export PATH=$$PATH:$$(go env GOPATH)/bin && golangci-lint run

lint-fix: ## Run linting with auto-fix where possible
	@echo "Running golangci-lint with fixes..."
	@export PATH=$$PATH:$$(go env GOPATH)/bin && golangci-lint run --fix

lint-verbose: ## Run linting with verbose output
	@echo "Running golangci-lint (verbose)..."
	@export PATH=$$PATH:$$(go env GOPATH)/bin && golangci-lint run --verbose

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html

install: build ## Install the provider binary to local Terraform plugin directory
	@echo "Installing provider locally..."
	@mkdir -p ~/.terraform.d/plugins/registry.terraform.io/armagankaratosun/kkp/$(VERSION)/$(GOOS)_$(GOARCH)/
	@cp $(BUILD_DIR)/$(BINARY_NAME)_v$(VERSION) ~/.terraform.d/plugins/registry.terraform.io/armagankaratosun/kkp/$(VERSION)/$(GOOS)_$(GOARCH)/$(BINARY_NAME)_v$(VERSION)
	@echo "Provider installed to ~/.terraform.d/plugins/"

dev-deps: ## Install development dependencies
	@echo "Installing development dependencies..."
	@export PATH=$$PATH:$$(go env GOPATH)/bin && which golangci-lint > /dev/null || (echo "Installing latest golangci-lint..." && curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin latest)
	@echo "Development dependencies installed"

fmt: ## Format Go code
	@echo "Formatting code..."
	@go fmt ./...
	@goimports -w -local github.com/armagankaratosun/terraform-provider-kkp .

fmt-go-s: ## Run gofmt -s (simplify) across repo
	@echo "Running gofmt -s ..."
	@gofmt -s -w .

fmt-examples: ## Run terraform fmt recursively on examples/
	@echo "Formatting Terraform examples..."
	@terraform fmt -recursive examples || true

tidy: ## Tidy up go.mod
	@echo "Tidying go.mod..."
	@go mod tidy

check: lint test ## Run both linting and tests

pre-commit: fmt tidy lint test ## Run pre-commit checks (format, tidy, lint, test)
	@echo "Pre-commit checks completed successfully!"

# Development workflow targets
dev: clean fmt tidy build ## Full development build (clean, format, tidy, build)

release: clean test lint build ## Release build (clean, test, lint, build)

# Tagging helpers
tag: ## Create an annotated git tag v$(VERSION) on HEAD
	@git tag -a v$(VERSION) -m "Release v$(VERSION)"
	@echo "Created tag v$(VERSION)"

tag-push: ## Push tag v$(VERSION) to origin
	@git push origin v$(VERSION)

tag-delete: ## Delete local and remote tag v$(VERSION)
	@echo "Deleting local tag v$(VERSION)"
	@-git tag -d v$(VERSION)
	@echo "Deleting remote tag v$(VERSION)"
	@-git push --delete origin v$(VERSION) || git push origin :refs/tags/v$(VERSION)

tag-repush: tag-delete tag ## Delete and recreate tag v$(VERSION), then push
	@$(MAKE) tag-push

# Changelog helpers (requires git-cliff installed locally)
changelog: ## Generate Unreleased changelog into CHANGELOG.md using git-cliff
	@git-cliff --unreleased --output CHANGELOG.md

changelog-release: ## Generate CHANGELOG for a release, e.g., make changelog-release VERSION=0.1.0
	@test -n "$(VERSION)" || (echo "VERSION is required, e.g., make changelog-release VERSION=0.1.0" && exit 1)
	@git-cliff --tag $(VERSION) --output CHANGELOG.md

.DEFAULT_GOAL := help
