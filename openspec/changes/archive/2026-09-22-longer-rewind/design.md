## Context

See proposal.md for why. The recording is a ring of `RewindFrames` world frames
in `rewind.go`; `captureFrame` fills the slot at `rewindHead` once per fixed
step by truncating and appending into that slot's own slices, and `ResetTime`
zeroes the ring. Nothing about the ring depends on its length beyond the
constant.

Scrubbing lives in `UpdateTime`: `scrubDirection(dx)` turns one rendered
frame's mouse movement into -1, 0 or +1; a zero stops time at once
(`TimeMode = TimePaused`, `scrubAcc = 0`); otherwise `dt` is banked in
`scrubAcc` and every whole `FixedTimeStep` in it seeks `ScrubDir *
ScrubFrames` frames. Counting in fixed quanta is what makes the rate
independent of the frame rate, and the ramp keeps that.

## Goals / Non-Goals

**Goals:**
- A minute of recording with no change to what a frame is or how it is
  captured and restored.
- A ramp that depends only on how long the hand has kept moving one way,
  counted in fixed steps, so two frame rates cover the same frames.

**Non-Goals:**
- Compression, delta frames, keyframes, a serializer, a new dependency.
- A recording that grows with the match.
- Any change to what a restored frame contains.

## Decisions

### Raise the constant; keep the ring as it is

`RewindFrames = 60 * 60`. The ring is an array of `worldFrame` headers
(three slice headers, a pointer and two integers: about 100 bytes), so 3600
of them are about 0.4 MB, allocated with the package. Each slot's `actors`
slice is allocated on the first lap and reused after, as now: about 2 KB a
slot for the player and three bots, 7 MB for the full minute. `ResetTime`
already drops all of it.

Alternative considered: compress frames (zstd, or Braid's delta against the
previous frame with a full keyframe every N). A frame keeps the fighter
objects themselves and a value copy with live pointers (`W`, `Attacker`,
`PushStuck`, `Engaged`, `Bot.target`, `AnimTop`). Bytes a compressor holds
are invisible to the garbage collector, so the pointers would have to become
IDs on capture and be relinked on restore, by a hand-written serializer that
every new `Actor` field must join or be silently lost. At 7 MB there is
nothing to win that is worth that. If the window grows to several minutes or
the whole match, the route is delta-encode then compress, and the
serializer is the price.

### The ramp counts fixed steps held in one direction

```
  state:  scrubHeld  int      fixed steps scrubbed in the current direction
          scrubStill float32  seconds the mouse has reported no movement
          scrubLast  int      the direction the ramp belongs to (-1 / +1)

  each rendered frame:
    dir := scrubDirection(dx)
    dir == 0:  pause (as now); scrubStill += dt
               if scrubStill >= ScrubRampGrace: scrubHeld = 0
    dir != 0:  scrubStill = 0
               if dir != scrubLast: scrubHeld = 0; scrubLast = dir
               bank dt; for each whole FixedTimeStep:
                   rewindSeek(dir * scrubRate(scrubHeld)); scrubHeld++

  scrubRate(held) = ScrubFrames
                  + (ScrubFramesMax - ScrubFrames) * min(held, ScrubRampSteps)
                    / ScrubRampSteps            (integer arithmetic)

  ScrubFrames = 2, ScrubFramesMax = 16, ScrubRampSteps = 120 (2 s),
  ScrubRampGrace = 0.1 s
```

Counting quanta, not seconds of `dt`, means the rate at the Nth quantum is
the same whatever frame rate delivered it, so the frame-rate scenario holds
exactly, not approximately. Integer steps of the rate keep the frames covered
exact for tests.

The grace is measured in rendered-frame `dt`, since it is about the mouse and
not the match: a mouse polled at 125 Hz on a 144 Hz display reports zero on
some frames while the hand is moving, and without the grace every such frame
would drop the rate back to 2x. The pause itself stays immediate, as the spec
has always required.

Entering or leaving the mode and `ResetTime` clear all three.

Alternative considered: rate from mouse speed, as a jog wheel. The spec's
rule that hand speed does not matter was deliberate - speed would make
jitter and a light touch move time unpredictably - and time held is a
calmer signal that gets the same result: long drags go far, short ones stay
precise.

### What reaches the playhead is unchanged

`rewindSeek` already clamps to the recording; a 16-frame step at the oldest
frame or the live edge stops there, as a 2-frame one does. Restore is still
called once per rendered frame whose playhead moved, so a faster scrub costs
one restore a frame, not one per frame skipped.

## Risks / Trade-offs

- [7 MB held per match] → The minute is held for the life of the match and
  keeps removed fighters alive with it; fine for single player, and
  `ResetTime` releases it on restart and on leaving the level.
- [First lap allocates] → Each slot allocates its `actors` slice once, 3600
  small allocations spread over the first minute; steady play after that
  allocates nothing, which the existing allocation test checks once it has
  gone round the larger ring.
- [16x skips frames] → At full speed the playhead lands on every 16th frame;
  releasing the mouse and dragging again starts at 2x from where it stopped,
  which is how a single frame is found.
- [Tests that record the whole ring get slower] → They record 3600 frames of
  an empty or tiny world; milliseconds.
