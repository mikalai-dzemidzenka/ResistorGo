# attack-target-outline Specification

## Purpose
Shows the local player where their next swing goes and whether it reaches: the fighter it would strike, marked across the band of its body the current attack height would land on.

## Requirements

### Requirement: The outline traces the body across the band the swing would strike
A marked fighter SHALL be outlined along the edge of its own sprite — its silhouette, not a box around it — across the horizontal band that the local player's swing would land on at its current attack height. Rows of the fighter outside that band SHALL carry no outline, so the mark reads as the slice of that fighter the swing would take and follows the shape of what it would hit.

The outline SHALL be drawn outside the fighter's own pixels, so the sprite stays visible and entirely unaltered: no pixel the fighter draws SHALL be recoloured, dimmed or covered by the mark.

The outline SHALL follow the sprite the fighter is actually showing, changing with its animation from frame to frame.

#### Scenario: The band sits at the attack's height
- **WHEN** the local player's attack height is mid and a fighter is marked
- **THEN** the outlined band sits at the height a mid attack lands at, below the band a high attack would mark and above the band a low attack would mark

#### Scenario: The fighter is not covered
- **WHEN** a fighter is marked
- **THEN** its sprite is still drawn in full, in its own colours, with none of its own pixels replaced by the mark

#### Scenario: It follows the body, not a box
- **WHEN** a marked fighter's body is narrower at one row of the band than another
- **THEN** the outline is correspondingly narrower at that row, tracing the sprite rather than a rectangle around it

#### Scenario: Nothing outside the band
- **WHEN** a fighter is marked at one height
- **THEN** the parts of its body above and below that band carry no outline

#### Scenario: Moving with the animation
- **WHEN** a marked fighter's animation advances to a frame with a different pose
- **THEN** the outline traces the new pose

### Requirement: The outline is red and fades across the band
The outline SHALL be red. It SHALL be drawn at full strength along the middle of the band — the row the swing lands on — and SHALL fade out towards the band's top and bottom edges, reaching nothing at each, so the mark has no hard edge where the band stops and its brightest row is the one that would be hit.

The fade SHALL stay centred on the row the swing lands on even where the band runs past the top or bottom of the fighter's sprite, so a mark clipped by the edge of the body is not a mark whose strongest row has moved.

#### Scenario: Brightest where it lands
- **WHEN** a fighter is marked
- **THEN** the outline is strongest along the middle of the band and weakest at its top and bottom edges

#### Scenario: No hard edge
- **WHEN** a fighter is marked
- **THEN** the outline fades to nothing at the band's edges rather than stopping at full strength

#### Scenario: Clipped by the body
- **WHEN** the band runs past the top or the bottom of the marked fighter's sprite
- **THEN** the part of the outline that is still on the body is drawn at the strength that row calls for, unmoved by the clipping

### Requirement: The outline follows the attack height
The outlined band SHALL follow the local player's current attack height as it changes, without waiting for a swing to be thrown. Moving the aim from one height to another SHALL move the band on the next rendered frame.

Guard height 3 (block-all) SHALL mark nothing: it has no attack of its own, so while the local player holds it no fighter SHALL be outlined.

#### Scenario: Aiming up the body
- **WHEN** the local player moves its aim from low to high with a fighter in range at every height
- **THEN** the outlined band moves up that fighter's body, low then mid then high

#### Scenario: Nothing to aim with
- **WHEN** the local player guards with block-all
- **THEN** no fighter is outlined, however close any of them are

#### Scenario: Back from block-all
- **WHEN** the local player leaves block-all for an attack height with a fighter in range
- **THEN** that fighter is outlined again at the new height

### Requirement: Only a fighter in range is marked
A fighter SHALL be outlined only while the local player's swing at its current attack height would reach it. A fighter the swing would not reach SHALL carry no outline, so the mark appearing tells the player the swing connects from where they stand.

Range SHALL be judged by the same geometry the attack itself uses, so a marked fighter is one the swing would actually reach and an unmarked one is not.

#### Scenario: Walking into range
- **WHEN** the local player walks toward a fighter until its swing would reach
- **THEN** that fighter is outlined from the moment it would be reached, and not before

#### Scenario: Walking out of range
- **WHEN** a marked fighter moves away until the swing no longer reaches
- **THEN** its outline disappears

#### Scenario: Height changes what is in range
- **WHEN** a fighter is within reach at one attack height but not at another, because of where it stands
- **THEN** it is outlined at the height that reaches it and not at the height that does not

#### Scenario: Facing away
- **WHEN** a fighter stands behind the local player, who is facing the other way
- **THEN** it is not outlined, because the swing goes the other way

#### Scenario: Two in range
- **WHEN** the swing at the current height would reach two fighters at once
- **THEN** both are outlined

### Requirement: The outline says nothing about the target's defence
The outline SHALL be the same whatever the marked fighter is doing. It SHALL NOT change with that fighter's guard height, whether the swing would be blocked, its structure, its energy, or how many hits it has taken. Whether an attack lands SHALL remain something the player reads from the fighter itself.

#### Scenario: A guard that would block it
- **WHEN** the marked fighter's guard is at the height the local player is aiming at, so the swing would be blocked
- **THEN** its outline is exactly as it is when the guard is at a different height

#### Scenario: Block-all on the target
- **WHEN** the marked fighter guards with block-all
- **THEN** its outline is unchanged

#### Scenario: A hurt target
- **WHEN** the marked fighter has lost structure and taken hits toward a stun
- **THEN** its outline is unchanged, and its structure bar and pips are drawn as they always are

### Requirement: The outline belongs to the fighter and the scene
The outline SHALL be drawn at the marked fighter's position and move with it, including a remote fighter drawn from interpolated snapshots. It SHALL be drawn at the scene's pixel scale, so it zooms with the camera and stays on the pixel grid, and its colour SHALL NOT change with the scene's lighting.

It SHALL NOT be drawn for a fighter in the Death state, nor for any fighter while time manipulation is showing a recorded frame.

#### Scenario: Follows a moving target
- **WHEN** a marked fighter runs while staying in range
- **THEN** its outline stays on it for the whole run

#### Scenario: Zooms with the camera
- **WHEN** the camera zooms in by one step
- **THEN** the outline grows with the fighter, and its edges stay on the pixel grid

#### Scenario: Unaffected by lighting
- **WHEN** one marked fighter stands under a lamp and another in shadow
- **THEN** both outlines are drawn in identical colours

#### Scenario: A dying target
- **WHEN** a marked fighter enters the Death state
- **THEN** its outline is gone while it falls and flickers

#### Scenario: Scrubbing time
- **WHEN** time manipulation is showing a recorded frame
- **THEN** no fighter is outlined, and the outline returns when the mode is left

### Requirement: The outline is the local player's alone
The outline SHALL be drawn only for the fighter the player at this screen controls, and SHALL mark other fighters, never that fighter itself. A viewer with no fighter of their own — a spectator — SHALL see no outlines, and a networked client SHALL be shown its own fighter's reach rather than anyone else's.

#### Scenario: Spectating
- **WHEN** a spectator watches a match in which two fighters are trading blows in range of each other
- **THEN** no fighter is outlined

#### Scenario: Never itself
- **WHEN** the local player's own fighter is drawn
- **THEN** it carries no outline

#### Scenario: Each client sees its own reach
- **WHEN** two players in a networked match stand in range of each other
- **THEN** each sees the other outlined and neither sees its own fighter outlined

#### Scenario: No fighter to aim with
- **WHEN** the local player's fighter has died and has not yet respawned
- **THEN** no fighter is outlined
