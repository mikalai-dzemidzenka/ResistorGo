## Purpose

How a fighter moves along the ground: position that keeps its fraction so slow
speeds really move, movement while swinging, and stopping when the keys are
let go.

## ADDED Requirements

### Requirement: Horizontal position keeps its fraction
A fighter's horizontal position SHALL advance by exactly its speed times the time step, keeping the fraction of a pixel from one tick to the next, so that any non-zero horizontal speed moves the fighter over time. A fighter SHALL be drawn at its position rounded to the nearest whole pixel. Being stopped by a wall, snapped to a ledge or placed by a clash SHALL still put the fighter at the position those rules give.

#### Scenario: Slow speeds move
- **WHEN** a fighter moves at a steady 20 px/s for one second, unobstructed
- **THEN** it has moved 20 px, give or take one

#### Scenario: The same walk everywhere
- **WHEN** a fighter walks for one second from a stand, once on each side of the level's origin and once in each direction on each side
- **THEN** the two walks in the same direction cover the same distance, whichever side of the origin they are on

#### Scenario: Drawn on whole pixels
- **WHEN** a fighter is at a position with a fraction of a pixel
- **THEN** its sprite is drawn at a whole pixel

### Requirement: A swinging fighter moves slowly
A fighter in an attack SHALL be able to move left or right under its own input at up to 20 px/s, and SHALL keep facing the way it faced when the attack started while doing so, whichever way it moves. This SHALL hold for bots and players alike. The dash each blow opens with SHALL be unchanged. Guarding everything SHALL still prevent moving, as it does outside an attack.

#### Scenario: Backing out of reach mid-swing
- **WHEN** two fighters stand just inside each other's reach, one swings, and the other swings back while holding the key away from its opponent
- **THEN** the second fighter moves away during its own swing, still facing its opponent

#### Scenario: Distance decides a swing
- **WHEN** a fighter whose own swing is already under way, guarding a height other than the one coming at it, stands 2 px inside an opponent's reach as the opponent starts a swing, and holds the key away from the opponent until the opponent's blow reaches its hit frame
- **THEN** the opponent's blow misses it; the same fighter holding no key is hit

#### Scenario: Pressing in
- **WHEN** a fighter swinging at an opponent just out of its reach holds the key toward it
- **THEN** it closes the gap slowly during the swing, at no more than 20 px/s

#### Scenario: Very slow
- **WHEN** a fighter holds a direction through a whole low swing
- **THEN** it moves no more than 6 px, plus the swing's own dash

#### Scenario: A bot creeps too
- **WHEN** a bot swings while its steering asks it to move toward its target
- **THEN** it moves toward the target at the same slow pace a player would

### Requirement: Letting go stops a walk
A fighter moving under its own power that has no movement key held SHALL come to a stop once its speed has fallen below 60 px/s, rather than drifting on. A speed given to the fighter by something else - a flinch's knockback, a stun's throw, a shove - SHALL NOT be stopped by this rule and SHALL run out under friction.

#### Scenario: A released walk stops short of a glide
- **WHEN** a fighter walks at full speed and the key is released
- **THEN** it stops once its speed falls below 60 px/s, about 7 px sooner than friction alone would stop it

#### Scenario: A knockback runs out
- **WHEN** a standing fighter with no key held is flinched
- **THEN** it keeps sliding until friction stops it, and is not stopped the moment its speed falls under 60 px/s
