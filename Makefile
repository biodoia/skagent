# SkAgent Makefile
# AI-Powered Spec-Driven Development Assistant

VERSION := 1.0.0
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Go parameters
GOCMD := go
GOBUILD := $(GOCMD) build
GOCLEAN := $(GOCMD) clean
GOTEST := $(GOCMD) test
GOMOD := $(GOCMD) mod

# Build flags
LDFLAGS := -ldflags "-s -w -X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME) -X main.gitCommit=$(GIT_COMMIT)"

# Binary names
BINARY_NAME := skagent
BINARY_LINUX := $(BINARY_NAME)-linux-amd64
BINARY_DARWIN := $(BINARY_NAME)-darwin-amd64
BINARY_DARWIN_ARM := $(BINARY_NAME)-darwin-arm64
BINARY_WINDOWS := $(BINARY_NAME)-windows-amd64.exe

# Directories
BUILD_DIR := build
CMD_DIR := ./cmd/skagent

.PHONY: all build clean test deps install uninstall cross-compile help

# Default target
all: deps build

# Build for current platform
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)
	@echo "Built: $(BUILD_DIR)/$(BINARY_NAME)"

# Build optimized release
release:
	@echo "Building optimized release..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)
	@echo "Built: $(BUILD_DIR)/$(BINARY_NAME)"

# Cross-compile for all platforms
cross-compile: clean
	@echo "Cross-compiling for all platforms..."
	@mkdir -p $(BUILD_DIR)

	# Linux AMD64
	@echo "  Building for Linux (amd64)..."
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_LINUX) $(CMD_DIR)

	# macOS AMD64
	@echo "  Building for macOS (amd64)..."
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_DARWIN) $(CMD_DIR)

	# macOS ARM64 (Apple Silicon)
	@echo "  Building for macOS (arm64)..."
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_DARWIN_ARM) $(CMD_DIR)

	# Windows AMD64
	@echo "  Building for Windows (amd64)..."
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_WINDOWS) $(CMD_DIR)

	@echo "All binaries built in $(BUILD_DIR)/"
	@ls -la $(BUILD_DIR)/

# Install to system
install: build
	@echo "Installing $(BINARY_NAME)..."
	@sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)
	@sudo chmod +x /usr/local/bin/$(BINARY_NAME)
	@echo "Installed to /usr/local/bin/$(BINARY_NAME)"

# Install to user bin (no sudo)
install-user: build
	@echo "Installing $(BINARY_NAME) to ~/bin..."
	@mkdir -p ~/bin
	@cp $(BUILD_DIR)/$(BINARY_NAME) ~/bin/$(BINARY_NAME)
	@chmod +x ~/bin/$(BINARY_NAME)
	@echo "Installed to ~/bin/$(BINARY_NAME)"
	@echo "Make sure ~/bin is in your PATH"

# Uninstall from system
uninstall:
	@echo "Uninstalling $(BINARY_NAME)..."
	@sudo rm -f /usr/local/bin/$(BINARY_NAME)
	@rm -f ~/bin/$(BINARY_NAME)
	@echo "Uninstalled $(BINARY_NAME)"

# Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

# Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	@rm -rf $(BUILD_DIR)
	@echo "Clean complete"

# Run the application
run: build
	@$(BUILD_DIR)/$(BINARY_NAME)

# Run setup wizard
setup: build
	@$(BUILD_DIR)/$(BINARY_NAME) setup

# Show version info
version:
	@echo "Version: $(VERSION)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Git Commit: $(GIT_COMMIT)"

# Show help
help:
	@echo "SkAgent Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make              Build for current platform"
	@echo "  make build        Build for current platform"
	@echo "  make release      Build optimized release binary"
	@echo "  make cross-compile Build for all platforms"
	@echo "  make install      Install to /usr/local/bin (requires sudo)"
	@echo "  make install-user Install to ~/bin (no sudo)"
	@echo "  make uninstall    Remove installed binary"
	@echo "  make run          Build and run"
	@echo "  make setup        Run setup wizard"
	@echo "  make test         Run tests"
	@echo "  make deps         Download dependencies"
	@echo "  make clean        Remove build artifacts"
	@echo "  make version      Show version info"
	@echo "  make help         Show this help"
