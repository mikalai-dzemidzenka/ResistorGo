## ADDED Requirements

### Requirement: The flashlight shines only from a drawn pistol
With the flashlight setting on, a living fighter holding the pistol SHALL cast the flashlight's beam only while its pistol is drawn on screen. Whenever the fighter is shown in a pose that does not draw the pistol - climbing a ledge, climbing down, jumping or falling, ducking or rolling, taking a hit - it SHALL cast no beam, and the beam SHALL come back as soon as the pistol is drawn again. A fighter SHALL never cast a beam from a pistol that is not on screen.

#### Scenario: Climbing a ledge
- **WHEN** a fighter holding the pistol, with the flashlight on, catches a ledge and climbs it
- **THEN** it casts no beam for the whole climb, and casts one again once it stands with the pistol drawn

#### Scenario: A jump
- **WHEN** a fighter holding the pistol, with the flashlight on, jumps
- **THEN** it casts no beam while it is off its pistol legs, and casts one again once it has landed with the pistol drawn

#### Scenario: A roll
- **WHEN** a fighter holding the pistol, with the flashlight on, ducks or rolls
- **THEN** it casts no beam until it is back on its pistol legs

#### Scenario: Standing armed
- **WHEN** a fighter holding the pistol, with the flashlight on, stands or walks with the pistol drawn
- **THEN** it casts the beam along its aim, as before
