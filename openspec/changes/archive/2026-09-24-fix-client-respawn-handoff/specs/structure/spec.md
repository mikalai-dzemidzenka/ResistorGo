## ADDED Requirements

### Requirement: What a dying fighter's own client shows
A client SHALL show its own fighter dying exactly as it shows anyone else's: while the authority is still reporting that fighter, the client SHALL play the fall and the flicker from what it is told. When the authority stops reporting it, the client SHALL take it out of its world as it does any other fighter that has gone, and the player SHALL have no fighter until it is given one.

A fighter's flicker SHALL be bounded whether or not the fighter is removed on time. Past the end of its flicker the fighter SHALL stay hidden rather than keep blinking, so a client waiting to be told something never presents a strobing sprite.

#### Scenario: Death seen from the inside
- **WHEN** a player's own fighter dies in a networked match
- **THEN** that player sees it fall and flicker, and then sees it gone

#### Scenario: Nothing is left blinking
- **WHEN** a client holds a dead fighter for longer than its flicker lasts, whatever the reason
- **THEN** the fighter stays hidden from that point instead of blinking on and off

#### Scenario: The gap between
- **WHEN** a player's fighter has gone and the respawn has not arrived
- **THEN** the player has no fighter on screen and no fighter under its control, and the view stays where it was

#### Scenario: Single player hands over too
- **WHEN** a player's fighter dies and respawns in an offline match
- **THEN** the new fighter is the player's own: it takes the player's input, and the effects that follow the player's fighter follow it
