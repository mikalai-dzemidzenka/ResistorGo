## 1. Which tiles follow their texture

- [x] 1.1 Add the unexported `own bool` field to `LevelObject` and a `sameRects(a, b ObjectRects) bool` helper — nil equals only nil, values compare by value. Verify a test over both-nil, one-nil, equal and differing rectangles, and that `json.Marshal` of a `LevelObject` with `own` set is byte-identical to one without.
- [x] 1.2 Add `markOwnRects()`: a tile is own when its texture has no palette entry or its rectangles differ from the entry's, and follows otherwise; call it at the end of `LoadLevel`. Verify a test that a matching tile follows, a differing tile and a tile of an unknown texture are own, and no rectangle changes.
- [x] 1.3 Verify with a test that loading leaves the level fingerprint as it was: `Hash()` after `LoadLevel` equals the hash of the file's bytes as unmarshalled without classification, and `TestLoadLevelSource` still passes — the `The fingerprint is unaffected` and `An existing level loads unchanged` scenarios.
- [x] 1.4 Leave `own` false in `place`, and set it true in the tile branch of `setSelectedRects`. Verify a test that a placed tile follows, and that reshaping, adding, clearing and pasting onto it each make it own.

## 2. A texture's change reaches its tiles

- [x] 2.1 In the proto branch of `setSelectedRects`, give every following tile of the same texture, on every layer, a clone of the new rectangles, then rebuild the index once. Verify a test that following tiles of that texture change, own tiles of it do not, tiles of another texture do not, and no tile shares a rectangle pointer with the entry — the `Placed tiles follow their texture`, `Other textures are untouched` and `A hand-edited tile keeps its own` scenarios.
- [x] 2.2 Update `TestDefaultsAndTilesAreEditedSeparately` from the in-flight change, whose first half asserts the old "new tiles only" behaviour, to the new requirement. Verify it fails against the code before 2.1 and passes after.

## 3. Following again

- [x] 3.1 Add `followTexture()`: for a selected own tile whose texture has a palette entry, write the entry's rectangles (cloned) and `own = false` without going through the tile branch, marking dirty only if the rectangles changed. Verify a test that it resets an own tile, that a following tile, a palette selection and a tile of an unknown texture are left alone, and that a later texture change then reaches the tile — the `Back to the texture` scenario.
- [ ] 3.2 Show the selected tile's state in the panel: "follows texture", or "own rectangles (F: follow texture)". Verify by eye through reshape, then F — the `Shown` scenario.

## 4. Copy and paste

- [x] 4.1 Add `clip`, `clipFrom` and `hasClip` to `LevelEditor`, with `copySelection()` and `pasteSelection()`; paste goes through `setSelectedRects` with a clone, is guarded by `sameRects`, and does nothing with no copy or no selection. Verify tests for: texture to texture, a copy not a link, paste onto a texture reaching its following tiles, paste onto a tile making it own, absence copied, nothing to paste, and a paste that changes nothing leaving the editor clean.
- [ ] 4.2 In `Update`, read Ctrl once: with Ctrl, S saves, C copies, V pastes; the bare letter keys (C, L, N, Q, Tab, F) act only without Ctrl, Delete as before. Handle copy, paste and F before the picker's early return, and skip them all while a value box has the keyboard. Verify by hand that Ctrl+C while editing the light rectangle leaves it active — the `Copy does not retool` scenario — and that copy and paste work with the picker open.
- [ ] 4.3 Add the "Copied: `<name>`" line to the panel once something has been copied. Verify by eye.

## 5. Verify against the spec

- [x] 5.1 Run `go test ./...` and `go test -tags release ./...` and confirm both pass.
- [x] 5.2 Build with `./build.ps1` (to a scratch `-Out`, so the tracked `resistor.exe` is left alone) and confirm it packs.
- [ ] 5.3 In `go run ./cmd`: shape `mn_wall1`'s rectangles, copy them, and paste onto the other seven `mn_*` textures from the open picker; place tiles of several of them; reshape `mn_wall1` and confirm its tiles follow and the others do not.
- [ ] 5.4 Reshape one tile by hand, change its texture and confirm that tile keeps its shape; press F and confirm it takes the texture's; save, restart, and confirm a hand-edited tile is still own — the `A hand-edited tile survives a restart` scenario.
