## 1. Lights in the level file

- [x] 1.1 Before changing anything, take a screenshot of the level at a fixed camera position (a fresh single-player match, the fighter at spawn, default zoom) and keep it outside the repo, for task 2.5. Verify the file exists.
- [x] 1.2 Add `LevelLight` (`X`, `Y`, `Radius` int32, `Color` [3]uint8, `Intensity` float32, with the JSON names in the design) and `lights []LevelLight` on `LevelEditor`. Verify `go build ./...`.
- [x] 1.3 Make `LoadLevel` read the file as `map[string]json.RawMessage`: `"lights"` into `le.lights`, integer keys into layers, anything else logged and skipped; and make `SaveLevel` write the layers plus `"lights"`, leaving `marshal()` — the fingerprint — over the layers alone. Verify tests that a level with lights round-trips exactly, a file with no `"lights"` key loads with none, an unknown key is skipped, and `Hash()` is the same with and without lights.
- [x] 1.4 Add the three lights to `level_1.json` — (40, 30) radius 256, (120, 120) radius 128, (−40, 30) radius 128, each white at intensity 1 — leaving every tile as it is. Verify a test that the level loads with exactly those three, and that its `Hash()` equals the fingerprint of the same file with the `"lights"` key removed; `TestLoadLevelSource` still passes.
- [x] 1.5 Verify with a test that the dedicated server's load path (`LE.LoadLevel` on a file with lights) keeps the lights and the tiles, and that a client and server loading the same file agree on the fingerprint.

## 2. Drawing the level's lights

- [x] 2.1 Add `lightVisible(l LevelLight, view rl.Rectangle) bool` — the light's square against the camera's world rectangle — and a test for inside, outside, overlapping an edge, and overlapping only a corner.
- [x] 2.2 Add the camera's world rectangle: the corners of the scene target through `rl.GetScreenToWorld2D` with `Cam.View()`. Verify a test with a known camera that the rectangle is where it should be at zoom 1 and zoom 3.
- [x] 2.3 Replace `light1..3` and their init with one shared `ZimaLight` (512) and a `LightAccum` render texture the size of `ScreenFullTexture`; per frame compute the occluders once, clear `LightAccum`, and for each visible level light set the shared light's radius, color and intensity, `Compute`, and add the result into `LightAccum` inside `Cam.View()`. Verify `go build ./...`, and that `grep` finds no `light1`, `pos1` or `newZimaLight` call left in `game.go` but the one.
- [x] 2.4 In the scene pass, draw `LightAccum` once, additively, where the three `addLight` calls were. Verify by eye that the level is lit.
- [x] 2.5 Compare against the screenshot from 1.1, same spot and zoom: the lighting matches — the `The level looks as it did` scenario.

## 3. The light tool

- [x] 3.1 Add `toolLight` and make `Q` cycle place → rect → light → place, with `setTool` clearing the light drag too. Verify a test of the cycle.
- [x] 3.2 Add `lightSel`, `lightDrag`, and the pure operations `lightAt`, `placeLight`, `dragLight` and `deleteLight`, each setting `dirty`. Verify tests: `lightAt` prefers the radius handle, then the center, then nothing; `placeLight` inherits the selected light's radius, color and intensity, or the defaults; moving keeps the radius and resizing keeps the center; dragging the handle past the center gives radius 1; deleting the selected light clears the selection; each marks the editor unsaved.
- [ ] 3.3 Wire the light tool into `Update`: left click on nothing places, on a light selects and starts a drag, the drag follows the unrounded world cursor rounded to whole pixels at `grabTolerance(zoom)`, release ends it; right click on a light deletes it; Delete deletes the selection; all behind `overGui`. Verify by hand, and that the place and rect tools still do exactly what they did — the `Tools keep to their own` scenario.
- [x] 3.4 Draw the markers in edit mode with the light tool: circle, center cross and radius handle, the selected light in white and the others in their color at half alpha; none outside the light tool. Verify by eye — the `Markers` scenario.

## 4. The panel

- [ ] 4.1 Generalise `drawRectFields` and `nextEditField` to a slice of fields, move `TestNextEditField` to the slice form, and keep the rect tool on four fields. Verify the tests pass and, by eye, that the rect tool's fields behave as before.
- [ ] 4.2 Give the light tool its panel: which light is selected (or "LMB to place a light"), fields for X, Y, radius, R, G, B and intensity in percent, clamped to radius ≥ 1, color 0–255 and intensity 0–1000 %, and a help line. Verify by eye that typing a value sets it exactly and nothing overlaps at 1280×720 and 1920×1080 — the `Exact values` scenario.

## 5. Verify against the spec

- [x] 5.1 Run `go test ./...` and `go test -tags release ./...` and confirm both pass.
- [x] 5.2 Build with `./build.ps1` to a scratch `-Out`, and with `./build.ps1 -Linux`, and confirm both pack.
- [ ] 5.3 In `go run ./cmd`: place a red light and a dim white one, move and resize them, save, restart, and confirm they are back and lit as set — the `Color and intensity`, `Save` and `Quit without saving` scenarios.
- [ ] 5.4 Place a row of lights far from the start, walk the camera to them and back, and confirm the light at the edge of the screen neither pops in nor out, and that the frame rate with them off screen matches the frame rate without them — the `At the edge of the screen` and `Many lights` scenarios.
- [ ] 5.5 Host a match from a development build with a moved light and join it from the release build: the join is not refused — the `A changed light` scenario.

## 6. Level tests that survive editing the level

- [x] 6.1 `TestLevel1HasTheThreeLights` and `TestServerAndClientLoadTheSameLevel` pin what `level_1.json` holds (exactly three lights). Since `level_1.json` gained four more lights and a tile, both fail in both builds, so 5.1 no longer holds. Make them test behaviour rather than content. The level loads with at least one light, and its lights survive a save round trip unchanged. Its `Hash()` equals the fingerprint of the same file with `"lights"` removed. The server's load path keeps lights and tiles, and server and client agree on the fingerprint. Keep the tests of 1.3 and 1.5 as they are. Verify `go test ./...` and `go test -tags release ./...` pass with the edited `level_1.json`, and still pass after adding or moving a light in it.
