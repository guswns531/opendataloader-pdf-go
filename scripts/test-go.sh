#!/bin/bash

# Local development test script for the Go CLI
# For CI/CD builds, use build-go.sh instead

set -e

# Prerequisites
command -v go >/dev/null || { echo "Error: go not found"; exit 1; }

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT_DIR="$SCRIPT_DIR/.."
PACKAGE_DIR="$ROOT_DIR"
cd "$PACKAGE_DIR"

# Run all Go tests in the current module.
go test "$@" ./...
