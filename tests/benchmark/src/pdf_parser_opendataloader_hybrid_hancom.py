"""PDF parser using local opendataloader-pdf build with hancom hybrid mode."""

import os
import subprocess
import sys
from pathlib import Path


DEFAULT_URL = "https://dataloader.cloud.hancom.com/studio-lite/api"


def _find_local_binary() -> Path:
    """Find the locally built Go binary."""
    project_root = Path(__file__).parent.parent.parent.parent.resolve()

    binary = project_root / "bin" / "opendataloader-pdf"
    if binary.exists():
        return binary

    binary2 = project_root / "go" / "opendataloader-pdf"
    if binary2.exists():
        return binary2

    raise FileNotFoundError(
        f"Go binary not found at {binary}. "
        "Run `cd go && go build -o ../bin/opendataloader-pdf ./cmd/opendataloader-pdf/` first."
    )


def to_markdown(document_paths, input_path, output_dir):
    """Convert PDF to Markdown using hybrid mode with hancom backend.

    Args:
        _: Unused (for compatibility with engine dispatch signature).
        input_path: Input directory or single PDF file path.
        output_dir: Output directory for markdown files.

    Environment Variables:
        HANCOM_URL: Override URL for the Hancom API. Default: https://dataloader.cloud.hancom.com/studio-lite/api
        HYBRID_TIMEOUT: Request timeout in milliseconds. Default: 600000
    """
    binary_path = _find_local_binary()

    backend_url = os.environ.get("HANCOM_URL", DEFAULT_URL)
    timeout_ms = os.environ.get("HYBRID_TIMEOUT", "600000")

    # Build command - pass input path directly (directory or file)
    inputs = [str(path) for path in document_paths] if document_paths else [str(input_path)]

    command = [
        str(binary_path),
        *inputs,
        "--output-dir", str(output_dir),
        "--format", "markdown",
        "--image-output", "off",
        "--quiet",
        # Hybrid mode options
        "--hybrid", "hancom",
        "--hybrid-url", backend_url,
        "--hybrid-timeout", timeout_ms,
        "--hybrid-fallback",
        # Hancom uses full mode (no triage, all pages to backend)
        "--hybrid-mode", "full",
    ]

    # Run conversion
    result = subprocess.run(
        command,
        capture_output=True,
        text=True,
    )

    if result.returncode != 0:
        print(f"Error converting {input_path} (hancom hybrid mode):", file=sys.stderr)
        print(result.stderr, file=sys.stderr)
        # Don't raise - continue with other files
