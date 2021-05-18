#! /usr/bin/env nix-shell
#! nix-shell -p python3 -i python

QUADROS = int(input("Quantidade de quadros: "))

DEFAULT_VALIDATOR = lambda: True

class Algorithm():
    def __init__(self, items, quadros = []):
        self.faults = 0
        self.items = items
        self.quadros = quadros

    def next_item(self, i):
        print("stub")
        raise UnimplementedError()
        return 0

    def incr(self):
        self.faults += 1

    def run(self):
        for i in range(0, len(self.items)):
            is_hit = self.next_item(i)
            if is_hit:
                print(f"{self.items[i]} - HIT => Q = {self.quadros}")
            else:
                print(f"{self.items[i]} - MISS => Q = {self.quadros}")
        return self.faults


class AlgoOPT(Algorithm):
    def next_item(self, i):
        if len(self.quadros) < QUADROS:
            self.quadros.append(items[i])
            return False
        if self.items[i] in self.quadros:
            found_index = self.quadros.index(items[i])
            self.quadros.remove(self.items[i])
            self.quadros.append(self.items[i]) # bota no final
            return True
        MAX_DISTANCE = 9999
        distance = [MAX_DISTANCE]*len(self.quadros)
        for p in range(i, len(self.items)):
            item = self.items[p]
            for q in range(0, len(self.quadros)):
                if item == self.quadros[q] and distance[q] == MAX_DISTANCE:
                    distance[q] = p - i
        biggest_distance_index = 0
        for p in range(0, len(self.quadros)):
            if distance[p] > distance[biggest_distance_index]:
                biggest_distance_index = p
        # print(f"distances = {distance}, biggest_distance = {distance[biggest_distance_index]} ({biggest_distance_index})")
        self.quadros[biggest_distance_index] = self.items[i]
        self.quadros.remove(self.items[i])
        self.quadros.append(items[i]) # botar no final
        self.incr()
        return False
    
class AlgoFIFO(Algorithm):
    def next_item(self, i):
        hit = self.items[i] in self.quadros
        self.quadros.append(self.items[i])
        self.quadros = self.quadros[-QUADROS:]
        if not hit:
            self.incr()
        return hit

class AlgoLRU(Algorithm):
    def next_item(self, i):
        item = self.items[i]
        if self.items[i] in self.quadros:
            idx = self.quadros.index(item)
            self.quadros.remove(item)
            self.quadros.append(item)
            return True
        else:
            self.quadros.append(item)
            self.quadros = self.quadros[-QUADROS:]
            self.incr()
            return False

algorithms = {
    "OPT": AlgoOPT,
    "FIFO": AlgoFIFO,
    "LRU": AlgoLRU
}

def once_asker(query, validator = DEFAULT_VALIDATOR, breakable = False):
    while True:
        value = input(query)
        if breakable and len(value) == 0:
            return None
        try:
            validator_result = validator(value)
            if validator_result is False or validator_result is None:
                raise Exception()
            else:
                return value
        except Exception as e:
            print(e)
            print("ERRO: Valor inválido")


def continuous_asker(query, validator = DEFAULT_VALIDATOR):
    while True:
        value = once_asker(query, validator, True)
        if value is None:
            break
        yield value

def is_inside_list(l):
    l = list(l)
    def ret(value):
        return value in list(l)
    return ret

algorithm_name = once_asker("Selecione algoritmo (OPT, FIFO, LRU): ", validator = is_inside_list(algorithms.keys()))
algorithm = algorithms[algorithm_name]

ACESSOS = []
for acesso in continuous_asker("Referencia (numero, vazio termina): ", validator = int):
    item = int(acesso)
    ACESSOS.append(item)

page_faults = algorithm(ACESSOS).run()
print(f"Page faults: {page_faults}")

# print(QUADROS)
# print(ACESSOS)
