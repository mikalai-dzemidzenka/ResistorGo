## ADDED Requirements

### Requirement: A kill fills the killer's pool
When a fighter dies from an attack, the fighter whose attack killed it SHALL be set to full energy at that moment. This SHALL hold whether the killing attack landed or broke through a block-all, and for bots and players alike. An attack that does not kill SHALL NOT change the attacker's energy. The refill SHALL NOT restart or finish either regeneration wait. As with a block's payback, the refill SHALL be the authority's alone and SHALL reach a client only in a snapshot.

#### Scenario: A landed kill
- **WHEN** a fighter with 20 energy lands the hit that breaks its opponent's structure
- **THEN** the fighter is at full energy as its opponent falls

#### Scenario: Through a block-all
- **WHEN** a fighter's attack is blocked by an opponent guarding all heights, and paying for that block breaks the opponent's structure
- **THEN** the attacking fighter is at full energy

#### Scenario: A hit that does not kill
- **WHEN** a fighter lands a hit that leaves its opponent standing
- **THEN** the fighter's energy is unchanged by the hit

#### Scenario: Bots are rewarded too
- **WHEN** a bot kills a player
- **THEN** the bot is at full energy

#### Scenario: Online
- **WHEN** a client's fighter kills an opponent while its energy is low
- **THEN** its energy becomes full when the snapshot carrying the kill arrives, and the crosshair shows its full-pool flash
