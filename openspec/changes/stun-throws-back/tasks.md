## 1. The throw reaches the fighter

- [x] 1.1 Remove the `a.Speed.X = clamp(a.Speed.X, -AttackDash, AttackDash)` line from the `case StateAttacked` arm of the state switch in `Actor.Update`, leaving `a.updateAttacked()`. Leave the identical line on `case StateAttack` alone — that one caps the lunge of the fighter doing the swinging, which is what `AttackDash` is for. Verify `go build ./...` succeeds.
- [x] 1.2 Confirm by reading that nothing else caps a stunned fighter's horizontal speed: `updateHorSpeed`'s `canMove` still excludes `StateAttacked` so the throw cannot be steered, and the only remaining bound is `MaxSpeed`, which 150 is well inside.

## 2. It stays fixed

- [x] 2.1 Add a test that a stun carries the fighter a real, absolute distance — at least ten whole pixels from where it was hit, once the throw has run out. Assert on distance travelled, not on `Speed.X`: a speed assertion passes on a fighter that never moves, which is the bug. Do not compare against a flinch: `PunchSpeed` is under the truncation floor, so a flinch travels 0 px and the comparison would pass on a one-pixel throw. Verify it fails against the current code and passes after task 1.1.
- [x] 2.2 Add a test that the throw actually changes position: a stunned fighter's `Pos.X` after the freeze differs from its `Pos.X` when the hit landed by more than the whole-pixel truncation in `Update`. Verify it fails against the current code and passes after 1.1.
- [x] 2.3 Add a test that the throw follows the attacker's facing: a fighter stunned by an attacker facing left goes left, one stunned by an attacker facing right goes right. Verify it passes.
- [x] 2.4 Add a test that the attacker's cap survives: a fighter in `StateAttack` still has its horizontal speed held to `AttackDash`, confirming the right one of the two identical lines was removed. Verify it passes.

## 3. Nothing else moved

- [x] 3.1 Verify the knockback values are untouched: `git diff` shows no edit to `hitFeedback`, `enterAttackedState`, `enterFlinchState`, `StunKnockback` or `PunchSpeed`.
- [x] 3.2 Verify the lockout is untouched: the stun still ends with its animation, and the existing stun tests — the hit-run and hit-count suites included — pass unchanged.
- [x] 3.3 Verify nothing reached the wire: `git diff` shows no edit to `net_protocol.go` and `ProtocolVersion` is unchanged.
- [x] 3.4 Run `go test ./...` and `go test -tags release ./...` and confirm both pass.

## 4. Verify in play, and decide the number

- [ ] 4.1 Play a local match (`go run ./cmd`): land a third hit and confirm the fighter is visibly thrown, noticeably further than the two flinches before it, and that the stun now reads as the exchange resetting rather than as the defender standing still.
- [ ] 4.2 Judge the distance. The measured throw is 20 px on a 160 px virtual screen, an eighth of the width; decide by eye whether that is a stun or a nudge, and if it is wrong, change `StunKnockback` and record the new value and the reason in design.md.
- [ ] 4.3 Judge the overrun. Friction brings the throw under the 60 px/s truncation floor after about 0.36 s, against a 0.3 s stun animation, so the fighter is still creeping as it regains control. Decide whether that reads as being knocked off balance or as losing the first moment of the recovery, and record the verdict in design.md — if it needs stopping, that is a follow-up change, not this one.
- [ ] 4.4 Confirm the throw behaves at the edges: stun a fighter with its back to a wall and stun one standing at a ledge, and check it stops or falls the way a flinched fighter already does rather than doing anything new.
- [ ] 4.5 Play a networked match against the dedicated server (`go run ./cmd -dedicated` plus a client): confirm a stunned fighter is thrown the same distance on both screens under `F5`–`F8` latency simulation, with no correction spike at the moment of the stun.
