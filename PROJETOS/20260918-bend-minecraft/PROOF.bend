# Mine -- the proofs. Esc quits by computation; press_reads is an
# induction over the held list with a case analysis on the slot; the
# rest compute.
import Base
import ./main.bend as M
import ./LAWS.bend as Laws

def Laws.esc_quits(rest, s):
  {==}

def Laws.press_reads(held, k, v):
  match held k:
    case [] 0n:
      {==}
    case [] 1n+j:
      Laws.press_reads([], j, v)
    case h <> t 0n:
      {==}
    case h <> t 1n+j:
      Laws.press_reads(t, j, v)

def Laws.air_not_solid():
  {==}

def Laws.water_not_solid():
  {==}

def Laws.grass_is_solid():
  {==}

def Laws.edit_wins(id, x, y, z):
  {==}

def Laws.put_at():
  {==}

def Laws.miss_is_sky(dy):
  {==}
