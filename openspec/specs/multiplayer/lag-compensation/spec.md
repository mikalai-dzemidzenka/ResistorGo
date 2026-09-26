# multiplayer/lag-compensation Specification

## Purpose

Makes melee hits and blocks fair for the attacker under latency. The server evaluates each attack against where the attacker saw the other fighters, not where they are on the server right now.

## Requirements

### Requirement: Fighter history
For at least the last 1 second of server ticks, the server SHALL record every fighter's position, collision box, state, guard height and animation progress.

#### Scenario: History depth
- **WHEN** the server has been running for more than 1 second
- **THEN** it can reconstruct any fighter's recorded values for any tick within the last second

### Requirement: Rewound hit resolution
When executing a player's command during which that player's attack can hit, the server SHALL test the attack against all other fighters as they were at the command's view tick. The test SHALL use those fighters' recorded position, collision box and guard height at that tick, interpolated between recorded ticks when the view tick falls between them. The attacking fighter SHALL use its current state. Whether a target can be affected at all SHALL use the target's current state (for example, a fighter already in hit-stun or blocking is not hit again). The resulting hit or block SHALL be applied to the target's current state.

#### Scenario: Hit on what the attacker saw
- **WHEN** an attacker at 150 ms round-trip time sees an opponent inside attack range and swings, while on the server the opponent has since stepped back out of range
- **THEN** the server registers the hit, because the opponent was in range at the attacker's view tick

#### Scenario: Guard height as seen by the attacker
- **WHEN** the attacker swings high while the opponent's guard was low at the attacker's view tick, and the opponent has since switched to high guard on the server
- **THEN** the attack is resolved against the low guard and hits

#### Scenario: Targets are restored
- **WHEN** the server finishes resolving a rewound attack
- **THEN** every other fighter's current position and state are exactly as before the test, apart from the hit or block result applied to the target

### Requirement: Compensation limit
The server SHALL NOT rewind further back than a maximum compensation window (default 250 ms, configurable on the server). A view tick older than the window SHALL be clamped to the window's edge. A view tick newer than the current server tick SHALL be clamped to the current tick.

#### Scenario: Very high latency attacker
- **WHEN** an attacker's view tick is 600 ms behind the server
- **THEN** the attack is tested against fighters as they were 250 ms ago

#### Scenario: Forged future view tick
- **WHEN** a command carries a view tick ahead of the server's tick
- **THEN** the attack is tested against the current server state

### Requirement: Compensation can be disabled
The server SHALL provide a setting that turns lag compensation off, so that attacks are tested against the current server state. In debug mode on a listen server, the rewound collision boxes used for the last attack test SHALL be drawable for inspection.

#### Scenario: Disabled
- **WHEN** lag compensation is turned off and an attacker at 150 ms round-trip time swings at an opponent who has already left range on the server
- **THEN** the attack misses

### Requirement: Rewound shot resolution
When executing a player's command that fires a shot, the server SHALL test the shot against all other fighters as they were at the command's view tick. It SHALL use their recorded position and collision box, the same way a rewound swing is tested, within the same compensation limit. Level walls SHALL be tested as they are, since they do not move. The shooter SHALL fire from its current position with the command's aim. Whether a target can be hit at all SHALL use the target's current state, and the resulting hit SHALL be applied to the target's current state. When lag compensation is disabled, shots SHALL be tested against the current server state.

#### Scenario: Shot on what the shooter saw
- **WHEN** a shooter at 150 ms round-trip time fires at an opponent it sees in its aim, while on the server the opponent has since stepped out of it
- **THEN** the server registers the hit, because the opponent was in the aim at the shooter's view tick

#### Scenario: Targets are restored after a shot
- **WHEN** the server finishes resolving a rewound shot
- **THEN** every other fighter's current position and state are exactly as before the test, apart from the hit applied to the target
