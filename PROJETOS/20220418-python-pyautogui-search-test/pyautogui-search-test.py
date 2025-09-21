#!/usr/bin/env nix-shell
#!nix-shell -i python -p python3Packages.pysimplegui python3Packages.pystray python3Packages.pillow libappindicator python3Packages.pygobject3 gtk3-x11 gobject-introspection

import PySimpleGUI as sg
from pystray import Icon, Menu, MenuItem
import pystray as tray

from PIL import Image, ImageDraw
from io import BytesIO

# print(tray.Icon.__dict__.items())

def create_image(width, height, color1, color2):
    # Generate an image and draw a pattern
    image = Image.new('RGB', (width, height), color1)
    dc = ImageDraw.Draw(image)
    dc.rectangle(
        (width // 2, 0, width, height // 2),
        fill=color2)
    dc.rectangle(
        (0, height // 2, width // 2, height),
        fill=color2)
    return image


sg.theme("DarkAmber")

def show_search_box(icon, item):
    class SearchBox:
        def __init__(self):
            self.selected = 0
            self.NUM_ITEMS = 5
            self.btn_keys = [f"btn:{i}" for i in range(self.NUM_ITEMS)]
            self.all_items = []
            layout = [
                [sg.InputText(key = "input", enable_events = True, focus = True)],
                [[
                    sg.Button(key = f"btn_{key}", enable_events = True),
                    sg.Text(key = f"txt_{key}")
                ] for key in self.btn_keys]
            ]
            self.window = sg.Window('Search', layout, 
                finalize = True,
                element_justification = 'center',
                return_keyboard_events = True,
                no_titlebar = True
            )
            for key in self.btn_keys:
                self.window[f"btn_{key}"].hide_row()
                self.window[f"txt_{key}"].hide_row()
        def update_list(self):
            text = self.window['input'].get()
            self.all_items = text.split()
            for key in self.btn_keys:
                self.window[f"btn_{key}"].hide_row()
                self.window[f"txt_{key}"].hide_row()
            if len(self.all_items) > 0:
                items = self.all_items[self.selected:][:self.NUM_ITEMS]
                for k, v in enumerate(items):
                    img = create_image(32, 32, 'white', 'black')
                    with BytesIO() as output:
                        img.save(output, format = "PNG")
                        data = output.getvalue()
                    # print(k, v)
                    key = self.btn_keys[k]
                    self.window[f"btn_{key}"].update(image_data = data)
                    self.window[f"txt_{key}"].update(value = v)
                    self.window[f"btn_{key}"].unhide_row()
                    self.window[f"txt_{key}"].unhide_row()
                    self.window[f"btn_{key}"].expand(expand_y = True)
                    self.window[f"txt_{key}"].expand(expand_y = True, expand_x = True)
                # print('len items', len(items))
            self.window.finalize()

        def run(self):
            while True:
                event, values = self.window.read(timeout = 1000)
                if event == sg.WIN_CLOSED:
                    break
                if event == '__TIMEOUT__':
                    continue
                if event == 'input':
                    self.update_list()
                    continue
                if len(event.split(':')) == 2:
                    k, kcode = event.split(':')
                    if k == "Escape":
                        break
                    if k == "Return":
                        # print(self.all_items[self.selected])
                        # print("FOI")
                        break
                    if k == "btn":
                        print('btn', k, kcode)
                    if k == "Down":
                        # print(self.selected, self.all_items)
                        if self.selected < len(self.all_items) - 1:
                            self.selected += 1
                            self.update_list()
                    if k == "Up":
                        if self.selected > 0:
                            self.selected -= 1
                            self.update_list()
                # print('args', event, values)
            self.window.close()
    SearchBox().run()
def main():
    Icon(
        'test',
        icon=create_image(64, 64, 'black', 'white'),
        menu = Menu(lambda: (
            MenuItem("Teste", lambda icon, item: print("Hello, world"), default = True),
            MenuItem("Search", show_search_box),
        ))
    ).run()

if __name__ == '__main__':
    main()
