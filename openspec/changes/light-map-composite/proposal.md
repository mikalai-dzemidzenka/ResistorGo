## Why

The dark of the level is a black rectangle at alpha 50 drawn over the whole frame (`game.go`, between the scene and the lights), and every light is then added on top of it. Darkness is therefore a flat 20 % wash — `scene * 0.804` — that no shadow can deepen, no level can colour and nothing can tune, while light is added to whatever happens to lie under it: a light washes the black sky exactly as brightly as it washes the floor, and anything already bright clips to white. The shadow passes the light code goes to real trouble over (distance, distort, reduction, shadow) end up deciding only whether a pixel gets the wash or the wash plus a glow, which is why the shadows read as haze instead of dark.

## What Changes

- **The accumulation buffer becomes a light map.** It is cleared to the level's ambient light rather than to black, every light adds into it exactly as it does today, and the scene is then *modulated* by it: an unlit surface is the scene times the ambient, a lit one is the scene times ambient plus light. A shadow is genuinely dark and keeps the hue of what is in it, and a light can no longer make a black pixel white.
- **A small glow keeps lights visible in the air.** After the modulation, the light map is added back at a low weight, so a light still reads as a glow over empty space and a near-black tile still catches something — the look the current additive draw gives, at a fraction of its strength, rather than losing it entirely.
- **The black rectangle goes.** Nothing draws flat darkness over the frame any more; the ambient in the light map is the only thing that decides how dark an unlit pixel is.
- **Ambient light is level data.** The level carries an ambient colour and strength beside its lights, saved in `level_1.json`, so one level can be a cool night and another a warm dusk. A level file with no ambient in it loads with a default ambient, exactly as a level file with no lights loads with no lights.
- **The light tool edits the ambient.** With no light selected, the editor panel shows the level's ambient colour and strength as fields that set it exactly, the change shows on the next frame, and it saves with Ctrl+S and marks the level unsaved like every other edit.
- **The ambient stays out of the level fingerprint**, like the lights: changing how dark a level is must never turn into "level mismatch" between builds whose play is identical.
- **`level_1.json` is authored for the new composite.** Its ambient is set so the level reads at least as well as it does now, which is part of this change rather than a follow-up: the old wash and the new modulation cannot both be right for one set of numbers.
Not in scope: dithering the light falloff (the palette pass's `ditherStrength` stays 0 - it was tried during implementation and the Bayer pattern, which the palette shader applies to the whole frame rather than to the gradient alone, stippled the flat tiles), bloom or any HDR pass, normal maps, softer or coloured shadows, moving or flickering lights, and animating the ambient for a day/night cycle.

## Capabilities

### New Capabilities

- None.

### Modified Capabilities

- `lighting`: how the level's light reaches the screen — the light map and its modulation of the scene replace the black wash and the additive draw, and the level gains an ambient that is data like its lights and out of the fingerprint like them. This capability is introduced by the in-flight `add-level-lights` change, so this change has to be archived after it.
- `level-editor`: the light tool's panel edits the level's ambient when no light is selected, and an ambient edit saves like any other. Added requirements only. The capability is introduced by the in-flight `add-editor-textures-and-rect-editing` change and extended by `add-level-lights`; this change has to be archived after both.

## Impact

- `game.go`: the `rl.DrawRectangle(..., rl.NewColor(0, 0, 0, 50))` wash and the additive draw of the accumulation buffer are replaced by the composite — a multiplied draw and a weak additive one, in the same place in the scene pass. The attack outlines, the structure bars and the editor overlay go on after it exactly as they do now, so they stay untinted.
- `level_lights.go`: an ambient type beside `LevelLight`, the `"ambient"` key in `parseLevel` and `marshalFile`, the clear of `LightAccum` to the ambient, the composite itself, and the ambient fields in the light panel.
- `level_1.json` gains an `"ambient"` object. Its tiles and its lights are untouched, and so is its fingerprint, which is computed over the tiles alone.
- The whole level looks different — that is the point of the change. Every light's radius, colour and intensity was tuned against a wash that is going away, so `level_1.json`'s lighting is re-tuned with it.
- No new render texture and no new shader: the composite is two blended draws of the buffer that already exists, in place of a rectangle fill and one blended draw. The cost per frame is what it is today.
- The dedicated server draws nothing, so the ambient only rides along with the level it loads, as the lights do.
- A development build of an older revision reading a level saved with an ambient ignores the key it does not know and lights the level as it always did; every build carries its own level, so this only matters when switching revisions with a level saved on disk.
