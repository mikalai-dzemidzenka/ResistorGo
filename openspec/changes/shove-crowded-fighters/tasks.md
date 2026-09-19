> A first implementation, on the withdrawn premise, is saved in `recovered.md`:
> the state, `crowding`, `shove`, and `push_test.go`. Start from it, but the
> trigger, the direction and the distance differ - see design.md - Decisions,
> and the `// CHANGES:` markers in that file.

## 1. Shared reach

- [x] 1.1 Lift the reach computation out of `inSwingRange` into
      `swingReachFacing(t, o)`, keeping the block-all fallback; verify
      `inSwingRange` is a one-line caller and `go test ./...` passes unchanged
- [x] 1.2 Add a test sweeping `dx` over `[-20, 20]` for both facings that asserts
      the crowding window (overlap and not `swingReachFacing`) is exactly
      `[-13, -3]` facing right and `[3, 13]` facing left

## 2. The state

- [x] 2.1 Append `StatePush` to the `ActorState` enum before `stateCount`, append
      `resources/threefour/push.png` to the animation table and register it in
      `LoadTextures`; verify `TestAnimTableIsAppendOnly` (updated to pin the
      previous last entries and hold `push.png` to the end) and
      `TestEveryStateSurvivesTheWire` pass
- [x] 2.2 Add `PushSeconds`, `PushFrames`, `PushFrame`, `PushDistance = 16` and
      `PushSpeed`, and the `PushDone` field; verify with a test that for every
      offset in the crowding window, offset + `PushDistance` clears the overlap
      and stays inside the swing's reach
- [x] 2.3 Add `crowding()`: an overlapping, living fighter this one cannot reach
      facing the way it faces; verify the `Crowded from behind`, `Overlapping but
      in front` and `Only the fighter that cannot reach shoves` scenarios
- [x] 2.4 Hook it into `Actor.Update` before the state switch, gated on
      `StateIdle`, on the ground and not guarding everything; verify the `A chase
      does not set it off`, `A shove comes off the feet` and `Guarding everything
      stands its ground` scenarios
- [x] 2.5 Add `enterPushState(other)`, which faces the crowder and starts
      `push.png`, `updatePush` returning to standing when the animation ends, and
      the `StatePush` case; exclude `StatePush` from `canMove`; verify the
      `Turning to shove`, `No movement during the shove` and `The shove ends`
      scenarios, including that the mouse-facing of a standing player does not
      turn a shoving one

## 3. The shove itself

- [x] 3.1 Add `shove(other)`: move `Pos.X` away from the shover by
      `PushDistance`, set its speed, then `blockSides`; verify the `The crowder
      goes away from the shover`, `The two end up apart` and `The fight stays in
      reach` scenarios across every offset in the window, both facings
- [x] 3.2 Land it once, on `PushFrame`, only where `W.Authority`; verify the `A
      shove lands once` and `A client moves nobody` scenarios
- [x] 3.3 Verify the `Nothing else changes hands`, `A fighter mid-swing is still
      moved` and `The dead are left where they lie` scenarios
- [x] 3.4 Verify the `Hit while shoving` scenario: a hit landing before
      `PushFrame` puts the shover into `StateAttacked` and the displacement never
      happens; fix the hit path if it treats `StatePush` specially

## 4. Walls

- [x] 4.1 Add a `walls()` test helper beside `floor()` that builds a floor with a
      vertical face; verify a fighter walked into it rests flush against it
- [x] 4.2 Using `walls()`, crowd a fighter whose crowder is flush against the
      wall; verify the `A wall stops the throw` scenario - no second shove, and a
      tick where input is read within a bounded number of ticks - and that the
      crowder is left with no speed into the wall

## 5. Reconcile what the shove disturbs

- [x] 5.1 Space the fighters in `jump_test.go` apart so the jump tests measure
      jumping; verify they pass and still assert what they did - not needed:
      the two fighters there are running, and only a fighter standing still
      shoves, so they pass unchanged
- [x] 5.2 Run `go test ./...` at least three times; verify no failure, and that
      `TestABotFollowsAFighterTheLengthOfTheCourse` and the spectator tests pass
      every time - narrow any assertion that a shove now legitimately disturbs
      to what that test is actually about

## 6. Verify in play

- [x] 6.1 Build with `build.ps1` (never a plain `go build`) and confirm the
      packed `resistor.exe` is replaced and runs
- [ ] 6.2 Let a bot walk into you, from in front and from behind, standing and
      running; confirm the fighter turns to it, both arms go out, the bot is
      thrown to arm's length and the fight carries on face to face
- [ ] 6.3 Stand flush against a wall with a bot crowding in; confirm nobody locks
      into repeated shoves
- [ ] 6.4 Play against the dedicated server with latency simulation (`F5`-`F8`)
      and watch for a predicted turn the server does not make; if it shows,
      gate the crowding check on `W.Authority`, record the decision in
      design.md, and verify the `A client moves nobody` scenario still holds

## 7. The shove stuns (found in play: the shove did not finish)

- [x] 7.1 Reproduce the unfinished shove headlessly with a real bot walking onto
      a standing player's back; verify the two cut-offs - a block after landing,
      and the crowder's own swing landing inside the wind-up
- [x] 7.2 Add `enterShovedState`: the lockout and animation of a stun, ending a
      swing or a block, with no structure, hit count, flash, `EvHit` or hit-stop,
      and the shove's own speed; apply it on the authority when the shove starts;
      verify the `Stunned, not hurt`, `No hit is shown` and `A client moves
      nobody` scenarios
- [x] 7.3 Verify the `A swing already thrown does not cut the shove off`
      scenario with a crowder just short of its hit frame at the shover's guard
      height, and that the test fails with the stun moved to the landing frame;
      verify the reproduction then finishes every shove it throws
- [ ] 7.4 In play, confirm the shoved bot reads as stunned rather than hit, and
      that the shove now plays to its end

## 8. Any overlap is shoved (asked for in play)

- [x] 8.1 Trigger on any overlap of two fighters on the ground, in front, behind
      or on the same column, from standing, walking or running; not from a
      swing, a block, a stun, a flinch, the air or block-all; verify `Any
      overlap is shoved` and `Only a fighter on its feet and free shoves`,
      including that exactly one of two overlapping fighters shoves
- [x] 8.2 Skip a fighter in the air as a target; verify `Not the one in the air`
      and that `TestABotFollowsAFighterTheLengthOfTheCourse` - where a bot on a
      ledge was shoving a climber out of the air - passes repeatedly
- [x] 8.3 Re-derive `PushDistance`'s bounds for a crowder starting 0 to 13px
      away (14 to 21; 16 kept); verify the distance and arm's-length sweeps
      cover every overlapping offset in both facings
- [x] 8.4 Reconcile the tests the wider trigger disturbs - the jump runners
      spaced apart, the bot attack target updated so its stun wears off, the
      wall test's crowder guarding everything so it cannot part the two itself;
      verify `go test ./...` passes three times running
- [x] 8.5 `push.png` grew to 10 frames: set `PushFrames` 10 and `PushSeconds`
      0.5, keeping three ticks a frame and the landing on frame 3, where the
      arms are first fully out; verify the suite passes and the reproduction
      still lands every shove it throws
- [x] 8.6 Carry the crowder instead of moving it 16px in one tick: a whole-pixel
      ease-out glide (6, 4, 2, 2, 1, 1) on the crowder, wall-checked every step,
      replacing the displacement and `PushSpeed`; remember the fighter shoved
      from the landing on, so the overlap that lasts into the glide is not
      shoved again; verify `Carried there, not put there` and `A wall ends the
      carry`, and that the suite passes three times running
