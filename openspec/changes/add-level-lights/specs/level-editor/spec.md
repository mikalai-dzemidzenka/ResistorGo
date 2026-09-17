## ADDED Requirements

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
