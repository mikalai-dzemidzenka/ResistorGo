# level-collision Specification

## Purpose
Covers how fighters meet the level's collision rectangles from the side: walls that stop them, small steps they walk up, and ledges that still catch against a wall.

## Requirements

### Requirement: The side of a collision rectangle stops a fighter
A fighter moving sideways SHALL NOT pass through the side of a collision rectangle that is a wall to it. It SHALL stop with the side of its collision box against the rectangle's side, and its sideways speed SHALL become zero. This SHALL hold however the fighter is moving: walking, running, steering in the air, thrown by a hit, or dashing forward as it starts a swing. It SHALL hold at every speed a fighter can reach, however thin the rectangle. Moving away from a wall SHALL never be stopped by it.

A rectangle is a wall to a fighter when some part of it is level with the fighter's body, between its head and a step's height above its feet.

#### Scenario: Walking into a wall
- **WHEN** a fighter walks toward a collision rectangle that stands taller than a step
- **THEN** it stops with its collision box touching the rectangle's side, and does not end up on top of it

#### Scenario: Running into a wall
- **WHEN** a fighter runs at full speed into a wall
- **THEN** it stops against the wall's side, its sideways speed is zero, and it is no longer running

#### Scenario: A thin wall
- **WHEN** a fighter runs at full speed at a wall one pixel wide
- **THEN** it stops against it and does not come out on the other side

#### Scenario: Thrown into a wall
- **WHEN** a hit's knockback carries a fighter into a wall
- **THEN** it stops against the wall, and the hit's own reaction is unaffected

#### Scenario: A swing's dash
- **WHEN** a fighter standing against a wall and facing it starts a swing
- **THEN** it does not move into the wall, and the swing comes out

#### Scenario: Steering into a wall in the air
- **WHEN** a fighter in the air steers into the side of a wall
- **THEN** it stops against the wall's side and goes on rising or falling

#### Scenario: Walking away
- **WHEN** a fighter standing against a wall walks away from it
- **THEN** it moves away as it would anywhere else

### Requirement: A small step is walked up
A collision rectangle whose top is no more than 4 pixels above a fighter's feet SHALL NOT stop that fighter. A fighter walking or running onto it SHALL end up standing on it without jumping. A rectangle whose top is more than 4 pixels above the fighter's feet SHALL be a wall to it. Neighbouring floor rectangles with the same top SHALL be crossed without any stop.

#### Scenario: A four-pixel step
- **WHEN** a fighter walks onto a collision rectangle whose top is 4 pixels above the floor it stands on
- **THEN** it ends up standing on that rectangle, without jumping and without stopping

#### Scenario: A five-pixel rise
- **WHEN** a fighter walks toward a collision rectangle whose top is 5 pixels above the floor it stands on
- **THEN** it stops against the rectangle's side

#### Scenario: A level floor
- **WHEN** a fighter runs across a floor made of neighbouring tiles whose tops are level
- **THEN** it runs across without stopping, as it does now

### Requirement: A ledge is still caught against a wall
A fighter that is stopped by a wall while jumping or falling SHALL be able to catch the wall's top edge and climb it, on the same terms as a fighter reaching a ledge today. Being stopped by the wall SHALL NOT cost it the climb.

#### Scenario: Jumping at a wall
- **WHEN** a fighter facing a wall jumps against it, and reaches its top edge
- **THEN** it catches the ledge and climbs onto the wall

#### Scenario: Falling past a wall
- **WHEN** a fighter falling beside a wall and facing it drops past the wall's top edge
- **THEN** it catches the ledge and climbs, as it does at a tile's edge now

### Requirement: A ledge needs room to stand on
A fighter SHALL catch a ledge only if its body fits standing on top of it, where the climb would put it. A ledge with a solid collision rectangle over it, in the space the fighter would stand in, SHALL NOT be caught, and the fighter SHALL go on rising or falling as it would beside any wall. A drop-through rectangle SHALL NOT take that room away. Catching one ledge SHALL NOT carry the fighter on to catch another above it in the same moment; a fighter reaching up a wall of stacked rectangles SHALL end up on the top of the wall, not inside any of it.

#### Scenario: The stem of a T
- **WHEN** a fighter jumps beside a wall that has a ceiling resting on its top, reaching past the ceiling's edge on the fighter's side
- **THEN** it does not catch the wall's ledge, is stopped by the ceiling, and does not end up on top of or inside the ceiling

#### Scenario: A wall two tiles tall
- **WHEN** a fighter facing a wall of two stacked rectangles, with nothing above it, jumps and its head reaches the upper rectangle
- **THEN** it catches the top edge and climbs onto the top of the wall

#### Scenario: A drop-through platform over a ledge
- **WHEN** a fighter jumps at a ledge whose standing space is crossed by a drop-through platform
- **THEN** it catches the ledge and climbs as it would with nothing there

#### Scenario: An open ledge
- **WHEN** a fighter reaches a ledge with nothing above it
- **THEN** it catches the ledge and climbs, as it does now

### Requirement: A fighter is never pushed out of a rectangle
A fighter whose collision box already overlaps a collision rectangle SHALL NOT be moved out of it by that rectangle. It SHALL be free to move in any direction that does not take it further into the rectangle, and SHALL NOT move further in.

#### Scenario: Spawned overlapping
- **WHEN** a fighter is placed so that its collision box overlaps the side of a rectangle
- **THEN** it is not moved sideways by the rectangle, and it can walk out of it

#### Scenario: No deeper
- **WHEN** a fighter overlapping the side of a wall walks further toward the wall
- **THEN** it does not move further into it

### Requirement: Walls work the same online
Walls SHALL stop every fighter the same way in single player and online: the fighters players control, the fighters of bots, and the fighters a client predicts. A client's prediction of its own fighter against a wall SHALL agree with the server.

#### Scenario: Online against a wall
- **WHEN** a player in a networked match runs into a wall and stands against it
- **THEN** the player's fighter stops against the wall on both the client and the server, and the server does not correct the client for it
