## MODIFIED Requirements

### Requirement: A landed hit costs structure and flinches
A hit that lands SHALL cost the defender 1 structure and SHALL produce a flinch: the hit flash, the hit-stop freeze that both fighters already share, a small knockback and the hit animation. A flinch SHALL NOT lock the defender out of acting — it can move, attack, block and jump during one — and SHALL NOT interrupt an action the defender had already started.

The knockback SHALL move the fighter, not merely be given to it: after a flinch the fighter SHALL be somewhere it was not when the hit landed. It SHALL remain clearly smaller than the throw a stun gives, so the two read as different events.

#### Scenario: First hit does not stun
- **WHEN** a fighter at full structure takes one hit
- **THEN** its structure is 11, it flashes and is knocked back slightly, and it can act on the next tick

#### Scenario: The small knockback actually moves the fighter
- **WHEN** a fighter is flinched and then left alone until the knockback has run out
- **THEN** its position differs from where it was when the hit landed

#### Scenario: A flinch stays smaller than a stun
- **WHEN** one fighter is flinched and another stunned by attackers facing the same way, and both are left alone until their knockbacks have run out
- **THEN** the flinched one has travelled a visibly shorter distance than the stunned one

#### Scenario: Flinch does not interrupt a swing
- **WHEN** a fighter is hit while its own attack animation is playing
- **THEN** that attack continues and can still land
