## 1. The latch survives a tick that cannot use it

- [ ] 1.1 Add `func (a *Actor) frozen() bool` beside `Update` in `actor.go`, returning whether `FreezeTicks > 0`, and use it at `Update`'s existing early-return guard. Say in its comment that it is the per-actor, deferring counterpart to the global, discarding `matchHeld()`, so the two are not confused for one rule. Verify `go build ./...` succeeds and `go test ./` still passes with no behaviour change.
- [ ] 1.2 In `sampleInput`, clear `pressedAttack` and `pressedJump` only when the sampled actor is not frozen, leaving the returned `Input` and the emitted command untouched. Verify `go build ./...` succeeds, and confirm by reading that neither caller — `stepOffline` in `game.go` nor `Client.tick` in `net_client.go` — gained a condition of its own and that the client still produces one command per tick.

## 2. The bug stays fixed

- [ ] 2.1 Add a test that an attack latched while an actor is frozen survives every frozen tick and is consumed by the first live one: set `FreezeTicks`, set the latch, step `Update` through the freeze asserting the latch still stands, then assert the fighter attacks on the tick after. Verify it fails against the current code and passes after task 1.2.
- [ ] 2.2 Add a test for the attacker's chain: with the fighter in `StateAttack` and frozen, a latched attack reaches `AttackQueued` once the freeze ends, exactly as a press later in the same swing would. Verify it passes.
- [ ] 2.3 Add a test that one press is one action — a single latched attack spanning several frozen ticks produces exactly one attack, not one per tick. Verify it passes.
- [ ] 2.4 Add a test for the stun boundary: a latched attack survives the freeze and is then consumed and dropped by `StateAttacked` without firing, and does not fire when the stun ends either. Verify it passes, so the fix does not leak input through a lockout.

## 3. Nothing else moved

- [ ] 3.1 Verify the paused-mode tests still pass unchanged — `TestMatchHeld`, `TestLeavingTimeManipulationDropsLatchedInput` and `TestLeavingEditModeDropsLatchedInput` — confirming `matchHeld()` and both discard paths were not touched.
- [ ] 3.2 Verify the packet format and snapshot flags are unchanged: `go test ./` includes the protocol round-trip tests, and `git diff` shows no edit to `net_protocol.go`.
- [ ] 3.3 Run `go test ./` and `go test -tags release ./` and confirm both pass.

## 4. Verify against the spec

- [ ] 4.1 Play the defender case: take a hit and click during the freeze, and confirm the counter comes out rather than being swallowed — and that clicking a moment later still behaves the same, so the fast reaction is no longer the punished one.
- [ ] 4.2 Play the attacker case: land a swing and mash the attack button through the hit-stop, and confirm the chain follows as reliably as it does on a whiff.
- [ ] 4.3 Confirm a stun still locks out in play: take the third hit of a run, click during the freeze, and confirm nothing comes out during or at the end of the stun.
- [ ] 4.4 Confirm the fix holds online against a listen server with `F5`–`F8` latency simulation on: a click during hit-stop produces one attack and no correction spike, verifying client and server agree on the freeze.
