## MODIFIED Requirements

### Requirement: Gameplay input is suppressed while the mode is on
While time-manipulation mode is on, fighter input SHALL NOT be applied: movement, jumping, attacking and stance changes SHALL have no effect on the match until the mode is left. The mouse SHALL drive only the scrub. The crosshair MAY keep following the mouse. Input given while the mode is on SHALL NOT be held over and acted on once the match resumes.

#### Scenario: Attacking does nothing while scrubbing
- **WHEN** the player clicks the attack button while the mode is on
- **THEN** no attack is queued or resolved, and the fighter's animation stays on the scrubbed frame

#### Scenario: Movement keys do nothing while scrubbing
- **WHEN** the player holds a movement key while the mode is on
- **THEN** the fighter does not move and no input is buffered for when the mode is left

#### Scenario: Clicks do not carry over
- **WHEN** the player clicks the attack button several times while the mode is on and then leaves it
- **THEN** the fighter does not attack on leaving, and the match resumes with no action pending
