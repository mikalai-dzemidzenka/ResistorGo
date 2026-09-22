## Context

See proposal.md for why. `audio.go` holds one track: the menu's, streamed from
embedded bytes pinned for the life of the stream, faded by `musicFade`
toward 1 on the menu and 0 elsewhere over `musicFadeSec`, paused at silence
and resumed when it rises. `updateAudio` runs every frame in every stage.

## Goals / Non-Goals

**Goals:**
- Two tracks with one fade rule, decided by a function that needs no audio
  device to test.

**Non-Goals:**
- Per-level tracks, playlists, a separate level-music volume.
- Slowing or reversing the music with time manipulation.

## Decisions

### A track type, two of them

```
  type musicTrack struct {
      music  rl.Music
      data   []byte          // pinned: the decoder streams from it
      pin    runtime.Pinner
      ready  bool
      fade   float32         // 0..1, times the volume setting
      rewind bool            // stop at silence, so it plays from the top next time
  }
  menuTrack  = {path: Neon Sprint,     rewind: false}   pauses at silence, as now
  levelTrack = {path: Final Run (1),   rewind: true}    stops at silence
```

Loading, the fade step and the play/pause/stop decisions move onto the type;
the menu track behaves exactly as before.

### One pure function says where each track is heading

```
  musicTargets(stage, held) (menu, level float32)
    stage == main menu -> 1, 0
    stage == level     -> 0, held ? LevelMusicDuck : 1
    otherwise          -> 0, 0
```

`held` is `matchHeld()`: time manipulation on or edit mode. `LevelMusicDuck`
is 0.3. A duck is only a lower target: the track keeps playing and being fed,
so it never restarts for a hold.

### Starting over

A track with `rewind` set is stopped (`StopMusicStream`, which rewinds it)
when its fade reaches 0, and started (`PlayMusicStream`) when its target rises
above 0 while it is stopped. So leaving the level and coming back plays the
track from its beginning, while a restart with R - which never leaves the
level stage - does not touch it.

## Risks / Trade-offs

- [Exe size] +3.3 MB for the track. → Accepted; the spare tracks stay out of
  the build as before.
- [Two streams fed at once during a crossfade] Both are updated for the 0.6 s
  of the fade; raylib streams are cheap and it was already one.
- [The written requirement] "Music stops during a match" in the unarchived
  `add-menu-music-and-options` now describes old behaviour. → A task there
  and here: archive that change first, then turn this delta into a MODIFIED
  of that requirement.
