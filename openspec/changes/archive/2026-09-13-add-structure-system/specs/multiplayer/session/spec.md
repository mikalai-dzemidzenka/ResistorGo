## MODIFIED Requirements

### Requirement: Player slots and spawning
Every accepted player SHALL receive a unique player id and a fighter. The fighter SHALL spawn at a spawn position not occupied by another fighter's collision box. All other clients SHALL see the new fighter no later than the second snapshot after it spawns. A player whose fighter dies SHALL keep its slot and player id while it has no fighter, and SHALL be given a new one when it respawns; its slot SHALL NOT be offered to a joining player in the meantime.

#### Scenario: Second player joins
- **WHEN** a second player joins a match with one player in it
- **THEN** both clients show two fighters, and the fighters do not overlap at spawn

#### Scenario: Slot survives a death
- **WHEN** a player's fighter dies and is removed from the world
- **THEN** the player stays connected and keeps its slot and player id, and a client connecting during that gap is told the server is full if every slot is taken

#### Scenario: Respawn takes a free position
- **WHEN** a dead player's fighter respawns while another fighter is standing on the spawn position it used before
- **THEN** it spawns at a position not occupied by another fighter's collision box, and every client sees it no later than the second snapshot after it spawns
