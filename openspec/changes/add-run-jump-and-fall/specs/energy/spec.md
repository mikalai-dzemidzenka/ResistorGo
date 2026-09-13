## MODIFIED Requirements

### Requirement: Energy pool
Every fighter SHALL have an energy pool with a capacity of 200 points, full when the fighter enters the world. Energy SHALL never exceed that capacity or fall below zero. Energy SHALL gate attacking and the running jump: walking, running, the standing jump, climbing and blocking SHALL cost nothing and SHALL be available at any energy.

#### Scenario: Full at spawn
- **WHEN** a fighter spawns, whether at the start of a match or on respawn
- **THEN** its energy is 200

#### Scenario: Restoring cannot overfill
- **WHEN** a fighter at full energy blocks an attack at the attacker's height
- **THEN** its energy is still 200

#### Scenario: An empty fighter can still defend
- **WHEN** a fighter at 0 energy is attacked
- **THEN** it can move, jump from a standstill and block normally, and blocking still resolves as it always does

#### Scenario: An empty fighter still gets around
- **WHEN** a fighter at 0 energy runs, climbs a ledge and jumps from a standstill
- **THEN** none of them is refused and none of them costs anything
