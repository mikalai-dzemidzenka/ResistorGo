## Context

The requirements are in `specs/climb-down/spec.md` and the two delta specs. This section covers only the parts of the current code that decide how the change is built.

- **Collision is one flat list.** `LE.solids` is every collision rectangle in world space. `buildIndex` builds it, and `checkCollision`, `headBlocked`, `blockSides`, bot pathing, blood and the Lua API all walk it. Order matters: `checkCollision` depends on the order it visits the rectangles in. `rl.Rectangle` has no room for a flag.
- **`checkCollision` does three things per rectangle.** The *ground* branch holds the fighter up if its feet (`Pos.Y+48`) are anywhere from one pixel above the top down to the bottom, and snaps it to `top-48`. That is also how a fall lands. The *head* branch runs when the head (`Pos.Y+6`) is inside the rectangle: it zeroes vertical speed and snaps the fighter under the rectangle. `headBlocked` pre-scans for this and suppresses ledge catches. The *ledge* branches run when the head is level with the rectangle, it is within 32 px ahead, and the fighter is airborne.
- **The snapshot record has no free flag bits.** All 16 bits of the `ActorRec` flags are used, the last by `fBlockOdd`. Position (`X`, `Y`) and speed (`SX`, `SY`) are replicated. The command's button byte has three bits free after `bFaceLeft`.
- **An `Actor` is copied by value for time manipulation**, so it holds no slices or maps. Reconciliation applies a server record and replays commands, so any state that is not replicated has to be something the replay can work out again from the record.
- **`climb-down.png` is 240×96**: five 48-px-wide frames, twice the usual height. Measured figure bounds: frame 0 spans y 13–47, which is standing with feet on row 47, the same as the idle sheet. Frame 4 spans y 44–68, hanging by the hands at floor level. `fall.png` frame 0 spans y 4–47. `DrawAnimOnce` and `spriteDraw` already draw a play-once sheet at its full texture height from `Pos`, so the sheet draws correctly with `animYOffset: 0`. The outline and rim passes follow it too.
- **Existing jumping/fall behaviour** (from the unarchived `add-run-jump-and-fall` change, already in the code): `enterFallState`, `updateFalling` and `land` are the fall this change hands over to.

## Goals / Non-Goals

**Goals:**
- Tell drop-through rectangles apart from solid ones without changing the type of `LE.solids` or any of the code that reads it.
- Work out "is this fighter past the platform" from position and speed alone, so nothing new has to be replicated or recorded.
- Leave every existing level, its bytes and its fingerprint exactly as they are.

**Non-Goals:**
- Anything stopping a fighter below a drop-through platform. It started as solid from below and was changed in play: a jump has to be able to go up through it.
- Bots climbing down, or bot pathing planning routes through drop-through platforms.
- Buffering an S press across hit-stop the way attack and jump are buffered (`BufAttack`/`BufJump`). The snapshot has no bit left for a third buffer, and a climb-down asked for mid-fight is not worth growing the record for.
- A sound for the climb-down.

## Decisions

### A third rectangle kind, not a flag on collision rectangles
`ObjectRects` gets `Drop []rl.Rectangle`, `LevelObject` gets `DropRects`, and the wire structs get `dropRect`/`dropRects` keys using the existing `rectsToWire`/`rectsFromWire` helpers. The rect tool gets `rectDrop` as a third kind, appended after `rectLight` so the constants and the `ToggleGroup` index agree: `"Collision (C);Light (L);Drop (P)"`. It is picked with **P** ("platform"), which the editor does not use. D would be the obvious letter, but it is a movement key.
- *Why not a flag per collision rectangle:* `rl.Rectangle` cannot carry one, so a flag means a parallel `[]bool` inside `ObjectRects`. Every clone, compare, paste, follow and wire path would have to keep the two lists in step, and the `collisionRect` bytes of every existing level would change. A separate kind drops into all the machinery that already carries "every rectangle of every kind" (`clone`, `sameRects`, `list(which)`, the hit-test loop over kinds, copy/paste, follow) and saves nothing extra for objects that have none.
- *Why not per texture:* the user chose per rectangle, so one object can have a solid part and a drop-through part.

### `LE.solids` stays as it is; a parallel `LE.drop []bool` marks the drop-through ones
`buildIndex` appends each object's drop rectangles to `solids` right after its collision rectangles and records `true` at the same index in `LE.drop`. Bots, blood and the Lua API keep reading `LE.solids` unchanged, and for them a drop-through platform is ordinary floor. Only the actor's collision code reads `LE.drop`.
- *Alternative:* change `solids` to a struct slice. That touches every reader, including tests that assign `LE.solids` directly, and gains nothing.

### A drop-through rectangle is passed by a fighter that was already below its top
Let `feet = Pos.Y + CollisionBox.Y + CollisionBox.Height` and `was = feet - Speed.Y*dt`, which is where the feet were before this tick's vertical move. For a rectangle marked drop, `checkCollision` changes as follows:
- **Ground branch:** applies only when `was <= top + 1`. The fighter had to be at or above the top to be held by it. A standing fighter (feet exactly at the top, `Speed.Y == 0`) and one landing from above (it started the tick above the top) are both held. A fighter that climbed down (placed below the top with `Speed.Y == 0`) is not.
- **Head branch and `headBlocked`:** never apply. The platform is open from below, so a rising head passes up through it, and it is never a ceiling that blocks a ledge catch elsewhere.
- **Ledge branches:** apply only when `Speed.Y < 0`, meaning the fighter is rising, so a jump at the platform's edge can still be pulled up onto it. They live in `catchLedge`, shared with ordinary solids. A rising fighter whose feet are inside the platform is not stood on it: that waits until it comes down onto the top.

Everything here comes from `Pos` and `Speed`, which the snapshot already carries. So a client replaying after a correction reaches the same answer as the server, and the rewind buffer needs nothing new. Ordinary solids are untouched.
- *Alternative:* remember which platform the fighter dropped through (a `Dropping` flag or rectangle on the actor). That needs a replicated bit, and the record has none free, or a larger record. It would also need clearing rules. The position-and-speed rule needs neither.
- *Trade-off accepted:* a fighter falling past the edge of a drop-through platform does not catch that ledge. The ledge test needs the head level with the rectangle, which means the feet are already far below its top. The spec says this in so many words.

### Footing check: every rectangle holding the fighter up must be drop-through
The helper `a.onDropThrough()` repeats the ground branch's test (the centre column inside the rectangle, the feet at its top) over `LE.solids`. It reports true only if at least one rectangle holds the fighter and every rectangle that does is marked drop. If a solid and a drop-through rectangle share a top under the fighter, the fighter is on solid floor, and climbing down would only land it again in the same place.

### `StateClimbDown`, appended after `StateClash`
The state is entered from the `StateIdle` and `StateRun` arms when `a.Input.Down && a.OnGround && a.onDropThrough()`. It is checked after attack and jump, so if more than one is pressed on the same tick, attack wins, then jump, then climb-down. `enterClimbDownState` zeroes `Speed.X`, stops any flicker the way other voluntary actions do (`stopFlicker`), and sets `climb-down.png` with `once: true, reset: true, totalFrames: 5, animDuration: ClimbDownSeconds`. The duration starts at 0.5 s and is tuned by eye. In the state arm the fighter holds still (`Speed = {}`), and on `AnimFinished` it calls `finishClimbDown`. `updateHorSpeed`'s `canMove` excludes it while `climbingDown()`, so held keys neither steer nor turn it until the drop begins. `updateFalling` skips the state, so the drop stays in it rather than becoming an ordinary fall.

Stuns, shoves and deaths already take a fighter out of any state (`enterAttackedState`, `enterShovedState`, `enterDeathState`). The fighter has not moved, so it is left on top of the platform with no extra code. `enterFlinchState` does not change state, so a flinch leaves the climb-down running, which is what the spec asks for.

### Handover: move down `ClimbDownDrop = 40` px, capped at the next floor, then drop in the climb-down state
The figure's top in the last frame (row 44) lines up with fall frame 0's top (row 4) when the fighter moves 40 px down. That keeps the hands where they are rather than the feet, so the pose reads as letting go. `finishClimbDown` finds the highest rectangle top strictly below the current feet, under the centre column, within the next 40 px, and if there is one it places the feet on it instead. It then sets `Speed.Y = 0` and switches the animation to `fall.png`, but leaves the fighter in `StateClimbDown` until it lands. `climbingDown()` (state and climb-down animation together) tells the in-place part from the drop. The drop can steer, `updateFalling` leaves it alone, and `checkCollision` counts no fighter in the state as airborne for ledge catches. The arm lands it through `land()` once it is on the ground. The 40 is a starting value, tuned by eye with the other animation constants.

*Why not the ordinary fall:* found in play. A drop-through tile usually sits in a row of solid tiles. Tiles are 39 px wide, so a solid neighbour's edge is always within the 32 px a fall reaches for a ledge, and at the height the fighter just left. An ordinary fall caught it at once and climbed straight back up. Remembering which height to ignore would need a replicated field, and the record has no room for one. A state of its own is already replicated, so "no ledges until it lands" costs nothing on the wire. The price is that a drop never catches a lower ledge on the way down either.

### Input: `Input.Down`, latched from S, carried as `bDown`
`latchInput` latches `rl.IsKeyPressed(rl.KeyS)` into `pressedDown` beside `pressedJump`. `sampleInput` hands it on and clears it, and `clearLatchedInput` drops it. The command gets `bDown` (bit 5 of the button byte). Bots leave it false. The snapshot does not change except for the new state value and the new animation ID. `stateCount` already covers the decoder's range check.

### Registration: append `climb-down.png` to the end of `animTable`
It goes after `battle-low-1-right.png` and gets the next ID. `TestAnimTableIsAppendOnly` gets an entry pinning it, as the run-jump and fall did.

## Risks / Trade-offs

- [A drop-through platform thicker than 40 px leaves the fighter inside it after the handover] → The ground rule ignores the platform because the fighter started below its top, so it falls out of the bottom. It looks odd while it does, but it is harmless. Platforms are meant to be thin, and the editor makes the thickness visible.
- [A fighter landing very fast on a thin drop-through platform could start a tick above the top and end below the bottom, so the ground branch never sees its feet inside the rectangle] → This already applies to ordinary solids. `MaxFallSpeed` is 900 px/s, or 15 px a tick, so platforms at least 15 px thick are safe. A new rectangle starts as one grid cell (39×83), which is well over that. Anyone who shapes a platform thinner than 15 px takes the same risk they would with an ordinary solid.
- [`Pos.Y` is truncated to a whole pixel after each move, so `was` can be off by up to a pixel] → The `+1` in `was <= top + 1` matches the ground branch's own one-pixel allowance.
- [Bots misjudge falls near drop-through platforms] → They walk and land on them like ordinary floor, which is correct. They never climb down, so they never get below one by that route. A bot knocked off an edge falls as it does today.
- [The server and clients disagree about which rectangles are drop-through] → Drop rectangles are in the level JSON, so a mismatch changes the level fingerprint and the existing session check refuses the join.

## Migration Plan

Nothing to migrate. Existing levels and defaults files have no `dropRect` keys and load with no drop-through rectangles. Their bytes and fingerprints do not change. An older client joining a newer server fails the fingerprint check only if the level actually uses drop-through rectangles. Otherwise it plays, but it would not understand `bDown`, the new state or the new animation ID. The project already ships client and server together, as `resistor.exe` built by `build.ps1`, so there is no mixed-version play to support.

## Open Questions

- `ClimbDownSeconds` (0.5 s) and `ClimbDownDrop` (40 px) are starting values. Tune them by eye in play. The spec does not depend on either number.
- The overlay colour for drop-through rectangles in the editor and debug view. Pick one that reads against both the collision (sky-blue) and light outlines.
