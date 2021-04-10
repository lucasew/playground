#! /usr/bin/env nix-shell
#! nix-shell -p python3 -i python

class Test:
    def __init__(self):
        return
    def __enter__(self):
        print("Entrando...")

    def __exit__(self, t, v, traceback):
        print("Saindo...")
        print(t)
        print(v)
        print(traceback)

if __name__ == "__main__":
    with Test():
        print("Realizando ações")
