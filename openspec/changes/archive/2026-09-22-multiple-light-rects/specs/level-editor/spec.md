## MODIFIED Requirements

### Requirement: Rectangles are shaped on screen
With the rect tool, every collision rectangle and every light rectangle of the selected thing SHALL be drawn over the level, collision distinguishable from light, with exactly one of them active for editing. Clicking any of those rectangles SHALL make it the active one, and a drag begun on it SHALL shape it. The developer SHALL also be able to switch between the collision rectangles and the light rectangles, and to step from one rectangle of the active kind to the next, without losing the selection. The editor SHALL show which rectangle of the active kind is active, and how many of that kind the selection has.

Dragging an edge or a corner of the active rectangle SHALL resize it from that edge or corner. Dragging its body SHALL move it without changing its size. Only the active rectangle SHALL change. Both SHALL work in whole level pixels, and SHALL work at every camera zoom level. A rectangle's width and height SHALL never be dragged below one pixel. Releasing the mouse button SHALL end the drag. Numeric fields SHALL show the active rectangle's position and size and SHALL set them exactly.

An edit SHALL take effect at once: a reshaped light rectangle SHALL change the level's lighting on the next frame drawn, and a reshaped collision rectangle SHALL be what fighters collide with when the match resumes.

#### Scenario: Resize by a corner
- **WHEN** the developer drags the bottom-right corner of the active collision rectangle two pixels to the right
- **THEN** that rectangle is two pixels wider and its top-left corner has not moved

#### Scenario: Move the body
- **WHEN** the developer drags the middle of the active rectangle
- **THEN** it moves with the cursor and keeps its width and height

#### Scenario: Pick another collision rectangle
- **WHEN** a tile with three collision rectangles is selected and the developer clicks the third one
- **THEN** the third one becomes the active rectangle, and the editor says it is the third of three

#### Scenario: Pick another light rectangle
- **WHEN** a tile with two light rectangles is selected and the developer clicks the second one
- **THEN** the second light rectangle becomes the active rectangle, and the editor says it is the second of two

#### Scenario: Only the active one moves
- **WHEN** the developer drags the body of one of a tile's two light rectangles
- **THEN** that rectangle moves and the other stays where it was

#### Scenario: Step through
- **WHEN** the developer, with a tile's first collision rectangle active, steps to the next collision rectangle twice on a tile that has two
- **THEN** the second is active after the first step and the first again after the second

#### Scenario: Step through the light rectangles
- **WHEN** the developer, with a collision rectangle active on a tile with two light rectangles, switches to the light rectangles and then steps twice
- **THEN** the first light rectangle is active after the switch, the second after the first step, and the first again after the second

#### Scenario: Exact values
- **WHEN** the developer types a width into the numeric field for the active rectangle
- **THEN** the rectangle is exactly that wide

#### Scenario: Never inverted
- **WHEN** the developer drags an edge of the active rectangle past its opposite edge
- **THEN** the rectangle is one pixel in that dimension and is not inverted

#### Scenario: Zoomed in
- **WHEN** the developer zooms the camera in or out and drags a corner
- **THEN** the corner follows the cursor as it does at the default zoom

#### Scenario: The light follows
- **WHEN** the developer reshapes one of a placed tile's light rectangles
- **THEN** the shadow that tile casts changes as the rectangle is dragged

#### Scenario: The collision follows
- **WHEN** the developer reshapes a placed tile's collision rectangle and leaves edit mode
- **THEN** the fighter collides with the new rectangle

### Requirement: A rectangle can be added and cleared
The selected thing SHALL be able to have any number of collision rectangles and any number of light rectangles. The developer SHALL be able to add another rectangle of either kind at any time, and to remove the active rectangle, leaving every other rectangle as it was. A newly added rectangle SHALL start as one grid cell at the object's origin, so it is visible and can be dragged from there, and SHALL become the active one. An object with no collision rectangles SHALL NOT collide with fighters, and an object with no light rectangles SHALL NOT cast a shadow.

#### Scenario: Add one
- **WHEN** the developer selects a placed tile that has no light rectangle and adds one
- **THEN** a one-cell light rectangle appears at that tile's origin and can be dragged

#### Scenario: Add a second collision rectangle
- **WHEN** the developer selects a tile with one collision rectangle and adds a collision rectangle
- **THEN** the tile has two, the new one a grid cell at its origin and active, and the first unchanged

#### Scenario: Add a second light rectangle
- **WHEN** the developer selects a tile with one light rectangle and, with the light rectangles active, adds one
- **THEN** the tile has two light rectangles, the new one a grid cell at its origin and active, and the first unchanged

#### Scenario: Remove one of several
- **WHEN** the developer removes the active one of a tile's three collision rectangles
- **THEN** the tile has the other two, unchanged, and one of them is active

#### Scenario: Remove one of several light rectangles
- **WHEN** the developer removes the active one of a tile's two light rectangles
- **THEN** the tile has the other one, unchanged and active, and still casts its shadow

#### Scenario: Clear one
- **WHEN** the developer removes a placed tile's only collision rectangle and leaves edit mode
- **THEN** a fighter passes through that tile

#### Scenario: No shadow
- **WHEN** the developer removes a placed tile's only light rectangle
- **THEN** that tile stops casting a shadow

### Requirement: Every collision rectangle travels together
Whatever carries an object's rectangles SHALL carry all of them, every collision rectangle and every light rectangle, each kind in the same order:

- placing a tile from a palette entry
- a palette entry's saved defaults
- a placed tile following its texture, and the check of whether it still does
- copy and paste
- saving and loading the level

A tile SHALL follow its texture only while all of its collision rectangles and all of its light rectangles are exactly its texture's.

#### Scenario: Placed with all of them
- **WHEN** the developer gives a texture three collision rectangles and two light rectangles and places a tile of it
- **THEN** the tile has the same three collision rectangles and the same two light rectangles

#### Scenario: Copied with all of them
- **WHEN** the developer copies from a texture with two light rectangles and pastes onto a texture with one
- **THEN** the second texture has exactly the two

#### Scenario: Followed with all of them
- **WHEN** the developer adds a second light rectangle to a texture whose tiles follow it
- **THEN** each of those tiles has both light rectangles and casts a shadow from both

#### Scenario: No longer following
- **WHEN** the developer adds a light rectangle to one placed tile that follows its texture
- **THEN** that tile no longer follows its texture, and the texture's other tiles are unchanged

#### Scenario: Saved with all of them
- **WHEN** the developer gives a tile three collision rectangles and two light rectangles, saves, and restarts
- **THEN** the tile has the same three collision rectangles and the same two light rectangles

### Requirement: The level file keeps its shape
An object with no collision rectangle or one, and no light rectangle or one, SHALL be saved in the level file exactly as it was before objects could have several of either. Only the kind of which an object has two or more SHALL be saved differently. The texture defaults file SHALL follow the same rule. A level in which no object has more than one rectangle of either kind SHALL therefore keep its fingerprint.

#### Scenario: The existing level
- **WHEN** a development build loads a level in which no tile has more than one light rectangle, and saves it without an edit
- **THEN** the level file is byte-for-byte what it was, and its fingerprint is unchanged

#### Scenario: Only light rectangles changed shape
- **WHEN** a tile with one collision rectangle is given a second light rectangle and the level is saved
- **THEN** that tile's collision rectangle is saved exactly as before, and only its light rectangles are saved as a list

## ADDED Requirements

### Requirement: An object casts a shadow from every light rectangle
Every light rectangle of every object in the level SHALL cast a shadow from every light, each behaving on its own exactly as an object's single light rectangle does. Light SHALL pass through the space between an object's light rectangles. This SHALL hold in the development build and the release build alike, and the debug overlay SHALL draw every light rectangle of every object.

#### Scenario: A gap lets light through
- **WHEN** a tile has two light rectangles one above the other with a gap between them, and a light stands on one side of the tile
- **THEN** the floor beyond the tile is shadowed behind each rectangle and lit behind the gap

#### Scenario: The overlay shows them all
- **WHEN** the developer turns on the debug overlay over a tile with two light rectangles
- **THEN** both light rectangles are outlined
