## Context

See `proposal.md` — Why. What shapes the approach is where the editor already sits in the frame:

- `LE` is a package-level `LevelEditor` built by `NewLevelEditor(39, 83)` **at package init**, before a window or a GL context exists, and built a second time in `Game()` after `LoadTextures()`. So building the palette must not touch the GPU: a palette entry can only be paths and rectangles, with the texture looked up in the `resources` map when something is drawn.
- `resources` is `map[string]rl.Texture2D` keyed by asset path, filled by `LoadTextures()` through `loadTex`, which goes through `asset()` — a panic on anything not in `assetFS`. Nothing can load a texture from disk today.
- `assetFS` already embeds `resources/blocks/*.png`, so that folder is the one place a new texture can land and be in the next build without touching `assets.go`.
- `LE.Update` runs per frame from `runLevel` only while `Editing`; `LE.Draw` runs in world space inside the camera, with the existing collision/light rectangle overlay behind `DevTools && Debug`; `LE.DrawGui` runs in screen space after the palette pass, inside the 1280x720 virtual screen.
- `present.letterbox()` already calls `rl.SetMouseOffset`/`SetMouseScale`, so `rl.GetMousePosition()` is in virtual-screen coordinates at any window size: raygui and `rl.GetScreenToWorld2D(MousePos, Cam.View())` both agree with what is drawn.
- `Cam.View()` zoom is a whole number 1..5 (`PixelSize`), so one world pixel is 1 to 5 virtual-screen pixels. Any grab tolerance has to be derived from the zoom, not fixed in world pixels.
- `MouseWorldPos` is a `Vector2i` produced by `v2i`, which truncates — fine for snapping to a grid cell, wrong for dragging a rectangle edge, where truncation biases every drag on the negative side of the origin.
- Rectangles in `LevelObject` are `*rl.Rectangle` stored **relative to the object's origin**, and nil already means "no rectangle": `actor.go` skips a nil collision rect, `LightOccluders` skips a nil light rect. Add/clear is therefore nil-handling, not a new flag.
- `go test ./...` builds without the release tag, so `DevTools` is true in tests and every path here is reachable — but nothing with a GL context is.

## Goals / Non-Goals

**Goals:**

- Keep the GPU out of palette construction, so `LE` can still be a package-level var built twice.
- Put the arithmetic that can be tested without a window — palette merge order, handle hit-testing, drag clamping, the defaults file round trip, the import name check — in plain functions over plain values.
- Give the mouse exactly one owner per click: raygui, or the world.
- Keep every disk read and write behind the `DevTools` constant so the release build drops them at compile time.

**Non-Goals:**

- Undo/redo, multi-select, copy/paste, rectangle snapping to the grid.
- Per-texture grid sizes: the grid stays 39x83.
- Reapplying a changed palette default to tiles already placed.
- Any change to how a level is stored, hashed or replicated.
- Importing anything but a PNG (sprite sheets, models, audio).

## Decisions

### The palette is a merged, sorted path list, built without GL

`LoadLevelObjects()` stops being the list of what exists. The palette is built by `blockTextures()`: `fs.Glob(assetFS, "resources/blocks/*.png")`, plus — only when `DevTools` — the `*.png` names from `os.ReadDir("resources/blocks")`, merged, deduplicated and sorted by path. Sorting gives the stable picker order the spec asks for for free, and the merge function is a pure `[]string, []string -> []string` that tests can drive without a filesystem.

Each entry is a `LevelObject` carrying the path and the rectangles resolved for it (see below). `NewLevelEditor` must tolerate an empty list rather than indexing `[0]` — an empty `resources/blocks` is then a level with nothing to place, not a panic.

*Alternative considered:* a manifest file listing placeable textures. Rejected: it is the hardcoded list again, one indirection further away, and it would have to be maintained alongside a folder that already says exactly the same thing.

`LoadTextures()` loads every path `blockTextures()` returns instead of three names, so anything placeable is drawable. Seven small PNGs today; eager loading keeps the draw path free of "is this loaded yet" state.

*Alternative considered:* lazy load on first use. Rejected as unnecessary state for a handful of tiles; revisit if a level ever carries hundreds of textures.

### Textures can come from disk, and a missing one is visible

`loadTex(name)` splits into `loadTexBytes(name string, data []byte)` and two callers: `loadTex` (embedded, via `asset()`, unchanged behaviour) and `loadTexDisk(name, diskPath)` for a texture that is not in the build. `embeddedBlocks` — a set built once from the `assetFS` glob — is what tells the editor whether an entry is "in the build", and drives both the picker marking and the outline on placed objects.

`LE.Draw` looks up with `tex, ok := resources[obj.Resource]`. When `!ok`, a development build draws a marker rectangle with the object's path; a release build draws nothing, as it does today. Collision is untouched either way — collision reads the rectangle, never the texture.

### Prototype defaults live in `level_objects.json`, not in the level and not in the binary

A new file beside `level_1.json`:

```json
{ "resources/blocks/e1m1_wall1.png": { "collisionRect": {...}, "lightRect": {...} } }
```

Keyed by texture path, values holding the same two `*rl.Rectangle` fields a `LevelObject` has, nil meaning none. Resolution order for a palette entry: the defaults file, then the built-in rectangles still returned by `LoadLevelObjects()`, then none. The file is read only when `DevTools`, and is **not** embedded: the release build has no editor, so nothing there reads it, and keeping it out of `assetFS` keeps it out of any question about the level fingerprint.

*Alternative considered:* folding defaults into `level_1.json`. Rejected: the level's fingerprint is a network contract, and editor preferences have no business changing it.

### Two tools, and the keys that switch them

`LevelEditor` gains a tool (`toolPlace`, `toolRect`) and a selection. In `toolPlace` the mouse behaves exactly as it does today. In `toolRect` the left mouse button only selects and drags; nothing places, nothing deletes.

Keys, all free today (movement and `R` are inert while editing, `Ctrl+S` and `Tab` keep their meanings):

- `Q` — switch tool
- `C` / `L` — make the collision / light rectangle the active one
- `N` — give the selection the active rectangle when it has none, as one grid cell (39x83) at the object's origin
- `Delete` — clear the active rectangle

While a raygui value box has keyboard focus these letters are ignored, or typing a number would also retool the editor.

### Selection is one of two things, and the prototype gets a parked ghost

`selection` is either the palette entry (`le.obj`) or one placed object, addressed by layer and grid position so it survives the index rebuild. Picking in the picker selects the entry; a left click in `toolRect` that hits no handle selects the topmost placed object whose cell is under the cursor, and selecting nothing leaves the previous selection alone rather than silently editing the entry.

The prototype has a problem the instances do not: its ghost follows the mouse, so its rectangles would run away from the cursor mid-drag. So in `toolRect` the ghost is **parked**: `updateCursor` stops following the mouse, and the prototype is drawn at the cell it was on when the tool was entered. Leaving `toolRect` lets it follow again.

*Alternative considered:* editing prototype rectangles only in a separate screen-space widget, with no world dragging. Rejected — dragging against the texture you are shaping is the whole point of the change.

### Dragging: unrounded mouse, zoom-derived tolerance, clamped result

Dragging uses `rl.GetScreenToWorld2D(MousePos, Cam.View())` directly and rounds, rather than the truncating `MouseWorldPos`.

Hit-testing picks, in order: the four corners, then the four edges, then the body. Tolerance is `max(1, 4/zoom)` world pixels, so a handle stays about four virtual-screen pixels wide at every zoom level.

A drag stores the grabbed handle and the offset between the cursor and the rectangle at grab time, so the rectangle does not jump on the first frame. Applying a drag is a pure function `dragRect(r rl.Rectangle, h handle, to rl.Vector2) rl.Rectangle`: a move translates, an edge or corner moves only that side and clamps against the opposite one so width and height never go below 1 and can never invert. The stored value stays relative to the object origin — the world position is `origin + rect`, so the drag converts once on the way in and once on the way out.

The numeric fields are `rlg.ValueBox` over `int32` copies of the active rectangle's X/Y/W/H, written back when a box reports a change, with the same minimum-1 clamp. `rl.Rectangle` is float32 but every value here is a whole pixel, which is what the level has always stored.

### One owner per click

`LE.DrawGui` draws the rect panel in the bottom-left of the virtual screen, clear of the picker strip, the debug text and the EDIT marker. The editor records the panel's bounds (and the picker's, which today is handled by the `isObjectsViewerEnabled` check) and `LE.Update` ignores world mouse input while the cursor is inside them. Panel bounds are computed in one place and used for both drawing and the guard, so they cannot drift apart.

### Dropped files are always drained, acted on only in edit mode

The drop check lives in the main loop. `rl.IsFileDropped()` stays true until `rl.UnloadDroppedFiles()`, so the loop **always** loads and unloads the drop list and only *acts* on it when `DevTools && Editing` — otherwise a file dropped during a match would be waiting to be imported the moment edit mode was entered, which is exactly what the spec forbids.

Import order matters for what is left on disk when something goes wrong: check the extension is `.png`, load the dropped file from its original path to prove it decodes, and only then copy it into `resources/blocks/` and register it under that path. A file that does not decode never gets copied, so a bad drop cannot poison the next run's palette. A name already present in `resources/blocks` is never overwritten; the existing entry is selected instead. Then the palette is rebuilt and the new entry selected.

### One dirty flag, one save

Placements, deletions, instance rectangle edits and prototype default edits all set the existing `dirty` flag, and `Ctrl+S` writes `level_1.json` and then `level_objects.json`, clearing it. Two flags and two markers would buy nothing: the answer to "is there anything unsaved" is the only thing the marker shows.

## Risks / Trade-offs

- **A level can reference a texture the shipped binary will not have.** → The editor marks disk-only entries and placed objects, the development build draws a marker where a missing texture would be invisible, and `build.ps1` builds from the repo where the file already sits in `resources/blocks`, so the normal path embeds it. Not prevented, only made loud.
- **Release build accidentally gaining a disk read or write.** → Every disk path is behind the `DevTools` constant so it is compiled out, and the dev-tools delta spec carries scenarios for the dropped file, the extra PNG on disk and the defaults file.
- **raygui keyboard focus fighting the editor's letter keys.** → Letter keys are ignored while a value box is being edited; this is the one piece of interaction that cannot be covered by a test without a window, so it is on the manual check list.
- **Two mouse owners, one button.** → Single source of truth for GUI bounds, and world input suppressed inside them. A click that lands on a handle never also selects a different object, because hit-testing runs handles first.
- **The parked prototype ghost is a mode surprise**: in the rect tool the ghost stops following the cursor. → It is drawn with its selection marking, and the tool is shown in the panel. The alternative — a ghost that runs away while its rectangle is dragged — is worse.
- **A defaults file written by hand can disagree with the built-in rectangles.** → The file wins where it has an entry and the built-ins fill the rest; a malformed file falls back entirely and logs, on the same terms `LoadSettings` already uses for `settings.json`.
- **`level_objects.json` is another file to keep in the repo.** → Accepted: it is authoring data that belongs beside `level_1.json`, and it is written only on an explicit save.

## Migration Plan

No data migration. `level_1.json` is unchanged in shape and loads as it does today. `level_objects.json` is absent until the first save, and absent means "use the built-in rectangles". Rollback is reverting the commit: the level file written after this change is still readable by the code before it, because nothing about its format changed.
