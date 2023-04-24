#!/usr/bin/env nix-shell
#!nix-shell -i python3 -p python3Packages.opencv3

import cv2
import numpy as np
from argparse import ArgumentParser
from pathlib import Path

parser = ArgumentParser()
parser.add_argument('input', type=Path)

args = parser.parse_args()

img = cv2.imread(str(args.input), 0)

print(np.count_nonzero(img)/np.size(img))
