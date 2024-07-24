#!/usr/bin/env python3
from argparse import ArgumentParser, ArgumentDefaultsHelpFormatter
from urllib.request import urlretrieve
import tempfile
import sys
from pathlib import Path
from collections import defaultdict
import json

def hash_string(s):
    import hashlib
    if isinstance(s, str):
        s = s.encode('utf-8')
    h = hashlib.md5(s)
    return h.hexdigest()

parser = ArgumentParser(formatter_class=ArgumentDefaultsHelpFormatter)

parser.add_argument("chord")
parser.add_argument('--chords-url', default="https://raw.githubusercontent.com/tombatossals/chords-db/master/lib/guitar.json")

args = parser.parse_args()

def get_chords():
    chord_filename = f"chordcache-{hash_string(args.chords_url)}.json"
    chordfile = Path(tempfile.gettempdir()) / chord_filename
    if not chordfile.exists():
        print("Downloading.", file=sys.stderr, end='')
        urlretrieve(args.chords_url, chordfile, reporthook=lambda _a, _b, _c: sys.stderr.write("."))
        sys.stderr.write('\n')
    with chordfile.open('r') as f:
        data = json.load(f)

    chord_specs = {}
    for group_name, group_values in data['chords'].items():
        for group_value in group_values:
            key = group_value.get('key', '')
            suffix = group_value.get('suffix', '')
            chord_specs["".join([key, suffix])] = group_value
    data['chords'] = chord_specs
    return data

def structure_position(position, title=""):
    baseFret = position['baseFret']
    items = defaultdict(lambda: " ")
    for i in range(6):
        items[i, 1] = '-'
        if position['frets'][i] < 0:
            for j in range(0, baseFret + 5):
                items[i, j] = 'x'
            continue
        items[i,baseFret + position['frets'][i]] = position['fingers'][i]

    lines = []
    for i in range(baseFret, baseFret + 5):
    # for i in range(0, baseFret + 5):
        line = [
            str(i).rjust(2, ' '),
            items[0, i],
            items[1, i],
            items[2, i],
            items[3, i],
            items[4, i],
            items[5, i],
        ]
        lines.append(" ".join([str(x) for x in line]))

    lines.append("{:^%is}".replace('%i', str(len(lines[0]))).format(title))
    return "\n".join(lines)
    # print(position)

selected_chord = args.chord

chords = get_chords()

chord = chords['chords'].get(selected_chord)
if chord is None:
    print("No such chord, available: ", chords['chords'].keys())
    exit(1)

for position in chord['positions']:
    print(structure_position(position, title=selected_chord))
    print()
    # print(selected_chord)


