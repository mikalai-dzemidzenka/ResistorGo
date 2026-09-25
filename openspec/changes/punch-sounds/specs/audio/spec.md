## ADDED Requirements

### Requirement: Punch sounds
Every landed hit SHALL play a punch sound, chosen at random from the punch sounds, on every machine that shows the match - in single player, when hosting, when joining and when spectating. A hit that flinches, stuns or kills SHALL sound alike. A swing that misses, a blow stopped by a guard or by block-all and a shove SHALL make no punch sound. A clash SHALL play a punch at each of its impacts - where a blow of the exchange meets the other fighter - once per impact however many fighters are in it, and none as it starts. Hits close together SHALL each be heard, overlapping rather than cutting one another off. While time manipulation shows a recorded frame no punch sound SHALL play. Punch sounds SHALL be played at the sound volume setting, and at 0 SHALL be inaudible.

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

### Requirement: Block sounds
Every blow stopped by a guard or by block-all SHALL play the block sound, on every machine that shows the match - in single player, when hosting, when joining and when spectating. A hit that lands, a swing that misses, a clash and a shove SHALL make no block sound. While time manipulation shows a recorded frame no block sound SHALL play. Block sounds SHALL be played at the sound volume setting.

#### Scenario: A blow is blocked
- **WHEN** a swing meets a guard at its height, or block-all
- **THEN** the block sound plays, and no punch sound

#### Scenario: A hit is no block
- **WHEN** a hit lands on a fighter
- **THEN** no block sound plays

#### Scenario: Scrubbing is silent
- **WHEN** the player scrubs back and forth across a block with time manipulation
- **THEN** no block sound plays
