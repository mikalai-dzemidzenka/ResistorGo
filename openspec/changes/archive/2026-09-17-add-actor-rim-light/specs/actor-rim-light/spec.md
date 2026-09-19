## Purpose

Lights the edges of a fighter's silhouette that face a nearby level light, in that light's colour, so fighters stand out from the level and you can tell which way the light falls on them.

## ADDED Requirements

### Requirement: A level light rims the fighter's outside edges that face it
A fighter within a level light's radius SHALL get a rim light from that light: a line drawn in the light's colour on the empty pixels immediately outside the fighter, along the sides of it that face the light. A rim pixel is a transparent pixel touching the fighter's silhouette.

A rim pixel SHALL face the light when the direction from the silhouette out to that pixel points towards the light. It SHALL be lit more strongly the more directly it faces the light, and not at all where it faces away. On the sides facing directly away from the light, nothing SHALL be drawn.

The rim SHALL be one pixel wide. It SHALL NOT recolour, dim or cover any pixel the fighter itself draws, so the sprite stays exactly as it was drawn and only the pixels around it change. It SHALL be drawn on every side of the fighter that faces the light, including where the fighter's own sprite reaches the edge of its animation frame.

#### Scenario: Light from one side
- **WHEN** a fighter stands with a level light level with it on its left
- **THEN** a rim in the light's colour runs down the outside of its left side, and its right side carries none

#### Scenario: Light from above
- **WHEN** a level light is directly above a fighter
- **THEN** the rim runs along the outside of its head and shoulders and not under its body

#### Scenario: The fighter is not covered
- **WHEN** a fighter is rimmed
- **THEN** its sprite is still drawn in full, in its own colours, with none of its own pixels replaced by the rim

#### Scenario: Against the frame's edge
- **WHEN** a rimmed fighter's sprite reaches the edge of its animation frame on a side that faces the light
- **THEN** the rim is still drawn along that side rather than being cut off there

### Requirement: The rim is as strong as the light at the fighter
A light's rim SHALL use that light's colour, scaled by its intensity and by the same falloff with distance that the light map uses for that light. The rim therefore SHALL fade as the fighter moves away from the light and SHALL vanish at or beyond the light's radius.

A light whose intensity is zero SHALL give no rim.

#### Scenario: Fading with distance
- **WHEN** a fighter walks away from a level light, staying inside its radius
- **THEN** the rim from that light gets steadily weaker

#### Scenario: Out of reach
- **WHEN** a fighter is entirely outside a level light's radius
- **THEN** that light gives the fighter no rim

#### Scenario: Coloured light
- **WHEN** a fighter is rimmed by a light whose colour is cyan
- **THEN** the rim is cyan, not white

### Requirement: A shadowed light gives no rim
A level light SHALL NOT rim a fighter when a tile's light rectangle lies between the light and the fighter's body, so a fighter standing in a light's shadow does not get that light's rim.

#### Scenario: Behind a wall
- **WHEN** a wall with a light rectangle stands between a level light and a fighter within that light's radius
- **THEN** that light gives the fighter no rim

#### Scenario: Clear line
- **WHEN** nothing stands between a level light and a fighter within its radius
- **THEN** that light rims the fighter

### Requirement: Several lights rim a fighter at once
Every level light that reaches a fighter SHALL add its own rim on the edges that face it, in its own colour, up to a fixed number of lights per fighter. When more lights reach a fighter than that number, the ones that are strongest at the fighter SHALL be used. Where two rims light the same pixel, their light SHALL add up.

#### Scenario: Two lamps, two sides
- **WHEN** a fighter stands between a red light on its left and a blue light on its right
- **THEN** its left edges carry a red rim and its right edges carry a blue rim

#### Scenario: Too many lights
- **WHEN** more lights reach a fighter than the per-fighter number
- **THEN** the lights that are strongest at that fighter are the ones that rim it

### Requirement: Only level lights make a rim
The ambient light and lit areas have no direction, so they SHALL NOT make a rim. A fighter lit only by the ambient light or standing inside a lit area SHALL show no rim.

#### Scenario: Inside a lit area
- **WHEN** a fighter stands inside a lit area and within no level light's radius
- **THEN** it has no rim

### Requirement: A fighter drawn in two layers is rimmed as one shape
A fighter may be drawn as a body sprite with a second sprite over it - as a walking fighter is, whose arms are drawn from a second sheet chosen by the guard height it is holding. Where a fighter is drawn this way, the rim SHALL treat the two together as one silhouette: it SHALL run around the outside of their combined shape, and SHALL NOT draw a rim along a seam where the upper sprite lies over the body.

#### Scenario: One outline around both
- **WHEN** a fighter drawn with a second sprite over its body is rimmed
- **THEN** the rim runs around the outside of the two together

#### Scenario: No seam
- **WHEN** the upper sprite covers part of the body sprite
- **THEN** no rim is drawn where they overlap

#### Scenario: Walking with a guard up
- **WHEN** a walking fighter is rimmed and its guard height changes, so its arms are drawn from a different sheet
- **THEN** the rim follows the arms it is now showing

### Requirement: The rim follows what the fighter shows
The rim SHALL trace the sprite frame the fighter is showing on that frame. It SHALL follow the animation frame by frame and the fighter's interpolated draw position. It SHALL stay on the correct side when the fighter faces left, whose sprite is drawn mirrored.

While time shows a recorded frame, the rim SHALL trace the recorded sprite that is shown. While a dead fighter is flickered out of view, it SHALL have no rim.

#### Scenario: Turning around
- **WHEN** a fighter lit from its left turns to face the other way
- **THEN** the rim is still on the edges on the light's side

#### Scenario: Animation
- **WHEN** a rimmed fighter's animation advances to a frame with a different pose
- **THEN** the rim traces the new pose

#### Scenario: Rewind
- **WHEN** time is being rewound and a recorded frame is shown
- **THEN** the rim traces the recorded pose that is on screen

#### Scenario: Death flicker
- **WHEN** a dead fighter's flicker hides it on a frame
- **THEN** no rim is drawn for it on that frame

### Requirement: Every fighter is rimmed for every viewer
The rim is scene lighting, not a cue for the local player. It SHALL apply to every fighter in the level: the local player, remote players and bots. It SHALL look the same offline, on a listen server, on a remote client and for a spectator.

It SHALL NOT change the simulation, the data sent over the network, the level file or the level fingerprint.

#### Scenario: Spectating
- **WHEN** a spectator watches two fighters near a level light
- **THEN** both fighters are rimmed as they would be for a player

#### Scenario: Opponents too
- **WHEN** a remote player or a bot stands near a level light
- **THEN** it is rimmed as the local player would be

### Requirement: The rim is lit light, under the overlays
The rim SHALL show on a fighter in the dark as clearly as on one under a lamp. The level's ambient darkness SHALL NOT dim it. The attack-target outline, the structure bars and the level editor's overlay SHALL be drawn above the rim.

The attack-target outline occupies the same pixels around a fighter as the rim. Where both fall on one pixel, the attack-target outline SHALL take precedence, so a marked fighter still reads as marked.

#### Scenario: Dark surroundings
- **WHEN** a fighter stands in an otherwise dark part of the level at the edge of a light's radius where the light map is dim
- **THEN** the rim on its lit side is still drawn at the strength the light's falloff gives there

#### Scenario: Marked and rimmed
- **WHEN** a fighter is both rimmed and marked by the attack-target outline
- **THEN** the mark is still plainly on it, and the rim shows on the sides the mark does not cover
