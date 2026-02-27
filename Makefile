.PHONY: build build-patient build-auth build-sync run test test-patient test-auth test-sync test-all proto-gen lint clean

BINARY := gateway
BUILD_DIR := bin

build:
	go build -o $(BUILD_DIR)/$(BINARY) ./cmd/gateway

build-patient:
	go build -o $(BUILD_DIR)/patient-service ./services/patient/cmd

build-auth:
	go build -o $(BUILD_DIR)/auth-service ./services/auth/cmd

build-sync:
	go build -o $(BUILD_DIR)/sync-service ./services/sync/cmd

run: build
	./$(BUILD_DIR)/$(BINARY)

test:
	go test -v -race ./...

test-patient:
	go test -v -race ./services/patient/... ./pkg/fhir/... ./pkg/gitstore/... ./pkg/sqliteindex/...

test-auth:
	go test -v -race ./services/auth/... ./pkg/auth/...

test-sync:
	go test -v -race ./services/sync/... ./pkg/merge/...

test-all:
	go test -v -race ./...

proto-gen:
	buf generate

lint:
	golangci-lint run ./...

clean:
	rm -rf $(BUILD_DIR)
