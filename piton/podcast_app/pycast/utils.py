
def fingerprint_string(s: str) -> str:
    from hashlib import sha256
    symboldict = [
        *map(chr, range(ord('A'), ord('Z') + 1)),
        *map(chr, range(ord('0'), ord('9') + 1)),
    ]
    m = sha256()
    l = []
    for char in s.upper():
        if char in symboldict:
            l.append(char)
    txt = "".join(l)
    m.update(txt.encode())
    return m.hexdigest()

def strduration2seconds(duration: str) -> int:
    parts = duration.split(":")
    multiplier = 1
    ret = 0
    while len(parts) > 0:
        part = int(parts.pop())
        ret += multiplier * part
        multiplier *= 60
    return ret

def reexport_url(url: str):
    from urllib.request import urlopen
    from io import BufferedReader
    return BufferedReader(urlopen(url))
