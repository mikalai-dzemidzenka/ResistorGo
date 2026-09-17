## 1. The side check

- [x] 1.1 Add `StepHeight = 4` to the movement constants in `actor.go`. Add the wall test (a solid overlapping the band from the head down to `StepHeight` above the feet) and `blockSides(fromX)`, as the design describes: limit the leading side to `max(wall face, where it started)` for every wall ahead, set `Pos.X` from the limit, and zero `Speed.X` only when it points into the wall. Verify unit tests that call `blockSides` directly: it stops at the face moving right and moving left; a one-pixel wall stops a move of `MaxSpeed*FixedTimeStep`; a fighter already overlapping a wall is not moved and cannot move deeper; a move away from a wall is untouched; a rectangle whose top is 4 px above the feet is not a wall and one 5 px above is.
- [x] 1.2 Split the move at the end of `Update` into X then Y: `Pos.X += Speed.X*dt`, truncate, `blockSides(fromX)`, then `Pos.Y += Speed.Y*dt`, truncate. Verify tests driving `Update` over a floor with a wall on it: a walking fighter and a running one stop against the wall, the runner with zero speed and out of the run state; a knocked-back fighter stops against it; a fighter steering in the air into the wall stops beside it and goes on falling. Verify `jump_test.go`, `hit_run_test.go` and `actor_test.go` still pass unchanged.
- [x] 1.3 In `enterAttackState`, record X before the forward dash and call `blockSides` after it. Verify a test: a fighter standing against a wall and facing it starts a swing from idle, is in the attack state, and its X is unchanged. Also verify that the same swing away from the wall still dashes `AttackDash` px.

## 2. Steps and ledges

- [x] 2.1 Verify with tests driving `Update`: a fighter walking onto a rectangle 4 px above its floor ends up standing on it (`OnGround`, feet on its top) without jumping; one 5 px up stops it; and a fighter running across a floor of equal-height tiles never loses speed at a seam. Also load `level_1.json` and run a fighter the length of its main floor row and back, checking it is never stopped: the proposal's *The current level plays as before*.
- [x] 2.2 Verify with tests driving `Update`: a fighter facing a wall taller than its jump reach, jumping against it, is stopped by the wall and then catches the ledge (enters the climb) when its head reaches the top edge. A fighter falling beside a wall and facing it catches the ledge on the way down. The tests are the spec's *Jumping at a wall* and *Falling past a wall* scenarios.
- [x] 2.3 Verify with a test that a fighter placed with its box overlapping a wall's side is not moved sideways over a tick with no input, and can walk out of it: the spec's *Spawned overlapping* scenario.

## 3. Online

- [ ] 3.1 Following `net_collision_rects_test.go`, add a networked test: a floor with a wall on it, and the client's own fighter held running into the wall. Verify that the fighter stops against the wall on the client and the server at the same X, and that `c.Corrections` is zero: the spec's *Online against a wall* scenario.

## 4. Verify against the spec

- [ ] 4.1 Run `go test ./...` and `go test -tags release ./...` and confirm both pass.
- [ ] 4.2 Build with `./build.ps1` to a scratch `-Out`, and with `./build.ps1 -Linux`, and confirm both pack.
- [ ] 4.3 In `go run ./cmd`, use the editor to give a tile a collision rectangle taller than a step, standing on a floor. Check that you run into it and stop, jump against it and climb it, walk up a 4 px step, and that `level_1.json` plays as before. Note how a fighter held against the wall looks, which is the design's first risk.
