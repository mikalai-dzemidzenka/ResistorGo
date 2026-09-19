## MODIFIED Requirements

### Requirement: Pixel-art menu scene
The animated menu scene (the sky with its sun, the neon grid floor, the rain and the CRT drawn over them, and the wireframe head, jab, kick and martelo models) SHALL be rendered into a texture smaller than the virtual screen and drawn into the virtual screen upscaled with nearest-neighbour filtering, so it keeps a blocky pixel look. The on-screen size of one scene pixel SHALL be a single setting that divides the virtual screen evenly, and SHALL be large enough to stay visibly pixelated. Changing it SHALL change only the pixel density, not the size or position of anything in the scene.

#### Scenario: Scene is upscaled, not re-rendered at window size
- **WHEN** the main menu is shown at the default scene pixel size of 2
- **THEN** the scene is rendered at 640x360 and each scene pixel appears as a 2x2 block in the 1280x720 virtual screen

#### Scenario: Models are detailed enough to read
- **WHEN** the head model is shown after it has faded in
- **THEN** its wireframe edges are distinguishable from one another rather than merging into solid blocks

### Requirement: Menu layout splits the window
The logo, the menu widgets and the error message SHALL be in a menu panel on the left of the window (the left 560 pixels of the 1280x720 virtual screen). The panel SHALL be drawn, as a filled and bordered rectangle inset from the left edge and standing between the terminal's two rules, and everything the menu column shows SHALL be inside it. The wireframe models SHALL be in the area to the right of the panel. The head SHALL be at the top centre of that area with a jab on each side, and the martelo and the kick SHALL be below them.

#### Scenario: Menu on the left, models on the right
- **WHEN** the main menu has been shown for more than 5 seconds, so every model has faded in
- **THEN** the logo and the rows are in the left 560 pixels of the virtual screen, the head and both jabs are in a row at the top of the area to the right of it, and the martelo and the kick are below them

#### Scenario: The panel holds the whole column
- **WHEN** any menu screen is shown
- **THEN** the panel's header, the logo, the name of the screen, every row with its label and the error message are all drawn inside the panel's rectangle

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

### Requirement: Menu widgets fit inside the menu panel
On every menu screen (Main, Host Game, Join Game, Options, Connecting), every row, text field, spinner, slider, selector and label SHALL be fully inside the menu panel, with at least 16 pixels to the left and bottom edges of the virtual screen. The error message SHALL start at the left edge of the label column. Rows and the error message SHALL NOT overlap the logo. The order of the entries on each screen SHALL NOT change.

#### Scenario: Host screen fits
- **WHEN** the user opens the Host Game screen
- **THEN** the Port, Max players, Bots and Max spectators rows with their labels, and the Start and Back rows, are all fully visible inside the menu panel, and Back ends at least 16 pixels above the bottom of the virtual screen

#### Scenario: Options screen fits
- **WHEN** the user opens the Options screen
- **THEN** the fullscreen, window size, music volume and HUD rows with their labels, and the Back row, are all fully visible inside the menu panel

#### Scenario: Error message placement
- **WHEN** hosting fails with an error
- **THEN** the error text is shown between the name of the screen and the first row, overlapping neither

## ADDED Requirements

### Requirement: The window is framed as a terminal
The menu SHALL be framed by a rule across the top and a rule across the bottom of the window, each stopping short of the window's side edges, and each carrying readouts of what the terminal is and how it is worked. The top rule SHALL carry the terminal's name, the connection it stands on, and a status that reflects the session that is actually up. The bottom rule SHALL carry the keys the screen in front of the player answers to, and the protocol version the build speaks. The readouts SHALL be drawn in the palette's reds, like everything else the menu shows.

#### Scenario: The frame is there on every screen
- **WHEN** any menu screen is shown
- **THEN** a rule is drawn across the top and across the bottom of the window, with the terminal's name and connection on the top one and the key hints and protocol version on the bottom one

#### Scenario: The status readout follows the session
- **WHEN** the menu is shown with no session, then while a client is connecting, then while a host is listening
- **THEN** the top right readout says the system is online, then that a link is active, then that a host is active

#### Scenario: The hints follow the screen
- **WHEN** the main screen is shown, and then any other screen
- **THEN** the bottom left readout offers the navigate and select keys on the main screen, and the select and back keys on the others

### Requirement: The menu column stands on a panel
The panel SHALL carry a header above the logo — a short rule, the terminal's place, and a line naming what the terminal is for — and, under the logo, the name of the screen that is up. The panel SHALL be darker than the rows drawn on it, and the rows in turn darker than the highlighted row, so the column reads as raised out of the scene rather than cut into it.

#### Scenario: The header is above the logo
- **WHEN** any menu screen is shown
- **THEN** the header's rule and its two lines are drawn above the logo, inside the panel, and the logo is not drawn over them

#### Scenario: The screen names itself
- **WHEN** the user opens the Host Game, Join Game, Options and Connecting screens in turn
- **THEN** the line under the logo names the screen that is up each time

### Requirement: Menu entries are rows of the terminal
Every menu entry SHALL be drawn as a row of the terminal: its position in the screen as a two-digit number, a marker down its left edge, a border, and a fill. Exactly one row of a screen SHALL be highlighted at a time, and the highlighted row SHALL be drawn with a brighter border and fill, a marker of its own, a glow that pulses around it, and a pointer at its right edge. Taking a row SHALL do what that entry has always done.

#### Scenario: A row carries its number
- **WHEN** the main screen is shown
- **THEN** its five entries are drawn as rows numbered 01 to 05, in the order they have always been in

#### Scenario: One row is lit
- **WHEN** the main screen is shown
- **THEN** exactly one row is drawn with the bright border, the pulsing glow and the pointer, and the other four are not

#### Scenario: Taking a row acts
- **WHEN** the player takes the first row of the main screen
- **THEN** a single-player match starts, as it did when that entry was a button

### Requirement: The main screen is worked by the keyboard
On the main screen the highlight SHALL move with W and S and with the up and down arrows, wrapping at both ends, and the highlighted row SHALL be taken with Enter or Space. Moving the mouse over a row, or clicking one, SHALL move the highlight to it, so the mouse and the keyboard never disagree about which row is lit. The screens that have text fields, spinners, sliders and selectors on their rows SHALL stay mouse-driven, so a keystroke meant for a field is never read as navigation.

#### Scenario: The highlight wraps
- **WHEN** the highlight is on the last row of the main screen and the player presses S
- **THEN** the highlight moves to the first row

#### Scenario: The keyboard alone is enough
- **WHEN** the player starts the game and, without touching the mouse, moves the highlight to an entry and presses Enter
- **THEN** that entry is taken

#### Scenario: The mouse takes the highlight
- **WHEN** the player moves the mouse over a row other than the highlighted one
- **THEN** the highlight moves to the row under the mouse, and only that row is lit

#### Scenario: Typing into a field is not navigation
- **WHEN** the player is editing the address on the Join Game screen and types a character that is also a navigation key
- **THEN** the character goes into the field and no highlight moves

### Requirement: Escape backs out of a screen
On every screen but the main one, Escape SHALL do exactly what that screen's last row does: it SHALL return to the main screen from Host Game and Join Game and clear any error, save the settings and return from Options, and cancel the attempt from Connecting. Any field left in edit mode SHALL stop being edited when the screen is left.

#### Scenario: Escape leaves the host screen
- **WHEN** the user presses Escape on the Host Game screen with an error shown
- **THEN** the main screen is up and the error is gone

#### Scenario: Escape saves the options
- **WHEN** the user changes the music volume and presses Escape
- **THEN** the main screen is up and the new volume has been saved, as it is by the Back row

#### Scenario: Escape cancels a connection
- **WHEN** the user presses Escape while the Connecting screen is up
- **THEN** the attempt is dropped and the screen it was started from is up again

### Requirement: Weather and a CRT over the scene
The scene SHALL have rain falling over it and a CRT laid over that. The rain SHALL be positions normalized to the scene, so it is laid out the same whatever the scene is drawn at, and SHALL be the same rain on every run. A streak that falls past the bottom SHALL come back above the top in a new column. The CRT SHALL be a band sweeping down the scene and scanline gaps across it, and the gaps SHALL be painted in the scene's darkest red rather than in black, so the scene keeps to the palette's reds.

#### Scenario: The same rain every run
- **WHEN** the menu is shown at the same moment of two separate runs
- **THEN** the streaks are in the same places in both

#### Scenario: The rain covers the scene and stays in it
- **WHEN** the menu is shown
- **THEN** every streak is inside the scene, none longer than a quarter of its height, and each falls at a speed of its own

#### Scenario: The CRT keeps to the reds
- **WHEN** a frame of the menu is captured
- **THEN** the scanline gaps have cut the grid, the sun and the rain, and every pixel of the scene is still one of the palette's reds

### Requirement: The widgets that remain match the menu
The text fields, spinners, sliders, toggles and selectors the menu still uses SHALL be drawn in the same reds as the rows around them, in every state — normal, focused and pressed — rather than in the level editor's style. Starting a level SHALL put the editor's own style back, so the editor and the level's overlays are unaffected by what the menu did.

#### Scenario: The host screen is all one picture
- **WHEN** the user opens the Host Game screen
- **THEN** the port field and the three spinners are drawn in the menu's reds, matching the Start and Back rows

#### Scenario: The editor keeps its own style
- **WHEN** the player starts a match and opens the level editor
- **THEN** the editor's widgets are in the editor's dark style, not the menu's reds
