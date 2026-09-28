## ADDED Requirements

### Requirement: Tiles are placed on the editor's current layer
The editor SHALL have a current layer, shown in its panel. A newly placed tile SHALL go on the current layer, and erasing SHALL remove only the tile on the current layer in that cell. The current layer SHALL start at -1. The developer SHALL be able to step it up and down one layer at a time, and stepping SHALL skip layer 0, which belongs to the fighters. Tiles on different layers SHALL be able to share a grid cell.

#### Scenario: Default layer
- **WHEN** the developer opens the editor and places a tile without changing the layer
- **THEN** the tile is on layer -1, and the panel shows layer -1

#### Scenario: Stepping skips the fighters' layer
- **WHEN** the current layer is -1 and the developer steps it up
- **THEN** the current layer is 1

#### Scenario: Stacked tiles
- **WHEN** the developer places a tile on layer -1, steps to layer 1 and places another tile in the same cell
- **THEN** both tiles are in the level, and erasing that cell on layer 1 removes only the layer-1 tile

### Requirement: Layers set the draw order around the fighters
Fighters SHALL be drawn on layer 0. Tiles on negative layers SHALL be drawn behind the fighters, and tiles on positive layers SHALL be drawn in front of them. Within each side, a lower layer SHALL be drawn before a higher one. Blood stains and holes SHALL stay drawn over the background tiles and under the fighters. Tiles on every layer SHALL be lit by the level's lights and ambient light, the same as today. Collision SHALL NOT depend on layer: every tile's collision and drop-through rectangles SHALL apply whatever layer it is on.

#### Scenario: A foreground pillar
- **WHEN** a fighter walks behind a tile placed on layer 1
- **THEN** the tile is drawn over the fighter

#### Scenario: Background behind fighters
- **WHEN** a fighter stands in front of a tile on layer -1
- **THEN** the fighter is drawn over the tile

#### Scenario: A foreground floor still holds
- **WHEN** a tile with a collision rectangle is on layer 1
- **THEN** fighters stand on it exactly as they would if it were on layer -1

### Requirement: Tiles from an older level load behind the fighters
A tile a level file holds on layer 0 SHALL be loaded onto layer -1, so the next save writes it there. Client and server SHALL apply the same rule, so they compute the same fingerprint for the same file.

#### Scenario: The existing level
- **WHEN** the game loads a level file whose tiles are all on layer 0
- **THEN** every tile is on layer -1, drawn behind the fighters exactly as before

#### Scenario: Networked match on an old file
- **WHEN** a client and a server load the same level file that has tiles on layer 0
- **THEN** their fingerprints match

## MODIFIED Requirements

### Requirement: The level file keeps its shape
An object with no collision rectangle or one, no light rectangle or one, and no drop-through rectangle SHALL be saved in the level file exactly as it was before objects could have several of either or any drop-through rectangle. Only a kind of which an object has two or more SHALL be saved as a list, and drop-through rectangles SHALL be saved under keys of their own, one or a list by the same rule, and only when an object has some. The texture defaults file SHALL follow the same rule. A level with no tile on layer 0, in which no object has more than one rectangle of either of the first two kinds and none has a drop-through rectangle, SHALL therefore keep its fingerprint.

#### Scenario: The existing level
- **WHEN** a development build loads a level with no tile on layer 0 in which no tile has more than one light rectangle, and saves it without an edit
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
