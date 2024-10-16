#!/usr/bin/env python3

class B:
    def get_type(self):
        return type(self)

class A(B):
    pass

if __name__ == "__main__":
    print(A().get_type())
