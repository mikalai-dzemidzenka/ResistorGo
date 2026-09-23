## 1. Sounds

- [x] 1.1 Embed `resources/sounds/*.mp3`; verify both punches are in the
      binary with a test
- [x] 1.2 Load each punch with `SoundVoices` aliases in `initAudio`, unload
      them in `closeAudio`; add `punchHeard()` playing the next alias of a
      random punch at `Cfg.SoundVolume`, and doing nothing without an audio
      device; verify with a test that the choice alternates voices round robin
      and draws from the audio's own random source only
- [x] 1.3 Call `punchHeard()` from `blood.watch` for every landed hit; verify
      with a counting hook in a test that a flinch, a stun and a kill each
      sound once and a block, block-all, clash and shove do not

## 2. Settings

- [x] 2.1 Add `SoundVolume` (default 80, clamped) to `Settings`; verify the
      defaults, clamping, the round trip, and that a settings file without
      the field gives 80 with the rest as saved
- [x] 2.2 Add the Sound volume slider to Options as row 3, moving HUD and Back
      down; verify the menu layout tests pass and the rows fit

## 3. The whole

- [x] 3.1 Run `go test ./...` and `go test -tags release ./...`; build with
      `build.ps1`

## 4. In the game

- [x] 4.1 Land hits on a bot: each one punches; whiffs and blocks are silent
- [x] 4.2 Land a quick run of hits: each is heard
- [x] 4.3 Scrub across a hit with T: silence
- [x] 4.4 Set Sound volume to 0: silent hits, music unaffected; restart and
      it is still 0

## 5. Block sound

- [x] 5.1 Embed `resources/sounds/block1.mp3`, load and play it like the
      punches; verify it is in the binary with a test
- [x] 5.2 Count blows stopped in `Actor.Blocks` from `confirmBlock`; carry its
      parity in the snapshot's last flag bit; verify a block crosses the wire
      and a snapshot without one changes nothing
- [x] 5.3 Play `blockHeard()` from `blood.watch` when the count changed; verify
      a guarded block, a block at full structure and block-all are heard once
      and as no punch, a hit, a clash and a shove are not, and a reseed is
      silent
- [x] 5.4 Run `go test ./...` and `go test -tags release ./...`; build with
      `build.ps1`
- [x] 5.5 In the game: block a bot's swings high, low and with block-all -
      each is heard; as a client and a spectator likewise
