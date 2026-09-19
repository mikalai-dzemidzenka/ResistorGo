## Context

See proposal.md — Why. What matters for the approach:

- `botInput` (`actor.go:900`) is a pure function from an actor to one `Input`. It runs on the authority only: `stepOffline` calls it for every `!IsPlayer` actor (`game.go:267`), `Server.Step` for every bot it owns (`net_server.go:306`). A client never calls it. Everything here can therefore live behind that one function without touching prediction, reconciliation or the wire.
- The level's collision is a flat slice, `LE.solids` (`level_editor.go:712`): world-space `rl.Rectangle`s, rebuilt by `buildIndex` on every editor edit and on load. There is no grid, no tile lookup and no spatial index — `checkCollision`, `headBlocked` and `blockSides` each walk the whole slice every tick per actor. Anything this change reads about the level reads that same slice the same way.
- Orientation is `int8`, `±1` only. There is no aim vector, so a "cone" is a wedge either side of the X axis in the facing direction.
- `Actor` is a plain struct; the snapshot is written field by field in `net_protocol.go` (`writeActor`/`readActor`, `actorSnap`). Adding fields to `Actor` does not change a byte on the wire unless someone writes them there.
- A headless test harness for level geometry already exists: `tiles(corners...)` and `floor()` in `actor_test.go` build an `LE` out of 48×48 solids with no window, and `bot_guard_test.go` is the precedent for bot-specific tests.
- The physics a bot has to reason about, all constants in `actor.go`: gravity 1500 px/s², standing jump take-off `MaxJumpSpeed * dt` = 250 px/s up, run-jump `MaxRunJumpSpeed * dt` = 437.5 px/s up, `MaxSpeed` 320 px/s, `StepHeight` 4 px, a fighter 48 px tall with a 14×42 collision box.

## Goals / Non-Goals

**Goals:**

- Put sight, memory and the step decision behind one seam, so `botInput` reads as *see → decide → step* and a route planner could later replace the "decide" half alone.
- Derive the bot's idea of what it can jump from the movement constants, not from typed-in pixel counts, so retuning gravity or the jump cannot leave the bot believing in a jump it no longer has.
- Keep every new read of the level to the same `LE.solids` walk the rest of the actor code already does. No index, no cache, no rebuild hook.
- Stay testable headless: every requirement in specs/bots/spec.md should be reachable from a `tiles(...)` layout and a hand-placed pair of actors.

**Non-Goals:**

- Determinism across client and server. None of this is predicted, so there is nothing for the two to agree about. `botInput` already uses unseeded `math/rand` and continues to.
- Sharing the probes with the player. A human does not need to know whether a gap is crossable; the probes are the bot's, and `checkCollision`/`blockSides` stay exactly as they are.
- Making the bot good at fighting. The approach distance, swing odds and guard behaviour are untouched.

## Decisions

### A new `bot.go`, with `botInput` moved into it

`actor.go` is 1,594 lines and already holds the state machine, physics, collision, animation and combat. Sight, memory and terrain probes are a fourth concern that does not need any of `Actor`'s internals beyond position, facing and state.

`botInput` and `nearest` move to a new `bot.go`; the new code joins them there, and `bot_test.go` sits beside `bot_guard_test.go`. The package is flat, so nothing about the move changes what can call what.

*Alternative:* leave it in `actor.go`. Rejected — the change roughly doubles the bot code, and the sight/probe helpers would then sit between combat functions they have nothing to do with.

### Bot memory as fields on `Actor`, not a side table

The bot needs to carry three things between ticks: the last position it saw a target at, how long it has been searching, and whether it is searching at all. A side map keyed by actor would have to be cleaned up on death and respawn, which `World.Remove` does not know about.

Fields on `Actor` follow the pattern the struct already uses for per-fighter bookkeeping (`HitRun`, `EnergyCalm`, `DenyUntil`). They are born zero on spawn, which is exactly "not searching", so a respawned bot starts with a clean mind for free. They are never written to a snapshot, so the wire is unchanged; the spec's *Sight and movement decisions belong to the authority* is satisfied by simply not adding them to `writeActor`.

Grouping them as one small `bot` struct field on `Actor` keeps the three together and makes it obvious at a glance that they are not replicated — the comment sits on one field instead of three.

*Rewind:* `ActorRewindFrame` (`rewind.go:44`) records position, speed, orientation and animation only, so rewinding time does not restore bot memory and never has restored any bot state. A bot that had a target in mind before a rewind keeps that mind after it. This is consistent with everything else rewind leaves alone and is not worth fixing here.

### Sight: a slope test, a squared range and a segment/AABB sweep

Three cheap tests in sequence, cheapest first, each rejecting before the next runs:

1. **Facing.** `sign(target.x - self.x) == Orientation`. One comparison, and it drops roughly half the candidates.
2. **Range and angle, without trigonometry.** Take `dx`, `dy` between the two fighters' collision-box centres. In range when `dx*dx + dy*dy <= SightRange²`. Inside the cone when `abs(dy) <= abs(dx) * SightSlope`, where `SightSlope` is the tangent of the half-angle written directly as a constant. A slope comparison is the same predicate as an angle comparison for a wedge about an axis, without `atan2` and without a normalise; it also keeps the tunable something that can be understood as "how far up and down it sees per unit of distance ahead".
3. **A clear line.** The segment between the two centres against each rect in `LE.solids`, by the slab method — clip the segment's parameter interval against the rect's X slabs and then its Y slabs; anything left means the segment crosses the rect. Branch-free enough, no allocations, and it handles the vertical and horizontal segments the level will produce without a special case as long as the division by a zero component is guarded.

Eye point is the centre of the collision box rather than `Pos`: `Pos` is the sprite's top-left, so two fighters standing on the same floor would sight each other along a line through their foreheads, and a waist-high ledge between them would not break it when it obviously should.

*Alternatives:* a true angle via `atan2` — more expensive per tick, and the constant becomes a number in radians that reads as nothing. Raycasting against the level by marching fixed steps — cheaper to write, but it steps over thin solids, and the level has plenty of thin collision rects.

*Cost:* one slab sweep over `LE.solids` per bot per visible candidate per tick. `level_1.json` currently resolves to 4 collision rects, and the same slice is already walked three or more times per actor per tick by collision. Even a level two orders of magnitude denser is well inside the budget, and if it ever is not, the answer is a spatial index for collision and sight together, not a cache here.

### The step decision as one function returning an intent

```
type step struct {
    dir  int8 // -1, 0, +1: which way to walk
    jump bool
    stop bool // nothing to be done; come to rest
}
```

`botStep(a *Actor, goal rl.Vector2) step` is the whole of "where do I put my feet". `botInput` calls it once, with the goal being either the seen target's position or the remembered one, and turns the result into `Input.Left`/`Right`/`Jump` and the `face()` call. Attacking stays in `botInput`, where it is now, gated on a seen target in reach.

That is the seam the proposal promises: a route planner would give `botStep` a waypoint instead of the target's own position and the rest of `botInput` would not know.

**What `botStep` reads, in order:**

- **Wall ahead** — walk the solids, keep those whose facing-side face lies within a lookahead window in front of the collision box and which pass a variant of `isWall`, and take the highest wall top among them. Reusing `isWall`'s definition matters: a rise is a step if and only if `blockSides` will let the fighter walk onto it, so the bot's idea of "wall" and the collision code's cannot drift apart.
- **Gap ahead** — probe a column one lookahead in front and ask for the highest solid top within a drop window under the feet. No such top means a gap; the far side is found by probing outward until ground appears within the jump's horizontal reach.
- The rest is the ladder the spec lays out — step, clearable wall, unclearable wall, crossable gap, uncrossable gap — with the vertical difference to the goal choosing between "keep walking" and "look for the way up/down".

*Alternative:* fold the decision into `botInput` as a run of `if`s. Rejected — a separate function is the seam, and it is the thing the tests can drive directly with a hand-built `tiles(...)` layout and no `Input` round-trip.

### Jump reach derived from the movement constants

A projectile leaving the ground at `v` under gravity `g` rises `v²/2g` and is airborne for `2v/g`. With the constants as they stand:

| | take-off | rise | airborne | horizontal at `MaxSpeed` |
|---|---|---|---|---|
| standing jump | 250 px/s | ~20.8 px | 0.33 s | ~107 px |
| run-jump | 437.5 px/s | ~63.8 px | 0.58 s | ~187 px |

These are written as constants derived from `Gravity`, `MaxJumpSpeed`, `MaxRunJumpSpeed`, `MaxSpeed` and `FixedTimeStep`, in the style the energy and structure tunables already use, with a safety margin factor below 1 applied to the horizontal reach so the bot does not commit to the exact edge of what is possible. They are the bot's *belief* about its jump, used only to decide; the jump itself is the same `enterJumpStartState`/`enterRunJumpState` any fighter takes, and the physics has the last word.

Deriving rather than typing them is the point: the run-jump lift was retuned once already (`RunJumpLift`), and a bot with a hardcoded 187 would have quietly started missing jumps.

*Note on the run-jump:* it costs `EnergyRunJump` (a third of the pool) and `enterRunJumpState` refuses when the bot cannot pay. `botStep` asks `canRunJump()` before choosing a jump that needs the longer reach, which is what makes the spec's *A jump it cannot afford is a jump it cannot make* true rather than accidental — without the check the bot would commit to the gap, be refused mid-run and walk off the edge.

### Search: a countdown, not a state machine

Losing sight sets the remembered position and a patience counter; the counter runs down each tick the bot does not see anyone. Reaching the remembered position — within a small tolerance on X, since Y is whatever ground the bot ends on — switches the counter to a shorter linger and the bot stands still for it. Either counter hitting zero clears the memory and the bot is at rest.

Seeing anyone clears the memory outright before the counter is read, so *Sight cancels the search* is one early return rather than a transition.

One counter and one position, no enum. The bot has exactly two modes — it has somewhere to be, or it does not — and an enum would be three names for two states.

*Tunables:* `SightSearchTicks` (patience, order of 3 s) and `SightLingerTicks` (the look-around at the end, order of 0.75 s), both feel numbers to be turned by eye, stated in ticks beside the other tick-count tunables.

### Turning is a consequence of moving, not a thing of its own

`face()` is currently called on every tick from `botInput` toward the nearest actor anywhere. It becomes: turn toward a seen target, turn the way `botStep` says to walk, and otherwise do not turn at all. That is what makes the cone mean anything — a bot that turned to face an unseen fighter would see it on the next tick.

The consequence, accepted: a bot that has never seen anyone faces one way forever, and something approaching from directly behind is never noticed. Patrolling and being alerted by a hit are both non-goals; if the bot facing a wall for the whole match turns out to read as broken in play, being alerted by a hit is the cheapest fix and is its own change.

### Two things the level's collision turned out to need

Both were found by building the thing, and neither was foreseen above.

**A tile seam dropped a fighter off the ground.** The ground test in `checkCollision` had both edges exclusive, so a fighter whose collision-box centre landed exactly on a tile boundary was claimed by neither of the two tiles: `OnGround` went false for that one tick, gravity applied, and `updateFalling` put it into `StateFall`. Walking an unbroken floor, that cancelled any jump still winding up — and the standing jump winds up for 42 ticks, so a bot walking at a wall essentially never got off the ground. The fix, agreed with the author, is one comparison made inclusive, so a boundary belongs to exactly one tile. It is core collision shared with the player, which is why it was asked about rather than assumed; no existing test changed behaviour.

**A bot cannot stop as quickly as it can decide.** From top speed, friction alone carries a fighter about 205 px — more than twice the distance the probes were reading ahead. A bot that decided at 96 px to stop at an uncrossable gap slid straight into it. So the hole probe looks ahead by its braking distance as well, and the bot stops pressing forward while it still has room to coast to a halt.

That is not quite enough on its own: a threshold that moves with the speed has an equilibrium in it, and the bot found it, pressing and releasing on alternate ticks while going nowhere — at that speed the step it takes is smaller than the pixel `Pos` is truncated to. The stop distance is therefore the braking distance *or* a fixed standoff, whichever is larger. The fixed floor is what makes it settle, and it settles about a stride short of the edge.

### Two more the course found

**Ground a step above the feet is still ground.** The hole probe took ground at or below the feet, so a 4px rise — exactly what `StepHeight` exists to let a fighter walk up — read as a hole, and the bot stopped dead at something it would have strolled over. The probe now allows a step's worth of headroom, which is the same rule `isWall` applies at the other end.

**A gap has to be judged at the speed the bot will leave it at.** Judging by the speed it has while deciding, a bot standing a stride from an edge finds it cannot cross, stops — and so never builds the speed that would have carried it over. It would stand there for ever. The crossing test now predicts the take-off speed from the run-up still available (`v² + 2·Acc·run`, capped at `MaxSpeed`), which is the same shape of reasoning as the braking distance, read the other way round.

## Risks / Trade-offs

- **The bot gets stuck somewhere reactive steering cannot solve** (a route needing two jumps chosen in advance, or a pit whose only way out is backward) → It comes to rest instead of walking into geometry, which is strictly better than today's walk-into-the-wall. The spec says as much rather than promising the bot arrives, and `botStep` is shaped so a planner can be dropped in behind it.
- **A bot oscillates at a decision boundary** — jumps, lands one pixel short, jumps again → This happened, at the edge of an uncrossable gap rather than at a jump, and is fixed by the fixed standoff described above. Watch for the jumping version in play; a short commit timer after a jump is the fallback if it appears.
- **`level_1.json` cannot show any of this off.** It resolves to four collision rects — essentially a flat floor. Almost nothing in the movement half of the spec is reachable in the game as authored today → Verification is headless tests over `tiles(...)` layouts, plus placing a few collision rects in the level by hand to try it in play. Worth saying out loud: "it works" here means the tests, not a play session on level_1.
- **Sight makes the offline bot feel passive.** Gated behind a cone, the bot that used to charge on sight now sometimes stands around → That is the feature, but the range and the angle decide whether it reads as alert or as blind. Both are single constants, deliberately.
- **Cost grows with the level.** Sight is another full walk of `LE.solids` per bot per tick, on top of the several collision already does → Linear in a slice that is currently four long; at the size where it matters, collision is the larger problem and an index serves both.
- **`math/rand` stays unseeded.** The guard wander and swing odds keep drawing from the global source, so a test cannot pin a bot's exact behaviour → The existing bot test (`bot_guard_test.go`) already handles this by asserting the distribution over many trials. The new tests are written to avoid the random paths entirely: sight, probes and `botStep` are deterministic and can be called directly.

## Migration Plan

Not applicable. No data, no protocol and no saved level changes; the feature is live the moment the code is in, for offline games and any server running the new build. A server and a client of different builds interoperate as before, because nothing here reaches the wire.

Rollback is reverting the commit.
