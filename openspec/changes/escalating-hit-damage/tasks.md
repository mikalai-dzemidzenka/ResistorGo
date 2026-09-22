## 1. The run and what it costs

- [x] 1.1 Add the escalation constants beside the structure constants in `actor.go`: the step a hit climbs by, the ceiling it stops at, and the run length at which the cost has capped. Derive the run length from the step and the ceiling rather than writing a second number, so retuning the curve cannot leave the two disagreeing. Verify `go build ./...` succeeds.
- [x] 1.2 Add the run to `Actor` beside the stun count, commented to say what tells them apart: the stun count clears every third hit, the run does not. Verify by reading that nothing else in the file already means this.
- [x] 1.3 Add one function that reports what the next hit on a fighter costs, from the run as it stands before that hit: the base, plus the step for each hit already in the run, never past the ceiling. Verify a test that it returns 1, 1.5, 2, 2.5, 3 and then 3 again.

## 2. Hits climb, blocks break it

- [x] 2.1 In `takeHit`, spend what that function reports rather than the flat cost, then count the hit into the run. Verify by reading that the cost is taken before the count moves, so the first hit of a run costs the base.
- [x] 2.2 Clear the run in `confirmBlock`, on both paths — a guard at the attacker's height and block-all — so any stopped attack ends it. Verify a test that a hit after either block costs the base again.
- [x] 2.3 Clear the run where the stun count is already cleared on recovery, in `regenStructure`, and say in the comment why both go together. Verify a test that a hit after recovery costs the base.
- [x] 2.4 Clear the run on death, beside the other counts a dying fighter drops, so a respawned fighter starts clean. Verify a test that a respawned fighter's first hit costs the base.

## 3. The run reaches every client

- [x] 3.1 Carry the run in the spare flag bits above the stun count, clamped on both sides of the wire to the run length the cost caps at. Verify `go build ./...` succeeds and that the record is still 44 bytes.
- [x] 3.2 Bump `ProtocolVersion`, since those bits meant nothing to the build before this one. Verify the existing version-mismatch test still refuses a client one version behind.
- [x] 3.3 Add a test that the run round-trips, is clamped when it is longer than the cost needs, and does not disturb the stun count or the flags it shares the word with. Verify it fails if the run is given the stun count's bits.
- [x] 3.4 Verify with a test that a broken run replicates as well as a growing one: a fighter that blocks out of a run carries a run of zero in the next record.

## 4. The whole curve

- [x] 4.1 Add a test for the full run against a fresh fighter: five hits cost 1, 1.5, 2, 2.5 and 3, leaving 2 structure, and a sixth kills — the `A run costs more as it goes` and `Six hits end a fight` scenarios. Verify it fails against a flat cost.
- [x] 4.2 Add a test that the run survives the stun: the fourth hit of a run costs 2.5 although the third stunned — the `The stun does not relieve it` scenario. Verify it fails if the run is cleared where the stun count is.
- [x] 4.3 Verify the existing structure tests still hold, with their expectations updated only where the escalation is what changed them: `TestEveryThirdHitStuns`, `TestFlinchKeepsControl`, `TestOneLandingPerSwing` and the block tests still describe the same behaviour.
- [ ] 4.4 Run `go test ./...` and `go test -tags release ./...` and confirm both pass.

## 5. Verify against the spec

- [ ] 5.1 Play a local match (`go run ./cmd`): take a run of hits and confirm the bar drops further each time, that six in a row ends it, and that blocking once puts the next hit back to a small step.
- [ ] 5.2 Confirm in play that block-all breaks a run as well as a matching guard, and that walking away long enough to start recovering does too.
- [ ] 5.3 Play a networked match against the dedicated server (`go run ./cmd -dedicated` plus a client): confirm the bar drops by the same amount on both screens under `F5`–`F8` latency simulation, with no correction spike as a run escalates, and that an older client is refused at the handshake.

## Before archiving

- [ ] Merge the `structure` delta by hand, not by replacing the block: `move-while-swinging` (archived 2026-09-23) rewrote "A landed hit costs structure and flinches" after this delta was written. Keep its knockback paragraph ("The knockback SHALL move the fighter ...") and its two scenarios (`The small knockback actually moves the fighter`, `A flinch stays smaller than a stun`) alongside this change's edits; verify with `openspec validate --specs --strict`
