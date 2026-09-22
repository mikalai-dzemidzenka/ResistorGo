# time-manipulation Specification

## Purpose

Lets a single player stop the match, travel through the last minute of it with the mouse, and pick the match back up from any moment they find.

## Requirements

### Requirement: Entering and leaving the mode
In a single-player match, T SHALL toggle time-manipulation mode: pressing it while the mode is off turns it on, pressing it while the mode is on turns it off. Holding T SHALL have no effect beyond the press that started it. Leaving the mode SHALL resume the match from the frame currently shown. The mode SHALL NOT be available in a network match, where T SHALL do nothing.

#### Scenario: Press to enter
- **WHEN** the player presses T during a single-player match
- **THEN** time-manipulation mode turns on and the match pauses on the current frame

#### Scenario: Press again to leave
- **WHEN** the player presses T while the mode is on
- **THEN** the mode turns off and the match continues from the frame that was on screen, at normal speed

#### Scenario: Holding does not rewind
- **WHEN** the player presses T and keeps it held without moving the mouse
- **THEN** the match stays paused on the frame it entered on and time does not move in either direction

#### Scenario: Not available online
- **WHEN** the player presses T while connected to a server, as a fighter or as a spectator
- **THEN** nothing happens and the match continues uninterrupted

### Requirement: Entering pauses the match
While time-manipulation mode is on and the player is not scrubbing, the simulation SHALL be fully stopped: fighter physics, the fighter state machine, animation, bot decisions, attack resolution and hit-stop SHALL NOT advance. The pause SHALL NOT bank up simulation time — leaving the mode SHALL NOT produce a burst of catch-up steps for the time spent in it, however long that was.

#### Scenario: Nothing moves while paused
- **WHEN** the mode has been on for several seconds without mouse movement
- **THEN** every fighter is in exactly the position, animation frame and state it was in when the mode was entered, and no attack has landed in the meantime

#### Scenario: No catch-up on leaving
- **WHEN** the player leaves the mode after holding it for 10 seconds
- **THEN** the match resumes at normal speed, with no fast-forward or jump in fighter positions

### Requirement: Mouse direction scrubs time
While the mode is on, horizontal mouse movement SHALL drive time. Movement to the right SHALL run time forward and movement to the left SHALL rewind it. Time SHALL move only while the mouse is moving; when the mouse is still, the match SHALL be paused on the current frame. The scrub rate SHALL start at twice the speed the match was recorded at, and SHALL rise the longer the mouse keeps moving the same way, reaching sixteen times that speed after two seconds of continuous movement and staying there. The rate SHALL fall back to its starting speed when the mouse reverses direction, or once it has been still for a tenth of a second; a still frame shorter than that SHALL pause time on that frame but SHALL NOT lower the rate. How fast the mouse moves SHALL NOT change how fast time moves, only for how long it has kept moving; the rate SHALL be the same in both directions, and SHALL NOT depend on the frame rate. Vertical mouse movement SHALL NOT move time.

#### Scenario: Right runs time forward
- **WHEN** the player moves the mouse to the right while the mode is on, after having rewound
- **THEN** the match replays forward from the current frame for as long as the mouse keeps moving right

#### Scenario: Left rewinds
- **WHEN** the player moves the mouse to the left while the mode is on
- **THEN** the match plays backwards from the current frame for as long as the mouse keeps moving left

#### Scenario: Stopping the mouse pauses
- **WHEN** the player stops moving the mouse while scrubbing in either direction
- **THEN** time stops on the frame reached and the match stays there

#### Scenario: A short drag is slow
- **WHEN** the player drags the mouse left for a quarter of a second from rest
- **THEN** the match rewinds by less than a second of play, slowly enough to pick out a single frame

#### Scenario: Holding the drag speeds it up
- **WHEN** the player keeps moving the mouse left for four seconds from rest
- **THEN** time runs backwards faster and faster for the first two seconds and at sixteen times the recorded speed for the last two, reaching a minute back within about five seconds

#### Scenario: Reversing starts slow again
- **WHEN** the player has rewound for three seconds and then moves the mouse right
- **THEN** the replay forward starts at twice the recorded speed, not sixteen

#### Scenario: A pause starts slow again
- **WHEN** the player rewinds for three seconds, holds the mouse still for half a second, and rewinds again
- **THEN** time stays still during the pause, and the second rewind starts at twice the recorded speed

#### Scenario: A dropped mouse report does not
- **WHEN** the mouse, moving steadily left, reports no movement on one rendered frame in the middle of a three-second drag
- **THEN** time does not move on that frame, and the rewind carries on at the speed it had reached

#### Scenario: Speed of the hand does not matter
- **WHEN** the player sweeps the mouse left quickly for one second from rest, and on another occasion drags it left slowly for one second from rest
- **THEN** both cover the same number of recorded frames

#### Scenario: The same in both directions
- **WHEN** the player rewinds for two seconds from rest, and on another occasion replays forward for two seconds from rest with enough recorded ahead
- **THEN** both cover the same number of recorded frames

#### Scenario: Frame rate does not matter
- **WHEN** the same three-second drag is made at 60 frames a second and at 150
- **THEN** both cover the same number of recorded frames

#### Scenario: Jitter does not move time
- **WHEN** the mouse rests on the table and reports sub-pixel sensor noise
- **THEN** time does not creep forward or backward, and the rate stays at its starting speed

#### Scenario: Vertical movement is ignored
- **WHEN** the player moves the mouse straight up or down while the mode is on
- **THEN** time does not move

### Requirement: Forward scrubbing is replay, bounded by the live edge
Forward scrubbing SHALL replay frames that were recorded before the mode was entered. It SHALL NOT simulate new frames at any speed. When the newest recorded frame is reached, forward scrubbing SHALL stop there and the match SHALL stay paused on it, whatever the mouse keeps doing, until the player leaves the mode or rewinds again.

#### Scenario: Stops at the newest recorded frame
- **WHEN** the player rewinds two seconds and then moves the mouse right continuously for ten seconds
- **THEN** the match replays forward to the frame it was on when the mode was entered and stays paused there

#### Scenario: The world does not run ahead
- **WHEN** the match is paused at the live edge and the player keeps moving the mouse right
- **THEN** no fighter moves, no bot acts and no new frame is recorded

### Requirement: Rewinding is bounded by the oldest recorded frame
Rewinding SHALL stop at the oldest frame still held in the recording and stay there, however much further the player scrubs left. Scrubbing SHALL never read outside the recorded history.

#### Scenario: Stops at the oldest frame
- **WHEN** the player moves the mouse left continuously for far longer than the recording is long
- **THEN** the match reaches the oldest recorded frame and stays paused on it, without error

#### Scenario: A short match has little to rewind
- **WHEN** the player enters the mode one second into a match and scrubs left continuously
- **THEN** rewinding stops at the first recorded frame of the match

### Requirement: Leaving rewrites the future
When the player leaves the mode on a frame earlier than the live edge, the match SHALL continue from that frame exactly as recorded, as if play had stopped on it: every fighter that was in the world then SHALL be in it again with the state it had then, every fighter that was not SHALL be gone, and fighters waiting to respawn or to be taken out SHALL be waiting as they were then. Nothing that happened after that frame SHALL carry over into the resumed match — not damage, deaths, stuns, attacks, spent energy, bot decisions or places at a fight. The recorded frames that were ahead of it SHALL be discarded. A later time-manipulation session SHALL only be able to scrub back through frames recorded up to the moment it is entered, including those recorded since the match resumed. The resumed match need not repeat the discarded future: the bots may choose differently than they did.

#### Scenario: Resuming from the past discards what came after
- **WHEN** the player rewinds two seconds, leaves the mode, plays for one second and enters the mode again
- **THEN** scrubbing forward stops at the frame where they entered the second time, and the discarded frames from the first timeline are never replayed

#### Scenario: History regrows after resuming
- **WHEN** the player resumes from a rewound frame and then plays on
- **THEN** the frames played after resuming are recorded and can be rewound through in a later session

#### Scenario: Resuming before a kill undoes it
- **WHEN** the player beats a bot in single player, rewinds to a frame where it is still standing, and leaves the mode
- **THEN** the bot is alive and fighting with the structure it had on that frame, and does not fall

#### Scenario: Resuming before the player's death undoes it
- **WHEN** the player's fighter dies, respawns, and the player rewinds to a frame before the death and leaves the mode
- **THEN** the player controls the fighter that had not yet died, with the structure it had then, and no respawn is pending

#### Scenario: Resuming before a hit undoes the damage
- **WHEN** a fighter takes a hit and the player rewinds to a frame before it landed and leaves the mode
- **THEN** the fighter has the structure, energy and hit count it had on that frame, and is not stunned unless it was then

#### Scenario: A swing in progress resumes where it was
- **WHEN** the player leaves the mode on a frame in the middle of a bot's attack
- **THEN** the attack continues from that point of its animation and can land, as it would have from that frame

#### Scenario: The line at a fight is as it was
- **WHEN** the player leaves the mode on a frame where two bots held the places at the player and a third was waiting
- **THEN** the same two bots hold the places and the third is waiting, regardless of how the line had changed since

### Requirement: Gameplay input is suppressed while the mode is on
While time-manipulation mode is on, fighter input SHALL NOT be applied: movement, jumping, attacking and stance changes SHALL have no effect on the match until the mode is left. The mouse SHALL drive only the scrub. The crosshair MAY keep following the mouse. Input given while the mode is on SHALL NOT be held over and acted on once the match resumes.

#### Scenario: Attacking does nothing while scrubbing
- **WHEN** the player clicks the attack button while the mode is on
- **THEN** no attack is queued or resolved, and the fighter's animation stays on the scrubbed frame

#### Scenario: Movement keys do nothing while scrubbing
- **WHEN** the player holds a movement key while the mode is on
- **THEN** the fighter does not move and no input is buffered for when the mode is left

#### Scenario: Clicks do not carry over
- **WHEN** the player clicks the attack button several times while the mode is on and then leaves it
- **THEN** the fighter does not attack on leaving, and the match resumes with no action pending

### Requirement: A recorded frame is the whole match
Every recorded frame SHALL hold the single-player match as it stood on that frame: which fighters were in the world, and for each its position, motion, facing, pose, state, structure, energy, stun, attack under way, buffered input, what a bot had in mind, and which fighters held the places at a fight; together with the fighters waiting to respawn and when, and the fighters dead and waiting to be taken out. While the mode is on, the frame on screen SHALL show exactly the fighters that were in the world on that frame, each as it was then.

#### Scenario: A removed fighter is back on its frame
- **WHEN** the player dies, is taken out of the world and respawns, and then scrubs back to a frame before the death
- **THEN** the fighter that died is shown where it stood on that frame, and the respawned fighter is not shown

#### Scenario: A fighter is absent before it existed
- **WHEN** the player scrubs back to a frame recorded before a fighter came into the world
- **THEN** that fighter is not shown on that frame, anywhere

#### Scenario: A body on the floor stands again on an earlier frame
- **WHEN** a bot is beaten in single player and the player scrubs back to a frame before its structure broke
- **THEN** the bot is shown standing and fighting as it was on that frame, not lying down

#### Scenario: Health bars follow the frame
- **WHEN** the player scrubs back to a frame before a fighter took a hit
- **THEN** that fighter's structure on the frame is what it was then, not what it is at the live edge

### Requirement: The recording holds the last minute
The recording SHALL hold every fixed step of the last 60 seconds of the single-player match, or the whole match when it is shorter. A frame older than that SHALL be dropped as each new one is recorded, and SHALL NOT be reachable by scrubbing. Recording SHALL NOT slow the match down, and a match played for longer than the recording holds SHALL NOT use more memory for it than one that has just filled it.

#### Scenario: A minute back
- **WHEN** the player has played for two minutes, enters the mode and rewinds as far as it goes
- **THEN** the match shown is the one from 60 seconds before the mode was entered

#### Scenario: Nothing older
- **WHEN** the player rewinds as far as the recording goes after a long match
- **THEN** rewinding stops at the frame 60 seconds back, and nothing from before it is shown

#### Scenario: Memory stops growing
- **WHEN** a single-player match runs for ten minutes
- **THEN** the memory the recording uses after ten minutes is the same as after one
