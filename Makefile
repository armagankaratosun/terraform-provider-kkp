.PHONY: help build test lint lint-fix clean install deps dev-deps fmt fmt-go-s fmt-examples tidy check pre-commit dev release tag tag-push tag-delete tag-repush changelog changelog-release

# Variables
BINARY_NAME=terraform-provider-kkp
VERSION?=v0.0.1
PLAIN_VERSION:=$(patsubst v%,%,$(VERSION))
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
		-o $(BUILD_DIR)/$(BINARY_NAME)_$(VERSION) .
	@echo "Binary built: $(BUILD_DIR)/$(BINARY_NAME)_$(VERSION)"

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
	@mkdir -p ~/.terraform.d/plugins/registry.terraform.io/armagankaratosun/kkp/$(PLAIN_VERSION)/$(GOOS)_$(GOARCH)/
	@cp $(BUILD_DIR)/$(BINARY_NAME)_$(VERSION) ~/.terraform.d/plugins/registry.terraform.io/armagankaratosun/kkp/$(PLAIN_VERSION)/$(GOOS)_$(GOARCH)/$(BINARY_NAME)_$(VERSION)
	@echo "Provider installed to ~/.terraform.d/plugins/"

deps: dev-deps ## Alias for dev-deps

dev-deps: ## Install development dependencies (golangci-lint, goimports, git-cliff)
	@echo "Installing development dependencies..."
	@# golangci-lint
	@export PATH=$$PATH:$$(go env GOPATH)/bin; which golangci-lint >/dev/null 2>&1 || \
		(echo "Installing latest golangci-lint..." && \
		 curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin latest)
	@# goimports
	@export PATH=$$PATH:$$(go env GOPATH)/bin; which goimports >/dev/null 2>&1 || \
		(echo "Installing goimports..." && \
		 GOBIN=$$(go env GOPATH)/bin go install golang.org/x/tools/cmd/goimports@latest)
	@# git-cliff (prebuilt binary)
	@which git-cliff >/dev/null 2>&1 || \
	  ( \
		echo "Installing git-cliff v2.2.1..."; \
		OS=$$(uname -s); ARCH=$$(uname -m); \
		case "$$OS/$$ARCH" in \
		  Linux/x86_64)  ASSET=git-cliff-2.2.1-x86_64-unknown-linux-gnu.tar.gz ;; \
		  Linux/aarch64|Linux/arm64) ASSET=git-cliff-2.2.1-aarch64-unknown-linux-gnu.tar.gz ;; \
		  Darwin/x86_64) ASSET=git-cliff-2.2.1-x86_64-apple-darwin.tar.gz ;; \
		  Darwin/arm64)  ASSET=git-cliff-2.2.1-aarch64-apple-darwin.tar.gz ;; \
		  *) echo "Unsupported OS/ARCH for git-cliff prebuilt: $$OS/$$ARCH"; ASSET="" ;; \
		esac; \
		if [ -n "$$ASSET" ]; then \
		  TMP=$$(mktemp -d); \
		  URL="https://github.com/orhun/git-cliff/releases/download/v2.2.1/$$ASSET"; \
		  echo "Downloading $$URL"; \
		  curl -sSL -o "$$TMP/git-cliff.tar.gz" "$$URL" && \
		  tar -xzf "$$TMP/git-cliff.tar.gz" -C "$$TMP"; \
		  BIN=$$(find "$$TMP" -maxdepth 3 \( -type f -name git-cliff -o -name git-cliff.exe \) | head -n1); \
		  if [ -z "$$BIN" ]; then echo "git-cliff binary not found in archive"; exit 1; fi; \
		  mkdir -p "$$HOME/.local/bin" && \
		  install -m 0755 "$$BIN" "$$HOME/.local/bin/git-cliff" && \
		  echo "Installed git-cliff to $$HOME/.local/bin. Ensure it is on your PATH."; \
		fi \
	  )
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

# Tagging helpers (use v-prefixed VERSION, e.g., VERSION=v0.1.0)
tag: ## Create an annotated git tag $(VERSION) on HEAD
	@git tag -a $(VERSION) -m "Release $(VERSION)"
	@echo "Created tag $(VERSION)"

tag-push: ## Push tag $(VERSION) to origin
	@git push origin $(VERSION)

tag-delete: ## Delete local and remote tag $(VERSION)
	@echo "Deleting local tag $(VERSION)"
	@-git tag -d $(VERSION)
	@echo "Deleting remote tag $(VERSION)"
	@-git push --delete origin $(VERSION) || git push origin :refs/tags/$(VERSION)

tag-repush: tag-delete tag ## Delete and recreate tag $(VERSION), then push
	@$(MAKE) tag-push

# Changelog helpers (requires git-cliff installed locally)
changelog: ## Generate Unreleased changelog into CHANGELOG.md using git-cliff
	@git-cliff --unreleased --output CHANGELOG.md

changelog-release: ## Generate CHANGELOG for a release, e.g., make changelog-release VERSION=v0.1.0
	@test -n "$(VERSION)" || (echo "VERSION is required, e.g., make changelog-release VERSION=v0.1.0" && exit 1)
	@git-cliff --tag $(VERSION) --output CHANGELOG.md

.DEFAULT_GOAL := help
