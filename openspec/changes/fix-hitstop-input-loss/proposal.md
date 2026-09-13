## Why

A click made during hit-stop is destroyed. The attack button is latched for the next fixed step (`pressedAttack`), `sampleInput` clears the latch as it builds the tick's command, and `Actor.Update` returns at its first line while `FreezeTicks > 0` — before any state reads `Input`. So on each of the six `HitStopTicks` the latch is set, consumed and thrown away.

The `structure` spec forbids exactly this. *A landed hit costs structure and flinches* says a flinch "SHALL NOT lock the defender out of acting — it can move, attack, block and jump during one", and its scenario says the fighter "can act on the next tick". The freeze sits in front of the flinch and eats the first 100 ms of it, so the defender who counters fastest gets nothing while a slower one is served. The requirement names the freeze and the right to act in the same breath without saying how they compose; the code resolves that by dropping the input.

The same tick also bypasses a buffer the game already has. `AttackQueued` exists to hold a click made mid-swing and fire the chain when the animation finishes, and it is set in `updateAttack`, downstream of the freeze return. `hitFeedback` freezes *both* fighters, so a swing that lands opens a dead window at its own start: clicks in it never reach the buffer built to catch them, and the same clicks on a whiff do.

This is the third appearance of one shape — input latched for a step that will not act on it. The two paused modes were fixed by `fix-paused-input-latch`; this is the remaining one, and it is not the same fix.

## What Changes

- **The latch is consumed only by a tick that will act on it.** While a fighter is frozen, `sampleInput` still reports the input and still produces a command, but leaves the latch standing. The first unfrozen tick consumes it.
- **A click during hit-stop acts when the freeze ends** instead of vanishing: the defender's counter comes out, and the attacker's mashed chain reaches `AttackQueued`.
- **Hit-stop is not folded into `matchHeld()`.** The two are different questions with opposite answers, and the spec delta says which is which.

### Why hit-stop defers where a paused mode discards

`fix-paused-input-latch` fixed its two cases by *dropping* the latch. Hit-stop needs the opposite, because the click means something different:

| Held by | The click meant | Right response |
| --- | --- | --- |
| Edit mode | place a tile | discard |
| Time manipulation | scrub | discard |
| Hit-stop | attack | defer |

The predicates differ in kind as well as in answer. `matchHeld()` is global and asks whether the *match* is held; `FreezeTicks` is per-actor and asks whether *one fighter* is. In a two-fighter match one side can be frozen while the other is not, so gating the fixed-step loop on a freeze would stop the world for both. The freeze test therefore stays beside the actor, and `matchHeld()` is left alone.

### Assumptions

- Nothing on a frozen tick reads `Input`: `Update` returns before the state switch, and `advanceAnim` is deferred after that return. Leaving the latch standing changes no other behaviour on that tick.
- The client keeps producing one command per fixed tick, as *Input commands* requires. The fix withholds the latch, it does not skip sampling — so `Seq` stays unbroken and the redundant-command window is unaffected.
- `Attack` therefore reads true on each frozen tick and on the first live one. That is safe: a frozen tick ignores it, `enterAttackState` can only fire once because the state changes, and `AttackQueued` is idempotent. Client and server run the same commands through the same freeze, and `FreezeTicks` is already replicated, so prediction and the authority agree without a protocol change.
- A stun still locks the defender out. The latch survives the six freeze ticks and is then consumed and dropped by `StateAttacked`, which reads no input — as *Every third landed hit stuns* requires.
- The latch's life is bounded: at most the freeze, then the first live tick either acts on it or drops it per its own state.
- Entering a paused mode during a freeze is unaffected. The preserved latch is cleared by `clearLatchedInput()` on leaving, so discard still wins where discard is right.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `structure`: *A landed hit costs structure and flinches* gains the rule that input given during the shared hit-stop freeze is held rather than lost, and a scenario for it. The requirement already grants the defender the right to act during a flinch; it does not currently say how that composes with the freeze it names in the same sentence, which is the gap the code fell into.

## Impact

- **Code**
  - `actor.go`: `sampleInput` leaves `pressedAttack` and `pressedJump` standing while the sampled actor is frozen.
- **Tests**: a regression test that an attack latched while `FreezeTicks > 0` survives the frozen ticks and fires on the first live one; and one that a stun still swallows it.
- **Not affected**: `matchHeld()` and both paused modes, `HitStopTicks` and every other gameplay value, the packet format and the snapshot flags, the rewind buffer, the camera effects.

## Non-goals

- **The wider input model.** Every state that does not read `Input` consumes and drops the latch: `StateRun` ignores `Jump`, `StateJumpStart` and `StateJump` buffer no attack for the landing, `StateIdle` drops an attack while guarding with block-all. Turning the latch into a decaying buffer would fix all of those at once and change the feel of the game everywhere; it is a design change with its own tuning pass, not this bug fix.
- **The missing running jump.** `enterJumpStartState` has one caller, in `case StateIdle`, so W does nothing above `RunReach`. Whether that is intended is a gameplay question and is not settled here.
- **`HitStopTicks` itself.** How long the freeze lasts is a feel decision; this change is only about what happens to input during it.
- Hit-stop's visual hold, the knockback, the flinch animation, and the chain rules beyond the click reaching `AttackQueued`.
