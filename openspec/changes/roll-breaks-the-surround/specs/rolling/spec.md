## MODIFIED Requirements

### Requirement: Landing recovers from the roll
A rolling fighter that reaches the ground SHALL play the roll-recover animation once. During the recovery it SHALL NOT act. Any speed it lands with SHALL run out under friction. When the recovery finishes, the fighter SHALL go back to running if it is still moving fast enough to run, and to standing otherwise, following the same rule a landing from a fall uses. Reaching the ground SHALL also close the place ahead of the roll for 2 seconds (see the engagement capability, "A landed roll closes the place ahead of it").

#### Scenario: Recovering on landing
- **WHEN** a rolling fighter touches the ground
- **THEN** it plays the roll-recover animation once and then stands

#### Scenario: Nothing during recovery
- **WHEN** a recovering fighter presses attack or jump
- **THEN** neither happens

#### Scenario: Landing closes the place ahead
- **WHEN** a rolling fighter touches the ground
- **THEN** the place on the side it was rolling toward closes for 2 seconds
