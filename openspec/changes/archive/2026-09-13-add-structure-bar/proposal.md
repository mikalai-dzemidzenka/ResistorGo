## Why

Structure is now the whole health model, and nothing on screen shows it. A fighter can be one hit from dying or one hit from a stun and look exactly as it did at full structure; the number only exists in the debug overlay. The structure change called a real meter out as the expected immediate follow-up (`openspec/changes/archive/2026-09-13-add-structure-system/proposal.md`, Non-goals). Without one, pressure has no visible payoff until the moment a fighter falls, and a player cannot tell whether blocking is buying anything.

## What Changes

- **A structure bar above every fighter's head,** posture-style: empty at full structure and filling as structure is lost, so a full bar means the guard is about to break. It follows the fighter as it moves and is part of the pixel-art scene, not an overlay on top of it.
- **Shown only while structure is below full.** The bar appears on the first loss and disappears again once regeneration has refilled the pool. An untouched fighter carries no bar.
- **Stun pips above the bar:** up to two small marks in a row centred just above it, one for each landed hit toward the next stun. They clear when the stun lands and when regeneration starts, exactly as the hit count does.
- **Not a debug feature.** The bar and pips are drawn whether or not the debug overlay is on, for the local fighter, bots, and every fighter a networked client or spectator can see.
- **The hit count is replicated.** Structure already travels in the snapshot; the count toward the next stun does not, so a client could draw a remote fighter's bar but never its pips. It is added to the snapshot, packed into spare bits of the existing flag field, and the protocol version goes up.

### Assumptions

- The bar fills from its centre outward, as Sekiro's posture bar does, and is 16 world pixels wide and 2 tall, sitting a few pixels above the top of the fighter's collision box. It scales with the camera zoom like the rest of the scene.
- One fill colour and one track colour, both taken from the game palette so the palette pass leaves them unchanged. The bar does not change colour as it fills.
- The bar is drawn after the scene's lighting, so a fighter standing in shadow or under a lamp gets the same bar.
- No bar during the Death state: a broken structure has already been shown by the break itself, and the fighter is out of the fight.
- No bars while time manipulation is showing a recorded frame. A recorded frame does not hold structure, so a bar there would show the present over a pose from the past.

## Capabilities

### New Capabilities
- `structure-bar`: the per-fighter structure bar and stun pips — what they show, when they are visible, where they sit, and which fighters get them.

### Modified Capabilities
- `multiplayer/replication`: *Snapshots* — the per-fighter state a snapshot carries now includes the count of landed hits toward the next stun, so a client can show a remote fighter's pips.

## Impact

- **Code**
  - `actor.go`: a draw function for the bar and pips, reading `Structure`, `HitCount` and the drawn position the sprite already uses.
  - `game.go`: a second pass over the draw list after the light pass in `runLevel`, inside the same render texture and camera, so the bars are palette-quantized and pixel-snapped but not lit.
  - `net_protocol.go`: the hit count in two spare bits of `ActorRec.Flags`, applied on decode; protocol version 4. The record size does not change.
- **Compatibility**: the protocol version rises, so builds from before this change are rejected with the existing "version mismatch".
- **Not affected**: the structure rules themselves, the simulation, prediction, lag compensation, the level editor and the menus.

## Non-goals

- A screen-space HUD (a fixed meter in a screen corner), numbers, or a structure readout outside the debug overlay.
- Animating the bar: no easing, no flash on hit, no drain effect as regeneration refills it.
- Showing which fighter is about to break in any way other than the bar itself.
- Changing any structure value or rule.
