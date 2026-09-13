## Why

Every attack in single player is also a level edit. The debug overlay is on by default (`Debug = true`), and with it the in-game level editor takes input on every frame of an offline match. The editor and the fight share one mouse:

- **LMB** attacks, and at the same time places the editor's selected object at the aimed grid cell.
- **RMB** is bound to nothing in play, and deletes whatever tile is at the aimed cell, the floor included.
- **Quitting** saves `level_1.json` if anything was touched.

The level loads disk-first, so the edited file then replaces the embedded level on every later run. It also changes the level fingerprint the multiplayer handshake compares, so that player is refused with "level mismatch" when joining anyone else's server. This is live in the shipped exe, and it bit development too: automated playtests of this project changed `level_1.json` by attacking.

The shipped exe also carries the rest of the developer kit. That is the `` ` `` overlay (collision boxes, state text, fps, net stats) and the F5–F8 keys that change a client's simulated latency and packet loss.

## What Changes

- **An explicit edit mode.** In single player, in a development build, F2 enters and leaves edit mode. While it is on:
  - the match is paused: no simulation, no bot decisions, no attack input;
  - LMB places the selected object, RMB deletes, Tab opens the object picker;
  - a marker on screen says the game is in edit mode, and whether there are unsaved edits.

  Outside edit mode the editor takes no input at all: LMB only attacks, and RMB and Tab do nothing.
- **BREAKING (development workflow): saving is explicit.** Ctrl+S in edit mode writes `level_1.json`. Quitting no longer saves anything; unsaved edits are dropped.
- **A release build without the developer kit.** `build.ps1` builds with a `release` build tag. The shipped exe has no editor or edit mode, never writes `level_1.json`, has no F5–F8 network-simulation keys and no `` ` `` overlay. It loads the embedded level only, ignoring any `level_1.json` lying beside it, whether it runs as a game or as a dedicated server.
- **Development builds keep everything.** `go run ./cmd` and `go test ./...` build without the tag. The overlay stays on by default there, which is now harmless, because it no longer brings the editor's input with it.

### Assumptions

- F2 is the edit-mode key. It is unbound today; Tab stays the object picker, now only inside edit mode.
- Edit mode and time manipulation exclude each other. R does nothing in edit mode, and F2 does nothing while time manipulation is on.
- Leaving edit mode keeps unsaved edits in memory for the session. Only quitting drops them.
- Editing uses the same captured cursor as play; the selected object is placed on the grid cell under it, as the editor does today.
- A release dedicated server loses the per-player view-tick log line it prints every five seconds while debug is on. That line is debug output.

## Capabilities

### New Capabilities
- `dev-tools`: the edit mode and what it does to play, explicit saving, and what a release build contains, omits and loads.

### Modified Capabilities
- `multiplayer/session`: *Offline-only features are disabled online*. Its statement that debug overlays remain available online now applies to development builds only, since a release build has none. Its rewind scenario is updated from holding R to pressing it, matching time manipulation as it now works.

## Impact

- **Code**
  - `game.go`: the edit-mode gate in `runLevel`, beside the time-manipulation one; save-on-quit removed; the overlay, stats and editor GUI behind the build tag.
  - New `devtools_dev.go` / `devtools_release.go`: a constant `DevTools`, true without the `release` tag and false with it, which every developer-only path checks.
  - `level_editor.go`: edit-mode entry and exit, Ctrl+S, the unsaved marker, and `LoadLevel` going straight to the embedded level in a release build.
  - `net_client.go`: F5–F8 behind the build tag. `net_server.go`: its debug logging behind it too.
  - `build.ps1`: builds with `-tags release`.
- **Tests**: `go test ./...` keeps testing the development build. A release-tag build and test run is added, so a tag mistake is caught at build time and not in a player's hands.
- **Not affected**: gameplay rules, structure, time manipulation, the network protocol and the level file format.

## Non-goals

- New editor features: undo, multi-select, editor UI beyond the existing object picker.
- Moving the editor out of the game binary into a separate tool.
- Any diagnostics for the release build to replace the overlay.
