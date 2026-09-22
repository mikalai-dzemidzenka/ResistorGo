## ADDED Requirements

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

## MODIFIED Requirements

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
