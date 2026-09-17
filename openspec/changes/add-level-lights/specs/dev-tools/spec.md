## MODIFIED Requirements

### Requirement: The editor takes input only in edit mode
Outside edit mode the level editor SHALL take no input: the left mouse button SHALL only attack, the right mouse button and Tab SHALL do nothing to the level, and a file dropped on the window SHALL do nothing.

In edit mode Tab SHALL open the object picker, and the editor SHALL have three tools with exactly one of them active, switched explicitly by the player: the place tool, the rect tool and the light tool. With the place tool the left mouse button SHALL place the selected object at the grid cell under the cursor and the right mouse button SHALL delete the object there. With the rect tool the left mouse button SHALL select an object and shape rectangles only, and SHALL neither place nor delete an object. With the light tool the mouse SHALL place, select, shape and remove lights only, and SHALL neither place, delete nor select an object. One click SHALL never act as two tools at once.

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
