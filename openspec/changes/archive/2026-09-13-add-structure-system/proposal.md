## Why

Combat has no arc. Every landed hit is identical — the defender is locked into hit-stun, knocked back a fixed distance, and then nothing accumulates. There is no way to lose a fight and no reason to keep pressure on rather than trade single hits, because the tenth hit costs exactly what the first one did. A structure pool, in the manner of Sekiro's posture or Sifu's structure, turns a sequence of hits into a story: pressure builds, the guard buckles, and eventually it breaks.

## What Changes

- **Structure pool on every fighter.** A fighter has 12 hit-units of structure. It is the fighter's whole health model — there is no separate hit-point pool.
- **BREAKING** A landed hit no longer stuns on its own. It costs 1 structure and produces a flinch — the hit flash, the hit-stop freeze and a small knockback, with the hit animation playing — but the defender keeps control and can act through it.
- **Every third landed hit stuns.** The third hit of a run puts the defender in the hit-stun state it has today and throws them back noticeably further than a flinch does. The counter resets when the stun lands.
- **Blocks feed the pool.** A block that holds through the swing's hit frame restores 0.5 structure, capped at full. Block-all (guard 3) is the exception: it holds any height but still costs 0.4 structure, less than the 1 a clean hit costs.
- **Structure recovers on its own.** After 1.5 seconds without losing any, structure refills at 2 per second. Sustained pressure breaks a guard; trading one hit a minute never will.
- **Structure break kills.** At zero, the fighter enters a new Death state: `death_fall.png` plays once, then the body flickers for about 0.6 seconds and is removed from the world.
- **Respawn.** Two seconds after removal, the fighter returns at a free spawn position with full structure — a bot as a new bot, a player as a new fighter for that player.
- **One hit per swing, made explicit.** A swing can land on a given fighter at most once. Today that falls out of the defender changing state on hit; with flinches the defender no longer does, so the swing itself has to remember that it connected.
- **Networked.** Structure and the Death state travel in the snapshot, and breaking, dying and respawning are authoritative-server decisions like hits are. The protocol version goes up.

### Assumptions

- Numbers are first values to tune, not balance decisions: 12 structure, 1 per hit, 0.4 per block-all, 0.5 restored per block, 2 per second of regeneration after 1.5 seconds, a 0.6 second flicker, a 2 second respawn delay.
- Block-all does not advance the every-third-hit counter — it is not a landed hit — but structure it drains can still reach zero and kill, exactly as a hit can.
- The stun counter counts hits that land on that fighter, and is reset by a stun, by death, and when structure starts regenerating. It is not a global combo counter.
- A flinch does not interrupt what the defender was doing: a fighter hit mid-swing keeps swinging, so trades are possible.
- Bots do not play around structure: `botInput` is unchanged, and a bot neither guards its own structure nor targets the opponent's.

## Capabilities

### New Capabilities
- `structure`: the structure pool and what spends and restores it, the flinch, the stun on every third hit, structure break, the Death state and its flicker, and respawning.

### Modified Capabilities
- `multiplayer/replication`: *Snapshots* — the per-fighter state a snapshot carries now includes structure, so a client can display and predict it.
- `multiplayer/session`: *Player slots and spawning* — a player's fighter can now leave the world on death and return on respawn, so "every accepted player has a fighter" needs the gap between the two spelled out.

## Impact

- **Code**
  - `actor.go`: a `Structure` field and the hit counter; `enterAttackedState` splits into a flinch and a stun; a new `StateDeath` with its update and entry; the per-swing hit flag; regeneration in the fixed step; block credit where a block is confirmed in `updateBlock`; the flicker in `Draw`.
  - `world.go`: removing a dead fighter and the pending-respawn list, on the authoritative world only.
  - `net_protocol.go`: structure in the actor record (record size and protocol version both change), `death_fall.png` registered in the animation table so the wire can name it.
  - `net_server.go`: death, removal and respawn are server decisions; a player whose fighter died gets a new one.
  - `resources.go`: `death_fall.png` loaded and added to the animation table — it is the only unregistered file in `resources/threefour/` and a missing texture panics on draw.
- **Compatibility**: the protocol version rises, so older builds are rejected with the existing "version mismatch". No saved data is affected.
- **Not affected**: the level editor, the camera, time manipulation, prediction and lag compensation beyond the added field.

## Non-goals

- A visible structure bar or any HUD for it. The debug overlay can show the number; a real meter is its own change.
- Bot AI that understands structure — no guarding when low, no pressing an opponent who is nearly broken.
- Round or match structure: scores, win conditions, or an end to the fight. Respawning keeps the fight going forever, exactly as it does now.
- Different structure costs per attack height or per attack in a chain, and any deathblow or execution mechanic.
