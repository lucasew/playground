#! /usr/bin/env nix-shell
#! nix-shell -p python3 -i python
import sys
from copy import deepcopy
from random import randint

class StackAutomata:
    def __init__(self, start_state=None, final_state=None, start_stack = []):
        self.start_stack = start_stack
        self.states = {}
        assert (final_state is not None)
        self.start_state = start_state
        self.final_state = final_state

    def route(self, from_state, ch, stack_top):
        if self.states.get(from_state) is None:
            return None
        if self.states[from_state].get(ch) is None:
            return None
        if self.states[from_state][ch].get(stack_top) is None:
            return None
        return self.states[from_state][ch][stack_top]

    def add_route(self, origin, destination, char, stack_char, push):
        if self.states.get(origin) is None:
            self.states[origin] = {}
        if self.states[origin].get(char) is None:
            self.states[origin][char] = {}
        if self.states[origin][char].get(stack_char) is None:
            self.states[origin][char][stack_char] = {}
        self.states[origin][char][stack_char] = [push, destination]
        if self.start_state is None:
            self.start_state = self.start_state = origin

    def check(self, stmt):
        stack = deepcopy(self.start_stack)

        def push_stack(items):
            items = list(items)
            for item in items:
                stack.append(item)

        def peek_stack():
            if len(stack) == 0:
                return ""
            return stack[-1]

        def pop_stack():
            if len(stack) == 0:
                return ""
            return stack.pop()

        state = self.start_state
        for ch in stmt:
            stack_top = peek_stack()
            route = self.route(state, ch, stack_top)
            if route is None:
                return False
            to_push, next_state = route
            pop_stack()
            push_stack(to_push)
            # print(ch, state, next_state, stack_top, to_push, stack)
            state = next_state
        # is_final = len(stack) == 0
        is_final = state == self.final_state
        return is_final

def our_states():
    ret = StackAutomata(start_state = 1, final_state = 9, start_stack = ["Z"])
    def r(sa, sb, ch, st, sp):
        ret.add_route(sa, sb, ch, st, sp)
    r(1, 2, "a", "Z", "ZAA")
    r(1, 3, "b", "Z", "Z")
    r(2, 2, "a", "A", "AAA")
    r(2, 3, "b", "A", "A")
    r(3, 3, "c", "A", "")
    r(3, 4, "d", "Z", "Z")
    r(4, 4, "d", "Z", "Z")
    r(4, 5, "e", "Z", "Z")
    r(5, 6, "e", "A", "A")
    r(5, 6, "e", "Z", "Z")
    r(6, 7, "e", "A", "A")
    r(6, 7, "e", "Z", "Z")
    r(7, 5, "e", "A", "AA")
    r(7, 5, "e", "Z", "ZA")
    r(7, 8, "f", "A", "")
    r(7, 9, "f", "Z", "")
    r(8, 8, "f", "A", "")
    r(8, 9, "f", "Z", "")
    return ret

states = our_states()
print(states.__dict__)

def check(stmt):
    ret = states.check(stmt)
    return ret

def pass_testcase(stmt):
    if check(stmt):
        print(f"OK valid: {stmt}")
    else:
        print(f"FAIL valid: {stmt}")

def fail_testcase(stmt):
    if not check(stmt):
        print(f"OK invalid: {stmt}")
    else:
        print(f"FAIL invalid: {stmt}")

def sentence_generator(n, x, m):
    # assert(n >= 0)
    # assert(m >= 1)
    # assert(x >= 1)
    return ("a"*n) + "b" + ("c" * 2 * n) + ("d" * x) + ("e" * 3 * m)+ ("f" * m)


MAX_SIZE=1
for n in range(0, MAX_SIZE):
    for m in range(0, MAX_SIZE):
        for x in range(0, MAX_SIZE):
            fail_testcase(sentence_generator(n,x,m) + ("f" * randint(1, MAX_SIZE)))
            if m == 0 or x == 0:
                fail_testcase(sentence_generator(n, x, m))
            else:
                pass_testcase(sentence_generator(n, x, m))


pass_testcase("bdeeef")
