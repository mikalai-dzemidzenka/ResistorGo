## ADDED Requirements

### Requirement: The scene is lit by a light map
The frame SHALL be lit by one light map: the level's ambient light with the light of every visible level light added into it. The scene SHALL be modulated by that map — an unlit pixel SHALL be what the scene draws there times the ambient, and a lit one what the scene draws there times the ambient plus the light reaching it — so that a shadow keeps the colour of what stands in it and no light can turn a black pixel white. A weak addition of the same map SHALL keep a light readable over empty space and let a near-black surface catch something. Nothing SHALL draw flat darkness over the frame.

What is drawn over the lit scene — the attack outlines, the structure bars, the energy bars, the stun pips and the editor's overlay — SHALL NOT be lit or darkened by the light map.

#### Scenario: A shadow keeps its colour
- **WHEN** a coloured tile stands in the shadow another tile casts
- **THEN** it is a darker shade of its own colour, not a grey wash over it

#### Scenario: Light does not make black white
- **WHEN** a bright light is placed against a black background with one white tile in reach of it
- **THEN** the tile is lit to white and the background beside it stays near black

#### Scenario: A light in the air
- **WHEN** a light is placed in open space with no tile in reach of it
- **THEN** its glow is visible where it hangs, dimmer than the light it casts on a surface

#### Scenario: Darkness deepens with the shadows
- **WHEN** the level is played after this change, from the same camera position as before it
- **THEN** the parts no light reaches are darker than they were, and the shadows the tiles cast are read as shadows rather than as haze

#### Scenario: The overlays keep their colours
- **WHEN** a fighter with a structure bar and an attack outline stands in an unlit part of the level
- **THEN** the bar and the outline are drawn in their own colours, as bright as they are in a lit part

### Requirement: The level carries its ambient light
A level SHALL hold an ambient light: a colour and a strength. It SHALL be saved in the level file and loaded with it, in the development build, the release build and the dedicated server alike. A level file with no ambient in it SHALL load with a default ambient. The ambient SHALL be what every unlit part of the scene is lit by, and a change to it SHALL show in the level on the next frame drawn. The ambient SHALL NOT change the level fingerprint: a level that gains an ambient, and a level whose ambient is changed, SHALL keep the fingerprint it had.

#### Scenario: Saved and loaded
- **WHEN** a level with an ambient is saved and loaded again
- **THEN** it has the same ambient, the same colour and the same strength

#### Scenario: A level without an ambient
- **WHEN** a level file with no ambient in it is loaded
- **THEN** the level loads with all its tiles and lights, and is lit by the default ambient

#### Scenario: A coloured night
- **WHEN** a level's ambient is set to a deep blue
- **THEN** every part of it no light reaches is tinted that blue, and the lit parts are unchanged in the middle of each light

#### Scenario: No ambient at all
- **WHEN** a level's ambient strength is set to zero
- **THEN** everything outside the reach of a light is black

#### Scenario: A changed ambient
- **WHEN** a development build with its level's ambient changed joins a server whose level differs from it only in that ambient
- **THEN** it is not refused with "level mismatch"

## MODIFIED Requirements

### Requirement: Every light in the level is drawn
Every light in the level SHALL be drawn every frame the level is on screen, adding its light in its colour at its intensity into the frame's light map, and shadowed by the tiles' light rectangles as the game's lights are today. A light whose area is entirely off screen SHALL NOT cost any drawing; a light whose area is partly on screen SHALL be drawn in full, so that no light pops as the camera reaches it. The memory lighting uses SHALL NOT grow with the number of lights in the level.

#### Scenario: Color and intensity
- **WHEN** a level has one red light at full intensity and one white light at half intensity
- **THEN** the first lights the scene red, and the second lights it white and dimmer

#### Scenario: Shadows
- **WHEN** a tile with a light rectangle stands between a light and the floor beyond it
- **THEN** the floor beyond it is in shadow

#### Scenario: Two lights on one spot
- **WHEN** two lights reach the same piece of floor
- **THEN** it is lit by both together, no dimmer than either alone lights it

#### Scenario: At the edge of the screen
- **WHEN** the camera moves until only the edge of a light's area is on screen
- **THEN** that edge is lit, and the light does not appear or disappear as the camera crosses the boundary

#### Scenario: Many lights
- **WHEN** a level has twenty lights, most of them far off screen
- **THEN** the game draws the ones on screen and runs as it does with the lights on screen alone
