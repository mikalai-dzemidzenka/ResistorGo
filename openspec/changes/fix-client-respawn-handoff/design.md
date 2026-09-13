## Context

See proposal.md — Why for the defect and its measurement. What shapes the fix is where the client's idea of "my fighter" lives:

- `c.ActorID` is written in exactly one place, `net_client.go:228`, from the `Welcome` packet, under `if !c.Welcomed` — so it is set once per connection and never again. `Welcome` is the only message in the protocol carrying an actor id.
- Everything downstream keys off it. `onSnapshot` reconciles the record whose `r.ID == c.ActorID` (`net_client.go:259`); `applyInterp` skips that id so the local fighter is predicted rather than interpolated (`net_client.go:335`); `reconcile` creates `c.W.Local` the first time it runs and applies to it forever after.
- `applySnapshot` removes actors that stopped arriving, but only `a.Remote` ones (`net_client.go:357`). The local fighter is never `Remote`, so nothing can remove it.
- `updateDeath` counts `DeathTicks` up on every world, and calls `w.Kill` only on the authority. On a client the count therefore has no ceiling, and `deathHidden()` reads `(DeathTicks/DeathFlickerSpan)%2`, which keeps alternating.
- The server already builds a **per-peer** snapshot: `sendSnapshots` loops over peers and puts that peer's `p.Cmds.last` in `Ack`. `p.Actor` is right there in the same loop, and is exactly the fighter the peer should be told about. It is `nil` while the player is dead, which is the "none" case.
- Offline, `World.Remove` nils `w.Local` and nothing re-points it; `stepOffline` routes input by `actor.IsPlayer`, which is why control keeps working there while the effects keyed on `w.Local` stop.

## Goals / Non-Goals

**Goals:**

- One place decides whose fighter is whose — the server — and says so often enough that no single lost packet can leave a client wrong about it.
- The client's own fighter obeys the same lifetime rule as every other: it exists while the authority reports it and goes when it stops.
- A flicker that cannot strobe forever no matter what else goes wrong.

**Non-Goals:**

- A spectator view, free camera, or any "you are dead" presentation for the two-second gap. The view holds where it was; what fills that gap is a design question of its own.
- Changing when or where the server respawns a fighter. `Kill`, `updateRespawns` and `spawnPos` are correct and are not touched.
- Predicting the handover. A client is told which fighter is its own; it never decides.

## Decisions

### The snapshot names the fighter, rather than a message announcing it

`Snapshot` gains `Yours uint16` beside `Ack` — the receiving player's current fighter, `0` for none. Both fields are already per-peer, so the server fills it in the loop it already has, from the `p.Actor` it already holds.

The alternative was a reliable `MsgRespawn{ActorID}` at the moment of the handover. It is fewer bytes, but it puts the client's most important piece of state behind a single delivery and introduces an ordering question the current code has never had to answer: a respawn message arriving before or after the snapshot that first carries the new fighter are different situations, and only one of them is easy. Carrying it in the snapshot makes the two arrive together by construction, and makes the state self-healing — a client that misses one snapshot is told again 50 ms later. It also answers a case no message would have covered: a client connecting while its player is between fighters is told `0`, instead of inheriting a stale id from `Welcome`.

Cost is 2 bytes per snapshot, not per fighter: the header goes 10 → 12 bytes, so a 16-fighter snapshot lands at 716 bytes against the 1200 the test asserts.

`Welcome` keeps carrying `ActorID`, so the first fighter still arrives the way it does today and the join path is untouched; `Yours` simply becomes the thing that keeps it true afterwards.

### Adopting a new id is a removal plus a fresh reconcile

`onSnapshot` sets `c.ActorID` from `s.Yours` before it looks for the record to reconcile. When the id it is told differs from the one it holds, the fighter it was predicting is no longer its own, and `c.W.Local` is dropped — `World.Remove` already nils `w.Local` for exactly this, so the next `reconcile` takes its existing `local == nil` branch and builds the new fighter from the record. That branch is already written and already tested; the fix reaches it rather than adding a second way to swap a fighter in place.

Applying the new record onto the old actor object would have been fewer lines, but the old actor is a corpse mid-death-animation, and its `DeathTicks`, `Dying`, `Attacker` and rewind history are not in `ActorRec` — so it would carry state the new fighter never had. The spec's *Nothing carries over* is easier to keep by not keeping the object.

Being told `0` is the same path without the rebuild: drop the local fighter, predict nothing, and wait.

### The local fighter is removed by the same rule as every other

`applySnapshot`'s removal loop drops the `a.Remote` condition and instead keeps any actor the snapshot still names. The local fighter is then removed when the server reaps it, which is what makes the corpse disappear on time and the player have no fighter during the gap.

This is safe for the one case the `Remote` test was protecting: the local fighter is absent from `applyInterp`'s loop by the `id == c.ActorID` skip, so `visible` must be built from the snapshot's own ids rather than from what interpolation drew, or the client would delete its own fighter every frame.

### The flicker gets a ceiling in `updateDeath`

`updateDeath` stops incrementing once `DeathTicks` reaches `DeathFlickerTicks`, and `deathHidden` reports hidden from that point rather than alternating. On the authority nothing changes — the fighter is killed on the tick it reaches the limit. On a client it means a dead fighter held for any reason ends up hidden and still, which is the failure that is least bad to look at.

This is belt-and-braces: with the two fixes above, a client stops holding a dead fighter. It is kept because the strobe is the symptom that made the bug unbearable rather than merely wrong, and one clamped counter closes it for any cause, including ones not yet found.

### Offline hands over where it already routes input

`stepOffline` adopts the player's fighter into `CL.Local` when it has none, in the loop that already asks each actor whether it `IsPlayer`. That is the offline equivalent of `Yours`: the same question answered by the only authority an offline match has.

## Risks / Trade-offs

- **Dropping `Remote` from the removal test deletes the local fighter wrongly** → The test becomes "did this snapshot name it", built from `s.Actors` rather than from the interpolation ids, which never include the local fighter. A test that runs a client through a full match asserting its fighter is never removed while alive covers the regression.
- **Two protocol bumps close together** (`add-actor-energy` is already moving 4 → 5) → They collapse into one if the two changes ship together; if they do not, the handshake rejects mismatches cleanly either way. Nothing persists a snapshot, so there is no stored data to migrate.
- **A client adopts a fighter id the server later contradicts** → It cannot linger: `Yours` is in every snapshot, so the correction is at most one snapshot behind, and the fighter it predicts in the meantime is one the server really did hand it.
- **Reconcile's rebuild path runs more often than it used to** → It runs once per handover, roughly once per death, and it is the path every client already takes on its first snapshot.

## Migration Plan

Nothing outside the running process holds a snapshot or an actor id, so there is no data to migrate and rollback is reverting the change. The wire boundary is the `ProtocolVersion` bump, which the existing handshake already turns into a readable "version mismatch" rather than a malformed packet.
