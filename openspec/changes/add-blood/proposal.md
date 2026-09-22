## Why

A landed hit today is a flash, a nudge and a bar going down. Nothing in the
world remembers the fight: once it is over the floor is as clean as before
it. Katana Zero shows what blood adds to a fighter of this size - a spray
thrown away from the blow, droplets arcing and landing, and stains left on the
floor and walls that tell where each fight happened. At 160x90 a droplet is a
single pixel, which is exactly the scale that works.

## What Changes

- A landed hit sprays a handful of blood droplets from the fighter it landed
  on, thrown away from the attacker; a stun sprays more, and a death the most.
  Blocks, block-all, clashes and shoves spill nothing.
- Droplets fly a short, random distance and splash onto the wall behind the
  fight, Katana Zero style: where one stops over a painted pixel of a level
  tile it sticks as a short streak along its flight. Over empty background or
  a tile's transparent gaps it flies on under gravity, and where it meets the
  floor, a solid wall or a ceiling it sticks as a one-pixel stain. Everything
  stays for the rest of the match. Droplets that leave the level are dropped.
  The number of stains is capped; past the cap the oldest go first.
- Blood is drawn in the level's own reds (already in the palette), inside the
  lit scene, so shadows darken it like everything else; stains under the
  fighters, droplets over them.
- Time manipulation takes blood back with the match: a frame restored shows
  exactly the stains that had landed by then, and nothing sprays while the
  recording is shown.
- Blood is presentation only and each machine makes its own: a client sprays
  from the structure drops its snapshots show, so nothing new is sent and the
  simulation is untouched. Stains need not match between machines.

## Capabilities

### New Capabilities
- `blood`: what spills blood and how much, how it flies and where it stays,
  how it is drawn, and how it follows time manipulation.

### Modified Capabilities

<!-- none: hits, structure and time manipulation keep their requirements; the
     blood's own rewind behaviour is stated in the new capability -->

## Impact

A new `blood.go` holds the droplets, the stains and their drawing. `game.go`
calls it once per rendered frame (spawn from structure drops, fly, land) and
draws it in `renderLevel`. `rewind.go`'s restore truncates the stains to the
frame's tick and clears the droplets; entering a match or restarting clears
both. `World` gains nothing; the blood belongs to the client-side view of the
match. No change to the wire, the dedicated server or the fixed step.
