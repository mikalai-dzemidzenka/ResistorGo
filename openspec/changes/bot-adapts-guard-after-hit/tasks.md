## 1. A landed hit can move a bot's guard

- [x] 1.1 Add `GuardAdaptOdds = 3` to the combat constants in `actor.go`, beside `StunEveryHits`, with a comment reading it as one in three and saying it is a feel number to be turned in play. Verify `go build ./...` succeeds.
- [x] 1.2 In `takeHit`, between the `enterDeathState` return and `a.HitCount++`, roll the odds for a bot and move its guard to the attacker's height: gate on `!a.IsPlayer`, set `a.PrevUpdown` only when the value actually changes, then set `a.Updown = attacker.Updown`. Say in the comment why it is bot-only — a player's guard follows their aim, and an ungated write would be read by another attacker's hit test on the same tick. Verify `go build ./...` succeeds and that a killing hit cannot reach the block.

## 2. It does what the spec says and no more

- [x] 2.1 Add a test that a bot's guard both moves and stays: hit a bot at a height it is not guarding many times over, resetting it between hits, and confirm at least one hit moved the guard to the attacker's height and at least one left it alone. Verify it passes, and confirm it asserts on the spread rather than on a particular draw from `math/rand`.
- [x] 2.2 Add a test that a player's guard is never touched: hit an `IsPlayer` fighter the same number of times and confirm its `Updown` is unchanged after every one. Verify it passes.
- [x] 2.3 Add a test that the killing hit adapts nothing: put a bot on the last of its structure, guard it away from the attacker's height, hit it, and confirm it died and its guard did not move. Verify it passes.
- [x] 2.4 Add a test that a stun adapts on the same terms as a flinch: drive a bot to the stunning hit of a run and confirm the guard can move on it. Verify it passes.
- [x] 2.5 Verify `PrevUpdown` stays consistent: after an adaptation that changed the guard, `PrevUpdown` is the height the bot had before it, as `updateUpdown` would have left it.

## 3. Nothing else moved

- [x] 3.1 Verify the wander is untouched: `git diff` shows no edit to `botInput`, so the one-in-sixty-one guard re-roll and the swing odds are as they were.
- [x] 3.2 Verify nothing reached the wire: this change edits only `actor.go`. Note the working tree also carries `fix-hitstop-input-loss`, whose `ProtocolVersion` 6 -> 7 shows in `git diff net_protocol.go`; that bump is not this change's, and `actorRecSize` is still 44.
- [x] 3.3 Verify blocks are untouched: `git diff` shows no edit to `confirmBlock`.
- [x] 3.4 Run `go test ./...` and `go test -tags release ./...` and confirm both pass.

## 4. Verify in play

- [ ] 4.1 Play a local match (`go run ./cmd`): hit the bot repeatedly at one height and confirm it starts guarding there some of the time rather than never, and that the same attack stops being reliable.
- [ ] 4.2 Confirm it is not every hit: watch a run of hits and check the guard sometimes stays where it was, so the bot is still readable rather than a mirror.
- [ ] 4.3 Confirm the transition looks right: watch the bot's guard-change animation when it adapts, and check it plays from the height it actually had rather than snapping or drawing the wrong transition.
- [ ] 4.4 Confirm your own guard is never taken: take hits at every height and check the guard follows the mouse throughout, including on the tick the hit lands.
- [ ] 4.5 Play a networked match against the dedicated server (`go run ./cmd -dedicated` plus a client) with bots: confirm a bot's adapted guard looks the same on the client as on the host under `F5`–`F8` latency simulation, with no correction spike when it changes.
- [ ] 4.6 Judge the one-in-three by eye and record any change to `GuardAdaptOdds` in design.md.
