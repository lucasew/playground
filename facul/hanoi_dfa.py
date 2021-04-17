#! /usr/bin/env nix-shell
#! nix-shell -p python3 -i python
import sys
from copy import deepcopy


"""
# Torres de Hanoi
## Abstração da transição
- Primeira letra é a coluna que vai ter o ultimo elemento removido
- Segunda letra é a coluna que vai receber o elemento
"""


def log(*args, **kwargs):
    print(*args, file=sys.stderr, **kwargs)

def write_visual_feedback(text):
    sys.stderr.write(text)
    sys.stderr.flush()

def counter(i = 0):
    class Counter:
        def __init__(self, i = 0):
            self.i = i
        def next(self):
            self.i += 1
            return self.i
    return Counter(i = 0).next

def get_dfa():
    ALPHABET = ["a", "b", "c"]
    # pinos são uma tupla, do menor para o maior e a tupla é de posições
    SMALLBAR="_  "
    MEDBAR=  "__ "
    BIGBAR=  "___"

    def num2char(n):
        return ALPHABET[n]
    ENDSTATE="endstate"

    START_STATE = [
        [BIGBAR, MEDBAR, SMALLBAR],
        [],
        []
    ]

    visited = {}
    transitions = {}
    def is_endstate(state):
        if len(state) != 3:
            raise ValueError("state must have 3 items")
        last_column = state[2]
        if len(last_column) == 3 and last_column[0] == BIGBAR and last_column[1] == MEDBAR and last_column[2] == SMALLBAR:
            return True
        last_column = state[1]
        return len(last_column) == 3 and last_column[0] == BIGBAR and last_column[1] == MEDBAR and last_column[2] == SMALLBAR

    def transition(start_state, origin, destination):
        if (origin == destination):
            return None
        end_state = deepcopy(start_state)
        if len(end_state[origin]) == 0:
            return None
        v = end_state[origin].pop()
        if len(end_state[destination]) > 0:
            if len(end_state[destination][-1].strip()) < len(v.strip()):
                return None
        end_state[destination].append(v)
        return end_state

    def generate_state(state):
        state_name = str(state).replace(' ', '')
        ret = "\n".join([",".join(line) for line in state])
        return state_name, ret


    def traverse_states(current_state):
        name, label = generate_state(current_state)
        if visited.get(name) != None:
            # log("state já visitado")
            return None
        visited[name] = current_state
        transitions[name] = {}
        if is_endstate(current_state):
            log("estado final encontrado")
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
                traverse_states(new_state)

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
    return DFA(ALPHABET, dfa_keys.values(), dfa_endstates, dfa_transitions)

class DFA():
    def __init__(self, alphabet, states, endstates, transitions):
        self.alphabet = list(alphabet)
        self.endstates = list(endstates)
        self.transitions = transitions
        self.states = list(states)
    def num2char(self, n):
        return self.alphabet[n]
    def print_graphviz(self):
        print("digraph dfa {")
        print('"" [shape=none]')
        print('"" -> "1"')
        for key in self.transitions:
            print(f'"{key}" [shape=circle label="{key}"]')
        for endstate in self.endstates:
            print(f'"{endstate}" [shape=doublecircle label="{endstate}"]')
        for node_key in self.transitions.keys():
            for trans_key in self.transitions[node_key].keys():
                trans_destination = self.transitions[node_key][trans_key]
                print(f'"{node_key}" -> "{trans_destination}" [label="{trans_key}"]')
        print("}")
    def check_match(self, match):
        state = 1
        for ch in match:
            if ch not in self.alphabet:
                log("not in alphabet")
                return False
            next_state = self.transitions[state].get(ch)
            if next_state == None:
                log(f'state {state} does not have path {ch}')
                return False
            state = next_state
        return state in self.endstates
    def minimize(self):
        # remover estados sem saída
        while True: # sai só quando não houver mais estados pra remover
            to_remove = []
            for transition in self.transitions.keys():
                if len(self.transitions[transition]) == 0 and transition not in self.endstates:
                    to_remove.append(transition)
            if len(to_remove) == 0:
                return
            for r in to_remove:
                self.states.remove(r)
                del self.transitions[r]
                for transition in self.transitions.keys():
                    delete = []
                    for path in self.transitions[transition]:
                        if self.transitions[transition][path] in to_remove:
                            delete.append(path)
                    for d in delete:
                        del self.transitions[transition][d]

    def get_prev_states(self, state):
        ret = []
        for k in list(self.transitions.keys()):
            if state in self.transitions[k].values():
                ret.append(k)
        return ret

    def get_pos_states(self, state):
        return list(self.transitions[state].values())

    def get_if_single_hop(self, orig, dest):
        return dest in self.transitions[orig].values()

    def to_regex(self):
        # tá usando ram bagaray
        dict_states = {r: {c: RegexpAST.new_empty() for c in self.states} for r in self.states}
        for i in self.states:
            for j in self.states:
                for k in self.transitions[i].keys():
                    if self.transitions[i][k] == j:
                        dict_states[i][j] = RegexpAST.new_summing(k)
        init_state = 1
        non_intermediates = [init_state, *self.endstates]
        intermediates = [state for state in self.states if state not in non_intermediates]
        for inter in list(intermediates):
            log(f"Eliminating state: {inter}")
            before_me = []
            for ki in dict_states.keys():
                for kj in dict_states[ki].keys():
                    if kj == inter and not dict_states[ki][kj].is_empty():
                        before_me.append(ki)
            write_visual_feedback(">")
            after_me = []
            for ki in dict_states[inter].keys():
                if not dict_states[inter][ki].is_empty():
                    after_me.append(ki)
            write_visual_feedback(">")
            for before in list(before_me):
                write_visual_feedback("+")
                for after in list(after_me):
                    write_visual_feedback("-")
                    inter_loop = dict_states[inter][inter]
                    before_inter = dict_states[before][inter]
                    before_after = dict_states[before][after]
                    inter_after = dict_states[inter][after]
                    if (len(before_inter) + len(inter_after) + len(inter_loop)) == 0:
                        # log("nothing to add, continuing")
                        continue
                    cur = dict_states[before][after]
                    first = dict_states[before][inter]
                    second = RegexpAST.star_it(cur)
                    third = RegexpAST.star_it(dict_states[inter][inter])
                    fourth = dict_states[inter][after]
                    dict_states[before][after] = cur + RegexpAST.concat(first, second, third, fourth)
            dict_states = {r: {c: v for c, v in val.items() if c != inter} for r, val in dict_states.items() if r != inter}
        init_loop = dict_states[init_state][init_state]
        init_to_final = RegexpAST.concat(dict_states[init_state][self.endstates[0]], RegexpAST.star_it(dict_states[self.endstates[0]][self.endstates[0]]))
        final_to_init = dict_states[self.endstates[0]][init_state]
        return RegexpAST.concat(RegexpAST.star_it(init_loop + init_to_final + final_to_init), init_to_final)

    def __len__(self):
        return len(self.states)

class RegexpAST():
    SYM_STAR = '*'
    SYM_PLUS = '+'
    SYM_EMPTY = ''

    def new_empty():
        return RegexpAST(RegexpAST.SYM_EMPTY, '')
    def new_summing(*items):
        return RegexpAST(RegexpAST.SYM_PLUS, *items)
    def star_it(regexp):
        return RegexpAST(RegexpAST.SYM_EMPTY, regexp, star = True)
    def concat(*items):
        return RegexpAST(RegexpAST.SYM_EMPTY, *items)

    def __init__(self, root_term, *children, star = False):
        children = list(children)
        if root_term not in [self.SYM_PLUS, self.SYM_EMPTY]:
            raise TypeError(f"invalid root_term")
        if len(children) == 0:
            raise TypeError("at least one regexp children is required")
        self.root_term = root_term
        self.star = star
        self.children = []
        for child in children:
            if type(child) == str:
                self.children.append(child)
                continue
            if type(child) == RegexpAST:
                self.children.append(child)
                continue
            raise TypeError("all children should be string or RegexpAST")
    def is_empty(self):
        if len(self.children) > 1:
            return False
        if type(self.children[0]) == str:
            return len(self.children[0]) == 0
        return self.children[0].is_empty()
    def __add__(self, another):
        return RegexpAST(self.SYM_PLUS, self, another)
    def __len__(self):
        childlen = 0
        for ch in self.children:
            childlen += len(ch)
        star_size = 0
        if self.star:
            star_size = 1
        return 2 + childlen + (len(self.root_term) * (len(self.children) - 1)) + star_size # parenthesis + children + item separators + 1 if star at the end
    def print_item(self, f = sys.stdout):
        childlen = len(self.children)
        f.write("(")
        for i in range(0, childlen):
            if i >= 0 and i <= (childlen - 1):
                f.write(self.root_term)
            chtype = type(self.children[i])
            if chtype == str:
                f.write(self.children[i])
            if chtype == RegexpAST:
                self.children[i].print_item(f)
        f.write(")")
        if self.star:
            f.write(self.SYM_STAR)


dfa = get_dfa()
log(f"quantidade de estados: {len(dfa)}")
assert(dfa.check_match("acabcbacbabcac")) # optimum
dfa.minimize()
log(f"quantidade de estados minimizada: {len(dfa)}")

dfa_regex = dfa.to_regex()
dfa_regex.print_item(sys.stdout)
# log(dfa_regex)
# log(len(dfa_regex))
# dfa.print_graphviz()

# rg = RegexpAST.concat("a", "b")
# rg.print_item(sys.stdout)
# log(len(rg))
