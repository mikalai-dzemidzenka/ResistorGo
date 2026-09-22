## Purpose

Lets a developer build a level out of the art that actually exists: every block texture in the build is placeable, a new one can be brought in by dropping it on the window, and each object's collision box and light occluder are shaped on screen instead of typed into Go source.

## ADDED Requirements

### Requirement: The palette is every block texture in the build
The editor's object palette SHALL list every block texture the build carries, one entry per texture, and every entry SHALL be placeable. A development build SHALL additionally list any block texture present on disk but not in the build. No texture SHALL need a code change to become placeable. The order of the palette SHALL be stable between runs, so the same entry is in the same place every time the picker is opened.

#### Scenario: Every texture is offered
- **WHEN** the developer opens the object picker in edit mode
- **THEN** the picker shows an entry for every block texture in the build, not only the three that had rectangles written in code

#### Scenario: New art needs no code
- **WHEN** a new block texture is added to the game's block textures and the game is started
- **THEN** it appears in the picker and can be placed, with no edit to any Go file

#### Scenario: Stable order
- **WHEN** the picker is opened, closed and opened again, and the game is restarted and the picker opened once more
- **THEN** the entries are in the same order every time

### Requirement: A texture can be imported by dropping it
In a development build, in edit mode, dropping one or more image files onto the window SHALL import them: each file SHALL be copied in beside the game's other block textures, loaded, and added to the palette, and the last one imported SHALL become the selected palette entry. An imported texture SHALL be placeable immediately, without restarting the game.

A dropped file that is not a supported image, or that fails to load, SHALL be ignored: the palette SHALL be unchanged and the failure SHALL be reported in the log. A dropped file whose name is already taken among the block textures SHALL NOT overwrite the existing one; the existing entry SHALL be selected instead.

#### Scenario: Import and place
- **WHEN** the developer drops a PNG on the window in edit mode and left-clicks an empty grid cell
- **THEN** the dropped texture is the selected object and a tile of it is placed in that cell

#### Scenario: Import survives a restart
- **WHEN** the developer imports a texture and restarts the development build
- **THEN** the imported texture is still in the picker

#### Scenario: Not an image
- **WHEN** the developer drops a text file on the window in edit mode
- **THEN** the palette is unchanged, the selected object is unchanged, and the game continues

#### Scenario: Name already taken
- **WHEN** the developer drops a PNG whose name matches a block texture already in the game
- **THEN** the file on disk is not overwritten and the existing palette entry is selected

### Requirement: A texture that is not yet in the build is marked
The editor SHALL mark a palette entry whose texture is not part of the build, and SHALL mark a placed object drawn with such a texture, so it is clear that a build made from the current binary would not carry it. The marking SHALL go away once the texture is part of the build.

#### Scenario: Freshly imported
- **WHEN** the developer imports a texture and opens the picker
- **THEN** that entry is marked as not being part of the build

#### Scenario: Rebuilt
- **WHEN** the developer imports a texture, rebuilds, and starts the new build
- **THEN** that entry is no longer marked

### Requirement: One thing is selected for editing
Exactly one thing SHALL be selected for rectangle editing at a time: either a palette entry or one object placed in the level. Picking an entry in the object picker SHALL select that palette entry. Clicking a placed object with the rect tool SHALL select that object. The editor SHALL show which of the two is selected and which object it is.

#### Scenario: Picking selects the entry
- **WHEN** the developer picks a texture in the object picker
- **THEN** that palette entry is what rectangle editing acts on, and the editor says so

#### Scenario: Clicking selects the tile
- **WHEN** the developer clicks a placed tile with the rect tool
- **THEN** that tile is what rectangle editing acts on, and the editor says so

#### Scenario: Only one at a time
- **WHEN** the developer clicks a placed tile with the rect tool and then picks a texture in the picker
- **THEN** only the palette entry is selected, and the tile is no longer selected

### Requirement: Rectangles are shaped on screen
With the rect tool, the selected thing's collision rectangle and light rectangle SHALL both be drawn over the level, distinguishable from one another, with exactly one of them active for editing. The developer SHALL be able to switch which one is active without losing the selection.

Dragging an edge or a corner of the active rectangle SHALL resize it from that edge or corner. Dragging its body SHALL move it without changing its size. Both SHALL work in whole level pixels, and SHALL work at every camera zoom level. A rectangle's width and height SHALL never be dragged below one pixel. Releasing the mouse button SHALL end the drag. Numeric fields SHALL show the active rectangle's position and size and SHALL set them exactly.

An edit SHALL take effect at once: a reshaped light rectangle SHALL change the level's lighting on the next frame drawn, and a reshaped collision rectangle SHALL be what fighters collide with when the match resumes.

#### Scenario: Resize by a corner
- **WHEN** the developer drags the bottom-right corner of the active collision rectangle two pixels to the right
- **THEN** that rectangle is two pixels wider and its top-left corner has not moved

#### Scenario: Move the body
- **WHEN** the developer drags the middle of the active rectangle
- **THEN** it moves with the cursor and keeps its width and height

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
The selected thing SHALL be able to have either rectangle added when it has none, and either rectangle cleared when it has one. A newly added rectangle SHALL start as one grid cell at the object's origin, so it is visible and can be dragged from there. An object with no collision rectangle SHALL NOT collide with fighters, and an object with no light rectangle SHALL NOT cast a shadow.

#### Scenario: Add one
- **WHEN** the developer selects a placed tile that has no light rectangle and adds one
- **THEN** a one-cell light rectangle appears at that tile's origin and can be dragged

#### Scenario: Clear one
- **WHEN** the developer clears a placed tile's collision rectangle and leaves edit mode
- **THEN** a fighter passes through that tile

#### Scenario: No shadow
- **WHEN** the developer clears a placed tile's light rectangle
- **THEN** that tile stops casting a shadow

### Requirement: Palette defaults and placed objects are edited separately
Editing a palette entry's rectangles SHALL change what new placements of that texture are born with, and SHALL NOT change any object already placed in the level. Editing a placed object's rectangles SHALL change only that object, and SHALL NOT change the palette entry or any other object placed with the same texture.

A palette entry with no stored rectangles SHALL start with none, and an object placed from it SHALL be born with none until that entry is given some.

#### Scenario: Defaults apply to new tiles only
- **WHEN** the developer changes a palette entry's collision rectangle and places a new tile of that texture
- **THEN** the new tile has the new rectangle and the tiles placed earlier still have the old one

#### Scenario: One tile only
- **WHEN** the developer reshapes one placed tile's collision rectangle
- **THEN** every other tile of the same texture is unchanged, and so is the palette entry

### Requirement: Palette defaults persist
Palette rectangle defaults SHALL be kept in a file of their own, separate from the level. They SHALL be written only in a development build, and only when the developer presses Ctrl+S in edit mode — the same terms the level file is written on. Unsaved default edits SHALL be shown as unsaved edits and SHALL be lost if the game quits without saving.

A missing or malformed defaults file SHALL NOT stop the editor: the palette SHALL fall back to the rectangles built into the game, and every texture without any SHALL start with none.

#### Scenario: Saved defaults come back
- **WHEN** the developer changes a palette entry's rectangles, presses Ctrl+S, and restarts the development build
- **THEN** that entry still has the changed rectangles

#### Scenario: Quit without saving
- **WHEN** the developer changes a palette entry's rectangles and quits without pressing Ctrl+S
- **THEN** the defaults file on disk is unchanged

#### Scenario: Malformed defaults file
- **WHEN** the development build starts with a corrupt defaults file
- **THEN** the editor opens with the built-in rectangles and the game plays

### Requirement: A missing texture is visible
An object in the level whose texture the build does not have SHALL NOT be drawn as an invisible tile in a development build: a marker SHALL be drawn in its place. Its collision rectangle SHALL still apply, so the level plays as its data says.

#### Scenario: Level outran its assets
- **WHEN** a development build loads a level referring to a texture that is not in the build
- **THEN** a marker is drawn where that object is, and the fighter still collides with it
