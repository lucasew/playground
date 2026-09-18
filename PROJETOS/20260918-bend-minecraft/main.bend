# Mine: a first-person voxel world. WASD fly, Q/E up/down, arrows look,
# 1-4 pick a block, left click breaks, right click places, Esc or Close
# quits. The land is a 32 x 16 x 32 heightmap of grass, dirt and stone
# with a few trees; edits sit in a Map so a lookup stays cheap as you
# dig. A frame is a 512 x 512 Image quadtree, built on the GPU by the
# one bang in the file, by DDA through the voxels.
#
#   bend main.bend -o mine && ./mine
import Base

type Mine.Hit is Data:
  Miss{}
  At{t: F32, face: U32, bx: U32, by: U32, bz: U32, id: U32}

type Mine.Rgb is Data:
  Rgb{r: F32, g: F32, b: F32}

type Mine.Pos is Data:
  Pos{x: U32, y: U32, z: U32}

type Mine.Cam is Data:
  Cam{ex: F32, ey: F32, ez: F32, fx: F32, fy: F32, fz: F32, rx: F32, rz: F32,
    ux: F32, uy: F32, uz: F32, hand: U32, fps: U32, edits: Map<&2, U32>}

type Mine is Data:
  Mine{x: F32, y: F32, z: F32, yaw: F32, pitch: F32, held: List<&2, Bool>,
    edits: Map<&2, U32>, hand: U32, t0: Nat, frames: U32, fps: U32}

def Mine.empty() -> Map<&2, U32>:
  Map.new(&2, U32)

def Mine.init() -> Mine:
  Mine{16.5, 8.5, 8.5, 0.0, (0.0 - 0.15 : F32), [], Map.new(&2, U32), 1, 0n, 0,
    0}

def Mine.pick(c: Bool, a: F32, b: F32) -> F32:
  match c:
    case True{}:
      a
    case False{}:
      b

def Mine.upick(c: Bool, a: U32, b: U32) -> U32:
  match c:
    case True{}:
      a
    case False{}:
      b

# Keys
# ----
# Slots: W A S D Q E Left Right Up Down Space 1 2 3 4. Space is unused
# in flight but kept so the list stays one shape.

def Mine.get(held: List<&2, Bool>, k: Nat) -> Bool:
  match held k:
    case [] _:
      False{}
    case h <> t 0n:
      h
    case h <> t 1n+j:
      Mine.get(t, j)

def Mine.set(held: List<&2, Bool>, k: Nat, v: Bool) -> List<&2, Bool>:
  match held k:
    case [] 0n:
      [v]
    case [] 1n+j:
      False{} <> Mine.set([], j, v)
    case h <> t 0n:
      v <> t
    case h <> t 1n+j:
      h <> Mine.set(t, j, v)

def Mine.slot(code: U32) -> Nat:
  match code:
    case 119:
      0n
    case 97:
      1n
    case 115:
      2n
    case 100:
      3n
    case 113:
      4n
    case 101:
      5n
    case 63234:
      6n
    case 63235:
      7n
    case 63232:
      8n
    case 63233:
      9n
    case 32:
      10n
    case 49:
      11n
    case 50:
      12n
    case 51:
      13n
    case 52:
      14n
    case _:
      15n

def Mine.held(held: List<&2, Bool>, k: Nat) -> F32:
  Mine.pick(Mine.get(held, k), 1.0, 0.0)

def Mine.handof(+code: U32, old: U32) -> U32:
  match code:
    case 49:
      1
    case 50:
      2
    case 51:
      4
    case 52:
      5
    case _:
      old

# World
# -----
# A column is stone, then dirt, then grass. Water fills any cell under
# y = 4 that the land left open. A trunk and a leaf cube sit on a few
# peaks. An edit list overrides a cell.

def Mine.height(+x: U32, +z: U32) -> U32:
  +h = (4.5 + 2.2 * F32.sin((U32.to_f32(x) * 0.45 : F32))
    + 1.6 * F32.cos((U32.to_f32(z) * 0.38 : F32)) : F32)
  F32.to_u32(F32.max(1.0, h))

def Mine.tree(+x: U32, +z: U32) -> Bool:
  U32.is_eq((x % 8 : U32), 3) && U32.is_eq((z % 8 : U32), 5)
    && (x > 2 : U32) && (x < 30 : U32) && (z > 2 : U32) && (z < 30 : U32)

def Mine.land(+x: U32, +y: U32, +z: U32) -> U32:
  +h = Mine.height(x, z)
  +wood = Mine.tree(x, z) && (y > h : U32) && (y < h + 4 : U32)
  +leaf = (Mine.tree(x, z) && U32.is_eq(y, (h + 4 : U32))
    || Mine.tree((x + 1 : U32), z)
      && U32.is_eq(y, (Mine.height((x + 1 : U32), z) + 4 : U32))
    || Mine.tree((x - 1 : U32), z)
      && U32.is_eq(y, (Mine.height((x - 1 : U32), z) + 4 : U32))
    || Mine.tree(x, (z + 1 : U32))
      && U32.is_eq(y, (Mine.height(x, (z + 1 : U32)) + 4 : U32))
    || Mine.tree(x, (z - 1 : U32))
      && U32.is_eq(y, (Mine.height(x, (z - 1 : U32)) + 4 : U32)) : Bool)
  Mine.upick(wood, 4, Mine.upick(leaf, 5,
    Mine.upick(U32.is_eq((y + 1 : U32), h), 1,
      Mine.upick(U32.is_eq((y + 2 : U32), h) || U32.is_eq((y + 3 : U32), h), 2,
        Mine.upick((y + 3 < h : U32), 3, Mine.upick((y < 4 : U32), 6, 0))))))

def Mine.inside(+x: U32, +y: U32, +z: U32) -> Bool:
  (x < 32 : U32) && (y < 16 : U32) && (z < 32 : U32)

def Mine.key(+x: U32, +y: U32, +z: U32) -> String:
  U32.show((x .|. (y << 5n) .|. (z << 9n) : U32))

def Mine.edget.id(miss: Bool, +id: U32) -> Maybe<&2, U32>:
  match miss:
    case True{}:
      None{}
    case False{}:
      Some{id}

def Mine.edget.from(got: Map<&2, U32> & U32) -> Maybe<&2, U32>:
  (m, +id) = got
  Mine.edget.id(U32.is_eq(id, 255), id)

def Mine.edget(
  edits: Map<&2, U32>, +x: U32, +y: U32, +z: U32
) -> Maybe<&2, U32>:
  Mine.edget.from(Map.get(U32, 255, edits, Mine.key(x, y, z)))

def Mine.block.go(got: Maybe<&2, U32>, +x: U32, +y: U32, +z: U32) -> U32:
  match got:
    case Some{id}:
      id
    case None{}:
      Mine.land(x, y, z)

def Mine.block(
  +edits: Map<&2, U32>, +x: U32, +y: U32, +z: U32
) -> U32:
  Mine.upick(Mine.inside(x, y, z),
    Mine.block.go(Mine.edget(edits, x, y, z), x, y, z), 0)

def Mine.solid(+id: U32) -> Bool:
  (id > 0 : U32) && U32.is_ne(id, 6)

# DDA
# ---
# Advance one voxel on the nearest axis. skip drops the cell the eye
# sits in. found ends the walk on the first solid. A miss is a sky.

def Mine.delta(+d: F32) -> F32:
  +a = F32.abs(d)
  Mine.pick((a < 0.0001 : F32), 1000.0, (1.0 / a : F32))

def Mine.tmax(+o: F32, +d: F32) -> F32:
  +cell = F32.floor(o)
  +a = F32.abs(d)
  +next = Mine.pick((d < 0.0 : F32), cell, (cell + 1.0 : F32))
  Mine.pick((a < 0.0001 : F32), 1000.0, ((next - o) / d : F32))

def Mine.dir(+d: F32) -> F32:
  Mine.pick((d < 0.0 : F32), (0.0 - 1.0 : F32), 1.0)

def Mine.faceof(axis: U32, +d: F32) -> U32:
  match axis:
    case 0:
      Mine.upick((d < 0.0 : F32), 1, 0)
    case 1:
      Mine.upick((d < 0.0 : F32), 3, 2)
    case _:
      Mine.upick((d < 0.0 : F32), 5, 4)

def Mine.cell(+p: F32) -> U32:
  F32.to_u32(F32.max(0.0, F32.floor(p)))

def Mine.mkhit(
  yes: Bool, +t: F32, +vx: F32, +vy: F32, +vz: F32, +face: U32, +id: U32
) -> Mine.Hit:
  match yes:
    case True{}:
      At{t, face, Mine.cell(vx), Mine.cell(vy), Mine.cell(vz), id}
    case False{}:
      Miss{}

def Mine.skipor(skip: Bool, +id: U32) -> Bool:
  match skip:
    case True{}:
      False{}
    case False{}:
      Mine.solid(id)

def Mine.advance(
  i: Nat, skip: Bool, found: Bool, hit: Mine.Hit, xy: Bool, xz: Bool,
  yz: Bool, +vx: F32, +vy: F32, +vz: F32, +tx: F32, +ty: F32, +tz: F32,
  +dtx: F32, +dty: F32, +dtz: F32, +sx: F32, +sy: F32, +sz: F32, +dx: F32,
  +dy: F32, +dz: F32, +t: F32, +face: U32, +edits: Map<&2, U32>
) -> Mine.Hit:
  match i:
    case 0n:
      hit
    case 1n+k:
      match found:
        case True{}:
          hit
        case False{}:
          match xy xz:
            case True{} True{}:
              +id = Mine.block(edits, Mine.cell(vx), Mine.cell(vy),
                Mine.cell(vz))
              +now = Mine.skipor(skip, id)
              +ntx = (tx + dtx : F32)
              Mine.advance(k, False{}, now,
                Mine.mkhit(now, t, vx, vy, vz, face, id), (ntx < ty : F32),
                (ntx < tz : F32), (ty < tz : F32), (vx + sx : F32), vy, vz, ntx,
                ty, tz, dtx, dty, dtz, sx, sy, sz, dx, dy, dz, tx,
                Mine.faceof(0, dx), edits)
            case _ _:
              match yz:
                case True{}:
                  +id = Mine.block(edits, Mine.cell(vx), Mine.cell(vy),
                    Mine.cell(vz))
                  +now = Mine.skipor(skip, id)
                  +nty = (ty + dty : F32)
                  Mine.advance(k, False{}, now,
                    Mine.mkhit(now, t, vx, vy, vz, face, id), (tx < nty : F32),
                    (tx < tz : F32), (nty < tz : F32), vx, (vy + sy : F32), vz,
                    tx, nty, tz, dtx, dty, dtz, sx, sy, sz, dx, dy, dz, ty,
                    Mine.faceof(1, dy), edits)
                case False{}:
                  +id = Mine.block(edits, Mine.cell(vx), Mine.cell(vy),
                    Mine.cell(vz))
                  +now = Mine.skipor(skip, id)
                  +ntz = (tz + dtz : F32)
                  Mine.advance(k, False{}, now,
                    Mine.mkhit(now, t, vx, vy, vz, face, id), (tx < ty : F32),
                    (tx < ntz : F32), (ty < ntz : F32), vx, vy, (vz + sz : F32),
                    tx, ty, ntz, dtx, dty, dtz, sx, sy, sz, dx, dy, dz, tz,
                    Mine.faceof(2, dz), edits)

def Mine.trace(
  +ox: F32, +oy: F32, +oz: F32, +dx: F32, +dy: F32, +dz: F32,
  +edits: Map<&2, U32>
) -> Mine.Hit:
  +dtx = Mine.delta(dx)
  +dty = Mine.delta(dy)
  +dtz = Mine.delta(dz)
  +tx = Mine.tmax(ox, dx)
  +ty = Mine.tmax(oy, dy)
  +tz = Mine.tmax(oz, dz)
  Mine.advance(48n, True{}, False{}, Miss{}, (tx < ty : F32), (tx < tz : F32),
    (ty < tz : F32), F32.floor(ox), F32.floor(oy), F32.floor(oz), tx, ty, tz,
    dtx, dty, dtz, Mine.dir(dx), Mine.dir(dy), Mine.dir(dz), dx, dy, dz, 0.0,
    0, edits)

# Color
# -----
# 1 grass, 2 dirt, 3 stone, 4 wood, 5 leaves, 6 water. A face darkens
# the sides; fog mixes the sky in.

def Mine.rgb(+id: U32) -> Mine.Rgb:
  match id:
    case 1:
      Rgb{0.34, 0.62, 0.22}
    case 2:
      Rgb{0.52, 0.35, 0.20}
    case 3:
      Rgb{0.48, 0.48, 0.50}
    case 4:
      Rgb{0.45, 0.28, 0.14}
    case 5:
      Rgb{0.18, 0.46, 0.16}
    case 6:
      Rgb{0.18, 0.38, 0.72}
    case _:
      Rgb{0.90, 0.78, 0.22}

def Mine.shade(+face: U32) -> F32:
  match face:
    case 2:
      1.0
    case 3:
      0.45
    case 0:
      0.80
    case 1:
      0.70
    case 4:
      0.90
    case _:
      0.60

def Mine.pack(+r: F32, +g: F32, +b: F32) -> U32:
  (F32.to_u32((255.0 * F32.clamp(r, 0.0, 1.0) : F32)) << 16n
    .|. F32.to_u32((255.0 * F32.clamp(g, 0.0, 1.0) : F32)) << 8n
    .|. F32.to_u32((255.0 * F32.clamp(b, 0.0, 1.0) : F32)) : U32)

def Mine.sky(+dy: F32) -> U32:
  +u = F32.clamp((dy * 1.6 + 0.15 : F32), 0.0, 1.0)
  Mine.pack((0.55 - 0.20 * u : F32), (0.72 - 0.18 * u : F32),
    (0.95 - 0.05 * u : F32))

def Mine.paint.go(
  +t: F32, +face: U32, +dy: F32, rgb: Mine.Rgb
) -> U32:
  Rgb{+r, +g, +b} = rgb
  +k = Mine.shade(face)
  +fog = F32.clamp((t * 0.035 : F32), 0.0, 1.0)
  +sr = (0.55 - 0.20 * F32.clamp((dy * 1.6 + 0.15 : F32), 0.0, 1.0) : F32)
  +sg = (0.72 - 0.18 * F32.clamp((dy * 1.6 + 0.15 : F32), 0.0, 1.0) : F32)
  +sb = {0.95 : F32}
  Mine.pack((r * k + (sr - r * k) * fog : F32),
    (g * k + (sg - g * k) * fog : F32),
    (b * k + (sb - b * k) * fog : F32))

def Mine.paint(hit: Mine.Hit, +dy: F32) -> U32:
  match hit:
    case Miss{}:
      Mine.sky(dy)
    case At{+t, +face, bx, by, bz, +id}:
      Mine.paint.go(t, face, dy, Mine.rgb(id))

# Aim
# ---
# A click traces the same ray as the centre pixel. Break writes air on
# the hit cell; place writes the hand on the face next to it.

def Mine.camof(
  +x: F32, +y: F32, +z: F32, +yaw: F32, +pitch: F32,
  +edits: Map<&2, U32>, +hand: U32, +fps: U32
) -> Mine.Cam:
  +sy = F32.sin(yaw)
  +cy = F32.cos(yaw)
  +sp = F32.sin(pitch)
  +cp = F32.cos(pitch)
  Cam{x, y, z, (sy * cp : F32), sp, (cy * cp : F32), cy, (0.0 - sy : F32),
    (0.0 - sp * sy : F32), cp, (0.0 - sp * cy : F32), hand, fps, edits}

def Mine.look(+s: Mine) -> Mine.Cam:
  Mine{+x, +y, +z, +yaw, +pitch, held, +edits, +hand, t0, frames, +fps} = s
  Mine.camof(x, y, z, yaw, pitch, edits, hand, fps)

def Mine.beside(+bx: U32, +by: U32, +bz: U32, face: U32) -> Mine.Pos:
  match face:
    case 0:
      Pos{(bx + 1 : U32), by, bz}
    case 1:
      Pos{(bx - 1 : U32), by, bz}
    case 2:
      Pos{bx, (by + 1 : U32), bz}
    case 3:
      Pos{bx, (by - 1 : U32), bz}
    case 4:
      Pos{bx, by, (bz + 1 : U32)}
    case _:
      Pos{bx, by, (bz - 1 : U32)}

def Mine.put(
  edits: Map<&2, U32>, +x: U32, +y: U32, +z: U32, +id: U32
) -> Map<&2, U32>:
  Map.set(&2, U32, edits, Mine.key(x, y, z), id)

def Mine.breakat(hit: Mine.Hit, edits: Map<&2, U32>) ->
  Map<&2, U32>:
  match hit:
    case Miss{}:
      edits
    case At{t, face, +bx, +by, +bz, id}:
      Mine.put(edits, bx, by, bz, 0)

def Mine.place.go(
  ok: Bool, edits: Map<&2, U32>, +px: U32, +py: U32, +pz: U32,
  +hand: U32
) -> Map<&2, U32>:
  match ok:
    case False{}:
      edits
    case True{}:
      Mine.put(edits, px, py, pz, hand)

def Mine.place.at(
  nxt: Mine.Pos, edits: Map<&2, U32>, +hand: U32
) -> Map<&2, U32>:
  Pos{+px, +py, +pz} = nxt
  Mine.place.go(Mine.inside(px, py, pz), edits, px, py, pz, hand)

def Mine.placeat(
  hit: Mine.Hit, edits: Map<&2, U32>, +hand: U32
) -> Map<&2, U32>:
  match hit:
    case Miss{}:
      edits
    case At{t, +face, +bx, +by, +bz, id}:
      Mine.place.at(Mine.beside(bx, by, bz, face), edits, hand)

def Mine.aim.go(cam: Mine.Cam) -> Mine.Hit:
  Cam{+ex, +ey, +ez, +fx, +fy, +fz, rx, rz, ux, uy, uz, hand, fps, +edits} = cam
  Mine.trace(ex, ey, ez, fx, fy, fz, edits)

def Mine.aim(+s: Mine) -> Mine.Hit:
  Mine.aim.go(Mine.look(s))

def Mine.click.go(
  left: Bool, hit: Mine.Hit, x: F32, y: F32, z: F32, yaw: F32, pitch: F32,
  held: List<&2, Bool>, edits: Map<&2, U32>, +hand: U32, t0: Nat,
  frames: U32, fps: U32
) -> Mine:
  match left:
    case True{}:
      Mine{x, y, z, yaw, pitch, held, Mine.breakat(hit, edits), hand, t0,
        frames, fps}
    case False{}:
      Mine{x, y, z, yaw, pitch, held, Mine.placeat(hit, edits, hand), hand, t0,
        frames, fps}

def Mine.click.of(left: Bool, s: Mine) -> Mine:
  Mine{+x, +y, +z, +yaw, +pitch, held, +edits, +hand, t0, frames, +fps} = s
  Mine.click.go(left, Mine.aim.go(Mine.camof(x, y, z, yaw, pitch, edits, hand,
    fps)), x, y, z, yaw, pitch, held, edits, hand, t0, frames, fps)

def Mine.click(+button: U32, s: Mine) -> Mine:
  Mine.click.of(U32.is_eq(button, 0), s)

# Tick
# ----

def Mine.key.esc(down: Bool, s: Mine) -> Maybe<Mine>:
  match down:
    case True{}:
      None{}
    case False{}:
      Some{s}

def Mine.key.go(
  esc: Bool, +code: U32, +down: Bool, s: Mine
) -> Maybe<Mine>:
  match esc:
    case True{}:
      Mine.key.esc(down, s)
    case False{}:
      Mine{x, y, z, yaw, pitch, held, edits, hand, t0, frames, fps} = s
      Some{Mine{x, y, z, yaw, pitch, Mine.set(held, Mine.slot(code), down),
        edits, Mine.handof(code, hand), t0, frames, fps}}

def Mine.mouse(down: Bool, +button: U32, s: Mine) -> Maybe<Mine>:
  match down:
    case True{}:
      Some{Mine.click(button, s)}
    case False{}:
      Some{s}

def Mine.event(e: Event, s: Mine) -> Maybe<Mine>:
  match e:
    case Key{+code, +down}:
      Mine.key.go(U32.is_eq(code, 27), code, down, s)
    case Close{}:
      None{}
    case Mouse{x, y, +button, down}:
      Mine.mouse(down, button, s)
    case _:
      Some{s}

def Mine.events(events: List<Event>, s: Mine) -> Maybe<Mine>:
  match events:
    case []:
      Some{s}
    case e <> rest:
      Maybe.bind(&1, Mine, Mine, Mine.event(e, s), s2 => Mine.events(rest, s2))

def Mine.step(+s: Mine) -> Mine:
  Mine{x, y, z, yaw, pitch, +held, edits, hand, t0, frames, fps} = s
  +fwd = (Mine.held(held, 0n) - Mine.held(held, 2n) : F32)
  +side = (Mine.held(held, 3n) - Mine.held(held, 1n) : F32)
  +rise = (Mine.held(held, 4n) - Mine.held(held, 5n) : F32)
  +turn = (Mine.held(held, 7n) - Mine.held(held, 6n) : F32)
  +tilt = (Mine.held(held, 8n) - Mine.held(held, 9n) : F32)
  +yaw2 = (yaw + 0.035 * turn : F32)
  Mine{(x + 0.10 * (fwd * F32.sin(yaw2) + side * F32.cos(yaw2)) : F32),
    F32.clamp((y + 0.10 * rise : F32), 0.4, 15.4),
    (z + 0.10 * (fwd * F32.cos(yaw2) - side * F32.sin(yaw2)) : F32), yaw2,
    F32.clamp((pitch + 0.025 * tilt : F32), (0.0 - 1.4 : F32), 1.4), held,
    edits, hand, t0, frames, fps}

def Mine.count(within: Bool, now: Nat, s: Mine) -> Mine:
  match within:
    case True{}:
      s
    case False{}:
      Mine{x, y, z, yaw, pitch, held, edits, hand, t0, frames, fps} = s
      Mine{x, y, z, yaw, pitch, held, edits, hand, now, 0, frames}

def Mine.clock(+now: Nat, +s: Mine) -> Mine:
  Mine{x, y, z, yaw, pitch, held, edits, hand, t0, frames, fps} = s
  Mine.count(Nat.is_lt(Nat.sub(now, t0), 1000n), now,
    Mine{x, y, z, yaw, pitch, held, edits, hand, t0, (frames + 1 : U32), fps})

def Mine.tick(events: List<Event>, s: Mine) -> IO(Maybe<Mine>):
  do IO<Maybe<Mine>>:
    now : Nat <- IO.now()
    return Maybe.bind(&1, Mine, Mine, Mine.events(events, s),
      s2 => Some{Mine.clock(now, Mine.step(s2))})

# Draw
# ----

def Mine.glyph(d: Nat) -> U32:
  match d:
    case 0n:
      480599
    case 1n:
      467506
    case 2n:
      464711
    case 3n:
      476999
    case 4n:
      280405
    case 5n:
      476951
    case 6n:
      481047
    case 7n:
      279623
    case 8n:
      481111
    case 9n:
      477015
    case _:
      0

def Mine.lead(blank: Bool, d: U32) -> U32:
  match blank:
    case True{}:
      10
    case False{}:
      d

def Mine.digit(i: Nat, +fps: U32) -> U32:
  match i:
    case 0n:
      Mine.lead((fps < 100 : U32), (fps / 100 : U32))
    case 1n:
      Mine.lead((fps < 10 : U32), (fps / 10 % 10 : U32))
    case _:
      (fps % 10 : U32)

def Mine.hud(+x: U32, +y: U32, +fps: U32) -> U32:
  +bit = U32.to_nat((y / 8 * 4 + x % 32 / 8 : U32))
  (16777215 * (Mine.glyph(U32.to_nat(Mine.digit(U32.to_nat((x / 32 : U32)),
    fps))) >> bit .&. 1) : U32)

def Mine.bar.go(sel: Bool, rgb: Mine.Rgb) -> U32:
  Rgb{+r, +g, +b} = rgb
  Mine.upick(sel, 16777215, Mine.pack(r, g, b))

def Mine.bar(+x: U32, +hand: U32) -> U32:
  +i = (x / 40 : U32)
  +id = Mine.upick(U32.is_eq(i, 0), 1, Mine.upick(U32.is_eq(i, 1), 2,
    Mine.upick(U32.is_eq(i, 2), 4, 5)))
  Mine.bar.go(U32.is_eq(id, hand), Mine.rgb(id))

def Mine.cross(+x: U32, +y: U32) -> Bool:
  +dx = Mine.upick((x < 256 : U32), (256 - x : U32), (x - 256 : U32))
  +dy = Mine.upick((y < 256 : U32), (256 - y : U32), (y - 256 : U32))
  U32.is_eq(dy, 0) && (dx < 8 : U32) || U32.is_eq(dx, 0) && (dy < 8 : U32)

def Mine.cast(+x: U32, +y: U32, +cam: Mine.Cam) -> U32:
  Cam{+ex, +ey, +ez, +fx, +fy, +fz, +rx, +rz, +ux, +uy, +uz, hand, fps,
    +edits} = cam
  +sx = ((U32.to_f32(x) - 255.5) / 320.0 : F32)
  +sy = ((255.5 - U32.to_f32(y)) / 320.0 : F32)
  +vx = (fx + sx * rx + sy * ux : F32)
  +vy = (fy + sy * uy : F32)
  +vz = (fz + sx * rz + sy * uz : F32)
  +il = (1.0 / F32.sqrt((vx * vx + vy * vy + vz * vz : F32)) : F32)
  Mine.paint(Mine.trace(ex, ey, ez, (vx * il : F32), (vy * il : F32),
    (vz * il : F32), edits), (vy * il : F32))

def Mine.over.bar(bar: Bool, +x: U32, +y: U32, +cam: Mine.Cam) -> U32:
  match bar:
    case True{}:
      Cam{ex, ey, ez, fx, fy, fz, rx, rz, ux, uy, uz, +hand, fps, edits} = cam
      Mine.bar(x, hand)
    case False{}:
      Mine.upick(Mine.cross(x, y), 16777215, Mine.cast(x, y, cam))

def Mine.over.hud(hud: Bool, +x: U32, +y: U32, +cam: Mine.Cam) -> U32:
  match hud:
    case True{}:
      Cam{ex, ey, ez, fx, fy, fz, rx, rz, ux, uy, uz, hand, +fps, edits} = cam
      Mine.hud(x, y, fps)
    case False{}:
      Mine.over.bar(((y > 496 : U32) && (x < 160 : U32)), x, y, cam)

def Mine.over(+x: U32, +y: U32, +cam: Mine.Cam) -> U32:
  Mine.over.hud(((x < 96) && (y < 40) : U32), x, y, cam)

def Mine.scene(+d: Nat, +x: U32, +y: U32, +cam: Mine.Cam) -> Image:
  match d:
    case 0n:
      Pix{Mine.over(x, y, cam)}
    case 1n+e:
      +h = U32.shln(1, e)
      +x1 = U32.add(x, h)
      +y1 = U32.add(y, h)
      tl tr bl br = Mine.scene(e, x, y, cam) Mine.scene(e, x1, y, cam)
        Mine.scene(e, x, y1, cam) Mine.scene(e, x1, y1, cam)
      Qua{tl, tr, bl, br}

def Mine.view(s: Mine) -> Mine & Image:
  +s2 = s
  (s2, Mine.scene!(9n, 0, 0, Mine.look(s2)))

# A column of the world as digits, for the ASCII slice and the laws.

def Mine.column.go(
  +edits: Map<&2, U32>, +x: U32, +z: U32, y: Nat, acc: String
) -> String:
  match y:
    case 0n:
      acc
    case 1n+p:
      +p = p
      Mine.column.go(edits, x, z, p,
        acc ++ U32.show(Mine.block(edits, x, U32.from_nat(p), z)))

def Mine.column(+edits: Map<&2, U32>, +x: U32, +z: U32) -> String:
  Mine.column.go(edits, x, z, 16n, "")

def Mine.row.go(+edits: Map<&2, U32>, +z: U32, x: Nat, acc: String) ->
  String:
  match x:
    case 0n:
      acc
    case 1n+p:
      +p = p
      Mine.row.go(edits, z, p, acc ++ Mine.column(edits, U32.from_nat(p), z)
        ++ " ")

def Mine.slice(+edits: Map<&2, U32>, +z: U32) -> String:
  Mine.row.go(edits, z, 32n, "")

def main() -> IO(Unit):
  App.run(~Mine, ~App{Mine.view, Mine.tick}, "Mine", 512, 512, Mine.init())
