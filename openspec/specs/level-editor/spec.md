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
With the light tool and no light selected, the panel SHALL show the level's ambient light — its colour and its strength — as numeric fields that set it exactly when typed into, and the level SHALL be lit by the new ambient on the next frame drawn. Selecting a light SHALL show that light's own fields, as it does today, and deselecting SHALL bring the ambient fields back. No other tool SHALL change the ambient.

The selected light SHALL always be possible to let go of: a right click on empty space SHALL clear the selection without removing any light and without marking the level unsaved, while a right click on a light SHALL remove it as it does today. Nothing else about the selection SHALL change: a left click on empty space still places a light.

An ambient edit SHALL mark the level unsaved, SHALL be written only by Ctrl+S in edit mode, and SHALL never be written by a release build — exactly as an edit to a light is.

#### Scenario: The ambient fields
- **WHEN** the developer takes out the light tool with no light selected and types a new colour and strength
- **THEN** the level's unlit parts take that colour and that strength as the values are typed

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
With the area tool out and a lit area selected, the panel SHALL show that area's position, size and feather as numeric fields that set it exactly when typed into, and the level SHALL be lit by the typed numbers on the next frame drawn. With no area selected, the panel SHALL say how many lit areas the level has and how to place one. A key typed into a numeric field SHALL go to that field alone and SHALL NOT also retool the editor or remove anything.

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
