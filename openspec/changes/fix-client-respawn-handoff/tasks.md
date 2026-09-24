## 1. The snapshot names the fighter

- [x] 1.1 Add `Yours uint16` to `Snapshot` beside `Ack`, commented as the receiving player's current fighter and 0 for none; carry it in `encodeSnapshot`/`decodeSnapshot` after `Ack`, and verify a round-trip test preserves it
- [x] 1.2 Update the snapshot header size from 10 to 12 bytes wherever it is assumed (the `encodeSnapshot` capacity hint and `TestActorRecSize`'s `header` constant), and verify `TestActorRecSize` and `TestSnapshotFitsPacket` both pass
- [x] 1.3 Fill it in `sendSnapshots` from each peer's own `p.Actor` (0 when nil), inside the loop that already sets `Ack` from `p.Cmds.last`; verify with a test that two connected players are each named their own fighter and never the other's
- [x] 1.4 Bump `ProtocolVersion` 9 → 10 and verify the existing version-mismatch test still rejects a client one version behind

## 2. The client adopts what it is told

- [x] 2.1 In `onSnapshot`, set `c.ActorID` from `s.Yours` before the loop that reconciles, and verify a test that a snapshot naming a different fighter changes what the client reconciles against
- [x] 2.2 When the named id differs from the one held, drop `c.W.Local` through `World.Remove` so the next `reconcile` takes its existing `local == nil` branch and builds the new fighter; verify the rebuilt fighter carries none of the old one's death state (`DeathTicks`, `Dying`, `Attacker` all clear)
- [x] 2.3 Handle being named `0`: drop the local fighter and predict nothing until a fighter is named again; verify the client has no `W.Local` for the whole gap and does not panic in any per-tick path that reads it
- [x] 2.4 Verify with a test that a client named the same fighter it already holds keeps predicting it untouched — no rebuild, no correction, on every snapshot of an ordinary match

## 3. The local fighter obeys the same lifetime as any other

- [x] 3.1 In `applySnapshot`, build `visible` from the snapshot's own actor ids rather than from the interpolation ids, so the local fighter is represented in it; verify a test that a client's own fighter is never removed while the server is still sending it
- [x] 3.2 Drop the `a.Remote` condition from the removal loop so the local fighter goes when the server stops sending it, and verify a test that the corpse leaves the client's world within two snapshots of the server reaping it

## 4. The flicker cannot run away

- [x] 4.1 Stop `updateDeath` incrementing `DeathTicks` past `DeathFlickerTicks`, leaving the authority's `Kill` on the tick it reaches the limit unchanged; verify a test that the count stops there on a non-authority world
- [x] 4.2 Make `deathHidden` report hidden rather than alternating once the flicker span is over, and verify a test that a dead fighter held past its flicker stops blinking instead of strobing

## 5. Offline hands over too

- [x] 5.1 In `stepOffline`, adopt the player's fighter into `CL.Local` when it has none, in the loop that already tests `actor.IsPlayer`; verify a test that after an offline death and respawn `CL.Local` is the new fighter
- [x] 5.2 Verify with a test that the effects keyed on the local fighter work again after an offline respawn — `presentEffects` sees the new fighter as `local` and its attack raises the camera zoom

## 6. Whole-change verification

- [x] 6.1 Tighten `TestNetDeathAndRespawn` to assert control rather than visibility: after the respawn the client's `W.Local` is the server's new fighter and `c.ActorID` matches it, which is what it misses today
- [x] 6.2 Add an integration test covering the reported bug end to end: a client's fighter dies, the corpse leaves the client's world, the client has no fighter for the gap, and after the respawn it predicts and controls the new one — with `DeathTicks` never past `DeathFlickerTicks`
- [x] 6.3 Run `go test ./...` and confirm the replication, prediction, spectator and session suites all still pass
- [ ] 6.4 Play a networked match against the dedicated server (`go run ./cmd -dedicated` plus a client): die on purpose and confirm the fighter falls, flickers once, disappears, and that two seconds later a new fighter is under control with no blinking at any point
- [ ] 6.5 Play a single-player match (`go run ./cmd`): die on purpose and confirm the respawned fighter is controlled and the attack zoom and hit shake still happen
