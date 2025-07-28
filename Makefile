# Taskmaster Makefile

BINARY_NAME=taskmaster
MAIN_PATH=./cmd/taskmaster
CONFIG_DIR=configs

GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOMOD=$(GOCMD) mod

.PHONY: all build clean test deps run run-web help

all: deps build

build:
	$(GOBUILD) -o $(BINARY_NAME) $(MAIN_PATH)

clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f *.log

deps:
	$(GOMOD) download
	$(GOMOD) tidy

test:
	$(GOTEST) ./...

run: build
	./$(BINARY_NAME) -config $(CONFIG_DIR)/example.yml

run-web: build
	./$(BINARY_NAME) -config $(CONFIG_DIR)/example.yml -web-port 8080

help:
	@echo "Available targets:"
	@echo "  build    - Build the binary"
	@echo "  clean    - Clean build artifacts"
	@echo "  deps     - Download dependencies"
	@echo "  test     - Run tests"
	@echo "  run      - Build and run with example config"
	@echo "  run-web  - Build and run with web interface on port 8080"
	@echo "  help     - Show this help"