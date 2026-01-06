# VoiceType Makefile
# Linux Native Speech-to-Text App

# Variables
BINARY_NAME := VoiceType
BINARY_DIR := .
VERSION := 1.0.0
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "dev")
DATE := $(shell date -u +"%Y-%m-%d")
LDFLAGS := -ldflags "-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"

# Go commands
GO := go
GOFMT := gofmt
GOLINT := golangci-lint
GOTEST := go test

# Directories
SRC_DIRS := cmd internal pkg
BUILD_DIR := build

# Default target
.PHONY: all
all: build

# Build the binary
.PHONY: build
build:
	$(GO) build $(LDFLAGS) -o $(BINARY_NAME) ./cmd/voicetype/

# Build with debug symbols
.PHONY: build-debug
build-debug:
	$(GO) build -gcflags="all=-N -l" -o $(BINARY_NAME)-debug ./cmd/voicetype/

# Build for multiple platforms
.PHONY: build-all
build-all:
	@echo "Building for amd64..."
	GOOS=linux GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BINARY_NAME)-linux-amd64 ./cmd/voicetype/
	@echo "Building for arm64..."
	GOOS=linux GOARCH=arm64 $(GO) build $(LDFLAGS) -o $(BINARY_NAME)-linux-arm64 ./cmd/voicetype/

# Clean build artifacts
.PHONY: clean
clean:
	rm -f $(BINARY_NAME) $(BINARY_NAME)-debug
	rm -rf $(BUILD_DIR)
	rm -f $(BINARY_NAME)-linux-*

# Run tests
.PHONY: test
test:
	$(GOTEST) -v ./...

# Run tests with coverage
.PHONY: test-coverage
test-coverage:
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html

# Run tests for specific package
.PHONY: test-pkg
test-pkg:
	$(GOTEST) -v ./$(PKG)/

# Format code
.PHONY: fmt
fmt:
	$(GOFMT) -w $(SRC_DIRS)

# Check formatting
.PHONY: fmt-check
fmt-check:
	@if ! $(GOFMT) -d $(SRC_DIRS) | grep -q .; then \
		echo "Code is well formatted"; \
	else \
		echo "Code needs formatting:"; \
		$(GOFMT) -d $(SRC_DIRS); \
		exit 1; \
	fi

# Lint code
.PHONY: lint
lint:
	@if command -v $(GOLINT) >/dev/null 2>&1; then \
		$(GOLINT) run ./...; \
	else \
		echo "golangci-lint not installed. Install from https://golangci-lint.run/"; \
	fi

# Vet code
.PHONY: vet
vet:
	$(GO) vet ./...

# Install dependencies
.PHONY: deps
deps:
	$(GO) mod download
	$(GO) mod tidy

# Update dependencies
.PHONY: update-deps
update-deps:
	$(GO) get -u ./...
	$(GO) mod tidy

# Create release package
.PHONY: release
release: clean build-all
	mkdir -p $(BUILD_DIR)
	mv $(BINARY_NAME)-linux-* $(BUILD_DIR)/
	@echo "Release packages created in $(BUILD_DIR)/"

# Run the application
.PHONY: run
run: build
	./$(BINARY_NAME)

# Run with custom hotkey
.PHONY: run-hotkey
run-hotkey: build
	./$(BINARY_NAME) -hotkey=$(HOTKEY)

# Verbose run
.PHONY: run-verbose
run-verbose: build
	./$(BINARY_NAME) -v

# Generate coverage report
.PHONY: coverage
coverage: test-coverage
	@echo "Coverage report: coverage.html"

# Check dependencies
.PHONY: check-deps
check-deps:
	@echo "Checking Go installation..."
	@which $(GO) || (echo "Go not found!" && exit 1)
	@echo "Go version: $$($(GO) version)"

# Show help
.PHONY: help
help:
	@echo "VoiceType Build System"
	@echo ""
	@echo "Targets:"
	@echo "  all          - Build the binary (default)"
	@echo "  build        - Build the binary"
	@echo "  build-debug  - Build with debug symbols"
	@echo "  build-all    - Build for all platforms"
	@echo "  clean        - Clean build artifacts"
	@echo "  test         - Run all tests"
	@echo "  test-coverage- Run tests with coverage"
	@echo "  fmt          - Format code"
	@echo "  fmt-check    - Check formatting"
	@echo "  lint         - Lint code"
	@echo "  vet          - Vet code"
	@echo "  deps         - Install dependencies"
	@echo "  update-deps  - Update dependencies"
	@echo "  release      - Create release package"
	@echo "  run          - Run the application"
	@echo "  run-hotkey   - Run with custom hotkey"
	@echo "  run-verbose  - Run with verbose logging"
	@echo "  coverage     - Generate coverage report"
	@echo "  help         - Show this help"

# Docker build (optional)
.PHONY: docker-build
docker-build:
	docker build -t voicetype:latest .

# Verify build
.PHONY: verify
verify: build test
	@echo "Build and tests verified successfully!"

# Package for distribution
.PHONY: package
package: release
	@echo "Creating distribution package..."
	cd $(BUILD_DIR) && tar -czf VoiceType-$(VERSION)-linux.tar.gz *
	@echo "Distribution package: $(BUILD_DIR)/VoiceType-$(VERSION)-linux.tar.gz"

# CI/CD targets (for GitHub Actions)
.PHONY: ci-build
ci-build: deps build

.PHONY: ci-test
ci-test: deps test

.PHONY: ci-lint
ci-lint: deps vet fmt-check

.PHONY: ci-release
ci-release: deps test build-all

# Default to help if unknown target
%:
	@echo "Unknown target: $*. Run 'make help' for available targets."
	@exit 1
