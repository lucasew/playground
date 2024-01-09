#!/usr/bin/env nix-shell
#!nix-shell -i python3 -p python3Packages.flet

import flet
from flet import Page

def main(page: Page):
    page.title = "Teste"
    page.add(Text(value="Teste"))

flet.app(target=main)
