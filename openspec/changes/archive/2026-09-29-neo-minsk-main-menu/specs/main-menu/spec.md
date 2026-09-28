## MODIFIED Requirements

### Requirement: Pixel-art menu scene
The animated menu scene (the sky with its sun, the neon grid floor, the rain and the CRT drawn over them, and the wireframe head, jab, kick and martelo models) SHALL be rendered into a texture smaller than the virtual screen and drawn into the virtual screen upscaled with nearest-neighbour filtering, so it keeps a blocky pixel look. The on-screen size of one scene pixel SHALL be a single setting that divides the virtual screen evenly, and SHALL be large enough to stay visibly pixelated. Changing it SHALL change only the pixel density, not the size or position of anything in the scene.

#### Scenario: Scene is upscaled, not re-rendered at window size
- **WHEN** the main menu is shown at the default scene pixel size of 2
- **THEN** the scene is rendered at 640x360 and each scene pixel appears as a 2x2 block in the 1280x720 virtual screen

#### Scenario: Models are detailed enough to read
- **WHEN** the head model is shown after it has faded in
- **THEN** its wireframe edges are distinguishable from one another rather than merging into solid blocks

### Requirement: Menu scene composition is independent of the window scale
The size and position of the wireframe models and of the background layers — the sky, the neon grid floor and the weather over them — SHALL be defined in scene units only. Changing the window scale factor SHALL NOT change the scene's composition, only the size of its pixels. Once every model has faded in, all five models SHALL be fully inside the window. On every menu screen, none of them SHALL be covered by the logo, a widget, a row or the error message; the terminal's rules and their readouts are part of the picture rather than furniture and MAY cross the models. The background SHALL cover the whole window, with no edge of it and no empty area behind it ever showing.

#### Scenario: Models fit in the window
- **WHEN** the main menu has been shown for more than 5 seconds
- **THEN** the head, both jabs, the kick and the martelo are each fully inside the window, none clipped by a window edge, whatever their rotation

#### Scenario: Models are not hidden by the menu
- **WHEN** any menu screen (Main, Host Game with an error shown, Join Game, Options, Connecting) is shown after every model has faded in
- **THEN** no part of any model is drawn under the logo, a row, a text field, spinner, slider, selector, label or the error message

#### Scenario: The frame may cross the models
- **WHEN** the main menu is shown after every model has faded in
- **THEN** the top and bottom rules run the full width of the window across the model area, one pixel high, and their readouts sit in the corners clear of the models

#### Scenario: Models do not fill the screen
- **WHEN** the main menu is shown
- **THEN** the head is at most half the window height tall

#### Scenario: The background fills the window
- **WHEN** any main menu screen is shown
- **THEN** every pixel of the window is covered by the sky, the grid or something drawn over them

## ADDED Requirements

### Requirement: Neon grid floor
Below the horizon, the scene SHALL show a synthwave grid floor: glowing lines running from the viewer to the horizon, crossed by lines parallel to it, drawn in perspective so that they converge toward the horizon. The cross lines SHALL move steadily toward the viewer and repeat seamlessly, so the floor appears to scroll forever.

#### Scenario: Perspective
- **WHEN** the main menu is shown
- **THEN** the grid's lines converge toward the horizon, and cross lines near the viewer are further apart on screen than those near the horizon

#### Scenario: Endless scroll
- **WHEN** the main menu has been shown for a minute
- **THEN** the grid has scrolled toward the viewer the whole time, with no jump or visible restart

### Requirement: The background holds still
The sky, its sun and the grid SHALL keep a fixed position and orientation: the background SHALL NOT roll, sway or otherwise move as a whole. Only the grid's own scroll, the rain, the CRT's sweeping band and the cycling of the scene's colours SHALL change from frame to frame.

#### Scenario: The background does not move
- **WHEN** two frames of the main menu are captured ten seconds apart
- **THEN** the sun is at the same place in both, and the horizon is level in both

### Requirement: The scene's colours cycle
The colour of the grid SHALL cycle over time, stepping from one palette colour to the next. The cycle SHALL use only palette colours and SHALL NOT pass through a colour outside the palette.

#### Scenario: Colours move
- **WHEN** the main menu is watched for ten seconds
- **THEN** the grid has changed colour at least once

#### Scenario: Stepping, not blending
- **WHEN** a frame is captured in the middle of a colour change
- **THEN** every pixel of it is one of the palette's colours

### Requirement: The menu is drawn in Endesga 64, the scene in its reds
Every pixel of the main menu frame — the scene, the logo, the widgets and their text — SHALL be one of the 64 colours of the Endesga 64 palette. Fades, transparency and blending SHALL be resolved to the nearest palette colour. The scene itself — the sky, the sun, the grid and the wireframe models — SHALL be painted only from the palette's red entries: those whose hue is within 20 degrees of pure red. The logo and the widgets keep their own colours, so they stay legible. The level SHALL keep its own palette and SHALL NOT be drawn in Endesga 64.

#### Scenario: Only palette colours on the menu
- **WHEN** any main menu screen is captured, at any moment of its animation
- **THEN** every pixel of the capture is one of the 64 Endesga 64 colours

#### Scenario: The scene is red
- **WHEN** the colours the scene is painted with are listed
- **THEN** every one of them is an Endesga 64 entry whose hue is within 20 degrees of pure red

#### Scenario: The level is unchanged
- **WHEN** a single-player match is started from the menu
- **THEN** the level is drawn in its own palette exactly as before
