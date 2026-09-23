## Purpose

How a fighter gets down through a floor rather than off its edge: the drop-through platform, pressing S on it to climb down, and the fall that follows.

## ADDED Requirements

### Requirement: A drop-through platform is floor from above and open from below
A drop-through platform SHALL hold up a fighter that stands, walks, runs or lands on it, exactly as an ordinary solid does. A fighter rising into it from below SHALL pass up through it without its head being stopped, and SHALL be held up by it once it comes down onto its top. A rising fighter SHALL be able to catch its ledge and climb onto it, exactly as with an ordinary solid. A fighter moving sideways SHALL be stopped by it exactly as by an ordinary solid. This SHALL hold for players, bots and the fighters a client predicts.

#### Scenario: Walked on
- **WHEN** a fighter walks across a drop-through platform
- **THEN** it stays on top of the platform, as it would on an ordinary solid

#### Scenario: Landed on
- **WHEN** a fighter jumps and comes down onto a drop-through platform from above
- **THEN** it lands on the platform

#### Scenario: Jumped up through
- **WHEN** a fighter jumps from the floor under a drop-through platform, high enough for its feet to clear the platform's top
- **THEN** it rises up through the platform and lands on top of it

#### Scenario: Not high enough
- **WHEN** a fighter jumps up into a drop-through platform but its feet never clear the top
- **THEN** it passes up into the platform, comes back down through it, and lands where it jumped from

#### Scenario: Ledge caught on the way up
- **WHEN** a fighter jumps at the edge of a drop-through platform, facing it
- **THEN** it catches the ledge and climbs onto the platform

### Requirement: S climbs down through a drop-through platform
A fighter that is standing or walking (idle) or running, on the ground, and whose footing is a drop-through platform SHALL start climbing down when S is pressed. S SHALL be an edge-triggered press: holding it SHALL NOT start a second climb-down. On footing that is not a drop-through platform, in the air, or in any other state (swinging, blocking, stunned, shoving, clashing, turning over, climbing, dead), the press SHALL do nothing and SHALL NOT be kept for later. Climbing down SHALL cost no energy. Bots SHALL NOT climb down.

#### Scenario: Standing on the platform
- **WHEN** a fighter stands on a drop-through platform and S is pressed
- **THEN** it starts climbing down on that tick

#### Scenario: Running onto it
- **WHEN** a fighter runs across a drop-through platform and S is pressed while it is over the platform
- **THEN** it stops and starts climbing down

#### Scenario: Ordinary floor
- **WHEN** a fighter stands on an ordinary solid and S is pressed
- **THEN** nothing happens and the fighter stays standing

#### Scenario: Not while swinging
- **WHEN** a fighter on a drop-through platform presses S during a swing
- **THEN** the swing goes on, and the fighter does not climb down when the swing ends

#### Scenario: Not in the air
- **WHEN** a jumping or falling fighter presses S above a drop-through platform
- **THEN** nothing happens, and it lands on the platform as usual

### Requirement: The climb-down plays in place
While climbing down, the fighter SHALL play the climb-down animation once, from its first frame to its last, drawn where the fighter stands. The fighter SHALL NOT move sideways, turn or act on movement, attack or jump input for as long as the animation plays, and SHALL stay on top of the platform until it ends. The climb-down animation SHALL be drawn on every client and spectator that sees the fighter, not only by the player climbing down.

#### Scenario: Keys held while climbing down
- **WHEN** a fighter climbing down holds a movement key and presses attack and jump
- **THEN** it neither moves, turns, swings nor jumps, and the climb-down plays to its last frame

#### Scenario: Seen online
- **WHEN** one player climbs down in a networked match
- **THEN** the other players and any spectator see the climb-down animation play on that fighter

### Requirement: A climb-down ends in a fall below the platform
When the climb-down animation ends, the fighter SHALL be below the top of the platform it climbed down through, playing the fall animation and falling under gravity from rest. The fall SHALL pick up where the last climb-down frame leaves the fighter's body, without a visible jump of the sprite. From then on the fighter SHALL fall as any fighter does, steering and landing on whatever solid or drop-through platform is below, except that it SHALL NOT catch any ledge until it has landed. A drop-through platform usually stands in a row of solid ones, and a ledge beside it, level with the height the fighter left, would put it straight back up.

If the floor below is closer than the distance the climb-down lowers the fighter, the fighter SHALL be stood on that floor rather than placed inside or past it.

#### Scenario: Handed to the fall
- **WHEN** a fighter's climb-down animation reaches its end
- **THEN** on the next tick it is below the platform's top, playing the fall animation

#### Scenario: Solid tiles beside the platform
- **WHEN** a fighter climbs down through a drop-through tile in a row of solid tiles at the same height, facing either way
- **THEN** it falls past the solid tiles without catching their ledges and lands on the floor below

#### Scenario: Steered on the way down
- **WHEN** a fighter that has climbed down holds a movement key while it falls
- **THEN** it moves that way as it falls

#### Scenario: Lands below
- **WHEN** a fighter climbs down through a drop-through platform with open floor some distance below it
- **THEN** it falls to that floor and lands on it

#### Scenario: Floor right underneath
- **WHEN** a fighter climbs down through a drop-through platform with a solid floor closer below it than the climb-down lowers a fighter
- **THEN** it ends up standing on that floor, not inside it or below it

#### Scenario: Through a stack
- **WHEN** a fighter climbs down through one drop-through platform with a second drop-through platform further below it
- **THEN** it lands on the second platform, and can climb down through that one too

### Requirement: A fighter below a drop-through platform's top falls past it
A fighter that is falling and whose feet are below the top of a drop-through platform SHALL NOT land on that platform, SHALL NOT have its fall stopped by it, and SHALL NOT catch its ledge. This SHALL follow from where the fighter is and which way it is moving alone, so that a client and the server always agree on it. It SHALL NOT change how an ordinary solid treats a falling fighter.

#### Scenario: Not caught by the platform it came through
- **WHEN** a fighter has climbed down through a drop-through platform and is falling below it
- **THEN** it does not land back on the platform, and does not grab its ledge

#### Scenario: Agreed online
- **WHEN** a player climbs down through a drop-through platform in a networked match
- **THEN** the player's own view and the server both have the fighter falling past the platform, and the server does not correct the player back up onto it

### Requirement: A stun interrupts a climb-down
A fighter that is stunned while climbing down SHALL stop climbing down and SHALL be stunned on top of the platform, exactly as it would be from standing. A hit that only makes it flinch SHALL NOT interrupt the climb-down. A fighter killed while climbing down SHALL die on top of the platform.

#### Scenario: Stunned mid-climb
- **WHEN** a fighter halfway through its climb-down is stunned
- **THEN** it plays the stun on top of the platform and does not fall through

#### Scenario: Flinch does not stop it
- **WHEN** a fighter climbing down is hit by a blow that only makes it flinch
- **THEN** it flickers and still falls through the platform when the climb-down ends
