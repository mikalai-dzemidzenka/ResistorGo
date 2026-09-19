## ADDED Requirements

### Requirement: Every actor state survives the wire

A snapshot SHALL carry a fighter in any state the simulation can put it in.
The check that rejects an out-of-range state SHALL be expressed against the
number of states, not against any one named state, so that appending a state
cannot leave the check behind.

#### Scenario: Every state round-trips

- **WHEN** a snapshot carrying a fighter in any state the simulation defines is
  encoded and decoded
- **THEN** the snapshot SHALL be accepted and the fighter SHALL come back in the
  state it went in

#### Scenario: A state that does not exist is refused

- **WHEN** a snapshot carries a state value beyond the last state the simulation
  defines
- **THEN** the snapshot SHALL be rejected as malformed

#### Scenario: Appending a state does not break the check

- **WHEN** a new state is appended to the end of the state enumeration
- **THEN** a snapshot carrying a fighter in that state SHALL be accepted without
  any further change to the decoder

## MODIFIED Requirements

### Requirement: Robust packet handling
Malformed, truncated or unknown packets SHALL be ignored without crashing the server or client, and without affecting other players. A packet SHALL be judged malformed only on what the protocol actually disallows: a validity check SHALL NOT reject a value the simulation can legitimately produce.

#### Scenario: Garbage packet
- **WHEN** a peer sends a packet with an unknown message type or an invalid length
- **THEN** the packet is dropped and the match continues

#### Scenario: A valid snapshot is not dropped
- **WHEN** a snapshot carries only values the simulation can produce
- **THEN** the snapshot SHALL be accepted, and SHALL NOT be discarded on account of any fighter in it
