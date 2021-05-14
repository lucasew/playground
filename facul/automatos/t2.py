#! /usr/bin/env nix-shell
#! nix-shell -p python3 -i python
import sys

class StackAutomata:
    def __init__(self, start_state = None, final_state = None):
        self.states = {}
        assert(final_state != None)
        self.start_state = start_state
        self.final_state = final_state

    def route(self, from_state, ch, stack_top):
        if self.states.get(from_state) == None:
            return None
        if self.states[from_state].get(ch) == None:
            return None
        if self.states[from_state][ch].get(stack_top) == None:
            return None
        return self.states[from_state][ch][stack_top]

    def add_route(self, origin, destination, char, stack_char, push):
        if self.states.get(origin) == None:
            self.states[origin] = {}
        if self.states[origin].get(char) == None:
            self.states[origin][char] = {}
        if self.states[origin][char].get(stack_char) == None:
            self.states[origin][char][stack_char] = {}
        self.states[origin][char][stack_char] = [push, destination]
        if self.start_state == None:
            self.start_state = self.start_state = origin
    def check(self, stmt):
        if len(stmt) > 0 and stmt[-1] != "\0":
            stmt = stmt + "\0"
        stack = []
        def push_stack(items):
            items = list(items)
            # print(f'inserindo {items}')
            # if len(items) == 0:
            #     log(f"nada pra dar push: {items}")
            for item in items:
                stack.append(item)
        def peek_stack():
            if len(stack) == 0:
                return None
            return stack[-1]
        def pop_stack():
            return stack.pop()
        push_stack("Z") # pilha inicial
        state = self.start_state # estado inicial
        # print(stmt)
        for ch in stmt:
            stack_top = peek_stack()
            route = self.route(state, ch, stack_top)
            if route == None:
                log("no route")
                return False
            to_push, next_state = route
            pop_stack()
            push_stack(to_push)
            state = next_state
            # print(f"automata tick: ch = {ch}, stack = {stack}, state = {state}.")
        is_final = len(stack) == 0
        # print(f"final = {state == self.final_state}, stack = {len(stack)}")
        return is_final

def log(*args, **kwargs):
    print(*args, file=sys.stderr, **kwargs)

def outro_states():
    ret = StackAutomata(start_state = 0, final_state = 8)
    ret.add_route(0,0, "a", "Z", "ZA")
    ret.add_route(0,0, 'a', 'A', 'AA')
    ret.add_route(0,1, 'b', 'Z', 'Z')
    ret.add_route(0,1, 'b', 'A', 'A')
    ret.add_route(1,2, 'c', 'A', 'A')
    ret.add_route(2,1, 'c', 'A', '')
    ret.add_route(1,3, 'd', 'Z', 'Z')
    ret.add_route(3,3, 'd', 'Z', 'Z')
    ret.add_route(3,4, 'e', 'Z', 'Z')
    ret.add_route(4,5, 'e', 'Z', 'Z')
    ret.add_route(5,6, 'e', 'Z', 'ZE')
    ret.add_route(6,4, 'e', 'E', 'E')
    ret.add_route(4,5, 'e', 'E', 'E')
    ret.add_route(5,6, 'e', 'E', 'EE')
    ret.add_route(6,7, 'f', 'E', '')
    ret.add_route(7,7, 'f', 'E', '')
    ret.add_route(7,8, '\0', 'Z', '')
    return ret

def icaro_states():
    ret = StackAutomata(start_state = 0, final_state = 6)
    ret.add_route(0, 0, "a", "X", "XXX")
    ret.add_route(0, 0, "a", "Z", "XXZ")
    ret.add_route(0, 1, "b", "X", "X")
    ret.add_route(0, 1, "b", "Z", "Z")
    ret.add_route(1, 1, "c", "X", "")
    ret.add_route(1, 2, "d", "Z", "Z")
    ret.add_route(2, 2, "d", "Z", "Z")
    ret.add_route(2,3, "e", "Z", "X")
    ret.add_route(3,4,"e", "X", "X")
    ret.add_route(4,5,"e", "X", "X")
    ret.add_route(5,3, "e", "X", "XX")
    ret.add_route(5, 6, "f", "X", "")
    ret.add_route(6, 6, "f", "X", "")
    return ret

# states = icaro_states()
states = outro_states()
print(states.__dict__)

def check(stmt):
    ret = states.check(stmt)
    return ret

def pass_testcase(stmt):
    if not check(stmt):
        print(f"fail: {stmt} <=> valido")
    # else:
    #     print(f"pass valid")

def fail_testcase(stmt):
    if check(stmt):
        print(f"fail: {stmt} <=> inválido")
    # else:
    #     print(f"pass invalid")

def sentence_generator(n, x, m):
    # assert(n >= 0)
    # assert(m >= 1)
    # assert(x >= 1)
    return ("a"*n) + "b" + ("c" * 2 * n) + ("d" * x) + ("e" * 3 * m)+ ("f" * m)

pass_testcase("abccdeeef") # n = 1, x = 1, m = 1
pass_testcase("bdeeef") # n = 0, x = 1, m = 1
fail_testcase("bdeef") 

pass_testcase("bddeeef") # n = 0, x = 2, m = 1
fail_testcase("bddeef")

pass_testcase("bddeeeeeeff") # n = 0, x = 2, m = 2
fail_testcase("bddeeeeeef")
fail_testcase("bddeeeeeff")

pass_testcase("abccddeeef") # n = 1, x = 2, m = 1
fail_testcase("abccddeef")
fail_testcase("abcddeeef")

pass_testcase("abccddeeeeeeff") # n = 1, x = 2, m = 2
fail_testcase("abccddeeeeeff") 
fail_testcase("abccddeeeeeef")
fail_testcase("abccddeeeeff")

for n in range(0, 10):
    for m in range(1, 10):
        for x in range(1, 10):
            if m == 0 or x == 0:
                fail_testcase(sentence_generator(n, x, m))
            else:
                pass_testcase(sentence_generator(n, x, m))

fail_testcase("a")
fail_testcase("b")
fail_testcase("c")
fail_testcase("d")
fail_testcase("e")
fail_testcase("f")
fail_testcase("")
