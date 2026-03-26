"""PDF parser using local opendataloader-pdf build (Go binary)."""

import subprocess
import sys
from pathlib import Path


def _find_local_binary() -> Path:
    """Find the locally built Go binary.

    project_root resolves to the tripoli/ directory.
    """
    # tests/benchmark/src -> tests/benchmark -> tests -> tripoli
    project_root = Path(__file__).parent.parent.parent.parent.resolve()

    # Primary: bin/opendataloader-pdf (built by `make build`)
    binary = project_root / "bin" / "opendataloader-pdf"
    if binary.exists():
        return binary

    # Fallback: built directly in go/
    binary2 = project_root / "go" / "opendataloader-pdf"
    if binary2.exists():
        return binary2

    raise FileNotFoundError(
        f"Go binary not found at {binary}. "
        "Run `cd go && go build -o ../bin/opendataloader-pdf ./cmd/opendataloader-pdf/` first."
    )


def to_markdown(document_paths, input_path, output_dir):
    """Convert PDF to Markdown using local Go binary."""
    binary_path = _find_local_binary()
    inputs = [str(path) for path in document_paths] if document_paths else [str(input_path)]

    command = [
        str(binary_path),
        *inputs,
        "--output-dir", str(output_dir),
        "--format", "markdown",
        "--table-method", "cluster",
        "--image-output", "off",
        "--quiet",
    ]

    result = subprocess.run(
        command,
        capture_output=True,
        text=True,
    )

    if result.returncode != 0:
        print(f"Error converting {input_path}:", file=sys.stderr)
        print(result.stderr, file=sys.stderr)
        # Don't raise - continue with other files
