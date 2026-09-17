## Why

The level's lighting is not part of the level. There are exactly three lights, and both where they are and what they are live in `game.go`: three `ZimaLight`s created with fixed radii (256, 128, 128), all white, all at intensity 1, drawn every frame at world positions typed into the render loop — (40, 30), (120, 120) and (−40, 30). The editor can now place any texture and shape every tile's light occluder, but the lights those occluders exist to block cannot be added, moved, recoloured or removed without editing Go, and a level built anywhere but around those three points is simply dark.

## What Changes

- **The level carries its lights.** Each light has a position, a radius, a color and an intensity, and is saved in `level_1.json` with the rest of the level. The game draws every light the level has, with the shadows the tiles' light rectangles cast, exactly as the three built-in lights are drawn today.
- **The three built-in lights become ordinary level lights.** They move into `level_1.json` with the positions, radii, color and intensity they have now, so the level looks the same; nothing in the code places a light any more.
- **A light tool in edit mode.** A third tool joins place and rect: with it, a click on empty space puts a light there, a light can be selected and dragged, its radius dragged by a handle on its edge, and its numbers typed into the panel. Delete removes it. A new light takes the settings of the light selected when it is placed, so a row of matching lights is one shape and a series of clicks.
- **Lights are visible while editing.** In edit mode with the light tool, every light shows its center and the circle it reaches, the selected one marked.
- **Light edits save like any other edit**: shown as unsaved, written only by Ctrl+S in edit mode, never by a release build.
- **Lights do not change the level fingerprint.** The fingerprint goes on covering the tiles and nothing else, so a lighting change never turns into "level mismatch" between builds whose play is identical, and today's level keeps the fingerprint it has.

## Capabilities

### New Capabilities

- `lighting`: what the level's lights are and how they are drawn — lights as level data, drawing every light with its color and intensity, lights off screen costing nothing, and lights staying out of the fingerprint.

### Modified Capabilities

- `level-editor`: placing, selecting, moving, resizing, recolouring and removing lights, seeing them while editing, and saving light edits. Added requirements only. This capability is introduced by the in-flight `add-editor-textures-and-rect-editing` change.
- `dev-tools`: *The editor takes input only in edit mode* — edit mode now has three tools rather than two. That wording is itself being introduced by `add-editor-textures-and-rect-editing`, so this change has to be archived after it.

## Impact

- `level_editor.go`: the lights, the level file's `"lights"` key, and the light tool. `game.go` and `shaders.go`: the hardcoded lights and their globals go; one shared set of shadow textures is reused for every light and the results are added into one accumulation buffer, so memory no longer grows with the number of lights.
- `level_1.json` gains a `"lights"` array holding today's three lights. Its tiles are untouched, and so is its fingerprint: the fingerprint is computed over the tiles alone, as it is now. A development build of an older revision cannot read a level with lights; every build carries its own level, so this only matters when switching revisions with a level saved on disk.
- The dedicated server loads the lights with the level and never draws them.
- Every visible light costs a dozen GPU passes a frame, as each of today's three does. Lights wholly off screen are skipped, but a view crowded with lights will cost what it costs; there is no cap.
- Not in scope: spotlights (the light code has direction and cone fields that nothing uses yet), moving or flickering lights, copy and paste of lights, and lights attached to tiles.
