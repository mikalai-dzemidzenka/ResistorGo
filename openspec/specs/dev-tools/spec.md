# dev-tools Specification

## Purpose

Keeps the developer's tools out of the fight and out of players' hands: the level editor only listens in an explicit edit mode that only saves when asked to, and a release build ships without the editor, the debug overlay, the network-simulation keys or any level writing.

## Requirements

### Requirement: Edit mode is explicit
In a development build, in a single-player match, F2 SHALL toggle edit mode: pressing it while edit mode is off turns it on, pressing it again turns it off. F2 SHALL do nothing in a networked match and while time manipulation is on. While edit mode is on, T and R SHALL do nothing.

#### Scenario: Enter and leave
- **WHEN** the player presses F2 in a single-player match, then presses it again
- **THEN** edit mode turns on, then off, and the match continues

#### Scenario: Not online
- **WHEN** the player presses F2 in a networked match, as a fighter or as a spectator
- **THEN** nothing happens and the match continues

#### Scenario: Not during time manipulation
- **WHEN** the player presses F2 while time manipulation is on
- **THEN** edit mode stays off

#### Scenario: No time manipulation while editing
- **WHEN** the player presses T while edit mode is on
- **THEN** time manipulation does not start

#### Scenario: No restart while editing
- **WHEN** the player presses R while edit mode is on
- **THEN** the level is not restarted

### Requirement: Edit mode pauses the match
While edit mode is on, the match SHALL be paused: no fighter physics, no animation, no bot decisions, and no fighter input applied. No input SHALL be buffered for when edit mode is left, and leaving it SHALL NOT produce a burst of catch-up simulation.

#### Scenario: Nothing moves while editing
- **WHEN** edit mode has been on for several seconds
- **THEN** every fighter is exactly where it was when edit mode was entered, and no attack has landed

#### Scenario: Clicks do not carry over
- **WHEN** the player clicks the left mouse button several times in edit mode and then leaves it
- **THEN** the fighter does not attack on leaving

#### Scenario: No catch-up
- **WHEN** the player leaves edit mode after ten seconds in it
- **THEN** the match resumes at normal speed with no jump in fighter positions

### Requirement: The editor takes input only in edit mode
Outside edit mode the level editor SHALL take no input: the left mouse button SHALL only attack, the right mouse button and Tab SHALL do nothing to the level, and a file dropped on the window SHALL do nothing.

In edit mode Tab SHALL open the object picker, and the editor SHALL have four tools with exactly one of them active, switched explicitly by the player: the place tool, the rect tool, the light tool and the area tool. With the place tool the left mouse button SHALL place the selected object at the grid cell under the cursor and the right mouse button SHALL delete the object there. With the rect tool the left mouse button SHALL select an object and shape rectangles only, and SHALL neither place nor delete an object. With the light tool the mouse SHALL place, select, shape and remove lights only, and SHALL neither place, delete nor select an object. With the area tool the mouse SHALL place, select, shape and remove lit areas only, and SHALL neither place, delete nor select an object, nor touch a light. One click SHALL never act as two tools at once.

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

#### Scenario: The light tool does not touch tiles
- **WHEN** the player, in edit mode with the light tool, left-clicks a tile and right-clicks another
- **THEN** a light is placed where the first click was, and both tiles are unchanged

#### Scenario: Light clicks in play do nothing
- **WHEN** the player left-clicks and right-clicks on a light's center outside edit mode
- **THEN** the light is unchanged, and the left click is an attack

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

### Requirement: Edit mode is visible
While edit mode is on, the screen SHALL say so, and SHALL say when there are edits that have not been saved.

#### Scenario: Marker
- **WHEN** edit mode is on and nothing has been changed since the last save
- **THEN** the screen shows that the game is in edit mode

#### Scenario: Unsaved marker
- **WHEN** the player places a tile in edit mode
- **THEN** the screen shows that there are unsaved edits, until Ctrl+S is pressed

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

### Requirement: A release build plays the embedded level
A release build SHALL load the level embedded in the binary and SHALL ignore any level file on disk, when running as a game and when running as a dedicated server. Its level fingerprint SHALL therefore be the same as every other copy of the same release.

#### Scenario: A stale level file beside the exe
- **WHEN** a release build is started with a modified `level_1.json` in its working directory
- **THEN** it plays the embedded level, and joining a server from the same release is not rejected with "level mismatch"

#### Scenario: Dedicated server
- **WHEN** a release dedicated server starts with a modified `level_1.json` in its working directory
- **THEN** it serves the embedded level

### Requirement: A development build keeps the tools
A development build — the one built without the release tag, as `go run ./cmd` and `go test ./...` do — SHALL keep the debug overlay, on by default and toggled with `` ` ``, edit mode, and the network-simulation keys. It SHALL load the level file from disk when one is present, so saved edits take effect on the next run.

#### Scenario: Saved edits are used
- **WHEN** the developer saves an edit with Ctrl+S and restarts a development build
- **THEN** the edited level is loaded
