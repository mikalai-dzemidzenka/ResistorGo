> A first implementation, on the withdrawn premise, is saved in `recovered.md`:
> the state, `crowding`, `shove`, and `push_test.go`. Start from it, but the
> trigger, the direction and the distance differ - see design.md - Decisions,
> and the `// CHANGES:` markers in that file.

## 1. Shared reach

- [ ] 1.1 Lift the reach computation out of `inSwingRange` into
      `swingReachFacing(t, o)`, keeping the block-all fallback; verify
      `inSwingRange` is a one-line caller and `go test ./...` passes unchanged
- [ ] 1.2 Add a test sweeping `dx` over `[-20, 20]` for both facings that asserts
      the crowding window (overlap and not `swingReachFacing`) is exactly
      `[-13, -3]` facing right and `[3, 13]` facing left

## 2. The state

- [ ] 2.1 Append `StatePush` to the `ActorState` enum before `stateCount`, append
      `resources/threefour/push.png` to the animation table and register it in
      `LoadTextures`; verify `TestAnimTableIsAppendOnly` (updated to pin the
      previous last entries and hold `push.png` to the end) and
      `TestEveryStateSurvivesTheWire` pass
- [ ] 2.2 Add `PushSeconds`, `PushFrames`, `PushFrame`, `PushDistance = 16` and
      `PushSpeed`, and the `PushDone` field; verify with a test that for every
      offset in the crowding window, offset + `PushDistance` clears the overlap
      and stays inside the swing's reach
- [ ] 2.3 Add `crowding()`: an overlapping, living fighter this one cannot reach
      facing the way it faces; verify the `Crowded from behind`, `Overlapping but
      in front` and `Only the fighter that cannot reach shoves` scenarios
- [ ] 2.4 Hook it into `Actor.Update` before the state switch, gated on
      `StateIdle`, on the ground and not guarding everything; verify the `A chase
      does not set it off`, `A shove comes off the feet` and `Guarding everything
      stands its ground` scenarios
- [ ] 2.5 Add `enterPushState(other)`, which faces the crowder and starts
      `push.png`, `updatePush` returning to standing when the animation ends, and
      the `StatePush` case; exclude `StatePush` from `canMove`; verify the
      `Turning to shove`, `No movement during the shove` and `The shove ends`
      scenarios, including that the mouse-facing of a standing player does not
      turn a shoving one

## 3. The shove itself

- [ ] 3.1 Add `shove(other)`: move `Pos.X` away from the shover by
      `PushDistance`, set its speed, then `blockSides`; verify the `The crowder
      goes away from the shover`, `The two end up apart` and `The fight stays in
      reach` scenarios across every offset in the window, both facings
- [ ] 3.2 Land it once, on `PushFrame`, only where `W.Authority`; verify the `A
      shove lands once` and `A client moves nobody` scenarios
- [ ] 3.3 Verify the `Nothing else changes hands`, `A fighter mid-swing is still
      moved` and `The dead are left where they lie` scenarios
- [ ] 3.4 Verify the `Hit while shoving` scenario: a hit landing before
      `PushFrame` puts the shover into `StateAttacked` and the displacement never
      happens; fix the hit path if it treats `StatePush` specially

## 4. Walls

- [ ] 4.1 Add a `walls()` test helper beside `floor()` that builds a floor with a
      vertical face; verify a fighter walked into it rests flush against it
- [ ] 4.2 Using `walls()`, crowd a fighter whose crowder is flush against the
      wall; verify the `A wall stops the throw` scenario - no second shove, and a
      tick where input is read within a bounded number of ticks - and that the
      crowder is left with no speed into the wall

## 5. Reconcile what the shove disturbs

- [ ] 5.1 Space the fighters in `jump_test.go` apart so the jump tests measure
      jumping; verify they pass and still assert what they did
- [ ] 5.2 Run `go test ./...` at least three times; verify no failure, and that
      `TestABotFollowsAFighterTheLengthOfTheCourse` and the spectator tests pass
      every time - narrow any assertion that a shove now legitimately disturbs
      to what that test is actually about

## 6. Verify in play

- [ ] 6.1 Build with `build.ps1` (never a plain `go build`) and confirm the
      packed `resistor.exe` is replaced and runs
- [ ] 6.2 Stand still with the mouse facing away and let a bot walk onto your
      back; confirm the fighter turns, both arms go out, the bot is thrown to
      arm's length and the fight carries on face to face
- [ ] 6.3 Stand flush against a wall with a bot crowding in; confirm nobody locks
      into repeated shoves
- [ ] 6.4 Play against the dedicated server with latency simulation (`F5`-`F8`)
      and watch for a predicted turn the server does not make; if it shows,
      gate the crowding check on `W.Authority`, record the decision in
      design.md, and verify the `A client moves nobody` scenario still holds
