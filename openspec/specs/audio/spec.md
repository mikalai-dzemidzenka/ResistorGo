# audio Specification

## Purpose

Covers the game's sound: opening the audio device, the music that plays on the main menu, and how the player's volume setting is applied to it.

## Requirements

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

### Requirement: Music volume
The music volume SHALL be a value from 0 to 100. Changing it SHALL take effect immediately, without restarting the track. At 0 the music SHALL be inaudible.

#### Scenario: Volume applied live
- **WHEN** the player drags the music volume slider while the menu music is playing
- **THEN** the playing music gets quieter or louder as the slider moves, without a pause or restart

#### Scenario: Silence at zero
- **WHEN** the music volume is set to 0
- **THEN** no music is audible

### Requirement: Level music
While a match is on, a level track SHALL loop in place of silence: in single player, when hosting and when joining, and for a spectator. It SHALL start from its beginning each time a match starts, fading in while the menu music fades out, and SHALL fade out and stop when the player returns to the menu, while the menu music fades back in. While the match is held - time manipulation on, or edit mode - the level track SHALL keep playing at a low volume and SHALL return to full volume when the match resumes; a hold, a rewind or restarting the level SHALL NOT pause or restart it. The level track SHALL follow the same music volume setting as the menu music, and at 0 SHALL be inaudible.

#### Scenario: Music in the fight
- **WHEN** the player starts a single-player match from the menu
- **THEN** the menu music fades out and the level track fades in from its beginning, and keeps looping for as long as the match goes on

#### Scenario: Back to the menu
- **WHEN** the player leaves a match for the menu
- **THEN** the level track fades out and the menu music fades back in

#### Scenario: A new match starts the track over
- **WHEN** the player leaves a match partway through the level track and starts another
- **THEN** the level track starts again from its beginning

#### Scenario: Ducked while held
- **WHEN** the player turns time manipulation on and scrubs, then turns it off
- **THEN** the level track plays on quietly while the mode is on, then comes back up to full volume, without having stopped or restarted

#### Scenario: A restart keeps the music
- **WHEN** the player restarts the level with R
- **THEN** the level track carries on where it was

#### Scenario: Volume applies
- **WHEN** the music volume is set to 0
- **THEN** the level track is inaudible during a match
