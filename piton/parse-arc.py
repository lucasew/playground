#!/usr/bin/env python3

# https://wiki.haskell.org/FreeArc/Universal_Archive_Format

from pathlib import Path
import mmap
import lzma
from argparse import ArgumentParser
import ctypes
import enum
import struct

parser =ArgumentParser()
parser.add_argument("file", type=Path)

args = parser.parse_args()

ARC_FILE: Path = args.file
ARC_MAGIC = bytes([65,114,67,1])

ARC_MAGIC_REVERSED = list(ARC_MAGIC)
ARC_MAGIC_REVERSED.reverse()
ARC_MAGIC_REVERSED = bytes(ARC_MAGIC_REVERSED)

file_stat = ARC_FILE.stat() 

class BLOCKTYPE(enum.IntEnum):
    DESCR_BLOCK = 0
    HEADER_BLOCK = 1
    DATA_BLOCK = 2
    DIR_BLOCK = 3
    FOOTER_BLOCK = 4
    RECOVERY_BLOCK = 5

class BLOCK(ctypes.Structure):
    _fields_ = [
        ("type", ctypes.c_int),
        ("compressor", ctypes.c_char_p),
        ("pos", ctypes.c_int), # https://pubs.opengroup.org/onlinepubs/9699919799/basedefs/sys_types.h.html#tag_13_67 off_t shall be signed integer type
        ("origsize", ctypes.c_int),
        ("compsize", ctypes.c_int),
    ]


with ARC_FILE.open('rb') as f:
    f = mmap.mmap(f.fileno(), 0, access=mmap.ACCESS_READ)
    assert f[0:4] == ARC_MAGIC, "Not an arc file"
    version = list(list(f[4:8]))
    print(version)
    last_block = f[-4096:]

    last_signature = list(last_block).copy()
    last_signature.reverse()
    last_signature = bytes(last_signature)
    last_signature = 4096 - 4 - last_signature.index(ARC_MAGIC_REVERSED)

    block_descriptor = last_block[last_signature + 4:]

    unpacked = struct.unpack("ipiii", last_block[last_signature+4:][:20])
    print('unpacked', unpacked)
    # last_block_struct = BLOCK.from_buffer_copy(last_block, last_signature + 4)
    # print(last_block_struct.type, last_block_struct.compressor, last_block_struct.pos, last_block_struct.origsize, last_block_struct.compsize)
    # print('block_descriptor_dec', lzma.decompress(last_block[:last_signature]))
    print('block_descriptor', block_descriptor, [hex(b) for b in block_descriptor])
    

    print(last_signature, last_block[last_signature:last_signature+4])

