## Purpose

Puts each fighter's structure where the fight is happening: a posture-style bar over its head that fills as its guard is worn down, with marks counting the hits toward its next stun.

## ADDED Requirements

### Requirement: The bar fills as structure is lost
Each fighter's structure bar SHALL show how much structure it has lost, not how much it has left: empty at full structure, completely full at zero. The filled part SHALL grow from the centre of the bar outward, evenly on both sides, in proportion to the structure lost.

#### Scenario: A quarter lost
- **WHEN** a fighter is at 9 of its 12 structure
- **THEN** a quarter of its bar is filled, centred on the middle of the bar

#### Scenario: One hit from breaking
- **WHEN** a fighter is at 1 structure
- **THEN** its bar is almost completely filled

#### Scenario: Blocking shows
- **WHEN** a fighter at 8 structure blocks an attack at the attacker's height and goes to 8.5
- **THEN** the filled part of its bar shrinks

### Requirement: The bar is shown only while structure is below full
A fighter's bar and pips SHALL be drawn only while its structure is below full. They SHALL appear on the first loss and disappear once regeneration has refilled the pool. They SHALL NOT be drawn for a fighter in the Death state, nor for any fighter while time manipulation is showing a recorded frame.

#### Scenario: Untouched fighter
- **WHEN** a fighter has not lost any structure
- **THEN** no bar or pips are drawn over it

#### Scenario: First loss
- **WHEN** a fighter at full structure takes a hit
- **THEN** its bar appears on the next rendered frame

#### Scenario: Recovered
- **WHEN** a fighter's structure has regenerated back to full
- **THEN** its bar and pips are no longer drawn

#### Scenario: Dying fighter
- **WHEN** a fighter enters the Death state
- **THEN** no bar or pips are drawn over it while it falls and flickers

#### Scenario: Scrubbing time
- **WHEN** time manipulation is showing a recorded frame
- **THEN** no fighter has a bar or pips, and they return when the mode is left

### Requirement: Stun pips count toward the next stun
In a row centred just above the bar, a fighter SHALL show one pip for each landed hit toward its next stun: none, one or two. The row SHALL NOT extend past either end of the bar. The pips SHALL follow the same count the stun rule uses, clearing when a stun lands and when regeneration starts. A pip SHALL NOT be shown for a block.

#### Scenario: Two hits in
- **WHEN** a fighter has taken two hits and has not been stunned or regenerated since
- **THEN** two pips are shown in a row centred above its bar

#### Scenario: Close range
- **WHEN** two fighters below full structure, each with two pips, trade blows at arm's length
- **THEN** neither fighter's pips cover any part of the other fighter's bar

#### Scenario: Stunned
- **WHEN** the third hit stuns a fighter
- **THEN** its pips clear while the bar stays, showing the structure lost

#### Scenario: Blocks add no pips
- **WHEN** a fighter below full structure blocks an attack
- **THEN** its pip count does not change

### Requirement: The bar belongs to the fighter and the scene
The bar and pips SHALL be drawn above the fighter's head and move with it, including a remote fighter drawn from interpolated snapshots. They SHALL be drawn at the scene's pixel scale, so they zoom with the camera and stay on the pixel grid. Their colours SHALL NOT change with the scene's lighting: a fighter in shadow and a fighter under a lamp SHALL have identical bars.

#### Scenario: Follows a running fighter
- **WHEN** a fighter below full structure runs across the screen
- **THEN** its bar stays centred above its head for the whole run

#### Scenario: Zooms with the camera
- **WHEN** the camera zooms in by one step
- **THEN** the bar grows with the fighter, and its edges stay on the pixel grid

#### Scenario: Unaffected by lighting
- **WHEN** two fighters at the same structure stand one under a lamp and one in shadow
- **THEN** their bars are drawn in identical colours

### Requirement: Every fighter, whether or not debug is on
Every living fighter SHALL get a bar and pips under the rules above: the local fighter, bots, and every fighter a networked client or spectator can see. They SHALL be drawn whether the debug overlay is on or off.

#### Scenario: Debug off
- **WHEN** the debug overlay is toggled off and a fighter is below full structure
- **THEN** its bar and pips are still drawn

#### Scenario: Remote fighter
- **WHEN** a client in a networked match sees another player's fighter that has taken two hits
- **THEN** that fighter has the bar for the structure the server reports and two pips
