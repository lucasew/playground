#!/usr/bin/env nix-shell
#! nix-shell -i python -p python3Packages.matplotlib python3Packages.numpy python3Packages.scipy
from matplotlib import pyplot as plt
import numpy as np
from scipy.stats import gaussian_kde
from time import sleep

mu, sigma = 0, 0.1 
x = np.random.normal(mu, sigma, 1000)
y = np.random.normal(mu, sigma, 1000)
z = np.random.normal(mu, sigma, 1000)
xyz = np.vstack([x,y,z])
kde = gaussian_kde(xyz)
density = kde(xyz)
# print(density)

ax = plt.axes(projection='3d')
ax.scatter(x, y, z, s=2, c=density/ density.max())
plt.show()

while True:
    sleep(1)
