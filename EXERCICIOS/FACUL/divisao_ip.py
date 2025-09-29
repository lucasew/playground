#!/usr/bin/env nix-shell
#!nix-shell -p python3 -i python

"""
Programa criado como atividade na matéria de Redes 1
"""

def pad_bin(number, zfill = 8):
    return bin(number).__str__()[2:].zfill(zfill)

def is_binary_expr(expr):
    for c in expr:
        if c not in ["0", "1"]:
            return False
    return True

def parse_binary_expr(expr):
    if is_binary_expr(expr):
        return eval("0b%s" %(expr))
    raise TypeError("'%s' is not a binary expression", expr)

class IP():
    def __init__(self, expr, try_binary = False):
        if type(expr) == list and len(expr) == 4:
            self._ip = [int(str(v)) for v in expr]
            self._validate_input()
            return
        if type(expr) == int:
            expr = pad_bin(expr, zfill = 32)
        if type(expr) == str and len(expr) == 32:
            new_expr = []
            for i in range(0, 32):
                if i in [8, 16, 24]:
                    new_expr.append(".")
                new_expr.append(expr[i])
            expr = "".join(new_expr)
            try_binary = True
        if type(expr) == str:
            dotsplit = expr.split(".")
            if len(dotsplit) == 4:
                ip = []
                for part in dotsplit:
                    if is_binary_expr(part) and try_binary:
                        ip.append(parse_binary_expr(part))
                    else:
                        ip.append(int(part))
                self._ip = ip
                self._validate_input()
                return
        raise TypeError("not a valid input for IP")

    def _validate_input(self):
        for val in self._ip:
            if val > 255 or val < 0:
                raise TypeError("value must be between 0 and 255 but its %s" %(val))

    def bin_repr(self):
        r = []
        for octet in self._ip:
            r.append(pad_bin(octet))
        return ".".join(r)

    def bin_stream(self):
        return self.bin_repr().replace(".", "")

    def __int__(self):
        return parse_binary_expr(self.bin_stream())

    def is_mask(self):
        chars = self.bin_stream()
        if chars[0] == "0": # Máscaras não começam com 0
            return False
        justOnes = True # A ideia aqui é tipo implementar um automato
        for ch in chars:
            if justOnes and (not chars == "1"):
                justOnes = False
                continue
            if (not justOnes) and chars == "1":
                return False
        return True

    def mask_number_hosts(self):
        if self.is_mask():
            s = self.bin_stream()
            bits = 0
            for c in s:
                if c == "0":
                    bits += 1
            return (2**bits) - 2
        raise TypeError("%s is not a mask", self.__repr__())

    def mask_number_nets(self):
        if self.is_mask():
            s = self.bin_stream()
            bits = 0
            for c in s:
                if c == "1":
                    bits += 1
            return (2**bits)
        raise TypeError("%s is not a mask", self.__repr__())

    def get_class(self):
        [a, b, c, d] = self._ip
        if a <= 127:
            return "A"
        elif a <= 191:
            return "B"
        elif a <= 223:
            return "C"
        elif a <= 239:
            return "MULTICAST"
        elif a <= 255:
            return "EXPERIMENTAL"

    def classful_mask(self):
        cls = self.get_class()
        if cls == "A":
            return IP("255.0.0.0")
        if cls == "B":
            return IP("255.255.0.0")
        if cls == "C":
            return IP("255.255.255.0")

    def classful_netid(self):
        cls = self.get_class()
        if cls == "A":
            return "%d/8" %(self._ip[0])
        if cls == "B":
            return "%s/16" %(".".join([str(s) for s in self._ip[0:2]]))
        if cls == "C":
            return "%s/24" %(".".join([str(s) for s in self._ip[0:3]]))

    def __xor__(self, another):
        return IP(int(self) ^ int(another))

    def __and__(self, another):
        return IP(int(self) & int(another))

    def __repr__(self):
        ipstr = ".".join([str(s) for s in self._ip])
        cls = self.get_class()
        return "<IP(%s): %s>" %(cls, ipstr)

def subrede(level):
    if type(level) == int and level > 0 and level < 32:
        ones = level
        zeroes = 32 - level
        from itertools import repeat
        ret = []
        for i in range(0, 32):
            if i in [8, 16, 24]:
                ret.append(".")
            if i < level:
                ret.append("1")
            else:
                ret.append("0")
        return IP("".join(ret), try_binary = True)

def meu_ip():
    from urllib.request import urlopen
    with urlopen("http://ifconfig.me") as response:
        return IP(response.read().decode('utf-8'))

if __name__ == "__main__":
    print(IP("10.163.69.121") & subrede(29))

# usa `python -i script.py` pra chamar o script no repl
