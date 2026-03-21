"""
Low-level JAR runner for opendataloader-pdf.
"""
import os
import locale
import subprocess
import sys
import importlib.resources as resources
from typing import List

# The consistent name of the JAR file bundled with the package
_JAR_NAME = "opendataloader-pdf-cli.jar"


def run_jar(args: List[str], quiet: bool = False) -> str:
    """Run the opendataloader-pdf JAR with the given arguments."""
    try:
        binary_override = os.environ.get("OPENDATALOADER_PDF_CLI_BIN") or os.environ.get("OPENDATALOADER_GO_BIN")
        if binary_override:
            command = [binary_override, *args]
        else:
        # Access the embedded JAR inside the package
            jar_ref = resources.files("opendataloader_pdf").joinpath("jar", _JAR_NAME)
            with resources.as_file(jar_ref) as jar_path:
                command = ["java", "-jar", str(jar_path), *args]

        if quiet:
            # Quiet mode → capture all output
            result = subprocess.run(
                command,
                capture_output=True,
                text=True,
                check=True,
                encoding=locale.getpreferredencoding(False),
            )
            return result.stdout

        # Streaming mode → live output
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

    except FileNotFoundError:
        print(
            "Error: CLI command not found. Please ensure Java or the override binary is installed and in your PATH.",
            file=sys.stderr,
        )
        raise

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
