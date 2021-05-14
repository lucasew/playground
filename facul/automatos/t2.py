#! /usr/bin/env nix-shell
#! nix-shell -p python3 -i python

def log(*args, **kwargs):
    print(*args, file=sys.stderr, **kwargs)

def check(stmt):
    # TODO: implement
    return False

def pass_testcase(stmt):
    if not check(stmt):
        log(f"fail: {stmt} <=> valido")

def fail_testcase(stmt):
    if check(stmt):
        log(f"fail: {stmt} <=> inválido")

pass_testcase("abccdeeef") # n = 1, x = 1, m = 1
fail_testcase("abcdeeef")
fail_testcase("accdeeef")

pass_testcase("bdeeef") # n = 0, x = 1, m = 1
fail_testcase("bdeef") 

pass_testcase("bddeeef") # n = 0, x = 2, m = 1
fail_testcase("bddeef")

pass_testcase("bddeeeeeeff") # n = 0, x = 2, m = 2
fail_testcase("bddeeeeeef")
fail_testcase("bddeeeeeff")

pass_testcase("abccddeeef") # n = 1, x = 2, m = 1
fail_testcase("abccddeef")
fail_testcase("abccdeeef")
fail_testcase("abcddeeef")

pass_testcase("abccddeeeeeeff") # n = 1, x = 2, m = 2
fail_testcase("abccddeeeeeff") 
fail_testcase("abccddeeeeeef")
fail_testcase("abccddeeeeff")

fail_testcase("a")
fail_testcase("b")
fail_testcase("c")
fail_testcase("d")
fail_testcase("e")
fail_testcase("f")
fail_testcase("")
