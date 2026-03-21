"""PDF parser using the Go CLI build."""

from __future__ import annotations

import os
import shutil
import subprocess
import sys
from pathlib import Path
from typing import List


def _project_root() -> Path:
    """Return the repository root for the benchmark checkout."""
    return Path(__file__).parent.parent.parent.parent.resolve()


def _candidate_binaries() -> List[Path]:
    """Return likely locations for a compiled Go benchmark binary."""
    root = _project_root()
    return [
        root / "bin" / "opendataloader-pdf",
        root / "bin" / "opendataloader-pdf-go",
        root / "opendataloader-pdf",
        root / "opendataloader-pdf-go",
    ]


def _build_command(input_path: Path, output_dir: Path) -> List[str]:
    """Build the command used to invoke the Go CLI."""
    binary_override = os.environ.get("OPENDATALOADER_GO_BIN")
    if binary_override:
        binary_path = Path(binary_override).expanduser()
        if not binary_path.is_absolute():
            binary_path = _project_root() / binary_path
        if binary_path.exists():
            executable = str(binary_path)
        else:
            executable = binary_override
        return [
            executable,
            "--output-dir",
            str(output_dir),
            "--format",
            "markdown",
            "--table-method",
            "cluster",
            "--image-output",
            "off",
            "--quiet",
            str(input_path),
        ]

    for candidate in _candidate_binaries():
        if candidate.exists() and candidate.is_file() and os.access(candidate, os.X_OK):
            return [
                str(candidate),
                "--output-dir",
                str(output_dir),
                "--format",
                "markdown",
                "--table-method",
                "cluster",
                "--image-output",
                "off",
                "--quiet",
                str(input_path),
            ]

    go_executable = shutil.which("go")
    if go_executable is None:
        go_executable = "go"

    return [
        go_executable,
        "run",
        "./cmd/opendataloader-pdf",
        "--output-dir",
        str(output_dir),
        "--format",
        "markdown",
        "--table-method",
        "cluster",
        "--image-output",
        "off",
        "--quiet",
        str(input_path),
    ]


def to_markdown(_, input_path, output_dir):
    """Convert PDF to Markdown using the Go CLI."""
    command = _build_command(Path(input_path), Path(output_dir))

    result = subprocess.run(
        command,
        capture_output=True,
        text=True,
        cwd=_project_root(),
    )

    if result.returncode != 0:
        print(f"Error converting {input_path} with Go CLI:", file=sys.stderr)
        if result.stderr:
            print(result.stderr, file=sys.stderr)
        elif result.stdout:
            print(result.stdout, file=sys.stderr)
        # Don't raise - continue with other files
