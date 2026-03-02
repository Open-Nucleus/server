.PHONY: build build-all build-patient build-auth build-sync build-formulary build-anchor run test test-patient test-auth test-sync test-formulary test-anchor test-e2e test-all smoke proto-gen lint clean

BINARY := gateway
BUILD_DIR := bin

build:
	go build -o $(BUILD_DIR)/$(BINARY) ./cmd/gateway

build-all: build build-patient build-auth build-sync build-formulary build-anchor

build-patient:
	go build -o $(BUILD_DIR)/patient-service ./services/patient/cmd

build-auth:
	go build -o $(BUILD_DIR)/auth-service ./services/auth/cmd

build-sync:
	go build -o $(BUILD_DIR)/sync-service ./services/sync/cmd

build-formulary:
	go build -o $(BUILD_DIR)/formulary-service ./services/formulary/cmd

build-anchor:
	go build -o $(BUILD_DIR)/anchor-service ./services/anchor/cmd

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

test-formulary:
	go test -v -race ./services/formulary/...

test-anchor:
	go test -v -race ./services/anchor/... ./pkg/openanchor/...

test-e2e:
	go test -v -race -count=1 ./test/e2e/...

test-all:
	go test -v -race ./...

smoke:
	go run ./cmd/smoke

proto-gen:
	buf generate

lint:
	golangci-lint run ./...

clean:
	rm -rf $(BUILD_DIR)
