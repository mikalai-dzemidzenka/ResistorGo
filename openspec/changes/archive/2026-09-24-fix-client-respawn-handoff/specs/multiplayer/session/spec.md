## MODIFIED Requirements

### Requirement: Player slots and spawning
Every accepted player SHALL receive a unique player id and a fighter. The fighter SHALL spawn at a spawn position not occupied by another fighter's collision box. All other clients SHALL see the new fighter no later than the second snapshot after it spawns. A player whose fighter dies SHALL keep its slot and player id while it has no fighter, and SHALL be given a new one when it respawns; its slot SHALL NOT be offered to a joining player in the meantime.

A player's fighter SHALL be able to change during a session, and the client SHALL be told each time it does. Being given a fighter is not something that happens only when joining: a client SHALL end up predicting and controlling whichever fighter the server currently holds for it, whether that is the one it joined with or one it was given later.

#### Scenario: Second player joins
- **WHEN** a second player joins a match with one player in it
- **THEN** both clients show two fighters, and the fighters do not overlap at spawn

#### Scenario: Slot survives a death
- **WHEN** a player's fighter dies and is removed from the world
- **THEN** the player stays connected and keeps its slot and player id, and a client connecting during that gap is told the server is full if every slot is taken

#### Scenario: Respawn takes a free position
- **WHEN** a dead player's fighter respawns while another fighter is standing on the spawn position it used before
- **THEN** it spawns at a position not occupied by another fighter's collision box, and every client sees it no later than the second snapshot after it spawns

#### Scenario: Control follows the new fighter
- **WHEN** a player's fighter respawns after a death
- **THEN** that player's input drives the new fighter, and the client predicts it as it predicted the one before

#### Scenario: A second death in one session
- **WHEN** a player dies, respawns, and dies again without leaving
- **THEN** each respawn hands the client the fighter it is given, and control returns every time
