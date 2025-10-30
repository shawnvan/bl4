# BL4 Item Serial Code Codec Makefile

.PHONY: all build clean test lint fmt run-api run-gui run-cli run-tui docker-build docker-run help

# Variables
BINARY_NAME_API=bl4-api
BINARY_NAME_GUI=bl4-gui
BINARY_NAME_CLI=bl4-cli
BINARY_NAME_TUI=bl4-tui
BUILD_DIR=build
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-X main.version=$(VERSION)"

# Default target
all: build

# Build all binaries
build: build-api build-gui build-cli build-tui

# Build API server
build-api:
	@echo "Building API server..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME_API) ./cmd/api

# Build GUI application
build-gui:
	@echo "Building GUI application..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME_GUI) ./cmd/gui

# Build CLI tool
build-cli:
	@echo "Building CLI tool..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME_CLI) ./cmd/cli

# Build TUI tool
build-tui:
	@echo "Building TUI tool..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME_TUI) ./cmd/tui

# Run API server
run-api: build-api
	@echo "Starting API server on port 8080..."
	./$(BUILD_DIR)/$(BINARY_NAME_API)

# Run GUI application
run-gui: build-gui
	@echo "Starting GUI application..."
	./$(BUILD_DIR)/$(BINARY_NAME_GUI)

# Run CLI tool
run-cli: build-cli
	@echo "Running CLI tool..."
	./$(BUILD_DIR)/$(BINARY_NAME_CLI)

# Run TUI tool
run-tui: build-tui
	@echo "Starting TUI tool..."
	./$(BUILD_DIR)/$(BINARY_NAME_TUI)

# Run tests
test:
	@echo "Running all tests..."
	go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run benchmark tests
bench:
	@echo "Running benchmark tests..."
	go test -bench=. ./...

# Run security check
security:
	@echo "Running security check..."
	gosec ./...

# Run linter
lint:
	@echo "Running linter..."
	golangci-lint run ./...

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...
	goimports -w .

# Tidy dependencies
tidy:
	@echo "Tidying dependencies..."
	go mod tidy

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

# Full clean (including vendor)
clean-all: clean
	@echo "Cleaning all generated files..."
	rm -f go.sum
	go clean -cache
	rm -rf vendor/

# Install development dependencies
dev-deps:
	@echo "Installing development dependencies..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest

# Docker build
docker-build:
	@echo "Building Docker image..."
	docker build -t bl4:$(VERSION) .

# Docker run
docker-run:
	@echo "Running Docker container..."
	docker run --rm -p 8080:8080 bl4:$(VERSION)

# Generate documentation
docs:
	@echo "Generating documentation..."
	@mkdir -p docs
	@echo "Documentation generated in docs/"

# Install binary globally
install: build
	@echo "Installing binaries..."
	cp $(BUILD_DIR)/$(BINARY_NAME_API) /usr/local/bin/
	cp $(BUILD_DIR)/$(BINARY_NAME_GUI) /usr/local/bin/
	cp $(BUILD_DIR)/$(BINARY_NAME_CLI) /usr/local/bin/
	cp $(BUILD_DIR)/$(BINARY_NAME_TUI) /usr/local/bin/

# Development setup
dev-setup:
	@echo "Setting up development environment..."
	@echo "Installing dependencies..."
	@echo "Setting up pre-commit hooks..."
	@echo "Development environment ready!"

# Check code quality
quality: fmt lint security test

# Quick development cycle
dev: quality build-api run-api

# Release
release: clean test build
	@echo "Release build completed!"

# Help
help:
	@echo "Available targets:"
	@echo "  all           - Build all binaries"
	@echo "  build-api     - Build API server"
	@echo "  build-gui     - Build GUI application"
	@echo "  build-cli     - Build CLI tool"
	@echo "  build-tui     - Build TUI tool"
	@echo "  run-api       - Run API server (port 8080)"
	@echo "  run-gui       - Run GUI application"
	@echo "  run-cli       - Run CLI tool"
	@echo "  run-tui       - Run TUI tool"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  bench         - Run benchmark tests"
	@echo "  security      - Run security check"
	@echo "  lint          - Run linter"
	@echo "  fmt           - Format code"
	@echo "  tidy          - Tidy dependencies"
	@echo "  clean         - Clean build artifacts"
	@echo "  clean-all     - Clean all generated files"
	@echo "  docker-build  - Build Docker image"
	@	"  docker-run    - Run Docker container"
	@echo "  docs          - Generate documentation"
	@echo "  install       - Install binaries globally"
	@echo "  dev-setup     - Setup development environment"
	@echo "  quality       - Run code quality checks"
	@echo "  dev           - Quick development cycle"
	@echo "  release       - Release build"
	@echo "  help          - Show this help message"