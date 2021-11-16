#!/usr/bin/env nix-shell
#! nix-shell -p python3 -i python


from random import random
from pprint import pprint
from itertools import product

def sig2num(sig):
    if sig > 0:
        return 1
    return -1

class Perceptron:
    def __init__(self, 
                 w1 = random(), 
                 w2 = random(), 
                 bias = random(), 
                 learning_rate = 0.2, 
                 activation = sig2num
                ):
        self.weights = [w1, w2]
        self.bias = bias
        self.learning_rate = learning_rate
        self.activation = activation

    def __call__(self, a, b):
        [w1, w2] = self.weights
        return self.activation((a*w1) + (b*w2) + self.bias)
    def train(self, *lines):
        for i in range(len(lines)):
            line = lines[i]
            [a, b, right_result] = line
            predicted_result = self(a, b)
            error = right_result - predicted_result
            for j in range(2):
                delta_w = self.learning_rate*error*line[j]
                self.weights[j] += self.learning_rate * delta_w
                self.bias =+ self.learning_rate * delta_w
        return right_result - predicted_result

def generate_train_data():
    def b2i(b):
        if b:
            return 1
        else:
            return -1
    ret = []
    for item in product([False, True], [False, True]):
        [a, b] = item
        out = not (a and b)
        ret.append([b2i(b) for b in [a, b, out]])
    return ret

def test_data(perceptron):
    data = generate_train_data()
    for item in data:
        [a, b, res] = item
        ret = perceptron(a, b)
        print(f"x0={a} x1={b} res={res} pred={ret} err={res - ret}\
        w1={p.weights[0]} w2={p.weights[1]} bias={p.bias}")

train_data = generate_train_data()

p = Perceptron(learning_rate = .05)
while True:
    p.train(*train_data)
    test_data(p)
    input()


