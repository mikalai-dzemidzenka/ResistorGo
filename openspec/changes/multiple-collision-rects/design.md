## Context

See `proposal.md` — Why. What the approach has to work with, all as committed in `c0fb088`:

- **The data.** `LevelObject.CollisionRect *rl.Rectangle` is one optional rectangle, relative to the object's origin. The in-memory pair `ObjectRects{Collision, Light *rl.Rectangle}` (`level_objects.go`) carries a tile's or a texture's rectangles through `rects()`/`setRects()`, `clone()`, `sameRects()`, `builtinRects()`/`buildPalette()` and `level_objects.json`.
- **The file and the fingerprint.** `encoding/json` over `LevelObject` writes `resource, layer, x, y, collisionRect, lightRect` with `omitempty`. `marshal()` of the tiles is the fingerprint's input. `level_1.json` holds 55 tiles, each with at most one collision rectangle.
- **Collision.** `actor.go`'s `headBlocked` and `checkCollision` walk `LE.objectsIndexed` and read `obj.CollisionRect`. The order matters to the climb checks: `objectsIndexed` is sorted by layer, then Y and X descending, and `headBlocked` runs first because a neighbouring tile can be visited before the one the head is in. These two loops are the only readers outside the editor. The client's prediction and the server's simulation both go through them, over the same loaded level.
- **The rect tool.** One active rectangle, `activeRect` ∈ {collision, light}, reached through `active()` → `(ObjectRects, *rl.Rectangle, origin, ok)` and written through `setActive(*rl.Rectangle)`, where nil clears. `beginDrag` hit-tests only the active rectangle. `C`/`L` pick the kind, `N` adds when there is none, and Delete clears. The panel shows a Collision/Light toggle and four fields.
- **`actor.go` carries unrelated uncommitted work** alongside `c0fb088`.

## Goals / Non-Goals

**Goals:**

- Every object and texture has an ordered list of collision rectangles; every path that carried one carries the list.
- Levels and texture defaults with at most one collision rectangle per object keep their exact bytes and fingerprint.
- Collision behaves per rectangle exactly as it does per object today, climb order included.
- The change in `actor.go` is confined to the two loops.

**Non-Goals:**

- Several light rectangles, non-rectangular shapes, one-way platforms.
- Merging overlapping rectangles, or snapping them to each other.

## Decisions

### A list in memory, one or many on disk

`LevelObject.CollisionRect *rl.Rectangle` becomes `CollisionRects []rl.Rectangle`, values rather than pointers: nil or empty means no collision. `ObjectRects.Collision` becomes `[]rl.Rectangle` likewise. `clone()` copies the slice, and `sameRects()` compares length and each rectangle in order, as well as the light rectangle as before.

`LevelObject` and `ObjectRects` get their own `MarshalJSON`/`UnmarshalJSON` through a wire struct that keeps today's field order:

```
resource, layer, x, y, collisionRect, collisionRects, lightRect
```

- **One rectangle** is written as `"collisionRect"`, exactly as now.
- **Two or more** are written as `"collisionRects": [...]` holding all of them, with no `"collisionRect"`.
- **None** writes neither.

On load, `"collisionRect"` gives a list of one and `"collisionRects"` gives the list. A file carrying both is malformed; the list wins and the load logs it.

Because a one-rectangle object marshals to the same bytes as before, `marshal()` — and so the fingerprint — is unchanged for every existing level and texture defaults file. A test pins this against a copy of the old struct.

*Alternative considered:* always writing `"collisionRects"`. Rejected: it rewrites every tile in `level_1.json` and changes the fingerprint for a level nobody has edited.

*Alternative considered:* keeping `CollisionRect` for the first rectangle with an extra list for the rest. Rejected: every reader would have to remember to look in two places, which is exactly how one of them would get missed.

### Collision walks a flat list the index keeps

`buildIndex` also builds `le.solids []rl.Rectangle`: every collision rectangle of every object, in world space, in `objectsIndexed` order and, within an object, in list order. `headBlocked` and `checkCollision` loop over `LE.solids` with the same body they have now; `oColX`/`oColY`/width/height simply come from the rectangle. Nothing else in `actor.go` changes. The climb checks see the same order as before for one-rectangle levels, because the flattening preserves object order.

*Alternative considered:* a nested loop over each object's rectangles inside the existing loops. Rejected: the same code in two places and a world-space conversion per rectangle per tick, where the index can do it once per edit.

### The active rectangle is a kind and an index

`activeRect` stays collision or light, and `activeColl int` says which collision rectangle. `active()` returns the rectangle for the pair, and nil when that index does not exist. `setActive(nil)` removes a collision rectangle at `activeColl`, or clears the light rectangle, as now.

- **Picking.** `beginDrag` hit-tests the active rectangle first, then every other rectangle of the selection: collision in order, then light. The first hit becomes active and is dragged. This is what lets a click pick any rectangle, and it keeps "the active one wins" when rectangles overlap.
- **`C`.** Makes collision active, or steps `activeColl` to the next one, wrapping, if collision is already active. `L` is unchanged.
- **`N`.** With collision active, appends a one-cell rectangle at the origin and makes it active. With light active, adds the light rectangle when there is none, as now.
- **Delete.** Removes the active rectangle. `activeColl` then moves to the previous one, or to 0, so there is still an active one while any remain.
- **Selection changes.** `activeColl` is clamped whenever the selection changes, so a new selection never points past its list.

The panel's toggle row gains "Collision k of n (C: next)", and the four fields edit the active one. `drawSelection` draws every collision rectangle, the active one filled with handles and the rest as outlines. The debug overlay and `DrawCursor` draw every collision rectangle of each object rather than one.

### Built-in rectangles

`LoadLevelObjects()` keeps its three entries; their single `CollisionRect` literals become one-element `CollisionRects`. Nothing else about the built-in table changes.

## Risks / Trade-offs

- **About sixty test references move from a pointer to a list.** → Mechanical, and done before any behaviour changes, so a test that starts failing afterwards is a real regression rather than a rename.
- **Overlapping rectangles are easy to make and hard to pick apart.** → The active one wins a click, `C` steps through them all, and the panel says which one is active. Nothing stops overlap; it is harmless to collision.
- **A file with both keys** can only come from a hand edit. → The list wins and the load logs it, rather than refusing the level.
- **`actor.go` shares the working tree with unrelated uncommitted work.** → The edit is two loop headers, and committing it will need the same one-hunk staging `c0fb088` used.
- **Archive order.** → This restates two requirements introduced by `add-editor-textures-and-rect-editing`, so it must be archived after that change.

## Migration Plan

None. Existing files load unchanged, and are saved unchanged until a tile is given a second rectangle. Rolling back leaves any object saved with `"collisionRects"` unreadable by the old loader, which would drop its collision. So roll back only a level that has no such object, or give those objects one rectangle first.
