.PHONY: build build-patient run test test-patient proto-gen lint clean

BINARY := gateway
BUILD_DIR := bin

build:
	go build -o $(BUILD_DIR)/$(BINARY) ./cmd/gateway

build-patient:
	go build -o $(BUILD_DIR)/patient-service ./services/patient/cmd

run: build
	./$(BUILD_DIR)/$(BINARY)

test:
	go test -v -race ./...

test-patient:
	go test -v -race ./services/patient/... ./pkg/...

proto-gen:
	buf generate

lint:
	golangci-lint run ./...

clean:
	rm -rf $(BUILD_DIR)
