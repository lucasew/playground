#! /usr/bin/env nix-shell
#! nix-shell -p python3 -i python
from copy import deepcopy

ALPHABET = ["a", "b", "c"]

"""
# Torres de Hanoi
## Abstração da transição
- Primeira letra é a coluna que vai ter o ultimo elemento removido
- Segunda letra é a coluna que vai receber o elemento
"""

# pinos são uma tupla, do menor para o maior e a tupla é de posições
SMALLBAR="_  "
MEDBAR=  "__ "
BIGBAR=  "___"

ENDSTATE="endstate"

START_STATE = [
    [BIGBAR, MEDBAR, SMALLBAR],
    [],
    []
]

visited = {}
transitions = {}

def counter(i = 0):
    class Counter:
        def __init__(self, i = 0):
            self.i = i
        def next(self):
            self.i += 1
            return self.i
    return Counter(i = 0).next

def log(*args, **kwargs):
    import sys
    print(*args, file=sys.stderr, **kwargs)

def is_endstate(state):
    if len(state) != 3:
        raise ValueError("state must have 3 items")
    last_column = state[2]
    if len(last_column) != 3:
        return False
    return last_column[0] == BIGBAR and last_column[1] == MEDBAR and last_column[2] == SMALLBAR

def transition(start_state, origin, destination):
    if (origin == destination):
        return None
    end_state = deepcopy(start_state)
    if len(end_state[origin]) == 0:
        return None
    v = end_state[origin].pop()
    if len(end_state[destination]) > 0 and len(end_state[destination][-1]) > len(v):
        return None
    end_state[destination].append(v)
    return end_state

def generate_state(state):
    state_name = str(state).replace(' ', '')
    ret = "\n".join([",".join(line) for line in state])
    return state_name, ret

def num2char(n):
    return ALPHABET[n]

def traverse_states(current_state, depth = 12):
    name, label = generate_state(current_state)
    if visited.get(name) != None:
        # log("state já visitado")
        return None
    visited[name] = current_state
    transitions[name] = {}
    if is_endstate(current_state):
        log("estado final encontrado")
        return None
    if depth <= 0:
        log("depth tá mt baixo, saindo...")
        return None

    for i in range(0, 3):
        transition_state = name + num2char(i)
        visited[transition_state] = "intermediate"
        transitions[name][num2char(i)] = transition_state
        transitions[transition_state] = {}
        for j in range(0, 3):
            new_state = transition(current_state, i, j)
            if new_state == None:
                continue
            new_name, new_label = generate_state(new_state)
            transitions[transition_state][num2char(j)] = new_name
            traverse_states(new_state, depth - 1)

traverse_states(START_STATE)

dfa_keys = {}
transition_counter = counter()
for strkey in transitions.keys():
    state_id = transition_counter()
    dfa_keys[strkey] = state_id


dfa_transitions = {}

for strkey in transitions.keys():
    statekey = dfa_keys[strkey]
    strtransitions = transitions[strkey]
    dfa_transitions[statekey] = {}
    for transk in strtransitions.keys():
        dfa_transitions[statekey][transk] = dfa_keys[strtransitions[transk]]

dfa_endstates = []
for key in dfa_keys:
    state_value = visited.get(key)
    if state_value == None:
        continue
    if state_value == "intermediate":
        continue
    if is_endstate(state_value):
        dfa_endstates.append(dfa_keys[key])

def print_dfa():
    print("digraph dfa {")
    print('"" [shape=none]')
    print('"" -> "1"')
    for key in dfa_keys.values():
        print(f'"{key}" [shape=circle label="{key}"]')
    for endstate in dfa_endstates:
        print(f'"{endstate}" [shape=doublecircle label="{endstate}"]')
    for node_key in dfa_transitions.keys():
        for trans_key in dfa_transitions[node_key].keys():
            trans_destination = dfa_transitions[node_key][trans_key]
            print(f'"{node_key}" -> "{trans_destination}" [label="{trans_key}"]')
    print("}")

print_dfa()
