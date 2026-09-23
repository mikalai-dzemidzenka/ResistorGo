## Why

A landed blow flashes, shakes, spills blood and costs structure, and makes no
sound. `resources/sounds/punch1.mp3` and `punch2.mp3` were added for it.
A blow stopped by a guard is silent too; `block1.mp3` was added for that.

## What Changes

- Every landed hit plays a punch sound, one of the two at random, on every
  machine that shows the match: single player, hosting, joining and
  spectating. Whiffs, blocks, block-all, clashes and shoves make no punch.
- Every blow stopped by a guard or by block-all plays the block sound, on
  every machine likewise. Hits, whiffs, clashes and shoves make no block.
- Hits close together overlap rather than cutting each other off.
- Nothing plays while time manipulation shows a recorded frame, so scrubbing
  across a hit is silent, as it sprays no blood.
- Options gains a Sound volume slider under Music volume (0 to 100, default
  80), saved in `settings.json`; a settings file from before it gets the
  default.

## Capabilities

### New Capabilities

<!-- none -->

### Modified Capabilities
- `audio`: a punch sound plays for every landed hit, and a block sound for
  every blow stopped, at the sound volume.
- `settings`: the Options screen has a Sound volume slider, saved with the
  other settings.

## Impact

`assets.go` embeds `resources/sounds/*.mp3`. `audio.go` loads the two sounds
with a few aliases each so hits can overlap, and plays one at the sound
volume with its own random source. The trigger is the blood's hit detection
(`bled`), which already tells a landed hit from every other structure change
on every machine and is already silent while a recorded frame is shown.
A fighter counts the blows it stops (`Blocks`); the snapshot carries its
parity in the last free flag bit, and the same look hears a change in it.
`settings.go` gains `SoundVolume`; `main_menu.go` gains its row in Options,
moving HUD and Back down one. Nothing in the simulation reads the count; the snapshot record keeps its size.
