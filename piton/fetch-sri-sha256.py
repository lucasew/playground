#!/usr/bin/env nix-shell
#!nix-shell -i python -p python3

from urllib import request
from argparse import ArgumentParser
import hashlib
import base64

parser = ArgumentParser()
parser.add_argument('url')

args = parser.parse_args()

response = request.urlopen(args.url)
hasher = hashlib.sha256()

while True:
    buf = response.read(16*1024)
    if not buf:
        break
    hasher.update(buf)

hash_bytes = hasher.digest()
b64_hash = base64.b64encode(hash_bytes).decode('ascii')
print(f"sha256-{b64_hash}")
