## MODIFIED Requirements

### Requirement: Pixel-art menu scene
The animated menu scene (background line fan and the wireframe head, jab, kick and martelo models) SHALL be rendered into a texture smaller than the virtual screen and drawn into the virtual screen upscaled with nearest-neighbour filtering, so it keeps a blocky pixel look. The on-screen size of one scene pixel SHALL be a single setting that divides the virtual screen evenly, and SHALL be large enough to stay visibly pixelated. Changing it SHALL change only the pixel density, not the size or position of anything in the scene.

#### Scenario: Scene is upscaled, not re-rendered at window size
- **WHEN** the main menu is shown at the default scene pixel size of 4
- **THEN** the scene is rendered at 320x180 and each scene pixel appears as a 4x4 block in the 1280x720 virtual screen

#### Scenario: Models are detailed enough to read
- **WHEN** the head model is shown after it has faded in
- **THEN** its wireframe edges are distinguishable from one another rather than merging into solid blocks

### Requirement: Menu widgets fit inside the menu panel
On every menu screen (Main, Host Game, Join Game, Options, Connecting), every button, text field, spinner, slider, selector and label SHALL be fully inside the menu panel, with at least 16 pixels to the left and bottom edges of the virtual screen. The error message SHALL start at the left edge of the label column. Widgets and the error message SHALL NOT overlap the logo. The order of the entries on each screen SHALL NOT change.

#### Scenario: Host screen fits
- **WHEN** the user opens the Host Game screen
- **THEN** the Port, Max players, Bots and Max spectators rows with their labels, and the Start and Back buttons, are all fully visible inside the menu panel, and Back ends at least 16 pixels above the bottom of the virtual screen

#### Scenario: Options screen fits
- **WHEN** the user opens the Options screen
- **THEN** the fullscreen, window size and music volume rows with their labels, and the Back button, are all fully visible inside the menu panel

#### Scenario: Error message placement
- **WHEN** hosting fails with an error
- **THEN** the error text is shown between the logo and the first row, overlapping neither
