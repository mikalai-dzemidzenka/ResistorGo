## ADDED Requirements

### Requirement: Rewound shot resolution
When executing a player's command that fires a shot, the server SHALL test the shot against all other fighters as they were at the command's view tick. It SHALL use their recorded position and collision box, the same way a rewound swing is tested, within the same compensation limit. Level walls SHALL be tested as they are, since they do not move. The shooter SHALL fire from its current position with the command's aim. Whether a target can be hit at all SHALL use the target's current state, and the resulting hit SHALL be applied to the target's current state. When lag compensation is disabled, shots SHALL be tested against the current server state.

#### Scenario: Shot on what the shooter saw
- **WHEN** a shooter at 150 ms round-trip time fires at an opponent it sees in its aim, while on the server the opponent has since stepped out of it
- **THEN** the server registers the hit, because the opponent was in the aim at the shooter's view tick

#### Scenario: Targets are restored after a shot
- **WHEN** the server finishes resolving a rewound shot
- **THEN** every other fighter's current position and state are exactly as before the test, apart from the hit applied to the target
