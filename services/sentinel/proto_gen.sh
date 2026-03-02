#!/usr/bin/env bash
set -euo pipefail

# Generate Python gRPC stubs from proto files.
# Run from services/sentinel/ directory.

PROTO_ROOT="../../proto"
OUT_DIR="src/sentinel/gen"

rm -rf "$OUT_DIR"
mkdir -p "$OUT_DIR"

python -m grpc_tools.protoc \
  -I "$PROTO_ROOT" \
  --python_out="$OUT_DIR" \
  --grpc_python_out="$OUT_DIR" \
  common/v1/metadata.proto \
  sentinel/v1/sentinel.proto

# Fix imports in generated files (protoc generates absolute imports)
# sentinel_pb2_grpc.py imports sentinel.v1.sentinel_pb2 but we need relative
if [[ "$(uname)" == "Darwin" ]]; then
  SED_I="sed -i ''"
else
  SED_I="sed -i"
fi

# Fix import paths in grpc file
$SED_I 's/from sentinel\.v1 import sentinel_pb2/from sentinel.gen.sentinel.v1 import sentinel_pb2/' \
  "$OUT_DIR/sentinel/v1/sentinel_pb2_grpc.py" 2>/dev/null || true

# Create __init__.py files
find "$OUT_DIR" -type d -exec touch {}/__init__.py \;

echo "Proto generation complete. Output: $OUT_DIR"
