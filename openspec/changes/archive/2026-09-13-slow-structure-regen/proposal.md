## Why

Structure comes back too fast to matter. At 2 per second a fighter refills its whole 12-point pool in six seconds, and the four points a three-hit stun run costs are back two seconds after the pressure stops — often before the fighter that landed them has closed the distance again. Damage that repays itself that quickly turns an exchange into something you can walk away from rather than something you have to survive, and it flattens the stun count with it: the count clears when regeneration begins, so a short pause undoes the progress toward a stun as well as the structure.

Playtesting the game with energy in it made this plainer. Energy gives offence a price and a recovery of its own, so a fighter who commits a pool of swings and gets nothing for it is now paying twice — while the fighter who took those hits is already whole. The two pools should not recover at the same speed; the one that decides whether you go down should be the slower.

## What Changes

- Structure refills at **1 per second** instead of 2, so a full pool takes 12 seconds rather than 6, and the 4 points a stun run costs take 4 seconds rather than 2.
- The 1.5-second calm wait before refilling starts is **unchanged**. What is slow is the healing, not the reaction to being left alone — a fighter still starts recovering promptly once the pressure is off, it just gets less for the time.
- Nothing else about structure moves: a hit still costs 1, a block still restores 0.5, block-all still costs 0.4, the stun still lands every third hit, and the pool is still 12.

## Capabilities

### Modified Capabilities
- `structure`: the *Structure regenerates after a pause* requirement's rate, and the scenario whose timings are derived from it.

## Impact

- `actor.go`: `StructureRegen` 2 → 1. `regenStructure` already multiplies it by the fixed timestep, so no other line changes.
- `structure_test.go`: `TestStructureRegen` waits 3.5 seconds for 4 structure, which is exactly the old rate's timing and will fail at the new one; it needs the new timing, and is the test that proves the change landed.
- No wire, save-file or protocol impact: `StructureRegen` is a simulation constant, structure itself already replicates as tenths, and nothing persists it.
- Not affected: energy's own regeneration, which is a separate pool with its own rate and its own calm wait, and is deliberately left faster.
