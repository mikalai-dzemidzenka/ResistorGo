## 1. Load

- [x] 1.1 In `parseLevel`, move layer-0 tiles to layer -1 (the map key and each object's `Layer`). Verify with a test: `parseLevel` of `{"blocks":[{"resource":"a"}],"0":[[0,0,0]]}` gives one tile at `objects[-1]` with `Layer == -1` and nothing at `objects[0]`
- [x] 1.2 Verify that two loads of the same layer-0 file hash equal, and that a load, save, load round trip of a `"-1"` file is byte-for-byte identical (extend the existing level IO tests)

## 2. Editor layer

- [x] 2.1 Start `le.layer` at -1 in the editor's constructor. Like the current tool, it is kept across New and Load. Verify with a test that a fresh editor's `place` puts the tile in `objects[-1]`
- [x] 2.2 Add a layer step that skips 0, bound to PageUp/PageDown in edit mode. Verify with a test that stepping up from -1 gives 1, stepping down from 1 gives -1, and that no sequence of steps lands on 0
- [ ] 2.3 Show `layer: N` in the editor panel. Verify by running the editor and checking the panel shows the layer
- [x] 2.4 Verify with a test that a tile on -1 and a tile on 1 can share a cell, and that `erase` on layer 1 leaves the -1 tile
- [x] 2.5 Fade tiles on other layers while the place tool is out. Verified by `TestOtherLayersFadeWhilePlacing`; check by eye in the editor

## 3. Draw order

- [x] 3.1 Split `LevelEditor.Draw` into a background pass (`Layer <= 0`) and a foreground pass (`Layer > 0`). The debug and editor overlays go in the foreground pass. Verify with a test on a pure helper that picks which objects each pass draws: negatives and 0 in the background pass, positives in the foreground pass, each in `objectsIndexed` order
- [ ] 3.2 In `game.go`, call the background pass where `LE.Draw()` is now, and the foreground pass after `drawPuffs` and before `EndMode2D`/`compositeLights`. Remove the TODO. Verify in game: a tile on layer 1 covers a fighter walking behind it and is lit and shaded like the rest of the level
- [ ] 3.3 Verify that collision is unchanged: `go test ./...` passes, and in game a collision tile on layer 1 can be stood on

## 4. Ship

- [x] 4.1 Load `level_1.json` in the editor, save it with Ctrl+S, and check that its tiles are now under `"-1"`
- [ ] 4.2 Build with `build.ps1` (UPX), then check that `resistor.exe` starts and the level looks the same as before
