## Why

Every hit costs the same. The twelfth hit of a beating costs exactly what the first one did, so a fighter under sustained pressure is in no more danger on its fifth mistake than on its first, and a defender who has already lost the read four times running is punished at the same rate as one who slipped once. Structure accumulates, but nothing about the fight does.

That flattens the thing the guard system is for. Reading the attacker's height is how a defender gets out of a run, and the reward for reading it right is currently a flat 0.5 structure — worth having, but no more valuable after four hits than after none. Making each hit in a run cost more turns a run into something that has to be broken, and makes the block that breaks it the most valuable thing a defender can do.

## What Changes

- **A run of hits costs more as it goes.** The first hit of a run costs 1 structure as it does today, and each hit after it in the same run costs half a point more: 1, 1.5, 2, 2.5, 3.
- **The cost stops climbing at 3**, three times what a hit costs today. A long run stays lethal without becoming arbitrary, and the numbers stay small enough to carry.
- **A successful block ends the run**, whichever guard stopped it — a guard at the attacker's height or block-all. The next hit to land costs 1 again. This is the defender's way out, and the reason a correct read is worth more than the structure it restores.
- **The run survives a stun.** The stun every third hit lands does not relieve the pressure: a five-hit run really is five hits of climbing cost, not two runs of escalation separated by a lockout.
- **Time still ends a run**, on the same terms structure already recovers on: once a fighter has been left alone long enough for its structure to start refilling, the run is over and the next hit costs 1. Pressure has to be sustained to be worth anything.
- A clean six-hit run now ends a full-structure fighter, where twelve hits were needed before.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `structure`: *A landed hit costs structure and flinches* — a hit's cost is no longer flat. *Blocking and structure* — a block now also ends the run.
- `multiplayer/replication`: *Snapshots* — the run's length has to reach clients, or a client would predict the wrong cost for its own fighter and show the wrong bar for everyone else's.

## Impact

- The run is a new per-fighter count, separate from the count toward the next stun: that one clears every third hit by design, and this one must not. Both live on the fighter and both are replicated.
- **BREAKING on the wire.** The run's length rides in spare bits of the snapshot's existing flag word, so the record does not grow — but a client reading those bits as anything else would misread them, so the protocol version has to move and an older client is refused at the handshake, as it already is for any version difference.
- Structure spending itself is untouched: what changes is the number handed to it, not how a hit resolves, what it does to the defender, or when the third hit stuns.
- This lands on top of the in-flight flinch change, which is already rewriting the same requirement to take the hit animation off a flinch. The delta here carries that wording forward rather than reverting it.
