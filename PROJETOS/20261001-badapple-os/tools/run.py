#!/usr/bin/env python3
"""Boot Chimia on QEMU with the PC speaker attached to the default audio backend."""

import os
import shutil
import subprocess
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
from qemu import qemu_env

ROOT = Path(__file__).resolve().parents[1]
IMAGE = ROOT / "build" / "chimia.img"


def main():
    binary, bios, env = qemu_env()
    display = "gtk" if os.environ.get("DISPLAY") else "sdl"
    if shutil.which("qemu-system-x86_64") is None and not os.environ.get("DISPLAY"):
        display = "gtk"
    command = [binary]
    if bios:
        command += ["-L", bios]
    command += [
        "-m", "128",
        "-no-reboot",
        "-net", "none",
        "-boot", "order=c,menu=off",
        "-drive", f"file={IMAGE},format=raw,if=ide,index=0,media=disk",
        "-serial", "stdio",
        "-audiodev", "sdl,id=speaker",
        "-machine", "pc,pcspk-audiodev=speaker",
        "-display", display,
    ]
    print(" ".join(command))
    return subprocess.call(command, env=env)


if __name__ == "__main__":
    sys.exit(main())
