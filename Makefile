BINARY_NAME=taskmaster
BUILD_DIR=build

.PHONY: build clean test run

build:
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/taskmaster

clean:
	@rm -rf $(BUILD_DIR)
	go clean

test:
	go test ./...

run: build
	./$(BUILD_DIR)/$(BINARY_NAME) -configs/example.yml

install: build
	sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/

lint:
	golangci-lint run

format:
	go fmt ./...

