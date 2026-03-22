#!/bin/bash
# Open Nucleus — Full Demo Script
# Run from the open-nucleus/ project root.
#
# This script:
# 1. Seeds demo patient data (6 patients, cholera outbreak)
# 2. Starts the Go backend server
# 3. Runs the Sentinel AI outbreak detection
# 4. Triggers Hedera HCS anchoring
#
# Prerequisites:
#   - Go 1.25+, Python 3.11+, pnpm installed
#   - open-sentinel repo cloned alongside this repo
#   - Hedera testnet credentials in config.yaml

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
SENTINEL_DIR="$(dirname "$PROJECT_DIR")/open-sentinel"

cd "$PROJECT_DIR"

echo ""
echo "============================================================"
echo "  OPEN NUCLEUS — Demo Setup"
echo "============================================================"
echo ""

# Step 1: Seed demo data
echo "[1/4] Seeding demo data (6 patients + cholera outbreak)..."
rm -rf data/repo data/nucleus.db 2>/dev/null || true
go run ./cmd/seed
echo ""

# Step 2: Start the server
echo "[2/4] Starting server on :8080..."
NUCLEUS_BOOTSTRAP_SECRET=demo go run ./cmd/nucleus &
SERVER_PID=$!
echo "  Server PID: $SERVER_PID"
sleep 3

# Verify server is up
if curl -s http://localhost:8080/health | grep -q "healthy"; then
  echo "  Server is healthy."
else
  echo "  ERROR: Server failed to start!"
  kill $SERVER_PID 2>/dev/null
  exit 1
fi
echo ""

# Step 3: Run Sentinel outbreak detection
echo "[3/4] Running Sentinel AI outbreak detection..."
if [ -d "$SENTINEL_DIR" ]; then
  cd "$SENTINEL_DIR"
  python3 scripts/demo.py --repo "$PROJECT_DIR/data/repo" 2>&1
  cd "$PROJECT_DIR"
else
  echo "  SKIP: open-sentinel not found at $SENTINEL_DIR"
fi
echo ""

# Step 4: Trigger Hedera anchor
echo "[4/4] Triggering Hedera HCS anchor..."
# Login to get a token
TOKEN=$(curl -s http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"device_id":"demo-script-device-00000000000000000000","public_key":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA","challenge_response":{"nonce":"demo","signature":"demo","timestamp":"demo"},"practitioner_id":"demo-clinician","bootstrap_secret":"demo"}' \
  | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['token'])" 2>/dev/null)

if [ -n "$TOKEN" ]; then
  ANCHOR_RESULT=$(curl -s -X POST http://localhost:8080/api/v1/anchor/trigger \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -H "X-Break-Glass: true")
  echo "  Anchor result: $ANCHOR_RESULT" | head -1
else
  echo "  SKIP: Could not get auth token for anchoring"
fi
echo ""

echo "============================================================"
echo "  Demo Environment Ready!"
echo "============================================================"
echo ""
echo "  Server running on http://localhost:8080 (PID: $SERVER_PID)"
echo "  Bootstrap secret: demo"
echo ""
echo "  To start the desktop app:"
echo "    cd open-nucleus-app && pnpm tauri dev"
echo ""
echo "  To stop the server:"
echo "    kill $SERVER_PID"
echo ""
echo "  Hedera anchoring topic: 0.0.8334822"
echo "  Verify on HashScan: https://hashscan.io/testnet/topic/0.0.8334822"
echo ""
