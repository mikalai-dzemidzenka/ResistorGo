## 1. Settings

- [x] 1.1 Add `settings.go` with `Settings{Fullscreen, WindowSize, MusicVolume}`, the `windowSizes` list (1280x720, 1600x900, 1920x1080), defaults (windowed, index 0, volume 70), and a package-level `Cfg` holding the current settings. Verify `go build ./...` succeeds.
- [x] 1.2 Add `LoadSettings()` and `SaveSettings()` for `settings.json`, with per-field fallback: a missing file, a parse error or an out-of-range value uses that field's default and never returns a fatal error. Verify `go build ./...` succeeds.
- [x] 1.3 Add `settings_test.go` covering the fallback rules: missing file gives defaults; malformed JSON gives defaults; out-of-range `WindowSize` and `MusicVolume` are clamped; a valid round trip through save and load preserves values. Verify `go test -run TestSettings .` passes.
- [x] 1.4 Add `settings.json` to `.gitignore` (creating the file if absent). Verify `git status --short` does not list `settings.json` after the game writes one.

## 2. Virtual screen and letterbox

- [x] 2.1 Add `present.go` with `VirtualScreen` (1280x720 point-filtered render texture, created once), `letterbox()` computing the centred destination rectangle and scale from the current window size, and `present()` drawing the virtual screen to the window over a black background with the Y-flipped source. Verify `go build ./...` succeeds.
- [x] 2.2 In `letterbox()`, apply `rl.SetMouseOffset(-dst.X, -dst.Y)` and `rl.SetMouseScale(1/scale, 1/scale)` so `GetMousePosition` returns virtual coordinates for raygui and the cursor. Verify `go build ./...` succeeds.
- [x] 2.3 Add `present_test.go` checking `letterbox` geometry: 1280x720 gives scale 1 and no bars; 2560x1440 gives scale 2 and no bars; 1920x1200 letterboxes top and bottom with the image centred and inside the window; a window narrower than 16:9 bars the sides. Verify `go test -run TestLetterbox .` passes.
- [x] 2.4 Switch `runLevel` and `runMainMenu` from `rl.BeginDrawing`/`rl.EndDrawing` to `rl.BeginTextureMode(VirtualScreen)`/`rl.EndTextureMode`, and call `present()` once per frame from the main loop after the stage switch. Verify `go build ./...` succeeds and both stages still draw when run.

## 3. Audio

- [x] 3.1 Add `audio.go`: `initAudio()` opening the device and loading `resources/music/Neon Sprint.mp3` as a looping music stream, and `closeAudio()`, both guarded by `rl.IsAudioDeviceReady()` so a machine without audio runs silently. Verify `go build ./...` succeeds.
- [x] 3.2 Add `updateAudio(dt)`: call `rl.UpdateMusicStream` every frame, move `musicFade` toward 1 in the menu stage and 0 otherwise, set the stream volume to `musicFade * Cfg.MusicVolume/100`, and pause the stream at zero fade. Call it from the main loop for both stages. Verify `go build ./...` succeeds.
- [x] 3.3 Wire `initAudio()` into `Game()` before `LoadTextures()` and `closeAudio()` into the shutdown path next to `rl.CloseWindow()`. Verify the game starts with music on the menu and falls silent when a match starts.

## 4. Options screen

- [x] 4.1 Add `menuOptions` to the menu screen constants, an Options button on the main menu between Join Game and Quit, and shift Quit down a row. Verify the main menu shows five entries in that order.
- [x] 4.2 Draw the Options screen in `drawMenuUI`: a `CheckBox` for fullscreen, a `ComboBox` for window size, a `Slider` for music volume (0-100, live), and Back. Each control applies its change immediately via `ApplySettings()`, and Back saves the settings and returns to the main menu. Verify `go build ./...` succeeds.
- [x] 4.3 Extend `TestMenuLayout` to cover the Options rows the same way as the Host rows (inside the panel, above the window bottom). Verify `go test -run TestMenuLayout .` passes.

## 5. Manual verification

- [x] 5.1 Run the game. Check: menu music plays and loops; the music fades out when Single Player starts and fades back in on Escape to the menu; the menu and level look unchanged at 1280x720.
- [x] 5.2 Open Options. Check that each control applies at once: the volume slider changes the playing music without a restart; the window size list resizes the window and the game fills it; fullscreen fills the monitor letterboxed. Check that buttons still activate under the cursor in fullscreen and at 1920x1080, and that the in-level attack cursor still tracks the mouse.
- [x] 5.3 Set volume 30 and fullscreen on, leave Options, restart the game, and check it starts fullscreen at volume 30. Delete `settings.json`, restart, and check it starts with defaults. Put malformed text in `settings.json`, restart, and check it still starts.

## Superseded in part

- [x] Its "Music stops during a match" requirement (the level left silent) is replaced by `level-music`, which plays a level track during a match. Archive this change before `level-music`, whose delta then becomes a MODIFIED of that requirement
