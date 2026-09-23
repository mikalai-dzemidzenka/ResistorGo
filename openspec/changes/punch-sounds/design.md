## Context

See proposal.md for why. `blood.go`'s `watch` runs once per rendered frame,
compares each fighter's structure, state and guard with the last look, and
`bled` returns how many droplets a landed hit spills - zero for everything
that is not one. It already runs on every machine that renders a match and is
skipped while a recorded frame is shown. `audio.go` holds the music tracks.

## Decisions

### The blood's detection is the trigger

`watch` calls `punchHeard()` wherever `bled` returns more than zero. Hits and
blood are the same event; a second detector would only drift from the first.
`punchHeard` does nothing when the audio device is not open (tests, the
dedicated server).

### A block is a count that changed

`confirmBlock`, the one place a blow is stopped - by a guard or by block-all -
runs on the authority only, and a stopped blow leaves nothing reliable in what
the snapshot carries: a guarded block pays structure back only below the top,
and the block animation loops while the swing lasts. So a fighter counts the
blows it stops in `Blocks`, and `watch` plays `blockHeard()` when the count
differs from its last look. The count is copied with the fighter, so a
restore reseeds it silently like the structure. The record carries it as one
bit - odd or even - in the last free flag bit; `apply` bumps the count when the
bit disagrees. Blocks are a swing apart, far longer than snapshots, so a flip
is never missed.

### Overlap with aliases

`rl.PlaySound` restarts a sound that is still playing, so one `rl.Sound` per
file would cut a flurry short. Each punch is loaded once from its embedded
bytes and given `SoundVoices` (3) aliases with `LoadSoundAlias`, sharing the
sample data; a punch plays on the next alias of the chosen file, round robin.
The block sound is loaded and played the same way.

### Its own random source

Which punch plays uses the audio's own `rand.Rand`, as the blood does, so
hearing hits cannot change what a bot decides.

### Settings

`SoundVolume` is a new field, `json:"soundVolume"`, default 80, clamped to
0..100. `LoadSettings` unmarshals over the defaults, so a file without the
field keeps 80. Options gets the slider as row 3; HUD moves to row 4 and Back
to 5. The Host screen already uses six rows, so the panel holds it.

## Risks / Trade-offs

- [Two sounds only] Every hit sounds like one of two samples. Fine to start;
  more files can join the list.
- [Volume at startup] `SetSoundVolume` is applied on every play, so the
  slider takes effect on the next hit without any apply step.
