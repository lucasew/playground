#!/usr/bin/env python3
"""Pack the boot sector, stage2, and kernel into one bootable disk image."""

import struct
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
BUILD = ROOT / "build"


def main():
    boot = (BUILD / "boot.bin").read_bytes()
    stage2 = bytearray((BUILD / "stage2.bin").read_bytes())
    kernel = (BUILD / "kernel.bin").read_bytes()
    if len(boot) != 512 or boot[510:512] != b"\x55\xaa":
        raise SystemExit(f"boot sector is {len(boot)} bytes, expected 512 with 0xAA55")
    if len(stage2) != 8192:
        raise SystemExit(f"stage2 is {len(stage2)} bytes, expected 8192")
    sectors = (len(kernel) + 511) // 512
    struct.pack_into("<II", stage2, 8192 - 8, 17, sectors)
    padded = kernel + bytes(sectors * 512 - len(kernel))
    image = boot + bytes(stage2) + padded
    out = BUILD / "chimia.img"
    out.write_bytes(image)
    print(f"wrote {out} ({len(image)} bytes, kernel {len(kernel)} bytes, {sectors} sectors)")


if __name__ == "__main__":
    sys.exit(main())
