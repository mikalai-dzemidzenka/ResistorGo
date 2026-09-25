## ADDED Requirements

### Requirement: Touching a death area kills
A level SHALL be able to hold death areas: rectangles of the level, tied to no tile. A living fighter whose collision box overlaps a death area by any amount SHALL enter the Death state on that tick, whatever its structure and whatever it was doing, and from then on SHALL die, flicker and respawn exactly as a fighter whose structure broke does. No fighter SHALL be credited with the kill: nobody is rewarded as a killer is, and scripts are told of the death with no killer. A fighter already in the Death state SHALL be unaffected by a death area, and a death area SHALL NOT block, carry or otherwise move anything that passes through it.

Only the authoritative simulation SHALL decide a death-area death, as it decides every other; a predicting client SHALL show it only when it is told of it.

#### Scenario: Falling into a pit
- **WHEN** a fighter at full structure falls into a death area
- **THEN** it enters the Death state on the first tick its collision box overlaps the area, plays the death fall, flickers and is removed, and respawns as a broken fighter does

#### Scenario: One pixel is enough
- **WHEN** a fighter's collision box overlaps a death area by a single pixel
- **THEN** it dies

#### Scenario: Nobody gets the kill
- **WHEN** a fighter that was last hit by another one dies in a death area
- **THEN** the other fighter's energy and structure are unchanged by the death, and scripts are told of a death with no killer

#### Scenario: A body is left alone
- **WHEN** a fighter already in the Death state falls through a death area
- **THEN** its death goes on as it was, and it is not killed again

#### Scenario: Nothing to stand on
- **WHEN** a fighter walks along a floor with a death area just beside it that it does not touch
- **THEN** it lives, and its movement is exactly what it would be with no area there

#### Scenario: The client waits to be told
- **WHEN** a client's predicted world moves its fighter into a death area before the server has
- **THEN** the fighter does not die on the client until a snapshot says it has
