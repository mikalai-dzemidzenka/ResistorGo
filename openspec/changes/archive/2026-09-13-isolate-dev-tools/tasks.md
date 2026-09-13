## 1. The build switch

- [x] 1.1 Add `devtools_dev.go` (`//go:build !release`, `const DevTools = true`) and `devtools_release.go` (`//go:build release`, `const DevTools = false`). Initialise `Debug` from `DevTools`, and let `` ` `` toggle it only when `DevTools` is true. Verify that `go build ./...` and `go build -tags release ./...` both succeed, and that a unit test asserts `DevTools` is true in an untagged build.
- [x] 1.2 Add a `//go:build release` test file asserting that `DevTools` and `Debug` are both false. Verify that `go test -tags release ./...` passes.
- [x] 1.3 Add `-tags release` to the `go build` line in `build.ps1`. Verify that `build.ps1` still produces and packs `resistor.exe`.

## 2. Edit mode

- [x] 2.1 Add the `Editing` flag and its toggle, pressing F2 in `runLevel`. It only turns on when `DevTools` is true, the match is offline and time manipulation is off. R does nothing while `Editing` is on. Verify with unit tests on the toggle, which takes its inputs as parameters as `UpdateTime` does: it turns on and off; it refuses while time manipulation is on; it refuses in a networked match; and R while editing does not start time manipulation.
- [x] 2.2 Pause the match while `Editing` is on: skip `stepOffline` without accumulating time, skip `latchInput`, and clear `pressedAttack` and `pressedJump` when edit mode is left. Verify with a unit test that leaving edit mode clears a latched attack. Then verify in game: fighters stay frozen in edit mode, no attack fires on leaving it, and there is no catch-up after ten seconds in it.
- [x] 2.3 Call `LE.Update` and `LE.DrawGui` only while `Editing` is on, replacing the `Debug` gates at `game.go:312` and `:412`. Verify in game with the overlay on:
  - attacking many times leaves `level_1.json` unchanged by `git diff`, with no pending edit;
  - a right-click in play leaves the floor tile in place;
  - in edit mode, LMB places, RMB deletes and Tab opens the picker.
- [x] 2.4 Save on Ctrl+S inside `LE.Update`, clearing `dirty`, and remove the save on the way out of `Game()`. Verify with a unit test that the save path writes the level to a temporary path and clears `dirty`. Then verify in game:
  - placing a tile and pressing Ctrl+S puts it in the file;
  - placing a tile and quitting leaves the file unchanged;
  - an edit survives leaving and re-entering edit mode, still marked unsaved.
- [x] 2.5 Draw the marker after the palette pass, reading "EDIT", or "EDIT — unsaved" while `dirty` is set, only while `Editing` is on. Verify in game that it shows both states and disappears when edit mode is left.

## 3. What a release build contains

- [x] 3.1 Guard the remaining developer paths with `DevTools`: F5–F8 (`Cli.Sim.controls()`), the overlay and stats draws, the editor GUI, and `SaveLevel`, which returns without writing when `DevTools` is false. Verify that `go vet -tags release ./...` is clean, and add a release-tag unit test that `SaveLevel` to a temporary path writes nothing.
- [x] 3.2 Make `LoadLevel` read the embedded level directly when `DevTools` is false. Verify with a test that runs in both builds, from a temporary working directory holding a modified `level_1.json`. The development build loads the modified level, so its `LE.Hash()` differs from the embedded one. The release build loads the embedded level, so its hash equals the embedded one. The dedicated server loads through the same function.
- [x] 3.3 Run the `resistor.exe` built by `build.ps1` from a temporary directory. Verify that F2, `` ` `` and F5–F8 do nothing, and that attacking, right-clicking and quitting leave no `level_1.json` in that directory.

## 4. Verify against the spec

- [x] 4.1 Run `go test ./...` and `go test -tags release ./...` and confirm both pass.
- [x] 4.2 Play a development build through the `specs/dev-tools/spec.md` scenarios that need a window and are not covered above: F2 is refused during time manipulation; a saved edit is loaded after a restart; the overlay is still on by default and toggles with `` ` ``.
- [x] 4.3 Check the modified session scenarios in a networked match of a development build: F2 and R do nothing, the overlay toggles without affecting the match, and quitting leaves the level file unchanged.
