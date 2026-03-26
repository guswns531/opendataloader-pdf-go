#!/bin/bash
# Benchmark script for opendataloader-pdf
#
# Usage:
#   ./scripts/bench.sh                       # Run full benchmark (Go binary)
#   ./scripts/bench.sh --doc-id 01030...     # Run for specific document
#   ./scripts/bench.sh --check-regression    # Run with regression check (CI)
#   ./scripts/bench.sh --hybrid docling-fast # Run with hybrid mode (requires docling-serve)
#   ./scripts/bench.sh --skip-build          # Skip Java build step

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
BENCHMARK_DIR="$PROJECT_ROOT/tests/benchmark"

# Parse --skip-build flag
SKIP_BUILD=false
ARGS=()
for arg in "$@"; do
    if [[ "$arg" == "--skip-build" ]]; then
        SKIP_BUILD=true
    else
        ARGS+=("$arg")
    fi
done

# Build Go binary if needed
if [[ "$SKIP_BUILD" == "false" ]]; then
    echo "Building Go binary..."
    (cd "$PROJECT_ROOT/go" && make build)
else
    echo "Skipping Go build..."
fi

# Install Python dependencies and run benchmark
echo "Running benchmark..."
cd "$BENCHMARK_DIR"

# Check if uv is available
if ! command -v uv &> /dev/null; then
    echo "Error: uv is not installed. Please install it first."
    exit 1
fi

# Sync dependencies
uv sync --quiet

# Run benchmark with all passed arguments (excluding --skip-build)
uv run python run.py "${ARGS[@]}"
