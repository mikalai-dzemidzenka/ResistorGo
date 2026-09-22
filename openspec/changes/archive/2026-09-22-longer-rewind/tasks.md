## 1. A minute of recording

- [x] 1.1 Set `RewindFrames = 60 * 60` and update its comment (a minute of
      fixed steps, not "every actor keeps"); verify the existing
      `rewind_test.go` recording, seek and resume tests and
      `TestCaptureDoesNotAllocate` pass against the larger ring
- [x] 1.2 Add a test for `A minute back` and `Nothing older`: record more than
      3600 steps of an offline match, enter the mode, seek to the oldest
      frame, and verify its tick is exactly 3599 steps before the newest (3600
      frames held, counting both ends)
- [x] 1.3 Add a test for `Memory stops growing`: after one lap and after three
      laps of a four-fighter match, the total capacity of every slot's
      `actors`, `dying` and `pending` slices is the same

## 2. The ramp

- [x] 2.1 Add `ScrubFramesMax = 16`, `ScrubRampSteps = 120` and
      `ScrubRampGrace = 0.1` beside `ScrubFrames`, and `scrubRate(held int) int`
      with the integer formula from design.md; verify with a table test that
      it is 2 at 0, 16 at 120 and beyond, never decreases, and moves in whole
      frames
- [x] 2.2 Add the ramp's state (`scrubHeld`, `scrubStill`, `scrubLast`) and use
      it in `UpdateTime`: seek `dir * scrubRate(scrubHeld)` per fixed quantum
      and count `scrubHeld`; reset it on a reversal and after
      `ScrubRampGrace` of stillness; clear all three on entering and leaving
      the mode and in `ResetTime`; verify `TestScrubStopsWithTheMouse` and
      `TestScrubDirection` still pass unchanged
- [x] 2.3 Rewrite `TestScrubRateIsFixed` as the frame-rate test: the same
      three-second drag at 60 fps and at 150 fps covers the same frames, and
      a two-second drag covers the same frames rewinding as replaying forward
      over enough recording
- [x] 2.4 Rewrite `TestScrubSpeedOfHandDoesNotMatter` for the ramp: a
      one-second sweep and a one-second slow drag, each from rest, cover the
      same frames
- [x] 2.5 Add tests for `A short drag is slow` (a quarter second covers fewer
      than 60 frames), `Holding the drag speeds it up` (the frames covered per
      second rise through the first two seconds, then hold at 16 per step,
      and five seconds reach the whole minute back), `Reversing starts
      slow again`, `A pause starts slow again` (half a second still, then 2
      per step) and `A dropped mouse report does not` (one zero frame
      mid-drag moves nothing that frame and the drag carries on at the
      rate it had; each judged over ten frames, since one rendered frame may
      cover one fixed step or none)
- [x] 2.6 Verify `Jitter does not move time`: sub-pixel `dx` for three seconds
      moves nothing and leaves `scrubHeld` at zero

## 3. The spec's purpose, and the whole

- [x] 3.1 On archiving, change the Purpose in
      `openspec/specs/time-manipulation/spec.md` from "the last few seconds"
      to "the last minute" - a delta cannot carry a Purpose change
- [x] 3.2 Run `go test ./...` and `go vet .`; verify only the known flaky
      engagement and spectator tests ever fail, and at their usual rate

## 4. In the game

- [x] 4.1 Build with `build.ps1`; play a minute or more against the bots,
      press T and rewind all the way: it reaches a minute back in about five
      seconds of dragging, and stops there
- [x] 4.2 Find one exact frame after a long rewind by stopping and nudging
      the mouse: the first nudges step slowly enough to pick it
- [x] 4.3 Watch the process's memory over five minutes of play: it rises for
      the first minute and then stays flat
