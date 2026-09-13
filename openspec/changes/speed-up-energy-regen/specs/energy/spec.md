## MODIFIED Requirements

### Requirement: Energy regenerates after a pause
When a fighter has not spent energy for 1.5 seconds, its energy SHALL refill at 30 per second up to capacity. Spending energy SHALL restart that wait. Blocking, being hit, and restoring energy SHALL NOT restart it.

#### Scenario: Recovering between exchanges
- **WHEN** a fighter that has spent 60 energy is left alone for 3.5 seconds
- **THEN** its energy has refilled by 60 and it is back at full

#### Scenario: Part way back
- **WHEN** a fighter that has spent 60 energy is left alone for 2.5 seconds
- **THEN** its energy has refilled by 30, half of what it spent, and it is 30 short of full

#### Scenario: A swing is soon affordable again
- **WHEN** a fighter with no energy left is not attacked and does not swing for 2 seconds
- **THEN** it can swing again

#### Scenario: A combo denies recovery
- **WHEN** a fighter swings once per second
- **THEN** its energy never regenerates between those swings and it runs out after sixteen

#### Scenario: Being hit does not stop recovery
- **WHEN** a fighter that has not swung for 1.5 seconds is hit
- **THEN** its energy keeps refilling
