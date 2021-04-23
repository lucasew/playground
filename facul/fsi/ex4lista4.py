#! /usr/bin/env nix-shell
#! nix-shell -p python3 -i python
# vim:ft=python

from random import randint
from time import sleep

INITIAL_GENERATION = [
    0b001100, 
    0b010101,
    0b111000,
    0b000111,
    0b101011, 
    0b101000
]

def inverseBit(b):
    if b == "1":
        return "0"
    if b == "0":
        return "1"

def bin2str(b):
    return str(bin(b))[2:]

def mutate(ind):
    str_ind = list(bin2str(ind))
    idx = randint(0, len(str_ind) - 1)
    str_ind[idx] = inverseBit(str_ind[idx])
    return eval(f"0b{''.join(str_ind)}")

def fit(ind):
    return ind**2

if __name__ == "__main__":
    generation = INITIAL_GENERATION
    while True:
        generation = sorted(generation, key=fit, reverse=True) # greater first
        print(f"Best of the generation: {generation[0]}. Fit: {fit(generation[0])}")
        new_generation = []
        new_generation.append(generation[0])
        new_generation.append(generation[1])
        new_generation.append(mutate(generation[0]))
        new_generation.append(mutate(generation[0]))
        new_generation.append(mutate(generation[1]))
        new_generation.append(mutate(generation[1]))
        generation = new_generation
        sleep(1)
