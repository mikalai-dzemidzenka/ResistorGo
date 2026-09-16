## Why

Energy takes too long to come back. Energy is meant to price a swing, not to bench the player who threw it, and the refusal is the harshest thing in the game to sit through, because it is the one state the player cannot act their way out of: blocking pays, but only if the opponent attacks.

This change first doubled the single refill rate. Since then the pool was retuned (commit `32cebbf`) to a larger capacity with two rates — a slow one while the fighter is still being attacked and a fast one once nothing has reached it for a while — sized so that an empty pool took 20 and 10 seconds to refill. That brought the long wait back. This change now doubles both of those rates, and records the two-rate rule the spec never caught up with.

## What Changes

- Energy refills at **30 per second** while attacks are still reaching the fighter (was 15), and at **60 per second** once no attack has reached it for 2 seconds (was 30). An empty pool is full in 10 and 5 seconds of refilling rather than 20 and 10, and a single swing's 12 points come back in 0.2 seconds when left alone.
- The 1.5-second wait after a swing before refilling starts is **unchanged**, and so is the 2-second wait before the fast rate. Pressure still denies recovery entirely: a fighter swinging once a second never starts refilling at all, which is what stops a mashing fighter from being funded by the rule.
- The spec's regeneration requirement is brought up to the two-rate rule the code already has, with the doubled rates.
- Nothing else about energy moves: a swing still costs 12, a block at the attacker's height still restores 15, block-all still restores nothing, and the refusal still stands when a swing cannot be paid for.

Not in scope: the pool's capacity, which the main spec still states as 200 while the code holds 300 — the in-flight `add-run-jump-and-fall` change owns that requirement, and the scenarios here are worded so they hold for either.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `energy`: *Energy regenerates after a pause* — becomes the two-rate rule, at 30 per second under attack and 60 per second left alone.

## Impact

- `actor.go` — two constants: `EnergyRefillSeconds` 20 → 10 and `EnergyFreeRefillSeconds` 10 → 5, which halves both tick gaps (4 → 2 and 2 → 1). `regenEnergy` is unchanged, and so are both waits that gate it.
- `energy_test.go` — the regeneration tests already derive their timings from the constants; a test pins the two rates at 30 and 60 per second so a retune cannot move them silently.
- Nothing on the wire: the rates are constants both sides compile in, and everything regeneration is decided from — the tick counter and the two replicated waits — is unchanged, so client and server still refill on the same tick.
