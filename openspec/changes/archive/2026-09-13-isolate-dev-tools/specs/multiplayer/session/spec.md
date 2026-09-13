## MODIFIED Requirements

### Requirement: Offline-only features are disabled online
While in a networked match (host or client), time manipulation, the level editor and writing the level file SHALL be disabled. In a development build, debug overlays SHALL remain available; a release build has none.

#### Scenario: Rewind key online
- **WHEN** a player in a networked match presses R
- **THEN** nothing is paused or rewound and gameplay continues normally

#### Scenario: Level not saved
- **WHEN** a player exits the game after a networked match
- **THEN** the level file on disk is unchanged

#### Scenario: Overlay online in a development build
- **WHEN** a player of a development build toggles the debug overlay in a networked match
- **THEN** the overlay is shown and the match is unaffected
