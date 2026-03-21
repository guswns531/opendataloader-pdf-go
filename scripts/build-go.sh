#!/bin/bash

# CI/CD build script for the Go CLI
# For local development, use test-go.sh instead

set -e

# Prerequisites
command -v go >/dev/null || { echo "Error: go not found"; exit 1; }

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT_DIR="$SCRIPT_DIR/.."
PACKAGE_DIR="$ROOT_DIR"
OUTPUT_DIR="$ROOT_DIR/bin"
OUTPUT_BIN="$OUTPUT_DIR/opendataloader-pdf"
cd "$PACKAGE_DIR"

# Run tests first so the build only produces an artifact from a passing tree.
go test ./...

# Build the current Go CLI entrypoint.
mkdir -p "$OUTPUT_DIR"
go build -o "$OUTPUT_BIN" ./cmd/opendataloader-pdf

echo "Build completed successfully: $OUTPUT_BIN"
