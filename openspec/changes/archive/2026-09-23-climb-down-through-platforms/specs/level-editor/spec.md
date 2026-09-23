## MODIFIED Requirements

### Requirement: Rectangles are shaped on screen
With the rect tool, every collision rectangle, every drop-through rectangle and every light rectangle of the selected thing SHALL be drawn over the level, each kind distinguishable from the other two, with exactly one of them active for editing. Clicking any of those rectangles SHALL make it the active one, and a drag begun on it SHALL shape it. The developer SHALL also be able to switch between the collision rectangles, the drop-through rectangles and the light rectangles, and to step from one rectangle of the active kind to the next, without losing the selection. The editor SHALL show which rectangle of the active kind is active, and how many of that kind the selection has.

Dragging an edge or a corner of the active rectangle SHALL resize it from that edge or corner. Dragging its body SHALL move it without changing its size. Only the active rectangle SHALL change. Both SHALL work in whole level pixels, and SHALL work at every camera zoom level. A rectangle's width and height SHALL never be dragged below one pixel. Releasing the mouse button SHALL end the drag. Numeric fields SHALL show the active rectangle's position and size and SHALL set them exactly.

An edit SHALL take effect at once: a reshaped light rectangle SHALL change the level's lighting on the next frame drawn, and a reshaped collision or drop-through rectangle SHALL be what fighters collide with when the match resumes.

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

#### Scenario: Switch to the drop-through rectangles
- **WHEN** the developer, with a collision rectangle active on a tile with one collision rectangle and one drop-through rectangle, switches to the drop-through rectangles
- **THEN** the drop-through rectangle is active, the editor says it is the first of one drop-through rectangle, and it is drawn unlike the collision rectangle

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

#### Scenario: The drop-through follows
- **WHEN** the developer widens a placed tile's drop-through rectangle and leaves edit mode
- **THEN** the fighter can stand on, and climb down through, the whole of the widened rectangle

### Requirement: A rectangle can be added and cleared
The selected thing SHALL be able to have any number of collision rectangles, any number of drop-through rectangles and any number of light rectangles. The developer SHALL be able to add another rectangle of any of the three kinds at any time, and to remove the active rectangle, leaving every other rectangle as it was. A newly added rectangle SHALL start as one grid cell at the object's origin, so it is visible and can be dragged from there, and SHALL become the active one. An object with no collision rectangles and no drop-through rectangles SHALL NOT collide with fighters, and an object with no light rectangles SHALL NOT cast a shadow.

#### Scenario: Add one
- **WHEN** the developer selects a placed tile that has no light rectangle and adds one
- **THEN** a one-cell light rectangle appears at that tile's origin and can be dragged

#### Scenario: Add a second collision rectangle
- **WHEN** the developer selects a tile with one collision rectangle and adds a collision rectangle
- **THEN** the tile has two, the new one a grid cell at its origin and active, and the first unchanged

#### Scenario: Add a second light rectangle
- **WHEN** the developer selects a tile with one light rectangle and, with the light rectangles active, adds one
- **THEN** the tile has two light rectangles, the new one a grid cell at its origin and active, and the first unchanged

#### Scenario: Add a drop-through rectangle
- **WHEN** the developer selects a tile with one collision rectangle and, with the drop-through rectangles active, adds one
- **THEN** the tile has one drop-through rectangle, a grid cell at its origin and active, and its collision rectangle unchanged

#### Scenario: Remove one of several
- **WHEN** the developer removes the active one of a tile's three collision rectangles
- **THEN** the tile has the other two, unchanged, and one of them is active

#### Scenario: Remove one of several light rectangles
- **WHEN** the developer removes the active one of a tile's two light rectangles
- **THEN** the tile has the other one, unchanged and active, and still casts its shadow

#### Scenario: Clear one
- **WHEN** the developer removes a placed tile's only collision rectangle and leaves edit mode
- **THEN** a fighter passes through that tile

#### Scenario: Clear the drop-through
- **WHEN** the developer removes a placed tile's only drop-through rectangle, the tile has no collision rectangle, and the developer leaves edit mode
- **THEN** a fighter passes through that tile

#### Scenario: No shadow
- **WHEN** the developer removes a placed tile's only light rectangle
- **THEN** that tile stops casting a shadow

### Requirement: Every collision rectangle travels together
Whatever carries an object's rectangles SHALL carry all of them, every collision rectangle, every drop-through rectangle and every light rectangle, each kind in the same order:

- placing a tile from a palette entry
- a palette entry's saved defaults
- a placed tile following its texture, and the check of whether it still does
- copy and paste
- saving and loading the level

A tile SHALL follow its texture only while all of its collision rectangles, all of its drop-through rectangles and all of its light rectangles are exactly its texture's.

#### Scenario: Placed with all of them
- **WHEN** the developer gives a texture three collision rectangles and two light rectangles and places a tile of it
- **THEN** the tile has the same three collision rectangles and the same two light rectangles

#### Scenario: Placed with its drop-through
- **WHEN** the developer gives a texture one drop-through rectangle and places a tile of it
- **THEN** the tile has the same drop-through rectangle, and a fighter can climb down through it

#### Scenario: Copied with all of them
- **WHEN** the developer copies from a texture with two light rectangles and pastes onto a texture with one
- **THEN** the second texture has exactly the two

#### Scenario: Copied with the drop-through
- **WHEN** the developer copies from a texture with one drop-through rectangle and pastes onto a texture with none
- **THEN** the second texture has exactly that drop-through rectangle

#### Scenario: Followed with all of them
- **WHEN** the developer adds a second light rectangle to a texture whose tiles follow it
- **THEN** each of those tiles has both light rectangles and casts a shadow from both

#### Scenario: No longer following
- **WHEN** the developer adds a light rectangle to one placed tile that follows its texture
- **THEN** that tile no longer follows its texture, and the texture's other tiles are unchanged

#### Scenario: Saved with all of them
- **WHEN** the developer gives a tile three collision rectangles, one drop-through rectangle and two light rectangles, saves, and restarts
- **THEN** the tile has the same three collision rectangles, the same drop-through rectangle and the same two light rectangles

### Requirement: The level file keeps its shape
An object with no collision rectangle or one, no light rectangle or one, and no drop-through rectangle SHALL be saved in the level file exactly as it was before objects could have several of either or any drop-through rectangle. Only a kind of which an object has two or more SHALL be saved as a list, and drop-through rectangles SHALL be saved under keys of their own, one or a list by the same rule, and only when an object has some. The texture defaults file SHALL follow the same rule. A level in which no object has more than one rectangle of either of the first two kinds, and none has a drop-through rectangle, SHALL therefore keep its fingerprint.

#### Scenario: The existing level
- **WHEN** a development build loads a level in which no tile has more than one light rectangle, and saves it without an edit
- **THEN** the level file is byte-for-byte what it was, and its fingerprint is unchanged

#### Scenario: Only light rectangles changed shape
- **WHEN** a tile with one collision rectangle is given a second light rectangle and the level is saved
- **THEN** that tile's collision rectangle is saved exactly as before, and only its light rectangles are saved as a list

#### Scenario: A drop-through rectangle is added
- **WHEN** a tile with one collision rectangle is given a drop-through rectangle and the level is saved
- **THEN** that tile's collision rectangle is saved exactly as before, its drop-through rectangle is saved under its own key, and every other tile is saved as before

#### Scenario: Client and server disagree on a drop-through
- **WHEN** a client's level has a drop-through rectangle the server's level does not
- **THEN** the two levels have different fingerprints
