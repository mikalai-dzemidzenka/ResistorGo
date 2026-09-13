## 1. The pool

- [x] 1.1 Add the energy constants to the `Structure`/`MaxStructure` block in actor.go — `MaxEnergy = 200`, `EnergySwing = 12`, `EnergyBlockRestore = 15`, `EnergyRegenTicks = 4` (one point per 4 ticks is 15/s at 60 Hz), `EnergyCalmTicks = 90` — each with a comment in the style of the structure constants, and verify `go build ./...` is clean
- [x] 1.2 Add `EnergyCalm uint16` to `Actor` beside `CalmTicks` (field named like `CalmTicks`, against the `EnergyCalmTicks` constant), commented as ticks since energy was last *spent* (structure's counts ticks since structure was last *lost*), and set `Energy: MaxEnergy` in `NewActor` so every spawn path starts full; verify a new actor reports the cap
- [x] 1.3 Write `canSwing()`, `spendEnergy()` (deducts `EnergySwing`, resets `EnergyCalm`) and `restoreEnergy(n)` (clamped to `MaxEnergy`, does not touch the calm counter) next to `spendStructure`/`restoreStructure`, and verify a unit test covering the clamp at both ends passes

## 2. Spending it

- [x] 2.1 Change `enterAttackState()` to return bool: refuse and raise `EvNoEnergy` when `!canSwing()`, otherwise spend and enter the state as it does now; verify it returns false and changes no field when energy is below `EnergySwing`
- [x] 2.2 Update the two call sites in `Actor.Update` (`StateIdle`, `StateRun`) to `if a.Input.Attack && ... && a.enterAttackState() { return }` so a refusal falls through to the rest of that state's logic instead of returning; verify a fighter refused mid-run keeps running
- [x] 2.3 Add `EvNoEnergy` to the presentation-event bits beside `EvAttackStart`/`EvHit`/`EvBlocked`
- [x] 2.4 Gate the chain in `updateAttack`: at the `if !a.AttackQueued` branch, treat an unaffordable next swing as a released button — clear `AttackQueued`, raise `EvNoEnergy`, play the back-attack recovery; spend on the swing that does continue; verify a test holding attack from full energy gets exactly a pool's worth of swings and then the recovery
- [x] 2.5 Verify with a test that a swing costs once no matter how long its box overlaps a defender, and costs the same whether it lands, is blocked, or hits nothing

## 3. Earning it back

- [x] 3.1 Grant `EnergyBlockRestore` in `confirmBlock`'s matching-guard branch only, leaving the `Updown == 3` branch untouched; verify a test shows a matching guard at 40 energy goes to 55 and block-all at 40 stays at 40
- [x] 3.2 Add `regenEnergy()` beside `regenStructure` — `EnergyCalm` counting up to `EnergyCalmTicks`, then one point every `EnergyRegenTicks` ticks, and no authority check: a client predicts its refill (see design.md) — and call it from `Actor.Update` next to `regenStructure()`; verify a test that a fighter that spent 60 is back at full after 5.5 s and that a swing per second never regenerates
- [x] 3.3 Verify with a test that being hit and blocking do not restart the energy calm wait, and that only spending does

## 4. Over the wire

- [x] 4.1 Add `Energy uint8` and `EnergyCalm uint8` to `ActorRec`, bump `actorRecSize` 42 → 44, and carry both in `recOf` and `apply` (clamping with `min()` as `HitCount` does); verify a round-trip test preserves every energy value from 0 to 100 exactly and the refill wait with it
- [x] 4.2 Carry both bytes in `encodeSnapshot`/`decodeSnapshot` beside `Structure`, and verify the existing 16-fighter packet-size test still passes under 1200 bytes
- [x] 4.3 Bump `ProtocolVersion` 4 → 5 and verify the existing version-mismatch test still rejects an older client
- [x] 4.4 Verify on the prediction harness that a client holding attack through an empty pool throws the refilled swing on the same tick the server does, with no mispredicted swing and no energy mismatch, and that reconciliation replaying a swing leaves the pool 20 below the snapshot's value, not 40

## 5. On screen

- [x] 5.1 Add `EnergyBarHeight`, `EnergyBarDrop` and an exact-palette `energyBarFill` to structure_bar.go, and verify the energy bar matches the structure bar's width
- [x] 5.2 Add `showEnergyBar(a)` — `Energy < MaxEnergy`, not in the Death state, not while `TimeShowsHistory()` — deliberately separate from `showStructureBar`; verify a test that a fighter at full structure that has swung shows the bar, and one at full energy with a wrecked structure shows none
- [x] 5.3 Add the energy track and fill to `structureBarRects`/`structureBarLayout`: same width as the structure bar, directly below it, filled `roundf(StructureBarWidth * Energy / MaxEnergy)` from the left; verify a layout test that the fill never extends past either end at any energy and that every edge is a whole pixel
- [ ] 5.4 Draw the energy bar in `drawStructureBar`, and verify in the running game that it appears on the first swing, drains as swings are spent, and grows on a matching block
- [ ] 5.5 Consume `EvNoEnergy` in `presentEffects` into a short local-only deadline on the actor, and draw the empty energy track while it runs; verify in the running game that mashing on an empty pool flashes the bar

## 6. Whole-change verification

- [x] 6.1 Run `go test ./...` and confirm the existing structure, protocol, prediction and menu suites still pass alongside the new energy tests
- [ ] 6.2 Play a local match (`go run ./cmd`): confirm sixteen swings from full then refusal, that a matching block funds the next swing, that block-all funds nothing, and that a fighter left alone recovers
- [ ] 6.3 Play a networked match against the dedicated server (`go run ./cmd -dedicated` plus a client): confirm the refusal is immediate under latency, that the pips of a remote fighter track the server's value, and that no swing appears on a client that the server refuses
