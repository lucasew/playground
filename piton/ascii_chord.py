#!/usr/bin/env python3
from argparse import ArgumentParser, ArgumentDefaultsHelpFormatter
from urllib.request import urlretrieve
import tempfile
import sys
from pathlib import Path
from collections import defaultdict
import itertools
import json
import os

def hash_string(s):
    import hashlib
    if isinstance(s, str):
        s = s.encode('utf-8')
    h = hashlib.md5(s)
    return h.hexdigest()


def batched(iterable, n):
    # batched('ABCDEFG', 3) → ABC DEF G
    if n < 1:
        raise ValueError('n must be at least one')
    iterator = iter(iterable)
    while batch := tuple(itertools.islice(iterator, n)):
        yield batch

parser = ArgumentParser(formatter_class=ArgumentDefaultsHelpFormatter)

parser.add_argument("chord", nargs='+')
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

class Chord():
    def __init__(self, title="", **kwargs):
        assert 'frets' in kwargs and len(kwargs['frets']) == 6, "Faltando frets"
        assert 'fingers' in kwargs and len(kwargs['fingers']) == 6, "Faltando frets"

        baseFret = kwargs.get('baseFret', 1)
        assert isinstance(baseFret, int), 'baseFret deve ser inteiro'
        self.baseFret = baseFret

        self.frets = kwargs['frets']
        self.fingers = kwargs['fingers']
        self.title = title

        self.capo = 0
        if kwargs.get('capo'):
            self.capo = kwargs.get('barres', [0])[0]


    @property
    def suggested_capotraste(self):
        interest_frets = [ f  for f in self.frets if f > 0]
        min_fret = min(*interest_frets)
        return min_fret + self.baseFret

    def __repr__(self):
        items = defaultdict(lambda: " ")
        for i in range(6):
            items[i, self.capo + self.baseFret] = '='
            items[i, 1] = '-'
            if self.frets[i] < 0:
                for j in range(0, self.baseFret + 5):
                    items[i, j] = 'x'
                continue
            items[i, self.baseFret + self.frets[i]] = self.fingers[i]
        lines = []
        for i in range(self.baseFret, self.baseFret + 5):
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

        lines.append("{:^%is}".replace('%i', str(len(lines[0]))).format(self.title))
        return "\n".join(lines)

class ChordList():
    def __init__(self, *chords, title=""):
        # print(chords)
        self.chords = []
        for chord in chords:
            if isinstance(chord, dict):
                chord = Chord(**chord, title=title)
            assert isinstance(chord, Chord)
            self.chords.append(chord)
    def __repr__(self):
        blocks = [str(c).split('\n') for c in self.chords]
        block_column_size = len(blocks[0][0]) + 2
        block_amount = os.get_terminal_size().columns // block_column_size
        lines = []
        for line in list(batched(blocks, block_amount)):
            for items in zip(*line):
                lines.append("  ".join(items))
        return '\n'.join(lines)

chords = get_chords()

for selected_chord in args.chord:
    chord = chords['chords'].get(selected_chord)
    if chord is None:
        print("No such chord, available: ", chords['chords'].keys())
        continue
    print(ChordList(*chord['positions']))

