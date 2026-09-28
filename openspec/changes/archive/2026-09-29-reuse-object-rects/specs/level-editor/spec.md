## MODIFIED Requirements

### Requirement: Palette defaults and placed objects are edited separately
Editing a palette entry's rectangles SHALL change what new placements of that texture are born with, and SHALL change, at the same moment, every placed object of that texture that follows it. Objects of any other texture SHALL NOT change.

A placed object SHALL follow its texture from the moment it is placed. Editing a placed object's rectangles — reshaping, adding, clearing or pasting onto them — SHALL change only that object, SHALL NOT change the palette entry or any other object placed with the same texture, and SHALL make that object keep its own rectangles: a later change to its texture SHALL NOT change it.

A palette entry with no stored rectangles SHALL start with none, and an object placed from it SHALL be born with none until that entry is given some.

#### Scenario: Defaults apply to new tiles
- **WHEN** the developer changes a palette entry's collision rectangle and places a new tile of that texture
- **THEN** the new tile has the new rectangle

#### Scenario: Placed tiles follow their texture
- **WHEN** the developer places several tiles of a texture and then changes that palette entry's collision rectangle
- **THEN** every one of those tiles has the new collision rectangle, and the fighter collides with the new shape when the match resumes

#### Scenario: Other textures are untouched
- **WHEN** the developer changes one palette entry's rectangles
- **THEN** no tile of any other texture changes

#### Scenario: One tile only
- **WHEN** the developer reshapes one placed tile's collision rectangle
- **THEN** every other tile of the same texture is unchanged, and so is the palette entry

#### Scenario: A hand-edited tile keeps its own
- **WHEN** the developer reshapes one placed tile's collision rectangle and then changes that texture's collision rectangle
- **THEN** the reshaped tile keeps the shape it was given, and every other tile of the texture takes the new one

## ADDED Requirements

### Requirement: A placed object can follow its texture again
With a placed object selected that keeps its own rectangles, pressing F in edit mode SHALL give it its texture's current rectangles and make it follow its texture again. The editor SHALL show whether the selected placed object follows its texture or keeps its own rectangles. F SHALL do nothing when the selection is a palette entry, when the selected object already follows its texture, or when its texture has no palette entry.

#### Scenario: Back to the texture
- **WHEN** the developer selects a hand-edited tile, presses F, and then changes that texture's collision rectangle
- **THEN** on F the tile takes the texture's rectangles, and on the texture's change it takes the new one

#### Scenario: Shown
- **WHEN** the developer selects a tile, reshapes it, and presses F
- **THEN** the editor shows the tile as following its texture, then as keeping its own rectangles, then as following again

### Requirement: Rectangles can be copied and pasted
In edit mode, Ctrl+C SHALL copy the selected thing's collision and light rectangles, and Ctrl+V SHALL replace the selected thing's collision and light rectangles with the copied ones. A rectangle the source does not have SHALL be copied as absent, so pasting it clears that rectangle on the target. The copy SHALL be a copy and not a link: changing the source after copying SHALL NOT change the copy or anything it was pasted onto.

Pasting onto a palette entry SHALL be an edit of that entry, and pasting onto a placed object SHALL be an edit of that object, with everything the requirement on editing entries and placed objects says of each. Copy and paste SHALL work with either tool and with the object picker open or closed. The copy SHALL last until another copy is made or the game quits, and the editor SHALL show what was copied last. Ctrl+V with nothing copied, or with nothing selected, SHALL do nothing. A paste that changes nothing SHALL NOT count as an unsaved edit. Ctrl+C and Ctrl+V SHALL NOT also act as the editor's C key or any other key.

#### Scenario: One texture's rectangles to another
- **WHEN** the developer selects a texture in the picker, presses Ctrl+C, selects another texture in the picker and presses Ctrl+V
- **THEN** the second texture has exactly the first texture's collision and light rectangles

#### Scenario: A copy, not a link
- **WHEN** the developer copies one texture's rectangles, pastes them onto another, and then reshapes the first
- **THEN** the second texture keeps the rectangles that were pasted

#### Scenario: Pasting onto a texture reaches its tiles
- **WHEN** the developer pastes rectangles onto a texture that has tiles placed in the level
- **THEN** every tile of that texture that follows it has the pasted rectangles

#### Scenario: Pasting onto a tile
- **WHEN** the developer pastes rectangles onto one placed tile and then changes that tile's texture
- **THEN** the tile has the pasted rectangles and keeps them through the texture's change

#### Scenario: Absence is copied
- **WHEN** the developer copies from a texture with no light rectangle and pastes onto a texture that has one
- **THEN** the second texture has no light rectangle

#### Scenario: Nothing to paste
- **WHEN** the developer presses Ctrl+V before anything has been copied
- **THEN** nothing changes and no edit is pending

#### Scenario: Copy does not retool
- **WHEN** the developer is editing the light rectangle and presses Ctrl+C
- **THEN** the light rectangle is still the one being edited

### Requirement: Loading a level never changes it
Loading a level SHALL NOT change any object's rectangles. On load, a placed object SHALL follow its texture when its rectangles are exactly its texture's, and SHALL keep its own rectangles when they differ or when its texture has no palette entry. The level file SHALL NOT change shape, and a development build and a release build loading the same level file SHALL agree on its fingerprint.

#### Scenario: A hand-edited tile survives a restart
- **WHEN** the developer reshapes one tile, saves with Ctrl+S, restarts, and changes that tile's texture
- **THEN** the reshaped tile keeps its shape, and the texture's other tiles take the change

#### Scenario: An existing level loads unchanged
- **WHEN** a development build loads a level saved before this change
- **THEN** every tile has the rectangles the file gives it

#### Scenario: The fingerprint is unaffected
- **WHEN** a development build and a release build load the same level file
- **THEN** they report the same level fingerprint
