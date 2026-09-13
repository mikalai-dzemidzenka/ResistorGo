## MODIFIED Requirements

### Requirement: A landed hit costs structure and flinches
A hit that lands SHALL cost the defender 1 structure and SHALL produce a flinch: the hit flash, the hit-stop freeze that both fighters already share, and a small knockback. A flinch SHALL NOT play an animation of its own: the defender SHALL keep whatever animation it was already playing, and SHALL be shown to have been hit by the flash alone. A flinch SHALL NOT lock the defender out of acting — it can move, attack, block and jump during one — and SHALL NOT interrupt an action the defender had already started.

The flash SHALL hold for about 0.2 seconds and then stop on its own. It SHALL also stop early with whatever replaces the flinch: a fighter that swings, blocks, is stunned or dies while flashing SHALL stop flashing at that moment, so no fighter is ever left lit.

#### Scenario: First hit does not stun
- **WHEN** a fighter at full structure takes one hit
- **THEN** its structure is 11, it flashes and is knocked back slightly, and it can act on the next tick

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

### Requirement: Every third landed hit stuns
The third hit to land on a fighter SHALL stun it: the defender enters the hit-stun state, is locked out until the stun animation finishes, and is thrown back noticeably further than a flinch throws it. The stun SHALL be the only hit that plays the hit animation, so the animation tells the player their fighter has lost control rather than merely been hit; the flash SHALL cover the opening of that animation, not the whole of it. The count SHALL be per fighter and SHALL reset when the stun lands, when the fighter dies, and when its structure begins regenerating. A block SHALL NOT advance it.

#### Scenario: Third hit of a run
- **WHEN** a fighter takes three hits without its structure regenerating in between
- **THEN** the first two flinch it and the third stuns it, locking it out and throwing it back further

#### Scenario: Only the stun has an animation
- **WHEN** a fighter takes three hits in a row
- **THEN** neither of the first two changes the animation it was playing, and the third plays the hit animation

#### Scenario: The count restarts after a stun
- **WHEN** a fighter has just been stunned and takes another hit
- **THEN** that hit flinches it, and the stun comes two hits later

#### Scenario: Blocks do not count
- **WHEN** a fighter takes a hit, blocks two attacks, then takes another hit
- **THEN** neither hit stuns it
