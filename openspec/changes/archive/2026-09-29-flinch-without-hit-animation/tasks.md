## 1. The hit animation belongs to the stun

- [x] 1.1 Take the `SetAnim` call out of `hitFeedback` in `actor.go`, leaving it the shared part of a landing hit — attacker, hit event, hit-stop on both fighters, knockback, flash — and say in its comment that the animation is the stun's alone. Verify `go build ./...` succeeds.
- [x] 1.2 Play `hit-top.png` from `enterAttackedState` instead, so the stun is the only path that sets it. Verify by reading that no other call site sets that animation (`grep -n "hit-top" *.go`), and that `TestEveryThirdHitStuns` still passes.
- [x] 1.3 Leave `enterFlinchState` setting only the flinch hold, and verify a flinched fighter's `Anim` is whatever it was before the hit.

## 2. The flash decides itself

- [x] 2.1 Add a predicate on `Actor` that answers whether the fighter is drawn through the flash: false without the flash flag, the opening frames of the animation while in the hit-stun state, true otherwise. Verify by reading that every field it touches is one the snapshot carries, so a remote fighter answers the same as the server's.
- [x] 2.2 Have the draw loop in `game.go` ask that predicate instead of testing the flash flag and the animation frame itself. Verify `go build ./...` succeeds and that no other draw path tests the flag (`grep -rn "AnimWhite" *.go`).
- [x] 2.3 Add one call that clears the flinch hold and the flash together, and use it everywhere the hold was being zeroed — starting a swing, starting a block — leaving the stun and death paths to set the flash themselves afterwards. Verify a fighter that swings out of a flinch is not left flashing.
- [x] 2.4 Remove the idle branch that held a flinching fighter's animation, so a flinch no longer suppresses the fighter's own animation changes. Verify a running fighter keeps its running animation across a flinch.

## 3. The behaviour is pinned

- [x] 3.1 Add a test that the first two hits of a run leave the defender's own animation alone and flicker, and that the third plays the hit animation and flickers over its opening — the `Only the stun has an animation` scenario. Verify it fails against a build where `hitFeedback` still sets the animation.
- [x] 3.2 Add a test that the flash runs out with the flinch hold and not before, stepping the fighter tick by tick — the `The flash runs out` scenario. Verify it passes.
- [x] 3.3 Add a test that swinging out of a flinch ends the flash — the `Acting ends the flash` scenario. Verify it fails if the clearing call from task 2.3 is reverted to zeroing the hold alone.
- [x] 3.4 Verify the existing structure tests still hold unchanged: `TestEveryThirdHitStuns`, `TestFlinchKeepsControl` and `TestOneLandingPerSwing` — a flinch still keeps control, still costs 1 structure and still throws less far than a stun.
- [x] 3.5 Run `go test ./...` and `go test -tags release ./...` and confirm both pass, the protocol round-trip tests included — nothing was added to the wire.

## 4. Verify against the spec

- [ ] 4.1 Play a local match (`go run ./cmd`): take a hit while running and confirm the run continues and the fighter flashes, then take the third hit and confirm the hit animation plays and control is lost.
- [ ] 4.2 Confirm in play that the flash never sticks: swing, block and jump out of a flinch, and die during one, checking the fighter is drawn normally afterwards in each case.
- [ ] 4.3 Play a networked match against the dedicated server (`go run ./cmd -dedicated` plus a client): confirm a remote fighter's flinches and stuns look the same on the client as they do on the host, with `F5`–`F8` latency simulation on.

## Before archiving

- [ ] Merge the `structure` delta by hand, not by replacing the block: `move-while-swinging` (archived 2026-09-23) rewrote "A landed hit costs structure and flinches" after this delta was written. Keep its knockback paragraph ("The knockback SHALL move the fighter ...") and its two scenarios (`The small knockback actually moves the fighter`, `A flinch stays smaller than a stun`) alongside this change's edits; verify with `openspec validate --specs --strict`
