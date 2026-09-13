## MODIFIED Requirements

### Requirement: Structure regenerates after a pause
When a fighter has not lost structure for 1.5 seconds, its structure SHALL refill at 1 per second up to capacity. Losing structure SHALL restart that wait.

#### Scenario: Recovering between exchanges
- **WHEN** a fighter that has lost 4 structure is left alone for 5.5 seconds
- **THEN** its structure has refilled by 4 and it is back at full

#### Scenario: Half recovered
- **WHEN** a fighter that has lost 4 structure is left alone for 3.5 seconds, which once refilled all of it
- **THEN** its structure has refilled by 2, and it is still 2 short of full

#### Scenario: Pressure prevents recovery
- **WHEN** a fighter is hit once per second
- **THEN** its structure never regenerates, and it is stunned on every third hit
