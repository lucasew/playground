#!/usr/bin/env python3

from subprocess import PIPE, run
from argparse import ArgumentParser
from sys import stderr
from json import loads, dumps

parser = ArgumentParser(description="Dumps path relations of a Nix closure")
parser.add_argument('-i', type=str, help="Nix flake ref or nix store path", required=True)
args = parser.parse_args()

# proc = run(["cat", "/home/lucasew/TMP2/sysgraph.json"], stdout=PIPE)
# proc = run(["cat", "/home/lucasew/TMP2/sysgraph.json"], capture_output=True, shell=True, check=True)
proc = run(["nix", "path-info", "-Sh", args.i, "--json", "--recursive"], stdout=PIPE, stderr=stderr, check=True)
items = loads(proc.stdout)
links = {}
backlinks = {}
paths = {}
for item in items:
    path = item['path']
    paths[path] = {
        'nar': item['narSize'],
        'closure': item['closureSize']
    }
    for reference in item['references']:
        if path not in links:
            links[path] = []
        links[path].append(reference)

        if reference not in backlinks:
            backlinks[reference] = []
        backlinks[reference].append(path)

print(dumps(dict(links=links,backlinks=backlinks,paths=paths)))

# print(result_to_process[0])
