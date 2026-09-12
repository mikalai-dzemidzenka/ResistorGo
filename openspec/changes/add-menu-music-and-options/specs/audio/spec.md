## Purpose

Covers the game's sound: opening the audio device, the music that plays on the main menu, and how the player's volume setting is applied to it.

## ADDED Requirements

### Requirement: Audio device lifetime
The game SHALL open an audio device at startup, before any sound is loaded, and close it on exit. If the audio device cannot be opened, the game SHALL still start and run silently rather than failing.

#### Scenario: Game runs without a working audio device
- **WHEN** the game starts on a machine with no working audio output
- **THEN** the game starts, the menu and gameplay work normally, and no error is shown to the player

### Requirement: Main menu music
A single music track SHALL loop while the main menu is shown, on every menu screen. It SHALL start when the menu is first shown and resume when the player returns to the menu from a match.

#### Scenario: Music plays on the menu
- **WHEN** the game is started and the main menu appears
- **THEN** the menu track begins playing and loops when it reaches the end

#### Scenario: Music continues across menu screens
- **WHEN** the player moves between the Main, Host, Join, Options and Connecting screens
- **THEN** the music keeps playing without restarting

### Requirement: Music stops during a match
The menu music SHALL fade out and stop when a match starts, leaving the level silent. It SHALL fade back in when the player returns to the menu.

#### Scenario: Silence in the level
- **WHEN** the player starts a single player, host or join match
- **THEN** the menu music fades out and no music plays during the match

#### Scenario: Music returns
- **WHEN** the player leaves a match and returns to the main menu
- **THEN** the menu music fades back in

### Requirement: Music volume
The music volume SHALL be a value from 0 to 100. Changing it SHALL take effect immediately, without restarting the track. At 0 the music SHALL be inaudible.

#### Scenario: Volume applied live
- **WHEN** the player drags the music volume slider while the menu music is playing
- **THEN** the playing music gets quieter or louder as the slider moves, without a pause or restart

#### Scenario: Silence at zero
- **WHEN** the music volume is set to 0
- **THEN** no music is audible
