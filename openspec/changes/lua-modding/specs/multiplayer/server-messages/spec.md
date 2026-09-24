## Purpose

Lets the server put a line of text in front of one player or everyone watching the match, shown in a short message feed over the level.

## ADDED Requirements

### Requirement: Server text messages
The server SHALL be able to send a line of text to one connected player or to every accepted client, fighters and spectators alike. A line SHALL be delivered on the reliable channel, in the order sent, and SHALL NOT delay snapshots or input. A line longer than 128 bytes SHALL be cut to 128 bytes on a character boundary, and characters the game's font cannot draw SHALL be shown as `?`. A client that receives a malformed text message SHALL ignore it.

#### Scenario: To one player
- **WHEN** the server sends "you are on team 2" to player 3
- **THEN** player 3 receives it and no other client does

#### Scenario: To everyone
- **WHEN** the server sends "round over" to everyone while two fighters and a spectator are connected
- **THEN** all three receive it

#### Scenario: Under loss
- **WHEN** three lines are sent over a link with 20 % packet loss
- **THEN** all three arrive, in the order sent

### Requirement: Message feed
A client in a networked match SHALL show the lines it receives in a feed over the level, newest at the bottom, at most six lines at once, each staying on screen for ten seconds. The feed SHALL be drawn at the window's resolution, not the level's pixel grid, SHALL stay on screen when the HUD is hidden, and SHALL be cleared when the client leaves the match.

#### Scenario: A line appears and goes
- **WHEN** a line arrives
- **THEN** it is shown at the bottom of the feed and is gone ten seconds later

#### Scenario: More than six
- **WHEN** eight lines arrive within a second
- **THEN** the last six are shown
