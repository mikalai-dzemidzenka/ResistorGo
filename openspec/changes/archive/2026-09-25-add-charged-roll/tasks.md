## 1. Register the sprites and the states

- [x] 1.1 Add `resources/threefour/duck.png` (4 frames), `roll.png` (5) and `roll-recover.png` (4) to the texture list in `resources.go`, and append all three to the end of `animTable`. Verify that the existing `animTable` round-trip test still passes, and that no existing path changed its index.
- [x] 1.2 Append `StateDuck`, `StateRoll` and `StateRollRecover` after `StateClimbDown` and before `stateCount` in `actor.go`. Verify `go build ./...` succeeds, and check by reading that no existing state's number moved.

## 2. The roll button, end to end

- [x] 2.1 Add `Roll bool` to `Input`, and set it in `sampleInput` from `rl.IsMouseButtonDown(rl.MouseButtonRight)`. Verify `go build ./...` succeeds.
- [x] 2.2 Add `bRoll` to the button bits in `net_protocol.go`, and encode and decode it in `encodeCmds`/`decodeCmds`. Verify with a test that a command with `Roll` set survives the round trip, and that the test fails without the bit.

## 3. Duck, charge, launch

- [x] 3.1 Add the roll constants beside the other movement constants: `RollChargeTicks`, `RollMinX`, `RollMaxX`, `RollMinY`, `RollMaxY`. Add `rollLaunch(a) rl.Vector2` in a new `roll.go`. Verify with a test that the launch at tick 1 is the minimum, that a longer hold gives a larger |X| and a more negative Y, and that holds of 60, 180 and 600 ticks give the same result.
- [x] 3.2 Add `enterDuckState` in `roll.go`. It refuses with `EvNoEnergy` when `Energy < EnergyRunJump`. Otherwise it zeroes `Speed.X` and plays `duck.png` once with `reset: true`, so `AnimStartTick` marks the start of the hold. Call it from the idle and run branches when `Input.Roll && OnGround`, before the attack and jump checks. Verify with tests that a standing fighter and a running fighter both enter `StateDuck` with `Speed.X == 0`, and that a fighter below the cost stays in its state with its energy unchanged and `EvNoEnergy` raised.
- [x] 3.3 In `updateHorSpeed`, add `StateDuck` and `StateRollRecover` to the states that cannot move, and add `StateDuck` to the "face the cursor" condition next to `StateIdle`. Verify with a test that a ducking fighter holding a direction key does not move, and that one whose `FaceLeft` flips turns around.
- [x] 3.4 Add the duck's case to the state switch. When `!Input.Roll`, it spends `EnergyRunJump`, sets `Speed = rollLaunch(a)` and `OnGround = false`, and enters `StateRoll` playing `roll.png` in a loop. It does not return, so the take-off tick moves the fighter. If the fighter is no longer `OnGround`, it leaves the duck and lets the fall take over. Verify with tests that energy is unchanged while ducking, drops by `EnergyRunJump` on the release tick, and that the fighter is off the ground on that tick.

## 4. In the air and landing

- [x] 4.1 Make `updateHorSpeed` return early for `StateRoll`, and make `updateFalling` skip `StateRoll`. Verify with a test that a rolling fighter holding the opposite key keeps its horizontal speed until it lands, and is still in `StateRoll` after its apex.
- [x] 4.2 In the roll case, enter `StateRollRecover` on `OnGround`, playing `roll-recover.png` once. In the recovery case, call `land()` on `AnimFinished`. Verify with a test that a roll on flat ground goes Roll → RollRecover → Idle, and that attack or jump pressed during the recovery does nothing.
- [x] 4.3 Verify with a test that a hold of 15 ticks and a hold of 45 ticks, both on flat ground, give a higher peak and a farther landing for the longer hold. This is the scenario "Longer is higher and farther".

## 5. Hits interrupt it

- [x] 5.1 Add `rolling()` for the three states, and make the swing's hit loop treat a rolling fighter as not guarding. Verify with a test that a ducking fighter whose `Updown` matches the attacker's is hit rather than blocking, leaves the duck for `StateAttacked`, and spends no roll energy.

## 6. The arc

- [x] 6.1 Add `rollPath(a) []rl.Vector2` to `roll.go`. It steps a released copy of the fighter through the real `Update` until the roll lands, capped at `RollPathTicks`. Verify with a test that replays a real roll from the same state on a flat test level and lands within 1 px of the path's last point.
- [x] 6.2 Add `drawRollArc(a)`, which draws dots along the path in a palette colour. Call it in `game.go` for `w.Local` next to the `drawAttackOutline` loop, only when `w.Local.State == StateDuck`. Verify `go build ./...` succeeds. This is checked visually in 7.2.

## 7. Whole-change verification

- [x] 7.1 Run `go test ./...` and `go test -tags release ./...`, and confirm both pass.
- [ ] 7.2 Build with `build.ps1` and play `go run ./cmd`. Hold RMB and check that the duck plays and holds, and that the arc grows for about a second and then stops growing. Release and check that the roll follows the arc, the roll animation loops in the air, and `roll-recover.png` plays on landing. Check that the energy bar drops by a third, and that with an empty bar RMB flashes it instead of ducking. Tune the `Roll*` constants by eye.
- [ ] 7.3 Play against `go run ./cmd -dedicated` with a client, including with `F5`–`F8` latency. Check that the roll launches on release with no correction snap, and that a remote fighter's duck, roll and recovery look right.
