## MODIFIED Requirements

### Requirement: Options screen
The main menu SHALL offer an Options entry, opening a screen with a fullscreen toggle, a window size selector, a music volume slider, a sound volume slider and a Back button. Back SHALL return to the main menu. The screen SHALL be reachable from the main menu only, not during a match.

#### Scenario: Opening options
- **WHEN** the player selects Options on the main menu
- **THEN** a screen appears with the fullscreen toggle, the window size selector, the music volume slider, the sound volume slider and Back

#### Scenario: Returning from options
- **WHEN** the player selects Back on the options screen
- **THEN** the main menu is shown again

## ADDED Requirements

### Requirement: Sound volume setting
The sound volume SHALL be a value from 0 to 100, 80 unless the player has set it, saved and loaded with the other settings. A settings file that does not have it SHALL give the default, and a value out of range SHALL be brought into it.

#### Scenario: Survives a restart
- **WHEN** the player sets the sound volume to 25, leaves the options screen and restarts the game
- **THEN** the sound volume is 25

#### Scenario: An older settings file
- **WHEN** the game starts with a settings file written before the sound volume existed
- **THEN** the other settings are as saved and the sound volume is 80
