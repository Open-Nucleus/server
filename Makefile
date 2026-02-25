.PHONY: build run test proto-gen lint clean

BINARY := gateway
BUILD_DIR := bin

build:
	go build -o $(BUILD_DIR)/$(BINARY) ./cmd/gateway

run: build
	./$(BUILD_DIR)/$(BINARY)

test:
	go test -v -race ./...

proto-gen:
	buf generate

lint:
	golangci-lint run ./...

clean:
	rm -rf $(BUILD_DIR)
