# multiplayer/replication Specification

## Purpose

Defines how the authoritative server and clients exchange simulation data: the server tick, client input commands, world snapshots, and the reliability rules that keep them consistent over a lossy network.

## Requirements

### Requirement: Authoritative fixed-tick server
The server SHALL simulate the world at a fixed 60 ticks per second with a monotonically increasing tick number. Server results are authoritative: positions, states, hits and blocks. No client-reported state SHALL override them.

#### Scenario: Client cannot set its position
- **WHEN** a client sends anything other than input commands for its fighter
- **THEN** the server ignores it, and the fighter's state is determined only by server simulation

### Requirement: Input commands
Each client simulation tick, the client SHALL produce one input command. The command SHALL contain a sequence number, the movement buttons (left, right), the edge-triggered actions (attack, jump, climb down), the guard height (low, mid, high or block-all), the desired facing, and the server tick the client was displaying for other fighters (the view tick). Every packet sent SHALL also repeat the most recent previously sent commands (at least the last 2), so a single lost packet does not lose input.

#### Scenario: Lost input packet
- **WHEN** one input packet from a client is lost
- **THEN** the server still receives the commands it contained from the next packet, and executes each exactly once

#### Scenario: Duplicate commands
- **WHEN** the server receives a command whose sequence number it has already executed
- **THEN** that command is discarded

#### Scenario: Climb down reaches the server
- **WHEN** a client presses S while its fighter stands on a drop-through platform
- **THEN** the server receives a command with climb down set, and its fighter climbs down on the tick that command is executed

### Requirement: Per-player command execution
The server SHALL advance each human player's fighter by exactly one simulation tick per executed command of that player, in sequence order. Per server tick it SHALL execute the commands received for that player, limited so a player cannot average more than one command per server tick (a short burst allowance for jitter is permitted). If no commands arrive, the fighter SHALL not advance on its own, except for reactions forced on it by others (being hit or blocking).

#### Scenario: Normal play
- **WHEN** a client sends one command per tick at a steady rate
- **THEN** its fighter advances one tick per server tick

#### Scenario: Speed-hack attempt
- **WHEN** a client sends commands at twice the tick rate for several seconds
- **THEN** the server executes no more than one command per tick on average, and the excess is dropped

#### Scenario: Stalled client
- **WHEN** a client's commands stop arriving for 500 ms
- **THEN** its fighter holds position and state (unless hit), and resumes when commands arrive again

### Requirement: Snapshots
The server SHALL send every client a snapshot 20 times per second on an unreliable channel. A snapshot SHALL contain the server tick and, for every fighter, the full state needed to predict or display it: id, position, velocity, state, facing, guard height, structure, energy and the wait since its last swing, landed hits toward the next stun, animation and its progress, and combat flags. Each snapshot SHALL also carry the sequence number of the last command the server executed for the receiving player. A snapshot for up to 16 fighters SHALL fit in one packet of at most 1200 bytes.

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

#### Scenario: Still fits the packet
- **WHEN** a snapshot carries 16 fighters, each with structure and energy
- **THEN** the packet is at most 1200 bytes

### Requirement: Loss and reordering tolerance
Clients SHALL tolerate lost, duplicated and reordered snapshots. A snapshot older than the newest one already received SHALL be discarded. Gameplay SHALL never wait for a missing snapshot.

#### Scenario: Out-of-order snapshot
- **WHEN** a snapshot for tick 300 arrives after the snapshot for tick 303
- **THEN** the tick-300 snapshot is discarded

#### Scenario: Burst loss
- **WHEN** 5 consecutive snapshots are lost
- **THEN** the client keeps rendering and accepting input, and resumes normally with the next snapshot

### Requirement: Reliable control messages
Connection control messages SHALL be delivered on a reliable ordered channel separate from snapshots and input: handshake, welcome, rejection with reason, and orderly disconnect reason. Loss of snapshots or input SHALL NOT delay control messages, and control messages SHALL NOT delay snapshots or input.

#### Scenario: Welcome under loss
- **WHEN** a client connects over a link with 20 % packet loss
- **THEN** it still receives its welcome message and joins the match

### Requirement: Robust packet handling
Malformed, truncated or unknown packets SHALL be ignored without crashing the server or client, and without affecting other players.

#### Scenario: Garbage packet
- **WHEN** a peer sends a packet with an unknown message type or an invalid length
- **THEN** the packet is dropped and the match continues
