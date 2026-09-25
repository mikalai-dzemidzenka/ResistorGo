## Why

A level has nowhere a fighter can die except at another fighter's hands. The new pit under level 1's floor, and anything a fighter can fall out of, needs a way to say "falling in here is the end": a rectangle of the level that kills whoever touches it.

## What Changes

- A level SHALL be able to hold death areas: freestanding rectangles, not tied to any tile, saved in the level file under a key of their own and only when the level has some, so a level without them keeps its file and its fingerprint.
- A living fighter whose collision box overlaps a death area SHALL die, through the same Death state a broken structure leads to — fall, flicker, respawn — with nobody credited for the kill. Only the authority decides it; a predicting client shows the death it is told about.
- The editor SHALL get a fifth tool, the death tool, reached by the same key that cycles the others. It places, moves, resizes and removes death areas exactly as the area tool does lit areas, with a panel for their numbers, markers drawn while it is out, and edits saved and undone like any other.
- Death areas SHALL have no part in lighting, and SHALL NOT be drawn while the game is played outside debug mode.
- Out of scope: bots do not see death areas and may walk into one; teaching pathing to avoid them is a follow-up.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `structure`: a new way to enter the Death state — touching a death area — alongside a broken structure.
- `level-editor`: the death tool, its panel and markers, saving and undo of death areas, and the level file keeping its shape when a level has none.

## Impact

- `level_lights.go`: a `DeathArea` type, the level file's death-area key in `parseLevel` and in saving.
- `level_editor.go`: `toolDeath`, the editor's death-area list and selection, the tool's update, Delete key and panel, loading.
- `level_undo.go`: the death areas in the undo snapshot.
- A new file for the death tool's editing and drawing, alongside the area tool's.
- `world.go` / `actor.go`: the authority's overlap check, calling `enterDeathState(nil)`.
- `level_1.json`: unchanged until death areas are placed in it.
- Level fingerprint changes for any level that gains a death area, so client and server must run the same level file, as they already must.
