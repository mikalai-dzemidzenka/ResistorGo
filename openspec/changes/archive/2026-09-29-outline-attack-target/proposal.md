## Why

Nothing on screen says where a swing will go. Attack height comes from the attack cursor, a lagging green cross that trails the mouse, measured against *your own* fighter's body — so the three bands are invisible boundaries on a 48-pixel sprite, and the cursor sits somewhere between where the mouse is and where the height was read from. A player aiming for the head has no way to tell they are a few pixels into mid until the swing comes out at the wrong height and gets blocked.

Range is the same guess. The attack box reaches 24 pixels in front of the fighter, and there is nothing to say whether the opponent is inside it. The two together make the opening of an exchange a matter of feel: players learn the distances by dying at them.

Debug mode draws the boxes, but it is the wrong tool — it is developer scaffolding behind the backtick key, it draws all three heights at once in red rectangles, and it is compiled out of release builds entirely. The information a player needs is narrower: the one band the current height would strike, on the fighter it would strike, and only when a swing would actually reach.

## What Changes

- A fighter in range of the local player's swing at its current attack height SHALL be outlined across the horizontal band that swing would land on, so the player can see both where the swing goes and that it reaches.
- The outline follows the attack height as it changes: aiming from low to high slides the band up the target's body, and it slides while the guard height is being chosen rather than only once the swing starts.
- The outline is drawn only while a swing would actually reach: it appears when the target comes into range and disappears when it leaves. A fighter out of range carries nothing.
- Block-all (guard height 3) has no attack of its own, so nothing is outlined while the local player holds it.
- The outline says where the swing goes and **not** whether it would land. It does not react to the target's guard height, its structure, or its energy: reading the opponent's guard stays the player's job, which is the fight this game is built on.
- It belongs to the fighter and the scene, like the structure bar above it: drawn in world pixels so it zooms with the camera, and drawn above the lighting so it is never tinted.
- It is for the local player only. A spectator has no fighter and no attack height, so a spectated match carries no outlines at all.

## Capabilities

### New Capabilities

- `attack-target-outline`: which fighter is marked as the target of the local player's next swing, the band of it that is marked, and when the mark is shown at all.

### Modified Capabilities

None. The outline reads existing state and changes nothing about how attacks resolve, so the `structure` and `energy` specs are untouched.

## Impact

- A new drawing pass alongside the existing per-fighter overlay pass, reading the same geometry the attack resolution already computes — the three attack boxes in front of the attacker and the target's collision box.
- The attack box the renderer asks about is currently built inline inside the attack resolution and nowhere else, so that geometry has to become something both the simulation and the renderer can ask for, without changing what it returns.
- Presentation only: no change to hit resolution, no change to input, nothing added to the wire, and no new asset. The outline is drawn from state a client already holds, so it works the same offline, on a listen server and on a remote client.
- The debug overlay's attack boxes stay as they are — developer scaffolding is a separate thing from a player-facing cue.
