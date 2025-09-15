.PHONY: help build test lint lint-fix clean install deps dev-deps fmt fmt-check tidy check pre-commit dev release tag tag-push tag-delete tag-repush docs docs-check snapshot examples-validate examples-bump-version examples-version-check

# Variables
BINARY_NAME=terraform-provider-kkp
BUILD_DIR=./bin
GOOS?=$(shell go env GOOS)
GOARCH?=$(shell go env GOARCH)

help: ## Display this help message
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the Terraform provider binary
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@V=$${VERSION:-dev}; \
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build \
		-ldflags="-X main.version=$$V" \
		-o $(BUILD_DIR)/$(BINARY_NAME)_$$V .; \
	echo "Binary built: $(BUILD_DIR)/$(BINARY_NAME)_$$V"

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
	@if [ -z "$(VERSION)" ]; then \
	  echo "ERROR: VERSION is required for 'make install' (e.g., VERSION=v0.1.0)" >&2; exit 1; \
	fi
	@echo "Installing provider locally..."
	@VER_PLAIN=$${VERSION#v}; \
	mkdir -p $$HOME/.terraform.d/plugins/registry.terraform.io/armagankaratosun/kkp/$$VER_PLAIN/$(GOOS)_$(GOARCH)/; \
	cp $(BUILD_DIR)/$(BINARY_NAME)_$(VERSION) $$HOME/.terraform.d/plugins/registry.terraform.io/armagankaratosun/kkp/$$VER_PLAIN/$(GOOS)_$(GOARCH)/$(BINARY_NAME)_$(VERSION); \
	echo "Provider installed to $$HOME/.terraform.d/plugins/"

deps: dev-deps ## Alias for dev-deps

dev-deps: ## Install development dependencies (golangci-lint, goimports)
	@echo "Installing development dependencies..."
	@# golangci-lint
	@export PATH=$$PATH:$$(go env GOPATH)/bin; which golangci-lint >/dev/null 2>&1 || \
		(echo "Installing latest golangci-lint..." && \
		 curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin latest)
	@# goimports
	@export PATH=$$PATH:$$(go env GOPATH)/bin; which goimports >/dev/null 2>&1 || \
		(echo "Installing goimports..." && \
		 GOBIN=$$(go env GOPATH)/bin go install golang.org/x/tools/cmd/goimports@latest)
	@echo "Development dependencies installed"

fmt: ## Format Go and Terraform (writes changes)
	@echo "Formatting Go code..."
	@go fmt ./...
	@goimports -w -local github.com/armagankaratosun/terraform-provider-kkp .
	@if [ -d examples ]; then \
		echo "Formatting Terraform in examples/..."; \
		if command -v terraform >/dev/null 2>&1; then \
		  terraform fmt -recursive examples || true; \
		else \
		  echo "Terraform not found; skipping Terraform formatting."; \
		fi; \
	      fi

fmt-check: ## Check formatting (gofmt -s, terraform fmt -check on examples/)
	@echo "Checking Go formatting (gofmt -s)..."
	@unformatted=$$(gofmt -s -l .); \
	  if [ -n "$$unformatted" ]; then \
	    echo "Files not gofmt'ed:" >&2; \
	    echo "$$unformatted" >&2; \
	    exit 1; \
	  fi
	@if [ -d examples ]; then \
		echo "Checking Terraform formatting in examples/..."; \
		if command -v terraform >/dev/null 2>&1; then \
		  terraform fmt -check -recursive examples; \
		else \
		  echo "Terraform not found; please install Terraform CLI"; exit 1; \
		fi; \
	      fi

tidy: ## Tidy up go.mod
	@echo "Tidying go.mod..."
	@go mod tidy

# Examples
examples-validate: ## Validate all examples with terraform init/validate (no backend)
	@echo "Validating Terraform examples..."
	@command -v terraform >/dev/null 2>&1 || { echo "Terraform CLI is required for example validation" >&2; exit 1; }
	@set -e; \
	find examples -type f -name 'main.tf' | while read -r f; do \
	  d=$$(dirname "$$f"); \
	  echo "==> $$d"; \
	  (cd "$$d" && terraform init -backend=false -input=false -upgrade >/dev/null && terraform validate -no-color); \
	done; \
	echo "All examples validated."

# Usage: make examples-bump-version EXAMPLES_VERSION=0.1.0
examples-bump-version: ## Bump provider version constraint in examples (EXAMPLES_VERSION=0.1.0)
	@if [ -z "$(EXAMPLES_VERSION)" ]; then echo "ERROR: EXAMPLES_VERSION is required (e.g., EXAMPLES_VERSION=0.1.0)" >&2; exit 1; fi
	@echo "Bumping examples provider version to ~> $(EXAMPLES_VERSION) ..."
	@find examples -type f -name 'main.tf' -print0 | xargs -0 sed -i.bak -E 's/(version\s*=\s*")(~>\s*)?[^\"]+(\".*)/\1~> $(EXAMPLES_VERSION)\3/'; \
	find examples -type f -name '*.bak' -delete; \
	echo "Done. Run 'make examples-version-check' to verify."

examples-version-check: ## Print current provider version constraints used in examples
	@rg -n "required_providers|source\s*=\s*\"armagankaratosun/kkp\"|version\s*=\s*\"~>" examples -S || true

check: lint test ## Run both linting and tests

pre-commit: fmt tidy lint test ## Run pre-commit checks (format, tidy, lint, test)
	@echo "Pre-commit checks completed successfully!"

# Development workflow targets
dev: clean fmt tidy build ## Full development build (clean, format, tidy, build)

release: clean test lint build ## Release build (clean, test, lint, build)

.PHONY: docs docs-full

# Documentation
docs: ## Generate provider docs with tfplugindocs
	@echo "Generating provider docs with tfplugindocs..."
	@command -v tfplugindocs >/dev/null 2>&1 || { \
	  echo "tfplugindocs not found." >&2; \
	  echo "Install with: go install github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@latest" >&2; \
	  exit 1; \
	}
	@tfplugindocs generate

docs-full: docs ## Generate docs (alias)

docs-check: ## Verify docs are up-to-date (fails if changes)
	@echo "Checking provider docs are up-to-date..."
	@git diff --quiet -- docs || { \
	  echo "Docs are outdated. Run 'make docs' to regenerate." >&2; \
	  git --no-pager diff -- docs; \
	  exit 1; \
	}

# Tagging helpers (require v-prefixed VERSION, e.g., VERSION=v0.1.0)
tag: ## Create an annotated git tag $(VERSION) on HEAD
	@if [ -z "$(VERSION)" ]; then echo "ERROR: VERSION is required (e.g., VERSION=v0.1.0)" >&2; exit 1; fi
	@case "$(VERSION)" in v*) ;; *) echo "ERROR: VERSION must be v-prefixed (e.g., v0.1.0)" >&2; exit 1;; esac
	@git tag -a $(VERSION) -m "Release $(VERSION)"
	@echo "Created tag $(VERSION)"

tag-push: ## Push tag $(VERSION) to origin
	@if [ -z "$(VERSION)" ]; then echo "ERROR: VERSION is required (e.g., VERSION=v0.1.0)" >&2; exit 1; fi
	@git push origin $(VERSION)

tag-delete: ## Delete local and remote tag $(VERSION)
	@if [ -z "$(VERSION)" ]; then echo "ERROR: VERSION is required (e.g., VERSION=v0.1.0)" >&2; exit 1; fi
	@echo "Deleting local tag $(VERSION)"
	@-git tag -d $(VERSION)
	@echo "Deleting remote tag $(VERSION)"
	@-git push --delete origin $(VERSION) || git push origin :refs/tags/$(VERSION)

tag-repush: tag-delete tag ## Delete and recreate tag $(VERSION), then push
	@$(MAKE) tag-push


.DEFAULT_GOAL := help

snapshot: ## Build snapshot artifacts locally with GoReleaser (no publish)
	@command -v goreleaser >/dev/null 2>&1 || { echo "Installing goreleaser..."; \
		GO111MODULE=on go install github.com/goreleaser/goreleaser/v2@latest; }
	@GOCACHE=$(PWD)/.cache/gobuild GOMODCACHE=$(PWD)/.cache/gomod goreleaser release --clean --snapshot --skip=publish
