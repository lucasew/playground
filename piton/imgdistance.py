#!/usr/bin/env nix-shell
#! nix-shell -i python3 -p python3Packages.numpy python3Packages.opencv3 python3Packages.tqdm python3Packages.numba

from argparse import ArgumentParser
from pathlib import Path
import sqlite3
from concurrent.futures import ThreadPoolExecutor
from itertools import islice
from json import dumps

parser = ArgumentParser()
parser.add_argument('input', type=Path)
parser.add_argument('output', type=Path)
parser.add_argument('-j', dest='num_workers', type=int, default=2)

args = parser.parse_args()

import numpy as np
import cv2
from tqdm import tqdm
from numba import jit

def load_image(path):
  img = cv2.imread(str(path))
  if img is None:
    return None
  return cv2.cvtColor(img, cv2.COLOR_BGR2RGB)

@jit
def process_image(image, ksize=32):
  kernel = np.ones((ksize, ksize), np.float32)/(ksize**2)
  resized = cv2.resize(image, (256, 256))
  return cv2.filter2D(resized, -1, kernel) / 255

@jit(parallel=True)
def diff_images(img1, img2):
  # assert img1.dtype == img2.dtype
  subpix = np.abs(img1 - img2)
  avg = np.zeros((16,16,3), img1.dtype)
  std = np.zeros((16,16,3), img1.dtype)
  for i in range(16):
    for j in range(16):
      for k in range(3):
        patch = subpix[i*16:(i+1)*16, j*16:(j+1)*16,k].ravel()
        avg[i,j,k] = np.average(patch)
        std[i,j,k] = np.std(patch)
        # del patch
  return np.sum(np.array(avg**2).ravel()) + np.sum(np.array(std**2).ravel())
  # del subpix
  # del avg
  # del std

def product2diag(range_max):
  for i in range(range_max):
    for j in range(range_max):
      if i >= j:
        yield (i, j)

def batched(iterable, n):
    "Batch data into tuples of length n. The last batch may be shorter."
    # batched('ABCDEFG', 3) --> ABC DEF G
    if n < 1:
        raise ValueError('n must be at least one')
    it = iter(iterable)
    while (batch := tuple(islice(it, n))):
        yield batch


if __name__ == '__main__':
  if not args.input.is_dir() and not args.output.is_dir():
    img_in = load_image(str(args.input))
    img_out = load_image(str(args.output))
    print(diff_images(img_in, img_out))
  else:
    items=[]
    folder_items = list(args.input.iterdir())
    loaded_images = 0
    with ThreadPoolExecutor(max_workers=args.num_workers) as e:
      def transform_image(item):
        if item.is_file():
          try:
            loaded = load_image(item)
            if loaded is None:
              return None
            processed = process_image(loaded)
            return dict(file=str(item), image=processed)
          except Exception as e:
            print(item, e)
            return None

      for item in tqdm(e.map(transform_image, folder_items, chunksize=128), total=len(folder_items), desc="Loading and preprocessing images..."):
        if item is not None:
          loaded_images += 1
          items.append(item)

      with args.output.open('w') as f:
        print('a,b,distance', file=f)
        items_prod = product2diag(loaded_images)
        total = int(((loaded_images**2) / 2) + (loaded_images / 2))

        def process_item(item):
          (from_item, to_item) = item
          from_item = items[from_item]
          to_item = items[to_item]
          return dict(
            a=from_item['file'],
            b=to_item['file'],
            distance=diff_images(
              from_item['image'],
              to_item['image']
            )
          )
        @jit(parallel=True)
        def process_batch(batch):
          for item in batch:
            yield process_item(item)
        ops = tqdm(total=total, desc="Processing distances")
        for batch in batched(items_prod, 512):
          # for item in e.map(process_item, batch, chunksize=256):
          for item in process_batch(batch):
            ops.update(1)
            if item is not None:
              print(",".join(
                (
                  str(item['a']),
                  str(item['b']),
                  str(item['distance'])
                )
              ), file=f)
              print(",".join(
                (
                  str(item['b']),
                  str(item['a']),
                  str(item['distance'])
                )
              ), file=f)
