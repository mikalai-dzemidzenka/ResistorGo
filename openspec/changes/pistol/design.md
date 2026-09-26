## Context

- A fighter already has a two-layer draw. `topAnim`/`topDraw` put a second sheet over the body (the walk arms), and the rim light draws from `topDraw`. That second layer always uses the body's frame, while the pistol torso needs its own frame: the aim.
- The roll already turns the cursor into a byte in the command (`Input.RollAim`, `rollAim`), so the server and client get the same angle. Aim follows the same pattern.
- The authority decides melee hits in `updateAttack` through `takeHit`/`confirmBlock`. `Server.runCmd` rewinds other fighters only while `a.State == StateAttack`.
- Blocks are "heard" on every machine through a replicated counter (`Blocks`, parity flag), not an event. `Events` are local to a machine and are restored after a client replay.
- Lights reach the screen through `computeLights(LE.lights, ...)` in `game.go`. A `LevelLight` is a plain value with a centre, radius, colour and intensity.
- The command's button byte has one free bit (bit 7). The actor record's 16 flag bits are all taken.
- Sheets: `pistol-top.png` is 240×48, with five frames from up to down. `gun-bot-idle.png` is one 48×48 frame. Both share the fighter's 48×48 frame and its origin.

## Goals / Non-Goals

**Goals:**
- One `Weapon` value on the fighter that later guns extend: a new gun is a new value plus its sheet, muzzle table and numbers, not a new mechanism.
- Legs chosen in one place, so walk and run legs slot in later.
- Shots decided only by the authority, predicted locally for the tracer, and lag-compensated like swings.

**Non-Goals:**
- A weapon or inventory system, pickups, ammo, reload, recoil or spread.
- Bot gun use, and bots reacting to a gun beyond taking hits.
- Sound. No asset exists; a shot should go through the existing audio path once one does.

## Decisions

**The gun is a field on the fighter, not a state.** `Weapon uint8` (0 = none, 1 = pistol) stays set through every state. The gun only changes behaviour inside `StateIdle` (and the Run→Idle hand-off), which is where standing and walking already live. A `StateGun` was considered and rejected: jump, roll, hit and death all end by entering Idle, and each of them would have to learn to return to a gun state instead. As a field, the gun comes back for free when the fighter next stands.

**Standing and walking with the gun is `StateIdle` with a gun branch.** At the top of the Idle case, when `Weapon != 0`:
- skip `updateUpdown` and the updown, walk and block animations, and set the body to `gun-bot-idle.png` frame 0;
- LMB → `fire()`;
- jump, roll and climb down keep their existing checks;
- no attack, no block-all.

`updateHorSpeed` clamps speed to `RunReach` while `Weapon != 0`. Idle only enters Run when speed is above `RunReach`, so the fighter never runs. Drawing the gun from Run enters Idle. Drawing it while `Updown == 3` sets `Updown = 1`, so `canMove` (which refuses movement at `Updown == 3`) does not root the gunner. A guard check `&& actor.Weapon == 0` in the melee `guarding` expression makes the gunner unguarded.

**The toggle is an edge-triggered `Input.Gun`**, latched from Q in `latchInput` like the other presses, sent in button bit 7. It is acted on only in Idle and Run while on the ground. It is not added to the hit-stop buffer (`BufAttack`/`BufJump`): a toggle lost to hit-stop is harmless, and the record has no free flag bit to carry a third buffered press.

**Aim is a byte, `Input.Aim`: 0 = `AimUpDeg`, 255 = `AimDownDeg`.** `sampleInput` works it out from the shoulder point to the mouse, like `rollAim`: the vertical angle from horizontal toward the facing side, clamped to the sheet's range. The fighter copies it into `a.Aim` each Idle tick while armed, so the snapshot and the drawing read the fighter rather than the command. The frame is the nearest entry in a table of per-frame angles. The shot direction is the exact byte angle, not the snapped frame angle, so aim stays precise while the art steps in fives. Starting table, measured from the sheet and tuned by eye:

| frame | angle (° below horizontal) | muzzle (x, y) facing right |
|---|---|---|
| 0 | -55 | (38, 5) |
| 1 | -20 | (41, 11) |
| 2 | 0 | (40, 15) |
| 3 | 30 | (38, 23) |
| 4 | 55 | (37, 27) |

`AimUpDeg`/`AimDownDeg` are frame 0's and frame 4's angles. Facing left mirrors x as `47 - x`, since the sheet is flipped by a negative source width.

**Ray cast: one function for sim and view.** `shotEnd(a, w) (end rl.Vector2, hit *Actor)` casts from the muzzle along the aim for `ShotRange` (240). It uses a slab ray-vs-rectangle test against `LE.solids` minus drop-through platforms (`dropThrough(i)`), and against the collision box of every actor that is not the shooter, not dead and not an ally. The nearest entry point wins. The authority calls it in `fire()` and applies the hit. Every machine calls it again when presenting a shot, to find the tracer's end. The two can differ by lag compensation or interpolation, but the tracer is cosmetic and the hit is authoritative. Replicating the end point was the alternative, but it costs 8 bytes per fighter per snapshot to fix a sub-frame visual mismatch.

**`fire()`**: requires `ShotCooldown == 0`. It sets `ShotCooldown = ShotCooldownTicks` (15), increments `Shots`, and on the authority resolves the hit:
- a target in `StateAttacked` or `StateClash` stops the ray and takes nothing;
- a target in `StateBlock` gets `exitBlockState()` and then `takeHit(a)`;
- any other target gets `takeHit(a)`.

`takeHit` already costs structure, flinches, stuns, kills, credits and alerts bots, and its hit-stop freezing the shooter too reads fine as recoil. `ShotCooldown` counts down every live tick in `Update`.

**Shots are seen through the replicated `Shots` counter, not an event.** Every machine keeps a presentation-only `SeenShots` on the actor. In `presentEffects`, when `a.Shots != a.SeenShots`, it spawns a tracer and a flash at the actor's current muzzle and aim, then sets `SeenShots = Shots`. This one path covers the local predictor, the listen server, offline play, remote fighters and spectators. A client replay re-derives the same `Shots` value, so nothing plays twice, and no extra event bit is needed. `SeenShots` is set to `Shots` when an actor first appears, so joining mid-match does not replay old shots. It is a `uint8` and wraps, which is fine because shots are 15 ticks apart and snapshots 3 ticks apart.

**Tracer and flash are package-level presentation lists in `gun.go`**, aged by wall-clock `dt` like blood droplets. Neither is simulation state, so time manipulation and rewind do not touch them.
- Flashes are `LevelLight` values: radius about 80, warm white, intensity 1, lifetime `FlashSec` (0.05). They are passed as `append(slices.Clone(LE.lights), flashLights()...)` to `computeLights`, so they cast shadows through the existing Zima pass. They are not passed to the rim light, where a flash would be too brief to read.
- Tracers are drawn with `rl.DrawLineV` after `compositeLights()`, beside the rim pass, so the ambient does not darken them. They are 1 px, near-white yellow, with lifetime `TracerSec` (0.06). The palette pass quantizes the colour like everything else.

**Smoke and dust are more of the same presentation lists.** When a shot is shown, a `puff` list gets 5 smoke puffs at the muzzle and, if `shotEnd` stopped at a solid rather than a fighter or the range, 3 dust puffs at the end point. Each puff has a position, a velocity, a lifetime and a start and end radius. It slows under drag, rises with a constant lift and grows as it ages. It is drawn as a pixel disc that loses pixels in a fixed pattern as it ages: whole, then a checkerboard, then one pixel in four. That keeps the fade crisp under the palette pass, where transparency would band. The puffs are drawn inside the lit scene beside the blood droplets, so the flash lights them and the ambient darkens them. The alternative was to draw them over the lit scene like the tracer, but they would glow in the dark, and Animal Well's smoke works because it doesn't.

**Gun feel: recoil, fire, casings, sparks, holes.**
- *Recoil comes from `ShotCooldown`*, which is already replicated. While `ShotCooldown > ShotCooldownTicks - RecoilTicks` (6 ticks), `topDraw` draws `max(aimFrame-1, 0)` and shifts the torso a pixel back. So every machine shows it and needs no new state or message. The camera kick is local: a decaying offset added to `Cam.Offset` in `presentEffects` when the local fighter's shot is shown.
- *No shooter hit-stop*: `fire()` clears the shooter's `FreezeTicks` after `takeHit`, which sets it for both sides. Rewriting `hitFeedback` for this one caller was the alternative, but it is a bigger diff for the same result.
- *The muzzle fire* rides on the flash entry: the flash keeps the shot direction, and `drawTracers` draws a short star of pixels along it.
- *Casings and sparks* are one more presentation particle, `bit`. It has a position, a velocity, gravity, a bounce against `pixelSolid(LE.solids, …)` and a lifetime. Casings are brass, bounce and rest. Sparks are bright, fast and short-lived, and drawn over the lit scene like the tracer. Casings are drawn inside it.
- *Holes* are points just inside the wall along the shot. They are kept in a list capped at 64 (the oldest is dropped) for 20 s, and drawn as a dark pixel inside the lit scene. `clearBlood`'s call sites also call `clearShots`, which empties every gun presentation list.

**Replication.**
- Command: bit 7 carries `Gun`, and one byte is added after `RollAim` for `Aim`.
- Actor record: 4 bytes are added (`Weapon`, `Aim`, `Shots`, `ShotCooldown`), taking `actorRecSize` from 47 to 51. 16 × 51 is well inside 1200 bytes.
- `ProtocolVersion` becomes 13.
- `decodeCmds` needs no range check on `Aim`, since every byte is a valid aim. `apply` must reject an unknown `Weapon` by clamping it to 0.

**Lag compensation.** `runCmd`'s condition becomes "attacking, or (armed and `in.Attack`)". The rewind only moves other fighters' `Pos`, and `shotEnd` reads positions through `collisionRect()`, so no other lag-comp code changes.

**Sheets.** `gun-bot-idle.png` and `pistol-top.png` are appended to the end of `animTable`, which is append-only. That registers them as textures, and gives the legs a network id because they are the body `Anim`. The torso is chosen by `topAnim`: `Weapon == 1` and `Anim == gun-bot-idle` → `pistol-top.png`. `topDraw` then replaces `src.X` with `aimFrame(a.Aim) * 48` for the gun layer. Because the legs choice is just the body `Anim` set in the Idle gun branch, walk and run legs later become a `SetAnim` choice there.

## Risks / Trade-offs

- [The tracer end on a client can differ from where the server's shot stopped] → It is cosmetic and on screen for 0.06 s. The hit, structure and death that follow are authoritative and replicated.
- [The slab ray over every solid each shot] → A shot is at most every 15 ticks per fighter, and `LE.solids` is the same list `blockSides` already scans every tick. `ponytail:` a broad-phase is only needed if levels grow very large.
- [The gunner is unguarded, so a melee fighter up close wins] → Intended as the trade for range. It is tuned by the cooldown and walking pace, not by code.
- [Frame angles and muzzle points are estimates from the sheet] → They are one table in `gun.go`, tuned by eye in play, and a debug-mode muzzle dot makes the tuning quick.
- [`Input.Gun` is not buffered through hit-stop] → A Q pressed during the 6 frozen ticks is dropped and the player presses again. It could be carried later if the record gains a free bit.

## Migration Plan

This is a protocol version bump. Old clients are rejected at hello with the existing version-mismatch reason. Level files are untouched.
