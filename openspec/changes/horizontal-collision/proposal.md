## Why

Fighters only collide with the level at one point: the middle of their collision box. The floor is found under that point, the ceiling at the head above it, and a ledge within reach of it. Nothing ever tests the sides of the box, so the side of a collision rectangle stops nobody. A fighter walks into a wall until its middle is inside, and if its feet are anywhere within the rectangle's height it is snapped straight up onto the top, however tall that is. Now that the editor can give a tile several collision rectangles, walls are easy to build, and none of them work.

## What Changes

- **The side of a collision rectangle stops a fighter.** A fighter moving sideways into one stops with its collision box against it, and its sideways speed goes to zero. This applies however the fighter is moving: walking, running, steering in the air, thrown by a hit, or dashing forward into a swing.
- **A small step is still walked up.** A rectangle whose top is no more than a few pixels above the fighter's feet is a step, not a wall. The fighter walks onto it as it does today. Anything taller is a wall, and the way over it is a jump and a climb.
- **Ledges still catch.** A fighter jumping or falling against a wall still grabs its top and climbs it, as it does at a tile's edge now.
- **Nothing throws a fighter out of a wall.** A fighter that already overlaps a rectangle, as a spawn or a climb can leave it, is never pushed out. It can move out in any direction, but never deeper in.
- **The current level plays as before.** In `level_1.json`, neighbouring floor tiles share a top, and the floors are further apart than a fighter is tall, so no rectangle in it is a wall to anyone.

## Capabilities

### New Capabilities

- `level-collision`: how fighters collide with the level's collision rectangles from the side. It covers walls stopping a fighter, steps it can walk up, and ledges that still catch against a wall.

### Modified Capabilities

None. Standing on floors, head ceilings, landing and ledge catching are unchanged. Those last two are introduced by the in-flight `add-run-jump-and-fall` change's `jumping` capability, and this change keeps them as they are.

## Impact

- `actor.go`: a side check after the fighter moves horizontally in `Update`, and after the swing's forward dash in `enterAttackState`. One new constant, the step height. `checkCollision`, floors, ceilings and ledges are not changed.
- It depends on `LE.solids`, the flat list of world-space collision rectangles introduced by the in-flight `multiple-collision-rects` change, so it has to be applied after that change.
- Multiplayer: nothing new goes over the network. The client's prediction and the server run the same `Update` over the same level, so they agree on where a wall stops a fighter. The level fingerprint is unchanged.
- Bots have no pathfinding. A bot with a wall between it and its target walks into the wall and stays there.
- Not in scope: fighters blocking each other, one-way platforms, and floors or ceilings tested across the whole width of the box. Also left as it is: the reach of the ledge catch. It grabs whenever the head is anywhere within a rectangle's height, so a fighter can jump at a tall one-rectangle wall and climb it from part-way down. That is a separate change if walls are to be unclimbable.
