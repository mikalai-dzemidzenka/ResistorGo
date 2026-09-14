## MODIFIED Requirements

### Requirement: Every third landed hit stuns
The third hit to land on a fighter SHALL stun it: the defender enters the hit-stun state, is locked out until the stun animation finishes, and is thrown back noticeably further than a flinch throws it. The count SHALL be per fighter and SHALL reset when the stun lands, when the fighter dies, and when its structure begins regenerating. A block SHALL NOT advance it.

The throw-back SHALL move the fighter, not merely be given to it: a stunned fighter SHALL end up a real distance from where the hit landed, far enough that it reads as being thrown rather than as a jitter. Being locked out SHALL NOT cap how far the hit carries it — the lockout governs what the fighter may do, never how far the attack throws it — and the throw SHALL go the way the attacker was facing.

#### Scenario: Third hit of a run
- **WHEN** a fighter takes three hits without its structure regenerating in between
- **THEN** the first two flinch it and the third stuns it, locking it out and throwing it back further

#### Scenario: The throw carries a real distance
- **WHEN** a fighter is stunned and then left alone until the throw has run out
- **THEN** it has been carried at least ten whole pixels from where it was hit

#### Scenario: The lockout does not cap the throw
- **WHEN** a fighter is stunned
- **THEN** the distance it travels is set by the knockback the hit gave it and the friction that slows it, and by nothing belonging to the state it was put into

#### Scenario: The throw follows the attacker's facing
- **WHEN** a fighter is stunned by an attacker facing left, and another by an attacker facing right
- **THEN** each is carried the way its own attacker was facing

#### Scenario: The count restarts after a stun
- **WHEN** a fighter has just been stunned and takes another hit
- **THEN** that hit flinches it, and the stun comes two hits later

#### Scenario: Blocks do not count
- **WHEN** a fighter takes a hit, blocks two attacks, then takes another hit
- **THEN** neither hit stuns it
