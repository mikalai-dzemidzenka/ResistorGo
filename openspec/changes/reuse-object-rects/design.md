## Context

See `proposal.md` — Why. This builds directly on the in-flight `add-editor-textures-and-rect-editing` code (implemented, uncommitted), and the relevant parts of it are:

- `LevelEditor.levelObjects` is the palette; each entry's rectangles are resolved from `level_objects.json`, then the built-in table, then none. `setSelectedRects` is the single place both kinds of selection are written: the proto branch updates `le.obj`, the palette entry and `le.defaults`; the tile branch rewrites one `LevelObject` in `le.objects[layer][pos]` and calls `buildIndex`.
- Every rectangle edit — drag, value box, `N`, `Delete` — already funnels through `setActive` → `setSelectedRects`, so there is one choke point for "a texture changed" and one for "a tile changed".
- `place` clones the entry's rectangles into the new tile. `LevelObject` is a value type stored in maps and copied into `objectsIndexed`.
- The level fingerprint is `levelHash(marshal())`, where `marshal` is `encoding/json` over `le.objects`. A release build loads the embedded level with no editor state; `TestLoadLevelSource` pins that dev and release agree.
- Keys are read in `LevelEditor.Update` with `rl.IsKeyPressed`, before the early return taken while the picker is open. `C` already switches the active rectangle, and Ctrl+S is checked first. Letter keys are skipped while a value box has the keyboard (`editField >= 0`).

## Goals / Non-Goals

**Goals:**

- Keep follow-or-own out of the level file, so the fingerprint, the file format and every consumer of the level are untouched.
- Make loading pure: it may classify tiles, never alter them.
- Route every change through the existing `setSelectedRects` choke point, so no edit path can forget to propagate or to detach.

**Non-Goals:**

- Named presets, or textures linked to one another.
- The operating system clipboard.
- Pasting onto several selections at once, and undo.
- Recording "own" when it equals the texture — see Risks.

## Decisions

### Follow-or-own is an unexported field on the tile, inferred at load

`LevelObject` gains `own bool`. It is unexported, so `encoding/json` never writes it: `level_1.json`, `marshal` and therefore `Hash` are exactly what they were. Being on the value itself, it travels with the tile through the `objects` maps and into `objectsIndexed` with no side table to keep in step with placement, deletion and the index rebuild.

After `LoadLevel` unmarshals, `markOwnRects` classifies every tile: `own = true` when its texture has no palette entry or when `!sameRects(tile, entry)`; otherwise it follows. `sameRects` compares both rectangles, nil-aware — nil equals only nil, and two values compare by value. Classification writes the flag and nothing else.

*Alternative considered:* a persisted `"ownRects"` field in `level_1.json`. Rejected: inferring it at load would make a dev build's marshal differ from a release build's for the same file, breaking the fingerprint across builds; persisting it only on save leaves the same problem for any level saved from a dev build and played by an older client. *Alternative considered:* a side map in `level_objects.json` keyed by layer and position. Rejected: positions are what editing changes, so that table rots silently the first time a tile is moved by delete-and-place.

### Propagation and detaching happen in `setSelectedRects`

- **Proto branch:** after updating the entry and `le.defaults`, walk every layer and give each tile with the same `Resource` and `!own` a clone of the new rectangles, then `buildIndex` once. Clones, never the entry's pointers: tiles must stay independent of the entry for the next edit.
- **Tile branch:** set `own = true` on the tile being written. This covers drag, value box, `N`, `Delete` and paste with no change at any of those call sites.
- `place` leaves `own` false — a new tile follows.

A drag on a texture now rewrites every following tile each frame. With levels of tens to a few hundred tiles that is a map walk and one sort per frame while dragging, in a paused match. Not worth a cache.

### Following again

`followTexture()`: only for a tile selection with `own == true` whose texture has a palette entry. It writes the entry's rectangles (cloned) and `own = false` directly — not through the tile branch of `setSelectedRects`, which would immediately set `own` back to true. It marks the level dirty only if the rectangles actually changed, the same rule `setActive` applies. Bound to F, which is free in edit mode.

### Copy and paste are a value on the editor

`clip ObjectRects`, `clipFrom string` (the source's name for the panel), `hasClip bool`. Copy stores `rects().clone()` of the selection. Paste calls `setSelectedRects(clip.clone())` — so a paste onto an entry propagates and a paste onto a tile detaches, by the same code as any other edit — guarded by `sameRects` so a paste that changes nothing is not an unsaved edit. The clone on both sides is what makes it a copy and not a link.

### Keys: Ctrl chords first, bare letters only without Ctrl

`Update` reads `ctrl := IsKeyDown(LeftControl) || IsKeyDown(RightControl)` once. With Ctrl held: S saves (unchanged), C copies, V pastes. Without Ctrl: the existing letter keys, plus F. This ordering is what stops Ctrl+C from also making the collision rectangle active — `IsKeyPressed(KeyC)` is true for both. Copy, paste and F are handled before the picker's early return, so they work with the picker open, and all of them are skipped while a value box has the keyboard, as the other editor keys already are.

### The panel says which

The selection line in the panel gains, for a tile, "follows texture" or "own rectangles (F: follow texture)", and a clipboard line reads "Copied: `<name>`" once something has been copied. No new world-space drawing.

## Risks / Trade-offs

- **An own tile edited back to exactly its texture's rectangles counts as following after a reload.** → Accepted: it is indistinguishable in play, and the only effect is that a later texture change reaches it. The alternative is a persisted flag, rejected above for the fingerprint.
- **Texture defaults and a level can disagree** — a merged or hand-edited `level_objects.json`, or a level saved before this change. → Classification makes every disagreeing tile own, so nothing moves on load; the disagreement is only resolved by an explicit edit or F.
- **Propagation touches tiles the developer is not looking at.** → That is the feature, and the propagation is limited to the one texture and to following tiles; the unsaved marker already goes up, and nothing is written until Ctrl+S.
- **Ctrl+C / Ctrl+V inside a value box.** → Editor keys are already ignored while a field has the keyboard, so raygui keeps whatever it does with those keys there.
- **This change depends on uncommitted work.** → Its tasks assume the in-flight change's code as it stands; it must be archived after that change, since it modifies a requirement that change introduces.

## Migration Plan

None. The level file and `level_objects.json` keep their shape and content; existing levels load unchanged and are classified tile by tile. Rolling back leaves every tile's rectangles as they were last saved, since those were always stored resolved.
