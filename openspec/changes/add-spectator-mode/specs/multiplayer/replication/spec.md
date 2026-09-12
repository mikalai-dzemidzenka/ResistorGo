## MODIFIED Requirements

### Requirement: Input commands
Each client simulation tick, a client that controls a fighter SHALL produce one input command. The command SHALL contain a sequence number, the movement buttons (left, right), the edge-triggered actions (attack, jump), the guard height (low, mid, high or block-all), the desired facing, and the server tick the client was displaying for other fighters (the view tick). Every packet sent SHALL also repeat the most recent previously sent commands (at least the last 2), so a single lost packet does not lose input. A spectator client SHALL NOT send input commands.

#### Scenario: Lost input packet
- **WHEN** one input packet from a client is lost
- **THEN** the server still receives the commands it contained from the next packet, and executes each exactly once

#### Scenario: Duplicate commands
- **WHEN** the server receives a command whose sequence number it has already executed
- **THEN** that command is discarded

#### Scenario: Spectator sends no commands
- **WHEN** a spectator watches a match
- **THEN** it sends no input commands to the server
