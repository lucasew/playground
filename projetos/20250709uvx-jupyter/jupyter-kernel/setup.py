#!/usr/bin/env python3

import argparse
import json
import shutil
from pathlib import Path


UV = shutil.which('uv')
assert UV is not None, "uv not found in PATH"
UV_DIR = str(Path(UV).parent)


def main():
    parser = argparse.ArgumentParser(description="Setup Jupyter kernels for uv")
    parser.add_argument(
        "--versions", 
        nargs="+", 
        default=["3.13", "3.12"],
        help="Python versions to configure (default: 3.13 3.12)"
    )
    
    args = parser.parse_args()
    
    kernel_base = Path.home() / ".local" / "share" / "jupyter" / "kernels"
    
    for version in args.versions:
        kernel_file = kernel_base / f"uv-{version}" / "kernel.json"
        
        kernel_file.parent.mkdir(parents=True, exist_ok=True)
        
        kernel_config = {
            "env": {
                "PATH": "${PATH}:" + UV_DIR
            },
            "argv": [
                UV,
                "run",
                "--python", version,
                "--with", "ipykernel",
                "--no-project",
                "--isolated",
                "--refresh",
                "python", "-m", "ipykernel_launcher",
                "-f", "{connection_file}"
            ],
            "display_name": f"uv-{version}",
            "language": "python",
            "metadata": {
                "debugger": True
            }
        }
        
        kernel_file.write_text(
            json.dumps(kernel_config, indent=2, ensure_ascii=False),
            encoding='utf-8'
        )
        
        print(f"Kernel configured for Python {version} at: {kernel_file}")


if __name__ == "__main__":
    main()
