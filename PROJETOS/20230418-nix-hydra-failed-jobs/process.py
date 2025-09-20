#!/usr/bin/env python3
from argparse import ArgumentParser
from pathlib import Path
from collections import defaultdict
from urllib.request import urlopen
from pprint import pprint
import logging
import re

logging.basicConfig()
logger = logging.getLogger(__name__)
logger.setLevel(logging.DEBUG)

parser = ArgumentParser()
parser.add_argument(
    '-i,--input',
    dest="input",
    help="File with data downloaded from 'https://hydra.nixos.org/jobset/nixpkgs/trunk/jobs-tab?filter=%'. Will fetch manually if not defined",
    type=Path,
)
parser.add_argument(
    '-j,--job',
    dest="job",
    help="Job to get data from hydra",
    default="trunk"
)

args = parser.parse_args()
print(args)


regex = r"https:\/\/hydra.nixos.org\/job\/nixpkgs\/[^\/]*\/([^\"]*)"


logger.info(f"Fetching data from 'https://hydra.nixos.org/jobset/nixpkgs/{args.job}/jobs-tab?filter=%'")

data = args.input.open('rb') if args.input is not None else urlopen(f'https://hydra.nixos.org/jobset/nixpkgs/{args.job}/jobs-tab?filter=%')

logger.info("Parsing fetched data")

items = []

for line in data:
    line = line.decode('utf-8')
    if not 'title="Failed"' in line:
        continue
    line = line.strip()
    item = re.search(regex, line)
    if item is None:
        continue
    item = item.group(1)
    items.append(item)

for item in items:
    print(item)
