## 1. Register the two animations before anything uses them

- [x] 1.1 Add `resources/threefour/run-jump.png` and `resources/threefour/fall.png` to the texture loader in `resources.go`, so drawing them does not panic. Verify `go build ./...` succeeds and that both files exist at those paths (6 frames and 3 frames of 48 pixels respectively).
- [x] 1.2 Append both paths to the end of `animTable`, never beside their relatives, so no existing animation's network id changes. Verify by reading that every path before them keeps its index, and that `animID` no longer panics for either.
- [x] 1.3 Add a test that both paths round-trip through `animID`/`animPath` and that the ids of the animations that were already in the table are unchanged. Verify it fails if either new path is inserted rather than appended.

## 2. The two new states

- [x] 2.1 Append `StateRunJump` and `StateFall` to the end of the `ActorState` enum, after `StateDeath`, so the values the snapshot carries for the existing states are untouched. Verify by reading that no existing state's number moved.
- [x] 2.2 Add `enterRunJumpState`: commit the jump, spend its energy and play `run-jump.png` over the 6 frames it has, staying on the ground; and `launchRunJump`, the push-off on frame index 2, setting the same upward speed the standing jump sets and leaving horizontal speed alone. Verify `go build ./...` succeeds.
- [x] 2.3 Add `enterFallState`: set the state and play `fall.png` over its 3 frames, looping, since a fall lasts as long as it lasts. Verify it does not reset horizontal speed, so steering and momentum survive it.

## 3. Entering the air, and falling

- [x] 3.1 In the run branch of the state switch, begin the run-jump on the jump input, on the same tick, without the separate wind-up state; push off from the state's own case when the animation reaches frame index 2, reading the button only at the start so the jump is committed once begun. Verify a test that the fighter is on the ground for frames 0 and 1, leaves on frame 2, and leaves even if the button is released before then.
- [x] 3.2 Add the fall transition before the state switch: a fighter off the ground with a downward vertical speed enters the fall, from whatever state it is in. Exclude the climb, the death state and the stun lockout, so it cannot steal an animation from a fighter that is not free to act. Verify `go build ./...` succeeds.
- [x] 3.3 Verify with a test that every route into the air reaches the fall: a run-jump past its apex, a standing jump past its apex, and a fighter that walked off a ledge — the first three scenarios of the fall requirement. Verify the walked-off-a-ledge case fails against the old code, where the idle branch carried its TODO.
- [x] 3.4 Add a test that a rising fighter is not falling, so the transition reads the direction of travel and not merely being airborne. Verify it fails if the downward test is dropped.
- [x] 3.5 Remove the `// TODO state fall` comment from the idle branch, now that the transition covers it. Verify nothing else referred to it.

## 4. What the air allows, and landing

- [x] 4.1 Verify with a test that a run-jump keeps the run's horizontal speed at take-off, and that a fighter jumping at full speed covers more ground than one at half — the momentum requirement's two scenarios.
- [x] 4.2 Verify with a test that a fighter steers in both jumps and while falling, including reversing direction in the air, and that a fighter that walked off a ledge steers as it falls — the steering requirement. Confirm by reading that this needs no change to the movement rule, which already allows it.
- [x] 4.3 Add the landing transition out of the fall: on reaching the ground, run if still moving and stand if stopped, without locking the fighter out. Verify a test for both, and that an attack pressed on the landing tick comes out.
- [ ] 4.4 Keep the climb reachable from the fall: test the same ledge flags the jump tests, and verify with a test that a fighter reaching a ledge while falling climbs.

## 5. The run-jump costs half the pool

- [x] 5.1 Add the cost as a constant beside the other energy constants in `actor.go`, defined as half of `MaxEnergy` rather than as the number 100, so retuning the pool carries the jump with it. Verify `go build ./...` succeeds.
- [x] 5.2 Give `spendEnergy` the amount to spend instead of hard-wiring it to a swing's cost, and pass a swing's cost at its existing call sites. Verify the energy tests still pass unchanged — `TestEnergyRegen`, `TestEnergyCalmOnlyResetsOnSpending` and the swing-cost tests — so nothing about swinging moved.
- [x] 5.3 Make the run-jump report whether it happened, the way starting a swing does: refuse when the pool is short, raise the same refusal event, spend nothing and leave the fighter in the state it was in. Verify a refused jump leaves the fighter running and still on the ground.
- [x] 5.4 Have the run branch take the refusal: a jump that cannot be paid for leaves the fighter running rather than silently doing nothing else. Verify `go build ./...` succeeds.
- [x] 5.5 Add a test that two run-jumps come out of a full pool and leave it at 100 then 0, and that a third is refused — the `Two from a full pool` and `Not enough to leave the ground` scenarios. Verify it fails if the jump is free.
- [x] 5.6 Add a test that a refused jump raises the refusal event and takes nothing, and that the standing jump is still free at 0 energy — the `The refusal is shown` and `The standing jump is free` scenarios. Verify both pass.
- [x] 5.7 Add a test that a run-jump restarts the refill wait, as a swing does, and that a fighter can still swing after one — the `A jump delays the refill` and `Swinging is still affordable after one` scenarios. Verify they pass.

## 6. Whole-change verification

- [ ] 6.1 Verify the wire is unchanged for everything that already existed: run the protocol round-trip and replication tests, and confirm a fighter in each new state and animation survives `recOf`/`apply` with its state and animation intact.
- [ ] 6.2 Verify the standing jump is untouched: a fighter jumping from a standstill still plays the wind-up, still leaves the ground when it finishes, and reaches the same height as a run-jump.
- [x] 6.3 Run `go test ./...` and `go test -tags release ./...` and confirm both pass.
- [ ] 6.4 Play a local match (`go run ./cmd`): run, jump, and confirm the energy bar drops by half, that a third jump on an empty pool is refused with the bar flashing, that the take-off is immediate, the running-jump animation plays, the fighter steers in the air, the fall animation takes over at the apex, and landing at speed continues the run.
- [ ] 6.5 Confirm in play that walking off a ledge falls, that a ledge caught on the way down still climbs, and that a fighter knocked off an edge falls without disturbing the hit reaction.
- [ ] 6.6 Play a networked match against the dedicated server (`go run ./cmd -dedicated` plus a client): confirm a remote fighter's run-jump and fall look the same on the client as on the host, with no correction spike at take-off under `F5`-`F8` latency simulation.

## 7. The decoder's bound outran the states appended past it

Found after the fact: section 2 appended `StateRunJump` and `StateFall` to the
end of the enum, which was right, but `decodeSnapshot` spelled its validity check
`a.State > uint8(StateDeath)`. Every snapshot carrying a fighter in either new
state was rejected as malformed - and rejection throws away the whole snapshot,
so every other actor in it went too. A fighter falling anywhere in the match
stalled replication for everyone.

- [x] 7.1 Add `stateCount` as a sentinel one past the last state in the
      `ActorState` enum, with a comment saying what it is for; verify `go build
      ./...` succeeds and that no existing state's number moved
- [x] 7.2 Change `decodeSnapshot`'s bound from `a.State > uint8(StateDeath)` to
      `a.State >= uint8(stateCount)`; verify a snapshot carrying a fighter in
      `StateRunJump` or `StateFall` is now accepted
- [x] 7.3 Add a test that sweeps every state from zero to `stateCount` through
      `encodeSnapshot`/`decodeSnapshot` and asserts each survives with its state
      intact, and that `stateCount` itself is still refused; verify it fails
      against the old bound - the `Every state round-trips`, `A state that does
      not exist is refused` and `Appending a state does not break the check`
      scenarios
- [x] 7.4 Verify no other validity check in `net_protocol.go` names a state or an
      animation by value where it means the count; grep for `StateDeath` and for
      `len(animTable)` and confirm each use is deliberate
- [ ] 7.5 Play a networked match and confirm a remote fighter that run-jumps or
      falls no longer freezes every other fighter on the client for the duration;
      verify the client's correction counter does not spike when anyone falls
