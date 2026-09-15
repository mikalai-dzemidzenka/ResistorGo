## ADDED Requirements

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
