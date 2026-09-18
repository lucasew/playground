# Print one north-south strip of the land as 32 columns of 16 digits
# (y grows downward in the string). 0 air, 1 grass, 2 dirt, 3 stone,
# 4 wood, 5 leaves, 6 water. No window.
#
#   bend slice.bend
import Base
import ./main.bend as M

def main() -> IO(Unit):
  do IO<Unit>:
    Unit <- IO.print(M.Mine.slice(M.Mine.empty(), 8))
    IO.print(M.Mine.column(M.Mine.empty(), 3, 5))
