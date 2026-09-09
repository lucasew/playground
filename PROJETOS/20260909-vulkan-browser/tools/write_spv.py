#!/usr/bin/env python3
"""Write a tiny SPIR-V blob whose OpEntryPoint execution model the shim can read.

The original main.c mmaps vert.spv / frag.spv and hands the bytes to
vkCreateShaderModule. The WebGPU shim does not run the SPIR-V; it only
inspects the execution model (0 = Vertex, 4 = Fragment) and then uses
WGSL that matches shader.vert / shader.frag.
"""

import struct
import sys


def write_spv(path: str, execution_model: int) -> None:
    words = [
        0x07230203,  # magic
        0x00010000,  # version 1.0
        0x00000000,  # generator
        4,  # bound
        0,  # schema
        (2 << 16) | 17,  # OpCapability Shader
        1,
        (3 << 16) | 14,  # OpMemoryModel Logical GLSL450
        0,
        1,
        (5 << 16) | 15,  # OpEntryPoint model %1 "main"
        execution_model,
        1,
        0x6E69616D,  # "main"
        0x00000000,
    ]
    with open(path, "wb") as fh:
        fh.write(struct.pack("<" + "I" * len(words), *words))


if __name__ == "__main__":
    if len(sys.argv) != 3 or sys.argv[1] not in {"vert", "frag"}:
        sys.stderr.write("usage: write_spv.py vert|frag OUT.spv\n")
        sys.exit(2)
    write_spv(sys.argv[2], 0 if sys.argv[1] == "vert" else 4)
