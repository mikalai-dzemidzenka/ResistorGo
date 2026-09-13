## Why

Energy takes too long to come back. At 15 per second a fighter that has emptied its pool waits 13 seconds to be whole, and a burst of five swings costs four seconds of standing around before the fifth can be thrown again. Structure was slowed to 1 per second so that damage means something; energy was left where it was, and the two now pull the same way — a fighter that commits offence and gets nothing for it is out of the fight for as long as one that has been hurt by it, which is not what energy is for. Energy is meant to price a swing, not to bench the player who threw it.

The refusal is also the harshest thing in the game to sit through, because it is the one state the player cannot act their way out of: blocking pays, but only if the opponent attacks. Making the pool come back twice as fast keeps the price of a swing exactly what it was and shortens only the punishment for having spent.

## What Changes

- Energy refills at **30 per second** instead of 15, so an empty pool is full in about 6.7 seconds rather than 13.3, and a single swing's 12 points are earned back in 0.4 seconds rather than 0.8.
- The 1.5-second calm wait before refilling starts is **unchanged**. Pressure still denies recovery entirely: a fighter swinging once a second never starts refilling at all, which is what stops a mashing fighter from being funded by the rule.
- Nothing else about energy moves: a swing still costs 12, a block at the attacker's height still restores 15, block-all still restores nothing, the pool is still 200, and the refusal still stands when a swing cannot be paid for.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `energy`: *Energy regenerates after a pause* — the refill rate goes from 15 per second to 30.

## Impact

- `actor.go` — one constant: the tick interval between restored points halves. The rule reading it is unchanged, and so is the wait that gates it.
- `energy_test.go` — the regeneration test's timings move with the rate, and gain a check partway through so the rate is pinned from both sides rather than only at the moment the pool is full.
- Nothing on the wire: the rate is a constant both sides compile in, and everything regeneration is decided from — the tick counter and the replicated calm wait — is unchanged, so client and server still refill on the same tick.
