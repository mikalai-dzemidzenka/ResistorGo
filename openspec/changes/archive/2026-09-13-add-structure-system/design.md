## Context

See proposal.md — Why. What shapes the approach is how combat resolves today:

- Hits are decided in `Actor.updateAttack`, on the authoritative world only. It walks `a.W.Actors`, tests the attack box against `actor.collisionRect()`, and skips anyone already in `StateAttacked` or `StateBlock`. A guard at the attacker's height, or block-all, calls `enterBlockState`; otherwise, once the swing passes its hit frame (`a.AnimCurrentFrame > attackFrames[a.Anim]`), it calls `enterAttackedState`.
- A block is provisional. `updateBlock` re-tests at the swing's hit frame and converts the block into `enterAttackedState` if the guard no longer matches and is not block-all. So the hit frame is the one moment at which both outcomes are final — before it, a "block" may yet become a hit.
- `enterAttackedState` does five things at once: white flash, `EvHit`, `FreezeTicks = HitStopTicks` on *both* fighters, `PunchSpeed` knockback, the hit animation, and `State = StateAttacked`.
- Nothing stops a swing landing twice. It does not need to today: the defender changes state on the first landing, and the `State != StateAttacked && State != StateBlock` test then skips it for the rest of the swing.
- `ActorState` is a `uint8` const block and is replicated as its numeric value. Animations are replicated as a `uint8` index into `animTable` in `resources.go`; `death_fall.png` is the one file in `resources/threefour/` that is not in it, and a texture that is not loaded panics on draw.
- `World.Remove` already exists for disconnects and clears dangling `Attacker` pointers and lag-compensation history. Nothing respawns anything.
- The actor record on the wire is `actorRecSize = 41` bytes with a `uint16` flag field.

## Goals / Non-Goals

**Goals:**

- One place where an attack's outcome is priced, so hit, block and block-all cannot drift apart.
- A flinch that is genuinely not a lockout, without the animation being stamped over by the idle logic on the next tick.
- Death that clients can see — animation and flicker — before the fighter leaves the world.
- Keep hit authority where it is: the server decides, the client displays.

**Non-Goals:**

- Any HUD for structure (proposal — Non-goals). The debug overlay is the only readout.
- Reworking the block state machine. Blocks keep their current shape; this change only attaches structure to the moment they resolve.
- Making time manipulation replay a death faithfully.

## Decisions

### Structure is a float in the simulation, tenths on the wire

`Structure float32`, 0..12. Hits cost 1, block-all 0.4, blocks restore 0.5 — values that want a fraction. On the wire it is one `uint8` of tenths (0..120 covers the full range), taking the actor record from 41 to 42 bytes and meeting the spec's 0.1 precision.

*Alternative — integers scaled by ten everywhere:* rejected. It spreads a scale factor through the combat code for no gain; quantizing once, at the boundary that needs it, keeps the simulation readable next to `Pos` and `Speed`.

### Price the swing where its outcome becomes final

Two entry points own all structure accounting: `takeHit(attacker)` and `confirmBlock(attacker)`. Every path that resolves a swing — the hit branch in `updateAttack`, the block branch, and the block-to-hit conversion in `updateBlock` — goes through one of them. Nothing else touches `Structure` except regeneration.

This matters because a block is entered early (on box overlap) and only confirmed at the hit frame. Crediting structure at `enterBlockState` would pay the defender for a guard they then dropped.

### A swing remembers that it connected

The attacker carries a `SwingHit` flag, set by `takeHit` or `confirmBlock` and cleared where a swing begins — `enterAttackState` and the chained-swing branch of `updateAttack`. A swing whose flag is set resolves nothing further.

This replaces the implicit guard described in Context, which stops working the moment a flinch leaves the defender in its previous state. The existing `State != StateAttacked && State != StateBlock` test stays, as the separate question of whether a fighter can be engaged at all, and gains `StateDeath`.

*Alternative — record the attacker's `AttackCount` on the defender:* rejected. It needs per-pair bookkeeping to be correct with more than two fighters, and the flag says exactly what is meant.

### Flinch and stun are one entry split in two

`enterAttackedState` splits. The shared part — flash, `EvHit`, hit-stop on both fighters, knockback, hit animation — happens on every landed hit. Only the third adds the lockout (`State = StateAttacked`) and the larger knockback.

The wrinkle: `StateIdle` in `Actor.Update` re-selects an animation every tick from speed and stance, so a flinch animation set while idle is gone on the next tick. A short `FlinchTicks` counter suppresses that re-selection while it runs. Starting an action — attack, jump, block — sets its own animation and ends the flinch naturally, which is what "keeps control" should feel like.

*Alternative — give the flinch its own state:* rejected. A state is a lockout by construction, and the lockout is the thing being removed.

### Death is a state; removal waits for the flicker

`StateDeath` is appended to the `ActorState` consts, so every existing value keeps its number on the wire. The fighter stays in the world through the death animation and the flicker, then the authority calls `World.Remove`. Only then does the respawn timer start.

Keeping the actor alive is what lets clients see the death at all: they render what the snapshots contain, and a fighter removed at the moment of the break would simply vanish.

*Alternative — remove immediately and play the death as a detached effect:* rejected. There is no effect system, and building one to avoid a state is the larger change.

### The flicker counts ticks, not seconds

The flicker skips drawing on alternating spans of about four ticks. Ticks rather than wall-clock, because animation timing in this project already runs off `PhysicsTime`, and because time manipulation replays history by frame number — anything keyed to real time would flicker at the wrong rate while scrubbing.

### Respawning belongs to the world, not the corpse

A small pending-respawn list on `World` — due tick, and who it is for — checked once per fixed step on the authoritative world. `net_server.go` attaches the newly spawned actor to the waiting `netPlayer`; offline it just spawns a bot. Spawn placement reuses the existing free-position search rather than a second copy of it.

*Alternative — a timer on the dead actor:* rejected. The actor has to be gone from the world during the gap, which is the whole point of the removal.

### `death_fall.png` gets registered before anything draws it

The file exists but is in neither `LoadTextures` nor `animTable`. Both are required: an unloaded texture panics on draw, and the wire names animations by their index in the table. Appending to the end of the table leaves every existing index alone — worth being deliberate about, since the protocol bump here is for the record size, not for a reshuffled animation table.

### Protocol version 3

The record grows by a byte, so the version goes up and old builds are rejected by the existing "version mismatch" path. The handshake requirement does not name a version number, so only `multiplayer/replication` needs a spec delta for the field itself.

## Risks / Trade-offs

- **Structure is the entire health model and has no meter.** In play, nothing shows how close a fighter is to breaking until they die. → The debug overlay shows the number; a real meter is the expected immediate follow-up and is called out as a non-goal, not an oversight.
- **Hit-stop now fires on every landed hit, not just on stuns.** `HitStopTicks = 6` freezes both fighters, so a fast exchange could feel choppy where it used to feel continuous. → One constant; judge it in play, and consider a shorter freeze for flinches than for stuns if it drags.
- **A client sees its own death a round-trip late,** because structure and death are authority-only. → Consistent with how hits already behave; prediction of a death would have to be rolled back when wrong, which is worse.
- **Respawn can collide with a fighter standing on the spawn point.** The existing placement search handles the join case and is the only implementation of "free position"; respawn must reuse it rather than pick a fixed point. → Covered by a requirement and a task.
- **Time manipulation does not restore structure or resurrect anyone.** A recorded frame holds position, speed, orientation and animation only, so scrubbing back across a death will not bring the fighter back. → Pre-existing limitation of the recording, now more visible; widening the recorded frame is its own change.
- **Bots do not understand structure.** They will walk into a fourth stun and die without ever guarding. → Accepted and stated in the proposal; bot work is a separate change.

## Migration Plan

Nothing is persisted, so there is no data migration. The protocol version bump means client and server builds must move together; mismatched builds already get a clear rejection. Reverting is a single revert — no state outlives a session.

## Open Questions

- Whether a death should clear the killer's swing, so a chained swing cannot carry over into the respawned fighter. Deferrable: the respawn is 2 seconds later and no swing lasts that long, so it only matters if the delay is tuned down.
- Whether the debug overlay should show structure for every fighter or only the local one. Deferrable, one draw call either way.
