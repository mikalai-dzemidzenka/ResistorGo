## MODIFIED Requirements

### Requirement: Pixel-art menu scene
The animated menu scene (the sky, the Neo-Minsk skyline, the neon grid floor, and the wireframe head, jab, kick and martelo models) SHALL be rendered into a texture smaller than the virtual screen and drawn into the virtual screen upscaled with nearest-neighbour filtering, so it keeps a blocky pixel look. The on-screen size of one scene pixel SHALL be a single setting that divides the virtual screen evenly, and SHALL be large enough to stay visibly pixelated. Changing it SHALL change only the pixel density, not the size or position of anything in the scene.

#### Scenario: Scene is upscaled, not re-rendered at window size
- **WHEN** the main menu is shown at the default scene pixel size of 4
- **THEN** the scene is rendered at 320x180 and each scene pixel appears as a 4x4 block in the 1280x720 virtual screen

#### Scenario: Models are detailed enough to read
- **WHEN** the head model is shown after it has faded in
- **THEN** its wireframe edges are distinguishable from one another rather than merging into solid blocks

### Requirement: Menu scene composition is independent of the window scale
The size and position of the wireframe models and of the background layers — the sky, the Neo-Minsk skyline and the neon grid floor — SHALL be defined in scene units only. Changing the window scale factor SHALL NOT change the scene's composition, only the size of its pixels. Once every model has faded in, all five models SHALL be fully inside the window. On every menu screen, none of them SHALL be covered by the logo, a widget or the error message. The background SHALL cover the whole window, with no edge of it and no empty area behind it ever showing.

#### Scenario: Models fit in the window
- **WHEN** the main menu has been shown for more than 5 seconds
- **THEN** the head, both jabs, the kick and the martelo are each fully inside the window, none clipped by a window edge, whatever their rotation

#### Scenario: Models are not hidden by the menu
- **WHEN** any menu screen (Main, Host Game with an error shown, Join Game, Connecting) is shown after every model has faded in
- **THEN** no part of any model is drawn under the logo, a button, text field, spinner, label or the error message

#### Scenario: Models do not fill the screen
- **WHEN** the main menu is shown
- **THEN** the head is at most half the window height tall

#### Scenario: The background fills the window
- **WHEN** any main menu screen is shown
- **THEN** every pixel of the window is covered by the sky, the skyline, the grid or something drawn over them

## ADDED Requirements

### Requirement: Neon grid floor
Below the horizon, the scene SHALL show a synthwave grid floor: glowing lines running from the viewer to the horizon, crossed by lines parallel to it, drawn in perspective so that they converge toward the horizon. The cross lines SHALL move steadily toward the viewer and repeat seamlessly, so the floor appears to scroll forever.

#### Scenario: Perspective
- **WHEN** the main menu is shown
- **THEN** the grid's lines converge toward the horizon, and cross lines near the viewer are further apart on screen than those near the horizon

#### Scenario: Endless scroll
- **WHEN** the main menu has been shown for a minute
- **THEN** the grid has scrolled toward the viewer the whole time, with no jump or visible restart

### Requirement: Neo-Minsk skyline
Behind the grid, along the horizon, the scene SHALL show a night skyline of a future Minsk against a dusk sky. It SHALL include these landmarks, each recognisable by its silhouette: the National Library's cut-gem "diamond", the Hero City obelisk, the twin towers of the Gates of Minsk, the Stalinist blocks of Independence Avenue, and the TV tower. The buildings SHALL be flat silhouettes, without outlines, lit only by their windows, and SHALL carry no text. The skyline SHALL be drawn by the game itself, from no image asset.

#### Scenario: Landmarks present
- **WHEN** the main menu is shown
- **THEN** the library's diamond, the obelisk, the twin towers, the avenue blocks and the TV tower are each visible along the skyline, none hidden behind another

#### Scenario: Night city
- **WHEN** the main menu is shown
- **THEN** the buildings are solid silhouettes with lit windows against a dusk sky, and no building has an outline drawn around its edge

#### Scenario: Same city every time
- **WHEN** the game is started twice
- **THEN** both show the same skyline, landmark for landmark and window for window

### Requirement: The background holds still
The sky, the skyline and the grid SHALL keep a fixed position and orientation: the background SHALL NOT roll, sway or otherwise move as a whole. Only the grid's own scroll and the cycling of the scene's colours SHALL change from frame to frame.

#### Scenario: The city does not move
- **WHEN** two frames of the main menu are captured ten seconds apart
- **THEN** every building of the skyline is at the same place in both, and the horizon is level in both

### Requirement: The scene's colours cycle
The colours of the grid and of the skyline's lit windows SHALL cycle over time, stepping from one palette colour to the next. The cycle SHALL use only palette colours and SHALL NOT pass through a colour outside the palette.

#### Scenario: Colours move
- **WHEN** the main menu is watched for ten seconds
- **THEN** the grid and the lit windows have changed colour at least once

#### Scenario: Stepping, not blending
- **WHEN** a frame is captured in the middle of a colour change
- **THEN** every pixel of it is one of the palette's colours

### Requirement: The menu is drawn in Endesga 64, the scene in its reds
Every pixel of the main menu frame — the scene, the logo, the widgets and their text — SHALL be one of the 64 colours of the Endesga 64 palette. Fades, transparency and blending SHALL be resolved to the nearest palette colour. The scene itself — the sky, the sun, the skyline, its windows, the grid and the wireframe models — SHALL be painted only from the palette's red entries: those whose hue is within 20 degrees of pure red. The logo and the widgets keep their own colours, so they stay legible. The level SHALL keep its own palette and SHALL NOT be drawn in Endesga 64.

#### Scenario: Only palette colours on the menu
- **WHEN** any main menu screen is captured, at any moment of its animation
- **THEN** every pixel of the capture is one of the 64 Endesga 64 colours

#### Scenario: The scene is red
- **WHEN** the colours the scene is painted with are listed
- **THEN** every one of them is an Endesga 64 entry whose hue is within 20 degrees of pure red

#### Scenario: The level is unchanged
- **WHEN** a single-player match is started from the menu
- **THEN** the level is drawn in its own palette exactly as before
