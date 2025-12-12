.PHONY: all build build-linux dev install uninstall clean test lint help

# Variables
BINARY_NAME=monify
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT?=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE?=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS=-ldflags "-w -s -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(BUILD_DATE)"

# Directories
BUILD_DIR=build
INSTALL_DIR=/usr/local/bin
CONFIG_DIR=/etc/monify
SYSTEMD_DIR=/etc/systemd/system
LOG_DIR=/var/log/monify

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOMOD=$(GOCMD) mod
GORUN=$(GOCMD) run

all: clean build

## build: Build the binary for Linux
build:
	@echo "Building $(BINARY_NAME) $(VERSION) for Linux..."
	@mkdir -p $(BUILD_DIR)
	@ARCH=$${GOARCH:-amd64}; \
	GOOS=linux GOARCH=$$ARCH $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-$$ARCH ./cmd/agent; \
	echo "✓ Build complete: $(BUILD_DIR)/$(BINARY_NAME)-linux-$$ARCH"

## build-linux: Build for all Linux architectures
build-linux: clean
	@echo "Building for Linux platforms..."
	@mkdir -p $(BUILD_DIR)
	@echo "Building for linux/amd64..."
	@GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/agent
	@echo "Building for linux/arm64..."
	@GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 ./cmd/agent
	@echo "✓ Build complete for all Linux platforms"

## dev: Run agent in development mode (requires MONIFY_TOKEN env var)
dev:
	@echo "Running $(BINARY_NAME) in development mode..."
	@if [ -z "$$MONIFY_TOKEN" ]; then \
		echo "Error: MONIFY_TOKEN environment variable not set"; \
		echo ""; \
		echo "Usage:"; \
		echo "  make dev MONIFY_TOKEN=your_token_here"; \
		echo "  OR"; \
		echo "  export MONIFY_TOKEN=your_token_here"; \
		echo "  make dev"; \
		exit 1; \
	fi
	@echo "Token: $$MONIFY_TOKEN"
	@echo "Server: $${MONIFY_SERVER_URL:-https://api.monify.cloud/v1/agent/metrics}"
	@echo ""
	$(GORUN) ./cmd/agent start

## install: Install the agent (requires root)
install: build
	@echo "Installing $(BINARY_NAME)..."
	@if [ "$$(id -u)" != "0" ]; then \
		echo "Error: Installation requires root privileges. Please run with sudo."; \
		exit 1; \
	fi
	
	# Install binary
	install -m 755 $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(INSTALL_DIR)/$(BINARY_NAME)
	
	# Create directories
	mkdir -p $(CONFIG_DIR)
	mkdir -p $(LOG_DIR)
	
	# Install systemd service
	if [ -d $(SYSTEMD_DIR) ]; then \
		install -m 644 scripts/monify.service $(SYSTEMD_DIR)/monify.service; \
		systemctl daemon-reload; \
		echo "✓ Installed systemd service"; \
	fi
	
	@echo "✓ Installation complete!"
	@echo ""
	@echo "Next steps:"
	@echo "  1. Login: sudo $(BINARY_NAME) login"
	@echo "  2. Start service: sudo systemctl start $(BINARY_NAME)"
	@echo "  3. Enable auto-start: sudo systemctl enable $(BINARY_NAME)"
	@echo "  4. Check status: $(BINARY_NAME) status"

## uninstall: Uninstall the agent (requires root)
uninstall:
	@echo "Uninstalling $(BINARY_NAME)..."
	@if [ "$$(id -u)" != "0" ]; then \
		echo "Error: Uninstallation requires root privileges. Please run with sudo."; \
		exit 1; \
	fi
	
	# Stop and disable service
	-systemctl stop $(BINARY_NAME) 2>/dev/null
	-systemctl disable $(BINARY_NAME) 2>/dev/null
	
	# Remove files
	rm -f $(INSTALL_DIR)/$(BINARY_NAME)
	rm -f $(SYSTEMD_DIR)/$(BINARY_NAME).service
	
	# Ask before removing config and logs
	@echo ""
	@echo "Remove configuration and logs? [y/N]"
	@read -r response; \
	if [ "$$response" = "y" ] || [ "$$response" = "Y" ]; then \
		rm -rf $(CONFIG_DIR); \
		rm -rf $(LOG_DIR); \
		echo "✓ Removed config and logs"; \
	else \
		echo "⚠ Kept config and logs"; \
	fi
	
	systemctl daemon-reload 2>/dev/null || true
	@echo "✓ Uninstallation complete"

## clean: Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	$(GOCLEAN)
	@echo "✓ Clean complete"

## test: Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v -race -coverprofile=coverage.out ./...
	@echo "✓ Tests complete"

## test-coverage: Run tests with coverage report
test-coverage: test
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "✓ Coverage report generated: coverage.html"

## lint: Run linter
lint:
	@echo "Running linter..."
	@if command -v golangci-lint > /dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not found. Install it from https://golangci-lint.run/usage/install/"; \
		exit 1; \
	fi
	@echo "✓ Lint complete"

## fmt: Format code
fmt:
	@echo "Formatting code..."
	$(GOCMD) fmt ./...
	@echo "✓ Format complete"

## vet: Run go vet
vet:
	@echo "Running go vet..."
	$(GOCMD) vet ./...
	@echo "✓ Vet complete"

## mod-download: Download dependencies
mod-download:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	@echo "✓ Dependencies downloaded"

## mod-tidy: Tidy dependencies
mod-tidy:
	@echo "Tidying dependencies..."
	$(GOMOD) tidy
	@echo "✓ Dependencies tidied"

## help: Show this help message
help:
	@echo "Monify Agent - Makefile commands:"
	@echo ""
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'
	@echo ""
	@echo "Variables:"
	@echo "  VERSION     - Version tag (default: git describe)"
	@echo "  COMMIT      - Git commit hash (default: git rev-parse)"
	@echo "  BUILD_DATE  - Build timestamp (default: current UTC time)"
	@echo ""
	@echo "Examples:"
	@echo "  make build                           # Build for Linux amd64"
	@echo "  make build-linux                     # Build for all Linux platforms"
	@echo "  make dev MONIFY_TOKEN=xxx            # Run in dev mode"
	@echo "  make install                         # Install (requires sudo)"
	@echo "  make test                            # Run tests"
	@echo "  make lint                            # Run linter"

.DEFAULT_GOAL := help
