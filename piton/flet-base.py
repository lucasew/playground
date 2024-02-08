#!/usr/bin/env nix-shell
#!nix-shell -i python3 -p python3Packages.flet

import logging
logging.basicConfig(level=logging.DEBUG)

import flet as ft
from flet import Page, Text

def main(page: Page):
    page.title = "Teste"
    page.add(Text(value="Teste"))

ft.app(target=main, view=ft.AppView.FLET_APP)
