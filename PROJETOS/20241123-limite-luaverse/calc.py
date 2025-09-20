#!/usr/bin/env python3
from fractions import Fraction as F

def f(x):
    return ((F(7)+(F(x)**F(1/3) - F(3)))**F(1/2) - F(3))/(F(x) - F(8))

diff = F(1)
base = F(8)
while diff > 0.00000000001:
    sup = f(base + diff)
    sub = f(base - diff)

    print('diff=', diff, sup, sub, )
    diff = diff / 10
