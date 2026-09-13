## Context

See proposal.md — Why. What shapes the approach is how the developer kit is wired today:

- `Debug` is a package-level `var` in `game.go`, `true` at start. The main loop flips it on `` ` `` for every stage, in every build.
- Everything developer-only hangs off that one variable:

  ```
  game.go:312        offline && Debug      --> LE.Update(dt)   editor input, every frame
  game.go:315        Debug && Cli != nil   --> Cli.Sim.controls()   F5-F8 latency/loss
  game.go:354        Debug && Srv != nil   --> lag-comp rewind boxes
  game.go:412        Debug && offline      --> LE.DrawGui()   object picker
  game.go:421        Debug                 --> fps/time text, drawNetDebug()
  actor.go:393       Debug                 --> collision/attack boxes, state text
  level_editor.go:165 Debug                --> tile collision/light rects
  net_server.go:318  Debug                 --> per-player view-tick log line
  net_server.go:328  Debug                 --> records rewind boxes for the overlay
  game.go:154        LE.dirty on quit      --> SaveLevel("level_1.json")
  ```

- `LE.Update` reads Tab (object picker), RMB (delete) and LMB (place) itself, and gameplay reads LMB in `latchInput`. Nothing arbitrates between them, so one click does both.
- `LoadLevel` reads `level_1.json` from disk first and falls back to the embedded copy. The game (`game.go:130`, `:258`) and the dedicated server (`net_server.go:395`) both load through it, and `LE.Hash()` of whatever was loaded is the handshake fingerprint.
- Pausing already has a precedent. Time manipulation stops the match by skipping `stepOffline` and not adding to the accumulator. But `latchInput()` runs every frame before that gate, so a click made while paused stays latched in `pressedAttack` and is consumed by the first tick after the pause.
- `build.ps1` is the only release path: `go build -trimpath -ldflags "-s -w" -o resistor.exe ./cmd`, then UPX.

## Goals / Non-Goals

**Goals:**

- The editor listens only when asked to, in every build.
- A release binary that does not contain the developer kit at all, rather than one that hides it behind a flag.
- One switch for "is this a development build", so no gate can be forgotten in a second place.

**Non-Goals:**

- Editor features, or moving the editor into a separate tool (proposal — Non-goals).
- Changing what the overlay draws in a development build.

## Decisions

### A build-tag constant, `DevTools`

Two small files define one constant: `devtools_dev.go` (`//go:build !release`) sets `const DevTools = true`, and `devtools_release.go` (`//go:build release`) sets it `false`. `build.ps1` passes `-tags release`; `go run`, `go build` and `go test` without the tag get a development build.

Every developer-only path is guarded by `DevTools`. That includes the `` ` `` toggle, F2, F5–F8, the overlay draws, the editor, and level writes. Because it is a constant, the compiler drops those branches from a release binary instead of carrying them behind a check. `Debug` stays a `var` for the overlay toggle, initialised from `DevTools` and toggled only when `DevTools` is true, so it is simply never true in a release build. That also silences the two server-side `Debug` paths without touching them.

*Alternative — `-ldflags "-X"`:* rejected. It only sets string variables, so it is a runtime check with all the developer code still compiled in.

*Alternative — a command-line flag:* rejected. The decision is that nobody opens the editor in a shipped build, and a flag is exactly the door that decision closes.

### Edit mode as its own pause, next to time manipulation

An `Editing` flag in `level_editor.go`, toggled by F2 in `runLevel` only when `DevTools`, the match is offline and `TimeMode == TimeOff`. While R's handler sees `Editing`, it does nothing. The simulation gate becomes `TimeMode == TimeOff && !Editing`, reusing the precedent: no `stepOffline`, no accumulation.

Unlike time manipulation today, the edit-mode gate also stops `latchInput()` from latching anything, and clears `pressedAttack` and `pressedJump` when edit mode is left. Otherwise the editing clicks would fire an attack on the first tick back, breaking "Clicks do not carry over".

### The editor listens only inside edit mode

`LE.Update` and `LE.DrawGui` are called only while `Editing`, instead of whenever `Debug` is on. That one gate is the whole fix for the shared mouse: outside edit mode no editor code reads a button. Tab keeps its meaning, inside edit mode only.

### Saving on Ctrl+S, never on quit

`LE.Update` handles Ctrl+S — left or right Control held, S pressed — by calling `SaveLevel` and clearing `dirty`. The `if LE.dirty { LE.SaveLevel(...) }` on the way out of `Game()` goes. `dirty` stays as the "unsaved" signal for the marker.

S is not a gameplay key (movement is A/D/W), and it is only read inside edit mode anyway.

### The marker in screen space

"EDIT", or "EDIT — unsaved" while `dirty`, drawn after the palette pass like the spectator HUD, so it is legible and unaffected by the scene's lighting or palette. It is drawn only while `Editing`.

### The release build ignores the disk level

`LoadLevel` reads the embedded copy directly when `!DevTools`. A release exe therefore plays the level it was built with, whatever sits beside it, and every copy of the same release computes the same fingerprint. Development builds keep disk-first loading, which is what makes a saved edit take effect on the next run.

This also repairs players who already have an edited `level_1.json` from the current build: a release build from this change simply ignores it.

## Risks / Trade-offs

- **An unsaved edit is lost on quit.** That is the point of explicit saving, and the marker says "unsaved" the whole time. → Accepted; it is the development workflow's only cost.
- **A release built without `build.ps1` is a development build.** `go build ./cmd` without the tag produces an exe with the full kit. → `build.ps1` is already the only documented release path. The release-tag test run in the tasks keeps that build compiling and passing.
- **The release binary can no longer show anything diagnostic.** No fps, no net stats. → Accepted in the proposal's Non-goals; a developer reproduces reports in a development build.
- **Tests exercise the development build.** A gate that is wrong only under the tag would not show up in `go test ./...`. → A `-tags release` vet and test run is a task, and the release-only behaviour that matters (no level writes, embedded level only) is tested under that tag.

## Migration Plan

Nothing to migrate. `level_1.json` in the repository stays the source of the embedded level. Stray copies beside shipped executables become irrelevant to release builds from this change. Reverting is a single revert; the only behaviour a revert brings back is the problem this change fixes.

## Open Questions

- **The same latched-click problem appears to affect time manipulation.** Its gate also sits after `latchInput()`, so an attack clicked while time is paused may fire on resume. That would contradict the time-manipulation spec's "no input is buffered for when the mode is left". This is unverified: it was found by reading the code while designing edit mode, not reproduced. Fixing it means applying the same latch rule to time manipulation, which is outside this change unless its scope is widened.
