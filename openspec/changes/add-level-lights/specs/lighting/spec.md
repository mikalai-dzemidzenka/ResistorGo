## Purpose

Makes a level's lights part of the level: each light is placed data with a position, radius, color and intensity, saved and loaded with the level, drawn with the shadows its tiles cast, and kept out of the level fingerprint so lighting never decides who can play together.

## ADDED Requirements

### Requirement: The level carries its lights
A level SHALL hold any number of lights, including none. Each light SHALL have a position in whole level pixels, a radius, a color and an intensity. The lights SHALL be saved in the level file and loaded with it, in the development build, the release build and the dedicated server alike. A level file with no lights in it SHALL load with no lights. Nothing but the level SHALL place a light.

The level SHALL carry the three lights the game has always drawn, with the positions, radii, color and intensity they have had, so that the level looks as it did.

#### Scenario: Saved and loaded
- **WHEN** a level with lights is saved and loaded again
- **THEN** it has the same lights, each with the same position, radius, color and intensity

#### Scenario: A level without lights
- **WHEN** a level file with no lights in it is loaded
- **THEN** the level loads with all its tiles and no lights

#### Scenario: The level looks as it did
- **WHEN** the game is played after this change, from the same camera position as before it
- **THEN** the level is lit by the same three lights, in the same places, as before

#### Scenario: The dedicated server
- **WHEN** a dedicated server starts with a level that has lights
- **THEN** it serves the level, and players who join see its lights

### Requirement: Every light in the level is drawn
Every light in the level SHALL be drawn every frame the level is on screen, lighting the scene in its color at its intensity, and shadowed by the tiles' light rectangles as the game's lights are today. A light whose area is entirely off screen SHALL NOT cost any drawing; a light whose area is partly on screen SHALL be drawn in full, so that no light pops as the camera reaches it. The memory lighting uses SHALL NOT grow with the number of lights in the level.

#### Scenario: Color and intensity
- **WHEN** a level has one red light at full intensity and one white light at half intensity
- **THEN** the first lights the scene red, and the second lights it white and dimmer

#### Scenario: Shadows
- **WHEN** a tile with a light rectangle stands between a light and the floor beyond it
- **THEN** the floor beyond it is in shadow

#### Scenario: At the edge of the screen
- **WHEN** the camera moves until only the edge of a light's area is on screen
- **THEN** that edge is lit, and the light does not appear or disappear as the camera crosses the boundary

#### Scenario: Many lights
- **WHEN** a level has twenty lights, most of them far off screen
- **THEN** the game draws the ones on screen and runs as it does with the lights on screen alone

### Requirement: Lights do not change the level fingerprint
The level fingerprint SHALL be computed from the level's tiles alone. Adding, removing or changing a light SHALL NOT change it, and a level that gains lights SHALL keep the fingerprint it had without them.

#### Scenario: A changed light
- **WHEN** a development build with a light moved in its level joins a server whose level differs from it only in that light
- **THEN** it is not refused with "level mismatch"

#### Scenario: The level keeps its fingerprint
- **WHEN** the level's three lights are added to it
- **THEN** its fingerprint is the same as before they were added
