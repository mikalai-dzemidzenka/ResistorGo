## Context

See proposal.md - Why. What shapes the approach is that two of the three things being added are numbers on the wire, and the third is a rule the state machine has no place for yet.

A fighter's state and its animation both travel in the snapshot as small integers. The state is the enum's own value; the animation is an index into a table whose comment says, in as many words, that client and server must agree on the order and that it is append-only. The lookup that turns a path into that index panics when the path is missing — so an animation used before it is registered does not degrade, it takes the server down as it builds its first snapshot containing it. Drawing an unregistered texture panics the same way. Both new animations are in this position right now: present on disk, absent from the loader and absent from the table.

The state machine's airborne half is thin. Jump is offered from the idle branch only; the run branch does not look at the jump input. Take-off is split across two states because of the wind-up — one waits out the animation, the next sets the upward speed — and the airborne state sets no animation at all, which is why the wind-up's last frame is what a jumping fighter wears all the way down. The idle branch carries a `// TODO state fall` for the case it cannot handle.

The ground rules that matter are already in place and shared: gravity and the ground flag are applied every tick for every fighter before the state switch runs, horizontal acceleration is allowed in every state except attacking, blocking and being hit, and the ledge flags are recomputed by the collision pass rather than by the jump.

## Goals / Non-Goals

**Goals:**
- Falling decided in one place, from the speed and ground flag the simulation already keeps, so every way of ending up in the air reaches it.
- New states and animations appended, never inserted, so no existing value on the wire changes meaning.
- The run-jump distinguished from the standing jump by its take-off only, not by how high it goes.

**Non-Goals:**
- Retuning jump height, gravity, air acceleration or running speed. The arc is what it is today; this change is about which state and which animation the fighter is in while flying it.
- Double jumps, air attacks, landing recovery, or a landing animation. A landing returns the fighter to running or standing, as the spec says, and nothing more.
- Touching how the wind-up jump works from a standstill.

## Decisions

**Both states are appended to the end of the enum.** Its values are what the snapshot carries, so putting a run-jump next to the jump it resembles would renumber the climb, the turnover and death underneath it — every fighter a client is already drawing would change state the moment the two sides disagreed. Appending costs nothing but the reading order of a declaration. The same reasoning, already written down in the animation table's own comment, applies to the two new paths: they go on the end.

**Falling is a transition tested before the state switch, not a case inside it.** The alternative is a check in each airborne branch — the jump, the run-jump, and the idle branch's TODO — which is three places to keep in step and still leaves out anything else that ends up airborne, like a knockback over an edge. One test, before the switch, covering "off the ground and descending", reaches all of them and is why the spec can say the fall is entered from any state. It must not fire for a fighter that is rising, climbing, or dead.

**The run-jump sets the upward speed itself and skips the two-step take-off.** The wind-up state exists only to wait out an animation the run-jump does not play, so the run-jump enters the airborne state directly, on the tick the input arrives. It sets the same upward speed the standing jump does — the spec requires equal height — and leaves horizontal speed alone, which is all "keeping the run's momentum" needs, since nothing in the take-off path resets it today.

**The fall keeps the climb by being tested for the same ledge flags the jump is.** The flags are maintained by the collision pass for every fighter regardless of state, so the fall's branch asks the same question the jump's branch asks. This is a decision rather than a detail: it is the one place where the fall could quietly differ from the jump, and the spec pins it because a mistimed jump staying recoverable is what makes the new state feel like a state rather than a punishment.

**Landing reads the horizontal speed to choose between running and standing.** The turnover state already ends this way, so the fall borrows the shape rather than inventing one: still moving, run; stopped, stand. It also keeps a landing from being a lockout, which the spec requires.

## Risks / Trade-offs

- **A new animation used before it is registered panics rather than degrading** → this is a crash on the first snapshot or the first draw, not a visual glitch. Registration is the first task, before any state uses either path, and both builds are run before the change is called done.
- **Two more states on the wire between versions that disagree** → the handshake already refuses a client whose protocol version differs, so the exposure is a build talking to a build it was never meant to; the appended ordering means a stale client misreads nothing that existed before.
- **The idle branch's TODO disappearing changes behaviour nobody asked about** — a fighter that walks off a ledge currently keeps its idle animation, and will now fall. That is the fix the TODO was asking for and the spec requires it, but it is a visible change in a case outside the literal request.
- **A fall entered from a knockback could fight the hit reaction for the animation** → the hit reaction no longer sets an animation for a flinch, so the two do not collide today; a stun does set one, and the fall must not steal it while the fighter is locked out. The spec keeps the stun's own state untouched, and that boundary is worth checking in play as well as in a test.
