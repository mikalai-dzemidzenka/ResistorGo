# level-editor Specification

## Purpose

Lets a developer build a level out of the art that actually exists: every block texture in the build is placeable, a new one can be brought in by dropping it on the window, and each object's collision box and light occluder are shaped on screen instead of typed into Go source.

## Requirements

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

### Requirement: Lights are placed and removed with the light tool
Edit mode SHALL have a light tool. With it, a left click on an empty spot SHALL place a light there, in whole level pixels, and select it; a left click on a light SHALL select it; and Delete, or a right click on a light, SHALL remove it. A newly placed light SHALL take the radius, color and intensity of the light selected when it is placed, or a default radius, white and full intensity when none is selected. The light tool SHALL NOT place, delete or select tiles, nor shape their rectangles; the place tool and the rect tool SHALL NOT place, select or change lights.

#### Scenario: Place
- **WHEN** the developer, with the light tool, left-clicks an empty spot
- **THEN** a light appears there, selected, and lights the scene around it

#### Scenario: A matching row
- **WHEN** the developer selects a blue light of radius 64, then left-clicks three empty spots
- **THEN** there are three new blue lights of radius 64

#### Scenario: Remove
- **WHEN** the developer selects a light and presses Delete, and right-clicks another light
- **THEN** both lights are gone

#### Scenario: Tools keep to their own
- **WHEN** the developer left-clicks a tile with the light tool, and a light with the place tool
- **THEN** the first click places a light rather than selecting the tile, and the second places a tile without selecting or moving the light

### Requirement: A light is shaped on screen
With the light tool, dragging a light's center SHALL move it, and dragging the handle on the edge of its circle SHALL change its radius, both in whole level pixels and at every camera zoom level, and releasing the mouse button SHALL end the drag. A light's radius SHALL never be dragged below one pixel. Numeric fields SHALL show the selected light's position, radius, color and intensity and SHALL set them exactly. Every change SHALL show in the level's lighting on the next frame drawn.

#### Scenario: Move
- **WHEN** the developer drags a light's center ten pixels to the right
- **THEN** the light, and the shadows it casts, move ten pixels to the right

#### Scenario: Resize
- **WHEN** the developer drags a light's edge handle outward
- **THEN** the light's radius grows with the drag, and it lights further

#### Scenario: Exact values
- **WHEN** the developer types a color and an intensity into the selected light's fields
- **THEN** the light has exactly that color and intensity

#### Scenario: Never below a pixel
- **WHEN** the developer drags a light's edge handle past its center
- **THEN** its radius is one pixel

### Requirement: Lights are visible while editing
In edit mode with the light tool, every light SHALL show where its center is and the circle its light reaches, with the selected light distinguished from the others. Outside the light tool the markers SHALL NOT be shown.

#### Scenario: Markers
- **WHEN** the developer switches to the light tool in a level with three lights
- **THEN** each of the three shows its center and its circle, and the selected one is marked as selected

### Requirement: Light edits are saved like any other edit
Placing, moving, resizing, recolouring, changing the intensity of and removing a light SHALL each count as an unsaved edit, SHALL be written to the level file only when Ctrl+S is pressed in edit mode, and SHALL be lost if the game quits without saving. A release build SHALL never write them.

#### Scenario: Save
- **WHEN** the developer places a light, presses Ctrl+S and restarts the development build
- **THEN** the light is there

#### Scenario: Unsaved
- **WHEN** the developer moves a light
- **THEN** the editor shows that there are unsaved edits, until Ctrl+S is pressed

#### Scenario: Quit without saving
- **WHEN** the developer places a light and quits without pressing Ctrl+S
- **THEN** the level file on disk is unchanged

### Requirement: The level's ambient light is edited with the light tool
With the light tool and no light selected, the panel SHALL show the level's ambient light — its colour and its strength — as numeric fields that set it exactly when typed into, and as a darkness slider from 0% (the ambient at full strength) to 100% (no ambient at all) that sets its strength as it is dragged, and the level SHALL be lit by the new ambient on the next frame drawn. Selecting a light SHALL show that light's own fields, as it does today, and deselecting SHALL bring the ambient fields back. No other tool SHALL change the ambient.

The selected light SHALL always be possible to let go of: a right click on empty space SHALL clear the selection without removing any light and without marking the level unsaved, while a right click on a light SHALL remove it as it does today. Nothing else about the selection SHALL change: a left click on empty space still places a light.

An ambient edit SHALL mark the level unsaved, SHALL be written only by Ctrl+S in edit mode, and SHALL never be written by a release build — exactly as an edit to a light is.

#### Scenario: The ambient fields
- **WHEN** the developer takes out the light tool with no light selected and types a new colour and strength
- **THEN** the level's unlit parts take that colour and that strength as the values are typed

#### Scenario: The darkness slider
- **WHEN** the developer takes out the light tool with no light selected and drags the darkness slider
- **THEN** the level's unlit parts darken or brighten as it is dragged, the strength field follows it, and the whole drag is one step to undo

#### Scenario: Letting go of a light
- **WHEN** the developer, with a light selected, right-clicks a spot with no light on it
- **THEN** the light stays where it is, the panel goes back to the ambient fields, and the level is no more unsaved than it was

#### Scenario: A light takes the panel
- **WHEN** the developer selects a light while the ambient fields are shown, and then deselects it
- **THEN** the panel shows the light's position, radius, colour and intensity while it is selected, and the ambient fields again once it is not

#### Scenario: Saved with the level
- **WHEN** the developer changes the ambient, sees the level marked unsaved, presses Ctrl+S and restarts the game
- **THEN** the level is lit by the ambient that was saved

#### Scenario: Quit without saving
- **WHEN** the developer changes the ambient and quits without pressing Ctrl+S
- **THEN** the level file is unchanged and the level is lit as it was before the change

### Requirement: Lit areas are placed and shaped with the area tool
The editor SHALL have an area tool, reached by the same key that cycles its tools and neighbouring the light tool, so the two lighting tools are next to one another. Only the area tool SHALL place, shape, select or remove a lit area, and it SHALL NOT place tiles, shape an object's rectangles or touch a light — one click is never two tools at once.

With the area tool out, a left click on empty space SHALL place a lit area there and select it; a left click on an area SHALL take hold of it — by its body to move it, by an edge or a corner to resize it, as the rect tool's rectangles are held — and select it. A dragged area SHALL follow the cursor in whole level pixels, SHALL never be dragged thinner than a pixel and SHALL never be dragged inside out. A right click on an area SHALL remove it; a right click on empty space SHALL clear the selection without removing anything and without marking the level unsaved. The Delete key SHALL remove the selected area, unless a numeric field has the keyboard.

A newly placed area SHALL take the size and feather of the area selected before it, so a row of matching areas is a row of clicks, and a default size and feather when none is selected.

#### Scenario: Reaching the tool
- **WHEN** the developer cycles the editor's tools from the light tool
- **THEN** the area tool comes up, the panel says so, and no light, tile or rectangle is changed on the way

#### Scenario: Placing an area
- **WHEN** the developer, with the area tool out, left-clicks a spot with no area on it
- **THEN** a lit area appears there, selected, and the level is lit through it on that frame

#### Scenario: Moving and resizing
- **WHEN** the developer drags an area by its body, and then by a corner
- **THEN** the first drag moves it and the second resizes it, both in whole pixels, and the level is lit through the new rectangle as it is dragged

#### Scenario: Never inside out
- **WHEN** the developer drags an area's right edge past its left one
- **THEN** the area stops at a pixel wide instead of turning inside out

#### Scenario: Matching areas
- **WHEN** the developer places one area, sets its size and feather, and then clicks three more spots
- **THEN** each new area has that size and that feather

#### Scenario: Removing an area
- **WHEN** the developer right-clicks an area, and then presses Delete with another one selected
- **THEN** the first area is removed by the click and the second by the key

#### Scenario: Letting go
- **WHEN** the developer, with an area selected, right-clicks a spot with no area on it
- **THEN** every area stays where it is, nothing is selected, and the level is no more unsaved than it was

#### Scenario: Not the other tools
- **WHEN** the developer left-clicks and right-clicks over a tile with the area tool out
- **THEN** no tile is placed or erased and no object rectangle is changed

### Requirement: The area tool's panel edits the selected area's numbers
With the area tool out and a lit area selected, the panel SHALL show that area's position, size and feather as numeric fields that set it exactly when typed into, and as a darkness slider from 0% (the ambient at full strength) to 100% (no ambient at all) that sets its strength as it is dragged, and the level SHALL be lit by the typed numbers on the next frame drawn. With no area selected, the panel SHALL say how many lit areas the level has and how to place one. A key typed into a numeric field SHALL go to that field alone and SHALL NOT also retool the editor or remove anything.

#### Scenario: Typing an area exactly
- **WHEN** the developer selects an area and types a position, a size and a feather into the panel's fields
- **THEN** the area takes those numbers exactly and the level is lit through it as they are typed

#### Scenario: Nothing selected
- **WHEN** the developer takes out the area tool with no area selected
- **THEN** the panel says how many lit areas the level has and how to place one

#### Scenario: Typing is not a shortcut
- **WHEN** the developer types digits into a panel field while the area tool is out
- **THEN** the digits land in the field, and the editor's tool and its areas are unchanged

### Requirement: Lit areas are shown while the area tool is out
While the area tool is out, the editor SHALL mark every lit area over the lit scene: its rectangle, the extent of its feather, and the selected one told apart from the rest. The markers SHALL be legible over both a fully lit area and the dark around it, and SHALL survive the palette pass at any zoom. No other tool and no release build SHALL draw them, and nothing SHALL be drawn for an area while the game is played.

#### Scenario: The markers
- **WHEN** the developer takes out the area tool in a level with several lit areas
- **THEN** each area's rectangle and the extent of its feather are marked, and the selected one is told apart from the others

#### Scenario: Only this tool
- **WHEN** the developer switches from the area tool to any other tool
- **THEN** the markers go, and the level is lit exactly as it was

#### Scenario: Legible at any zoom
- **WHEN** the developer zooms the camera all the way in and all the way out with the area tool out
- **THEN** the markers stay visible and unbroken at both

### Requirement: An area edit is saved with the level
An edit to a lit area — placing, moving, resizing, retyping or removing one — SHALL mark the level unsaved, SHALL be written only by Ctrl+S in edit mode, and SHALL never be written by a release build, exactly as an edit to a light is. Selecting or deselecting an area SHALL NOT mark the level unsaved.

#### Scenario: Saved with the level
- **WHEN** the developer places a lit area, sees the level marked unsaved, presses Ctrl+S and restarts the game
- **THEN** the level is lit through the area that was saved

#### Scenario: Quit without saving
- **WHEN** the developer places a lit area and quits without pressing Ctrl+S
- **THEN** the level file is unchanged and the level is lit as it was before

#### Scenario: Selecting is not an edit
- **WHEN** the developer clicks an area and then clicks away from it, on a level with nothing else unsaved
- **THEN** the level is still marked saved

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

### Requirement: An object casts a shadow from every light rectangle
Every light rectangle of every object in the level SHALL cast a shadow from every light, each behaving on its own exactly as an object's single light rectangle does. Light SHALL pass through the space between an object's light rectangles. This SHALL hold in the development build and the release build alike, and the debug overlay SHALL draw every light rectangle of every object.

#### Scenario: A gap lets light through
- **WHEN** a tile has two light rectangles one above the other with a gap between them, and a light stands on one side of the tile
- **THEN** the floor beyond the tile is shadowed behind each rectangle and lit behind the gap

#### Scenario: The overlay shows them all
- **WHEN** the developer turns on the debug overlay over a tile with two light rectangles
- **THEN** both light rectangles are outlined
