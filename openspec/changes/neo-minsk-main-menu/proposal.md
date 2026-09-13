## Why

The main menu is a red wireframe sketch: a perspective fan of red lines, five red wireframe fighter models, and the whole scene tinted red when it is drawn. It says "prototype", not "place". The game has a setting that the menu never shows. A neon night over a future Minsk gives the first screen an identity, and drawing it in a strict palette gives the menu a coherent pixel-art look that the free RGB of today's menu does not have.

## What Changes

- **A synthwave neon grid floor** replaces the red line fan: a glowing grid plane stretching to the horizon, scrolling toward the viewer.
- **The Neo-Minsk skyline behind it**, drawn in code as pixel art. It is a recognisable silhouette of Minsk — the National Library's cut-gem "diamond", the Hero City obelisk, the twin towers of the Gates of Minsk at the station square, the Stalinist blocks of Independence Avenue, the TV tower — pushed into the future with lit windows against a dusk sky. The buildings are flat silhouettes, with no outlines.
- **The colours cycle.** The grid and the skyline's lit windows step from one palette colour to the next over time, so the scene keeps moving even though the city does not.
- **Every menu colour comes from Endesga 64,** and the scene is painted only from that palette's reds. The finished menu frame — scene, logo and widgets — is snapped to the 64 Endesga colours; the sky, city, grid and models use only the red end of it. The logo and the widgets keep their own colours so they stay legible. The level keeps its own palette, untouched.
- The wireframe fighter models stay where they are, in the area right of the menu panel, recoloured from plain red to the palette's reds.

### Assumptions

- The background is fixed: it does not roll, sway or drift. Only the grid's scroll and the colour cycling move.
- Colour cycling steps between palette entries. It never blends to a colour outside the palette, so the frame stays within Endesga 64 at every moment.
- "The palette's reds" means its entries whose hue is within 20 degrees of pure red. That is a testable rule, and it covers the palette's dark maroons through to its bright crimsons.
- The raygui widget colours from `style_dark.rgs` are snapped to their nearest Endesga 64 entries along with everything else. They are not restyled by hand.
- The city carries no text. The raylib default font has no Cyrillic, and the landmarks are what make it Minsk.
- The menu music, Options screen and every menu screen keep working as they do now.

## Capabilities

### New Capabilities

None. The menu's look is part of the existing `main-menu` capability.

### Modified Capabilities

- `main-menu`:
  - *Pixel-art menu scene* names the scene's parts. The background line fan becomes the sky, the Neo-Minsk skyline and the neon grid floor. This builds on the version of the requirement in the still-active `add-menu-music-and-options` change, whose code has already shipped, not on the older main-spec text.
  - *Menu scene composition is independent of the window scale* also names the line fan, and now covers the new background layers.
  - New requirements for the neon grid floor, the Neo-Minsk skyline, the background holding still, colour cycling, and the Endesga 64 palette with its red-only scene.

## Impact

- **Code**
  - `main_menu.go`: the line fan is replaced by the sky, skyline and grid drawing, the colour cycling, and the model colours.
  - A new `menu_city.go`: the procedural Neo-Minsk skyline. A new `menu_scene.go`: the scene's palette, its colour cycle and the background painter.
  - `palette.go`: the Endesga 64 table beside the existing level palette.
  - `game.go` / `present.go`: a menu-only palette pass, reusing the level's existing palette-quantization shader with the Endesga 64 table. The shader is currently loaded only when a level starts, so it has to be loaded before the menu first draws.
- **Tests**: the Endesga 64 table, the red set, the skyline layout and the colour cycle are checked without a window. The look itself is checked by running the game.
- **Not affected**: the level, its palette and lighting, gameplay, networking, the menu layout and widgets' positions, the Options screen and the menu music.

## Non-goals

- Switching the level to Endesga 64.
- New menu screens, widgets or layout changes.
- An intro sequence, or animation of the logo or the widgets.
- Hand-drawn art assets. The skyline is procedural.
