## Why

Attacking is currently free: a fighter can hold the attack button and swing forever, so the only thing an exchange costs is the risk of being hit back. Structure already makes taking damage accumulate, but nothing makes *throwing* offence accumulate, which leaves mashing as strong as reading the opponent. Energy gives offence a price and makes a correct guard pay for the next swing, so the fight has a rhythm of spending and earning instead of one continuous swing.

## What Changes

- Every fighter gets an energy pool, full at spawn, capped at 200 and never below zero.
- Starting a swing costs 12 energy — sixteen swings from a full pool. This applies to the first swing of an attack and to every chained swing that follows it, since each is a new attack.
- A swing that cannot be paid for does not happen: the attack input is refused, the fighter stays in the state it was in, and a presentation event fires so the renderer can show the refusal.
- A successful block at the attacker's height restores 15 energy. Blocking with block-all (guard height 3) restores nothing — it already trades structure for covering every height, and it must not also fund the counter-attack.
- Energy regenerates on its own once the fighter has not spent any for a moment, on the same calm-then-refill shape structure already uses, so a fighter that is not blocking still returns to fighting strength instead of being locked out for good.
- Each fighter shows its energy as a second, thinner bar beneath its structure bar, draining as energy is spent, so the cost of a swing and the payoff of a block are readable in play.
- Energy is decided by the authoritative simulation and replicated in the snapshot, like structure; a predicting client displays it and never invents it.

## Capabilities

### New Capabilities
- `energy`: the energy pool — its capacity, what a swing costs, what a block pays back, how it regenerates, what happens when a fighter cannot afford a swing, and who decides it.

### Modified Capabilities
- `structure-bar`: the bar over a fighter's head gains an energy bar below it, with its own visibility rule — energy is worth showing when it is *below* full, and the bar's existing "only while structure is below full" rule cannot govern it.
- `multiplayer/replication`: the snapshot's per-fighter record gains energy, and the packet budget must still hold with it.

## Impact

- `actor.go`: the already-declared but unused `Actor.Energy` field becomes live; `enterAttackState` and the chain branch of `updateAttack` gain a cost check; `confirmBlock` grants energy on a matching guard; a `regenEnergy` step joins `regenStructure` in `Update`; a new `EvNoEnergy` presentation event joins `EvAttackStart`/`EvHit`/`EvBlocked`.
- `structure_bar.go`: an energy track and fill added to `structureBarLayout` and `drawStructureBar`, in the same world-pixel, palette-exact style as the bar above it.
- `net_protocol.go`: `ActorRec` gains an energy byte, `actorRecSize` grows from 42 to 43, and `recOf`/`apply`/`encodeSnapshot`/`decodeSnapshot` carry it.
- `game.go`: `presentEffects` consumes the new event.
- Tests: `structure_test.go` and `net_protocol_test.go` gain energy cases; the 16-fighter packet-size test must still pass at the larger record size.
- Not affected: the rewind ring buffer, which records only position, speed, orientation and animation and holds no pool of either kind.
