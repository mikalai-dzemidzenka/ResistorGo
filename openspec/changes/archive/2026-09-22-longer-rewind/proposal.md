## Why

Time manipulation reaches back 256 fixed steps, about 4.3 seconds: enough to
undo a single exchange, not enough to go back to before a fight went wrong.
Now that a recorded frame is the whole match, going further back is worth a
lot more - and costs little: a frame of the single-player match (the player
and three bots) is about 2 KB, so a full minute is about 7 MB held as it is,
with no compression.

A longer recording brings its own problem. Scrubbing runs at a fixed 2x, so a
minute back would take thirty seconds of dragging the mouse.

## What Changes

- The recording holds the last 60 seconds of the single-player match (3600
  fixed steps) instead of the last 256.
- Scrubbing speeds up the longer the mouse keeps moving the same way: it
  starts at 2x, as now, and ramps up to 16x over two seconds of continuous
  movement in one direction. Stopping or reversing brings it back to 2x.
  How fast the hand moves still does not matter, only for how long it keeps
  moving; the rate is the same in both directions and does not depend on the
  frame rate.
- A mouse that reports no movement for a frame or two while the hand is still
  moving pauses time on those frames, as it does now, but does not reset the
  ramp; the ramp resets once the mouse has been still for a tenth of a
  second.
- Frames are not compressed. Compressing them (zstd, or Braid-style delta
  frames with keyframes) was considered and set aside: a frame holds the
  fighter objects and the pointers between them, which a byte compressor
  cannot hold without a serializer that turns every pointer into an ID and
  back - a serializer every new `Actor` field would have to be added to. At a
  minute it would save memory the game does not need saved. It is the route
  if the window ever grows to several minutes or the whole match.

## Capabilities

### New Capabilities

<!-- none -->

### Modified Capabilities
- `time-manipulation`: the recording reaches back 60 seconds, and the scrub
  rate ramps up with continuous movement instead of being fixed.

## Impact

`rewind.go`: `RewindFrames` becomes 3600; the scrub loop in `UpdateTime`
takes its frames per step from the ramp, and gains the ramp's state (steps
held in the current direction, and how long the mouse has been still), which
`ResetTime` clears. Memory: the ring's slot headers grow to about 0.4 MB, and
its recorded fighters to about 7 MB for a four-fighter match once the minute
has filled; a slot's storage is reused from lap to lap as now, so steady play
still allocates nothing.

Tests: `TestScrubRateIsFixed` and `TestScrubSpeedOfHandDoesNotMatter` describe
the fixed rate and are rewritten for the ramp; the recording tests are
written against `RewindFrames` and carry over. No new dependency.
