#! /usr/bin/env nix-shell
#! nix-shell -p python3 -i python

# dot -Tsvg teste.graph -o out.svg

from copy import deepcopy
# torres: a b c
A=0
B=1
C=2
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

def log(*args, **kwargs):
    import sys
    print(*args, file=sys.stderr, **kwargs)

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

def traverse_states(current_state, depth = 999):
    if visited.get(str(current_state)) != None:
        # log("state já visitado")
        return None
    visited[str(current_state)] = True
    current_state_name, current_state_label = generate_state(current_state)
    def get_color(tower, piece):
        pieces = current_state[tower]
        if piece >= len(pieces):
            return "white"
        if pieces[piece] == SMALLBAR:
            return "green"
        if pieces[piece] == MEDBAR:
            return "yellow"
        if pieces[piece] == BIGBAR:
            return "red"
    print(f"""
"{current_state_name}" [
shape=box
label=<
<TABLE>
<TR>
    <TD BGCOLOR="{get_color(0,0)}"></TD>
    <TD BGCOLOR="{get_color(0,1)}"></TD>
    <TD BGCOLOR="{get_color(0,2)}"></TD>
</TR>
<TR>
    <TD BGCOLOR="{get_color(1,0)}"></TD>
    <TD BGCOLOR="{get_color(1,1)}"></TD>
    <TD BGCOLOR="{get_color(1,2)}"></TD>
</TR>
<TR>
    <TD BGCOLOR="{get_color(2,0)}"></TD>
    <TD BGCOLOR="{get_color(2,1)}"></TD>
    <TD BGCOLOR="{get_color(2,2)}"></TD>
</TR>
</TABLE>>
]
    """)
    if depth <= 0:
        # log("depth tá mt baixo, saindo...")
        return None

    for i in range(0, 3):
        for j in range(0, 3):
            new_state = transition(current_state, i, j)
            if new_state == None:
                continue
            if is_endstate(new_state):
                log("encontrado um estado final")
                print("""
"%s" -- "ENDSTATE"
                """ %(current_state_name))
                continue
            name, label = generate_state(new_state)
            print(f"""
"{current_state_name}" -- "{name}"
            """)
            traverse_states(new_state, depth - 1)


print("strict graph G {")
if __name__ == "__main__":
    print("""
"STARTSTATE" [
label =<<B>Começo</B>>
]
            """)
    name, label = generate_state(START_STATE)
    print('STARTSTATE -- "%s"' %(name))
    print("""
"ENDSTATE" [
label = <<B>FIM</B>>
]
            """)
    traverse_states(START_STATE)


print("}")
log("nós visitados: ", len(visited))

assert(transition([[BIGBAR], [MEDBAR], [SMALLBAR]], 0, 2) == None)
