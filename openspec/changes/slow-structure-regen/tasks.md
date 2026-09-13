## 1. The rate

- [x] 1.1 Change `StructureRegen` from 2 to 1 in actor.go, keeping `StructureCalmTicks` at 90, and verify `go build ./...` is clean
- [x] 1.2 Confirm `regenStructure` needs no other edit — it already adds `StructureRegen * FixedTimeStep` per tick, so the rate is the only thing that moves

## 2. The tests that encode the old rate

- [x] 2.1 Update `TestStructureRegen`'s first case to the new timing: 4 structure back after 1.5s of waiting plus 4s of refilling (`90+240` ticks), and verify it passes
- [x] 2.2 Add the half-recovered case to the same test: after the 3.5s that once refilled all 4, exactly 2 are back — this is what distinguishes the new rate from the old, where the first case alone would pass at either
- [x] 2.3 Verify the pressure case is untouched: a hit per second still denies recovery entirely, since the 1.5s calm wait has not moved

## 3. Whole-change verification

- [x] 3.1 Run `go test ./...` and confirm the structure, structure-bar, protocol and prediction suites all still pass — `TestShowStructureBar` did depend on the rate incidentally (it refilled one hit in 60 ticks, which only reached full because the old rate overshot into the cap) and was given headroom
- [ ] 3.2 Play a local match (`go run ./cmd`): take a stun run, then break off, and confirm the bar refills at about half the speed it did — roughly 4 seconds to clear three hits' worth rather than 2
