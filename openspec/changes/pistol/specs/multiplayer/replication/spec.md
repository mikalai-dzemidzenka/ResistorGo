## MODIFIED Requirements

### Requirement: Input commands
Each client simulation tick, the client SHALL produce one input command. The command SHALL contain a sequence number, the movement buttons (left, right), the edge-triggered actions (attack, jump, climb down, draw or holster the gun), whether the roll button is held, the guard height (low, mid, high or block-all), the gun's aim, the desired facing, and the server tick the client was displaying for other fighters (the view tick). Every packet sent SHALL also repeat the most recent previously sent commands (at least the last 2), so a single lost packet does not lose input.

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

#### Scenario: Drawing the gun reaches the server
- **WHEN** a client presses Q while its fighter stands
- **THEN** the server receives a command with the draw set, and its fighter draws the pistol on the tick that command is executed

#### Scenario: A shot reaches the server along the client's aim
- **WHEN** a client holding the pistol clicks LMB with the mouse above its fighter
- **THEN** the server's fighter fires on the tick that command is executed, along the same aim the client predicted

### Requirement: Snapshots
The server SHALL send every client a snapshot 20 times per second on an unreliable channel. A snapshot SHALL contain the server tick and, for every fighter, the full state needed to predict or display it: id, position, velocity, state, facing, guard height, structure, energy and the wait since its last swing, landed hits toward the next stun, animation and its progress, combat flags, the gun it holds (if any), its aim, the ticks left before it can fire again, and a count of the shots it has fired. Each snapshot SHALL also carry the sequence number of the last command the server executed for the receiving player. A snapshot for up to 16 fighters SHALL fit in one packet of at most 1200 bytes.

#### Scenario: Snapshot rate
- **WHEN** a client is connected for 1 second without packet loss
- **THEN** it receives 20 snapshots (±1)

#### Scenario: Acknowledgement
- **WHEN** the server has executed a client's commands up to sequence 500 before building a snapshot
- **THEN** the snapshot sent to that client reports 500 as its last executed command

#### Scenario: Structure is replicated
- **WHEN** a fighter's structure changes on the server
- **THEN** the next snapshot carries the new value and every client shows it, to a precision of 0.1 structure or better

#### Scenario: Energy is replicated
- **WHEN** a fighter's energy changes on the server, whether by a swing, a block or regeneration
- **THEN** the next snapshot carries the new value, exactly, over the whole range from 0 to 200

#### Scenario: The refill wait is replicated
- **WHEN** a fighter has been waiting half a second since its last swing for its energy to start refilling
- **THEN** the next snapshot carries that wait, so a client refills on the same tick the server does

#### Scenario: Stun progress is replicated
- **WHEN** a fighter has taken two hits toward its next stun on the server
- **THEN** the next snapshot carries a count of two, and when the stun lands the following snapshot carries zero

#### Scenario: The gun and its aim are replicated
- **WHEN** another player holds the pistol and aims low
- **THEN** every client draws that fighter with the pistol at its lowest direction

#### Scenario: Shots are heard from the count
- **WHEN** another player fires between two snapshots
- **THEN** the next snapshot's shot count has gone up by one, and every client shows that shot once

#### Scenario: Still fits the packet
- **WHEN** a snapshot carries 16 fighters, each with structure, energy and gun state
- **THEN** the packet is at most 1200 bytes
