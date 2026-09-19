## Purpose

Fixes how far a swing reaches, so that the arm covers the same ground whichever
way a fighter faces, and so that a fighter standing on top of another is someone
it can hit rather than only shove.

## ADDED Requirements

### Requirement: A swing reaches the same distance in both facings

The distance a swing covers in front of a fighter SHALL NOT depend on which way
that fighter is facing. Reach is measured from the fighter's collision box, not
from the sprite frame it is drawn in, so the two facings mirror one another
exactly.

#### Scenario: Equal reach forward and backward

- **WHEN** a fighter facing right can just reach a target at some distance in
  front of it
- **THEN** the same fighter facing left SHALL just reach a target at that same
  distance in front of it

#### Scenario: Equal near edge

- **WHEN** a target is close enough that a fighter facing right can no longer
  reach it
- **THEN** a fighter facing left SHALL likewise fail to reach a target at that
  same distance in front of it

#### Scenario: Turning around does not change what is in range

- **WHEN** two fighters stand a fixed distance apart and both turn to face one
  another
- **THEN** each SHALL reach the other, or neither SHALL, rather than reach
  depending on who turned which way

### Requirement: A fighter standing on another can be hit

A fighter whose collision box exactly coincides with another's SHALL be within
reach of that other's swing, in either facing. Standing on someone is a position
a swing answers.

#### Scenario: Coincident fighters are in reach facing right

- **WHEN** two fighters occupy the same position and one faces right
- **THEN** the right-facing fighter SHALL reach the other

#### Scenario: Coincident fighters are in reach facing left

- **WHEN** two fighters occupy the same position and one faces left
- **THEN** the left-facing fighter SHALL reach the other

### Requirement: The shove only fires where no swing can land

A fighter SHALL shove only another fighter it overlaps and cannot reach with a
swing. The window in which that holds SHALL be the same width on either side of
the fighter.

#### Scenario: Symmetric crowding window

- **WHEN** a fighter overlaps another at some offset behind it and shoves rather
  than swinging
- **THEN** a fighter facing the other way SHALL shove at the mirrored offset,
  and at no other

#### Scenario: A wall clamp does not produce a shove that repeats forever

- **WHEN** a fighter stands flush against a wall and shoves another fighter whose
  displacement the wall cuts short
- **THEN** the fighter that was shoved SHALL end up somewhere its shover can
  reach with a swing, so the shove is not thrown again from the same position

#### Scenario: Neither fighter is locked out of acting

- **WHEN** two fighters are pressed together against a wall
- **THEN** each SHALL, within a bounded number of ticks, reach a state in which
  its movement, jump and attack inputs are read again
