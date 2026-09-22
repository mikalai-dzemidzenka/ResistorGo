## Why

Two fighters swinging low at each other in the same instant today resolve as
whichever one the update loop reaches first: its swing throws the other into a
low block, and a moment of mutual aggression reads as one fighter guarding.
`battle-low-1-left.png` and `battle-low-1-right.png` were drawn for that moment
instead - a short exchange of blows, a lunging punch answered by a knee, that
both fighters play out together.

## What Changes

- A new actor state, `StateClash`, appended to the state enum. When two
  opposing fighters facing each other, each within the other's reach, open a
  low swing within 6 ticks (100 ms) of one another, both enter it together.
  The second swing can come from a fighter already thrown into a low block by
  the first: that block is where the clash is caught.
- Only an opening low swing counts: not a chained blow, not a change of height
  mid-chain, not mid or high. Clashes at other heights are left for their own
  animations.
- The fighter standing on the left plays `battle-low-1-left.png` (9 frames),
  the one on the right plays `battle-low-1-right.png` (10 frames) mirrored to
  face left. Both run for the same time and end together, and each fighter goes
  back to idle.
- A clash costs each fighter 0.5 structure and one swing's energy (12): the
  swing that opened it, or, for a fighter caught from a block, a swing charged
  at the clash. It is not a hit: no flash, no hit-stop, no knockback, no
  flinch or stun, no hit counted and no run of hits extended. Structure broken
  by a clash kills as any structure break does, the other fighter taking the
  kill.
- While clashing, neither fighter can be hit, blocked against or shoved by
  anyone else: a third fighter's swing passes through them.
- Clashes are the authority's alone, as hits are. A client sees one in the
  snapshot.
- Both sheets are appended to the animation table and registered as textures.

## Capabilities

### New Capabilities
- `clash`: two fighters opening low swings at each other at the same moment -
  what starts a clash, who plays which side, what it costs, what it protects
  from, and who decides it.

### Modified Capabilities

<!-- none: the clash is a new outcome of two swings, and changes no existing
     hit, block, structure or energy rule -->

## Impact

`actor.go` carries the state, its constants, a record of the tick each swing
opened, the clash test in the attack and block updates, and the exclusions in
hit resolution and the shove. `resources.go`'s animation table and
`LoadTextures` gain the two sheets; both the table and the state enum are
append-only because their values travel in every snapshot. Time manipulation
records the new state with the rest of the fighter and needs nothing.

Multiplayer: the server decides the clash; a predicting client's own fighter
is shown its swing for the round trip before the snapshot puts it into the
clash.
