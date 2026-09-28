## MODIFIED Requirements

### Requirement: Snapshots
The server SHALL send every client a snapshot 20 times per second on an unreliable channel. A snapshot SHALL contain the server tick and, for every fighter, the full state needed to predict or display it: id, position, velocity, state, facing, guard height, structure, energy and the wait since its last swing, team, landed hits toward the next stun, the length of the run of hits against it, animation and its progress, combat flags, the gun it holds (if any), its aim, the ticks left before it can fire again, and a count of the shots it has fired. Each snapshot SHALL also carry the sequence number of the last command the server executed for the receiving player. A snapshot for up to 16 fighters SHALL fit in one packet of at most 1200 bytes.

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

#### Scenario: The run of hits is replicated
- **WHEN** a fighter has taken three hits in a row on the server
- **THEN** the next snapshot carries a run of three, so every client agrees on what the next hit costs

#### Scenario: A broken run is replicated
- **WHEN** a fighter under a run of hits blocks an attack
- **THEN** the next snapshot carries a run of zero

#### Scenario: Still fits the packet
- **WHEN** a snapshot carries 16 fighters, each with structure, energy and gun state
- **THEN** the packet is at most 1200 bytes

#### Scenario: Team is replicated
- **WHEN** a fighter's team changes on the server
- **THEN** the next snapshot carries the new team, and every client treats that fighter as an ally or not as the server does
