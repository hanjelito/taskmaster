# Taskmaster Makefile

# Variables
BINARY_NAME=taskmaster
MAIN_PATH=./cmd/taskmaster
BUILD_DIR=.
CONFIG_DIR=configs
LOG_FILE=taskmaster.log

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Build flags
LDFLAGS=-ldflags "-X main.Version=$(shell git describe --tags --always --dirty 2>/dev/null || echo 'dev')"

.PHONY: all build clean test deps run help install uninstall

# Default target
all: deps build

# Build the binary
build:
	@echo "🔨 Building $(BINARY_NAME)..."
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "✅ Build complete: ./$(BINARY_NAME)"

# Build using the Go script
build-script:
	@echo "🔨 Building using build script..."
	$(GOCMD) run scripts/build.go

# Clean build artifacts
clean:
	@echo "🧹 Cleaning..."
	$(GOCLEAN)
	rm -f $(BUILD_DIR)/$(BINARY_NAME)
	rm -f $(LOG_FILE)
	rm -f nohup.out
	@echo "✅ Clean complete"

# Download dependencies
deps:
	@echo "📦 Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy
	@echo "✅ Dependencies updated"

# Run tests
test:
	@echo "🧪 Running tests..."
	$(GOTEST) -v ./...

# Run the program with default config
run: build
	@echo "🚀 Running $(BINARY_NAME)..."
	./$(BINARY_NAME) -config $(CONFIG_DIR)/example.yml

# Run with custom config
run-config: build
	@if [ -z "$(CONFIG)" ]; then \
		echo "❌ Usage: make run-config CONFIG=path/to/config.yml"; \
		exit 1; \
	fi
	@echo "🚀 Running $(BINARY_NAME) with config $(CONFIG)..."
	./$(BINARY_NAME) -config $(CONFIG)

# Create example configuration
create-config:
	@echo "📝 Creating example configuration..."
	@mkdir -p $(CONFIG_DIR)
	@if [ ! -f $(CONFIG_DIR)/example.yml ]; then \
		echo "Creating $(CONFIG_DIR)/example.yml..."; \
		cat > $(CONFIG_DIR)/example.yml << 'EOF'; \
programs:; \
  test_program:; \
    cmd: "sleep 30"; \
    numprocs: 2; \
    autostart: true; \
    autorestart: unexpected; \
    exitcodes: [0]; \
    starttime: 3; \
    startretries: 3; \
    stopsignal: TERM; \
    stoptime: 10; \
    stdout: /tmp/test_program.stdout; \
    stderr: /tmp/test_program.stderr; \
    env:; \
      TEST_VAR: "hello_world"; \
    workingdir: /tmp; \
    umask: "022"; \
  failing_program:; \
    cmd: "exit 1"; \
    numprocs: 1; \
    autostart: false; \
    autorestart: never; \
    exitcodes: [0]; \
    starttime: 1; \
    startretries: 3; \
    stopsignal: TERM; \
    stoptime: 5; \
EOF; \
		sed -i 's/;/\n    /g' $(CONFIG_DIR)/example.yml; \
		echo "✅ Created $(CONFIG_DIR)/example.yml"; \
	else \
		echo "⚠️  $(CONFIG_DIR)/example.yml already exists"; \
	fi

# Install binary to system
install: build
	@echo "📦 Installing $(BINARY_NAME)..."
	sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	@echo "✅ $(BINARY_NAME) installed to /usr/local/bin/"

# Uninstall binary from system
uninstall:
	@echo "🗑️  Uninstalling $(BINARY_NAME)..."
	sudo rm -f /usr/local/bin/$(BINARY_NAME)
	@echo "✅ $(BINARY_NAME) uninstalled"

# Development mode - build and run with live reload simulation
dev: clean build create-config
	@echo "🔧 Starting development mode..."
	./$(BINARY_NAME) -config $(CONFIG_DIR)/example.yml

# Check for required tools
check-tools:
	@echo "🔍 Checking required tools..."
	@command -v go >/dev/null 2>&1 || { echo "❌ Go is required but not installed. Please install Go."; exit 1; }
	@echo "✅ Go found: $$(go version)"

# Format code
fmt:
	@echo "📐 Formatting code..."
	$(GOCMD) fmt ./...
	@echo "✅ Code formatted"

# Lint code (requires golangci-lint)
lint:
	@echo "🔍 Linting code..."
	@command -v golangci-lint >/dev/null 2>&1 || { echo "⚠️  golangci-lint not found, skipping lint"; exit 0; }
	golangci-lint run
	@echo "✅ Linting complete"

# Show help
help:
	@echo "🚀 Taskmaster Makefile Commands:"
	@echo ""
	@echo "Build commands:"
	@echo "  build          - Build the binary"
	@echo "  build-script   - Build using Go build script"
	@echo "  clean          - Clean build artifacts"
	@echo "  deps           - Download and update dependencies"
	@echo ""
	@echo "Run commands:"
	@echo "  run            - Build and run with default config"
	@echo "  run-config     - Build and run with custom config (use CONFIG=path)"
	@echo "  dev            - Development mode (clean build + run)"
	@echo ""
	@echo "Configuration:"
	@echo "  create-config  - Create example configuration file"
	@echo ""
	@echo "System commands:"
	@echo "  install        - Install binary to /usr/local/bin/"
	@echo "  uninstall      - Remove binary from /usr/local/bin/"
	@echo ""
	@echo "Development:"
	@echo "  test           - Run tests"
	@echo "  fmt            - Format code"
	@echo "  lint           - Lint code (requires golangci-lint)"
	@echo "  check-tools    - Check required tools"
	@echo ""
	@echo "Other:"
	@echo "  help           - Show this help message"
	@echo ""
	@echo "Examples:"
	@echo "  make build"
	@echo "  make run"
	@echo "  make run-config CONFIG=configs/production.yml"
	@echo "  make dev"