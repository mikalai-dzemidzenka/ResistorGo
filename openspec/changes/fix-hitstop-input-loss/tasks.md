## 1. The freeze holds input instead of eating it

- [x] 1.1 Add `BufAttack` and `BufJump` bools to `Actor` in `actor.go`, beside `FreezeTicks`, with a comment saying they hold an attack or jump pressed while the fighter is frozen and are handed to its first live tick. Verify `go build ./...` succeeds and `go test ./` still passes with no behaviour change.
- [x] 1.2 In `Actor.Update`, capture into the buffer on a frozen tick (`BufAttack |= Input.Attack`, `BufJump |= Input.Jump`) before the existing early return, and on the first unfrozen tick OR the buffer back into `a.Input` and clear it, above the state switch. Say in the comment that a press read below may be up to `HitStopTicks` old, since the six read sites cannot tell. Verify `go build ./...` succeeds and that no read site in the state switch was edited.
- [x] 1.3 Clear `BufAttack` and `BufJump` in `enterDeathState`, on the line that already clears `AttackQueued`. Verify `go build ./...` succeeds.

## 2. The buffer survives a correction

- [x] 2.1 Add `fBufAttack` and `fBufJump` to the flag constants in `net_protocol.go` at bits 13 and 14, after the `HitRun` field, with a comment noting they ride in free bits so the record does not grow. Verify `actorRecSize` is still 44 and `go build ./...` succeeds.
- [x] 2.2 Set both flags in `recOf` beside `fAttackQueued` and read them in `apply`. Verify by reading that `reconcile`'s `r.apply` now restores the buffer before it replays the command ring, so a freeze spanning the acknowledged tick keeps its press through a correction.
- [x] 2.3 Move `ProtocolVersion` 6 → 7. Check first whether `fix-client-respawn-handoff` has already moved it; if so, take the next value and note the ordering in design.md rather than both changes claiming 7.
- [x] 2.4 Extend the protocol round-trip test to carry `BufAttack` and `BufJump` set, and confirm they survive encode and decode. Verify `go test -run TestProtocol ./` passes.

## 3. The bug stays fixed

- [x] 3.1 Add a test that an attack pressed while a fighter is frozen survives every frozen tick and fires on the first live one: set `FreezeTicks`, set `Input.Attack` on the first frozen tick only, step `Update` through the freeze asserting `BufAttack` stands and the fighter has not attacked, then assert it enters `StateAttack` on the tick after. Verify it fails against the current code and passes after task 1.2.
- [x] 3.2 Add the same test for jump: a jump pressed while frozen leaves the ground on the first live tick. Verify it fails against the current code and passes after task 1.2.
- [x] 3.3 Add a test for the attacker's chain: with the fighter in `StateAttack` and frozen, a press during the freeze reaches `AttackQueued` once the freeze ends, exactly as a press later in the same swing would. Verify it passes.
- [x] 3.4 Add a test that one press is one action — a single press captured on one frozen tick produces exactly one attack across the whole freeze and the tick after, not one per tick — and that the buffer is false afterwards. Verify it passes.
- [x] 3.5 Add a test for the stun boundary: a press buffered during the freeze of a stunning hit is dropped by `StateAttacked` and does not fire when the stun ends. Verify it passes, and note in the test why it holds without a condition in the fix — `updateAttacked` reads no input — so a future change that adds one is caught here.
- [x] 3.6 Add a test that death discards the buffer: a press buffered during the freeze of the killing hit leaves `BufAttack` false after `enterDeathState`. Verify it passes.
- [x] 3.7 Add a test that a fighter with exactly `EnergySwing` energy, hit and pressing attack during the freeze, still gets the swing — confirming nothing regenerated or spent while frozen. Verify it passes.

## 4. Nothing else moved

- [x] 4.1 Verify the input layer is untouched: `git diff` shows no edit to `sampleInput`, `latchInput`, `matchHeld` or `clearLatchedInput`, and the paused-mode tests still pass unchanged — `TestMatchHeld`, `TestLeavingTimeManipulationDropsLatchedInput` and `TestLeavingEditModeDropsLatchedInput`.
- [x] 4.2 Verify `AttackQueued` is unchanged: `git diff` shows no edit to its assignment in `updateAttack` or its consumption in the `AnimFinished` branch, so chaining inside a swing behaves exactly as before.
- [x] 4.3 Verify no packet grew: `actorRecSize` is still 44 and the snapshot size assertions in the replication tests pass unchanged.
- [x] 4.4 Run `go test ./...` and `go test -tags release ./...` and confirm both pass.

## 5. Verify against the spec

- [ ] 5.1 Play the defender case (`go run ./cmd`): take a hit and click during the freeze, and confirm the counter comes out rather than being swallowed — and that clicking a moment later still behaves the same, so the fast reaction is no longer the punished one.
- [ ] 5.2 Play the attacker case: land a swing and mash the attack button through the hit-stop, and confirm the chain follows as reliably as it does on a whiff.
- [ ] 5.3 Play the jump case: take a hit and press W during the freeze, and confirm the fighter leaves the ground as the freeze ends.
- [ ] 5.4 Confirm a stun still locks out in play: take the third hit of a run, click during the freeze, and confirm nothing comes out during or at the end of the stun.
- [ ] 5.5 Confirm chaining feels unchanged: run a few chains with no hit involved and check the click still registers from anywhere in the swing, not only late in it.
- [ ] 5.6 Play a networked match against the dedicated server (`go run ./cmd -dedicated` plus a client): confirm a click during hit-stop produces one attack on both screens with no correction spike, under `F5`–`F8` latency simulation, and that an older client is refused at the handshake.

## Before archiving

- [x] Merge the `structure` delta by hand, not by replacing the block: `move-while-swinging` (archived 2026-09-23) rewrote "A landed hit costs structure and flinches" after this delta was written. Keep its knockback paragraph ("The knockback SHALL move the fighter ...") and its two scenarios (`The small knockback actually moves the fighter`, `A flinch stays smaller than a stun`) alongside this change's edits; verify with `openspec validate --specs --strict`
