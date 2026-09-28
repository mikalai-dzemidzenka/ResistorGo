## Why

A click made during hit-stop is destroyed. The attack button is latched for the next fixed step (`pressedAttack`, `actor.go:810`), `sampleInput` clears the latch as it builds the tick's command (`actor.go:844`), and `Actor.Update` returns while `FreezeTicks > 0` (`actor.go:230`) — before any state reads `Input`. So on each of the six `HitStopTicks` the latch is set, consumed and thrown away.

The `structure` spec forbids exactly this. *A landed hit costs structure and flinches* says a flinch "SHALL NOT lock the defender out of acting — it can move, attack, block and jump during one", and its scenario says the fighter "can act on the next tick". The freeze sits in front of the flinch and eats the first 100 ms of it, so the defender who counters fastest gets nothing while a slower one is served. The requirement names the freeze and the right to act in the same breath without saying how they compose; the code resolves that by dropping the input.

The same tick also bypasses a buffer the game already has. `AttackQueued` exists to hold a click made mid-swing and fire the chain when the animation finishes, and it is set in `updateAttack` (`actor.go:1165`), downstream of the freeze return. `hitFeedback` freezes *both* fighters (`actor.go:1023`), so a swing that lands opens a dead window at its own start: clicks in it never reach the buffer built to catch them, and the same clicks on a whiff do.

Jump has the identical hole and no buffer at all. `pressedJump` runs the same path (`actor.go:814`, `:842`) and dies at the same line, and there is no `JumpQueued` to fall back on. A jump pressed during hit-stop is gone with no recourse, in every state that would have taken it.

## What Changes

- **The freeze holds the input instead of eating it.** A frozen tick moves `Input.Attack` and `Input.Jump` into two new per-actor bools (`BufAttack`, `BufJump`) before its early return. The first unfrozen tick ORs them back into `a.Input` and clears them, then runs the state switch as it always has.
- **A press during hit-stop acts when the freeze ends** instead of vanishing: the defender's counter comes out, the attacker's mashed chain reaches `AttackQueued`, and a buffered jump is taken by whatever state can take it.
- **The buffer's lifetime is derived, not tuned.** It is exactly the freeze — at most `HitStopTicks`, 6 ticks, 100 ms. There is no new window constant and nothing to balance.
- **The six `Input` read sites are untouched.** Folding the buffer back into `a.Input` before the switch means `StateIdle`, `StateRun`, `StateAttack`, `land()` and the two jump reads keep their own conditions and need no edit.
- **Death drops the buffer.** `enterDeathState` already clears `AttackQueued` (`actor.go:1074`); it clears `BufAttack` and `BufJump` beside it, so a press buffered on the killing blow cannot survive into the respawned fighter.
- **BREAKING** (wire): the buffer is replicated in two of the three free bits of the existing `Flags uint16`, so `ProtocolVersion` moves 6 → 7. The record does not grow — `actorRecSize` stays 44 and no packet gets larger — but the bits change meaning, so client and server must be built together, which the existing version handshake already enforces.

### Routing, not discard

The earlier reading of this bug called it the third appearance of one shape, alongside the two cases `fix-paused-input-latch` fixed, and sorted the three by what the click meant. Sorted by mechanism they are two shapes, not one:

| Held by | A fixed step runs? | Was the input consumed? | Shape |
| --- | --- | --- | --- |
| Edit mode | no | yes — by the editor (LMB places a tile) | routing |
| Time manipulation | no | yes — by the scrubber (the drag) | routing |
| Hit-stop | **yes** | no — nothing took it, it was dropped | buffering |

Nothing is discarded in the first two rows. The click *was* consumed, by a second consumer that owns the device while that mode holds the match; `clearLatchedInput()` tidies a latch that has already been serviced, which is why it is right there and wrong here. Hit-stop is the first appearance of the other shape: a fixed step ran, read the input, and threw it away.

That gives two rules instead of one table:

1. **Routing** — while another consumer owns the device, the simulation latch is tidied up. Already correct, and untouched by this change: `matchHeld()` and `clearLatchedInput()` stay exactly as they are.
2. **Buffering** — input that reaches the simulation is held *in* the simulation until a tick can act on it. `AttackQueued` is already this; the freeze buffer is the same idea one layer earlier, before the state gate rather than inside it.

### Why the buffer lives in the simulation, not in `sampleInput`

Withholding the latch in `sampleInput` puts the buffer in the input layer, and on a client `sampleInput` is not input-layer only — it mints a networked command (`net_client.go:310`). A standing latch would repeat `Attack: true` on each of the seven ticks it spans, and the server's freeze need not be tick-aligned with the client's, because lag compensation resolved the hit at its own tick. Whichever tick the server unfreezes on it *probably* has an `Attack` waiting; if its window opens before the click it does not, and the client has already predicted a swing. That is a correction on the most feel-sensitive input in the game.

Keeping the buffer in the simulation removes the question. The command stays a one-tick edge, so *Input commands* keeps meaning what it says; client and server run the same fold-back on the same replicated `FreezeTicks`; and the buffer is in the snapshot, so `reconcile`'s rollback (`net_client.go:283`) restores it before replaying.

### Assumptions

- A frozen tick reads nothing else: `Update` returns before the state switch, and `advanceAnim` is deferred after that return. Capturing the two bools there changes no other behaviour on that tick.
- A fighter's energy cannot move during its own freeze — `regenEnergy` is at `actor.go:243`, after the return, and nothing spends. So a buffered press cannot become unaffordable while it waits, and `enterAttackState`'s `canSwing` check (`actor.go:873`) is reached with the same pool the player pressed against.
- A stun still locks the defender out, with no extra code. The fold-back happens on the first unfrozen tick; on that tick a stunned fighter is in `StateAttacked`, whose arm calls `updateAttacked()` and reads no input, so the press is folded in and dropped by that tick. A flinch leaves the fighter in `StateIdle` or `StateRun`, where `actor.go:288` and `:310` do read it. The requirement's split between flinch and stun falls out of the state machine rather than being spelled out in the fix.
- Two free bits exist. `Flags uint16` uses bits 0–7 for the named flags, 8–9 for `HitCount` and 10–12 for `HitRun`; 13, 14 and 15 are unused.
- The latch itself, `sampleInput`, `matchHeld()` and both paused modes are unchanged. Entering a paused mode during a freeze still ends in `clearLatchedInput()`, so routing still wins where routing is right.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `structure`: *A landed hit costs structure and flinches* gains the rule that input given during the shared hit-stop freeze is held rather than lost, and scenarios for the defender's counter, the attacker's chain, jump, one-press-one-action, the stun boundary and death. The requirement already grants the defender the right to act during a flinch; it does not currently say how that composes with the freeze it names in the same sentence, which is the gap the code fell into.

`multiplayer/replication` is deliberately **not** listed. *Snapshots* already requires "the full state needed to predict or display" each fighter "and combat flags", and the buffer is exactly that; two more bits inside an existing field change no requirement, and *Still fits the packet* holds unchanged because the record does not grow. *Input commands* keeps calling attack and jump edge-triggered, which stays true under this design.

## Impact

- **Code**
  - `actor.go`: two `bool` fields on `Actor`; capture and fold-back around the freeze guard in `Update`; clear in `enterDeathState` beside `AttackQueued`.
  - `net_protocol.go`: `fBufAttack` and `fBufJump` at bits 13 and 14; set in `recOf`, read in `apply`; `ProtocolVersion` 6 → 7. `actorRecSize` unchanged at 44.
- **Tests**: regressions that a press latched while `FreezeTicks > 0` survives the freeze and fires on the first live tick, for attack and for jump; that one press is one action; that a stun swallows it; that death drops it; and a protocol round-trip carrying the two new flags.
- **Not affected**: `sampleInput`, `latchInput`, `matchHeld()`, `clearLatchedInput()` and both paused modes; `AttackQueued` and its window; `HitStopTicks` and every other gameplay value; packet size; the rewind buffer; the camera effects; the six `Input` read sites in the state switch.

## Non-goals

- **A general input buffer with a tunable window.** Turning every `Input` read into a decaying buffer would also fix the cases outside a freeze — an attack pressed during `StateJumpStart` and wanted on landing, an attack pressed while guarding with block-all — and would subsume `AttackQueued`. It cannot be done without picking a window, and any window short enough to feel tight is shorter than the 18 ticks `AttackQueued` effectively allows today, so it changes how chaining feels. That is a design pass with its own playtest, not this bug fix.
- **`AttackQueued`'s own semantics.** It stays a bool with no expiry inside the swing. Chaining feels exactly as it does now.
- **`HitStopTicks` itself.** Whether the defender's freeze should simply be shorter is a live design question and a different answer to the same complaint; this change is only about what happens to input during the freeze, and does not foreclose tuning the length later.
- Hit-stop's visual hold, the knockback, the flinch flicker, and the chain rules beyond the press reaching `AttackQueued`.
