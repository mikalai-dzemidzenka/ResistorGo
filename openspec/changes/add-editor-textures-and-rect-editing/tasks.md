## 1. Textures the editor can reach

- [x] 1.1 Split `loadTex` in `resources.go` into `loadTexBytes(name string, data []byte)` and keep `loadTex(name)` as the embedded caller, so the decode path has one implementation. Verify `go build ./...` succeeds and the game still starts with every texture drawn.
- [x] 1.2 Add `loadTexDisk(name, diskPath string)` beside it, reading with `os.ReadFile` and registering under `name`, guarded so a release build never calls it. Verify `go build -tags release ./...` succeeds and that nothing in the release path references it.
- [x] 1.3 Add `mergeTextureLists(embedded, disk []string) []string` — merge, deduplicate, sort by path — and a test that a disk-only name appears, a name in both appears once, and the order is the same whatever order the inputs arrive in.
- [x] 1.4 Add `blockTextures()` using `fs.Glob(assetFS, "resources/blocks/*.png")`, plus `os.ReadDir("resources/blocks")` only when `DevTools`, through `mergeTextureLists`. Verify a test that it returns every PNG in `resources/blocks` — seven today, not three — and that it touches no GL call, so it is safe from package init.
- [x] 1.5 Add `embeddedBlocks` — the set built once from the `assetFS` glob — and a helper that answers whether a path is part of the build. Verify a test that a name from `resources/blocks` is in the build and an invented one is not.
- [x] 1.6 Make `LoadTextures()` load every path `blockTextures()` returns instead of the three hardcoded block names, leaving the animation and menu loads alone. Verify the game starts and a texture that had no Go literal (for example `e1m1_bg1.png`) is in the registry.

## 2. The palette, with defaults from a file

- [x] 2.1 Add the defaults type — the two `*rl.Rectangle` fields keyed by texture path — and `loadObjectDefaults()` / `saveObjectDefaults()` for `level_objects.json`, reading and writing only when `DevTools`, falling back and logging on a missing or malformed file the way `LoadSettings` does. Verify a test that a round trip preserves both rectangles and a nil one, and that malformed JSON yields the fallback rather than an error path that stops the editor.
- [x] 2.2 Build the palette from `blockTextures()`, resolving each entry's rectangles in order: the defaults file, then the built-in rectangles still returned by `LoadLevelObjects()`, then none. Verify a test that a texture with a built-in keeps it, a texture in the defaults file takes the file's, and a texture in neither gets nil rectangles.
- [x] 2.3 Make `NewLevelEditor` tolerate an empty palette instead of indexing `[0]`, and verify a test that it returns a usable editor when no block texture is found.
- [ ] 2.4 Verify by starting the game and opening the picker with Tab that every block texture is offered, in the same order across two runs — the `Every texture is offered` and `Stable order` scenarios.

## 3. Placing, drawing and what is missing

- [ ] 3.1 Have `LE.Draw` look up textures with `tex, ok := resources[obj.Resource]` and, when `!ok` in a development build, draw a marker with the object's path where the tile would be, leaving collision untouched. Verify by hand-editing a level copy to name a texture that does not exist: the marker is drawn there and the fighter still collides with it — the `Level outran its assets` scenario.
- [ ] 3.2 Mark palette entries and placed objects whose texture is not part of the build, using the helper from 1.5. Verify a test that the marking follows the embedded set, and by eye in the picker after an import.
- [x] 3.3 Make a placement copy the selected entry's rectangles including their absence — a nil rectangle on the entry places an object with no rectangle, rather than dereferencing it as `Update` does today. Verify a test that placing an entry with nil rectangles yields an object with nil rectangles and does not panic.

## 4. Importing a dropped texture

- [x] 4.1 Add `droppedTextureName(path string) (string, bool)`: accept `.png` only, case-insensitively, and reject anything else. Verify a test over a `.png`, a `.PNG`, a `.txt` and a path with no extension.
- [x] 4.2 Add the import step: load the file from the dropped path first to prove it decodes, then copy it into `resources/blocks/` unless the name is already taken, register it, rebuild the palette and select it. Verify a test of the name-taken branch — no write, existing entry selected — and of the decode-fails branch leaving nothing copied.
- [ ] 4.3 Poll drops in the main loop in `game.go`: always `LoadDroppedFiles`/`UnloadDroppedFiles` so nothing stays queued, and only act when `DevTools && Editing`. Verify by dropping a PNG during a match, then entering edit mode: nothing is imported — the `A dropped file in play does nothing` scenario.
- [ ] 4.4 Verify by hand: drop a PNG in edit mode, place it immediately, save with Ctrl+S, restart — the `Import and place` and `Import survives a restart` scenarios; drop a text file and confirm the palette and selection are unchanged.

## 5. Two tools and a selection

- [ ] 5.1 Add the tool state to `LevelEditor` with `Q` to switch, and keep every existing place-tool binding as it is. Verify by hand that the place tool still places on the left button and deletes on the right, and that the rect tool does neither — the `Editing` and `The rect tool does not place` scenarios.
- [x] 5.2 Add the selection: either the palette entry or one placed object addressed by layer and grid position, so it survives `buildIndex`. Verify a test that a selected object is still the selected object after a placement elsewhere rebuilds the index.
- [x] 5.3 Make a left click in the rect tool that hits no handle select the placed object under the cursor, and leave the previous selection alone when it hits nothing. Verify a test of both branches over a small object map.
- [ ] 5.4 Park the ghost while the rect tool is active — `updateCursor` stops following the mouse and the prototype stays on the cell it was on when the tool was entered — and let it follow again on leaving. Verify by hand that the prototype's rectangles can be dragged without the ghost moving.
- [ ] 5.5 Add `C` / `L` to choose the active rectangle, keeping the selection, and draw both rectangles distinguishably with the active one emphasised. Verify by eye that switching does not clear the selection.

## 6. Shaping a rectangle

- [x] 6.1 Add `rectHandleAt(r rl.Rectangle, p rl.Vector2, tol float32) handle` — corners, then edges, then body, then none — and a test over a known rectangle for all nine outcomes, including a point just outside the tolerance.
- [x] 6.2 Add `dragRect(r rl.Rectangle, h handle, to rl.Vector2) rl.Rectangle` and a test that a body drag translates without resizing, an edge drag moves only that side, a corner drag leaves the opposite corner fixed, and dragging an edge past its opposite leaves a one-pixel, non-inverted rectangle — the `Resize by a corner`, `Move the body` and `Never inverted` scenarios.
- [x] 6.3 Wire dragging into `LE.Update`: unrounded `rl.GetScreenToWorld2D(MousePos, Cam.View())` rounded to whole pixels, tolerance `max(1, 4/zoom)`, the grab offset stored so the rectangle does not jump, the release ending the drag, and the stored value kept relative to the object origin. Verify a test of the world-to-relative conversion at a negative origin, where the old truncating conversion is wrong.
- [ ] 6.4 Add `N` to give the selection the active rectangle as one grid cell at its origin, and `Delete` to clear it. Verify a test that adding yields a 39x83 rectangle at the origin and clearing yields nil, and by hand that a cleared collision rectangle lets the fighter through and a cleared light rectangle stops the shadow — the `Clear one` and `No shadow` scenarios.
- [ ] 6.5 Make every edit — an instance rectangle, a palette default, a placement, a deletion — set the existing `dirty` flag, and make `Ctrl+S` write `level_1.json` and then `level_objects.json`. Verify a test that a rectangle edit marks the editor dirty, and by hand that the EDIT marker says unsaved until Ctrl+S — the `A rectangle edit is an edit` scenario.
- [ ] 6.6 Verify by hand that a palette default change applies only to tiles placed after it and that an instance change applies only to that tile — the `Defaults apply to new tiles only` and `One tile only` scenarios.

## 7. The panel, and one owner per click

- [ ] 7.1 Draw the rect panel in `LE.DrawGui` in the bottom-left of the virtual screen, clear of the picker strip, the debug text and the EDIT marker: the tool, what is selected, whether its texture is in the build, which rectangle is active, and `rlg.ValueBox` fields for X/Y/W/H. Verify by eye at 1280x720 and at 1920x1080 that nothing overlaps and the fields track a drag.
- [ ] 7.2 Compute the panel and picker bounds in one place, use them for both drawing and input, and have `LE.Update` ignore world mouse input while the cursor is inside them. Verify a test of the bounds check, and by hand that clicking a value box neither places a tile nor starts a drag.
- [ ] 7.3 Write back a changed value box with the same minimum-1 clamp as a drag, and ignore the editor's letter keys while a box has keyboard focus. Verify by hand that typing a width sets it exactly and does not switch tools — the `Exact values` scenario.

## 8. Verify against the spec

- [x] 8.1 Run `go test ./...` and `go test -tags release ./...` and confirm both pass.
- [ ] 8.2 Build the release binary with `./build.ps1` and confirm it packs; then with that binary confirm F2, `` ` ``, Tab and F5–F8 do nothing, a dropped PNG does nothing and is not copied anywhere, an extra PNG placed beside the others on disk is ignored, and neither `level_1.json` nor `level_objects.json` is written — the release scenarios in the dev-tools delta.
- [ ] 8.3 Play a development build (`go run ./cmd`): shape a placed tile's light rectangle and confirm the shadow follows it as it is dragged, then shape its collision rectangle, leave edit mode and confirm the fighter collides with the new shape — the `The light follows` and `The collision follows` scenarios.
- [ ] 8.4 Confirm dragging works at every zoom level from 1 to 5, including on the negative side of the origin, and that no drag can invert a rectangle — the `Zoomed in` and `Never inverted` scenarios.
- [ ] 8.5 Confirm the whole loop end to end: drop a new PNG, shape its palette defaults, place several tiles, reshape one of them, save with Ctrl+S, restart and confirm the level, the defaults and the imported texture all came back.
