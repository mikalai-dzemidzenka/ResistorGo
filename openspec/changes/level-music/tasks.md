## 1. Two tracks

- [x] 1.1 Embed `resources/music/Final Run (1).mp3` in `assets.go`; verify
      `asset` returns it in a test
- [x] 1.2 Move the menu track's loading, fade and play/pause onto a
      `musicTrack` type and add the level track with `rewind` set; verify the
      game builds and `go vet` is clean
- [x] 1.3 Add `musicTargets(stage, held)` and `LevelMusicDuck = 0.3`, and drive
      both tracks' fades from it in `updateAudio`; verify with a table test:
      menu -> (1, 0), level -> (0, 1), level held -> (0, 0.3), anything else
      -> (0, 0)
- [x] 1.4 Verify with a test of the track's decision function (not the device)
      that a rewinding track stops at silence and starts when its target
      rises, and a non-rewinding one pauses and resumes

## 2. Records

- [x] 2.1 Add a task to `add-menu-music-and-options`: its "Music stops during
      a match" is replaced by `level-music`, which must be archived after it
- [ ] 2.2 When archiving: archive `add-menu-music-and-options` first, then turn
      this delta into a MODIFIED of "Music stops during a match" and sync it
- [x] 2.3 Run `go test ./...` and `go test -tags release ./...`; build with
      `build.ps1`. `TestOnlyTheMenuTrackIsEmbedded` pinned the menu
      track as the only one embedded; it now allows the menu and level tracks
      and is renamed `TestOnlyTheTracksPlayedAreEmbedded`. The packed exe is
      8.5 MB, up from 5.3

## 3. In the game

- [x] 3.1 Start a single-player match: the menu music fades out and Final Run
      fades in; leave and start again: it starts over; the menu music returns
      in between
- [x] 3.2 Press T and scrub, then T again: the music ducks and comes back
      without restarting; press R: it carries on
- [x] 3.3 Set the music volume to 0 in Options: silence in the match too
