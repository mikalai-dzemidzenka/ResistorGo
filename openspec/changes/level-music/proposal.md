## Why

The level plays in silence: the menu's track fades out as a match starts and
nothing takes its place. `Final Run (1).mp3` is already in `resources/music`
for the level, and a fight needs a beat under it.

This reverses a written requirement. `add-menu-music-and-options`, not yet
archived, says "Music stops during a match: the menu music SHALL fade out and
stop when a match starts, leaving the level silent." The menu music still
fades out; the level is no longer silent.

## What Changes

- `Final Run (1).mp3` is embedded and plays, looping, whenever a match is on
  (level_1, the only level): single player, hosting and joining alike. It
  starts from the top each time a match starts and fades in as the menu track
  fades out; on returning to the menu it fades out and the menu track fades
  back in.
- While the match is held - time manipulation on, or edit mode - the level
  music ducks to a low volume, and comes back up when the match resumes. It
  never pauses or restarts for a hold, a rewind or a restart with R.
- It follows the existing music volume setting, as the menu track does.
- The exe grows by the track: the packed build goes from 5.3 MB to 8.5 MB.

## Capabilities

### New Capabilities

<!-- none -->

### Modified Capabilities
- `audio`: a level track plays during a match, instead of the level being
  silent. The capability is introduced by the unarchived
  `add-menu-music-and-options`, so this delta is written as an ADDED
  requirement and carries a task to become a MODIFIED one of "Music stops
  during a match" once that change is archived.

## Impact

`assets.go` embeds the track. `audio.go` goes from one track to two: each
keeps its stream, its pinned bytes and its own fade; the menu track pauses at
silence as now, and the level track stops at silence so the next match starts
it from the top. The volume each track fades toward is worked out by one pure
function of the stage and whether the match is held, so it can be tested
without an audio device. No change to the simulation, the network or the
dedicated server, which opens no audio.
