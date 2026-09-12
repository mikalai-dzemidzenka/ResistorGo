## MODIFIED Requirements

### Requirement: Snapshots
The server SHALL send every client a snapshot 20 times per second on an unreliable channel. A snapshot SHALL contain the server tick and, for every fighter, the full state needed to predict or display it: id, position, velocity, state, facing, guard height, structure, animation and its progress, and combat flags. Each snapshot SHALL also carry the sequence number of the last command the server executed for the receiving player. A snapshot for up to 16 fighters SHALL fit in one packet of at most 1200 bytes.

#### Scenario: Snapshot rate
- **WHEN** a client is connected for 1 second without packet loss
- **THEN** it receives 20 snapshots (±1)

#### Scenario: Acknowledgement
- **WHEN** the server has executed a client's commands up to sequence 500 before building a snapshot
- **THEN** the snapshot sent to that client reports 500 as its last executed command

#### Scenario: Structure is replicated
- **WHEN** a fighter's structure changes on the server
- **THEN** the next snapshot carries the new value and every client shows it, to a precision of 0.1 structure or better

#### Scenario: Still fits the packet
- **WHEN** a snapshot carries 16 fighters, each with structure
- **THEN** the packet is at most 1200 bytes
