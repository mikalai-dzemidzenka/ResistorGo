## Context

The fighter's state machine is in `Actor.Update` (`actor.go`). The run-jump and fall work (`add-run-jump-and-fall`, merged in code but not archived) already gives the patterns this change needs:

- **A priced mobility move.** `enterRunJumpState` refuses with `EvNoEnergy` when `canRunJump` fails, and spends through `spendEnergy(EnergyRunJump)`.
- **A launch.** `launchRunJump` sets `OnGround=false` and a negative `Speed.Y`. The take-off tick still applies movement.
- **A landing.** `land()` runs or stands depending on `abs(Speed.X) > RunReach`.
- **Appending to the enums.** States are appended before `stateCount`, animations are appended to `animTable`, and textures are listed in `resources.go`.

Some existing behaviour gets in the way of a roll and has to be worked around:

- `updateHorSpeed` applies steering, applies friction whenever no key is held (in the air too), and clamps to `MaxSpeed` (320).
- `updateFalling` puts any airborne fighter that is moving down into `StateFall`.
- A fighter blocks whenever `Updown` matches the attacker's (`actor.go` ~1778), whatever state it is in.
- `sampleInput` reads the mouse. `latchInput` latches edge presses per rendered frame.

## Goals / Non-Goals

**Goals:**
- The client prediction, the server and time rewind all compute the same charge from state that is already replicated or recorded.
- The drawn arc and the real roll run the same maths, so they cannot drift apart.

**Non-Goals:**
- Bots rolling.
- Aiming the roll's angle with the cursor.
- Catching a ledge during a roll.
- Sound.
- Fixing the energy spec's stale numbers. It says 200 while `MaxEnergy` is 300. That drift is older than this change.

## Decisions

**Three states: `StateDuck`, `StateRoll`, `StateRollRecover`.** They are appended after `StateClimbDown`. One "roll" state with sub-phases would hide its phase in the animation name, the way `climbingDown()` does. That works, but it is harder to read and to test than one state per phase. Each phase has its own animation and its own exits, so each gets its own state.

**`Input.Roll` is a held bit, not an edge.** The duck lasts while RMB is down, so the command carries `rl.IsMouseButtonDown(rl.MouseButtonRight)` sampled in `sampleInput`, like Left and Right. On the wire it is `bRoll`, the next free bit in the button byte. A held bit also fixes release timing without any latching: the first command that has the bit clear is the release. An edge "press" bit would cost another wire bit and a replicated "previous held" field. Its only use would be to stop a fighter ducking when it lands with the button still held, and the spec accepts that duck instead.

**The charge is `Ticks - AnimStartTick` in the duck.** `enterDuckState` resets the animation, so `AnimStartTick` marks the start of the hold. Both fields are already in `ActorRec`, and the rewind ring records them because it copies the whole `Actor`. So the client's reconcile and the server agree without a new field. The alternative was a `ChargeTicks` counter. It would need a snapshot field, and it would say nothing that isn't already known.

```
t      = min(Ticks-AnimStartTick, RollChargeTicks) / RollChargeTicks   // 0..1, RollChargeTicks = 60
launch = (lerp(RollMinX, RollMaxX, t) * Orientation, lerp(RollMinY, RollMaxY, t))
```

The starting values below are knobs to tune by eye, named beside the other movement constants:

| Constant | Value (px/s) | Compared with |
|---|---|---|
| `RollMinX` | 260 | a click is an escape: it lands about 62 px out and slides to about 129 px |
| `RollMaxX` | 300 | under `MaxSpeed` |
| `RollMinY` | -150 | |
| `RollMaxY` | -420 | about the run-jump's -437 |

One function, `rollLaunch(a) rl.Vector2`, returns the launch vector, and both the release and the arc call it.

**The roll is exempt from steering, friction and the clamp.** At the top of `updateHorSpeed`, `StateRoll` returns early, before steering, friction and the clamp. `updateVertSpeed` still applies gravity and its clamp. `blockSides` still stops the fighter at walls. `StateDuck` and `StateRollRecover` join the `canMove == false` list, so the duck stands still and the recovery slides out under friction. `StateDuck` joins the "face the cursor" condition, next to `StateIdle`, so the roll direction is simply `Orientation`.

**`updateFalling` skips `StateRoll`.** It skips the other states it already ignores the same way. The roll animation plays through the apex.

**Ducking is entered from the idle and run branches, and nowhere else.** `if a.Input.Roll && a.OnGround { a.enterDuckState() }` goes before the attack and jump checks, since a held RMB outranks them. `enterDuckState` refuses with `EvNoEnergy` when `Energy < EnergyRunJump`, and zeroes `Speed.X`. If the button is held with too little energy, the fighter raises `EvNoEnergy` every tick. That only keeps extending the same flash, which is acceptable.

**Launch on release, in the duck's case.** When `!a.Input.Roll`, the duck case calls `spendEnergy(EnergyRunJump)`, sets `Speed = rollLaunch(a)` and `OnGround = false`, and enters `StateRoll`. There is no `return`, so the take-off tick moves the fighter, as the run-jump comment in the run branch explains. Energy is spent at launch and not at the press, so a duck that a hit interrupts costs nothing. The case also leaves the duck if the fighter stops being on the ground, which happens for example when a shove carries it off an edge. `updateFalling` then takes over as normal.

**Landing:** the roll case does `if a.OnGround { enterRollRecoverState() }`, playing the recovery once. The recovery case calls `land()` when `AnimFinished`. `land()` can fire an attack pressed on its tick, which the spec allows once the recovery is over.

**A ducking, rolling or recovering fighter does not guard.** Guarding is decided in the swing's hit loop: `guarding := (a.Updown == actor.Updown || actor.Updown == 3) && !actor.rolling()`, where `rolling()` reports the three states. Without this, a ducking fighter whose stance happened to match would block.

**The arc lives in a new file, `roll.go`, and it runs the real simulation.** `rollPath(a)` copies the fighter by value, releases the copy (`Input.Roll = false`), and steps it with `Update` until it leaves `StateRoll`, collecting the feet position at every tick. A second loop written only for the arc (gravity, clamp and a check against solids) would be a copy of the physics that could drift from the real thing. Stepping a copy can't drift, and copying is safe because `Actor` is built to be recorded by value, as rewind already does. A roll touches no other fighter, and whatever the copy spends or raises stays on the copy. `drawRollArc(a)` dots every second point. `game.go` calls it for `w.Local` just before the `drawAttackOutline` loop. It draws nothing unless the fighter is ducking. The test `TestTheArcIsWhereTheRollGoes` replays a real roll and checks it against the path.

**A roll does not catch ledges.** `catchLedge` snaps an airborne fighter's X to the ledge, so `StateRoll` is left out of its `airborne` test, next to `StateClimbDown`.

## Risks / Trade-offs

- [The arc runs up to 300 fixed steps of one fighter each rendered frame while ducking] → That is cheap for a single fighter, and nothing runs when not ducking.
- [RMB also deletes in the level editor] → `matchHeld()` stops the fixed step while editing, so the fighter never sees it. `sampleInput` is not reached while editing.
- [Old clients cannot read the `bRoll` bit or the three new states] → Client and server already have to be the same build. `stateCount` keeps decoding safe for appended states.
- [Hold-to-charge makes a roll slow to start] → A tap still rolls at the minimum charge.
