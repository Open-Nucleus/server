#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$ROOT"

export NUCLEUS_BOOTSTRAP_SECRET="${NUCLEUS_BOOTSTRAP_SECRET:-demo}"

echo "=== Open Nucleus Demo ==="
echo

# Build
echo "[1/3] Building..."
go build -o bin/nucleus ./cmd/nucleus

# Seed
echo "[2/3] Seeding demo data..."
go run ./cmd/seed

# Run
echo "[3/3] Starting server on :8080..."
echo
echo "  Health:  http://localhost:8080/health"
echo "  API:     http://localhost:8080/api/v1/"
echo "  FHIR:    http://localhost:8080/fhir/metadata"
echo
exec ./bin/nucleus
