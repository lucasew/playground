#!/usr/bin/env python3
from argparse import ArgumentParser
from pathlib import Path
from collections import defaultdict
import re

parser = ArgumentParser()

parser.add_argument('input', type=Path)

regex = r"https:\/\/hydra.nixos.org\/job\/nixpkgs\/trunk\/([^\"]*)"

args = parser.parse_args()
print(parser)

packages = defaultdict(lambda: [])


for line in args.input.open('r'):
    if not 'title="Failed"' in line:
        continue
    line = line.strip()
    item = re.search(regex, line)
    if item is None:
        continue
    item = item.group(1)
    parts = item.split('.')
    arch = parts.pop()
    packages[".".join(parts)].append(arch)

for package, archs in packages.items():
    print(package, archs)

