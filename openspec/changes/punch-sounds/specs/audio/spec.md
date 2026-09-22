## ADDED Requirements

### Requirement: Punch sounds
Every landed hit SHALL play a punch sound, chosen at random from the punch sounds, on every machine that shows the match - in single player, when hosting, when joining and when spectating. A hit that flinches, stuns or kills SHALL sound alike. A swing that misses, a blow stopped by a guard or by block-all, a clash and a shove SHALL make no punch sound. Hits close together SHALL each be heard, overlapping rather than cutting one another off. While time manipulation shows a recorded frame no punch sound SHALL play. Punch sounds SHALL be played at the sound volume setting, and at 0 SHALL be inaudible.

#### Scenario: A punch lands
- **WHEN** a hit lands on a fighter
- **THEN** a punch sound plays

#### Scenario: Missed and blocked blows are silent
- **WHEN** a swing misses, or is blocked
- **THEN** no punch sound plays

#### Scenario: A flurry
- **WHEN** three hits land in quick succession
- **THEN** three punch sounds are heard, the later ones not cutting the earlier ones short

#### Scenario: Scrubbing is silent
- **WHEN** the player scrubs back and forth across a hit with time manipulation
- **THEN** no punch sound plays

#### Scenario: Muted
- **WHEN** the sound volume is set to 0
- **THEN** hits make no sound, and the music is unaffected
