#!/usr/bin/env nix-shell
#!nix-shell -i python -p python3Packages.pypresence

from pypresence import Presence
from time import sleep

# client_id = '69420694206969'
client_id = '937827341332267028'
RPC = Presence(client_id, pipe = 0)
RPC.connect()


while True:
    RPC.update(state="Vida real", details = "Acha que eu fico só no PC garai")
    sleep(15)
