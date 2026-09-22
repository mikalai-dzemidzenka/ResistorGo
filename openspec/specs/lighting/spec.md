# lighting Specification

## Purpose

Makes a level's lights part of the level: each light is placed data with a position, radius, color and intensity, saved and loaded with the level, drawn with the shadows its tiles cast, and kept out of the level fingerprint so lighting never decides who can play together.

## Requirements

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

### Requirement: Lights do not change the level fingerprint
The level fingerprint SHALL be computed from the level's tiles alone. Adding, removing or changing a light SHALL NOT change it, and a level that gains lights SHALL keep the fingerprint it had without them.

#### Scenario: A changed light
- **WHEN** a development build with a light moved in its level joins a server whose level differs from it only in that light
- **THEN** it is not refused with "level mismatch"

#### Scenario: The level keeps its fingerprint
- **WHEN** the level's three lights are added to it
- **THEN** its fingerprint is the same as before they were added

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

### Requirement: A lit area holds part of the level at full brightness
A level SHALL be able to hold lit areas: axis-aligned rectangles, in whole level pixels, inside which the light map is at full brightness. Inside a lit area the scene SHALL be drawn as it is, undarkened — neither the level's ambient nor any shadow cast by a tile's light rectangle SHALL darken what stands there. A lit area SHALL NOT change what is drawn there, only how brightly it is lit, and it SHALL NOT dim anything: a part of the level already lit brighter than the area lifts it SHALL stay that bright.

A lit area SHALL light everything in its rectangle alike — tiles, fighters and anything else the scene draws. Whatever is drawn over the lit scene — the attack outlines, the bars, the editor's overlay — SHALL be no more affected by a lit area than it is by a light.

A lit area SHALL be nothing but light: it SHALL NOT be visible in its own right in play, SHALL NOT block or shape any light, and SHALL NOT affect collision, hits or anything else the match simulates.

#### Scenario: Out of the shadow
- **WHEN** a lit area covers a floor that a tile's light rectangle shadows
- **THEN** that floor is drawn at full brightness, with no shadow across it

#### Scenario: Away from every light
- **WHEN** a lit area covers a corner of the level no light reaches
- **THEN** that corner is drawn at full brightness instead of in the ambient dark

#### Scenario: A fighter walks through
- **WHEN** a fighter walks out of an unlit part of the level into a lit area and on out of it
- **THEN** the fighter is lit while inside the area, in the colours the sprite is drawn in, and is back in the dark once out

#### Scenario: A bright light inside an area
- **WHEN** a light bright enough to wash a wall white stands inside a lit area
- **THEN** the wall is as bright as that light alone makes it, not dimmed by the area

#### Scenario: Nothing to see in play
- **WHEN** a level with lit areas is played outside the editor
- **THEN** no outline, edge or marker of any area is drawn — only the brightness it gives

#### Scenario: No effect on the match
- **WHEN** a fighter runs, jumps and attacks across a lit area
- **THEN** movement, collision and hits are exactly what they are with no area there

### Requirement: A lit area's edge fades over its feather
A lit area SHALL carry a feather: the width in level pixels over which its edge fades from the full brightness inside it to the light the rest of the map gives just outside it. The fade SHALL lie inside the area's rectangle, so an area never lights anything beyond its own bounds, and it SHALL be even along every side and through every corner. A feather of zero SHALL give a hard edge, and a feather wider than half the area SHALL still leave the area brightest at its center rather than darkening it.

#### Scenario: A soft edge
- **WHEN** a lit area with a feather of sixteen pixels lies over an unlit floor
- **THEN** the floor is fully lit in the middle of the area and fades smoothly into the dark over the last sixteen pixels of it, with no visible straight line

#### Scenario: A hard edge
- **WHEN** a lit area's feather is zero
- **THEN** the area's brightness stops exactly at its rectangle

#### Scenario: The corners fade too
- **WHEN** a lit area with a wide feather is viewed at a corner
- **THEN** the corner fades as evenly as the sides do, with no brighter or darker wedge where two sides meet

#### Scenario: A feather wider than the area
- **WHEN** a lit area sixteen pixels across is given a feather of sixty-four
- **THEN** the area is still brightest at its center and still fades to nothing at its bounds

### Requirement: The level carries its lit areas
A level SHALL hold its lit areas — each one's position, size and feather — saved in the level file and loaded with it, in the development build, the release build and the dedicated server alike. A level file with no lit areas in it SHALL load exactly as it does today, with none. A change to an area SHALL show in the level on the next frame drawn. Lit areas SHALL NOT change the level fingerprint: a level that gains a lit area, and a level whose areas are changed, SHALL keep the fingerprint it had.

#### Scenario: Saved and loaded
- **WHEN** a level with lit areas is saved and loaded again
- **THEN** it has the same areas, each at the same position, the same size and the same feather

#### Scenario: A level without lit areas
- **WHEN** a level file with no lit areas in it is loaded
- **THEN** the level loads with all its tiles, lights and ambient, and nothing is lit beyond what they light

#### Scenario: Not a level mismatch
- **WHEN** a development build with a lit area added to its level joins a server whose level differs from it only in that area
- **THEN** it is not refused with "level mismatch"

#### Scenario: The server carries them along
- **WHEN** the dedicated server loads a level with lit areas
- **THEN** it starts and runs the match exactly as it does with the same level without them

### Requirement: Lit areas cost nothing off screen
A lit area whose rectangle is entirely off screen SHALL NOT cost any drawing, and an area partly on screen SHALL be lit in full, so that nothing pops as the camera reaches it. The memory lighting uses SHALL NOT grow with the number of lit areas in the level, and no lit area SHALL add a render target or a shader to the frame.

#### Scenario: At the edge of the screen
- **WHEN** the camera moves until only the edge of a lit area is on screen
- **THEN** that edge is lit as it is from any other camera position, and nothing appears or disappears as the camera crosses the boundary

#### Scenario: Many areas
- **WHEN** a level has twenty lit areas, most of them far off screen
- **THEN** the game draws the ones on screen and runs as it does with the areas on screen alone
