#!/usr/bin/env python3
"""
# O problema da mochila (Knapsack)

## Overview
Dado um conjunto de coisas o objetivo é colocar elas em uma mochila hipotética de capacidade limitada.

A questão é que as coisas tem um valor, logo o objetivo é otimizar a capacidade da mochila levando quanto
mais valor melhor, o ideal é achar um ponto ótimo.

## Onde é usado?
É um problema relativamente comum em programação linear, é um dos motivos de existir algoritmos como o simplex.

## Modelagem
A abordagem aqui é a de selecionar transações de dada criptomoeda __proof of work__ a fim de maximizar os ganhos
com taxas de transação na mineração de um bloco.

Um bloco tem um tamanho máximo que é ocupado por transações, as transações tem cada uma um tamanho em bytes
e um valor de taxa.

Um bloco tem um custo constante de mineração, a questão é fazer esse custo valer mais a pena, o algoritmo recebe um
conjunto de transações e monta o bloco buscando a melhor taxa.

Pode ser reduzido a um problema de ordenação por custo/beneficio permitindo inclusão até preencher o espaço disponível.

Como aqui latência não é algo crítico é possível ordenar as transações por custo/beneficio e inserir conforme possível
na lista de transações que será processada até fechar a quantidade, como não é possível passar do tamanho o algoritmo
tenta otimizar o pouco espaço restante

Também é possível usar uma estratégia iterativa, ou até uma dividir para conquistar que usa n workers para buscar as melhores transaçoes em um pedaço de um conjunto de transações mas nesse caso vamos com o naive approach

Esse algoritmo é NP completo pois reduz com a seleção (P, O(n¹)) para ordenação (NP, O(n log(n))).

"""

from dataclasses import dataclass

@dataclass
class Transaction():
    size: int
    tax: float

def mine(block_size: int, txs: list[Transaction] = []):
    s = [*txs]
    s.sort(key = lambda t: float(t.size) / t.tax)
    ret = []
    acumulado = 0
    for item in s:
        if acumulado + item.size > block_size:
            if acumulado >= block_size:
                break
            continue
        ret.append(item)
        acumulado += item.size
    return ret

txs = [
    Transaction(3, 0.1),
    Transaction(4, 0.1),
    Transaction(4, 0.2),
    Transaction(3, 0.1),
    Transaction(1, 0.1),
    Transaction(1, 0.4),
]

block_size = 5

print(mine(block_size, txs))
