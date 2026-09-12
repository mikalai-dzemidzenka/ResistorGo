## Purpose

Defines how the main menu is laid out and presented: the animated wireframe scene, the game logo and the menu widgets are scaled for the game's virtual resolution, placed side by side, and always fit inside the game window.

## ADDED Requirements

### Requirement: Pixel-art menu scene
The animated menu scene (background line fan and the wireframe head, jab, kick and martelo models) SHALL be rendered into a texture smaller than the window and shown upscaled with nearest-neighbour filtering, so it keeps a blocky pixel look. The on-screen size of one scene pixel SHALL be a single setting that divides the window evenly, and SHALL be large enough to stay visibly pixelated. Changing it SHALL change only the pixel density, not the size or position of anything in the scene.

#### Scenario: Scene is upscaled, not re-rendered at window size
- **WHEN** the main menu is shown in the default 1280x720 window at the default scene pixel size of 4
- **THEN** the scene is rendered at 320x180 and each scene pixel appears as a 4x4 block of identical screen pixels

#### Scenario: Models are detailed enough to read
- **WHEN** the head model is shown after it has faded in
- **THEN** its wireframe edges are distinguishable from one another rather than merging into solid blocks

### Requirement: Menu layout splits the window
The logo, the menu widgets and the error message SHALL be in a menu panel on the left of the window (the left 560 pixels of the default 1280x720 window). The wireframe models SHALL be in the area to the right of the panel. The head SHALL be at the top centre of that area with a jab on each side, and the martelo and the kick SHALL be below them.

#### Scenario: Menu on the left, models on the right
- **WHEN** the main menu has been shown for more than 5 seconds, so every model has faded in
- **THEN** the logo and the buttons are in the left 560 pixels of the window, the head and both jabs are in a row at the top of the area to the right of it, and the martelo and the kick are below them

### Requirement: Menu scene composition is independent of the window scale
The size and position of the wireframe models and the background line fan SHALL be defined in scene units only. Changing the window scale factor SHALL NOT change the scene's composition, only the size of its pixels. Once every model has faded in, all five models SHALL be fully inside the window. On every menu screen, none of them SHALL be covered by the logo, a widget or the error message.

#### Scenario: Models fit in the window
- **WHEN** the main menu has been shown for more than 5 seconds
- **THEN** the head, both jabs, the kick and the martelo are each fully inside the window, none clipped by a window edge, whatever their rotation

#### Scenario: Models are not hidden by the menu
- **WHEN** any menu screen (Main, Host Game with an error shown, Join Game, Connecting) is shown after every model has faded in
- **THEN** no part of any model is drawn under the logo, a button, text field, spinner, label or the error message

#### Scenario: Models do not fill the screen
- **WHEN** the main menu is shown
- **THEN** the head is at most half the window height tall

### Requirement: Logo fits inside the menu panel
The "RESISTOR" logo SHALL be drawn at a whole-number scale (at least 1), with nearest-neighbour filtering. It SHALL be centred horizontally over the menu's widget column, above the widgets and the error message. The scale SHALL be the largest one at which the logo stays inside the menu panel, and never more than the window scale factor. Its fade-in timing is unchanged.

#### Scenario: Whole logo visible
- **WHEN** the logo has fully faded in, in the default 1280x720 window
- **THEN** every letter of the logo is visible, and the logo is drawn at 2x (418x56 pixels) inside the menu panel, centred over the buttons

### Requirement: Menu widgets fit inside the menu panel
On every menu screen (Main, Host Game, Join Game, Connecting), every button, text field, spinner and label SHALL be fully inside the menu panel, with at least 16 pixels to the left and bottom edges of the window. The error message SHALL start at the left edge of the label column. Widgets and the error message SHALL NOT overlap the logo. The order of the entries on each screen SHALL NOT change.

#### Scenario: Host screen fits
- **WHEN** the user opens the Host Game screen in the default 1280x720 window
- **THEN** the Port, Max players, Bots and Max spectators rows with their labels, and the Start and Back buttons, are all fully visible inside the menu panel, and Back ends at least 16 pixels above the bottom of the window

#### Scenario: Error message placement
- **WHEN** hosting fails with an error
- **THEN** the error text is shown between the logo and the first row, overlapping neither
