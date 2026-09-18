# Mine

A first-person voxel world in [Bend](https://bend-lang.com). You fly a 32 x 16 x 32 land of grass, dirt, stone, water and trees, break blocks, and place new ones.

## Play

Needs Bend 2.0.6, clang, and a desktop session. On this machine Bend is:

```
~/.local/share/workspaced/tools/registry-bend/2.0.6/bin/bend
```

Build and run. Keep `mine.gpu` next to the binary; Bend writes it for the GPU.

```
bend main.bend -o mine && ./mine
```

| Key | Action |
|-----|--------|
| W A S D | Fly |
| Q / E | Up / down |
| Arrows | Look |
| 1 2 3 4 | Hand: grass, dirt, wood, leaves |
| Left click | Break the aimed block |
| Right click | Place the hand on the aimed face |
| Esc | Quit |

The top-left corner shows frames per second. The bottom-left strip is the hand.

## Check

`bend PROOF.bend` checks the laws (Esc quits, a key reads back, air is empty, a write is what a later read returns).

`bend slice.bend` prints one strip of the land as digits, then the tree at (3, 5). No window.

## Layout

| File | Role |
|------|------|
| `main.bend` | The game |
| `LAWS.bend` | Claims. Do not edit to make a proof pass |
| `PROOF.bend` | Proofs of those claims |
| `slice.bend` | ASCII strip of the land |
