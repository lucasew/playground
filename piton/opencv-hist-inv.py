#!/usr/bin/env nix-shell
#!nix-shell -p python3Packages.opencv4 python3Packages.matplotlib -i python

from argparse import ArgumentParser
from pathlib import Path
from matplotlib import pyplot as plt
import cv2

parser = ArgumentParser()
parser.add_argument("-i", metavar = "img", type = Path, help = "Imagem de entrada", required = True)
args = parser.parse_args()

img=cv2.imread(str(args.i))
img=cv2.cvtColor(img, cv2.COLOR_BGR2RGB)
img_inv=cv2.flip(img, 1)

fig, axis = plt.subplots(2, 2)
axis[0][0].imshow(img)
axis[0][1].hist(img.ravel())
axis[1][0].imshow(img_inv)
axis[1][1].hist(img_inv.ravel())

plt.show()
