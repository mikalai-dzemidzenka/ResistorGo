## Why

A fighter can climb up through a ceiling. Jump beside the stem of a T - a wall
with a ceiling bar resting on it - and the fighter grabs the stem's ledge, is
snapped into the bar, grabs the bar's ledge on the same tick and ends up
standing on top of the T. A ledge is caught without asking whether there is
room to stand on it, and one grab moves the fighter far enough to set off the
next.

## What Changes

- A ledge is caught only where the fighter's body fits standing on top of it.
  A ledge with a solid over it - the stem of a T, the lower tile of a stacked
  wall - is not a ledge.
- With that, one tick can no longer chain grabs up a column. A wall two tiles
  tall is still climbed, onto its top, once the head reaches the top tile: a
  few pixels of rise later than today, from standing beside it.
- Drop-through rectangles are not in the way: a ledge under a drop-through
  platform is still caught, as the platform does not stop a fighter from below.

## Capabilities

### New Capabilities

### Modified Capabilities
- `level-collision`: a ledge needs room above it to be caught.

## Impact

- `actor.go`: `catchLedge` checks the spot it would snap the fighter to against
  the level's solids before taking hold. `headBlocked` stays.
- `actor_test.go`: a T-shaped test next to `TestNoClimbThroughCeiling`.
- Bots: `clears`/`wallAhead` in `bot.go` already model a climb as the head
  reaching the top of the whole column, so they are unchanged. A bot will now
  fail to climb a wall with a ceiling over it rather than pass through it.
- Online: collision runs the same on client and server; no protocol change.
