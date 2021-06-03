#!/usr/bin/env nix-shell
#! nix-shell -p python3 -i python

def early_return(item):
    try:
        if item:
            print("true")
            return 1
        else:
            print("false")
            return 0
    finally:
        print("defer")

early_return(True)
early_return(False)
