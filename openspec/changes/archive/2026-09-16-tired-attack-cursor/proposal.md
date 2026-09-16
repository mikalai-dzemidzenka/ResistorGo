## Why

Energy only matters at the moment a swing is refused: a fighter with 13 points left handles exactly like one with a full pool, and nothing on the crosshair says how close the refusal is. Spending the pool should be felt before it runs out, and it should be visible where the player is already looking — at the green attack crosshair that chooses the guard and attack height.

## What Changes

- **The green crosshair pales as energy drains.** Its colour follows the local fighter's energy: the green it is today at a full pool, fading continuously towards a pale grey-green as the pool empties. It stays opaque, so the guard height is always readable.
- **The green crosshair follows the red one more slowly when energy is not full.** The follow speed falls continuously with energy, from today's feel at a full pool to a noticeably sluggish one at empty. Because the green crosshair is what picks the height, a tired fighter is slower to change guard and attack height alike — blocking stays available at any energy, but reaching the right height takes longer.
- **The follow becomes frame-rate independent.** Today the green crosshair closes half the gap every rendered frame whatever the frame time; it becomes a time-based exponential follow tuned to feel the same as now at 60 fps and full energy, so a slower machine is not also a more tired fighter.
- With no local fighter — dead, not yet spawned — the crosshair behaves as at full energy.

Not in scope: slowing the bots' guard changes (they set the height directly, not through the crosshair; a follow-up can add an energy-dependent delay if tired players find the bots unfair); changing the red crosshair; changing energy costs, capacity or regeneration; any network or protocol change.

## Capabilities

### New Capabilities

- None.

### Modified Capabilities

- `energy`: a new requirement that a player's attack crosshair shows and is slowed by their energy, making explicit that a tired fighter can still block at any height but takes longer to get there. Added rather than a modification of `Energy pool`, which the in-flight `add-run-jump-and-fall` change is already rewriting.

## Impact

- `cursor.go`: `Cursor.Update` takes the local fighter's energy fraction and follows with `1 - exp(-k·dt)`, `k` interpolated between a tired and a full rate; `DrawGUI` blends the crosshair colour by the same fraction. `CursorAttackSpeed` gives way to the two rates.
- `game.go`: the `Cur.Update(dt)` call passes the local fighter (or its energy), and the cursor's initialisation drops `AtkSpeed`.
- Gameplay: guard and attack height reach the server later for a tired player, since `sampleInput` reads the slowed crosshair. Only the resulting `Updown` is sent, so the server, the protocol and the replays are untouched.
- Bots (`botInput`) are unaffected.
