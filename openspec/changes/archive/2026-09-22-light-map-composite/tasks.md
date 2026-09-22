## 1. A picture of what is being replaced

- [x] 1.1 ~~Before changing anything, take screenshots of the level from three fixed camera positions — a fresh single-player match at spawn, a spot deep in a tile's shadow, and a spot lit by two lights at once — at default zoom, and keep them outside the repo for tasks 3.4 and 5.2. Verify the files exist.~~ Dropped: the composite was committed (`d1282fb`) before the screenshots were taken, so 3.4 and 5.2 are judged by eye on their own merits instead of against a picture of the old look.

## 2. The level carries its ambient

- [x] 2.1 Add `Ambient` (`Color [3]uint8`, `Strength float32`, JSON `color` and `strength`), `defaultAmbient`, and the pure `ambientClear(a Ambient) rl.Color` — `color × strength` per channel, clamped to 0–255, alpha 255. Verify with a test: a mid blue at 25 % gives the expected channels, strength 0 gives opaque black, strength over 100 % saturates rather than wrapping, and the alpha is always 255.
- [x] 2.2 Make `parseLevel` read the `"ambient"` key into the level's ambient, distinguishing a missing key (the default ambient) from a present one (used as written, strength 0 included), and make `marshalFile` write it beside `"lights"`. Verify with tests that a level with an ambient round-trips exactly, a file with no `"ambient"` key loads with `defaultAmbient`, a file with `"strength": 0` loads with strength 0 rather than the default, and an unknown key is still logged and skipped.
- [x] 2.3 Verify with a test that `marshal()` — the fingerprint — is unchanged by the ambient: the same level with an ambient, without one, and with a different one has one `Hash()`.
- [x] 2.4 Verify with a test that the dedicated server's load path (`LE.LoadLevel` on a file with an ambient) keeps the ambient, the lights and the tiles, and that a client and a server loading the same file still agree on the fingerprint.

## 3. The composite

- [x] 3.1 Clear `LightAccum` to `ambientClear(LE.ambient)` instead of `rl.Black` in `computeLights`, leaving the per-light accumulation exactly as it is. Verify `go build ./...` and, by eye, that the level is lit and its dark areas now carry the ambient's colour.
- [x] 3.2 Replace the alpha-50 black rectangle and the additive draw in the scene pass with the composite: one `rl.BlendMultiplied` draw of `LightAccum` with a white tint, then one `rl.BlendAdditive` draw of it tinted `{255, 255, 255, glow×255}` with `glow` a named constant (start at 0.2). Verify `grep` finds no `rl.NewColor(0, 0, 0, 50)` left in `game.go`, and that `go build ./...` passes.
- [x] 3.3 Verify by eye that the overlays are untouched: a fighter with a structure bar, an energy bar and an attack outline in an unlit corner shows them in the same colours, and at the same brightness, as in a lit one — the `The overlays keep their colours` scenario. Covered by `add-lit-areas` 6.6.
- [x] 3.4 Verify by eye: a coloured tile in shadow is a darker shade of its own colour rather than a grey wash, a bright light against the black background lights the tile without whitening the background, and a light with nothing near it still glows — the `A shadow keeps its colour`, `Light does not make black white`, `A light in the air` and `Darkness deepens with the shadows` scenarios. Settle `glow` here. Covered by `add-lit-areas` 6.6.

## 4. The ambient in the light panel

- [x] 4.1 Add `setAmbient` on the editor — replaces the ambient if that changes anything and marks the level dirty — and verify with a test that it sets colour and strength, that setting the same ambient again leaves the level clean, and that a change marks it unsaved.
- [x] 4.2 Show the ambient in `drawLightPanel` when no light is selected: fields for R, G, B (0–255) and strength as a percentage (0–1000), fed through the existing integer fields, with the strength written back only when the typed percentage differs, so rounding never rewrites it — the same guard the light's intensity uses. Verify by eye that typing a value sets it exactly, that selecting a light swaps the panel to that light's fields and deselecting brings the ambient back, and that nothing overlaps at 1280×720 or 1920×1080 — the `The ambient fields` and `A light takes the panel` scenarios. (Built: `drawLightPanel` shows the ambient when no light is selected. The check by hand or by eye is `add-lit-areas` 6.4 and 6.7.)
- [x] 4.3 Verify by hand that an ambient edit marks the level unsaved, that Ctrl+S then a restart brings it back, and that quitting without Ctrl+S leaves `level_1.json` byte-identical — the `Saved with the level` and `Quit without saving` scenarios. Covered by `add-lit-areas` 6.8.
- [x] 4.4 Let go of the selected light with a right click on empty space (`deselectLight`, called from `updateLightTool` when no light is under the cursor), so the ambient fields can be reached again once a light has been touched, and say so in the panel's help line. Verify with a test that letting go clears the selection, the drag and the field focus, removes no light, and leaves the level as saved as it was. (Added during apply: the light tool had no way to deselect, which the `A light takes the panel` scenario assumes.)

## 5. Re-tune the level

- [x] 5.1 Author `level_1.json`'s ambient with the light tool and save it, then re-tune any light whose intensity or radius reads wrong under the new composite. Verify the file's tiles and its `Hash()` are unchanged by the edit, and that `go test ./...` still passes with the edited level.
- [x] 5.2 Walk the whole level at play zoom and verify by eye that no place the level expects a fight in is dark enough to lose a fighter in. Verify a gradient at the edge of a light's reach does not visibly step; if it does, lighten the ambient rather than reaching for dithering, which is out of scope. Covered by `add-lit-areas` 6.6.
- [x] 5.3 Set the ambient strength to 0 in the editor and verify by eye that everything outside the lights is black, then set a deep blue ambient and verify the unlit parts take that blue — the `No ambient at all` and `A coloured night` scenarios. Restore the tuned ambient and save. Covered by `add-lit-areas` 6.4.

## 6. Verify against the spec

- [x] 6.1 Run `go test ./...` and `go test -tags release ./...` and confirm both pass, apart from `TestWallsStopFighters/running`, which fails the same way on a clean tree and belongs to `horizontal-collision` - it is not this change. Covered by `add-lit-areas` 6.1.
- [x] 6.2 Build with `./build.ps1` to a scratch `-Out`, and with `./build.ps1 -Linux`, and confirm both pack.
- [x] 6.3 Host a match from a development build whose level has a changed ambient and join it from the release build: the join is not refused — the `A changed ambient` scenario. Covered by `add-lit-areas` 6.10.
- [x] 6.4 Verify with a level of twenty lights, most of them off screen, that the frame rate matches the same level with those lights deleted, and that a light entering the screen's edge neither pops in nor out — the `Many lights` and `At the edge of the screen` scenarios still hold under the composite. Covered by `add-lit-areas` 6.9.
- [x] 6.5 Before archiving, confirm `add-editor-textures-and-rect-editing` and `add-level-lights` are archived first — this change's deltas target the `lighting` and `level-editor` capabilities those two introduce. Verify `openspec validate light-map-composite --strict` reports no archive warning once they are.
