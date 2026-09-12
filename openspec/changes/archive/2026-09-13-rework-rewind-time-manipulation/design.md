## Context

See proposal.md — Why. What matters for the approach is how rewinding is wired today:

- `rewind.go` holds three globals: `IsRewinding`, a free-running `RewindIdx` and `RewindLeft`. `UpdateRewind` is called once per **rendered** frame from `runLevel`, and only when `NetMode == NetOffline`.
- Recording happens somewhere else and at a different rate: `Actor.updateRewind`, called from `Actor.Update`, runs once per **fixed step** (60 Hz). It writes `RewindData[RewindIdx%RewindFrames]` on a live step and reads it back on a rewinding one. `RewindFrames` is 256, so the history is 256 fixed steps ≈ 4.3 seconds.
- Those two rates only agree when the game runs at exactly 60 FPS. At the ~150 FPS the game actually reaches, the index advances about 150 times a second while frames are recorded 60 times a second, so "hold R for one second" does not rewind one second.
- `RewindIdx` is also unbounded in both directions. It decrements freely, and Go's `%` keeps the sign, so a long enough rewind indexes the array with a negative number.
- Fighter simulation is driven by `stepOffline`, which drains a `dt` accumulator in fixed steps. `Actor.Update` is the only place gameplay advances; input is sampled inside that loop (`sampleInput`), not per frame.
- `FreezeTicks` already exists as a per-actor freeze, but it is hit-stop: it is replicated, decremented per actor, and does not stop bots deciding or the world ticking.

## Goals / Non-Goals

**Goals:**

- One place that owns "what moment is on screen" and enforces both ends of the recorded history, so the spec's two limits are comparisons rather than emergent behavior.
- Scrub speed that does not depend on the frame rate.
- A pause that stops the whole simulation without touching the fighter state machine or the netcode.
- Stay in the project's style: package-level state in the existing files, no new abstraction layers, no dependency injection.

**Non-Goals:**

- Changing what a recorded frame holds, or how long the history is (proposal — Non-goals).
- Any change to the network path. `net_lagcomp.go` rewinds server-side for hit tests; it is a separate mechanism and is not touched.
- A HUD for the mode beyond the existing debug text.

## Decisions

### Replace the free-running index with an explicit ring window

Track three values in `rewind.go`: `head` (the newest recorded frame), `count` (how many frames of history are valid, capped at `RewindFrames`) and `cursor` (the frame currently on screen). The invariant is `head-count < cursor <= head`, and every ring access goes through one helper that maps an absolute frame number onto `[0, RewindFrames)` with a non-negative modulo.

The spec's two limits fall out directly: rewinding clamps `cursor` at `head-count+1`, forward scrubbing clamps it at `head`. "Leaving rewrites the future" becomes `head = cursor; count = cursor - oldest + 1` at the moment the mode is left.

*Alternative — keep `RewindIdx`/`RewindLeft`:* rejected. The pair cannot express the live edge independently of the playhead, which is exactly what forward scrubbing needs, and `RewindLeft` doubles as both "how much history exists" and "how far back we are".

### Drive scrubbing from the fixed-step accumulator, not from rendered frames

Scrub steps are consumed from the same accumulator `stepOffline` uses, in the same 1/60 s quanta. The rate constant is then expressed in recorded frames per fixed step (default 2), and the spec's "speed of the hand does not matter" and frame-rate independence both hold by construction.

Mouse delta, though, arrives per rendered frame. So each rendered frame latches a direction from the delta, and the accumulator loop applies that direction to as many scrub steps as it has time for. At 150 FPS most frames contribute no scrub step and some contribute one; over a second the count is the same 60 × rate either way.

*Alternative — step the cursor once per rendered frame (today's behavior):* rejected. It is the bug described in Context.

### Mode as a state, not a bool

`IsRewinding` becomes a three-state mode — off, on-and-paused, on-and-scrubbing — plus a direction. `Actor.updateRewind` keys off it: record on a live step, restore on a scrub step, and do not run at all while paused (a paused frame must neither overwrite history nor re-read it). Keeping the state in package-level vars in `rewind.go` follows the project's existing convention.

Call sites that only ask "is the simulation being replayed?" (`Actor.Update`'s early return, `Actor.Draw`'s rewind branch) keep working against a small helper rather than the raw state.

### Pause by not stepping, and by not accumulating

While the mode is on, `runLevel` skips `stepOffline` entirely and does not add `dt` to the accumulator. Skipping the call alone would satisfy "nothing moves", but the accumulator would keep growing and leaving the mode after ten seconds would fire hundreds of catch-up steps — the second scenario under "Entering pauses the match" exists to pin that down.

*Alternative — freeze each actor with `FreezeTicks`:* rejected. It is hit-stop's mechanism, it is replicated over the network, and it would still run the world tick, bot decisions and the accumulator.

### Input suppression comes for free

Fighter input is only ever read inside the fixed step, so skipping `stepOffline` means `sampleInput` is never called while the mode is on: nothing is applied and nothing is buffered for later. No separate gate is needed, and the spec's two suppression scenarios hold as long as no gameplay input is read per frame. `Cur.Update` keeps running per frame, so the crosshair still follows the mouse, which the spec permits.

### Mouse direction from the frame delta, with a dead zone

Use the per-frame mouse delta's X component, as `spectator.go` already does for panning (the cursor is captured in a level, so an absolute position is not meaningful). Direction is its sign; magnitude is deliberately discarded. A small dead zone on `|dx|` keeps sensor noise from creeping time forward, per the jitter scenario.

## Risks / Trade-offs

- **A recorded frame is partial — position, speed, orientation, animation frame — but not state, stance or hit-stop.** Resuming from a scrubbed frame can therefore leave a fighter with a restored pose and a stale state machine (mid-attack state showing an idle frame). → Pre-existing today and explicitly out of scope; the spec says only what is restored. If it shows up in play, widen `ActorRewindFrame` in a follow-up change rather than here.
- **Rate of 2× is a guess until it is felt.** → It is one constant, symmetric in both directions; tune it once in the game.
- **Discarding the future on resume is destructive and has no undo.** → It is what "resume from here" means; the alternative (keeping a branch) is a much bigger feature. The spec pins the behavior so it is at least predictable.
- **The debug overlay prints `RewindIdx` and `RewindLeft`.** → Those fields are going away; the overlay moves to the new values in the same change, or it will not build.
- **The mode is toggled, so it is possible to leave the game paused and wonder why nothing responds.** → The debug overlay already shows the rewind state; keeping a visible indication of the mode is worth doing even without a full HUD.

## Migration Plan

Nothing persists and nothing crosses the network, so there is no data migration and no compatibility window: the change lands in one commit and reverts cleanly. The only player-visible break is the control itself — R is a toggle now, not a hold — which the proposal marks BREAKING and which should be reflected wherever the controls are listed.

## Open Questions

- Whether the mode deserves an on-screen indicator outside debug mode (a "PAUSED / REWIND" marker). Deferrable: it changes no requirement and no task beyond adding one draw call.
- Whether the scrub rate should differ between directions once it is played (the spec fixes them as identical; changing that would need a spec revision).
