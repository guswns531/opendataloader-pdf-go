"""Custom build hook for hatch to copy the Go binary and license files."""

import os
import platform
import shutil
from pathlib import Path

from hatchling.builders.hooks.plugin.interface import BuildHookInterface


class CustomBuildHook(BuildHookInterface):
    def initialize(self, version, build_data):
        root_dir = Path(self.root)
        pkg_dir = root_dir / "src/opendataloader_pdf"
        binary_name = (
            "opendataloader-pdf.exe" if platform.system() == "Windows" else "opendataloader-pdf"
        )
        dest_bin_dir = pkg_dir / "bin"
        dest_bin_path = dest_bin_dir / binary_name
        license_path = pkg_dir / "LICENSE"
        notice_path = pkg_dir / "NOTICE"
        third_party_dest = pkg_dir / "THIRD_PARTY"

        readme_path = root_dir / "README.md"

        # Check if all required files already exist (building from sdist)
        if (
            dest_bin_path.exists()
            and license_path.exists()
            and notice_path.exists()
            and third_party_dest.exists()
            and readme_path.exists()
        ):
            print("All required files already exist (building from sdist), skipping copy")
            return

        # --- Copy Go binary ---
        print(f"Root DIR: {root_dir}")
        env_binary = os.environ.get("OPENDATALOADER_PDF_BIN")
        candidate_paths = [
            Path(env_binary) if env_binary else None,
            (root_dir / "../../bin" / binary_name),
            (root_dir / "../../go" / binary_name),
        ]
        source_binary_path = None
        for candidate in candidate_paths:
            if candidate and candidate.exists() and candidate.is_file():
                source_binary_path = candidate.resolve()
                break
        if source_binary_path is None:
            searched = [str(p.resolve()) for p in candidate_paths if p is not None]
            raise RuntimeError(
                "Could not find the opendataloader-pdf binary. "
                "Set OPENDATALOADER_PDF_BIN or run "
                "`cd go && go build -o ../bin/opendataloader-pdf ./cmd/opendataloader-pdf/` first. "
                f"Searched: {searched}"
            )

        print(f"Found source binary: {source_binary_path}")

        dest_bin_dir.mkdir(parents=True, exist_ok=True)
        print(f"Copying binary to {dest_bin_path}")
        shutil.copy(source_binary_path, dest_bin_path)
        dest_bin_path.chmod(0o755)

        # --- Copy LICENSE, NOTICE, README ---
        shutil.copy(root_dir / "../../LICENSE", license_path)
        shutil.copy(root_dir / "../../NOTICE", notice_path)
        shutil.copy(root_dir / "../../README.md", readme_path)
        third_party_src = root_dir / "../../THIRD_PARTY"
        print(f"Copying THIRD_PARTY directory to {third_party_dest}")
        if third_party_dest.exists():
            shutil.rmtree(third_party_dest)
        shutil.copytree(third_party_src, third_party_dest)
