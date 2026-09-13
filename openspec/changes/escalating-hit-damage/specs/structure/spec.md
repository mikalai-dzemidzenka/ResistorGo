## MODIFIED Requirements

### Requirement: A landed hit costs structure and flinches
A hit that lands SHALL cost the defender structure and SHALL produce a flinch: the hit flash, the hit-stop freeze that both fighters already share, and a small knockback. A flinch SHALL NOT play an animation of its own: the defender SHALL keep whatever animation it was already playing, and SHALL be shown to have been hit by the flash alone. A flinch SHALL NOT lock the defender out of acting — it can move, attack, block and jump during one — and SHALL NOT interrupt an action the defender had already started.

The flash SHALL hold for about 0.2 seconds and then stop on its own. It SHALL also stop early with whatever replaces the flinch: a fighter that swings, blocks, is stunned or dies while flashing SHALL stop flashing at that moment, so no fighter is ever left lit.

What a hit costs SHALL depend on how many hits have already landed on that fighter in the same run. The first hit of a run SHALL cost 1 structure; each hit after it SHALL cost half a point more than the one before — 1, 1.5, 2, 2.5, 3 — and the cost SHALL NOT climb past 3 however long the run goes on. The run SHALL be per fighter: it counts what has landed on that fighter, whoever landed it.

#### Scenario: First hit does not stun
- **WHEN** a fighter at full structure takes one hit
- **THEN** its structure is 11, it flashes and is knocked back slightly, and it can act on the next tick

#### Scenario: A run costs more as it goes
- **WHEN** a fighter at full structure takes five hits in a row with nothing to break the run
- **THEN** they cost 1, 1.5, 2, 2.5 and 3, leaving it at 2 structure

#### Scenario: The cost stops climbing
- **WHEN** a run reaches a sixth and a seventh hit
- **THEN** each costs 3, the same as the fifth

#### Scenario: Six hits end a fight
- **WHEN** a fighter at full structure takes six hits in a row
- **THEN** its structure reaches zero and it dies

#### Scenario: The stun does not relieve it
- **WHEN** a fighter takes four hits in a row, the third of which stuns it
- **THEN** the fourth costs 2.5, as the fourth hit of a run does, rather than starting the run again

#### Scenario: Time ends a run
- **WHEN** a fighter is left alone long enough for its structure to start refilling, and is then hit again
- **THEN** that hit costs 1, as the first hit of a run does

#### Scenario: Flinch does not interrupt a swing
- **WHEN** a fighter is hit while its own attack animation is playing
- **THEN** that attack continues and can still land

#### Scenario: A flinch leaves the animation alone
- **WHEN** a running fighter takes a hit that does not stun it
- **THEN** it goes on running, with the running animation uninterrupted, and flashes

#### Scenario: The flash runs out
- **WHEN** a fighter has been flinched and half a second passes with nothing else happening to it
- **THEN** it is no longer flashing

#### Scenario: Acting ends the flash
- **WHEN** a flinched fighter swings while it is still flashing
- **THEN** the flash stops as the swing starts

### Requirement: Blocking and structure
A block that holds until the swing's hit frame SHALL restore 0.5 structure to the defender, except when the defender is guarding with block-all (guard height 3), which SHALL instead cost 0.4 structure — less than a landed hit costs. A block SHALL NOT flinch or stun the defender.

A successful block SHALL end the run of hits against the defender, whichever guard stopped the attack: a guard at the attacker's height or block-all. The next hit to land on that fighter SHALL cost 1, as the first hit of a run does. Stopping an attack is therefore worth more than the structure it restores — it is how a defender gets out from under a run.

#### Scenario: Matching guard restores
- **WHEN** a fighter at 8 structure blocks an attack with a guard at the attacker's height
- **THEN** its structure is 8.5 and it is not flinched

#### Scenario: Block-all costs, but less than a hit
- **WHEN** a fighter at 8 structure blocks an attack while guarding with block-all
- **THEN** its structure is 7.6, which is more structure than the 7 it would have had if the attack had landed

#### Scenario: A block breaks the run
- **WHEN** a fighter that has taken three hits in a row blocks the next attack at the attacker's height, and is then hit again
- **THEN** that hit costs 1 rather than the 2.5 a fourth hit of the run would have cost

#### Scenario: Block-all breaks it too
- **WHEN** a fighter that has taken three hits in a row stops the next attack with block-all, and is then hit again
- **THEN** that hit costs 1

#### Scenario: A guard that drops mid-swing is a hit
- **WHEN** a fighter starts blocking but changes guard height before the swing's hit frame, so that it no longer matches and is not block-all
- **THEN** the attack lands as a hit, with a hit's structure cost, instead of restoring any
