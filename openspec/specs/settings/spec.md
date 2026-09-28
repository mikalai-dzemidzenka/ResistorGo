# settings Specification

## Purpose

Covers the options screen the player uses to change display and sound settings, and the file those choices are saved in so they survive a restart.

## Requirements

### Requirement: Options screen
The main menu SHALL offer an Options entry, opening a screen with a fullscreen toggle, a window size selector, a music volume slider, a sound volume slider and a Back button. Back SHALL return to the main menu. The screen SHALL be reachable from the main menu only, not during a match.

#### Scenario: Opening options
- **WHEN** the player selects Options on the main menu
- **THEN** a screen appears with the fullscreen toggle, the window size selector, the music volume slider, the sound volume slider and Back

#### Scenario: Returning from options
- **WHEN** the player selects Back on the options screen
- **THEN** the main menu is shown again

### Requirement: Settings apply immediately
Every setting SHALL take effect as soon as it is changed on the options screen, with no confirmation step and no restart.

#### Scenario: Immediate effect
- **WHEN** the player toggles fullscreen, picks a different window size or moves the volume slider
- **THEN** that change is visible or audible at once, while the options screen is still open

### Requirement: Settings persist across runs
The player's settings SHALL be saved to a settings file when the options screen is left, and loaded and applied at startup before the window is shown.

#### Scenario: Settings survive a restart
- **WHEN** the player sets the music volume to 30, turns fullscreen on, leaves the options screen and restarts the game
- **THEN** the game starts fullscreen with the music volume at 30

### Requirement: Invalid settings fall back to defaults
A missing, unreadable or malformed settings file SHALL NOT stop the game from starting. The game SHALL fall back to defaults. Values outside their valid range SHALL be clamped or replaced by the default for that setting.

#### Scenario: No settings file
- **WHEN** the game is started with no settings file present
- **THEN** it starts with default settings and writes the file the next time the options screen is left

#### Scenario: Corrupt settings file
- **WHEN** the settings file contains malformed content or an out-of-range value
- **THEN** the game starts using defaults for whatever could not be read, without crashing

### Requirement: Sound volume setting
The sound volume SHALL be a value from 0 to 100, 80 unless the player has set it, saved and loaded with the other settings. A settings file that does not have it SHALL give the default, and a value out of range SHALL be brought into it.

#### Scenario: Survives a restart
- **WHEN** the player sets the sound volume to 25, leaves the options screen and restarts the game
- **THEN** the sound volume is 25

#### Scenario: An older settings file
- **WHEN** the game starts with a settings file written before the sound volume existed
- **THEN** the other settings are as saved and the sound volume is 80
