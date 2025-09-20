#!/usr/bin/env -S sd nix shell --really
#!nix-shell -i python3 -p unstable.python3Packages.ollama
from argparse import ArgumentParser, ArgumentDefaultsHelpFormatter
from pathlib import Path
from sys import stderr
import re

import ollama

parser = ArgumentParser(description="Get information from images")
parser.add_argument("images", type=Path, nargs='+')
parser.add_argument('--model', default='llama3.2-vision')

args = parser.parse_args()

classes = dict(
    TEXT="The image has only text inside",
    TABLE="The image is a table",
    WHITESPACE="The image has nothing but a single color or a filled surface",
    PICTURE="The image is some kind of picture, choose this as last resort"
)

class_descriptions = []
for cls_name, cls_description in classes.items():
    class_descriptions.append(f"- !{cls_name}!: {cls_description}")

system_prompt = f'''
You are a image classifier. You must choose a class. These are your classes:

{"\n".join(class_descriptions)}
'''

system_prompt = system_prompt.strip()

print(system_prompt, file=stderr)
for progress in ollama.pull(args.model, stream=True):
    print(progress, file=stderr)

history = [
    {
        'role': 'user',
        'content': system_prompt,
        'images': args.images
    }
]
while True:
    ret = ollama.chat(
        model=args.model,
        messages=history,
        options = {
            'num_predict': 128*1024,
            'temperature': 0,
            # 'top_p': 0.99,
        }
    )
    print(ret, file=stderr)
    result = re.findall(r"!([A-Z]*)!", ret['message']['content'], re.MULTILINE)
    if len(result) > 0:
        print(result[0])
        break
    history.append(ret['message'])
    history.append({
        'role': 'user',
        'content': f'''
            Please choose a class:

            {"\n".join(class_descriptions)}
        '''
    })
