## 1. Branch and registration

- [x] 1.1 Create a `climb-down` branch from an up-to-date `master` and bring the untracked `resources/threefour/climb-down.png` onto it. Verify with `git status`, which should show the branch and the PNG.
- [x] 1.2 Append `resources/threefour/climb-down.png` to the end of `animTable` in `resources.go`, after `battle-low-1-right.png`, and make sure `LoadTextures` loads it. Add it to `TestAnimTableIsAppendOnly` with its new ID. Verify with `go test -run TestAnimTableIsAppendOnly`.

## 2. Drop-through rectangles in the level data

- [x] 2.1 Add `Drop []rl.Rectangle` to `ObjectRects` and `DropRects` to `LevelObject`. Carry them through `clone`, `sameRects`, `rects()`/`setRects` and `ObjectRects.list`/`count`. Verify that `go build ./...` succeeds.
- [x] 2.2 Add `dropRect`/`dropRects` to `levelObjectWire` and `objectRectsWire`, after the light keys, marshalled with `rectsToWire` and read with `rectsFromWire`. Add a test like `TestCollisionRectsOnDisk` that round-trips one drop rectangle and then two. Check that `TestLevel1KeepsItsBytes` and `TestOneCollisionRectKeepsItsBytes` still pass with no change.
- [x] 2.3 In `buildIndex`, append each object's drop rectangles to `LE.solids` straight after its collision rectangles, and fill a parallel `LE.drop []bool`. Extend `TestSolidsFollowTheCollisionOrder` to check the order and that the flags line up.
- [x] 2.4 Add a test that a level with a drop rectangle has a different fingerprint from the same level without it. Verify with `go test ./...`.

## 3. Editor: the third rectangle kind

- [x] 3.1 Add `rectDrop` after `rectLight`. Pick it with **P** in the key handling, next to C and L, with P also stepping to the next drop rectangle the way C and L do. Add it to the kinds loop the click hit-test walks, and change the rect-tool panel's toggle group to `"Collision (C);Light (L);Drop (P)"` with a "Drop n of m (P: next, N: add)" label. Add editor tests like `TestStepAddRemoveCollisionRects`, `TestClickPicksACollisionRect` and `TestAddAndClearRect`, for drop rectangles.
- [ ] 3.2 Draw drop rectangles in the rect tool and in the debug overlay in a colour distinct from both the collision and the light outlines. Verify by eye in `go run ./cmd` with a tile that has all three kinds.
- [x] 3.3 Extend `TestEveryCollisionRectTravels` (and the paste and follow tests it relies on) so a texture's drop rectangles travel through placing, copy and paste, following, and save and load. Verify with `go test ./...`.

## 4. Collision: floor from above, solid from below

- [x] 4.1 In `checkCollision`, for rectangles marked in `LE.drop`: apply the ground branch only when the feet were at or above the top before this tick's move (`feet - Speed.Y*dt <= top+1`). Apply the head and ledge branches only while rising (`Speed.Y < 0`). Make `headBlocked` skip drop rectangles unless the fighter is rising. Add actor tests that a fighter walks across a drop rectangle, lands on one from above, has its head stopped by one from below, catches a drop rectangle's ledge on the way up, and falls past one when placed below its top with no upward speed.
- [x] 4.2 Add `a.onDropThrough()`: true only when at least one rectangle holds the fighter's centre column at its feet and every one that does is marked drop. Test it on a drop rectangle, on a solid, and on a solid and a drop rectangle sharing a top.

## 5. Input and the command

- [x] 5.1 Add `Down` to `Input`. Latch `rl.IsKeyPressed(rl.KeyS)` into `pressedDown` in `latchInput`, hand it on and clear it in `sampleInput`, and clear it in `clearLatchedInput`. Verify that `go build ./...` succeeds and that bots still send `Down == false` (`bot_wire_test.go` still passes).
- [x] 5.2 Add `bDown` (after `bFaceLeft`) to `encodeCmds`/`decodeCmds`, and extend the command round-trip test in `net_protocol_test.go` so `Down` survives it. Verify with `go test -run Cmd`.

## 6. The climb-down state

- [x] 6.1 Append `StateClimbDown` after `StateClash`, before `stateCount`, and add the constants `ClimbDownSeconds = 0.5`, `ClimbDownFrames = 5` and `ClimbDownDrop = 40` with a short comment on what the 40 lines up. Add `enterClimbDownState`: set the state, zero `Speed.X`, call `stopFlicker`, and play `climb-down.png` once with `reset`. Check that the snapshot decoder accepts the new state (existing `stateCount`-based protocol tests still pass).
- [x] 6.2 In the `StateIdle` and `StateRun` arms, after the attack and jump checks, enter the climb-down when `Input.Down && OnGround && onDropThrough()`. Exclude `StateClimbDown` from `canMove` in `updateHorSpeed`. Add tests: S on a drop rectangle enters the state on that tick; S on a solid does nothing; S from a run enters it with the fighter stopped; S held through the climb-down does not start a second one; S while swinging does nothing and is not acted on when the swing ends.
- [x] 6.3 Add the `StateClimbDown` arm: hold `Speed` at zero, and on `AnimFinished` call `finishClimbDown`. That moves the fighter down `ClimbDownDrop` px, or onto the highest top closer than that under its centre column, sets `Speed.Y = 0`, and calls `enterFallState`. Add tests: movement, attack and jump held during the climb-down change nothing and the animation reaches its last frame; after it the fighter is in `StateFall` below the platform's top; it lands on open floor below; with a solid 20 px below it ends up standing on that solid; through a stack of two it lands on the second and can climb down again.
- [x] 6.4 Add tests for interruptions: a stun (`enterAttackedState`) and a shove during the climb-down leave the fighter stunned on top of the platform; a flinch leaves the climb-down running to the fall; a death leaves the fighter on top. Verify with `go test ./...`.

## 7. Network

- [x] 7.1 Add a networked test (like `net_integration_test.go`) on a level with a drop rectangle. A client presses S, and the server fighter climbs down and falls past the platform. The client's predicted fighter matches it with no correction back onto the platform, and a second client sees the climb-down animation ID on that fighter. Verify with `go test -run` for the new test.

## 8. Play and ship

- [ ] 8.1 In a development build, give a tile in `level_1.json` (or a new one) a drop-through rectangle. Play: stand, walk and land on it; bump it from below; catch its ledge; press S to climb down and fall. Tune `ClimbDownSeconds` and `ClimbDownDrop` by eye until the fall picks up from the last climb-down frame without a jump.
- [x] 8.2 Run `go test ./...` and `go vet ./...`, then build with `build.ps1` (UPX) without truncating its output. Confirm that `resistor.exe` was replaced, and commit with a short lowercase message.
