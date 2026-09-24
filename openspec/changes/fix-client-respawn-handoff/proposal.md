## Why

A networked player who dies never comes back. Their fighter blinks on their screen forever and they control nothing, from the moment they die until they quit.

Three things compose into it. A client only removes actors marked `Remote` when they stop appearing in a snapshot, and a client's own fighter is not one of them — so the corpse stays in the client's world after the server has reaped it. `updateDeath` keeps counting `DeathTicks` up on a client, and the `Kill` that would stop it is gated on being the authority, so the count runs away; `deathHidden()` is `(DeathTicks/4)%2 == 1`, which blinks the sprite at about 7.5 Hz for as long as the match lasts. Then the server does respawn the player two seconds later, but a client learns its fighter's id only from the `Welcome` it received when it joined, and nothing tells it about the new one — so `reconcile` never matches the new fighter, `W.Local` stays the corpse, and control never returns.

Measured against a real server: `DeathTicks` reached 360 and was still climbing six seconds after death, while the server had been holding a fresh fighter for the player since the three-second mark.

This is already specified. The `structure` capability's *Respawning* requirement says a player whose fighter died "SHALL be given the new fighter and control it", and its scenario spells out that "the player has no fighter while it is dead, and 2 seconds after it disappears they control a new fighter". The behaviour above meets neither half. `TestNetDeathAndRespawn` passes over it because it asserts the client can *see* the respawned fighter by id — which it can, as a remote actor it does not control.

Single player has the quiet half of the same bug: `World.Remove` nils `w.Local`, and nothing offline re-points it at the respawned fighter, so after dying the player keeps control (input is routed by `IsPlayer`) while `CL.Local` stays nil and the attack zoom and hit shake silently stop.

## What Changes

- **The snapshot says which fighter is the receiving player's.** Its header already carries a per-player `Ack`; it gains the id of that player's current fighter beside it, `0` while they have none. A client adopts what it is told, so the handoff repairs itself on the next snapshot after any loss, and a client that connects while dead is told the truth too.
- **A client drops its own fighter when the server stops sending it**, as it already does for every other fighter. The corpse plays its fall and its flicker while the server is still reporting it, then goes, and the player has no fighter for the rest of the wait.
- **The death flicker cannot run away.** A fighter past the end of its flicker span stops blinking and stays hidden rather than counting up forever, so a client that is waiting to be told something can never present a strobing sprite.
- **Offline hands the fighter over too**, so `CL.Local` follows the respawned fighter instead of staying nil.
- **BREAKING** (wire): the snapshot header grows by two bytes, so `ProtocolVersion` moves again. Client and server must be built together, which the existing version handshake already enforces.

## Capabilities

### Modified Capabilities
- `multiplayer/replication`: the snapshot carries the receiving player's current fighter id, and the packet budget holds with it.
- `multiplayer/session`: a player's fighter changing during a session — dying, then being given a new one — becomes part of what the session guarantees, rather than something only the join handshake covers.
- `structure`: the *Respawning* requirement's existing promise is unchanged, but what a client shows and controls between the death and the respawn is currently undefined; it gains the scenarios that pin it down.

## Impact

- `net_protocol.go`: `Snapshot` gains a `Yours uint16`; `encodeSnapshot`/`decodeSnapshot` carry it; the snapshot header goes 10 → 12 bytes; `ProtocolVersion` 9 → 10.
- `net_server.go`: `sendSnapshots` fills it per peer from `p.Actor`.
- `net_client.go`: `onSnapshot` adopts `Yours` into `c.ActorID`; `applySnapshot` removes the local fighter when it is no longer in the snapshot; `reconcile` handles being given a different fighter than the one it holds.
- `actor.go`: `updateDeath` stops the count at the end of the flicker instead of counting on.
- `game.go`: `stepOffline` points `CL.Local` at the player's fighter when it has none.
- Tests: `net_integration_test.go` gains a death-to-respawn-to-control case; `net_protocol_test.go` covers the new field and the packet size; the existing `TestNetDeathAndRespawn` is tightened to assert control, not just visibility.
