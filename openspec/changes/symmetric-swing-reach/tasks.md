> This change lands **before** `shove-crowded-fighters`, so nothing here touches
> the shove. The reach asymmetry exists without it and is corrected on its own
> terms; the shove then arrives on corrected reach and never has the wall
> deadlock at all. The tasks that guard the deadlock, the `walls()` test helper
> and the in-play wall check live in that change.

## 1. Pin the current behaviour before changing it

- [ ] 1.1 Add a test that measures the near and far reach limits of a swing in
      both facings by sweeping `dx` against `swingReachFacing`, and assert they
      mirror; verify it fails today reporting 38 forward / 32 backward
- [ ] 1.2 Add a test that two fighters at exactly `dx = 0` reach one another in
      both facings; verify it fails today only for the right-facing fighter

## 2. Correct the reach

- [ ] 2.1 Change `attackBoxFacing` to mirror the attack box about the collision
      box's centre with `d = 0`, per design.md - Decisions; verify tests 1.1 and
      1.2 now pass, reporting 35 both ways

## 3. Re-check what consumes reach

- [ ] 3.1 Add an assertion that a bot stops within swing range of its target
      rather than short of or past it, in both facings; verify it passes with
      the corrected reach
- [ ] 3.2 Resolve design.md's open question: determine whether `tooClose` needs
      its own adjustment or falls out of the corrected reach, and record the
      answer in a comment beside it
- [ ] 3.3 Run the full suite (`go test ./...`) and confirm `bot_course_test.go`,
      `jump_test.go` and the net tests still pass
- [ ] 3.4 Confirm server-side lag-compensated hit tests use the same corrected
      box: verify the lag-comp tests pass and that no reach constant is
      duplicated outside `attackBoxFacing`

## 4. Verify in the real game

- [ ] 4.1 Build with `build.ps1` (never a plain `go build`) and play a match:
      confirm the forward swing still feels right 3px shorter, and that the
      backward swing no longer feels short
- [ ] 4.2 In debug mode, confirm the drawn attack-box outline is mirror-symmetric
      about the collision box when a fighter turns around
