## MODIFIED Requirements

### Requirement: Rectangles are shaped on screen
With the rect tool, every collision rectangle and the light rectangle of the selected thing SHALL be drawn over the level, collision distinguishable from light, with exactly one of them active for editing. Clicking any of those rectangles SHALL make it the active one, and a drag begun on it SHALL shape it. The developer SHALL also be able to switch between the collision rectangles and the light rectangle, and to step from one collision rectangle to the next, without losing the selection. The editor SHALL show which collision rectangle is active, and how many the selection has.

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

#### Scenario: Only the active one moves
- **WHEN** the developer drags the body of one of a tile's two collision rectangles
- **THEN** that rectangle moves and the other stays where it was

#### Scenario: Step through
- **WHEN** the developer, with a tile's first collision rectangle active, steps to the next collision rectangle twice on a tile that has two
- **THEN** the second is active after the first step and the first again after the second

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
- **WHEN** the developer reshapes a placed tile's light rectangle
- **THEN** the shadow that tile casts changes as the rectangle is dragged

#### Scenario: The collision follows
- **WHEN** the developer reshapes a placed tile's collision rectangle and leaves edit mode
- **THEN** the fighter collides with the new rectangle

### Requirement: A rectangle can be added and cleared
The selected thing SHALL be able to have any number of collision rectangles and at most one light rectangle. The developer SHALL be able to add another collision rectangle at any time, to add a light rectangle when there is none, and to remove the active rectangle, leaving every other rectangle as it was. A newly added rectangle SHALL start as one grid cell at the object's origin, so it is visible and can be dragged from there, and SHALL become the active one. An object with no collision rectangles SHALL NOT collide with fighters, and an object with no light rectangle SHALL NOT cast a shadow.

#### Scenario: Add one
- **WHEN** the developer selects a placed tile that has no light rectangle and adds one
- **THEN** a one-cell light rectangle appears at that tile's origin and can be dragged

#### Scenario: Add a second collision rectangle
- **WHEN** the developer selects a tile with one collision rectangle and adds a collision rectangle
- **THEN** the tile has two, the new one a grid cell at its origin and active, and the first unchanged

#### Scenario: Remove one of several
- **WHEN** the developer removes the active one of a tile's three collision rectangles
- **THEN** the tile has the other two, unchanged, and one of them is active

#### Scenario: Clear one
- **WHEN** the developer removes a placed tile's only collision rectangle and leaves edit mode
- **THEN** a fighter passes through that tile

#### Scenario: No shadow
- **WHEN** the developer clears a placed tile's light rectangle
- **THEN** that tile stops casting a shadow

## ADDED Requirements

### Requirement: An object collides wherever any of its collision rectangles is
A fighter SHALL collide with every collision rectangle of every object in the level. Each rectangle SHALL behave on its own exactly as an object's single collision rectangle does: it can be stood on, it blocks a fighter's head from below, and its ledges can be climbed. This SHALL be true in single player and online, for the fighters a player controls, the fighters of bots, and the fighters a client predicts.

#### Scenario: Stand on either
- **WHEN** a tile has two collision rectangles side by side with a gap between them, and a fighter walks across it
- **THEN** the fighter stands on the first, falls into the gap, and can stand on the second

#### Scenario: A ledge on the second
- **WHEN** a fighter jumps up at the edge of a tile's second collision rectangle
- **THEN** it catches the ledge and climbs, as it would at the edge of a tile's only rectangle

#### Scenario: Online
- **WHEN** two players in a networked match both walk over a tile with two collision rectangles
- **THEN** both collide with both rectangles, and neither is corrected by the server for doing so

### Requirement: Every collision rectangle travels together
Whatever carries an object's collision rectangle SHALL carry all of them, in the same order:

- placing a tile from a palette entry
- a palette entry's saved defaults
- a placed tile following its texture, and the check of whether it still does
- copy and paste
- saving and loading the level

A tile SHALL follow its texture only while all of its collision rectangles, and its light rectangle, are exactly its texture's.

#### Scenario: Placed with all of them
- **WHEN** the developer gives a texture three collision rectangles and places a tile of it
- **THEN** the tile has the same three

#### Scenario: Copied with all of them
- **WHEN** the developer copies from a texture with two collision rectangles and pastes onto a texture with one
- **THEN** the second texture has exactly the two

#### Scenario: Followed with all of them
- **WHEN** the developer adds a second collision rectangle to a texture whose tiles follow it
- **THEN** each of those tiles has both rectangles

#### Scenario: Saved with all of them
- **WHEN** the developer gives a tile three collision rectangles, saves, and restarts
- **THEN** the tile has the same three

### Requirement: The level file keeps its shape
An object with no collision rectangle, or with one, SHALL be saved in the level file exactly as it was before objects could have several. Only an object with two or more SHALL be saved differently. The texture defaults file SHALL follow the same rule. A level in which no object has more than one collision rectangle SHALL therefore keep its fingerprint.

#### Scenario: The existing level
- **WHEN** a development build loads the level, which has no tile with more than one collision rectangle, and saves it without an edit
- **THEN** the level file is byte-for-byte what it was, and its fingerprint is unchanged
