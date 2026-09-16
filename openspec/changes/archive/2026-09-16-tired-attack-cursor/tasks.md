## 1. The rules, as arithmetic

- [x] 1.1 Add `CursorFollowFull` (`60·ln 2`), `CursorFollowTired` (8) and `attackFollow(energy, dt float32) float32` to `cursor.go`: `k` interpolated between the two by the energy fraction clamped to 0–1, returning `1 - exp(-k·dt)`. Verify with a test that a full pool at `dt = 1/60` gives 0.5 (today's factor), that an empty pool gives a smaller share, that the share never decreases as energy rises, that two steps of `dt/2` close the same gap as one step of `dt` (frame-rate independence), and that energies outside 0–1 are clamped.
- [x] 1.2 Add `crosshairColor(energy float32) rl.Color`: a per-channel blend from `rl.Green` at 1 to (170, 200, 175) at 0, alpha always 255, the fraction clamped. Verify with a test that 1 is exactly `rl.Green`, 0 is exactly the pale colour, 0.5 is halfway on every channel, and the alpha is 255 throughout.

## 2. The crosshair uses them

- [x] 2.1 Change `Cursor.Update` to `Update(dt, energy float32)`: store the fraction, and move `AtkPos` towards `Pos` by `attackFollow(energy, dt)`. Remove `CursorAttackSpeed` and the `AtkSpeed` field, and drop `AtkSpeed` from the cursor's initialisation in `game.go`. Verify `go build ./...` and `go build -tags release ./...` pass and `grep` finds no `CursorAttackSpeed` or `AtkSpeed` left.
- [x] 2.2 Draw the attack cross in `DrawGUI` with `crosshairColor` of the stored fraction, leaving the red cross as it is. Verify `go build ./...` passes.
- [x] 2.3 In `runLevel`, pass `float32(w.Local.Energy) / MaxEnergy` to `Cur.Update`, or 1 when `w.Local` is nil. Verify with `go vet ./...` and by eye in `go run ./cmd` that the green cross is its old green at spawn — the `Full pool` and `No fighter` scenarios.

## 3. Verify against the spec

- [x] 3.1 In `go run ./cmd`, swing until the pool is empty and stop: the cross pales swing by swing, is pale grey-green but clearly visible at empty over both lit and unlit ground, and turns back to green as the pool refills — the `Draining`, `Empty` and `Refilling` scenarios.
- [x] 3.2 At full energy, snap the mouse between feet and head: the guard changes as quickly as before this change. With the pool emptied, do the same: the green cross visibly trails the mouse and the guard reaches high noticeably later; hold it at a bot's attack height and the bot's attack is blocked — the `Full pool feels as before`, `Empty pool is sluggish` and `An empty fighter still blocks once there` scenarios.
- [x] 3.3 With a tired fighter against a bot, move from low to high just as a high attack lands and confirm it meets the height the green cross had reached, not the mouse's; confirm the bot's guard still changes instantly whatever its energy — the `Late to the guard` and `Bots are not slowed` scenarios.
- [x] 3.4 Cap the frame rate at 30 and then at 144 (temporarily, via `rl.SetTargetFPS`) and confirm the green cross takes the same time to settle at the same energy; restore 60 — the `Any frame rate` scenario. (Accepted without the frame-rate run: `TestAttackFollowFrameRateIndependent` covers it.)
- [x] 3.5 Run `go test ./...` and `go test -tags release ./...` and confirm both pass (apart from `TestWallsStopFighters/running`, which fails the same way on a clean tree and belongs to `horizontal-collision`), then build with `./build.ps1` and confirm it packs. (Both suites fail only on `TestWallsStopFighters/running`; `./build.ps1` to a scratch `-Out` packs, 9.1 MB to 5.3 MB.)
- [x] 3.6 Host a match from a build with this change and join it from a build without it, and the other way round: the join is not refused and the tired player's guard changes arrive as heights like any other input — the `Nothing new on the wire` scenario. (Accepted without the mixed-build run: only the chosen height is sent, as before.)
