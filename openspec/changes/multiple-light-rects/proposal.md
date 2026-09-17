## Why

A tile can now have any number of collision rectangles but still only one light rectangle, the rectangle that casts its shadow. The part of a tile that should block light is no more always one box than its solid part is: a wall with a window, an arch, or a vent with a gap in it shadows everything behind it or nothing, where it should let light through the opening. `multiple-collision-rects` left this out of scope. Now that tiles with several collision rectangles exist (`level_1.json` has one), the light rectangle is the one kind that still cannot follow the shape.

## What Changes

- **An object can have any number of light rectangles**, including none, where today it has zero or one. Every one of them casts a shadow exactly as a tile's single light rectangle does now.
- **The rect tool edits each of them, the way it edits collision rectangles.** Every light rectangle of the selection is drawn. Clicking any of them makes it the active one. `L` switches to the light rectangles or, if they are already active, steps to the next one. `N` adds another rectangle of whichever kind is active, a light one included even when the object already has one. Delete removes only the active rectangle. The panel says which light rectangle is active and how many there are, as it does for collision.
- **They all travel together.** Placing a tile, a texture's defaults, a tile following its texture (and the check of whether it still does), copy and paste, and saving and loading all carry the whole set of light rectangles, as they already do for collision rectangles.
- **The level file keeps its shape.** An object with one light rectangle is saved exactly as now, as `"lightRect"`. Only objects with two or more are saved with a `"lightRects"` list. The texture defaults file does the same. `level_1.json` and its fingerprint are unchanged until a tile is actually given a second light rectangle.
- **The debug overlay** draws every light rectangle of every tile, not just one.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `level-editor`: *Rectangles are shaped on screen* and *A rectangle can be added and cleared* are restated for several light rectangles per object. *Every collision rectangle travels together* and *The level file keeps its shape* are restated to cover light rectangles too. A new requirement says an object casts a shadow from every light rectangle it has. The capability is introduced by the in-flight `add-editor-textures-and-rect-editing` change, and the restated requirements were last restated by the in-flight `multiple-collision-rects`. This change has to be archived after both.

## Impact

- `level_editor.go`, `level_objects.go`, `resources.go`: `LevelObject.LightRect *rl.Rectangle` becomes `LightRects []rl.Rectangle`, and `ObjectRects.Light` becomes a list. The file's JSON handling writes one light rectangle as `"lightRect"`, as the collision handling already does for `"collisionRect"`. The rect tool's collision index becomes one index for whichever kind is active. `LightOccluders()`, the debug overlay and `drawSelection` walk the list.
- `actor.go`, the network code and the shadow passes: unchanged. `computeLights` already takes a flat list of occluders, and it has no limit on how many.
- Tests that read `LightRect` or `ObjectRects.Light` (about 19 uses, in `level_editor_test.go`, `collision_rects_test.go` and `reuse_rects_test.go`) move to the list.
- Multiplayer: nothing new travels. A level that uses several light rectangles has a different fingerprint, as any edit to its tiles does, so a build that cannot read the list is refused rather than drawing different shadows.
- Not in scope: shapes other than rectangles, and light rectangles that let some light through.
