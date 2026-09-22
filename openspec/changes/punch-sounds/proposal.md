## Why

A landed blow flashes, shakes, spills blood and costs structure, and makes no
sound. `resources/sounds/punch1.mp3` and `punch2.mp3` were added for it.

## What Changes

- Every landed hit plays a punch sound, one of the two at random, on every
  machine that shows the match: single player, hosting, joining and
  spectating. Whiffs, blocks, block-all, clashes and shoves are silent.
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
- `audio`: a punch sound plays for every landed hit, at the sound volume.
- `settings`: the Options screen has a Sound volume slider, saved with the
  other settings.

## Impact

`assets.go` embeds `resources/sounds/*.mp3`. `audio.go` loads the two sounds
with a few aliases each so hits can overlap, and plays one at the sound
volume with its own random source. The trigger is the blood's hit detection
(`bled`), which already tells a landed hit from every other structure change
on every machine and is already silent while a recorded frame is shown.
`settings.go` gains `SoundVolume`; `main_menu.go` gains its row in Options,
moving HUD and Back down one. No change to the simulation or the network.
