## Why

A running fighter cannot jump. The jump is offered from the idle state only, so pressing jump at a sprint does nothing at all — the input is swallowed and the fighter keeps running into whatever it was running at. There is no way to clear a gap or a crate without stopping first, taking the 0.7-second crouch, and setting off again.

The jump the game does have is also built for standing still. It opens with a 13-frame wind-up that reads as gathering yourself, which is right from a stop and wrong at full speed, where it would read as a stumble.

And nothing in the game falls. Once the wind-up finishes, the airborne fighter is left holding the last frame of the crouch for the whole arc — up, over the top and all the way down, the same still pose. Walking off a ledge does not even change state: the idle branch has carried a `// TODO state fall` for as long as it has existed, so a fighter steps off a platform still playing its idle. Both animations are already drawn and sitting in the resources folder unused.

## What Changes

- **A run-jump.** A running fighter that presses jump begins the running-jump animation on that tick and pushes off on its third frame, rather than taking the standing jump's separate 13-frame crouch first. The anticipation is the two frames the animation already draws, not a state of its own. The standing jump keeps its crouch exactly as it is.
- **The run's momentum is kept.** A run-jump carries the speed the run had into the air rather than launching from a standstill, and the fighter can still steer while airborne, as it can in a standing jump today.
- **A fall state, for everything that falls.** Any fighter that is off the ground and moving downward falls: after a run-jump, after a standing jump's apex, when walking off a ledge, and when it is put off an edge by a knockback. It plays the falling animation, which nothing currently does. This closes the `// TODO state fall` the idle branch has been carrying.
- **A fall still catches a ledge.** The climb a jump can reach today stays reachable while falling, so a jump that overshoots can still be saved on the way down.
- **A run-jump costs energy: half the pool.** Leaving the ground at a run spends 100 of the 200-point pool, so a fighter gets two run-jumps from full and has to earn the next one back. A fighter that cannot pay does not jump: the input is refused the way an unaffordable swing is, nothing is spent, and the refusal is raised so the empty bar flashes.
- **The standing jump stays free**, as it is today. What is priced is the mobility this change adds, not leaving the ground as such.
- **Landing puts the fighter back into what it was doing**: still moving, it runs; stopped, it stands.

## Capabilities

### New Capabilities

- `jumping`: how a fighter leaves the ground, what it can do while airborne, when it is falling rather than rising, what it catches on the way down, and what it lands into. It covers both jumps, the fall and the landing.

### Modified Capabilities

- `energy`: *Energy pool* — the pool no longer gates attacking alone. Its rule that "moving, jumping, climbing and blocking SHALL cost nothing and SHALL be available at any energy" is what the run-jump's cost contradicts, and its scenario that an empty fighter can still jump has to say which jump it means.

## Impact

- Two new actor states and the transitions between them, in the state machine in `actor.go`. The states travel to clients inside the snapshot as a number, and the animations travel as an index into an append-only table, so both have to be **appended** rather than slotted in beside their relatives — inserting either one silently renumbers every state or animation that follows it.
- `run-jump.png` (6 frames) and `fall.png` (3 frames) already exist in the resources folder but are registered nowhere: they are absent from both the texture loader and the network animation table. Drawing an unregistered texture panics, and sending an animation that is not in the table panics on the server as it builds a snapshot — so registration is not optional tidying, it is the first thing that has to happen.
- Falling is judged from the vertical speed the simulation already keeps and the ground flag it already sets; no new field, and nothing new on the wire beyond the state and animation values themselves.
- The run-jump's cost rides on the energy machinery that already exists — the pool, the calm wait that spending restarts, the refusal event and the bar that flashes on it — so pricing the jump is a new caller of that machinery rather than new machinery. The one thing it does need is for spending to stop being hard-wired to what a swing costs.
- The airborne rules apply to bots and networked fighters as much as to the local player, since they live in the shared fixed step rather than in input handling.
