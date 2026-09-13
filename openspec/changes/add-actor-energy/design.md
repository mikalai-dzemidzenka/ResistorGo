## Context

See proposal.md — Why. What shapes the approach is the code energy has to fit into:

- `Actor` already declares `Energy int32` (actor.go:101). It is written nowhere and read nowhere — a field left for exactly this.
- Structure is the working model for a fighter's pool: a `float32` on `Actor`, spent through `spendStructure`, restored through `restoreStructure`, refilled by `regenStructure` gated on `CalmTicks`, replicated as tenths in a byte, and decided only where `World.Authority` is true.
- A swing starts in two different ways. The first swing goes through `enterAttackState()`, called from the `StateIdle` and `StateRun` arms of `Actor.Update`. A chained swing never calls it: `updateAttack` sets `AttackQueued` while attack is held and, on `AnimFinished`, sets the next attack animation inline. Any cost has to cover both paths.
- Blocks resolve in `confirmBlock(attacker)`, reached from `updateAttack` (the attacker's view) and `updateBlock` (the defender's view). Both call sites are already inside `if a.W.Authority` checks.
- The simulation runs at 60 Hz (`FixedTimeStep = 1.0/60.0`).
- `ActorRec` is a fixed-size record: `actorRecSize = 42`, hand-encoded field by field, with a 16-fighter packet asserted under 1200 bytes by `net_protocol_test.go`.
- The bar over a fighter's head is laid out in whole world pixels by `structureBarLayout`, which already draws a centred row of stun pips above the bar using the `4n-2` width trick to stay on the pixel grid.

## Goals / Non-Goals

**Goals:**

- One pool, spent in one place and granted in one place, so the rule cannot drift between the two ways a swing starts.
- Exact integer arithmetic end to end, so the value on the wire is the value in the simulation with nothing lost to rounding.
- A refusal that is felt on the same tick as the input, under latency, without letting a client invent energy.
- HUD that reuses the bar's existing pixel-grid machinery rather than starting a second HUD system.

**Non-Goals:**

- Per-move costs (a heavy attack costing more than a jab). The 200-point scale leaves room for it; this change gives every swing the same price.
- Energy as a second defeat condition. Running out refuses swings; it never kills, never stuns, and never interacts with structure.
- Spending energy on anything but attacks — no dash cost, no block cost, no jump cost.
- Recording energy in the rewind buffer. `ActorRewindFrame` holds position, speed, orientation and animation only, and the HUD is already hidden while time manipulation shows a recorded frame.

## Decisions

### Integer points, and a regen tick divisor instead of a rate

Energy stays the declared `int32` and every value in the system is a whole number: cap 200, swing 12, block 15. Regeneration is expressed as one point every `EnergyRegenTicks = 4` ticks, which at 60 Hz is the 15 per second the spec requires, rather than as a float rate added per tick.

Structure needed `float32` because a block pays 0.5 and regen adds a fraction each tick; that is also why it replicates as tenths and why a client can only ever show it "to a precision of 0.1 or better". Energy has no fractional value, so making it integer means the byte on the wire is the pool exactly, at every value from 0 to 200 — which is what the replication delta now promises, and which is why the cap cannot go past the 255 a byte holds. Alternative considered: mirror structure with a `float32` and `EnergyRegen * FixedTimeStep` per tick. Rejected — it reintroduces rounding on the wire for no gain, and `55.24999` on the server versus `55` on the client is a difference that can flip an affordability check.

### `enterAttackState` owns the cost and reports refusal

`enterAttackState()` becomes `enterAttackState() bool`: it checks affordability first, and on success spends and enters the state as it does today. Its two call sites become `if a.Input.Attack && a.Updown != 3 && a.enterAttackState() { return }` — refused, control falls through to the rest of the idle/run logic untouched.

The chain in `updateAttack` cannot call it (it continues an attack rather than entering the state), so it gets the same check at its one branch point: where it reads `a.AttackQueued` to decide between the next swing and the back-attack recovery, an unaffordable swing is treated exactly as a released button — `AttackQueued` is cleared, the recovery animation plays, and `EvNoEnergy` is raised. This is what makes "a chain that runs out ends cleanly" fall out rather than needing its own path.

Alternative considered: check affordability at the input sites (`if a.Input.Attack && a.Energy >= EnergySwing`). Rejected — three call sites, each free to forget, and the spec's "no energy is taken by a refused attack" then depends on every one of them agreeing.

### The block restore lives in `confirmBlock`'s matching-guard branch, which is already server-only

`confirmBlock` splits on `a.Updown == 3` — the block-all branch spends structure, the other restores it. Energy is granted in the second branch only, next to `restoreStructure(BlockStructure)`. Block-all therefore restores nothing by construction rather than by a check, which is the rule the proposal asked for.

Both call sites are already behind `w.Authority`, so "a client does not grant itself the block restore" needs no new guard. Regeneration gets the same early return `regenStructure` already uses.

### A client predicts spending and regeneration; the block restore stays the server's

Spending is the one part a client must run locally: the prediction spec's "Immediate local response" requires the local fighter to start and chain attacks on the input tick, and a gate on those attacks is part of that answer. So the cost is taken in the ordinary simulation path with no authority check, which means reconciliation replays it naturally from the snapshot's value — the same mechanism that already keeps predicted position honest.

Regeneration has to be predicted for the same reason, which is not obvious and was got wrong first. The reasoning that failed: restores and regeneration are authority-only, so a client's energy can only be *lower* than the server's, and that is the safe direction. It is not. Lower means the client refuses a swing the server allows — and the server, running the same command, swings anyway, so the attack arrives in the next snapshot as a swing the player never saw start. Measured on the prediction harness (client 9 ticks ahead, snapshots every 3 ticks, attack held down), that happened on every refill cycle: five swings out of five, each with the client one point short of what a swing then cost and the server past it (measured when a swing cost 20; the tuning since has not changed the shape). A pool refilling under a held button is not a corner case; it is what a fighter who has spent their pool does next.

So regeneration runs on the client too, and everything it is decided from is replicated. `Ticks` already is. The missing input was `EnergyCalm` (ticks since energy was last *spent*, distinct from structure's `CalmTicks`, which counts ticks since structure was last *lost*), so the record carries it as a byte — it is capped at `EnergyCalmTicks = 90` and never needs more.

The block restore stays authority-only, which is right rather than merely tolerable: it is granted inside `confirmBlock`, a combat outcome the client is already forbidden to decide, and it moves energy only on a tick where the server has also decided a block that the client was waiting to be told about anyway.

### Energy rides in the record as two plain bytes

`ActorRec` gains `Energy uint8` and `EnergyCalm uint8`; `actorRecSize` goes 42 → 44, putting a 16-fighter snapshot at 714 bytes including its 10-byte header, well under the 1200 the test asserts. `apply` clamps with `min(r.Energy, MaxEnergy)` rather than rejecting the packet, following what `HitCount` already does — a malformed value degrades to a sane one instead of dropping a whole snapshot.

`ProtocolVersion` goes 4 → 5. The record's size and layout change, so an old client and a new server must not agree to talk; the existing version handshake already rejects the mismatch with a readable message.

### Energy is a second bar, not a row of pips

`structureBarLayout` gains an energy track and fill below the structure bar's track, the same width, drained from full as energy is spent — it shows what the fighter has, where the bar above it shows what the fighter has lost.

Pips were tried first, one per affordable swing, and the tuning pass killed them. The row cannot be wider than the 16px bar it sits under, and at the tightest readable pitch (2px pips, 1px gaps) that caps it at five. Five pips was exactly a full pool at the original 20-per-swing, so the HUD silently encoded the cost: making a swing cheaper — which is what play asked for — would have needed the pips to mean something other than one press, or to shrink to 1px and read as noise beside the 2px stun pips above them. A bar has no such ceiling. What a swing costs can now be retuned without touching the HUD at all, which is the property the first design lacked.

The fill is `roundf(StructureBarWidth * Energy / MaxEnergy)`, so its edge lands on a whole world pixel like every other edge in the layout. The colour must be an exact palette entry, like `structureBarFill` and `stunPipColor`: the palette pass maps every pixel to its nearest entry, so an off-palette colour would shift under it.

Visibility is its own predicate (`showEnergyBar`), keyed on `Energy < MaxEnergy` plus the Death-state and `TimeShowsHistory()` exclusions — deliberately not reusing `showStructureBar`, because a fighter that has swung at full structure must show energy and a fighter at full energy with a wrecked structure must not.

### The refusal is shown by flashing the empty row

`EvNoEnergy` joins the event bits in actor.go. `presentEffects` consumes it into a local-only deadline on the actor (not replicated, like the camera effects it sits beside), and while it runs the energy bar's empty track is drawn even though the bar would otherwise be hidden. This reuses the bar the player is already reading instead of adding a HUD element that exists only to say "no".

The client's reconciliation already saves and restores `local.Events` around replay (net_client.go:281), so a replayed refusal does not flash twice.

## Risks / Trade-offs

- **A fighter can be locked out of offence for several seconds** → That is the mechanic, but the floor matters: from empty, regeneration gives a swing back 1.5 s + 0.8 s later, and a single correct block (15, more than a swing costs) cuts that to nothing. The first playtest found the original 20-per-swing drained too fast; 12 is the retune, and the numbers remain three named constants in one block of actor.go.
- **Bots mash attack and never block** → Bots drive the same `Actor.Update` path, so they will spend into refusals and idle through the regen wait, which will read as hesitation. No AI change is in scope here; if it looks bad in play, bot input is a separate change.
- **Predicted energy diverges from the server** → Bounded by reconciliation: every snapshot resets the pool and its refill wait to the server's values and replays only unacknowledged commands. With regeneration predicted from replicated inputs, the only divergence left is a block restore in flight, which lasts until the snapshot that reports the block itself.
- **Version bump disconnects old clients** → Intended, and already the handshake's job. Local play and the dedicated server are updated together from one build.
- **A 16 px bar gives energy about six points per pixel** → Individual swings move it two pixels, which is visible; single points of regeneration are not. That is the right way round — the player needs to read roughly how much is left, and exactly how much is a number the HUD has never shown for structure either.

## Migration Plan

No persisted data holds energy: levels (`level_1.json`) carry tiles and objects, and the rewind buffer holds no pools. A fighter constructed by `NewActor` starts at `MaxEnergy`, so every existing spawn path — match start, respawn, bots, the editor — gets a full pool with no migration step.

The only compatibility boundary is the wire, handled by the `ProtocolVersion` bump. Rollback is reverting the change: nothing outside the build persists a value that a reverted build would misread.
