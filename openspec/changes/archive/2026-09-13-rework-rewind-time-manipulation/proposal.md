## Why

Rewinding today is a single gesture: hold R and the last few seconds play backwards, let go and play resumes. There is no way to stop on a frame, look at it, or move forward again — the moment you stop rewinding the match is live. That makes the feature useless for the thing it should be good at: finding an exact moment and picking it back up from there. The controls also fight the rest of the game, since R both enters and drives the rewind, leaving nothing to scrub with.

## What Changes

- **BREAKING** Hold-R-to-rewind is replaced by a time-manipulation mode. R is now a toggle: press it to enter, press it again to leave. Holding R does nothing on its own.
- **Entering pauses the match.** The instant R is pressed the simulation stops: no physics, no animation, no bots, no attack resolution. Nothing moves until the player scrubs.
- **The mouse is the scrub control.** While the mode is on, horizontal mouse movement drives time: moving right runs time forward, moving left rewinds. Stopping the mouse pauses again. Direction alone decides the direction of travel — the rate is fixed, so a slow drag and a fast sweep scrub at the same speed.
- **Forward scrubbing is replay, not simulation.** It replays recorded frames and stops at the newest one (the live edge), where the match stays paused. Time never runs past what was recorded.
- **Leaving resumes from where you stopped.** Pressing R again continues the match from the frame on screen. Frames that were ahead of that point are discarded — the match's future is rewritten from there, and the next time-manipulation session can only scrub back into what has been recorded since.
- **Rewinding stops at the oldest recorded frame** instead of running off the end of the buffer.
- Gameplay input is ignored while the mode is on: moving, jumping, attacking and stance changes do nothing until it is left.
- Unchanged: the mode is single-player only (it is not offered in a network match), the recording is still the existing 256-frame per-actor ring buffer, and a recorded frame still holds position, speed, orientation and the current animation frame.

### Assumptions

- Forward and backward scrubbing share one fixed rate, expressed as recorded frames per fixed step and defaulting to 2 (twice real time in both directions). It is a single constant to tune once it is felt in the game.
- "Mouse is moving" means the pointer moved horizontally at all since the previous frame. A small dead zone keeps sensor jitter from creeping time forward while the player's hand is still.
- The crosshair keeps following the mouse while scrubbing; only the effect of the mouse on the fighter is suppressed.

## Capabilities

### New Capabilities
- `time-manipulation`: entering and leaving the mode, the pause it imposes, mouse-driven scrubbing in both directions, the limits at the oldest recorded frame and at the live edge, what resuming does to recorded history, and which inputs are suppressed while the mode is on.

### Modified Capabilities

None. Rewinding has no spec today; this change specifies the capability for the first time and replaces the old hold-R behavior as part of it.

## Impact

- **Code**
  - `rewind.go`: `UpdateRewind` is rewritten around a mode flag and a scrub direction. `IsRewinding` becomes one of several states (paused, rewinding, replaying forward); the live edge and the oldest recorded frame become explicit bounds rather than a free-running index.
  - `game.go`: `runLevel` must skip the fixed-step simulation (`stepOffline`) while the mode is on, and must not let the accumulator bank up time during the pause, or leaving the mode would fire a burst of catch-up ticks.
  - `actor.go`: `Actor.Update` / `updateRewind` follow the new state — record on a live step, restore on a scrub step, do nothing while paused.
- **Controls**: R changes from hold to toggle. No other key changes.
- **Not affected**: network play (the mode stays offline-only, and the server's own rewind for lag compensation is a separate mechanism), the level editor, the fighter state machine and combat rules, rendering.

## Non-goals

- Extending how much history is kept, or what a recorded frame contains. A rewind still restores position, speed, orientation and animation frame only; state, stance and hit-stop are not restored, exactly as today.
- Simulating past the live edge, at any speed.
- Time manipulation in a network match.
- A scrub bar, timeline or any other HUD for the mode beyond what debug already draws.
