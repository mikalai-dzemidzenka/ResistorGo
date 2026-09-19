## Purpose

Gives a standing fighter an answer to another on its back - a position where the
one behind can hit and the one in front cannot - by turning on it and throwing
it off to arm's length, facing it.

## ADDED Requirements

### Requirement: A standing fighter shoves what is on its back

A fighter that is standing still on the ground SHALL shove another fighter whose
body overlaps its own and which its swing cannot reach. A fighter that can reach
an overlapping fighter with a swing SHALL NOT shove it, so the shove never takes
the place of a swing.

#### Scenario: Crowded from behind

- **WHEN** a standing fighter has another overlapping it from behind, out of
  reach of its swing
- **THEN** the standing fighter SHALL shove

#### Scenario: Overlapping but in front

- **WHEN** a standing fighter has another overlapping it that its swing reaches
- **THEN** it SHALL NOT shove

#### Scenario: Only the fighter that cannot reach shoves

- **WHEN** one fighter stands on another's back, so the one behind can reach and
  the one in front cannot
- **THEN** the one in front SHALL shove and the one behind SHALL NOT

### Requirement: Only a standing fighter shoves

A shove SHALL begin only from standing. A fighter doing anything else - running,
attacking, blocking, recovering from a hit, jumping, falling, climbing, turning
over, or guarding at every height - SHALL NOT begin one.

#### Scenario: A chase does not set it off

- **WHEN** a running fighter has another overlapping it from behind
- **THEN** it SHALL NOT shove

#### Scenario: A shove comes off the feet

- **WHEN** a fighter is airborne
- **THEN** it SHALL NOT shove

#### Scenario: Guarding everything stands its ground

- **WHEN** a standing fighter is guarding at every height
- **THEN** it SHALL NOT shove

### Requirement: The shover turns on the fighter it shoves

A shove SHALL be thrown facing the fighter it moves, and SHALL move that fighter
away from the shover. When the shove has landed, the fighter that was moved SHALL
be in front of the shover.

#### Scenario: Turning to shove

- **WHEN** a fighter facing one way shoves a fighter behind it
- **THEN** it SHALL face the other way while the shove plays

#### Scenario: The crowder goes away from the shover

- **WHEN** a shove lands
- **THEN** the fighter moved SHALL end up further from the shover, on the side it
  was already on

### Requirement: A shove ends at arm's length

A shove SHALL move the fighter it lands on far enough that the two no longer
overlap, and not so far that the shover's swing cannot reach it, from anywhere a
shove can be thrown.

#### Scenario: The two end up apart

- **WHEN** a shove lands on a fighter anywhere in the position that sets a shove
  off
- **THEN** their bodies SHALL no longer overlap

#### Scenario: The fight stays in reach

- **WHEN** a shove has landed on a fighter anywhere in the position that sets a
  shove off
- **THEN** that fighter SHALL be within reach of the shover's swing

### Requirement: A shove moves ground and nothing else

A shove SHALL take no structure and no energy from the fighter it moves, SHALL
NOT count as a hit, and SHALL NOT interrupt what that fighter was doing.

#### Scenario: Nothing else changes hands

- **WHEN** a shove lands
- **THEN** the fighter moved SHALL keep its structure, its energy and its hit
  count, and SHALL NOT enter the state a hit puts it in

#### Scenario: A fighter mid-swing is still moved

- **WHEN** a shove lands on a fighter part-way through a swing
- **THEN** that fighter SHALL be moved, and SHALL go on swinging

#### Scenario: The dead are left where they lie

- **WHEN** a fighter is dead
- **THEN** it SHALL NOT be shoved, and SHALL NOT cause another fighter to shove

### Requirement: A shove cannot repeat on its own

A shove SHALL NOT be followed by another shove of the same fighter from the same
position. A shove cut short by a wall SHALL still leave the shover able to act.

#### Scenario: A wall stops the throw

- **WHEN** a fighter shoves another that is flush against a wall and cannot be
  moved
- **THEN** the shover SHALL NOT shove again from that position, and SHALL within
  a bounded number of ticks reach a state in which its movement, jump and attack
  inputs are read

### Requirement: A shove is interrupted by a hit

A shove SHALL be interrupted by a hit landing on the shover, which SHALL then
react to the hit as it would from standing.

#### Scenario: Hit while shoving

- **WHEN** a hit lands on a fighter that is shoving
- **THEN** that fighter SHALL react to the hit, and a shove that has not yet
  landed SHALL NOT land

### Requirement: A shove cannot be steered

While a shove plays, the fighter throwing it SHALL NOT move or turn under its own
control, and the shove SHALL end on its own animation.

#### Scenario: No movement during the shove

- **WHEN** a fighter is shoving and movement is pressed
- **THEN** the fighter SHALL NOT move under that input

#### Scenario: The shove ends

- **WHEN** the shove's animation finishes
- **THEN** the fighter SHALL return to standing

### Requirement: A shove is decided by the authority

A shove's displacement SHALL be resolved only in the authoritative simulation,
as a hit is, and SHALL land at most once per fighter reached. A client SHALL be
shown the shove and the ground it changed by the snapshot, and SHALL NOT move any
fighter by a shove of its own.

#### Scenario: A client moves nobody

- **WHEN** a predicting client simulates a fighter crowded from behind
- **THEN** no fighter's position SHALL change by a shove on that client

#### Scenario: A shove lands once

- **WHEN** a shove plays out in the authoritative simulation
- **THEN** it SHALL move each fighter it reaches at most once
