## Context

See proposal.md for why. What exists to build on:

- `takeHit` raises `EvHit`, but nothing consumes it, and events are never sent
  over the network. What every machine does see is each fighter's
  `Structure` (replicated in tenths), its `State`, and its `Attacker`
  (replicated as `AttackerID`).
- `renderLevel` draws the level (`LE.Draw()`), then the fighters, inside
  `ScreenFullTexture` in world space, then `compositeLights()` multiplies the
  scene by the light map, then overlays that must not be lit. Anything drawn
  before `compositeLights()` is lit.
- `LE.solids` is every collision rectangle in world space.
- The level palette carries four reds: (87,28,39), (137,30,43), (196,36,48),
  (234,50,60).
- The whole-match rewind restores `World` from a recorded frame, including
  `Tick`; `rewindResumeHere` makes the shown frame the newest.

## Goals / Non-Goals

**Goals:**
- Blood from what every machine already knows, with no wire change and no
  effect on the simulation.
- Stains that follow time manipulation without being copied into 3600 frames.
- Bounded memory and draw cost however long the match.

**Non-Goals:**
- Blood on fighters' sprites, pools that spread, drips, splatter sprites.
- Sound.
- Blood that matches between machines.
- Blood in the menu or on the dedicated server.

## Decisions

### Detect hits from structure drops, not from events

Each rendered frame, for each fighter, the blood keeps what it last saw of
that fighter (`structure`, `state`, guard height, keyed by the `*Actor`). A
drop in structure is a landed hit unless the fighter was guarding everything
before it fell, or is clashing:

```
  drop, and the guard seen before was not block-all, and state is not Clash:
    now Death     -> kill spray    (60 droplets)
    now Attacked  -> stun spray    (30)
    otherwise     -> flinch spray  (10)
```

Guarded blocks raise structure; block-all's cost and a death from it happen to
a fighter guarding everything, which nothing can hit; a clash's cost happens
in `StateClash`; a shove costs no structure. Rising structure (regeneration, a
kill's reward) is ignored.

Two versions of this rule were wrong and are recorded here so they are not
tried again. Excluding a drop out of `StateBlock` misses a real hit: a
fighter thrown into a block at the wrong height is hit straight out of it.
Excluding a fighter guarding everything *after* the drop misses real hits
too: a bot that is hit moves its guard to the attacker's height a third of the
time, in the same instant, and a fighter can reach block-all within the same
rendered frame. The guard seen before the drop is the one that decides.

Alternative considered: replicate `EvHit`. It would need a wire change and a
way to not lose events between snapshots, to say what structure already says.

### Spray away from the attacker

The spray's direction is away from `Attacker` along x (from the attacker's
side, or its facing when they stand on the same column), falling back to
against the hit fighter's facing when there is no attacker. Droplets leave
from the middle of the fighter's collision box with a random speed in a cone
tilted upward - mostly away, some straight up, a few back - so a hit reads as
directional without looking like a hose.

### Droplets fly in the render frame; stains are stamped with the world tick

```
  droplet { pos, vel, color }          at most 512; the oldest dropped past it
  stain   { x, y int16; color; tick }  a ring of 4096, oldest overwritten
```

Droplets step with the rendered frame's `dt` under gravity, and not at all
while `matchHeld()` (edit mode or time manipulation). Each step tests the
droplet's pixel against `LE.solids`; on entering one it becomes a stain at its
last free pixel, stamped with `CL.Tick`. A droplet older than two seconds, or
past the level's lowest solid, is dropped.

Blood is not in the fixed step and not in `World`: it is the view's, the
rewind needs only the stamp, and on a client there is no authoritative step to
hang it on.

### Splashing onto the backdrop

Each droplet is given a reach when sprayed: a flight time between
`BloodReachMin` and `BloodReachMax` (0.04 to 0.3 s). When it has flown that
long, it asks whether the pixel under it is painted on any level tile; if so
it sticks there as a streak and stops, and if not it flies on as before and
asks again each step. A droplet that meets a solid first lands on it as a
one-pixel stain, as before.

"Painted" is a real pixel test, not the tile's square. `loadTexBytes` already
decodes every texture on the CPU before uploading it; it now also keeps an
opacity mask for it (one bit a pixel, alpha over half), in a map by resource
name. A tile's pixel is painted when the pixel under the droplet, relative to
the tile's origin, is set in its texture's mask. The level has about 90
tiles, so the droplets that are ready to splash walk `LE.objectsIndexed`
directly; nothing has to be rebuilt when the editor changes the level.

A streak is a stain with a tail: it keeps its head pixel and an offset of 1
to 4 pixels back along the droplet's flight (so 2 to 5 pixels long), and is
drawn as a line of pixels, the head in the droplet's red and the tail one red
darker. It counts once against the stain cap. The rewind treats it as any
stain: it has a tick.

The tiles are drawn in one pass and blood over all of them, so a streak on a
backdrop tile that a decorative, non-solid tile covers in front would show
over that tile. The level has no such overlaps where fights happen; it is
left as it is.

### Rewind hides, resume deletes

Copying stains into every recorded frame would cost 4096 x 8 bytes x 3600
frames. Stains never move once stuck, so the stamp is enough:

- While a recorded frame is shown, stains are drawn only if `tick <= CL.Tick`
  (the restored frame's tick). Scrubbing back hides them; scrubbing forward
  shows them again. Droplets are cleared on entering the mode and neither
  fly nor spawn while it is on.
- On the first live frame after a recorded one - leaving the mode, which
  `rewindResumeHere` makes the newest frame - stains with `tick > CL.Tick` are
  deleted: that future is gone, and ticks from here on count up from the
  resumed frame. The blood sees the change for itself in `updateBlood`, so
  `rewind.go` needs no hook.
- The blood's memory of each fighter's structure and state is reseeded from
  the restored world whenever a frame is restored and on leaving the mode, so
  structure that rises or falls through the rewind does not read as a hit.

Stains are appended in tick order, so deleting from the newest end is enough.
A stain overwritten by the cap before a rewind is not brought back; at 4096
stains that is only ever old blood.

### Drawing

Stains are drawn right after `LE.Draw()`, droplets right after the fighters,
both before `compositeLights()` so they are lit. Each is one level pixel,
`DrawRectangle` with a size of 1, which raylib batches: 4096 quads a frame is
well inside what it draws already. Only those inside the camera's view are
drawn.

`startOffline`, `restartOffline` and `leaveMatch` clear the blood; entering a
networked match starts it empty.

### Constants

```
  BloodFlinch = 10, BloodStun = 30, BloodKill = 60   droplets per spray
  BloodSpeedMin/Max = 40 / 140 px/s                  launch speed
  BloodGravity = 400 px/s^2                          droplets fall faster than fighters feel
  BloodDroplets = 512, BloodStains = 4096            caps
  BloodLife = 2 s                                    a droplet not landed by then is dropped
  BloodReachMin/Max = 0.04 / 0.3 s                   flight before splashing on the backdrop
  BloodStreakMax = 4                                 tail pixels behind a streak's head
```

All are feel numbers, tuned in play.

## Risks / Trade-offs

- [A structure drop that is not a hit] Anything new that spends structure
  other than block-all or a clash will spray. → The rule is written in one
  function, `bled`, with what it excludes and why.
- [A clash that kills] A clash that breaks a fighter's structure kills it out
  of its swing or block, before either enters `StateClash`, so that one death
  sprays as a kill. It needs a fighter with half a hit of structure left to
  clash, and is left as it is.
- [Tunnelling] A fast droplet can step past a thin solid in one frame. → Step
  a droplet in sub-steps of at most one pixel of travel.
- [Client ticks] A client's `Tick` follows its own world, and a client never
  rewinds, so its stamps are only ever compared with themselves.
- [Random spray] Blood uses its own random source, not the global one the
  bots draw from, so turning blood on cannot change what a bot does.
