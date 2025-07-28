# Taskmaster Makefile

# Variables
BINARY_NAME=taskmaster
DAEMON_NAME=taskmaster
CLIENT_NAME=taskmasterctl
MAIN_PATH=./cmd/taskmaster
DAEMON_PATH=./cmd/taskmaster
CLIENT_PATH=./cmd/taskmasterctl
BUILD_DIR=.
CONFIG_DIR=configs
LOG_FILE=taskmaster.log
SOCKET_PATH=/tmp/taskmaster.sock

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Build flags
LDFLAGS=-ldflags "-X main.Version=$(shell git describe --tags --always --dirty 2>/dev/null || echo 'dev')"

.PHONY: all build clean test deps run help install uninstall daemon client

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

# Build daemon (nuevo)
daemon:
	@echo "🔨 Building daemon $(DAEMON_NAME)..."
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(DAEMON_NAME) $(DAEMON_PATH)
	@echo "✅ Daemon build complete: ./$(DAEMON_NAME)"

# Build client (nuevo)
client:
	@echo "🔨 Building client $(CLIENT_NAME)..."
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(CLIENT_NAME) $(CLIENT_PATH)
	@echo "✅ Client build complete: ./$(CLIENT_NAME)"

# Build both daemon and client (nuevo)
build-cs: daemon client

# Clean build artifacts
clean:
	@echo "🧹 Cleaning..."
	$(GOCLEAN)
	rm -f $(BUILD_DIR)/$(BINARY_NAME)
	rm -f $(BUILD_DIR)/$(CLIENT_NAME)
	rm -f $(LOG_FILE)
	rm -f $(SOCKET_PATH)
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

# Run with web interface
run-web: build
	@echo "🚀 Running $(BINARY_NAME) with web interface..."
	./$(BINARY_NAME) -config $(CONFIG_DIR)/example.yml --web-port 8080

# Run with web interface and custom config
run-web-config: build
	@if [ -z "$(CONFIG)" ]; then \
		echo "❌ Usage: make run-web-config CONFIG=path/to/config.yml [PORT=8080]"; \
		exit 1; \
	fi
	@PORT=$${PORT:-8080}; \
	echo "🚀 Running $(BINARY_NAME) with web interface on port $$PORT..."; \
	./$(BINARY_NAME) -config $(CONFIG) --web-port $$PORT

# Run daemon mode (nuevo)
run-daemon: daemon
	@echo "🚀 Running daemon mode..."
	./$(DAEMON_NAME) --daemon -config $(CONFIG_DIR)/example.yml

# Run daemon with web (nuevo)
run-daemon-web: daemon
	@echo "🚀 Running daemon with web interface..."
	./$(DAEMON_NAME) --daemon --web-port 8080 -config $(CONFIG_DIR)/example.yml

# Run daemon with custom config (nuevo)
run-daemon-config: daemon
	@if [ -z "$(CONFIG)" ]; then \
		echo "❌ Usage: make run-daemon-config CONFIG=path/to/config.yml"; \
		exit 1; \
	fi
	@echo "🚀 Running daemon with config $(CONFIG)..."
	./$(DAEMON_NAME) --daemon -config $(CONFIG)

# Run daemon with web and custom config (nuevo)
run-daemon-web-config: daemon
	@if [ -z "$(CONFIG)" ]; then \
		echo "❌ Usage: make run-daemon-web-config CONFIG=path/to/config.yml [PORT=8080]"; \
		exit 1; \
	fi
	@PORT=$${PORT:-8080}; \
	echo "🚀 Running daemon with web interface on port $$PORT..."; \
	./$(DAEMON_NAME) --daemon --web-port $$PORT -config $(CONFIG)

# Create example configuration (original)
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

# Install both binaries (nuevo)
install-all: build-cs
	@echo "📦 Installing both binaries..."
	sudo cp $(BUILD_DIR)/$(DAEMON_NAME) /usr/local/bin/
	sudo cp $(BUILD_DIR)/$(CLIENT_NAME) /usr/local/bin/
	@echo "✅ Both binaries installed to /usr/local/bin/"

# Uninstall binary from system
uninstall:
	@echo "🗑️  Uninstalling $(BINARY_NAME)..."
	sudo rm -f /usr/local/bin/$(BINARY_NAME)
	@echo "✅ $(BINARY_NAME) uninstalled"

# Uninstall both binaries (nuevo)
uninstall-all:
	@echo "🗑️  Uninstalling both binaries..."
	sudo rm -f /usr/local/bin/$(DAEMON_NAME)
	sudo rm -f /usr/local/bin/$(CLIENT_NAME)
	@echo "✅ Both binaries uninstalled"

# Development mode - build and run with live reload simulation
dev: clean build create-config
	@echo "🔧 Starting development mode..."
	./$(BINARY_NAME) -config $(CONFIG_DIR)/example.yml

# Development mode with daemon (nuevo)
dev-daemon: clean daemon create-config
	@echo "🔧 Starting development mode with daemon..."
	./$(DAEMON_NAME) --daemon -config $(CONFIG_DIR)/example.yml

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
	@echo "  daemon         - Build daemon only"
	@echo "  client         - Build client only"
	@echo "  build-cs       - Build both daemon and client"
	@echo "  clean          - Clean build artifacts"
	@echo "  deps           - Download and update dependencies"
	@echo ""
	@echo "Run commands:"
	@echo "  run            - Build and run with default config"
	@echo "  run-config     - Build and run with custom config (use CONFIG=path)"
	@echo "  run-web        - Build and run with web interface on port 8080"
	@echo "  run-web-config - Build and run with web interface and custom config"
	@echo "  dev            - Development mode (clean build + run)"
	@echo ""
	@echo "Run commands (Daemon Mode):"
	@echo "  run-daemon           - Run in daemon mode"
	@echo "  run-daemon-web       - Run daemon with web interface"
	@echo "  run-daemon-config    - Run daemon with custom config"
	@echo "  run-daemon-web-config - Run daemon with web and custom config"
	@echo "  dev-daemon           - Development mode with daemon"
	@echo ""
	@echo "Configuration:"
	@echo "  create-config  - Create example configuration file"
	@echo ""
	@echo "System commands:"
	@echo "  install        - Install binary to /usr/local/bin/"
	@echo "  install-all    - Install both binaries to /usr/local/bin/"
	@echo "  uninstall      - Remove binary from /usr/local/bin/"
	@echo "  uninstall-all  - Remove both binaries from /usr/local/bin/"
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
	@echo "  make run-config CONFIG=configs/example.yml"
	@echo "  make run-web"
	@echo "  make run-web-config CONFIG=configs/example.yml PORT=9000"
	@echo "  make run-daemon-web"
	@echo "  make run-daemon-web-config CONFIG=my.yml PORT=9000"
	@echo "  make dev"