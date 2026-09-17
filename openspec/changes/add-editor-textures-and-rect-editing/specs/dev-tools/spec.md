## MODIFIED Requirements

### Requirement: The editor takes input only in edit mode
Outside edit mode the level editor SHALL take no input: the left mouse button SHALL only attack, the right mouse button and Tab SHALL do nothing to the level, and a file dropped on the window SHALL do nothing.

In edit mode Tab SHALL open the object picker, and the editor SHALL have two tools with exactly one of them active, switched explicitly by the player. With the place tool the left mouse button SHALL place the selected object at the grid cell under the cursor and the right mouse button SHALL delete the object there. With the rect tool the left mouse button SHALL select an object and shape rectangles only, and SHALL neither place nor delete an object. One click SHALL never both place an object and shape a rectangle.

#### Scenario: Attacking does not edit
- **WHEN** the player attacks many times in a single-player match, with the debug overlay on
- **THEN** the level is unchanged and no edit is pending

#### Scenario: Right-click in play does nothing
- **WHEN** the player right-clicks with the cursor over the floor, outside edit mode
- **THEN** the floor tile is still there

#### Scenario: A dropped file in play does nothing
- **WHEN** the player drops an image on the window outside edit mode
- **THEN** nothing is imported, the level is unchanged, and the match continues

#### Scenario: Editing
- **WHEN** the player, in edit mode with the place tool, left-clicks an empty grid cell and right-clicks an existing tile
- **THEN** the selected object appears in the empty cell and the existing tile is removed

#### Scenario: The rect tool does not place
- **WHEN** the player, in edit mode with the rect tool, left-clicks an empty grid cell and then a tile's rectangle
- **THEN** no object is placed, and the click on the rectangle selects and shapes it

### Requirement: Saving the level is explicit
The level file and the editor's palette defaults file SHALL be written only when the player presses Ctrl+S in edit mode. Quitting the game SHALL NOT write either of them. Edits SHALL survive leaving and re-entering edit mode for the rest of the session, and unsaved edits SHALL be lost when the game quits. A placement, a deletion and a rectangle edit SHALL all count as an edit on these terms.

#### Scenario: Save
- **WHEN** the player places a tile in edit mode and presses Ctrl+S
- **THEN** the level file on disk contains the new tile

#### Scenario: Quit without saving
- **WHEN** the player places a tile in edit mode and quits without pressing Ctrl+S
- **THEN** the level file on disk is unchanged

#### Scenario: Edits kept for the session
- **WHEN** the player places a tile, leaves edit mode, plays, and enters edit mode again
- **THEN** the tile is still there and still unsaved

#### Scenario: A rectangle edit is an edit
- **WHEN** the player reshapes a tile's collision rectangle in edit mode and presses Ctrl+S
- **THEN** the level file on disk contains the reshaped rectangle

### Requirement: A release build carries no developer tools
A release build SHALL contain no edit mode, no level editor input or GUI, no debug overlay, no network-simulation keys and no texture import, and SHALL never write the level file or the editor's palette defaults file. The keys that open these tools in a development build — F2, the `` ` `` key, Tab in the editor's sense, the editor's tool keys, and F5 to F8 — SHALL do nothing in a release build, and a file dropped on its window SHALL do nothing. A release build SHALL use only the block textures it was built with and SHALL NOT look for block textures on disk.

#### Scenario: Keys do nothing
- **WHEN** a player of a release build presses F2, `` ` ``, and F5 to F8 during a match
- **THEN** no edit mode starts, no overlay appears, and the network conditions are unchanged

#### Scenario: Never writes the level
- **WHEN** a player of a release build plays a match, attacking and right-clicking throughout, and quits
- **THEN** no level file and no palette defaults file has been written anywhere

#### Scenario: Dropping art does nothing
- **WHEN** a player of a release build drops a PNG on the window during a match
- **THEN** nothing is imported, nothing is copied anywhere, and the match continues

#### Scenario: Ignores textures on disk
- **WHEN** a release build is started with an extra PNG beside its other block textures on disk
- **THEN** it plays with the textures it was built with and that file is never read
