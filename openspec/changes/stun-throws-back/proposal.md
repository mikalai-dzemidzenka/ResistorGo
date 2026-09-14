## Why

A stun does not throw a fighter back, and the `structure` spec says it must. *Every third landed hit stuns* requires the defender to be "thrown back noticeably further than a flinch throws it", and its scenario repeats it. The fighter does not move at all.

`hitFeedback` sets the speed correctly — `StunKnockback` 150. Then the state switch takes it away:

```
  case StateAttack:
      a.Speed.X = clamp(a.Speed.X, -AttackDash, AttackDash)   // the attacker's lunge cap
      a.updateAttack()
  case StateAttacked:
      a.Speed.X = clamp(a.Speed.X, -AttackDash, AttackDash)   // <- copied onto the defender
      a.updateAttacked()
```

`AttackDash` is 4. On the first tick after the hit-stop freeze ends, the defender's 150 is cut to 4 — and 4 px/s is where it dies, because of how the game moves things:

```
  Update, last lines:
      a.Pos.X += a.Speed.X * dt
      a.Pos.X = float32(int(a.Pos.X))     <- whole pixels, every tick
```

The fractional part is discarded sixty times a second, so **any speed below 60 px/s moves a fighter zero pixels, forever**. Clamped to 4, a stunned fighter is at 0.067 px/tick and stands exactly where it was hit for the whole lockout. Measured, not derived: 150 uncapped carries it 20 px; capped at 4 it carries it 0.

So the game's one big punctuation mark does not read as one. A stun is meant to be the moment the exchange resets and both fighters get space; instead the defender is locked out standing still, and the cap doing it belongs to the fighter throwing the punch.

### The flinch is broken too, and not by this

`PunchSpeed` is 50, which is 0.833 px/tick — under the same truncation floor. A flinched fighter is in `StateIdle` or `StateRun`, neither of which clamps anything, and it still travels 0 px. *A landed hit costs structure and flinches* promises it is "knocked back slightly" and that is not happening either.

That is the same symptom with a different cause: the stun is capped, the flinch is merely too slow to survive rounding. Fixing it means changing `PunchSpeed` or the way position is integrated, neither of which is removing a clamp. It is filed as its own change so that each can be judged on its own, and it is why the scenarios below pin an absolute distance rather than comparing against a flinch that does not move.

## What Changes

- **The stun's throw-back reaches the fighter.** The `AttackDash` clamp is removed from the `StateAttacked` arm. `StunKnockback` then decays against the ordinary friction, and the fighter travels about 20 px over roughly 0.36 s before falling under the truncation floor and stopping.
- **`StunKnockback` keeps its value of 150.** It has never once taken effect, so it is an untested number rather than a tuned one; this change lets it be seen for the first time and leaves retuning to what that looks like in play.
- **The attacker's cap is untouched.** `case StateAttack` keeps its clamp: capping the lunge of the fighter that is swinging is what `AttackDash` is for.
- **Nothing else about a stun changes** — the lockout, its length, the hit animation, the flash, the structure cost and the hit count all stay exactly as they are.
- Not breaking. No wire change, no protocol version, no new state.

### What this exposes

- **The distance.** 20 px on a 160 px virtual screen is an eighth of the width. Play decides whether that is a stun or a nudge.
- **The overrun.** Friction takes the speed from 150 to the 60 px/s floor in about 0.36 s, while the stun animation runs 0.3 s — so the fighter is still moving, barely, as it regains control. Less of an overrun than the raw decay time suggests, because the floor stops it early.

Both are playtest tasks with a note to bring any change back into design.md, rather than pre-empted by a guess.

## Capabilities

### Modified Capabilities

- `structure`: *Every third landed hit stuns* keeps its requirement word for word — it was already right, and the code was wrong about it. It gains scenarios that pin the throw-back down so it cannot silently go missing again: that the fighter's position actually changes by a real distance, and that the throw goes the way the attacker was facing. The scenarios name a distance rather than comparing against a flinch, because the flinch's own knockback is currently 0 and would make the comparison pass on a one-pixel throw.

## Impact

- **Code**
  - `actor.go`: one line removed from the `StateAttacked` arm of the state switch in `Update`.
- **Tests**: that a stun carries the fighter a real, absolute distance; that the direction follows the attacker's facing; and that the attacker's own `AttackDash` cap still applies during `StateAttack`.
- **Not affected**: `hitFeedback`, `enterAttackedState`, `enterFlinchState` and every constant they use; the hit-stop freeze; the wire format; the position integration and its whole-pixel truncation.

## Non-goals

- **Tuning `StunKnockback`.** The number ships unchanged so that what it does can be judged before it is changed.
- **The flinch's knockback.** Same symptom, different cause, filed as its own change — see above.
- **The truncation floor itself.** Accumulating sub-pixel position and rounding only for drawing would fix both knockbacks and every other sub-pixel motion at once, but it changes the pixel-perfect movement rule the project is built on and needs its own look at prediction and rewind. Not smuggled in here.
- **Stopping the slide when the stun ends.** A real option, but a second behaviour decision; going in blind alongside the first would make neither judgeable.
- **Knockback into walls, off ledges or onto other fighters.** Existing collision handles the fighter wherever the throw takes it.
