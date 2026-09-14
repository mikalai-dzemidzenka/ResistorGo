## Context

See proposal.md — Why. The mechanism, in the order one tick runs it:

```
  enterAttackedState -> hitFeedback(attacker, StunKnockback)
      a.Speed.X     = orientation * 150
      a.FreezeTicks = HitStopTicks (6)

  ticks 1..6   Update returns at the freeze guard: the speed is set but
               nothing integrates it, so the fighter has not moved yet

  tick 7       checkCollision
               updateHorSpeed   friction: 150 -> ~145.8  (canMove is false
                                for StateAttacked, so no steering)
               updateVertSpeed / updateFalling
               switch a.State
                 case StateAttacked:
                     Speed.X = clamp(Speed.X, -AttackDash, AttackDash)  <- 145.8 -> 4
               a.Pos.X += a.Speed.X * dt          4/60 = 0.067 px
               a.Pos.X = float32(int(a.Pos.X))    discarded
```

Constants: `AttackDash` 4, `PunchSpeed` 50, `StunKnockback` 150, `DefaultFriction` 250, `MaxSpeed` 320.

Three facts decide the fix, all measured against the tree rather than derived:

- **There is a floor on movement, not just on rendering.** `Update` truncates `Pos` to whole pixels every tick, so the fractional part never accumulates. Below `1 px / FixedTimeStep` = **60 px/s** a fighter does not move at all, however long you wait. This is the fact that makes the clamp fatal rather than merely restrictive: 4 px/s is not "slow", it is stationary.
- **The clamp is the attacker's, not the defender's.** `AttackDash` is the lunge a swing gives the fighter making it — a position nudge in `enterAttackState` and the cap on `case StateAttack`. Nothing about it describes being hit.
- **The flinch is broken independently.** `PunchSpeed` 50 is 0.833 px/tick, under the floor, in states that clamp nothing. Removing this clamp does not touch it. Filed separately; it is why the tests here measure an absolute distance instead of comparing the two.

Measured travel, freeze excluded, left to decay:

```
  speed 50   (flinch)  0.833 px/tick   ->   0 px    under the floor
  speed 150  (stun)    2.500 px/tick   ->  20 px    stops when friction
                                                    brings it under 60 px/s,
                                                    after ~0.36 s
```

## Goals / Non-Goals

**Goals:**

- Make the throw-back that `hitFeedback` already computes actually reach the fighter.
- Change one thing, so that what `StunKnockback = 150` feels like can be judged on its own.
- Leave tests behind that fail if the throw is capped again, and that cannot pass on a fighter which barely twitches.

**Non-Goals:**

- See proposal.md — Non-goals. At design level: no new state, no new constant, no new branch — this removes a line rather than adding a mechanism, and deliberately does not touch the truncation floor.

## Decisions

### Remove the clamp rather than special-case it

The `StateAttacked` arm keeps `a.updateAttacked()` and loses the clamp above it. Friction then decelerates the knockback the way it decelerates every other speed, and `canMove` in `updateHorSpeed` (`actor.go:670`) already excludes `StateAttacked`, so the player cannot steer out of the throw. The lockout stays what it was — it governs what the fighter may do, not how far the hit carries it.

*Why not clamp to a larger number* — a cap sized to let 150 through is not a cap, it is a second knockback constant hiding as one. If the distance is wrong, `StunKnockback` is the honest place to change it.

*Why not fix the truncation floor here* — it would fix the flinch too, and every other sub-pixel motion in the game, which is exactly why it is not a side effect of a stun fix. Sub-pixel accumulation changes the pixel-perfect movement rule the project states as a core convention, and `Pos` is replicated and rewound, so it needs its own look at prediction and the rewind ring.

*Why not zero the speed when the stun ends* — a defensible second behaviour, but shipping it alongside the first would mean neither could be judged. The floor already cuts the slide short at about 0.36 s, close enough to the 0.3 s animation that the overrun may not need anything.

*Why not keep a clamp for the airborne case* — a fighter stunned in the air is carried by the same speed and falls under the same gravity; `updateFalling` and `checkCollision` already handle where it lands.

### Testing measures position, and an absolute distance

Two traps, both hit while writing these:

- **Asserting on `Speed.X` passes on a fighter that never moves.** That is the bug itself: the speed was always set. So the tests read `Pos.X`.
- **Comparing the stun against a flinch is degenerate** while the flinch travels 0 px: the comparison would pass on a one-pixel throw. So the tests name a distance — comfortably below the measured 20 px, comfortably above the couple of pixels a broken-but-not-clamped throw would manage.

The defenders in the tests are bots. The `StateIdle` arm turns an `IsPlayer` fighter to face the mouse, and a turn divides `Speed.X` by three (`actor.go:697`), which would quietly eat part of what is being measured.

A test that the attacker's own `AttackDash` cap still applies during `StateAttack` guards the other half — that this change removed the right one of the two identical lines.

## Risks / Trade-offs

- **20 px may be too little or too much.** → One constant, and the change is written so that turning it is the follow-up. Two playtest tasks decide it.
- **The measured distance depends on the truncation floor**, so changing how position is integrated later will change this throw without anyone touching `StunKnockback`. → Recorded here, and the absolute-distance test will catch the day it moves.
- **A stunned fighter can now be carried into a wall, off a ledge, or through where another fighter stands.** → All three already apply to any moving fighter; collision, `updateFalling` and the absence of actor-to-actor collision treat this no differently.
- **Networked play sees a larger correction if client and server ever disagree about a stun.** → They do not disagree: hits are resolved on the authority only, and `Speed` and `Pos` are both replicated. Worth an eyes-on check under latency simulation anyway.

## Migration Plan

None. One line, no data, no protocol, no saved format. Revert is putting the line back.
