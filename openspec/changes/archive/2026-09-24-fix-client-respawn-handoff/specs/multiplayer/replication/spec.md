## ADDED Requirements

### Requirement: A snapshot names the receiving player's fighter
Every snapshot SHALL tell the client which fighter is its own: the id of the receiving player's current fighter, or none when that player has no fighter. It SHALL be carried in the snapshot itself rather than announced once, so a client that loses the packet carrying a change is told again by the next one, and a client that connects while its player has no fighter is told the truth at once.

A client SHALL take the fighter it is named as its own, replacing whatever it held before, and SHALL predict and control that fighter and no other. A client named no fighter SHALL control none.

#### Scenario: A new fighter after a death
- **WHEN** a player's fighter is replaced on the server by a respawn
- **THEN** the next snapshot names the new fighter, and the client predicts and controls that fighter from then on

#### Scenario: Named none while dead
- **WHEN** a player's fighter has been removed and its respawn has not come yet
- **THEN** every snapshot in that gap names no fighter for that player

#### Scenario: A lost packet costs nothing
- **WHEN** the first snapshot naming a player's new fighter is dropped in the network
- **THEN** the next snapshot names it again and the client picks it up there, without any other message being needed

#### Scenario: Each player is told about its own
- **WHEN** two players are in a match and one of them respawns
- **THEN** the snapshot sent to each names that player's own fighter, and neither is told the other's is theirs

#### Scenario: Still fits the packet
- **WHEN** a snapshot carries 16 fighters and names the receiving player's own
- **THEN** the packet is at most 1200 bytes
