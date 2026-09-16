## 1. The rate

- [x] 1.1 Halve the tick interval between restored energy points in `actor.go` (`EnergyRegenTicks` 4 -> 2) and correct its comment to say one point every 2 ticks, so 30 per second. Verify `go build ./...` succeeds and that `EnergyCalmTicks` is untouched.
- [x] 1.2 Verify by reading `regenEnergy` that nothing else had to change: the wait still gates the refill, and the rate is still one point per interval on the fighter's own tick counter, so a client refills on the same tick as the server.

## 2. The rate is pinned

- [x] 2.1 Update `TestEnergyRegen` for the new timings and check the pool part way through as well as at the end — 60 spent is 30 back after 2.5 seconds and full after 3.5 — so a rate that is too fast fails as well as one that is too slow. Verify it fails against `EnergyRegenTicks = 4`.
- [x] 2.2 Add a check that a fighter with an empty pool can swing again 2 seconds after its last swing, the `A swing is soon affordable again` scenario. Verify it passes, and that it would not have at the old rate.
- [x] 2.3 Verify the rule that holds the refill back is unmoved: the swing-a-second half of `TestEnergyRegen` still ends with the pool empty and nothing regenerated, and `TestEnergyCalmOnlyResetsOnSpending` still passes.
- [x] 2.4 Run `go test ./...` and `go test -tags release ./...` and confirm both pass, including the prediction tests that replay a swing against a snapshot.

Sections 1 and 2 doubled the single rate the game had then. The pool has since been retuned to two rates (commit `32cebbf`), which undid that doubling; sections 3 and 4 double both of the current rates.

## 3. Both rates, twice as fast

- [x] 3.1 Set `EnergyRefillSeconds` 20 → 10 and `EnergyFreeRefillSeconds` 10 → 5 in `actor.go`, and correct the rate comments to one point every 2 ticks (30 per second) and every tick (60 per second). Verify `go build ./...` passes and that `EnergyCalmTicks`, `EnergyUntouchedTicks` and `regenEnergy` are untouched.
- [x] 3.2 Add `TestEnergyRegenRates` pinning the rates in points per second — 30 under attack and 60 left alone — independently of the tick constants, so a retune that moves them fails. Verify it fails against the old `EnergyRefillSeconds = 20`.
- [x] 3.3 Verify `TestEnergyRegen`, `TestEmptyPoolAffordsASwingAgain`, `TestEnergyRegenSlowsUnderAttack`, `TestEnergyFreeRateFromSpawn` and `TestEnergyCalmOnlyResetsOnSpending` pass at the new rates, and that the swing-a-second half of `TestEnergyRegen` still ends with nothing regenerated — the `Recovering between exchanges`, `Part way back`, `Refilling under attack`, `A swing is soon affordable again`, `A combo denies recovery`, `Being hit does not stop recovery` and `Fresh fighters refill fast` scenarios.
- [x] 3.4 Run `go test ./...` and `go test -tags release ./...` and confirm both pass, including the prediction tests that replay a swing against a snapshot (apart from `TestWallsStopFighters/running`, which fails the same way on a clean tree and belongs to `horizontal-collision`). (Both suites fail only on `TestWallsStopFighters/running`; `./build.ps1` to a scratch `-Out` packs, 9.1 MB to 5.3 MB.)

## 4. Verify against the spec

- [x] 4.1 Play a local match (`go run ./cmd`): empty the pool, break off, and confirm the energy bar refills from empty in about 5 seconds after the wait when nothing is attacking the fighter, and in about 10 while a bot keeps swinging at it.
- [x] 4.2 Confirm in play that pressure still denies recovery: keep swinging about once a second and confirm the pool runs down to a refusal with nothing coming back in between.
