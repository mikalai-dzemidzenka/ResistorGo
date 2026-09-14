## Why

A flinch does not move a fighter. *A landed hit costs structure and flinches* promises "a small knockback", and its scenario says the fighter "is knocked back slightly". It travels zero pixels.

`enterFlinchState` calls `hitFeedback(attacker, PunchSpeed)`, which sets `Speed.X = ±50`. Nothing clamps it — a flinched fighter stays in `StateIdle` or `StateRun`, and neither arm caps horizontal speed. The speed is simply too small to survive how the game moves things:

```
  Update, last lines:
      a.Pos.X += a.Speed.X * dt
      a.Pos.X = float32(int(a.Pos.X))     <- whole pixels, every tick
```

The fractional part is discarded sixty times a second rather than carried, so a fighter never accumulates its way across a pixel boundary. Below `1 px / FixedTimeStep` = **60 px/s** it does not move at all, however long the speed lasts. `PunchSpeed` is 50, which is 0.833 px/tick. Measured: 0 px, at every duration.

This was found while fixing the stun's throw-back (`stun-throws-back`). That was the same symptom — a knockback that never moved anyone — with a different cause: the stun's 150 was being clamped to 4 by a cap belonging to the attacker. Removing the clamp fixed the stun, because 150 is above the floor. It does nothing for the flinch, which was never clamped and was never fast enough.

The floor is worth stating plainly, because it is invisible at the call site: **every knockback, nudge and drift in the game under 60 px/s is dead code.** `PunchSpeed` is the one the spec notices.

## What Changes

The change is a choice between two shapes, and this proposal does not settle it — see Open Questions. Both satisfy the requirement; they differ in blast radius.

- **Either** `PunchSpeed` is raised above the floor, so a flinch moves a pixel or two per tick like the stun does. One constant, nothing else in the game changes, and the floor stays where it is for everything else.
- **Or** position accumulates sub-pixel and is rounded only for drawing, so a speed of 50 travels the distance it describes. Every sub-pixel motion in the game starts working, `PunchSpeed` keeps its meaning, and the pixel-perfect look is preserved at the point it actually matters — the frame.
- Either way the spec gains a scenario pinning that a flinch moves the fighter at all, so it cannot silently go back to zero.

## Open Questions

- **Which shape?** Raising `PunchSpeed` is a few characters and no risk, but it leaves the floor in place, so the next person to write a slow nudge hits the same wall with no warning. Accumulating sub-pixel position is the real fix and makes `PunchSpeed = 50` mean 50, but it changes the movement rule the project states as a core convention, and `Pos` is both replicated and recorded into the rewind ring — so it needs a look at prediction, at `recOf`/`apply`, and at whether the rewind ring should store the accumulator.
- **If `PunchSpeed` is raised, to what?** It must clear 60 px/s to move at all, and enough above it that friction does not drop it under the floor on the first or second tick. It also has to stay clearly under the stun's throw, which measures 20 px, or the flinch and the stun stop reading as different things.
- **Does anything depend on the flinch not moving?** Nothing found, but the flinch has never moved anyone, so no test or playtest has ever seen the alternative.

## Capabilities

### Modified Capabilities

- `structure`: *A landed hit costs structure and flinches* keeps its requirement — it was already right. It gains a scenario that the knockback actually changes the fighter's position, and one that a flinch stays clearly smaller than a stun's throw, so fixing the first cannot quietly erase the difference between them.

## Impact

- **Code**: `actor.go` — either the `PunchSpeed` constant, or the position integration at the end of `Update` plus wherever a sub-pixel remainder would have to live on `Actor`.
- **Tests**: that a flinch moves the fighter; that it moves it less than a stun does. If the integration changes, also that whole-pixel rendering is unaffected and that `recOf`/`apply` and the rewind ring round-trip whatever new state exists.
- **Possibly affected, and the reason the second shape is not free**: replication (`Pos` is on the wire), prediction and reconciliation, the rewind ring, and every other movement in the game.

## Non-goals

- **The stun's throw-back.** Fixed in `stun-throws-back`; this change must not change it.
- **Retuning knockback feel generally.** This is about a knockback that does nothing doing something.
