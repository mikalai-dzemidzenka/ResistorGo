## ADDED Requirements

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
