"""Low-level Go binary runner for opendataloader-pdf."""
import os
import locale
import platform
import shutil
import subprocess
import sys
from typing import List


def _get_binary_path() -> str:
    system = platform.system()
    binary_name = "opendataloader-pdf.exe" if system == "Windows" else "opendataloader-pdf"

    env_path = os.environ.get("OPENDATALOADER_PDF_BIN")
    if env_path and os.path.isfile(env_path):
        return env_path

    package_bin = os.path.join(os.path.dirname(__file__), "bin", binary_name)
    if os.path.isfile(package_bin):
        return package_bin

    path_binary = shutil.which("opendataloader-pdf")
    if path_binary:
        return path_binary

    raise RuntimeError(
        "opendataloader-pdf binary not found. "
        "Set OPENDATALOADER_PDF_BIN environment variable or install the Go binary."
    )


def _build_command(args: List[str]) -> List[str]:
    """Build the subprocess command for the Go binary."""
    return [_get_binary_path(), *args]


def run_jar(args: List[str], quiet: bool = False) -> str:
    """Backward-compatible alias for running the Go binary."""
    try:
        command = _build_command(args)

        if quiet:
            result = subprocess.run(
                command,
                capture_output=True,
                text=True,
                check=True,
                encoding=locale.getpreferredencoding(False),
            )
            return result.stdout

        with subprocess.Popen(
            command,
            stdout=subprocess.PIPE,
            stderr=subprocess.STDOUT,
            text=True,
            encoding=locale.getpreferredencoding(False),
        ) as process:
            output_lines: List[str] = []
            for line in process.stdout:
                sys.stdout.write(line)
                output_lines.append(line)

            return_code = process.wait()
            captured_output = "".join(output_lines)

            if return_code:
                raise subprocess.CalledProcessError(
                    return_code, command, output=captured_output
                )
            return captured_output

    except subprocess.CalledProcessError as error:
        print("Error running opendataloader-pdf CLI.", file=sys.stderr)
        print(f"Return code: {error.returncode}", file=sys.stderr)
        if error.output:
            print(f"Output: {error.output}", file=sys.stderr)
        if error.stderr:
            print(f"Stderr: {error.stderr}", file=sys.stderr)
        if error.stdout:
            print(f"Stdout: {error.stdout}", file=sys.stderr)
        raise
