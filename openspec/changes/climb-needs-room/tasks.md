## 1. Pin the bug

- [x] 1.1 In `horizontal_collision_test.go`, next to `TestLedgeAgainstAWall`,
      add `TestNoLedgeUnderACeiling` for both facings: a ceiling bar over a stem
      and over the fighter, a fighter jumping at the stem through `Update`.
      Assert no climb, the head never up through the bar, and back on the
      floor. Verify it fails today on the stem's ledge
- [x] 1.2 Dropped: a bar only over the stem is a taller wall with room on top,
      and climbing onto it is right (the "A wall two tiles tall" scenario)
- [x] 1.3 A two-tile stacked wall with nothing above: a fighter jumping at it
      ends up snapped onto the top tile, not the lower one; verify it passes
      today and after

## 2. Fix

- [x] 2.1 In `catchLedge`, compute the collision box the snap would give (feet
      on `c.Y`, X per facing) and refuse the ledge if it overlaps any solid
      that is not drop-through, before touching `Pos` or `CanClimb*`, per
      design.md - Decisions; verify 1.1-1.3, `TestClimbLedge` and
      `TestNoClimbThroughCeiling` pass
- [x] 2.2 Add a case with a drop-through platform crossing the ledge's standing
      spot: the ledge is still caught; verify it passes
- [x] 2.3 Run `go test ./...`; confirm `bot_course_test.go`,
      `bot_step_test.go`, the climb-down tests and the net tests pass

## 3. Verify in the real game

- [ ] 3.1 Build with `.\build.ps1` and play: build a T in the editor, jump
      beside the stem from both sides, and the fighter stays under the bar; a
      two-tile wall and the ledges in `level_1.json` still climb as before
