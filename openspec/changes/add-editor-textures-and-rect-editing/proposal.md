## Why

The level editor can place three things. `LoadLevelObjects()` hands it a hardcoded list of three block prototypes, each with a collision rectangle and a light rectangle typed into Go source, and the picker shows exactly those. `resources/blocks/` already holds seven PNGs — all of them embedded in the binary — and four of them can never be placed in a level because no Go literal mentions them. Drawing new level art is not enough to build with it: it takes an edit to `resources.go`, a guess at the right collision and light rectangles, a rebuild, and a look at the result. Get the numbers wrong and the fix is another edit, another rebuild, another look.

The rectangles are the worse half. Every tile's collision box and its light occluder are the things that decide whether a level plays and whether it is lit, and today they can only be authored by typing numbers into a struct literal for a texture you are looking at on the other side of a compile. Edit mode already pauses the match, draws both rectangles over every placed object, and knows where the mouse is — everything needed to shape them by hand is on screen already.

## What Changes

- **The picker shows every block texture, not a list in Go.** The editor builds its palette by enumerating `resources/blocks/*.png` — from the embedded asset FS, and additionally from disk in a development build — so a PNG dropped in that folder is placeable on the next run with no code change. The three prototypes in `LoadLevelObjects()` stop being the source of truth for what exists and become the starting rectangles for the textures that have them.
- **A texture can be brought in by dropping it on the window.** In edit mode, a development build accepts a PNG dragged onto the window: it is copied into `resources/blocks/`, loaded into the texture registry, added to the palette and selected, ready to place immediately. Dropping a file outside edit mode, or in a release build, does nothing.
- **A dropped texture is only in this build until the binary is rebuilt.** The editor says so: a level object whose texture is not in the embedded FS is marked in the editor as living on disk only, because a release build of today's binary would not have it.
- **Collision and light rectangles are edited on screen.** A new rect tool in edit mode shapes the selected object's rectangles by dragging their edges, corners and body in the world, with numeric X/Y/W/H fields beside them for exact pixel values. Both rectangles are editable; which one is being shaped is explicit, and the other stays visible.
- **Rectangles can be edited per texture and per placed tile.** Shaping the picked palette entry sets the defaults that new placements of that texture are born with; shaping a tile already in the level changes only that tile. An object with no rectangle at all can be given one, and a rectangle can be cleared — a texture that should not collide or should not cast a shadow needs to be able to say so.
- **Palette defaults persist.** They are written to a new `level_objects.json` beside `level_1.json`, on the same terms the level itself is written on: only in a development build, only when Ctrl+S is pressed in edit mode, never on quit. A missing or malformed file falls back to the rectangles in `LoadLevelObjects()`, so the editor always opens on something usable.
- **A missing texture draws a placeholder instead of nothing.** A level object whose texture is not in the registry currently draws a zero texture — an invisible tile that still collides. It becomes a visible marker in a development build, so a level that outran its assets is obvious rather than haunted.

## Capabilities

### New Capabilities

- `level-editor`: what the editor can author — the texture palette and where it comes from, importing a texture by dropping it, selecting a palette entry or a placed object, shaping and clearing collision and light rectangles, and where palette defaults live.

### Modified Capabilities

- `dev-tools`: *The editor takes input only in edit mode* — edit mode now has two tools, so a click means place-or-delete or rect-shaping but never both, and a dropped file is input the editor must ignore outside edit mode. *Saving the level is explicit* — Ctrl+S now also writes the palette defaults, and unsaved rectangle edits count as unsaved edits. *A release build carries no developer tools* — a release build additionally has no texture import and never writes `level_objects.json`.

## Impact

- `level_editor.go` takes the palette source, the tool state, rect hit-testing and dragging, and the `level_objects.json` read/write. `resources.go` loses its role as the list of what is placeable and keeps its role as the default rectangles, and gains loading a texture from disk at runtime — `loadTex` currently goes through `asset()`, which panics on anything not embedded.
- `game.go` polls for dropped files and already calls `LE.Update` and `LE.DrawGui` in the right places; the rect tool's world-space drawing goes with the existing debug rectangles in `LE.Draw`.
- `assets.go` needs no new embed pattern: `resources/blocks/*.png` already covers anything dropped into that folder, which is why that folder is the destination.
- Rect edits change the level fingerprint exactly as placing a tile does — the rectangles are already stored per object in `level_1.json` — so client and server agree as long as they run the same build, which is what the fingerprint is for. `level_objects.json` is an editor file and is not part of the fingerprint.
- The tile grid stays 39x83. A texture of another size can be placed, but it snaps to that grid; per-texture grid sizes are not part of this change. Neither is undo, nor pushing a changed palette default onto tiles already placed.
- Testable without a GL context: palette enumeration from a fake FS, rect handle hit-testing and drag maths, the `level_objects.json` round trip, and the dropped-file path check. The visual result — handles under the cursor at every zoom level, the lights reacting to a reshaped occluder — is verified by running the game.
