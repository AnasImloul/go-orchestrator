# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=go-orchestrator
BINARY_UNIX=$(BINARY_NAME)_unix

# Build the example application
.PHONY: build
build:
	$(GOBUILD) -o bin/example ./cmd/example

# Build for multiple platforms
.PHONY: build-all
build-all:
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o bin/$(BINARY_UNIX) ./cmd/example

# Run tests
.PHONY: test
test:
	$(GOTEST) -v ./...

# Run tests with coverage
.PHONY: test-coverage
test-coverage:
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

# Run the example application
.PHONY: run
run:
	$(GOBUILD) -o bin/example ./cmd/example && ./bin/example

# Clean build artifacts
.PHONY: clean
clean:
	$(GOCLEAN)
	rm -rf bin/
	rm -f coverage.out coverage.html

# Download dependencies
.PHONY: deps
deps:
	$(GOMOD) download
	$(GOMOD) tidy

# Format code
.PHONY: fmt
fmt:
	$(GOCMD) fmt ./...

# Lint code
.PHONY: lint
lint:
	golangci-lint run

# Install development tools
.PHONY: install-tools
install-tools:
	$(GOGET) -u github.com/golangci/golangci-lint/cmd/golangci-lint

# Generate go.sum
.PHONY: mod-tidy
mod-tidy:
	$(GOMOD) tidy

# Run all checks
.PHONY: check
check: fmt lint test

# Release commands
.PHONY: release-dry-run
release-dry-run:
	@echo "Running semantic-release in dry-run mode..."
	npx semantic-release --dry-run

.PHONY: release
release:
	@echo "Creating release..."
	npx semantic-release

.PHONY: commit
commit:
	@echo "Running conventional commit helper..."
	./scripts/commit.sh

# Help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build         - Build the example application"
	@echo "  build-all     - Build for multiple platforms"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  run           - Build and run the example application"
	@echo "  clean         - Clean build artifacts"
	@echo "  deps          - Download and tidy dependencies"
	@echo "  fmt           - Format code"
	@echo "  lint          - Lint code"
	@echo "  install-tools - Install development tools"
	@echo "  mod-tidy      - Tidy go.mod and go.sum"
	@echo "  check         - Run fmt, lint, and test"
	@echo "  commit        - Run conventional commit helper"
	@echo "  release-dry-run - Test semantic release without creating tags"
	@echo "  release       - Create a new release (use with caution)"
	@echo "  help          - Show this help message"
