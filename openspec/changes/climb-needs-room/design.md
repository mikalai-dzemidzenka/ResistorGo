## Context

`checkCollision` visits `LE.solids` lower rectangles first (Y descending, then
X descending) and calls `catchLedge` on every solid the fighter does not stand
on or hit with its head. `catchLedge` snaps `Pos` onto the ledge as soon as the
head is anywhere inside the rectangle's height band and the edge is within
reach. The only guard against ceilings is `hitTop`, from `headBlocked`: whether
the head, at the column the fighter started the tick in, is inside a solid.

Two things go wrong in a T (see proposal.md - Why):

```
  bar  | bar  | bar          tick: stem visited first -> snapped to its top,
 ------+------+------               body now inside the bar
       | stem |                     bar tile over stem -> head now in its band
  [F]->| stem |                     -> snapped again, on top of the bar
```

- nothing checks the spot the snap puts the fighter in;
- the snap changes `Pos.Y` mid-loop, and later solids are tested against the
  new height, so grabs chain up a column. The in-loop head check that could
  catch it uses the old `pColCenterX` and the new Y, and misses.

## Goals / Non-Goals

**Goals:**
- No ledge is caught where the fighter would stand inside a solid.
- Grabs no longer chain within a tick, as a consequence rather than a second
  rule.

**Non-Goals:**
- Changing the grab band, reach (`climbOffset`) or snap positions.
- Teaching bots about ceilings over walls: `wallAhead` walks the column up
  through the T's bar and a bot may try the jump and fail. A bot in a T is an
  edge case no level has today.
- Reworking the in-loop head check's stale `pColCenterX`.

## Decisions

**Room check in `catchLedge`, before it commits.** Work out the collision box
the snap would give - X from the same `c.X - 20` / `c.X + c.Width - 24`, feet on
`c.Y` - and refuse the ledge if it overlaps any solid that is not
drop-through. `catchLedge` then only changes `Pos` and `CanClimb*` when it
returns true.

- The box is taken with its feet exactly on `c.Y` (height 42 ending at `c.Y`),
  not the 1 px deeper the snap really leaves it, and overlaps are strict. That
  keeps `c` itself and floor tiles level with it beside the ledge from counting
  as in the way.
- A lower tile of a stacked wall is now refused (the upper one is in its
  standing spot), and the upper one is caught once the head rises into its band
  - about 6 px of rise for a wall two tiles tall from the floor. That replaces
  the cascade with the same end position.

Alternative: stop after the first grab in a tick (`break` on a true
`catchLedge`). Stops the chaining but still snaps the fighter into the bar of a
T; rejected as a symptom fix.

Alternative: extend `headBlocked` to look above the ledge. It answers a
different question (the head where it is now) and would still miss ceilings
made of several rectangles; the destination box covers every shape.

**Keep `headBlocked`.** It still does its own job - a fighter whose head is in
a ceiling row is pushed down rather than grabbing the neighbour - and
`TestNoClimbThroughCeiling` covers it. Removing it is a separate question.

## Risks / Trade-offs

- [A level with a ledge that only fit by a pixel] → the box uses the exact
  collision box size and strict overlap, so anything a fighter could stand in
  still qualifies. Playtest `level_1.json` and the bot course.
- [Walls climbed a few ticks later than before] → only walls of more than one
  rectangle, and only a few pixels of rise; the bot course test covers the high
  ledge.
- [Prediction mismatch online] → none: collision is pure over position, speed
  and the shared level, as today.
