## Why

A fighter can go up in the level (jump, catch a ledge, climb) but the only way down is to walk off an edge. There is no floor you can get down *through*, so every level has to be built with a way around the side. `climb-down.png` is now in the resources folder (5 frames, 48×96: the fighter crouches, grabs the floor it is standing on and lowers itself until it hangs from it), but nothing uses it yet.

## What Changes

- **A new kind of floor: the drop-through platform.** It works like any solid for standing on, walking along and bumping your head on from below. A jump can also catch its ledge. The one difference is that a fighter standing on it can climb down through it.
- **S climbs down.** A fighter that is standing or walking on a drop-through platform and presses S plays the climb-down animation where it stands. The fighter stays put while it plays, so no steering and no swinging. When the animation ends the fighter is below the platform and moves into the fall state that already exists, falling to whatever is underneath.
- **S does nothing anywhere else.** On ordinary floor, in the air, or while swinging, blocking or stunned, the press is ignored.
- **A fighter that has dropped through is not caught by the platform on the way down.** It does not land back on the platform, and it does not grab its ledge.
- **A stun interrupts a climb-down.** The fighter is left on top of the platform, stunned, as it would be from any other stance. A flinch only flickers, the same as it does everywhere else.
- **The level editor gets a third kind of rectangle** next to collision and light: *drop-through*. An object can have any number of them, mixed with ordinary collision rectangles. They are shaped, added, removed, copied, followed and saved exactly like the other two kinds. A level with none of them keeps its bytes and its fingerprint.
- **The climb-down action goes over the network.** The input command gets one more button bit, and the new state and animation are appended to the tables the snapshot indexes into.

## Capabilities

### New Capabilities

- `climb-down`: the drop-through platform as fighters meet it (standing on it, bumping it from below, catching its ledge), which fighters can climb down through it and when, what the climb-down looks like, and how it hands over to the fall.

### Modified Capabilities

- `level-editor`: *Rectangles are shaped on screen*, *A rectangle can be added and cleared*, *Every collision rectangle travels together* and *The level file keeps its shape* all say there are two kinds of rectangle. Each one gains the drop-through kind.
- `multiplayer/replication`: *Input commands*. The command's edge-triggered actions are listed by name (attack, jump), so climbing down has to be added to that list.

## Impact

- `actor.go`: a new state, `StateClimbDown`, appended after `StateClash` and before `stateCount` so no existing state number shifts. It is entered from Idle and Run, and finishes by handing over to `enterFallState`. `checkCollision` has to tell drop-through rectangles apart from solid ones for a falling fighter that is already below a platform's top. `Input` gets a `Down` field, and `latchInput`/`sampleInput` read S as an edge-triggered press, the same way W is read.
- `resources.go`: `climb-down.png` is registered at the **end** of the texture list. That list is also the network animation table, so inserting the file anywhere else would renumber every animation after it.
- `level_objects.go` / `level_editor.go`: `ObjectRects` and `LevelObject` get a third list. It goes on disk as `dropRect`/`dropRects`, with the same one-or-many rule the other two kinds follow. The rect tool gets a third kind to switch to, and the world-space rectangles collision walks now record which ones are drop-through.
- `net_protocol.go`: one new button bit (`bDown`) in the command. The snapshot record does not grow: whether a fighter is dropping is worked out from where it is and which way it is moving, so no new field is replicated.
- Bots do not climb down, and their pathing treats a drop-through platform as ordinary floor, which it is for anyone standing on it. Lua scripts and blood read the rectangles as they do today.
