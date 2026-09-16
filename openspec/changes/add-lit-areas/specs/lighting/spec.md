## ADDED Requirements

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
