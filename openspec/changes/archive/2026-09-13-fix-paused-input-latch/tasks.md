## 1. One rule, one helper

- [x] 1.1 Add `matchHeld()`, the single predicate for "a paused mode is holding the match" — true while edit mode is on or time manipulation is not off — and `clearLatchedInput()`, which drops the latched attack and jump, beside `latchInput` in `actor.go`. Verify with a unit test that `matchHeld` is false in normal play, true in edit mode, true while time manipulation is on, and false again once both are off.
- [x] 1.2 In `runLevel`, gate the input latch on `!spectating && !matchHeld()` and the fixed-step loop on `!matchHeld()`, replacing the two hand-maintained condition lists. Verify that `go build ./...` succeeds and that `go test ./` still passes, including the existing edit-mode and time-manipulation tests.
- [x] 1.3 Have both pause paths drop stale input through the shared helper: leaving time manipulation in `UpdateTime` now clears it, and edit mode's hand-written clearing line in `toggleEditing` becomes a call to `clearLatchedInput()`. Verify that `go build ./...` succeeds.

## 2. The bug stays fixed

- [x] 2.1 Add a regression test mirroring the existing `TestLeavingEditModeDropsLatchedInput`: with time manipulation on and the attack latched, leaving the mode clears the latch, so the first step back sends no attack. Verify it fails against the current code and passes after task 1.3.
- [x] 2.2 Verify the existing edit-mode test still passes unchanged, so routing it through the shared helper did not change edit mode's behaviour.

## 3. Verify against the spec

- [x] 3.1 Run `go test ./` and `go test -tags release ./` and confirm both pass.
- [x] 3.2 Play the scenarios that need a running game: press R, click the attack button several times while the match is held, press R again, and confirm the fighter does not attack on resuming and nothing is pending; then confirm an ordinary click still attacks in normal play, and that edit mode still drops its clicks.
