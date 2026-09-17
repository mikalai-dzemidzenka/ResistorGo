## Context

See `proposal.md` - Why. What the approach has to work with, as committed in `e2d3d98`:

- **The data.** `LevelObject.LightRect *rl.Rectangle` is one optional rectangle, relative to the object's origin, beside `CollisionRects []rl.Rectangle`. The in-memory pair `ObjectRects{Collision []rl.Rectangle, Light *rl.Rectangle}` (`level_objects.go`) carries a tile's or a texture's rectangles through `rects()`/`setRects()`, `clone()`, `sameRects()`, `builtinRects()`/`buildPalette()`, copy and paste, and `level_objects.json`. `markOwnRects` and `followTexture` compare and copy through the same pair.
- **The file and the fingerprint.** `levelObjectWire` and `objectRectsWire` write `resource, layer, x, y, collisionRect, collisionRects, lightRect`. `collisionToWire`/`collisionFromWire` write one collision rectangle as `"collisionRect"` and two or more as `"collisionRects"`, and on load keep the list if a file has both. `marshal()` of the tiles is the fingerprint's input. `level_1.json` has 88 tiles, 68 of them with one light rectangle each.
- **The shadows.** `LightOccluders()` turns each object's light rectangle into a world rectangle. `computeLights` hands that flat list to `sharedLight.Compute`, which draws every occluder with no limit on how many. Nothing else reads a light rectangle outside the editor.
- **The rect tool.** `activeRect` is `rectCollision` or `rectLight`, and `activeColl` is which collision rectangle. `ObjectRects.get`/`set` take `(which, i)`, and ignore `i` for light. `rectOrder` tries the active rectangle, then every collision rectangle, then the light one. `C` (`pickCollision`) switches to collision or steps through it. `L` only switches to light. `N` (`addActiveRect`) appends a collision rectangle, but adds a light rectangle only when there is none. Delete (`clearActiveRect`) steps the collision index back. The panel shows "Collision k of n" only for collision. `clampActiveColl` clamps against the collision list only.
- **Drawing.** `drawSelection` walks `rectOrder`. The debug overlay in `Draw` outlines every collision rectangle and the one light rectangle.

## Goals / Non-Goals

**Goals:**

- Light rectangles work exactly as collision rectangles already do: an ordered list in memory, one or many on disk, every one editable, every one carried.
- A level or defaults file with at most one light rectangle per object keeps its bytes and its fingerprint.
- The rect tool treats the two kinds the same way, with one code path rather than a light-shaped special case beside a collision-shaped one.

**Non-Goals:**

- Renaming the requirement *Every collision rectangle travels together*. It now covers both kinds, but a rename cannot be archived against a capability that is not in `openspec/specs` yet. It can be renamed once the capability is archived.
- Changing the shadow passes, how an occluder is drawn, or the order occluders are drawn in.

## Decisions

### A list for light, exactly as for collision

`LevelObject.LightRect *rl.Rectangle` becomes `LightRects []rl.Rectangle`, and `ObjectRects.Light` becomes `[]rl.Rectangle`. Values, not pointers: nil or empty means no shadow. `clone()` clones both slices, and `sameRects()` compares both in order. `sameRect`/`cloneRect` go if nothing else uses them.

The wire structs gain a `lightRects` field after `lightRect`, so the field order is:

```
resource, layer, x, y, collisionRect, collisionRects, lightRect, lightRects
```

`collisionToWire`/`collisionFromWire` become `rectsToWire`/`rectsFromWire`, used for both kinds. The load warning names the key it is about. One light rectangle is written as `"lightRect"` exactly as now, two or more as `"lightRects"`, none as neither. Because a one-rectangle object marshals to the same bytes as before, `marshal()` and so the fingerprint are unchanged for every existing file. `TestLevel1KeepsItsBytes` and its legacy struct, which still has a single `LightRect`, pin this. The legacy comparison skips a level that has a tile with more than one collision rectangle. It gains the same skip for more than one light rectangle.

*Alternative considered:* always writing `"lightRects"`. Rejected for the reason `multiple-collision-rects` rejected it: 68 tiles of `level_1.json` would be rewritten, and the fingerprint would change for a level nobody edited.

### One index for whichever kind is active

`activeColl` becomes `activeIdx`: the index into the active kind's list. `ObjectRects.get(which, i)` and `set(which, i, v)` index the light list the same way as the collision list, through one helper over a `*[]rl.Rectangle` for the kind. `set` still copies the list before writing, because a palette entry and its tiles can share it.

- **`C` and `L`.** Both go through one `pickKind(which)`. If the other kind is active, it switches and makes index 0 active. If this kind is already active, it steps `activeIdx` to the next one, wrapping. `C` behaves exactly as today.
- **`N`.** Appends a one-cell rectangle at the origin to the active kind's list and makes it active, for light as for collision. This drops the "only when there is none" rule for light.
- **Delete.** Removes the active rectangle and moves `activeIdx` to the previous one, or 0, for both kinds.
- **Clamping.** `clampActiveColl` becomes `clampActiveIdx` and clamps against the active kind's list. It is called at the same places as today, so a new selection or a switch of kind never points past its list.
- **Picking.** `rectOrder` tries the active rectangle, then every collision rectangle, then every light rectangle, each in order, skipping the active one. A click still goes to the active rectangle where rectangles overlap.
- **Panel.** The "k of n" line shows for whichever kind is active: "Collision k of n (C: next, N: add)" or "Light k of n (L: next, N: add)".

*Alternative considered:* keeping `activeColl` and adding a separate `activeLight` index. Rejected: every path that reads the index would have to choose one of the two. Only one rectangle is ever active, so one index is enough. Its only cost is renaming the 19 test uses in `collision_rects_test.go`, which is mechanical.

### Shadows and overlays walk the list

`LightOccluders()` appends one world rectangle per light rectangle per object, in `objectsIndexed` order and then list order. For a level with at most one per object, the list it returns is the same as now. The debug overlay outlines every light rectangle. `drawSelection` needs no change beyond `rectOrder`, and `resources.go`'s built-in wall gets a one-element `LightRects`.

## Risks / Trade-offs

- **[`L` changes meaning when light is already active]** It used to do nothing then, and now it steps to the next light rectangle. With one light rectangle, stepping wraps to the same one, so nothing visible changes for any existing tile.
- **[`N` with light active no longer does nothing on a tile that has a light rectangle]** A stray `N` now adds a second light rectangle and changes the fingerprint. Delete removes it, and the panel count shows it happened. This is the same trade the collision change already made.
- **[More occluders cost more per light]** Each rectangle is one extra draw in each light's shadow pass. Tiles that need several are the exception, and occluders already cost nothing when the light is off screen.
- **[A file with both `lightRect` and `lightRects`]** Only a hand edit makes one. The list wins and the load logs it, as for collision.
- **[Archive order]** This restates requirements last restated by `multiple-collision-rects`, in the capability `add-editor-textures-and-rect-editing` introduces. It must be archived after both, and its MODIFIED requirements are written against `multiple-collision-rects`' text.

## Migration Plan

None. Existing files load unchanged and are saved unchanged until a tile is given a second light rectangle. Rolling back leaves any object saved with `"lightRects"` unreadable by the old loader, which drops its shadow without refusing the level. So before rolling back a level, give such tiles one light rectangle each.
