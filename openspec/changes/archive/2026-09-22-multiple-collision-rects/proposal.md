## Why

A level object has exactly one collision rectangle, and the solid part of a tile is not always one box. A wall with a ledge, a block with a gap to walk through, or a vent that is solid top and bottom but open in the middle cannot be described by a single rectangle. Today that leaves two bad choices. One is a rectangle that covers too much, so fighters collide with air. The other is cutting the art into more tiles than it is, to get more rectangles. The editor can now shape any rectangle a tile has, but it cannot give a tile a second one.

## What Changes

- **An object can have any number of collision rectangles**, including none, where today it has zero or one. Fighters collide with every one of them, each behaving exactly as a tile's single rectangle does now: stood on, blocking the head, and with ledges that can be climbed.
- **The rect tool edits each of them.** All of the selected thing's rectangles are drawn. Clicking any of them makes it the active one and starts shaping it. `C` steps through the collision rectangles, `N` adds another one, and Delete removes the active one, leaving the rest. The panel says which collision rectangle is active, and how many there are.
- **They all travel together.** Placing a tile, a texture's defaults, a tile following its texture, copy and paste, and saving all carry the whole set of collision rectangles, where each carries one rectangle today.
- **The level file keeps its shape.** An object with one collision rectangle is saved exactly as now, so `level_1.json` and its fingerprint are unchanged until a tile is actually given a second rectangle. Only objects with two or more are saved with a list. The texture defaults file does the same.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `level-editor`: *Rectangles are shaped on screen* and *A rectangle can be added and cleared* are restated for several collision rectangles per object. New requirements cover collision with every rectangle, the set travelling together, and the file keeping its shape. This capability is introduced by the in-flight `add-editor-textures-and-rect-editing` change, so this change has to be archived after it.

## Impact

- `level_editor.go`, `level_objects.go`, `resources.go`: the collision rectangle becomes a list, with its own JSON handling so that one rectangle is still written as `"collisionRect"`. The rect tool learns which collision rectangle is active.
- `actor.go`: the two collision loops, `headBlocked` and `checkCollision`, walk the level's collision rectangles instead of its objects, and nothing else in the file changes. That file also carries unrelated uncommitted work, so this change touches only those two loops.
- Tests that read `CollisionRect` (about 60 uses, mostly in the editor tests) move to the list, mechanically.
- Multiplayer: nothing new travels. Both peers collide with the level they loaded. A level that uses several rectangles has a different fingerprint, as any edit to the level does, so a build that cannot read the list is refused rather than playing different geometry.
- Not in scope: several light rectangles per object, shapes other than rectangles, and one-way platforms.
