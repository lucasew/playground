#!/usr/bin/env python3
import itertools

def fib():
    a = 0
    b = 1
    while True:
        yield b
        a, b = b, a + b

print(list(itertools.islice(fib(), 20)))
