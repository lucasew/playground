# Mine -- the laws: Esc quits, a held key reads back, air is empty,
# and a written cell is the value that was written. The F32 DDA is
# not claimed.
import Base
import ./main.bend as M

# LAW: an Esc key-down quits, whatever follows it and whatever the state.
law esc_quits:
  for rest: List<Event>
  for s: M.Mine
  {M.Mine.events(Key{27, True{}} <> rest, s) == None{} : Maybe<M.Mine>}

# LAW: after a key is pressed (or released), that key reads pressed (or
# released), for any slot and any held keys.
law press_reads:
  for held: List<&2, Bool>
  for +k: Nat
  for +v: Bool
  {M.Mine.get(M.Mine.set(held, k, v), k) == v : Bool}

# LAW: air is not a solid block.
law air_not_solid:
  {M.Mine.solid(0) == False{} : Bool}

# LAW: water is not a solid block.
law water_not_solid:
  {M.Mine.solid(6) == False{} : Bool}

# LAW: grass is a solid block.
law grass_is_solid:
  {M.Mine.solid(1) == True{} : Bool}

# LAW: a found edit is the block at that cell.
law edit_wins:
  for +id: U32
  for +x: U32
  for +y: U32
  for +z: U32
  {M.Mine.block.go(Some{id}, x, y, z) == id : U32}

# LAW: writing 2 at (3,4,5) on an empty map is what a later read returns.
law put_at:
  {M.Mine.block(M.Mine.put(M.Mine.empty(), 3, 4, 5, 2), 3, 4, 5) == 2 : U32}

# LAW: a miss paints the sky of that ray.
law miss_is_sky:
  for +dy: F32
  {M.Mine.paint(M.Miss{}, dy) == M.Mine.sky(dy) : U32}
