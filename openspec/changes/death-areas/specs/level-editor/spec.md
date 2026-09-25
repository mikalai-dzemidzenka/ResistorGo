## ADDED Requirements

### Requirement: Death areas are placed and shaped with the death tool
The editor SHALL have a death tool, reached by the same key that cycles its tools and coming after the area tool. Only the death tool SHALL place, shape, select or remove a death area, and it SHALL NOT place tiles, shape an object's rectangles, or touch a light or a lit area; no other tool SHALL touch a death area.

With the death tool out, a left click on empty space SHALL place a death area there and select it; a left click on a death area SHALL take hold of it — by its body to move it, by an edge or a corner to resize it, as the rect tool's rectangles are held — and select it. A dragged death area SHALL follow the cursor in whole level pixels, SHALL never be dragged thinner than a pixel and SHALL never be dragged inside out. A right click on a death area SHALL remove it; a right click on empty space SHALL clear the selection without removing anything and without marking the level unsaved. The Delete key SHALL remove the selected death area, unless a numeric field has the keyboard. A newly placed death area SHALL take the size of the one selected before it, and a default size when none is selected.

#### Scenario: Reaching the tool
- **WHEN** the developer cycles the editor's tools from the area tool
- **THEN** the death tool comes up, the panel says so, and no light, lit area, tile or rectangle is changed on the way

#### Scenario: Placing a death area
- **WHEN** the developer, with the death tool out, left-clicks a spot with no death area on it
- **THEN** a death area appears there, selected

#### Scenario: Moving and resizing
- **WHEN** the developer drags a death area by its body, and then by a corner
- **THEN** the first drag moves it and the second resizes it, both in whole pixels

#### Scenario: Never inside out
- **WHEN** the developer drags a death area's right edge past its left one
- **THEN** the area stops at a pixel wide instead of turning inside out

#### Scenario: Removing a death area
- **WHEN** the developer right-clicks a death area, and then presses Delete with another one selected
- **THEN** the first is removed by the click and the second by the key

#### Scenario: Tools keep to their own
- **WHEN** the developer clicks a lit area with the death tool, and a death area with the area tool
- **THEN** neither is selected, moved or removed by the other's tool

### Requirement: The death tool's panel edits the selected death area's numbers
With the death tool out and a death area selected, the panel SHALL show that area's position and size as numeric fields that set it exactly when typed into. With none selected, the panel SHALL say how many death areas the level has and how to place one. A key typed into a numeric field SHALL go to that field alone and SHALL NOT also retool the editor or remove anything.

#### Scenario: Typing a death area exactly
- **WHEN** the developer selects a death area and types a position and a size into the panel's fields
- **THEN** the area takes those numbers exactly

#### Scenario: Nothing selected
- **WHEN** the developer takes out the death tool with no death area selected
- **THEN** the panel says how many death areas the level has and how to place one

### Requirement: Death areas are shown while editing and in debug mode
While the death tool is out, the editor SHALL mark every death area over the scene, the selected one told apart from the rest, in a colour that cannot be mistaken for a lit area's marker. The markers SHALL be legible over light and dark and SHALL survive the palette pass at any zoom. In debug mode every death area SHALL also be outlined in play, alongside the collision boxes. Outside debug mode, and in a release build's play, nothing SHALL be drawn for a death area, and a death area SHALL never change how the level is lit.

#### Scenario: The markers
- **WHEN** the developer takes out the death tool in a level with several death areas
- **THEN** each is marked, and the selected one is told apart from the others

#### Scenario: Seen in debug play
- **WHEN** a level with a death area is played in debug mode with another tool out
- **THEN** the death area is outlined with the collision boxes

#### Scenario: Unseen in play
- **WHEN** a level with a death area is played with debug mode off
- **THEN** nothing is drawn where the area is, and the level is lit exactly as it would be without it

### Requirement: Death areas are saved with the level
A death-area edit — placing, moving, resizing, retyping or removing one — SHALL mark the level unsaved, SHALL be undone and redone like any other edit, SHALL be written only by Ctrl+S in edit mode, and SHALL never be written by a release build. Selecting or deselecting a death area SHALL NOT mark the level unsaved. The level file SHALL hold the death areas under a key of their own, and only when the level has some, so a level without death areas is saved byte for byte as it was and keeps its fingerprint. A level with death areas SHALL have a different fingerprint from the same level without them.

#### Scenario: Saved with the level
- **WHEN** the developer places a death area, presses Ctrl+S and restarts the game
- **THEN** the death area is where it was placed, and a fighter touching it dies

#### Scenario: Undone like any edit
- **WHEN** the developer places a death area, moves it, and undoes twice
- **THEN** the area is back where it was placed, and then gone

#### Scenario: A level without them keeps its shape
- **WHEN** a development build loads a level with no death areas and saves it without an edit
- **THEN** the level file is byte-for-byte what it was, and its fingerprint is unchanged

#### Scenario: Client and server disagree on a death area
- **WHEN** a client's level has a death area the server's level does not
- **THEN** the two levels have different fingerprints
