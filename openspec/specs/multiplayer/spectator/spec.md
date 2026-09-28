# multiplayer/spectator Specification

## Purpose
Lets people watch an online match without taking part: a spectator joins without a fighter, sees every fighter, and moves a free or following camera over the level without affecting the game.

## Requirements

### Requirement: Joining as a spectator
A client that joins as a spectator SHALL receive the match like any other client but SHALL NOT be given a fighter. It SHALL enter the level once its first snapshot arrives and display every fighter in the match, including bots, the same way remote fighters are displayed to players. No other client SHALL see any new fighter because a spectator joined.

#### Scenario: Spectator joins
- **WHEN** a client joins a match with two fighters as a spectator
- **THEN** the spectator sees both fighters, and both players still see exactly two fighters

#### Scenario: Bots are visible
- **WHEN** a spectator joins a server running one human fighter and two bots
- **THEN** the spectator sees three fighters

### Requirement: Spectators cannot influence the match
The server SHALL ignore any input command received from a spectator. Joining or leaving as a spectator SHALL NOT change the fighters in the match, the number of bots, or the player slots available to fighters.

#### Scenario: Forged commands
- **WHEN** a spectator client sends input commands to the server
- **THEN** no fighter's state changes because of them

#### Scenario: Spectator leaves
- **WHEN** a spectator quits
- **THEN** the fighters, bots and free player slots are the same as before it left

#### Scenario: Bots are not replaced
- **WHEN** a spectator joins a server whose player slots are all taken by one human and bots
- **THEN** no bot is removed

### Requirement: Idle spectator stays connected
A spectator SHALL stay connected for as long as its connection to the server is alive, even though it sends no input commands.

#### Scenario: Watching for a minute
- **WHEN** a spectator watches without pressing anything for 60 seconds
- **THEN** it is still connected and still receiving the match

### Requirement: Free camera
In free mode, the spectator SHALL move the camera over the level:
- WASD or the arrow keys pan at a constant on-screen speed, regardless of zoom;
- dragging with the right mouse button moves the view with the mouse, so the point under the cursor stays under the cursor;
- the mouse wheel zooms within the same limits players have.

Free mode SHALL NOT follow any fighter.

#### Scenario: Pan with keys
- **WHEN** a spectator in free mode holds D
- **THEN** the view scrolls right steadily and keeps scrolling past fighters without snapping to them

#### Scenario: Same speed at any zoom
- **WHEN** a spectator pans for one second at minimum zoom and then for one second at maximum zoom
- **THEN** the view moves the same distance on screen both times

#### Scenario: Drag
- **WHEN** a spectator drags with the right mouse button 100 pixels to the left
- **THEN** the level moves 100 pixels to the left on screen

### Requirement: Follow camera
In follow mode, the camera SHALL track one fighter the way a player's camera tracks their own fighter. Q and E SHALL switch to the previous and next fighter in a fixed order, wrapping around, and SHALL enter follow mode when pressed in free mode. Space or any pan input (keys or right-mouse drag) SHALL switch to free mode without moving the view. When the spectator first enters the level, the camera SHALL follow the first fighter to become visible. If the followed fighter leaves the match, the camera SHALL follow the next fighter, or switch to free mode if no fighter is left.

#### Scenario: Cycle through fighters
- **WHEN** a spectator in a match with three fighters presses E three times
- **THEN** the camera follows each fighter in turn and ends on the one it started with

#### Scenario: Back to free mode
- **WHEN** a spectator following a fighter presses A
- **THEN** the camera stops following and pans left from where it was

#### Scenario: Followed fighter leaves
- **WHEN** the fighter a spectator follows disconnects
- **THEN** the camera follows another fighter, and if none is left it switches to free mode

### Requirement: Spectator HUD
While spectating, the screen SHALL show "SPECTATING", the current camera mode, the fighter being followed (if any), and the key hints for the camera controls. The HUD SHALL be shown whether or not debug mode is on. The player's attack cursor SHALL NOT be shown.

#### Scenario: HUD in follow mode
- **WHEN** a spectator follows a fighter
- **THEN** the HUD shows "SPECTATING", follow mode, which fighter is followed, and the Q/E, Space, WASD, right-drag and wheel hints
