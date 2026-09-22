## 1. A list of collision rectangles

- [x] 1.1 Change `LevelObject.CollisionRect *rl.Rectangle` to `CollisionRects []rl.Rectangle` and `ObjectRects.Collision` to `[]rl.Rectangle`, with `clone()` copying the slice and `sameRects()` comparing length and each rectangle in order. Turn the three built-in single rectangles in `LoadLevelObjects()` into one-element lists. Verify `go build ./...` fails only where the old field is read, which is task 1.2's list.
- [x] 1.2 Move every reader to the list, without changing behaviour: the editor and palette code, `drawSelection`, the debug overlay and `DrawCursor` drawing every rectangle, and the tests (about sixty references). A test asserting one rectangle asserts a one-element list. Verify `go test ./...` and `go test -tags release ./...` pass, with no test weakened.
- [x] 1.3 Give `LevelObject` and `ObjectRects` `MarshalJSON`/`UnmarshalJSON` through a wire struct in today's field order: one rectangle as `"collisionRect"`, two or more as `"collisionRects"`, none as neither; on load `"collisionRect"` gives a list of one and `"collisionRects"` the list, and with both the list wins and the load logs it. Verify tests of the round trip for zero, one and three rectangles, and of a file with both keys.
- [x] 1.4 Verify with a test that a one-rectangle object marshals to exactly the bytes a copy of the old struct gives, and that `level_1.json` loads and re-marshals to its own bytes with an unchanged `Hash()` — the `The existing level` scenario. Verify `TestLoadLevelSource` and `TestLevel1HasTheThreeLights` still pass.
## 2. Collision with every rectangle

- [x] 2.1 Build `le.solids` in `buildIndex`: every collision rectangle in world space, in `objectsIndexed` order and list order within an object. Verify a test that a level of one-rectangle tiles gives the same rectangles in the same order the old loops visited, and that a tile with three gives three in its list order.
- [x] 2.2 Make `headBlocked` and `checkCollision` in `actor.go` loop over `LE.solids` with their bodies unchanged, and change nothing else in that file. Verify with `git diff actor.go` that the only new hunks are those two loop headers, and that `jump_test.go` and `hit_run_test.go` still pass.
  - This change's hunk is only the two loops. The rest of `git diff actor.go` (run-jump lift, energy, `face()`) is other uncommitted work, so commit this hunk on its own.
- [x] 2.3 Add actor tests with a tile of two collision rectangles and a gap between them: a fighter stands on the first, falls into the gap, and stands on the second; and a fighter jumping at the second rectangle's edge catches the ledge — the `Stand on either` and `A ledge on the second` scenarios.
- [x] 2.4 Verify with the existing networking tests that client prediction and the server agree over a level holding a two-rectangle tile, with no correction from walking across it — the `Online` scenario.

## 3. The rect tool

- [x] 3.1 Add `activeColl`, make `active()` and `setActive()` address the collision rectangle at that index, and clamp it whenever the selection changes. Verify tests that `setActive(nil)` removes only that rectangle, and that a selection with fewer rectangles never leaves the index out of range.
- [x] 3.2 Make `beginDrag` hit-test the active rectangle first, then every other rectangle of the selection, collision in order and then light, making the first hit active. Verify tests: clicking the third of three makes it active, a drag moves only it, and on an overlap the already active one wins — the `Pick another collision rectangle` and `Only the active one moves` scenarios.
- [x] 3.3 Make `C` step through the collision rectangles when collision is already active, `N` append a one-cell rectangle at the origin and make it active, and Delete remove the active one with the index moving to the previous one. Verify tests of stepping with wrap-around, adding a second, and removing one of three — the `Step through`, `Add a second collision rectangle` and `Remove one of several` scenarios.
- [x] 3.4 Show "Collision k of n (C: next)" in the panel, draw every collision rectangle of the selection with the active one filled, and every collision rectangle of every object in the debug overlay. Verify by eye in a development build.

## 4. Everything carries the set

- [x] 4.1 Verify with tests that placing, texture defaults and their file, following a texture (including `markOwnRects` and `followTexture`), copy and paste, and saving and loading each carry all of an object's collision rectangles in order — the `Placed`, `Copied`, `Followed` and `Saved with all of them` scenarios.

## 5. Verify against the spec

- [x] 5.1 Run `go test ./...` and `go test -tags release ./...` and confirm both pass.
- [x] 5.2 Build with `./build.ps1` to a scratch `-Out`, and with `./build.ps1 -Linux`, and confirm both pack.
- [x] 5.3 In `go run ./cmd`: give a tile two collision rectangles with a gap, walk across it, jump at the second rectangle's ledge, save, restart and confirm the tile still has both; confirm a level with no multi-rectangle tile saves to the same bytes.
