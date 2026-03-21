#!/bin/bash
# Benchmark script for the Go CLI using the current benchmark harness.
#
# Usage:
#   ./scripts/bench-go.sh                    # Run full benchmark with opendataloader-go
#   ./scripts/bench-go.sh --doc-id 01030...  # Run for a specific document
#   ./scripts/bench-go.sh --check-regression  # Run with regression check
#   ./scripts/bench-go.sh --skip-build        # Skip Go build step
#
# Assumptions:
# - scripts/build-go.sh produces the binary at bin/opendataloader-pdf.
# - tests/benchmark/run.py already registers the opendataloader-go engine.

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
BENCHMARK_DIR="$PROJECT_ROOT/tests/benchmark"

# Parse --skip-build so the benchmark runner does not see it.
SKIP_BUILD=false
ARGS=()
for arg in "$@"; do
    if [[ "$arg" == "--skip-build" ]]; then
        SKIP_BUILD=true
    else
        ARGS+=("$arg")
    fi
done

if [[ "$SKIP_BUILD" == "false" ]]; then
    echo "Building Go..."
    "$SCRIPT_DIR/build-go.sh"
else
    echo "Skipping Go build..."
fi

echo "Running benchmark with opendataloader-go..."
cd "$BENCHMARK_DIR"

if command -v uv &> /dev/null; then
    uv sync --quiet
    uv run python run.py "${ARGS[@]}" --engine opendataloader-go
else
    python3 run.py "${ARGS[@]}" --engine opendataloader-go
fi
