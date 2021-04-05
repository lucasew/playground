#! /usr/bin/env nix-shell
#! nix-shell -p python3 -i python

# dot -Tsvg teste.graph -o out.svg

nodes_processed = 0

print("digraph G {")

def check_equality(board, pos_a, pos_b, pos_c):
    board[a] == board[b] == board[c]

def check_winner(board):
    win_conditions = [
        [0, 1, 2], # horizontal
        [3, 4, 5],
        [6, 7, 8],
        [0, 3, 6], # vertical
        [1, 4, 7],
        [2, 5, 8],
        [0, 4, 8], # diagonal
        [2, 4, 6],
    ]
    for condition in win_conditions:
        if board[condition[0]] == board[condition[1]] == board[condition[2]]:
            if board[condition[0]] != "_":
                return board[condition[0]]
    return None

def toggle_player(player):
    if player == "x":
        return "o"
    if player == "o":
        return "x"
    raise Error("%s is not a valid player" %(player))

def generate_2d_representation(board):
    ret = ""
    for i in range(len(board)):
        ret += board[i]
        if i in [2, 5]:
            ret += "\\n"
        else:
            ret += " "
    return ret

def print_board_state(board):
    print("""
%s [
label = "%s"
]
    """ %(board, generate_2d_representation(board)))


def moves(current_board = "_________", player = "o", recur = -1, symetry = False):
    global nodes_processed
    nodes_processed += 1
    for i in range(len(current_board)):
        if symetry and i in [5, 6, 7, 8]:
            return
        if current_board[i] == "_":
            new_board = list(current_board)
            new_board[i] = player
            new_board = "".join(new_board)
            print_board_state(current_board)
            print_board_state(new_board)
            print("%s -> %s" %(current_board, new_board))
            if check_winner(new_board) != None:
                return
            if recur != 0:
                moves(new_board, toggle_player(player), recur = recur - 1)

moves("_________", player = "x", recur = 1, symetry = True)
print("// nodes processed: %d" %(nodes_processed))
print("}")
