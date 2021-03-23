#!/usr/bin/env nix-shell
#!nix-shell -p python3Packages.kivy python3 -i python

from kivy.app import App
from kivy.uix.label import Label

class MainApp(App):
    def build(self):
        return Label(text="O JOGO")

MainApp().run()
