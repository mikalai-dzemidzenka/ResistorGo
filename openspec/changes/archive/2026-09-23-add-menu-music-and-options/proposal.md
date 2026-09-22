## Why

The game has no sound at all — no audio device is ever opened — and no way to change anything about how it runs. The window is locked to 1280x720 with no fullscreen, and there is nothing to adjust if the player wants it quieter or bigger. The main menu is silent and the only screens are Host, Join and Connecting.

## What Changes

- Audio: open an audio device at startup and loop a music track on the main menu (Neon Sprint). It fades out when a match starts and fades back in on return to the menu. The level stays silent, as it has no audio design yet.
- Presentation: the game renders into a fixed 1280x720 virtual screen, which is then scaled to fit the window or monitor and centred with black bars. Everything on screen — the level, the menu, the HUD and the cursor — goes through it. Every player sees exactly the same amount of the level whatever their resolution, so multiplayer stays fair, and no gameplay or layout code needs to know the window size.
- Display options: fullscreen (borderless, at the desktop resolution) and a window-size list (1280x720, 1600x900, 1920x1080), applied immediately.
- Audio options: a music volume slider, applied live as it is dragged.
- An Options entry on the main menu, opening a screen with those controls and a Back button, reachable from the menu only (not mid-match).
- Settings persist in `settings.json` next to `level_1.json`, written when the options screen is left and loaded at startup. A missing or corrupt file falls back to defaults rather than failing to start.
- **BREAKING** for code, not for players: the mouse position that gameplay and UI code reads becomes virtual-screen coordinates rather than window coordinates.

## Capabilities

### New Capabilities
- `audio`: opening the audio device, playing and looping menu music, and applying volume.
- `display`: the virtual screen, how it is scaled and letterboxed to the window, fullscreen and window size.
- `settings`: the options screen and the settings file that persists the player's choices.

### Modified Capabilities
- `main-menu`: the menu is drawn into the virtual screen rather than straight to the window, and gains an Options entry and an Options screen that must fit the menu panel.
- `multiplayer/session`: the main menu entries requirement lists the entries, so it must include Options.

## Impact

- New code: `settings.go` (load/save/apply), `audio.go` (device and music), `present.go` (virtual screen, letterbox, mouse mapping).
- Changed: `game.go` (startup, main loop, present step), `main_menu.go` (Options entry and screen), `resources.go` (music loading).
- `cursor.go`, `spectator.go` and `level_editor.go` keep working unchanged because they already use virtual coordinates; only the mouse source changes beneath them.
- New dependency: none. New asset use: `resources/music/Neon Sprint.mp3`, already in the repo.
- New file written at runtime: `settings.json` (gitignored).
