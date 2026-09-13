## Context

See proposal.md - Why. What shapes the approach is where the geometry lives and who is allowed to ask for it.

The attack box is built inline inside `updateAttack`: a `switch` on the attacker's guard height picks one of the three box offsets, adds the fighter's position and a facing term, and the result is a local variable consumed two lines later by the overlap loop. The debug overlay in `Actor.Draw` builds the same three rectangles again with the same expression copied out. So the geometry already exists twice, and the renderer needs a third copy unless it is given a way to ask.

Three constraints bound the answer. The overlap test runs inside the fixed simulation step on the authority; the outline is drawn per rendered frame on whatever machine is looking. The simulation reads `a.Pos`, truncated to whole pixels each step, while the renderer must read `drawPos()` — the interpolated position, rounded — or the outline will lag the sprite it is marking by up to a frame. And the local fighter may be absent: `World.Local` is nil for a spectator, and nil between a death and the respawn that follows it.

The existing per-fighter overlay pass is the model for all of this. `drawStructureBar` is a method on the drawn fighter, it decides its own visibility through small predicates that tests call directly, it lays out in world pixels from `drawPos()`, and it is invoked from a loop after the lighting so nothing tints it.

## Goals / Non-Goals

**Goals:**
- One expression for the attack box, asked by the simulation and the renderer alike, so the outline cannot drift from what the swing actually does.
- Visibility decided by predicates a test can call without a window, the way the bar's are.
- The outline reads only the aiming fighter and the marked one, so it costs nothing when there is no local fighter.

**Non-Goals:**
- Changing hit resolution, the boxes' sizes, or where they sit. The geometry moves; its values do not.
- Touching the debug overlay. It keeps drawing all three boxes in red — it answers a different question, and it is compiled out of release builds while this is not.
- Any new replicated field. The outline is derived per-frame from state a client already holds.
- Marking anything about the target's defence — see the spec requirement that forbids it.

## Decisions

**The attack box becomes one method on the actor, taking the height to build for.** The simulation asks for the box at its own guard height; the renderer asks for the box at the local fighter's height. Both get the same rectangle from the same expression, so "in range" in the outline means exactly "in range" in the swing. Alternative: give the renderer its own copy of the offsets — rejected, since that is the bug this feature would ship with, and the existing duplicate in the debug overlay is the evidence.

**The box is built from a position the caller supplies, not from `a.Pos`.** The simulation passes the simulated position; the renderer passes `drawPos()`. One expression, two positions, no interpolation lag on the outline and no interpolation leaking into hit resolution. Alternative: have the renderer offset the simulated box by the interpolation delta — the same thing, spelled worse.

**Range is the same overlap test the swing uses, run against the target's collision box.** `aabb` of the attack box and the defender's collision rectangle is what decides a hit today; the outline asks the same question of the same two rectangles. That makes the appearance of the outline honest by construction rather than by a tuned constant, and it is why the spec says range is judged by the geometry the attack uses.

**The mark is the sprite's silhouette, drawn by a shader, clipped to the band.** The band supplies the rows; the fighter's own alpha supplies the shape. The height of the band comes from the attack box (6 pixels at one of the three offsets), and the sprite frame supplies the width the shader then clips to the body. Alternative: a rectangle around the body — simpler, but it marks air either side of a narrow fighter and does not say what the swing meets. Alternative: outline the intersection of the two rectangles — visually a fragment that shrinks as the attacker approaches, which reads as a range meter rather than a target band.

**The outline is an outer one: it paints transparent pixels that touch the body.** An inner outline would recolour the fighter's own rim, which is the one thing the spec forbids. There is room for it because the fighter occupies well under the 48-pixel frame it is drawn in, so the outline lands inside the sprite's own destination rectangle and needs no expanded draw. Sampling is clamped to the frame's columns in the sheet: sprites are horizontal strips, and an unclamped neighbour read would trace the pose next door.

**The marked fighter is drawn twice.** The silhouette can only come from the sprite, and the sprite is drawn inside the lit pass — but the mark must not be tinted by a lamp or dimmed in shadow, exactly as the structure bars must not be. So the fighter is drawn once normally where it already is, and once more in the overlay pass through the outline shader, where every fragment but the outline is discarded. The cost is one extra textured quad per marked fighter, which is bounded by how many fighters a single swing can reach.

**The sprite's texture, source rectangle and position come from one place.** The outline pass has to draw exactly what the sprite pass drew, down to the frame and the facing; two copies of that expression would drift the moment either changed, and the outline would trace a pose the fighter is not in. This is the same reasoning as the attack box above, and the same fix.

**Visibility is a predicate taking both fighters, mirroring `showStructureBar`.** It answers: there is a local fighter, it is not this one, neither is dead, the local fighter's guard height is an attack height, time is not showing history, and the boxes overlap. Every clause is a spec scenario, and a test can drive it without a window — which is the only way most of this feature can be tested at all, since drawing needs a GL context.

**Drawn in the same pass as the bars, immediately before them.** The pass already runs after the lighting, so the outline is untinted, and drawing it before the bars keeps a bar readable if the two ever overlap. The loop is over the same sorted draw list, so the outline is drawn per marked fighter rather than once for the local one — which is what makes two targets in range both get marked, for free.

## Risks / Trade-offs

- **The outline hands intermediate players a distance meter they used to learn by feel** → that is the point of the feature, but it is a real change to how the opening of a fight plays. It is confined to presentation and to a single draw call site, so it can be put behind a setting or removed without touching the simulation if it plays worse than it reads.
- **Two fighters at arm's length each carry an outline plus a structure bar, a pip row and an energy bar** → the bar family is already tightly packed above the head; the band sits over the body instead, which is empty. Worth a look during the in-game verification, and the reason the band is an outline rather than a fill.
- **The lagging attack cursor means the height can change under the player's hand after they have committed to the click** → the outline makes that visible for the first time, which may read as the outline being wrong rather than the cursor being slow. If it does, the fix is the cursor's lerp speed, not the outline.
- **`Updown` is sampled per tick and the outline is drawn per frame** → at 60 Hz simulation and a higher frame rate the band can hold a stale height for part of a frame. It is the same state the swing will use, so the outline is never wrong about what would happen — only, at worst, one tick behind the mouse.
