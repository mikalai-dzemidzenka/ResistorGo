## 1. The rate

- [x] 1.1 Halve the tick interval between restored energy points in `actor.go` (`EnergyRegenTicks` 4 -> 2) and correct its comment to say one point every 2 ticks, so 30 per second. Verify `go build ./...` succeeds and that `EnergyCalmTicks` is untouched.
- [x] 1.2 Verify by reading `regenEnergy` that nothing else had to change: the wait still gates the refill, and the rate is still one point per interval on the fighter's own tick counter, so a client refills on the same tick as the server.

## 2. The rate is pinned

- [x] 2.1 Update `TestEnergyRegen` for the new timings and check the pool part way through as well as at the end — 60 spent is 30 back after 2.5 seconds and full after 3.5 — so a rate that is too fast fails as well as one that is too slow. Verify it fails against `EnergyRegenTicks = 4`.
- [x] 2.2 Add a check that a fighter with an empty pool can swing again 2 seconds after its last swing, the `A swing is soon affordable again` scenario. Verify it passes, and that it would not have at the old rate.
- [x] 2.3 Verify the rule that holds the refill back is unmoved: the swing-a-second half of `TestEnergyRegen` still ends with the pool empty and nothing regenerated, and `TestEnergyCalmOnlyResetsOnSpending` still passes.
- [x] 2.4 Run `go test ./...` and `go test -tags release ./...` and confirm both pass, including the prediction tests that replay a swing against a snapshot.

## 3. Verify against the spec

- [ ] 3.1 Play a local match (`go run ./cmd`): empty the pool, break off, and confirm the energy bar refills at about twice the speed it did — roughly 7 seconds from empty to full rather than 13, with the first swing available under half a second after the wait ends.
- [ ] 3.2 Confirm in play that pressure still denies recovery: keep swinging about once a second and confirm the pool runs down to a refusal with nothing coming back in between.
