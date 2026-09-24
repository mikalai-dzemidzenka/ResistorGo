## Why

A fighter has only two ways to get around: walking or running, and the two jumps. Neither lets a player choose how far to go. Three sprites for a charged roll are already in the resources folder but nothing uses them: `duck.png` (4 frames), `roll.png` (5) and `roll-recover.png` (4). This change adds a Gunpoint-style roll. Holding the right mouse button winds the roll up, the game draws the path it will take, and letting go sends the fighter along that path.

## What Changes

- **Holding RMB ducks.** A fighter standing or running on the ground that presses the right mouse button stops and crouches (`duck.png`, played once and then held on its last frame). It stays down for as long as the button is held.
- **The hold charges the roll.** The longer the button is held, the higher and longer the roll goes, up to a cap after about one second. The roll goes toward the side the cursor is on, the same side the fighter already faces. The cursor still chooses the guard height; it does not aim the roll.
- **The path is drawn while charging.** While the local player is ducking, a dotted arc shows where the roll would go if released now: the same launch speed, gravity and time step the simulation uses. The arc stops at the first solid it meets. It is drawn for the local player only, over the lighting, like the attack outline.
- **Letting go rolls.** On release the fighter leaves the ground on that arc and plays `roll.png` in a loop while in the air. It does not steer or catch ledges while rolling.
- **Touching the ground recovers.** Landing plays `roll-recover.png` once. When it finishes, the fighter stands, or runs if it still has speed, the same way a landing from a fall does.
- **A roll costs the same as a run-jump.** It costs a third of the pool. A fighter that cannot pay does not duck. The press is refused the way an unaffordable run-jump is: nothing is spent and the bar flashes. The cost is taken when the roll launches, and a duck that is interrupted costs nothing.
- **Being hit interrupts it.** A stun or death takes over from the duck, roll or recovery as it does from any other state.

## Capabilities

### New Capabilities

- `rolling`: the duck, the charge and its cap, the drawn arc, the roll through the air, the recovery on landing, what interrupts each stage, and the price.

### Modified Capabilities

- `energy`: *Energy pool*. The pool gates the roll as well as swinging and the run-jump.
- `multiplayer/replication`: *Input commands*. The command carries whether the roll button is held. It is a held button, not an edge.

## Impact

- `actor.go`: three new states, `StateDuck`, `StateRoll` and `StateRollRecover`, appended after `StateClimbDown` and before `stateCount`. A `Roll` field on `Input`. Sampling RMB in `sampleInput`. The duck is entered from the idle and run branches. The charge is read from the ticks since the duck animation started (`Ticks - AnimStartTick`). Both values are already replicated, so the client can reconcile the charge without a new field.
- `resources.go` / `animTable`: the three textures are registered, and the three paths are appended to the network animation table.
- `net_protocol.go`: a `bRoll` bit in the command's button byte. Two of its eight bits are still free.
- `game.go`: draws the arc for `w.Local` next to `drawAttackOutline`.
- The level editor already uses RMB to delete. The match does not run while editing (`matchHeld`), so the two uses do not overlap.
- Out of scope: bots do not roll.
