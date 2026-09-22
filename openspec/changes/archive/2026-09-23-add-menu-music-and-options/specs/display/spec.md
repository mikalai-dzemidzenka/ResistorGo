## Purpose

Covers how the rendered game reaches the player's screen: the fixed virtual screen everything is drawn into, how it is scaled and centred in the window, and the fullscreen and window-size options.

## ADDED Requirements

### Requirement: Fixed virtual screen
The game SHALL render everything — the level, the main menu, the HUD, the cursor and any debug overlays — into a single virtual screen of a fixed size (1280x720). Nothing SHALL be drawn straight to the window. The amount of the game world visible SHALL depend only on the virtual screen, never on the window size, the resolution setting or fullscreen.

#### Scenario: Same view at every resolution
- **WHEN** two players run the game at different window sizes, one windowed and one fullscreen, and stand at the same place in the level
- **THEN** both see exactly the same area of the level, differing only in how large it appears

### Requirement: Scaled and letterboxed presentation
The virtual screen SHALL be drawn to the window scaled by the largest factor that fits both dimensions, preserving its aspect ratio, and centred. Any leftover area SHALL be filled with black bars. Scaling SHALL use nearest-neighbour filtering, so the pixel-art look is preserved.

#### Scenario: Wider window letterboxes at the sides
- **WHEN** the window is wider than 16:9 for the virtual screen
- **THEN** the game is centred at the largest fitting scale with black bars to its left and right, and is not stretched

#### Scenario: Exact multiple fills the window
- **WHEN** the window is exactly 1280x720
- **THEN** the virtual screen fills it completely at scale 1 with no bars

### Requirement: Mouse input follows the presentation
The mouse position read by the game SHALL be in virtual screen coordinates. A click SHALL act on whatever the player sees under the cursor at any scale, fullscreen or window size. A cursor over a black bar SHALL fall outside the virtual screen and SHALL NOT activate anything.

#### Scenario: Clicking a button when scaled
- **WHEN** the game is fullscreen at a scale other than 1 and the player clicks a menu button
- **THEN** the button under the cursor activates, the same one it would at 1280x720

### Requirement: Fullscreen
The player SHALL be able to turn fullscreen on and off. Fullscreen SHALL use the desktop resolution of the monitor the window is on, with the virtual screen scaled and letterboxed inside it. Leaving fullscreen SHALL restore the window at the configured window size.

#### Scenario: Entering fullscreen
- **WHEN** the player turns fullscreen on
- **THEN** the game fills the monitor at its desktop resolution, letterboxed, with the menu and gameplay fully usable

#### Scenario: Leaving fullscreen
- **WHEN** the player turns fullscreen off
- **THEN** the game returns to a window of the configured size

### Requirement: Window size
The player SHALL be able to choose the window size from a list of 16:9 sizes (1280x720, 1600x900 and 1920x1080). Choosing one SHALL resize the window immediately. The setting SHALL apply when not fullscreen.

#### Scenario: Changing window size
- **WHEN** the player picks 1600x900 while windowed
- **THEN** the window becomes 1600x900 immediately and the game is drawn scaled to fill it

#### Scenario: Window size while fullscreen
- **WHEN** the player changes the window size while fullscreen is on
- **THEN** the display does not change until fullscreen is turned off, at which point the window takes the chosen size
