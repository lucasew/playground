#!/usr/bin/env -S sd nix shell --really
#!nix-shell -i python3 -p unstable.python3Packages.ollama
from argparse import ArgumentParser, ArgumentDefaultsHelpFormatter
from pathlib import Path
from sys import stderr

import ollama

parser = ArgumentParser(description="Get information from images")
parser.add_argument("images", type=Path, nargs='+')
parser.add_argument('--model', default='llama3.2-vision')

args = parser.parse_args()

system_prompt = '''
Act as an assistant that converts a document structure to Markdown.

You always follow these rules:
- Use only data from the document itself.
- Put the final result between "[final]" and "[/final]".
- Never take the meaning of words into consideration. You only see structure.
- All numbers, links and spelling of words must be double checked.
- Make sure all the elements are in the right place.
'''


system_prompt = system_prompt.strip()

print(system_prompt, file=stderr)
for progress in ollama.pull(args.model, stream=True):
    print(progress, file=stderr)

ret = ollama.chat(
    model=args.model,
    messages=[
        {
            'role': 'user',
            'content': system_prompt,
            'images': args.images
        }
    ],
    options = {
        'num_predict': 128*1024,
        'temperature': 0,
        # 'top_p': 0.99,
        'stop': ['[/final]']
    }
)
print(ret, file=stderr)
print(ret['message']['content'])
