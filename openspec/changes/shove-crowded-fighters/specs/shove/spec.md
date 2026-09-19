## Purpose

Keeps two fighters from standing in one another: whichever of them is on its
feet and free turns on the other, shoves it off to arm's length and stuns it,
doing it no damage.

## ADDED Requirements

### Requirement: Any overlap is shoved

A fighter on the ground and free to act SHALL shove another fighter on the
ground whose body overlaps its own, whether that fighter is in front of it,
behind it or on the same spot. Only one of the two SHALL shove: the first to
qualify shoves, and the stun it throws SHALL keep the other from shoving back.

#### Scenario: Overlapping from behind

- **WHEN** a free fighter has another on the ground overlapping it from behind
- **THEN** it SHALL shove

#### Scenario: Overlapping in front

- **WHEN** a free fighter has another on the ground overlapping it in front
- **THEN** it SHALL shove

#### Scenario: Not both

- **WHEN** two free fighters overlap
- **THEN** exactly one of them SHALL shove

#### Scenario: Not the one in the air

- **WHEN** a fighter in the air passes through another on the ground
- **THEN** it SHALL NOT be shoved until it has landed

### Requirement: Only a fighter on its feet and free shoves

A shove SHALL begin only from standing, walking or running on the ground. A
fighter swinging, blocking, stunned, flinching from a hit, in the air, climbing,
turning over, or guarding at every height SHALL NOT begin one; it finishes what
it is doing first.

#### Scenario: Standing, walking and running shove

- **WHEN** a fighter standing, walking or running has another overlapping it
- **THEN** it SHALL shove

#### Scenario: A committed fighter finishes first

- **WHEN** a fighter mid-swing, blocking, stunned or flinching has another
  overlapping it
- **THEN** it SHALL NOT shove

#### Scenario: A shove comes off the feet

- **WHEN** a fighter is airborne
- **THEN** it SHALL NOT shove

#### Scenario: Guarding everything stands its ground

- **WHEN** a fighter is guarding at every height
- **THEN** it SHALL NOT shove

### Requirement: The shover turns on the fighter it shoves

A shove SHALL be thrown facing the fighter it moves, and SHALL move that fighter
away from the shover. When the shove has landed, the fighter that was moved SHALL
be in front of the shover. Two fighters on the same spot are shoved the way the
shover already faces.

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

### Requirement: A shove stuns and does no damage

A shove SHALL stun the fighter it is thrown at from the moment it begins: that
fighter SHALL be locked out as a stun locks it out, and whatever it was doing,
a swing included, SHALL end. The shove SHALL take no structure and no energy
from it, SHALL NOT count as a hit, and SHALL NOT be shown or felt as one - no
hit flash and no hit-stop.

#### Scenario: Stunned, not hurt

- **WHEN** a shove is thrown at a fighter
- **THEN** that fighter SHALL be stunned, and SHALL keep its structure, its
  energy and its hit count

#### Scenario: No hit is shown

- **WHEN** a shove is thrown
- **THEN** neither fighter SHALL be frozen by hit-stop, and the fighter shoved
  SHALL NOT flash as a hit makes it flash

#### Scenario: A fighter mid-swing is stunned out of it and moved

- **WHEN** a shove is thrown at a fighter part-way through a swing
- **THEN** that swing SHALL end, and the fighter SHALL be moved when the shove
  lands

#### Scenario: A swing already thrown does not cut the shove off

- **WHEN** the fighter shoved was part-way through a swing that would have
  reached the shover before the shove lands
- **THEN** that swing SHALL NOT land, and the shove SHALL play to its end

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

A shove's displacement and its stun SHALL be resolved only in the authoritative
simulation, as a hit is, and SHALL land at most once per fighter reached. A
client SHALL be shown the shove, the ground it changed and the stun by the
snapshot, and SHALL NOT move or stun any fighter by a shove of its own.

#### Scenario: A client moves nobody

- **WHEN** a predicting client simulates a fighter crowded from behind
- **THEN** no fighter's position SHALL change by a shove on that client, and no
  fighter SHALL be stunned by one

#### Scenario: A shove lands once

- **WHEN** a shove plays out in the authoritative simulation
- **THEN** it SHALL move each fighter it reaches at most once
