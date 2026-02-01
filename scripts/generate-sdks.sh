#!/usr/bin/env bash
# Generate client SDKs from the OpenAPI specification.
# Requires: npx (Node.js), or pre-installed openapi-generator-cli
#
# Usage: ./scripts/generate-sdks.sh [python|node|go|all]

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"
SPEC="$ROOT_DIR/schemas/openapi.yaml"
SDK_DIR="$ROOT_DIR/sdks"
GENERATOR="npx @openapitools/openapi-generator-cli generate"

if [ ! -f "$SPEC" ]; then
  echo "Error: OpenAPI spec not found at $SPEC"
  exit 1
fi

generate_python() {
  echo "Generating Python SDK..."
  $GENERATOR \
    -i "$SPEC" \
    -g python \
    -o "$SDK_DIR/python" \
    --additional-properties=packageName=civitas_ai,projectName=civitas-ai-sdk,packageVersion=0.1.0 \
    --skip-validate-spec
  echo "Python SDK generated at $SDK_DIR/python"
}

generate_node() {
  echo "Generating Node.js SDK..."
  $GENERATOR \
    -i "$SPEC" \
    -g typescript-fetch \
    -o "$SDK_DIR/node" \
    --additional-properties=npmName=@civitas-ai/sdk,npmVersion=0.1.0,supportsES6=true \
    --skip-validate-spec
  echo "Node.js SDK generated at $SDK_DIR/node"
}

generate_go() {
  echo "Generating Go SDK..."
  $GENERATOR \
    -i "$SPEC" \
    -g go \
    -o "$SDK_DIR/go" \
    --additional-properties=packageName=civitasai,isGoSubmodule=true \
    --skip-validate-spec
  echo "Go SDK generated at $SDK_DIR/go"
}

TARGET="${1:-all}"

mkdir -p "$SDK_DIR"

case "$TARGET" in
  python) generate_python ;;
  node)   generate_node ;;
  go)     generate_go ;;
  all)
    generate_python
    generate_node
    generate_go
    ;;
  *)
    echo "Usage: $0 [python|node|go|all]"
    exit 1
    ;;
esac

echo "SDK generation complete."
