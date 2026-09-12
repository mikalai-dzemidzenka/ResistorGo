## Context

See proposal.md (Why). What the code looks like now:

- `game.go` opens a 1280x720 window (`SW*WindowScale` x `SH*WindowScale`) and never touches it again. No audio device is ever opened — there is no audio code in the package at all.
- The level renders the scene into `ScreenFullTexture` (1280x720), then runs the palette shader while blitting that straight to the backbuffer.
- The menu renders its 3D scene into `MainMenuTexture` (320x180), blits it to the backbuffer at `menuPixel`, then draws the logo and the raygui widgets straight to the backbuffer.
- 1280x720 is hardcoded as `SW*WindowScale` in ten places: `CamOffset`, the cursor's start position, the level render texture, the dim rectangle, the palette shader's `resolution` uniform, the blit source and destination rectangles, the spectator HUD position and its drag clamps, and the menu's own panel constants.
- `MousePos` is `rl.GetMousePosition()`, in window pixels, and everything downstream (raygui, `Cursor`, the level editor, spectator drag) assumes those are the same as the 1280x720 drawing space. Today they are.

Checked in raylib-go v0.60: `SetMouseOffset`, `SetMouseScale`, `ToggleBorderlessWindowed`, `GetCurrentMonitor`, `GetMonitorWidth/Height`, `SetWindowSize`, `InitAudioDevice`, `LoadMusicStream`, `UpdateMusicStream`, `SetMusicVolume` all exist, as do raygui's `Slider`, `CheckBox`, `ComboBox` and `Toggle`.

## Goals / Non-Goals

**Goals:**
- One place that decides how the finished frame reaches the window, so no gameplay or layout code learns about window size.
- Resolution and fullscreen that cannot change how much of the level a player sees.
- Menu music and a music volume that applies live.
- Settings that survive a restart and never stop the game from starting.

**Non-Goals:**
- Sound effects, and music during a match. The level has no audio design; this change only stops the menu track when a match starts.
- Non-16:9 virtual screens, or a resolution that changes the virtual screen. The virtual screen stays 1280x720; the list only changes the window it is scaled into.
- Key rebinding, graphics quality settings, or an in-match pause menu with options.
- Per-monitor or multi-monitor selection beyond "the monitor the window is on".
- Changing `menuPixel`, `WindowScale` or the palette pipeline.

## Decisions

### 1. A virtual screen, with one present step
Add `VirtualScreen`, a 1280x720 render texture with point filtering. Every frame becomes:

```
BeginTextureMode(VirtualScreen)
    ... the whole frame: level or menu, HUD, cursor, debug text ...
EndTextureMode()
BeginDrawing()
    ClearBackground(Black)          // the letterbox bars
    DrawTexturePro(VirtualScreen, srcFlipped, dstLetterbox, ...)
EndDrawing()
```

`present()` owns the second half. `runLevel` and `runMainMenu` keep their existing draw code almost verbatim; their `BeginDrawing`/`EndDrawing` pairs become `BeginTextureMode(VirtualScreen)`/`EndTextureMode`.

Because the virtual screen is exactly the old window size, every one of the ten hardcoded `SW*WindowScale` uses stays correct and needs no edit. That is the main reason for choosing 1280x720 as the virtual size rather than, say, 160x90.

*Nesting:* the level already renders into `ScreenFullTexture` before blitting. Raylib does not allow nested `BeginTextureMode`, so the existing texture pass must end before the virtual-screen pass begins. The level's structure already does this: build `ScreenFullTexture`, end it, then draw. Only the target of that second half changes.

*Alternative:* keep drawing to the backbuffer and letterbox by scaling each draw call. Rejected: every call site would need the scale factor, and the pixel grid would break up.

### 2. Letterbox geometry
```
scale = floor?(min(winW/1280, winH/720))   // see below
dst   = { x: (winW - 1280*scale)/2, y: (winH - 720*scale)/2, w: 1280*scale, h: 720*scale }
```
The scale is **not** rounded to an integer. All three offered window sizes (1280x720, 1600x900, 1920x1080) are exact 16:9, so they give exactly 1.0, 1.25 and 1.5 and letterbox to nothing. Fullscreen at a non-16:9 desktop (for example 1920x1200) gives a fractional scale and real bars. Forcing an integer scale there would waste a lot of screen (1.0 instead of 1.66 on a 1920x1200 display), so a fractional scale plus point filtering is the better trade: some pixels are one screen pixel larger than others, which is far less visible than a much smaller image.

*Alternative:* integer-only scaling for a perfect pixel grid. Rejected for the reason above; can be revisited as a "pixel-perfect" option later.

### 3. Mouse mapping via raylib's own offset/scale
```
SetMouseOffset(-dst.x, -dst.y)
SetMouseScale(1/scale, 1/scale)
```
applied whenever the letterbox is recomputed. `GetMousePosition` then returns virtual coordinates, so raygui (which calls it internally and cannot be intercepted), `Cursor`, the level editor and spectator drag all keep working untouched. This is why the mapping is done here rather than by adjusting `MousePos` ourselves.

A cursor over a black bar maps outside 0..1280 / 0..720, so it naturally hits nothing.

*Alternative:* convert `MousePos` in `Game()` and leave raylib alone. Rejected: raygui would still read raw window coordinates and every button would be offset in fullscreen.

### 4. Settings: a small struct, a JSON file, one apply function
`Settings{ Fullscreen bool, WindowSize int, MusicVolume int }` in `settings.go`, saved to `settings.json` with the same `encoding/json` approach `level_editor.go` uses for levels. `WindowSize` is an index into a fixed `windowSizes` list, so an unknown value clamps to a valid one.

- `LoadSettings()` — a missing file, a parse error or an out-of-range value each fall back to that field's default, never an error to the player. Defaults: windowed, 1280x720, volume 70.
- `ApplySettings()` — resizes or toggles the window, sets the music volume, recomputes the letterbox. Called at startup and whenever an option changes.
- Saved when the options screen is left (Back), not on every slider tick, so dragging the volume does not write the file 60 times a second.

`settings.json` is gitignored: it is per-machine state, like a save file.

*Alternative:* store the window size as width and height. Rejected: an index cannot express an unsupported size, and validating a free pair is more code for no benefit.

### 5. Fullscreen is borderless at the desktop resolution
`ToggleBorderlessWindowed()` rather than `ToggleFullscreen()`: it avoids a mode switch, alt-tabs cleanly, and is what a letterboxed 2D game wants. The monitor is `GetCurrentMonitor()`, so it follows the window. Leaving fullscreen restores the configured window size.

### 6. Music
`audio.go` owns the device and the menu track:
- `InitAudioDevice()` at startup, before `LoadTextures`; `CloseAudioDevice()` on exit. `IsAudioDeviceReady()` guards everything, so a machine with no audio runs silently instead of crashing.
- The track is `resources/music/Neon Sprint.mp3`, loaded as a music stream (streamed, not decoded into memory — the file is 3.3 MB) with `Looping = true`.
- `UpdateMusicStream` must be called every frame or the buffer starves, so `updateAudio(dt)` runs in the main loop for both stages, not inside `runMainMenu`.
- Fade: a `musicFade` value moved toward 1 on the menu and 0 in a match, at 1/`musicFadeSec` per second, multiplied by the volume setting. At 0 the stream is paused, so a silent match costs no decoding.

The duplicate `(1)` files in `resources/music` are ignored; the track is one constant, `menuMusicPath`.

## Risks / Trade-offs

- [A `BeginTextureMode` left unbalanced would silently draw into the wrong target] → There are only three pairs (`ScreenFullTexture`, `MainMenuTexture`, `VirtualScreen`) and they nest strictly. Verified by running both stages.
- [Fractional scaling makes some pixels larger than others] → Accepted, and invisible at the offered window sizes, which are all exact multiples. Only non-16:9 fullscreen is affected.
- [`SetMouseOffset`/`SetMouseScale` are global raylib state] → They are set once per letterbox change and never unset, so every reader sees virtual coordinates consistently. Anything genuinely needing window coordinates would have to undo them; nothing does today.
- [Resizing the window while a match is running] → The virtual screen is independent of window size, so nothing in the simulation or the render pipeline is affected; only the present step changes. Options are menu-only anyway.
- [`settings.json` written next to the executable's working directory] → Same convention as `level_1.json`, which the game already writes on exit.
- [No audio device on CI or a headless machine] → Everything is behind `IsAudioDeviceReady()`, and the tests added here touch only pure functions, so they do not need a device.
