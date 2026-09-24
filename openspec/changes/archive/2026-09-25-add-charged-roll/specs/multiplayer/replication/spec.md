## MODIFIED Requirements

### Requirement: Input commands
Each client simulation tick, the client SHALL produce one input command. The command SHALL contain a sequence number, the movement buttons (left, right), the edge-triggered actions (attack, jump, climb down), whether the roll button is held, the guard height (low, mid, high or block-all), the desired facing, and the server tick the client was displaying for other fighters (the view tick). Every packet sent SHALL also repeat the most recent previously sent commands (at least the last 2), so a single lost packet does not lose input.

#### Scenario: Lost input packet
- **WHEN** one input packet from a client is lost
- **THEN** the server still receives the commands it contained from the next packet, and executes each exactly once

#### Scenario: Duplicate commands
- **WHEN** the server receives a command whose sequence number it has already executed
- **THEN** that command is discarded

#### Scenario: Climb down reaches the server
- **WHEN** a client presses S while its fighter stands on a drop-through platform
- **THEN** the server receives a command with climb down set, and its fighter climbs down on the tick that command is executed

#### Scenario: A roll reaches the server
- **WHEN** a client holds the roll button for some ticks and releases it
- **THEN** the server's fighter ducks on the first command with the button held and rolls on the first command without it, with the same charge the client predicted
