#!/bin/bash
# Smoke benchmark wrapper for the Go CLI.
#
# Prerequisites:
# - `go` is installed and available on PATH.
# - Either `uv` is installed or the Python benchmark dependencies are already available to `python3`.
# - `tests/benchmark/pdfs` exists in the current checkout.
# - `scripts/build-go.sh` can build `bin/opendataloader-pdf`.
#
# Usage:
#   ./scripts/bench-go-smoke.sh
#   ./scripts/bench-go-smoke.sh --doc-id 01030000000001 --doc-id 01030000000002
#   BENCH_GO_SMOKE_DOC_IDS="01030000000001,01030000000002" ./scripts/bench-go-smoke.sh
#   ./scripts/bench-go-smoke.sh --doc-id 01030000000001 -- --check-regression
#
# The default document set is intentionally small and explicit so this stays a
# fast smoke check rather than a full benchmark run.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
BENCHMARK_PDF_DIR="$PROJECT_ROOT/tests/benchmark/pdfs"
DEFAULT_DOC_IDS=("01030000000001" "01030000000002" "01030000000003")

SKIP_BUILD=false
DOC_IDS=()
BENCH_ARGS=()

while [[ $# -gt 0 ]]; do
    case "$1" in
        --skip-build)
            SKIP_BUILD=true
            shift
            ;;
        --doc-id)
            if [[ $# -lt 2 ]]; then
                echo "Error: --doc-id requires a value" >&2
                exit 1
            fi
            DOC_IDS+=("$2")
            shift 2
            ;;
        --doc-id=*)
            DOC_IDS+=("${1#*=}")
            shift
            ;;
        --docs)
            if [[ $# -lt 2 ]]; then
                echo "Error: --docs requires a value" >&2
                exit 1
            fi
            read -r -a _SCRIPT_DOC_IDS <<< "${2//,/ }"
            DOC_IDS+=("${_SCRIPT_DOC_IDS[@]}")
            shift 2
            ;;
        --docs=*)
            DOC_SPEC="${1#*=}"
            read -r -a _SCRIPT_DOC_IDS <<< "${DOC_SPEC//,/ }"
            DOC_IDS+=("${_SCRIPT_DOC_IDS[@]}")
            shift
            ;;
        --)
            shift
            BENCH_ARGS=("$@")
            break
            ;;
        -*)
            BENCH_ARGS+=("$1")
            shift
            ;;
        *)
            DOC_IDS+=("$1")
            shift
            ;;
    esac
done

if [[ ${#DOC_IDS[@]} -eq 0 ]]; then
    if [[ -n "${BENCH_GO_SMOKE_DOC_IDS:-}" ]]; then
        read -r -a DOC_IDS <<< "${BENCH_GO_SMOKE_DOC_IDS//,/ }"
    else
        DOC_IDS=("${DEFAULT_DOC_IDS[@]}")
    fi
fi

if [[ ${#DOC_IDS[@]} -eq 0 ]]; then
    echo "Error: no document IDs were provided" >&2
    exit 1
fi

if [[ ! -d "$BENCHMARK_PDF_DIR" ]]; then
    echo "Error: benchmark PDFs not found at $BENCHMARK_PDF_DIR" >&2
    exit 1
fi

if [[ "$SKIP_BUILD" == "false" ]]; then
    if ! command -v go >/dev/null 2>&1; then
        echo "Error: go is not installed or not on PATH" >&2
        exit 1
    fi

    echo "Building Go CLI once for smoke benchmark runs..."
    "$SCRIPT_DIR/build-go.sh"
else
    if [[ ! -x "$PROJECT_ROOT/bin/opendataloader-pdf" ]]; then
        echo "Error: --skip-build was set but bin/opendataloader-pdf does not exist" >&2
        exit 1
    fi
fi

echo "Running smoke benchmark with opendataloader-go for: ${DOC_IDS[*]}"
for doc_id in "${DOC_IDS[@]}"; do
    echo "Benchmarking doc ID: $doc_id"
    if [[ ${#BENCH_ARGS[@]} -gt 0 ]]; then
        "$SCRIPT_DIR/bench-go.sh" --skip-build --doc-id "$doc_id" "${BENCH_ARGS[@]}"
    else
        "$SCRIPT_DIR/bench-go.sh" --skip-build --doc-id "$doc_id"
    fi
done
