#!/usr/bin/env python3

from subprocess import PIPE, run
from argparse import ArgumentParser
from sys import stderr
from json import loads, dump

parser = ArgumentParser(description="Dumps path relations of a Nix closure")
parser.add_argument('-i', type=str, help="Nix flake ref or nix store path", required=True)
parser.add_argument('-o', type=str, help="Where to save the generated HTML file", required=True)
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

with open(args.o, 'w') as w:
    with open("./web/dist/index.html", 'r') as r:
        while True:
            chunk = r.read(128*1024)
            if not chunk:
                break
            print("chunk")
            w.write(chunk)
    print("<script>window.setData(", file=w)
    dump(dict(links=links,backlinks=backlinks,paths=paths), w)
    print(")</script>", file=w)


# print(result_to_process[0])
