## Context

`Cursor.Update` runs once per rendered frame from `runLevel` and moves the attack crosshair with `Vector2Lerp(AtkPos, Pos, 0.5)`, ignoring `dt`. `sampleInput` reads `Cur.AtkPos` into `Input.Updown` when a fixed tick builds the local command, so the crosshair's position is the guard and attack height that gets predicted and sent. `DrawGUI` draws both crosses after the palette pass, so their colours reach the screen unquantized. The local fighter is `CL.Local` (nil when dead or not spawned) in single player and online alike; its `Energy` is the predicted value, corrected by the server. See proposal.md for why.

## Goals / Non-Goals

**Goals:**
- One energy fraction, taken once per frame, drives both the colour and the follow speed.
- The colour and speed rules are pure functions, testable without a window.
- At 60 fps and a full pool the crosshair moves exactly as it does today.

**Non-Goals:**
- Moving the crosshair into the fixed step. It is input sampling, and it stays per frame; only its dependence on frame time is removed.
- Easing the displayed energy. A swing is 12 of 300 points (4 %) and regeneration arrives a point at a time, so following the raw value already reads as gradual.

## Decisions

**Exponential follow, `t = 1 - exp(-k·dt)`.** This is the frame-rate independent form of "close a fixed share of the gap". With `k = 60·ln 2 ≈ 41.6 /s`, `t` is exactly 0.5 at `dt = 1/60`, so a full pool reproduces today's motion at the target frame rate. Alternative: a constant speed in pixels per second (`moveTo`) — rejected, because it changes how a large mouse jump feels even at full energy, and the lerp's quick start followed by a soft landing is the feel being kept.

**`k` linear in the energy fraction**, from `CursorFollowTired = 8 /s` at empty to `CursorFollowFull = 60·ln 2` at full, the fraction clamped to 0–1. Settling time (to 95 %) is `3/k`: about 72 ms full, 120 ms at half, 375 ms empty. The slowdown is therefore mild through most of the pool and bites as it empties, which matches when the player should start worrying. Alternative: interpolate the settling time linearly instead — that makes half a pool already 225 ms, which punishes ordinary play. Both constants are named so tuning is a one-line change.

**Colour as a straight RGB blend**, `rl.Green` (0, 228, 48) at full to (170, 200, 175) at empty, alpha 255. The two ends are close in hue, so a per-channel lerp has no muddy midpoint and needs no HSV round trip. Opaque, because a crosshair faded towards transparency would hide the guard height from a tired player exactly when it matters.

**The cursor takes a fraction, not an actor.** `Cursor.Update(dt, energy float32)` and `DrawGUI` use a stored fraction; `runLevel` computes it from `w.Local` (1 when nil) and passes it in. This keeps `cursor.go` free of the actor and the world, and lets the two pure helpers — `attackFollow(energy, dt) float32` and `crosshairColor(energy) rl.Color` — be unit-tested. `CursorAttackSpeed` and the `AtkSpeed` field go away, since nothing else reads them.

**The requirement is added, not modified.** `Energy pool`'s "An empty fighter can still defend" scenario is already being rewritten by the in-flight `add-run-jump-and-fall` change. A second MODIFIED block on the same requirement would conflict at archive, so the new requirement states the relationship itself: blocking stays available at any energy, reaching the height takes longer.

## Risks / Trade-offs

- [A tired fighter defends worse, not only attacks worse] → Deliberate (chosen during exploration); spelled out in the spec so it is not read as a regression.
- [Bots are not slowed, so a tired player faces a bot that changes guard instantly] → Accepted for now; bots already pick heights at random. A follow-up can add an energy-dependent delay in `botInput` if play shows it matters.
- [A server correction to the local energy moves the colour and speed a step] → Corrections are small and rare; both effects are continuous in energy, so the step is as small as the correction.
- [A long frame (a hitch) makes `t` close to 1 and the crosshair snaps] → Same as today at low frame rates, and correct for a time-based follow.
- [Hit-stop and pauses freeze the simulation but not the crosshair] → Unchanged from today: the crosshair is input and keeps following in real time.
- [The in-flight energy spec still says the pool is 200 while the code has 300] → The new requirement speaks in fractions of the pool, not points, so it holds either way.
