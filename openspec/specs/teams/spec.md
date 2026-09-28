# teams Specification

## Purpose
Shows each player whose side a fighter is on once a gamemode splits them into teams, so an ally is never taken for an enemy that cannot be hurt.

## Requirements

### Requirement: A line under a team's fighters
A fighter on team 2 SHALL stand on a red line and a fighter on team 3 on a blue one: a line one world pixel high and as wide as the fighter's body, on the ground directly under its feet, on the pixel grid. Both colours SHALL be exact entries of the level palette. A fighter on any other team - no team, or the bots' team - SHALL have no line, so a match without a gamemode, and single player, look as they did.

The line SHALL be drawn over the lighting, like the structure bar, so a fighter in a dark corner still shows its side. It SHALL be part of the HUD over the fighters: hidden with it, not drawn on a dying fighter, and not drawn over a recorded frame.

Every client SHALL draw the line from the team the snapshots carry, so all players see the same sides.

#### Scenario: Two sides
- **WHEN** a gamemode puts one fighter on team 2 and another on team 3
- **THEN** every client draws a red line under the first and a blue line under the second

#### Scenario: No gamemode
- **WHEN** a match runs with no gamemode, humans with no team and bots on the bots' team
- **THEN** no fighter has a line

#### Scenario: A side changes
- **WHEN** a script moves a fighter from team 2 to team 3
- **THEN** its line turns from red to blue on every client

#### Scenario: HUD hidden
- **WHEN** the player hides the HUD
- **THEN** the lines go with the bars, and come back with them

### Requirement: An ally is not marked as a target
The attack outline SHALL NOT mark an ally - a fighter on the same team as the local player, other than no team - because the swing passes through it. Two fighters with no team SHALL still mark each other.

#### Scenario: An ally in reach
- **WHEN** the local player on team 2 aims a swing that would reach a fighter also on team 2
- **THEN** that fighter carries no outline

#### Scenario: An enemy in reach
- **WHEN** the same swing would reach a fighter on team 3
- **THEN** that fighter is outlined as usual
