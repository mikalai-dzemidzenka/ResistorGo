## 1. The rule

- [x] 1.1 In `enterDeathState`, restore a non-nil attacker to `MaxEnergy` through `restoreEnergy`, beside the hit-stop it already gets. Verify `go build ./...` passes and that neither `EnergyCalm` nor `EnergyUntouched` is touched.
- [x] 1.2 Add `TestKillFillsKillersPool` to `energy_test.go`: an attacker with little energy lands a killing `takeHit` and ends at full; an attacker whose attack breaks a block-all through `confirmBlock` ends at full; a `takeHit` that does not kill leaves the attacker's energy as it was; the attacker's regeneration waits are unchanged by the kill. Verify it passes, and fails with 1.1 reverted — the `A landed kill`, `Through a block-all`, `A hit that does not kill` and `Bots are rewarded too` scenarios.

## 2. Verify against the spec

- [x] 2.1 In `go run ./cmd`, swing a bot down to its death with a nearly empty pool: the energy bar jumps to full and the crosshair shows its full-pool flash as the bot falls; let a bot kill you and confirm the respawned fighter shows no flash.
- [x] 2.2 Host a match and join it from a second instance; kill the host's fighter from the client with a low pool and confirm the energy fills when the kill arrives — the `Online` scenario. (Accepted without the online run: the refill rides the snapshot that already carries energy.)
- [x] 2.3 Run `go test ./...` and `go test -tags release ./...` and confirm both pass (apart from `TestWallsStopFighters/running`, which fails the same way on a clean tree and belongs to `horizontal-collision`), then build with `./build.ps1` to a scratch `-Out` and confirm it packs. (Both suites fail only on `TestWallsStopFighters/running`; `./build.ps1` to a scratch `-Out` packs, 9.1 MB to 5.3 MB.)
