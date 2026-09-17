## Why

A tile family is drawn to one shape and then has to be shaped again once per texture. The eight `mn_*` textures now in `resources/blocks/` are all 76x105 and want the same collision box and the same light occluder. The in-flight editor change (`add-editor-textures-and-rect-editing`) gives every texture its own default rectangles, so today that is eight rounds of dragging the same two rectangles to the same pixels.

The second half is worse. A tile copies its texture's rectangles when it is placed and keeps them from then on. Correcting a texture's collision box after it has been laid across a level corrects nothing already there: every one of those tiles has to be deleted and placed again, or reshaped one by one.

## What Changes

- **Rectangles can be copied and pasted.** Ctrl+C in edit mode copies the selected thing's collision and light rectangles — including the absence of one — and Ctrl+V replaces the selected thing's rectangles with that copy. The selected thing is a palette entry or a placed tile, as the editor already defines it. A paste is a copy, not a link: changing the source afterwards changes nothing it was pasted onto. Copy and paste work with the object picker open, so shaping one texture and handing its rectangles to seven more is copy once, then pick-and-paste seven times.
- **Placed tiles follow their texture.** A placed tile uses its texture's rectangles. Changing a texture's rectangles — by dragging, typing, adding, clearing or pasting — changes every placed tile of that texture along with it.
- **A tile edited by hand keeps its own rectangles.** Reshaping, adding, clearing or pasting onto one placed tile detaches it: from then on its texture's changes pass it by. Nothing else changes about editing one tile.
- **A detached tile can follow its texture again.** F on a selected tile gives it its texture's rectangles and makes it follow again. The editor's panel says which of the two the selected tile is.
- **Loading a level never changes it.** On load, a tile follows its texture when its rectangles match the texture's, and keeps its own when they differ. Nothing about the level file changes shape, and nothing extra is saved.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `level-editor`: *Palette defaults and placed objects are edited separately* — a texture's change now reaches the placed tiles that follow it, instead of only new placements. New requirements add copy and paste, following a texture again, and loading without changing the level. This capability is introduced by the in-flight `add-editor-textures-and-rect-editing` change, which is not yet archived; this change has to be archived after it.

## Impact

- `level_editor.go` only, on top of the in-flight editor change, which is implemented but not yet committed. Whether a tile follows lives on the tile in memory, never in the level file: `level_1.json` keeps exactly the shape and content it has, the level fingerprint is unaffected, and so are collision, lighting, the release build and multiplayer. They only ever see the rectangles that result.
- `level_objects.json` is unchanged in shape; a texture's rectangles are still its palette default.
- A tile edited by hand to exactly its texture's rectangles counts as following on the next load, since the level file does not record the difference. It looks the same either way; it only means a later change to the texture will reach it.
- Not in scope: named rectangle presets or textures linked to each other, the operating system clipboard, pasting onto several things at once, and undo.
