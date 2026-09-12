## 1. Recording window

- [x] 1.1 In `rewind.go`, replace `RewindIdx`/`RewindLeft` with the explicit window from design.md — `head`, `count`, `cursor` — plus one helper mapping an absolute frame number to a ring slot with a non-negative modulo. Verify with a unit test that the helper returns a slot in `[0, RewindFrames)` for negative, zero and past-wrap frame numbers.
- [x] 1.2 Add `record()` (advance `head`, grow `count` up to `RewindFrames`, move `cursor` to the new head) and `seek(delta)` (move `cursor`, clamped to `head` at the top and to the oldest valid frame at the bottom). Verify with unit tests: seeking far past either end lands exactly on the bound and never reports an out-of-range slot, and `count` stops growing at `RewindFrames`.
- [x] 1.3 Add `resumeHere()` — drop the frames ahead of `cursor` by setting `head = cursor` and shrinking `count` accordingly. Verify with a unit test that after recording 100 frames, seeking back 30 and calling it, seeking forward stops at the new head and the oldest bound is unchanged.

## 2. Mode state machine

- [x] 2.1 Replace the `IsRewinding` bool with the three-state mode (off / paused / scrubbing plus direction) and a small predicate for "a scrub step is replaying history", which is what `Actor.Update` and `Actor.Draw` actually ask. Verify the package builds and every existing reference to `IsRewinding` compiles against the new predicate.
- [x] 2.2 Add the toggle: a press of R turns the mode on (pausing on the current frame) and a second press turns it off, calling `resumeHere()`. Keep raylib input reading at the call site — the state machine takes "R was pressed" as a parameter so it stays testable without a window. Verify with a unit test that two presses return to the off state and that holding (press then no further press) neither scrubs nor resumes.
- [x] 2.3 Add direction latching from a mouse dx parameter: sign only, with a dead zone constant, and zero direction when the mouse is still. Verify with a unit test covering right, left, still, sub-dead-zone jitter and a vertical-only move (dx of 0).
- [x] 2.4 Drive scrub steps from fixed-step quanta at the rate constant (default 2 recorded frames per fixed step, same both directions), so a given latched direction produces the same number of steps per second at any frame rate. Verify with a unit test feeding 1 s of dt as 60 large steps and as 150 small ones and asserting the same cursor movement.

## 3. Wire into the level loop

- [x] 3.1 In `game.go`, gate `stepOffline` on the mode being off, and stop adding `dt` to the accumulator while it is on. Verify in game that after ten seconds in the mode the match resumes at normal speed with no catch-up burst, and that fighters are untouched while paused.
- [x] 3.2 Move the `UpdateRewind` call to feed the state machine the R press, the mouse dx (`rl.GetMouseDelta().X`) and `dt`, keeping the existing `NetMode == NetOffline` gate so the mode stays single-player only. Verify in a network match that R does nothing.
- [x] 3.3 Update `Actor.updateRewind` to the new mode: record into the current head slot on a live step, restore from the cursor slot on a scrub step, and do nothing while paused so a paused frame neither overwrites history nor re-reads it. Verify in game that pausing for several seconds and then rewinding still reaches the full history.
- [x] 3.4 Update the debug overlay in `actor.go`, which prints the now-removed `RewindIdx` and `RewindLeft`, to show the mode and the cursor's position within the window. Verify the game builds and the overlay reads sensibly while scrubbing in both directions.

## 4. Verify against the spec

- [x] 4.1 Run `go test ./` and confirm the whole suite passes, including the new rewind tests.
- [x] 4.2 Play through each scenario in `specs/time-manipulation/spec.md` that needs a running game: enter and pause, scrub both ways, stop the mouse mid-scrub, hold the mouse right at the live edge, scrub left past the oldest frame, resume from a rewound frame and confirm the discarded future never replays, and confirm movement and attack input do nothing while the mode is on.
- [ ] 4.3 Tune the scrub rate constant by feel in the game and record the value settled on in the change's design notes if it differs from the default of 2.
