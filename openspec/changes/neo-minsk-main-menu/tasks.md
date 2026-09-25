## 1. Palette

- [x] 1.1 Add `EndesgaPalette []rl.Color` to `palette.go`, with the 64 colours of the published Endesga 64 list on Lospec, in its order. Verify with a unit test that there are exactly 64 entries, all distinct, and that the first and last entries match the published list.
- [x] 1.2 Load `PaletteShader` and a new `MenuPaletteTexture`, built from `EndesgaPalette`, once at start-up beside `initPresent`, so they exist before the menu first draws. The level keeps using `ScreenPaletteTexture` built from `Palette`. Verify that `go build ./...` succeeds, and that a single-player match still starts and uses the level palette.

## 2. Scene pieces, without a window

- [x] 2.1 Add `menu_city.go` with a pure skyline layout: the five landmarks (library diamond, obelisk, Gates of Minsk twin towers, Independence Avenue blocks, TV tower) at fixed x positions, seeded filler buildings between them, and fixed lit windows per building. Add a separate function that draws the layout. Verify with unit tests:
  - every landmark is present, inside the scene, and not covered by another building;
  - every building top stays below the sky margin;
  - every base sits on the horizon;
  - the same seed gives an identical layout.
- [x] 2.2 Add `menuRedIndexes`, the Endesga 64 entries the scene may be painted from, and take every scene colour from it: sky bands, sun, silhouette, ground, both colour ramps and the model colours. Verify with a unit test that every entry of the list has a hue within 20 degrees of pure red, and that each of those scene colours is one of them.
- [x] 2.3 Add the colour cycle: a pure function from role (grid, window) and time to a colour, stepping through a short ramp of `menuRedIndexes` entries. Verify with a unit test that, sampled over a minute of time, every returned colour is in `EndesgaPalette` and in the red set, and that each role changes colour at least once within ten seconds. Remove the roll, sway and overscan functions and their tests.

## 3. Draw it

- [x] 3.1 Draw `MenuBackground`, a scene-sized texture: a dusk sky gradient with a sun, the skyline from 2.1 as flat silhouettes with lit windows and no outlines, then the perspective neon grid. Its cross lines sit at `y = horizon + depth/(k - frac)`, so the scroll is seamless. Colours come from 2.3. This replaces the red line fan. Verify in game that the grid converges toward the horizon and scrolls toward the viewer without a jump, that the skyline and sky show behind it, and that no building has an outline.
- [x] 3.2 Compose `MainMenuTexture`: the background drawn at the origin with no rotation or offset, then the five wireframe models drawn in palette reds instead of plain red. Remove the `rl.Red` tint on the upscale. Verify in game that the models are where they were and that the background holds still.
- [x] 3.3 Draw the menu frame into a 1280×720 `MenuFrame` texture: the scene upscaled by `menuPixel`, the logo, then `drawMenuUI()`. Then draw `MenuFrame` into `VirtualScreen` through `PaletteShader` with `MenuPaletteTexture`, `pixelSize` 1 and dithering off. Verify that `go build ./...` succeeds, and that in game every menu screen draws and its buttons respond.

## 4. Verify against the spec

- [x] 4.1 Run `go test ./` and `go test -tags release ./` and confirm both pass, including the existing `TestMenuLayout` unchanged and the new palette, red-set, skyline and cycle tests.
- [x] 4.2 Capture the menu at several moments, one of them mid colour change, on the Main, Host Game with an error shown, Join Game and Options screens. Check with a pixel script that every pixel of every capture is one of the 64 Endesga 64 colours. Capture a single-player match and check that its pixels are still from the level palette.
- [x] 4.3 Capture the menu ten seconds apart and check that the skyline is at the same place in both and the horizon is level, so the background is not moving. Check that no pixel of the window is left uncovered by the background.
- [ ] 4.4 Tune the cycle speed and the ramp entries by eye in the running game, and record any values that change in design.md.

## Before archiving

- [x] Merge the `main-menu` delta by hand, not by replacing blocks: `add-menu-music-and-options` (archived 2026-09-23) rewrote "Pixel-art menu scene" and "Menu widgets fit inside the menu panel" after this delta was written - the virtual screen instead of the window, and the Options screen with its fullscreen, window size and music volume rows. Keep those edits alongside this change's; verify with `openspec validate --specs --strict`
