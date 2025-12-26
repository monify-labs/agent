# Monify Agent Makefile
# Build and release automation for Linux amd64/arm64

# Variables (can be overridden by environment)
BINARY_NAME := monify
VERSION ?= $(shell grep 'Version   = ' internal/config/config.go | cut -d'"' -f2)
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Build flags
LDFLAGS := -s -w
LDFLAGS += -X 'github.com/monify-labs/agent/internal/config.Version=$(VERSION)'
LDFLAGS += -X 'github.com/monify-labs/agent/internal/config.Commit=$(COMMIT)'
LDFLAGS += -X 'github.com/monify-labs/agent/internal/config.BuildDate=$(BUILD_DATE)'

# Directories
BUILD_DIR := build
CMD_DIR := cmd/monify

# Go settings
GOOS := linux
CGO_ENABLED := 0

.PHONY: all build build-amd64 build-arm64 clean test lint fmt install uninstall dev help

# Default target
all: clean build

# Build for current architecture (or specified GOARCH)
build:
	@echo "Building $(BINARY_NAME) v$(VERSION) for $(GOOS)/$(GOARCH)..."
	@mkdir -p $(BUILD_DIR)
	GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=$(CGO_ENABLED) go build \
		-ldflags "$(LDFLAGS)" \
		-o $(BUILD_DIR)/$(BINARY_NAME)-$(GOOS)-$(GOARCH) \
		./$(CMD_DIR)
	@echo "Built: $(BUILD_DIR)/$(BINARY_NAME)-$(GOOS)-$(GOARCH)"

# Build for Linux amd64
build-amd64:
	@$(MAKE) build GOARCH=amd64

# Build for Linux arm64
build-arm64:
	@$(MAKE) build GOARCH=arm64

# Build for all platforms
build-all: build-amd64 build-arm64
	@echo "All builds complete!"
	@ls -la $(BUILD_DIR)/

# Clean build artifacts
clean:
	@echo "Cleaning build directory..."
	@rm -rf $(BUILD_DIR)
	@go clean

# Run tests
test:
	@echo "Running tests..."
	@go test -v -race -cover ./...

# Run linter
lint:
	@echo "Running linter..."
	@golangci-lint run ./...

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@goimports -w .

# Run locally (for development)
dev:
	@echo "Running in development mode..."
	MONIFY_DEBUG=true go run ./$(CMD_DIR) run

# Install locally (requires root)
install: build-amd64
	@echo "Installing $(BINARY_NAME)..."
	@if [ "$$(id -u)" != "0" ]; then \
		echo "Error: install requires root privileges"; \
		exit 1; \
	fi
	@cp $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 /usr/local/bin/$(BINARY_NAME)
	@chmod 755 /usr/local/bin/$(BINARY_NAME)
	@mkdir -p /etc/monify
	@chmod 700 /etc/monify
	@echo "Installed to /usr/local/bin/$(BINARY_NAME)"

# Uninstall (requires root)
uninstall:
	@echo "Uninstalling $(BINARY_NAME)..."
	@if [ "$$(id -u)" != "0" ]; then \
		echo "Error: uninstall requires root privileges"; \
		exit 1; \
	fi
	@systemctl stop monify 2>/dev/null || true
	@systemctl disable monify 2>/dev/null || true
	@rm -f /usr/local/bin/$(BINARY_NAME)
	@rm -f /etc/systemd/system/monify.service
	@rm -rf /etc/monify
	@rm -rf /var/log/monify
	@systemctl daemon-reload 2>/dev/null || true
	@echo "Uninstalled successfully"

# Check dependencies
deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

# Verify build
verify: clean build-all test
	@echo "Build verified successfully!"

# Show version info
version:
	@echo "Version: $(VERSION)"
	@echo "Commit: $(COMMIT)"
	@echo "Build Date: $(BUILD_DATE)"

# Help
help:
	@echo "Monify Agent Makefile"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  all          Clean and build for current architecture"
	@echo "  build        Build for current/specified architecture"
	@echo "  build-amd64  Build for Linux amd64"
	@echo "  build-arm64  Build for Linux arm64"
	@echo "  build-all    Build for all platforms"
	@echo "  clean        Remove build artifacts"
	@echo "  test         Run tests"
	@echo "  lint         Run linter"
	@echo "  fmt          Format code"
	@echo "  dev          Run locally in debug mode"
	@echo "  install      Install binary (requires root)"
	@echo "  uninstall    Remove installation (requires root)"
	@echo "  deps         Download dependencies"
	@echo "  verify       Clean, build, and test"
	@echo "  version      Show version info"
	@echo "  help         Show this help"
	@echo ""
	@echo "Examples:"
	@echo "  make build-all"
	@echo "  make build GOARCH=arm64"
	@echo "  sudo make install"
