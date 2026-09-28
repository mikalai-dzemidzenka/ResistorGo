## Why

`neo-minsk-main-menu` gave the menu's scene an identity — a red synthwave grid under a banded sun, every pixel of it an Endesga 64 colour — and then left the menu itself sitting on top of that picture in the level editor's dark raygui style, deliberately: restyling the widgets was one of its non-goals. The result is a good scene with a developer tool parked in front of it. The first screen also could not be worked without a mouse, which is a strange thing to ask of a fighting game.

`examples/example_main_menus/1` is the answer this change takes: the menu is a terminal you are looking at, framed by rules with readouts on them, with the entries as numbered rows on a panel. Everything it draws is already available in the palette's reds.

## What Changes

- **The window is framed as a terminal.** A rule across the top and the bottom, each carrying readouts: what this terminal is, what it is connected to, whether a link or a host is up, how the screen in front of you is worked, and the protocol it speaks.
- **The menu column stands on a drawn panel**, with a header above the logo and the name of the current screen under it. The logo gains an offset ghost of itself.
- **The entries are rows drawn by hand**, replacing every raygui button: a number, a marker down the left edge, a border, and a pulsing glow with a `>` on the highlighted one.
- **The main screen is worked by the keyboard as well as the mouse.** W/S or the arrows move the highlight, Enter or Space takes the row, and moving the mouse over a row takes the highlight so the two ways of picking never disagree. Escape backs out of every other screen exactly as its Back row does.
- **Weather and a CRT over the scene**: rain falling in normalized coordinates, a band sweeping down, scanline gaps, a sky gradient and the horizon the grid converges on.
- **The widgets that remain** — the text fields, spinners, the slider, the toggles and the dropdown — are painted in the scene's reds rather than left dark, and the editor's dark style is put back when a level starts.
- **Scene chrome is allowed over the models.** The rule that keeps the menu off the wireframe models is reworded: it holds the *widgets* off them, and lets the frame cross them. The rules are one pixel high at the window's edges and their readouts sit in the corners the models do not reach.

### Assumptions

- The frame is part of the picture, not furniture: a rule is not something you read or click, so the requirement that protects the models is about widgets, and says so now.
- The models stay exactly where they are. The example's right-hand status block and equalizer are dropped rather than drawn over them.
- The small readouts use the palette's most orange red. It is still inside the twenty degrees of pure red the scene's rule allows, and it gives the readouts an amber-terminal note against the crimson.
- The CRT's scanline gaps are painted in the scene's darkest red rather than black, so the scene keeps to its reds; they cut the grid, the sun and the rain and leave the flat sky as it is.
- The rain is laid out once from a fixed seed, so it is the same shower every run.
- The code has shipped. This change records the menu that is in the tree; its tasks are marked for what was done and verified, and only what is genuinely left is open.
- `neo-minsk-main-menu`'s *Neo-Minsk skyline* requirement describes code that was deleted in `9f84d10`, a commit about energy and structure regeneration. That drift belongs to that change, not this one: this change's scene requirement describes the scene as it is, and does not retire a requirement it did not write.

## Capabilities

### New Capabilities

None. The menu's look and the way it is worked are part of the existing `main-menu` capability.

### Modified Capabilities

- `main-menu`:
  - *Pixel-art menu scene* names the scene's layers; the rain, the sweeping band and the scanline gaps join them, and the scene pixel size is the 2 the code actually draws at rather than the 4 the scenario still names.
  - *Menu layout splits the window* names the panel as something drawn, with bounds, rather than a notional left column.
  - *Menu scene composition is independent of the window scale* keeps the widgets off the models and lets scene chrome cross them.
  - *Menu widgets fit inside the menu panel* holds them inside the drawn panel, not merely inside the left 560 pixels.
  - New requirements for the frame and its readouts, the panel and its header, the rows drawn by hand, the main screen being worked by the keyboard, Escape backing out, the weather and the CRT, and the widgets that remain being painted in the scene's reds.

## Impact

- **Code**
  - `main_menu.go`: the frame, the panel, the hand-drawn rows, the highlight and its input, the raygui red style, the logo's ghost.
  - `menu_scene.go`: the chrome's colours beside the scene's, the sky gradient, the horizon line, the rain and the CRT.
  - `game.go`: `enterLevel` puts the dark editor style back, as `leaveMatch` already did.
- **Tests**: the panel holding the whole column and the rain's layout are checked without a window; the existing red-set test covers the chrome's colours too. The look itself is checked by running the game and capturing it.
- **Not affected**: the palette pass and the Endesga 64 quantization, where the models sit and how they fade in, the logo's fade-in timing, the level and its palette, gameplay, networking, the menu music and what the Options screen contains.

## Non-goals

- New menu screens or entries, or reordering the ones there are.
- New art assets. The frame, the panel and the rows are drawn in code.
- The example's right-hand status block and equalizer.
- Retiring `neo-minsk-main-menu`'s skyline requirement, or archiving that change.
- Switching the level to Endesga 64.
