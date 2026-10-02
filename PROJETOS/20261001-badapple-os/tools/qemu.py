"""Locate the local QEMU build used to boot Chimia."""

import os
import shutil
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]


def qemu_env():
    lib = "/tmp/prefix/usr/lib/x86_64-linux-gnu:/tmp/prefix/lib/x86_64-linux-gnu"
    env = os.environ.copy()
    env["LD_LIBRARY_PATH"] = lib + (
        ":" + env["LD_LIBRARY_PATH"] if env.get("LD_LIBRARY_PATH") else ""
    )
    system = shutil.which("qemu-system-x86_64")
    local = Path("/tmp/prefix/usr/bin/qemu-system-x86_64")
    if system:
        binary = system
        bios = None
    elif local.exists():
        binary = str(local)
        bios = "/tmp/pc-bios" if Path("/tmp/pc-bios/bios-256k.bin").exists() else None
    else:
        raise SystemExit("qemu-system-x86_64 is not installed")
    return binary, bios, env


def qemu_command(image, serial_path, wav_path=None, monitor=None, display="none"):
    binary, bios, env = qemu_env()
    command = [binary]
    if bios:
        command += ["-L", bios]
    command += [
        "-m", "128",
        "-no-reboot",
        "-net", "none",
        "-boot", "order=c,menu=off",
        "-drive", f"file={image},format=raw,if=ide,index=0,media=disk",
        "-serial", f"file:{serial_path}",
        "-vga", "std",
        "-display", display,
    ]
    if wav_path:
        command += [
            "-audiodev", f"wav,id=speaker,path={wav_path}",
            "-machine", "pc,pcspk-audiodev=speaker",
        ]
    else:
        command += ["-machine", "pc"]
    if monitor:
        command += ["-monitor", f"unix:{monitor},server,nowait"]
    return command, env
