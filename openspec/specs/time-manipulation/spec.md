# time-manipulation Specification

## Purpose

Lets a single player stop the match, travel through the last few seconds of it with the mouse, and pick the match back up from any moment they find.

## Requirements

### Requirement: Entering and leaving the mode
In a single-player match, R SHALL toggle time-manipulation mode: pressing it while the mode is off turns it on, pressing it while the mode is on turns it off. Holding R SHALL have no effect beyond the press that started it. Leaving the mode SHALL resume the match from the frame currently shown. The mode SHALL NOT be available in a network match, where R SHALL do nothing.

#### Scenario: Press to enter
- **WHEN** the player presses R during a single-player match
- **THEN** time-manipulation mode turns on and the match pauses on the current frame

#### Scenario: Press again to leave
- **WHEN** the player presses R while the mode is on
- **THEN** the mode turns off and the match continues from the frame that was on screen, at normal speed

#### Scenario: Holding does not rewind
- **WHEN** the player presses R and keeps it held without moving the mouse
- **THEN** the match stays paused on the frame it entered on and time does not move in either direction

#### Scenario: Not available online
- **WHEN** the player presses R while connected to a server, as a fighter or as a spectator
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
While the mode is on, horizontal mouse movement SHALL drive time. Movement to the right SHALL run time forward and movement to the left SHALL rewind it. Time SHALL move only while the mouse is moving; when the mouse is still, the match SHALL be paused on the current frame. The scrub rate SHALL be fixed and identical in both directions: how fast the mouse moves SHALL NOT change how fast time moves, and the rate SHALL NOT depend on the frame rate. Vertical mouse movement SHALL NOT move time.

#### Scenario: Right runs time forward
- **WHEN** the player moves the mouse to the right while the mode is on, after having rewound
- **THEN** the match replays forward from the current frame for as long as the mouse keeps moving right

#### Scenario: Left rewinds
- **WHEN** the player moves the mouse to the left while the mode is on
- **THEN** the match plays backwards from the current frame for as long as the mouse keeps moving left

#### Scenario: Stopping the mouse pauses
- **WHEN** the player stops moving the mouse while scrubbing in either direction
- **THEN** time stops on the frame reached and the match stays there

#### Scenario: Speed of the hand does not matter
- **WHEN** the player sweeps the mouse left quickly and then drags it left slowly, each for one second
- **THEN** both cover the same number of recorded frames

#### Scenario: Jitter does not move time
- **WHEN** the mouse rests on the table and reports sub-pixel sensor noise
- **THEN** time does not creep forward or backward

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
When the player leaves the mode on a frame earlier than the live edge, the match SHALL continue from that frame and the recorded frames that were ahead of it SHALL be discarded. A later time-manipulation session SHALL only be able to scrub back through frames recorded up to the moment it is entered, including those recorded since the match resumed.

#### Scenario: Resuming from the past discards what came after
- **WHEN** the player rewinds two seconds, leaves the mode, plays for one second and enters the mode again
- **THEN** scrubbing forward stops at the frame where they entered the second time, and the discarded frames from the first timeline are never replayed

#### Scenario: History regrows after resuming
- **WHEN** the player resumes from a rewound frame and then plays on
- **THEN** the frames played after resuming are recorded and can be rewound through in a later session

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
