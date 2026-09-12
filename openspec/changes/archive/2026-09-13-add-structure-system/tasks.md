## 1. Structure pool

- [x] 1.1 Add `Structure float32` to `Actor`, full at 12 on spawn, with the capacity, hit cost, block-all cost, block restore, regeneration rate and regeneration delay as named constants beside the other combat constants. Verify with a unit test that a new actor starts full.
- [x] 1.2 Add `takeHit(attacker)` and `confirmBlock(attacker)` as the only places structure is spent or restored, clamping to 0..12. Verify with unit tests: a hit costs 1, a matching block restores 0.5 and cannot overfill, and block-all costs 0.4 — less than a hit.
- [x] 1.3 Add regeneration in the fixed step: refill at 2 per second once 1.5 seconds have passed with no structure lost, restarting that wait on every loss. Verify with unit tests that 3.5 idle seconds restore 4 structure and that a loss every second prevents any regeneration.

## 2. Flinch, stun and one landing per swing

- [x] 2.1 Add the per-fighter landed-hit counter, reset on stun, on death and when regeneration starts. Verify with a unit test that three hits produce flinch, flinch, stun, and that the fourth flinches again.
- [x] 2.2 Split `enterAttackedState` into the shared flinch (flash, `EvHit`, hit-stop on both, small knockback, hit animation) and the stun (adds the `StateAttacked` lockout and the larger knockback), routing both through `takeHit`. Verify in game that the first two hits leave the defender able to move and attack while the third locks it out and throws it further.
- [x] 2.3 Add `FlinchTicks` so a flinch animation is not overwritten by `StateIdle`'s per-tick animation selection, while movement, attack, jump and block input still work during it. Verify in game that a flinched fighter shows the hit animation and can still swing, and that starting a swing replaces the animation immediately.
- [x] 2.4 Add the attacker's `SwingHit` flag, set by `takeHit`/`confirmBlock` and cleared where a swing starts (`enterAttackState` and the chained-swing branch of `updateAttack`). Verify with a unit test that a swing whose box overlaps a defender for many ticks costs structure once, and that a chained second swing costs again.
- [x] 2.5 Move the block accounting to where a block is confirmed — the swing's hit frame in `updateBlock`, not `enterBlockState`. Verify with a unit test that a guard dropped before the hit frame is charged as a hit and restores nothing.

## 3. Death and respawn

- [x] 3.1 Register `death_fall.png` in `LoadTextures` and append it to `animTable`, leaving existing indices unchanged. Verify the game builds and the texture can be drawn without panicking.
- [x] 3.2 Add `StateDeath` at the end of the `ActorState` consts, entered when structure reaches zero from any source: plays `death_fall.png` once, takes no input, takes no further structure changes, and is skipped by attack resolution. Verify with a unit test that the twelfth hit enters it, that a block-all that empties structure enters it too, and that a swing at a dying fighter resolves nothing.
- [x] 3.3 Flicker for about 0.6 seconds once the death animation finishes, counted in ticks, then remove the fighter from the world on the authoritative side only. Verify in game that the body flickers and disappears, and that clients see both.
- [x] 3.4 Add the pending-respawn list on `World`, checked once per fixed step on the authoritative world: 2 seconds after removal, spawn a replacement at a free position with full structure and cleared combat state, reusing the existing free-position search. Verify in game that a killed bot is fighting again about 2 seconds later.
- [x] 3.5 Re-attach a respawned fighter to its waiting player in `net_server.go`, keeping the player's slot and id while it has no fighter. Verify with two clients that the dead player regains control after respawning and that their slot is not handed to a joining client in the meantime.

## 4. Networking

- [x] 4.1 Add structure to the actor record as a `uint8` of tenths, updating the record size, and raise the protocol version. Verify with the existing protocol round-trip test that structure survives encode/decode to 0.1, and that the record size constant matches what is written.
- [x] 4.2 Confirm a 16-fighter snapshot still fits 1200 bytes with the new field, extending the existing snapshot-size test. Verify that test passes.
- [x] 4.3 Keep structure, stuns, breaks, deaths and respawns authority-only: a client's predicted world must apply none of them locally. Verify with a unit test in the style of the existing non-authority test that a predicted swing deducts no structure and kills nobody.

## 5. Verify against the spec

- [x] 5.1 Add structure and the hit count to the debug overlay next to the other per-fighter debug text. Verify in game that the numbers move as expected while fighting.
- [x] 5.2 Run `go test ./` and confirm the whole suite passes, including the new structure and protocol tests.
- [x] 5.3 Play through the scenarios in `specs/structure/spec.md` that need a running game: flinch on hits one and two, stun on three, structure regenerating between exchanges, blocks restoring and block-all draining, death by hit and by block-all, the flicker and disappearance, and the respawn.
- [ ] 5.4 Play a two-client networked match and confirm structure, stuns, death and respawn agree on both ends, then tune the constants from task 1.1 by feel and record any values that changed.
