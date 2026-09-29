## ADDED Requirements

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
