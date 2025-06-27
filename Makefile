BINARY_NAME=taskmaster
BUILD_DIR=build

.PHONY: build clean test run deps

deps:
	go mod tidy
	go mod download

build: deps
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/taskmaster

clean:
	@rm -rf $(BUILD_DIR)
	@rm -f taskmaster.log
	@rm -f /tmp/test_echo.* /tmp/long_running.*
	go clean

test:
	go test ./...

run: build
	./$(BUILD_DIR)/$(BINARY_NAME) -config configs/example.yml