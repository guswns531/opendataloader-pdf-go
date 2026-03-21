"""Integration tests that run the packaged CLI or Go override (slow)."""

import subprocess
from pathlib import Path

import opendataloader_pdf


def test_convert_generates_output(input_pdf, output_dir, monkeypatch):
    """Verify that convert() actually generates output files"""
    root_dir = Path(__file__).resolve().parents[3]
    go_bin = output_dir / "opendataloader-pdf"
    subprocess.run(
        ["go", "build", "-o", str(go_bin), "./cmd/opendataloader-pdf"],
        cwd=root_dir,
        check=True,
    )
    monkeypatch.setenv("OPENDATALOADER_PDF_CLI_BIN", str(go_bin))

    opendataloader_pdf.convert(
        input_path=str(input_pdf),
        output_dir=str(output_dir),
        format="json",
        quiet=True,
    )
    output = output_dir / "1901.03003.json"
    assert output.exists(), f"Output file not found at {output}"
    assert output.stat().st_size > 0, "Output file is empty"


def test_legacy_run_generates_markdown_with_go_override(input_pdf, output_dir, monkeypatch):
    root_dir = Path(__file__).resolve().parents[3]
    go_bin = output_dir / "opendataloader-pdf"
    subprocess.run(
        ["go", "build", "-o", str(go_bin), "./cmd/opendataloader-pdf"],
        cwd=root_dir,
        check=True,
    )
    monkeypatch.setenv("OPENDATALOADER_PDF_CLI_BIN", str(go_bin))

    opendataloader_pdf.run(
        input_path=str(input_pdf),
        output_folder=str(output_dir),
        generate_markdown=True,
        no_json=True,
        debug=False,
    )

    output = output_dir / "1901.03003.md"
    assert output.exists(), f"Output file not found at {output}"
    assert output.stat().st_size > 0, "Output file is empty"
