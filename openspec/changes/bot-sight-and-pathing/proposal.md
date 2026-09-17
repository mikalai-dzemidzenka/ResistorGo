## Why

A bot has no eyes and no feet. `botInput` (`actor.go:900`) takes `a.nearest()` — the closest actor by Manhattan distance anywhere in the world — and walks straight at its X, through walls, across gaps and regardless of which way the bot is facing or what is between them. There is no `Jump` in the command it builds at all.

Two things follow, and both are visible the moment a level has any shape to it. A bot on the far side of a wall, in another room, a screen away and behind the bot's back, is still homing in on the player: there is nothing to hide from and nothing to flank. And a bot that meets the geometry on the way loses. `blockSides` stops it at the wall and `stopRun` drops it out of the run (`actor.go:848`), so it stands there pressing into the tile forever; a gap in the floor it walks off, falls into and — with the target now above it — walks back off again. The level got walls, ledges, climbable edges, a standing jump and a run-jump, and the one opponent available offline cannot use or avoid any of them.

## What Changes

- **An attack tells a bot where its attacker is.** A blow is the one thing that reaches a fighter from outside its sight. Whether it lands or is blocked, the attacker's position becomes the spot the bot last saw someone at, so a hit from behind turns the bot round and sends it to look. It learns a position and nothing else — the attacker is not made visible by hitting it.
- **A bot only reacts to what it can see.** Sight is a cone: a range, a half-angle around the direction the bot faces, and an unbroken line to the target. A solid collision rectangle between the two fighters breaks the line. A target outside the cone, past the range, or behind a wall is not seen, and a bot that sees nothing does not approach or swing at anything.
- **A bot that loses sight searches, then gives up.** The last position a target was seen at is remembered. Sight lost → the bot walks to that spot, waits there a moment looking around, and then goes back to standing and drifting its guard as it does today. Regaining sight at any point during the search cancels it.
- **A bot steers around the level reactively.** Each tick it looks at the solids immediately ahead of and under itself and answers one question — where do I step next:
  - a wall ahead it can clear → jump (run-jump when it is running and can pay for it, standing jump otherwise); the existing climb already carries it over the ledge it grabs
  - a gap ahead with ground on the far side it can reach → jump it
  - a gap ahead it cannot clear → stop at the edge rather than walk in
  - a wall ahead too tall to clear → stop, and treat the route as blocked
  - target above or below on another ledge → steer toward the near edge of the platform that gets it there
- **Nothing about combat changes.** The approach distance, the swing odds, the guard wander and the hit-adaptation from `bot-adapts-guard-after-hit` all keep working exactly as they do — they are simply gated behind seeing a target now, instead of firing at the world's nearest actor.
- **A human is unaffected.** Sight and stepping are read only inside `botInput`, which runs for `!IsPlayer` fighters on an authoritative world only (`game.go:267`, `net_server.go:306`).
- Not breaking. No wire change, no protocol version bump, no new replicated field: the bot's sight and search live on the server/offline side of the simulation and reach a client only as the movement and attacks they already produce.

### Scope

Reactive steering, not a navigation graph. The bot reads the solids in front of its own feet each tick; it does not plan a route, and a level shaped so that reaching the target needs a sequence of jumps chosen ahead of time will defeat it. That is deliberate — see Non-goals. The decision is shaped as a single "next step" answer so a planned route could later supply that answer without `botInput` changing.

## Capabilities

### New Capabilities

- `bots`: what a computer-controlled fighter does on its own. The capability is opened by the pending `bot-adapts-guard-after-hit` change and is not yet in `openspec/specs/`; this change adds sight, search and movement requirements to it. Both changes add to the same spec file and neither touches the other's requirements.

### Modified Capabilities

None. Approach, swinging, blocking, energy and structure keep the requirements they have; what changes is when the bot is looking for someone to do them to.

## Impact

- **Code**
  - `actor.go`: `botInput` reworked from "walk at the nearest actor" into see → decide → step. `nearest()` gains a sighted variant, or is replaced by one. New tunables beside the combat constants: sight range, sight half-angle, search patience.
  - `actor.go`, `checkCollision` — **added during implementation, agreed with the author**: one comparison in the ground test made inclusive. Both edges were exclusive, so a fighter whose collision-box centre landed exactly on a tile seam was claimed by neither tile, left the ground for that tick and was put into the fall — which cancelled any jump winding up, making the standing jump unusable to a bot walking at a wall. It hit the player too. No test regressed.
  - New per-bot state on `Actor` (last seen position, ticks since sight, whether a search is running). Local to the authority, not in the snapshot.
  - A small line-of-sight and ground/wall probe against `LE.solids`, likely its own file rather than more of `actor.go`.
  - **Added during implementation**: `level_bot_test.json`, a test course, and a `LevelPath` var behind a `-level` flag to reach it. `level_1.json` holds four collision rectangles and no floor, so none of this can be seen in the game as it stands. The flag routes loading *and* saving, so a level opened this way cannot be saved over `level_1.json`. Not embedded: the release binary and its level fingerprint are unchanged.
- **Tests**: sight — in cone/out of cone/out of range/behind a wall/behind the bot; search — walks to last seen, gives up after the patience runs out, cancels on regaining sight; stepping — jumps a clearable wall, stops at an unclearable one, jumps a crossable gap, stops at the edge of an uncrossable one. All of these run headless against a hand-built `LE.solids`, the way `floor()` in the existing tests already does.
- **Not affected**: the wire format and `ProtocolVersion`, prediction and reconciliation, lag compensation, spectators, the level editor, and `level_1.json` itself — which is left exactly as it was.

## Non-goals

- **A navigation graph and A\*.** Building a node graph from `LE.solids` with walk/jump/climb/fall links, keeping it in step with every editor edit, and pathing over it is a much larger change with its own file, its own rebuild hook and its own tests. Reactive steering gets a bot out of the two failure modes that are actually visible today.
- **Patrolling.** A bot that sees nothing stands and drifts its guard, as it does now. Walking a route looking for someone is a behaviour of its own and would need somewhere to walk to.
- **Hearing.** A bot is not told about anything by the noise of it. *(Being alerted by being hit started here too, and was moved into scope after a play session: a bot that can be hit from behind all day without noticing is worse than the one this replaces. An attack that reaches a bot now tells it where the attacker stood — see the spec's "An attack tells a bot where its attacker is". It learns a position and nothing else, so sight still decides everything else.)*
- **Darkness.** The level has shadow-casting lights and lit areas, and "a fighter standing in the dark is not seen" is a real idea. It would tie bot AI to the lighting and make the simulation answer a question only the shader answers today. Walls only, for now.
- **Bots cooperating, or noticing each other.** Sight is evaluated against every other fighter alike; whether two bots should gang up or ignore one another is a separate question.
- **Tuning the numbers.** The range, the angle and the patience are starting points to be turned by eye in play.
