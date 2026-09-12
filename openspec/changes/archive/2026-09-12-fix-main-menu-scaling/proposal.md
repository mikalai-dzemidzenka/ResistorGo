## Why

The main menu is drawn at the wrong scale. The "RESISTOR" logo overflows both edges of the window, so only the middle letters show. The wireframe head and fist/kick models are so large that they fill the background as blocky red noise. The Host Game screen's last button (Back) is also partly below the bottom edge of the window.

The cause is the commit that moved the menu to the 160x90 virtual resolution (f1454ab). It started upscaling the menu render texture by `WindowScale` (8) when drawing it to the window. It also kept multiplying every 3D model size, model position, background-line endpoint and the logo's draw scale by `WindowScale`. So everything is scaled twice. The logo is 209x28 px, and at ×8 it becomes 1672 px wide in a 1280 px window. The Host screen gained a sixth row in the spectator-mode change, which pushed its layout past 720 px.

Review during implementation led to a layout change. With the menu centred, the free space around the logo and buttons is too small for the models to read well. So the menu moves to the left and the models get the right side of the window.

## What Changes

- 3D menu scene (background line fan and the head/jab/kick/martelo wireframes): stop multiplying world-space sizes and positions by `WindowScale`. The ×8 upscale of the 160x90 texture stays, so the scene keeps its chunky pixel look.
- Layout: split the window. The logo, buttons and error messages move into a panel on the left (560 of 1280 px). The right side becomes the stage for the models: the head at the top with a jab on each side, and the martelo and kick below. Each model is as large as fits there without being cut off or covered: head 1.3, jabs 0.55, martelo and kick 0.7, up from the original 0.75 and 0.25.
- Scene pixel size: add a `menuPixel` setting, separate from the level's `WindowScale`, and set it to 4 (a 320x180 scene) instead of the old 8. The models keep hard pixel edges but have four times the pixels, so their wireframes read properly.
- Logo: drawn at an integer scale that fits the panel when centred over the buttons (×2 for the current 209 px asset), above the buttons.
- Menu buttons and fields: rows are 40 px tall with 8 px gaps, inside the panel. The tallest screen (Host Game, six rows) and the error line fit with a margin. The entries and their order are unchanged.
- Add a small unit test for the pure layout math (row rectangles, logo scale), so a future row or asset change that no longer fits fails in `go test`.

## Capabilities

### New Capabilities
- `main-menu`: How the main menu is laid out and presented: the animated 3D scene, the logo and the menu widgets are scaled for the virtual resolution, placed side by side, and fit inside the game window.

### Modified Capabilities
<!-- none: multiplayer/session "Main menu entries" (which entries exist, mouse usable) is unchanged -->

## Impact

- Code: `main_menu.go` only (`runMainMenu`, `drawMenuUI`, `head`, `jab1`, `jab2`, `kick`, `martelo`). A new `main_menu_test.go` covers the layout helpers.
- No change to menu entries, navigation, networking, or the level/game render pipeline.
- No new dependencies or assets.
