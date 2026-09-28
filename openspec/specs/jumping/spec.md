# jumping Specification

## Purpose
Covers leaving the ground and coming back down: the standing jump, the running jump, what a fighter can do while airborne, when it is falling rather than rising, and what it lands into.

## Requirements

### Requirement: A running fighter jumps into its own animation
A fighter that is running and presses jump SHALL begin the run-jump on that tick, playing the running-jump animation from its first frame. It SHALL NOT play the standing jump's separate wind-up.

The fighter SHALL leave the ground on the third frame of that animation (frame index 2), which is the frame it pushes off on, and SHALL stay on the ground for the two frames before it. The push-off SHALL therefore be part of the animation rather than a state of its own, and the fighter SHALL be committed from the tick it pressed jump: once the run-jump has begun it SHALL leave the ground when that frame arrives.

The standing jump SHALL be unchanged: a fighter that presses jump from a standstill still plays its wind-up before leaving the ground.

A run-jump SHALL take off with the same upward speed a standing jump does, so which jump a fighter used decides how it left the ground and not how high it goes.

#### Scenario: Jumping at a sprint
- **WHEN** a running fighter presses jump
- **THEN** it begins the run-jump on that tick, playing the running-jump animation, still on the ground

#### Scenario: Off the ground on the push-off frame
- **WHEN** a run-jump's animation reaches its third frame
- **THEN** the fighter is off the ground and moving upward, and it was on the ground for the two frames before it

#### Scenario: Committed once begun
- **WHEN** a fighter begins a run-jump and releases the jump button before the push-off frame
- **THEN** it still leaves the ground on that frame

#### Scenario: The standing jump is untouched
- **WHEN** a fighter presses jump from a standstill
- **THEN** it plays the wind-up first and leaves the ground when the wind-up finishes, as it did before

#### Scenario: Same height either way
- **WHEN** one fighter run-jumps and another jumps from a standstill
- **THEN** both leave the ground at the same upward speed

### Requirement: A run-jump costs half the pool
Taking a run-jump SHALL cost the fighter 100 energy — half of a full pool — spent as it leaves the ground. A fighter SHALL therefore get two run-jumps from a full pool and SHALL have to recover before a third. Spending it SHALL restart the wait before energy refills, exactly as spending on a swing does.

A fighter with less than 100 energy SHALL NOT run-jump. The input SHALL be refused: the fighter SHALL stay in the state it was in, it SHALL keep running, no energy SHALL be taken, and it SHALL NOT leave the ground. A refused run-jump SHALL raise the same refusal the game already raises for a swing that cannot be paid for, so the empty pool is shown to the player.

The standing jump SHALL remain free, and SHALL be available at any energy. Nothing else about moving SHALL cost energy.

#### Scenario: Two from a full pool
- **WHEN** a fighter at full energy run-jumps twice, with no time to recover in between
- **THEN** the first leaves it at 100 and the second at 0

#### Scenario: Not enough to leave the ground
- **WHEN** a running fighter with 99 energy presses jump
- **THEN** it does not leave the ground, its energy is still 99, and it goes on running

#### Scenario: The refusal is shown
- **WHEN** a running fighter without the energy for a jump presses jump
- **THEN** the refusal is raised, as it is for a swing that cannot be paid for

#### Scenario: The standing jump is free
- **WHEN** a fighter at 0 energy jumps from a standstill
- **THEN** it leaves the ground and its energy is still 0

#### Scenario: A jump delays the refill
- **WHEN** a fighter run-jumps
- **THEN** the wait before its energy starts refilling begins again, as it does after a swing

#### Scenario: Swinging is still affordable after one
- **WHEN** a fighter at full energy run-jumps once
- **THEN** it has 100 energy left and can still swing

### Requirement: A run-jump keeps the run's momentum
A run-jump SHALL carry the horizontal speed the fighter had when it left the ground. The jump SHALL NOT reset that speed, so a fighter jumping at a sprint travels further than one jumping at a walk.

#### Scenario: Speed carries into the air
- **WHEN** a fighter run-jumps at full running speed
- **THEN** its horizontal speed on the tick after take-off is the speed it was running at

#### Scenario: Faster run, longer jump
- **WHEN** two fighters run-jump, one at full speed and one at half
- **THEN** the faster one covers more ground before it lands

### Requirement: A fighter can steer while airborne
A fighter SHALL be able to move while airborne, in either jump and while falling: pressing left or right SHALL accelerate it as it does on the ground, and it SHALL be able to reverse the direction it is travelling. Steering SHALL NOT require the fighter to have jumped — a fighter that walked off a ledge SHALL steer as it falls.

#### Scenario: Steering mid-jump
- **WHEN** a fighter in the air holds a direction
- **THEN** its horizontal speed changes in that direction while it is held

#### Scenario: Turning around in the air
- **WHEN** a fighter that run-jumped to the right holds left for long enough
- **THEN** it travels left before it lands

#### Scenario: Steering while falling
- **WHEN** a fighter that walked off a ledge holds a direction as it falls
- **THEN** it moves in that direction

### Requirement: A fighter that is moving downward is falling
A fighter that is off the ground with a downward vertical speed SHALL be in the fall state and SHALL play the falling animation, whatever put it in the air: a run-jump past its apex, a standing jump past its apex, walking off a ledge, or being knocked off one. A fighter that is off the ground and still moving upward SHALL NOT be falling.

The fall SHALL be entered from any state that finds itself airborne and descending, rather than only from a jump, so no fighter is left in the air playing an animation it is no longer doing.

#### Scenario: Over the top of a jump
- **WHEN** a run-jump reaches its apex and begins to descend
- **THEN** the fighter is falling and plays the falling animation

#### Scenario: The standing jump falls too
- **WHEN** a standing jump reaches its apex and begins to descend
- **THEN** the fighter is falling and plays the falling animation, rather than holding the last frame of its wind-up

#### Scenario: Walking off a ledge
- **WHEN** a fighter walks off the edge of a platform without jumping
- **THEN** it is falling and plays the falling animation

#### Scenario: Still rising
- **WHEN** a fighter has just left the ground and is still moving upward
- **THEN** it is not falling and does not play the falling animation

#### Scenario: Knocked off an edge
- **WHEN** a hit's knockback carries a fighter off a ledge
- **THEN** it falls, and the hit's own reaction is unaffected

### Requirement: A falling fighter can still catch a ledge
A falling fighter SHALL be able to climb a ledge it reaches, on the same terms as a rising one. Falling SHALL NOT cost a fighter the climb.

#### Scenario: Catching the ledge on the way down
- **WHEN** a fighter overshoots a ledge and reaches it while falling
- **THEN** it climbs, exactly as it would have while rising

#### Scenario: Nothing to catch
- **WHEN** a falling fighter reaches no ledge
- **THEN** it keeps falling until it lands

### Requirement: Landing returns the fighter to what it was doing
A falling fighter that reaches the ground SHALL leave the fall state. If it is still moving, it SHALL run; if it has stopped, it SHALL stand. A landing SHALL NOT lock the fighter out of acting.

#### Scenario: Landing at speed
- **WHEN** a fighter lands while still travelling sideways
- **THEN** it runs on landing rather than stopping dead

#### Scenario: Landing stopped
- **WHEN** a fighter lands with no horizontal speed
- **THEN** it stands

#### Scenario: Acting out of a landing
- **WHEN** a fighter presses attack on the tick it lands
- **THEN** the swing comes out
