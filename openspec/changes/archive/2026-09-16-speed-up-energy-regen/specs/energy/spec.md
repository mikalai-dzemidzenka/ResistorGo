## MODIFIED Requirements

### Requirement: Energy regenerates after a pause
When a fighter has not spent energy for 1.5 seconds, its energy SHALL refill up to capacity at one of two rates: 30 per second while attacks are still reaching it, and 60 per second once no attack has reached it for 2 seconds. An attack reaches a fighter when it lands on it or is blocked by it, block-all included. A fighter that has not been attacked since it entered the world SHALL refill at the faster rate. Spending energy SHALL restart the 1.5-second wait. Being hit or blocking SHALL restart the 2-second wait but SHALL NOT restart the 1.5-second one, and restoring energy SHALL restart neither.

#### Scenario: Recovering between exchanges
- **WHEN** a fighter that has spent 60 energy is not attacked and does not swing for 2.5 seconds
- **THEN** its energy has refilled by 60 and it is back where it was before spending

#### Scenario: Part way back
- **WHEN** a fighter that has spent 60 energy is not attacked and does not swing for 2 seconds
- **THEN** its energy has refilled by 30, half of what it spent

#### Scenario: Refilling under attack
- **WHEN** a fighter that has not swung for 1.5 seconds is attacked at least once every 2 seconds
- **THEN** its energy refills at 30 per second, half the rate of a fighter left alone

#### Scenario: A swing is soon affordable again
- **WHEN** a fighter with no energy left is not attacked and does not swing for 2 seconds
- **THEN** it can swing again

#### Scenario: A combo denies recovery
- **WHEN** a fighter swings once per second
- **THEN** its energy never regenerates between those swings and it runs out

#### Scenario: Being hit does not stop recovery
- **WHEN** a fighter that has not swung for 1.5 seconds is hit
- **THEN** its energy keeps refilling, at 30 per second

#### Scenario: Fresh fighters refill fast
- **WHEN** a fighter that has not been attacked since it spawned swings and then waits 1.5 seconds
- **THEN** its energy refills at 60 per second from the first tick of refilling
