## ADDED Requirements

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
