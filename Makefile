.PHONY: build build-nucleus build-sentinel run test test-patient test-auth test-sync test-formulary test-anchor test-fhir test-sentinel test-e2e test-all smoke proto-gen proto-gen-python run-sentinel lint clean seed demo

BUILD_DIR := bin

# Default: build the monolith
build: build-nucleus

build-nucleus:
	go build -o $(BUILD_DIR)/nucleus ./cmd/nucleus

run: build-nucleus
	./$(BUILD_DIR)/nucleus

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
	go test -v -race ./services/anchor/... ./pkg/merge/openanchor/...

test-fhir:
	go test -v -race ./internal/handler/fhir/...

build-sentinel:
	cd services/sentinel && pip install -e . 2>/dev/null || true

test-sentinel:
	cd services/sentinel && PYTHONPATH=src python3 -m pytest tests/ -v

run-sentinel:
	cd services/sentinel && PYTHONPATH=src python3 -m sentinel.main

proto-gen-python:
	cd services/sentinel && bash proto_gen.sh

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

seed:
	go run ./cmd/seed

demo: build-nucleus seed
	NUCLEUS_BOOTSTRAP_SECRET=demo ./$(BUILD_DIR)/nucleus

clean:
	rm -rf $(BUILD_DIR) data/
