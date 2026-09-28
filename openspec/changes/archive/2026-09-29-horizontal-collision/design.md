## Context

See `proposal.md` — Why. What the approach has to work with:

- **Collision today.** `checkCollision` runs at the start of `Actor.Update`, before speeds and state. It tests only the middle X of the collision box against `LE.solids`, the world-space collision rectangles that the in-flight `multiple-collision-rects` change adds. With the middle inside a rectangle and the feet within its height, the fighter is snapped onto its top. That snap is both landing and today's accidental walking up. With the head inside, it hits a ceiling. A rectangle edge within 32 px of the middle, while airborne and facing it, catches a ledge and places the fighter at `oColX - 20` (right) or `oColX + width - 24` (left), 1 px below the top.
- **The fighter's box** is `CollisionBox{X: 14, Y: 6, W: 14, H: 42}` relative to `Pos`, so the feet are at `Pos.Y + 48`.
- **Where X changes.** In a normal tick, `Update` ends with `Pos += Speed*dt` and then truncates both axes to whole pixels. The swing's dash, `Pos.X += Orientation*AttackDash` in `enterAttackState`, is the other move. The Idle branch returns straight after it, before the end of `Update`. The ledge snap in `checkCollision` also writes `Pos.X`, but that places the fighter; it is not a move.
- **How far a tick moves.** `MaxSpeed` is 320, about 5.3 px per tick. `StunKnockback` is 150 and `AttackDash` is 4. No tick moves a fighter as far as its 14 px box is wide.
- **Determinism.** The client's prediction and the server both run `Update` over the same `LE.solids`, and rewind and lag compensation replay recorded positions. Anything decided inside `Update` from the level and the fighter's own state is predicted exactly.

## Goals / Non-Goals

**Goals:**

- One side check, used by every sideways move, that cannot be tunnelled through at any speed the game has.
- The existing floor, ceiling, landing and ledge code is left as it is: the side check runs alongside it, not instead of it.
- No fighter is ever moved by collision to somewhere it was not already heading.

**Non-Goals:**

- Sweeping the vertical axis, or testing floors and ceilings across the whole box.
- Sliding along a wall's top corner, or rounding corners.

## Decisions

### Stop the move at the face it crosses, rather than pushing out of overlaps

`blockSides(fromX)` takes the X the fighter moved from, with `Pos.X` already at where it moved to. Moving right, the box's right side goes from `fromRight` to `toRight`. For each wall ahead of the fighter, meaning one whose left face is beyond the box's left side, the right side is limited to `max(wall.X, fromRight)`. Moving left is the mirror image. If the move was cut short, `Pos.X` becomes the limit less the box offset, and `Speed.X` is zeroed if it points into the wall.

Two things follow from `max(wall.X, fromRight)`:

- **A fighter outside a wall stops at its face.**
- **A fighter already in a wall never goes deeper, and is never pushed.** That is the spec's *never pushed out* requirement, and it is why a spawn or a ledge snap that leaves the box overlapping a rectangle cannot throw the fighter across the screen.

Because the test is whether the side crossed the face during the tick, not whether the box overlaps a rectangle afterwards, it is independent of the wall's thickness. So a one-pixel wall stops a fighter at `MaxSpeed`.

*Alternative considered:* push the fighter out of any overlap along the shorter axis. Rejected. It moves fighters that were not heading anywhere: a spawn inside the upper platform, or the climb pose, which sits 8 px inside the rectangle it climbed. It also needs its own rule to tell a floor from a wall.

*Alternative considered:* sub-stepping the move one pixel at a time. Rejected. It costs up to five collision passes a tick per fighter to get the same answer the face test gives in one.

### What counts as a wall

A solid is a wall to the fighter when it overlaps the fighter's body band vertically: `wall.Y < feet - StepHeight` and `wall.Y + wall.Height > head`, with `head = Pos.Y + CollisionBox.Y` and `feet = head + CollisionBox.Height`. `StepHeight` is 4 px, a new constant next to the other movement constants.

- **A rectangle whose top is within 4 px of the feet** is let through. The fighter walks on until its middle is over it, and on the next tick `checkCollision` snaps it up onto it. That is today's behaviour, now limited to steps.
- **The floor the fighter stands on**, with its top level with the feet, is never a wall. Neither is a neighbouring floor tile with the same top. `level_1.json`'s tiles overlap their neighbours by a pixel at the same height.

The band uses the fighter's Y at the time of the X move, before this tick's vertical move.

### Where the check runs

- **The end of `Update`.** The one line moving both axes becomes X then Y: `Pos.X += Speed.X*dt`, truncate X, `blockSides(fromX)`, then `Pos.Y += Speed.Y*dt`, truncate Y. Truncating each axis separately gives the same numbers as before, since neither reads the other. Doing X before Y means a fighter falling past a wall's top corner is judged at the height it was at when it moved sideways.
- **`enterAttackState`**, around the dash: record X, dash, `blockSides(fromX)`. It is inside `enterAttackState` rather than at each of its callers, so a swing started from idle, a run or a landing is covered, including the idle branch that returns early.

The walls are the rectangles in `LE.solids`, which is what `checkCollision` already walks. The clamped position is a whole pixel, because solids are whole pixels and the box offsets are too. So truncation never moves it again.

### Ledges are left to `checkCollision`

A fighter stopped by a wall ends its tick with the box's side on the wall's face. Its middle is then 7 px from the face, well within the 32 px `climbOffset`. So on the next tick `checkCollision` catches the ledge exactly as it would at a tile's edge, whenever the head is within the wall's height, the fighter is airborne and it faces the wall. The ledge snap then places the fighter 8 px into the rectangle, which the *never deeper* rule tolerates, and the climb proceeds as now.

## Risks / Trade-offs

- **[A fighter held against a wall plays its walk animation in place.]** Each tick the input adds `Acc*dt` of speed, and the wall takes it away. → Acceptable for now; check it by eye. If it looks wrong, the Idle branch can read "blocked" as standing still, in a follow-up.
- **[Bots walk into walls and stay there.]** `botInput` heads straight at its target. → No pathfinding in this change. The current level has no walls between spawns.
- **[The ledge catch reaches down a tall wall.]** It fires whenever the head is anywhere within a rectangle's height. So jumping beside a tall one-rectangle wall snaps the fighter to its top from part-way down. → Unchanged by this change, and recorded as a non-goal in the proposal. A wall that must not be climbable needs a separate change to the catch's reach.
- **[Order of changes.]** This builds on `LE.solids`. → Apply and archive it after `multiple-collision-rects`.
- **[Existing tests assume walking through rectangles.]** → None found. The actor and jump tests use floors of equal-height tiles, and the ledge tests only call `checkCollision` or start next to the ledge. The full suite is run to confirm.

## Migration Plan

None. Levels need no change: `level_1.json` has no rectangle that is a wall to a fighter standing on any of its floors. Rolling back is reverting the commit.
