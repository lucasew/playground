#! /usr/bin/env nix-shell
#! nix-shell -p python3 -i python
# vim:ft=python

from random import randint
from time import sleep
from copy import deepcopy

INITIAL_GENERATION = [
    [1, 2, 3, 4, 5],
    [2, 1, 3, 4, 5]
]

custos = {}
custos[1, 1] = 0
custos[1, 2] = 2
custos[1, 3] = 9
custos[1, 4] = 3
custos[1, 5] = 6
custos[2, 1] = 2
custos[2, 2] = 0
custos[2, 3] = 4
custos[2, 4] = 3
custos[2, 5] = 8
custos[3, 1] = 9
custos[3, 2] = 4
custos[3, 3] = 0
custos[3, 4] = 7
custos[3, 5] = 3
custos[4, 1] = 3
custos[4, 2] = 3
custos[4, 3] = 7
custos[4, 4] = 0
custos[4, 5] = 3
custos[5, 1] = 6
custos[5, 2] = 8
custos[5, 3] = 3
custos[5, 4] = 3
custos[5, 5] = 0

def mutate(ind):
    new_ind = deepcopy(ind)
    xa = randint(0, len(ind) - 1)
    xb = randint(0, len(ind) - 1)
    new_ind[xa], new_ind[xb] = new_ind[xb], new_ind[xa]
    return new_ind

def fit(ind):
    a, b, c, d, e = ind
    return custos[a, b] + custos[b, c] + custos[c, d] + custos[d, e]

if __name__ == "__main__":
    generation = INITIAL_GENERATION
    while True:
        generation = sorted(generation, key=fit, reverse=True) # greater first
        print(f"Best of the generation: {generation[0]}, {generation[1]}. Fit: {fit(generation[0])}, {fit(generation[1])}")
        new_generation = []
        new_generation.append(generation[0])
        new_generation.append(generation[1])
        new_generation.append(mutate(generation[0]))
        new_generation.append(mutate(mutate(generation[0])))
        new_generation.append(mutate(generation[1]))
        new_generation.append(mutate(mutate(generation[1])))
        generation = new_generation
        sleep(1)
