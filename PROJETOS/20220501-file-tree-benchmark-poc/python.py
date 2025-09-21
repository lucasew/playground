from pathlib import Path

for f in Path('/home/lucasew').glob('**/*'):
    print(f)
