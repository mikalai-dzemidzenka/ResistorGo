## 1. Position keeps its fraction

- [x] 1.1 Remove the `Pos.X = float32(int(Pos.X))` truncation in `Actor.Update`
      and update the comment that states the whole-pixel rule for X; verify
      with a test that a fighter given a steady 20 px/s with friction held off
      moves 20 px (plus or minus one) in 60 ticks
- [x] 1.2 Add `WalkStopSpeed = 60` and the released-walk stop in
      `updateHorSpeed` (no key held, `State` Idle or Run, `FlinchTicks == 0`,
      `ShoveLeft == 0`); verify `A released walk stops short of a glide`: a
      full-speed walk released stops with its speed zeroed once under 60 px/s,
      travelling about 7 px less than the same walk with the stop disabled
- [x] 1.6 Verify `The same walk everywhere`: a one-second walk in each
      direction covers the same distance at x = 100 and at x = -300
- [x] 1.3 Verify `A knockback runs out` and the structure scenarios: a flinched
      standing fighter keeps sliding below 60 px/s and ends somewhere it was
      not (`The small knockback actually moves the fighter`), and a flinch
      travels visibly less than a stun (`A flinch stays smaller than a stun`)
- [x] 1.4 Verify `Drawn on whole pixels`: `drawPos` of a fighter at a fractional
      position is whole
- [x] 1.5 Run `go test ./...`; for each test that now fails on an exact
      position, decide whether the new value is the correct one and update the
      expectation, or whether it has found a real regression; list each in the
      task notes. Only `TestAttackerIsStillCappedToTheDash` failed: it pinned
      the old 4 px/s cap and now pins `AttackMoveSpeed`, renamed
      `TestAttackerIsStillCappedWhileSwinging`

## 2. Moving while swinging

- [x] 2.1 Add `AttackMoveSpeed = 20`; admit `StateAttack` in `canMove` without
      calling `face` while attacking, and replace the attack case's clamp to
      `AttackDash` with `AttackMoveSpeed`; verify `Pressing in` and `Very
      slow` (a whole low swing with a key held moves at most 6 px beyond the
      dash) and that the fighter's facing never changes during the swing
- [x] 2.2 Verify `Backing out of reach mid-swing` and `Distance decides a
      swing` with two fighters driven through `Update`: the backing fighter
      is missed, the one holding no key is hit
- [x] 2.3 Verify block-all still pins: a fighter guarding everything in an
      attack does not move under input
- [x] 2.4 Verify `A bot creeps too`: a bot swinging while its steering points at
      a target just out of reach moves toward it at the capped pace
- [x] 2.5 Re-run the engagement, clash and shove suites; confirm a line of bots
      still keeps two at the front and none creeps into the fighter it waits
      behind

## 3. Knockbacks after the floor

- [x] 3.1 Re-measure the stun's throw now that its tail moves; record the
      number in design.md and, if it has grown past what reads as a throw,
      lower `StunKnockback` to bring it back to about 20 px, with a test
      pinning the new distance
- [x] 3.2 Re-measure the shove's total travel including the friction after the
      glide; update the shove tests that pin it if it changed. Unchanged: a
      shove zeroes the crowder's speed and glides it in whole pixels, so it has
      no friction tail to grow

## 4. The whole

- [x] 4.1 Run `go test ./...` and `go test -tags release ./...`; verify only
      the known flaky engagement and spectator tests ever fail
- [ ] 4.2 Once archived, drop `flinch-knockback-is-lost-to-rounding`: this
      change carries its requirement and settles its open question

## 5. In the game

- [x] 5.1 Build with `build.ps1`; walk both ways on both sides of the level:
      the pace is the same everywhere now, where one direction used to be
      faster; judge whether that pace, and the stop, feel right
- [x] 5.2 Trade blows with a bot: step back during your swing and see its blow
      miss when you were near the edge of its reach; press in and see yours
      land from a little further out
- [x] 5.3 Take single hits: the flinch now visibly nudges you; a stun still
      reads as a throw, clearly further
- [x] 5.4 Tune `AttackMoveSpeed` by eye: slow enough that a swing still reads as
      planted, fast enough that distance can be won
