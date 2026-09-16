## Why

Finishing an opponent takes a string of swings, so the fighter who lands the kill usually does it on an emptying pool — and then faces the next opponent unable to attack until the pool refills. A kill is the biggest thing a fighter can achieve, and it should put the winner straight back into the fight rather than bench them.

## What Changes

- **A kill fills the killer's energy pool.** The moment a fighter's structure breaks and it dies, the fighter whose attack killed it is set to full energy — whether the killing attack landed or broke through a block-all.
- Bots are rewarded the same way as players; the rule applies to every fighter.
- The refill is decided by the authority, like every other kill: a client sees it when the snapshot carrying the kill arrives, and with it the crosshair's full-pool flash from `energy-restore-flash`.
- The wait before refilling, the swing that made the kill, and the killer's structure are untouched.

Not in scope: rewarding structure for a kill, a kill counter or score, and anything for a fighter that dies without an attacker.

## Capabilities

### New Capabilities

- None.

### Modified Capabilities

- `energy`: a new requirement that a kill fills the killer's pool, decided by the authority. Added rather than a modification, so it cannot conflict with the in-flight energy changes.

## Impact

- `actor.go`: `enterDeathState` restores the attacker to full energy, next to where it already freezes the attacker for hit-stop. Both death paths — a landed hit in `takeHit` and a broken block-all in `confirmBlock` — go through it.
- `energy_test.go`: tests that each death path fills the killer's pool and that a hit that does not kill gives the attacker nothing.
- Nothing on the wire: hits are only resolved in an authoritative world, and energy is already replicated in every snapshot.
