## Why

Since the light map replaced the black wash, an unlit part of the level is genuinely dark and a tile's light rectangle genuinely blocks light. That is what the change was for, and it leaves nowhere to author an exception: a corridor a light cannot reach, a room whose ceiling tiles shadow everything under them, or a strip of the level that simply has to read clearly for a fight, are all only fixable by moving lights and tiles until the darkness happens to lift. There is no way to say "this rectangle of the level is lit, whatever stands in it".

## What Changes

- **A level carries lit areas**: rectangles, in whole level pixels like the tiles and the lights, inside which the scene is drawn at full brightness — neither the level's ambient dark nor any shadow cast into them by a tile's light rectangle darkens what is there.
- **A lit area is soft-edged.** Each area carries a feather: how many pixels of falloff its edge fades over, from full brightness inside to the light the rest of the map gives just outside it, so an area never reads as a straight bright line across the level. A feather of zero is a hard edge.
- **Lit areas are level data**, saved in the level file beside the lights and the ambient, loaded with it in the development build, the release build and the dedicated server alike. A level file with no lit areas in it loads exactly as it does today.
- **Lit areas stay out of the level fingerprint**, like the lights and the ambient: how bright a corner of a level is must never turn into "level mismatch" between builds whose play is identical.
- **The editor gains an area tool**, fourth in the Q rotation: left click on empty space places an area, its edges and corners drag like the rect tool's rectangles, the panel shows its numbers — position, size, feather — as fields that set it exactly, right click removes it, and an edit marks the level unsaved and is written by Ctrl+S like any other.
- **Lit areas are drawn only by the editor**, as outlines while the area tool is out; in play nothing shows but the light itself.

Assumptions recorded here rather than asked about: an area is fully lit or not at all, so there is no per-area strength or colour (a dimmer lift is a light, and a coloured one is a light too); areas are axis-aligned rectangles, like every other rectangle the level holds; and an area lifts the dark for everything in it — tiles, fighters and effects alike — rather than for a chosen layer.

Not in scope: areas that darken rather than lighten; per-area ambient colour; areas that block or shape light rather than replace it; animating an area; and shapes other than a rectangle.

## Capabilities

### New Capabilities

- None.

### Modified Capabilities

- `lighting`: the light map gains lit areas — a rectangle of the level that the map holds at full brightness with a feathered edge, and that is level data saved and loaded with the lights and the ambient and kept out of the fingerprint as they are. This capability is introduced by the in-flight `add-level-lights` change and reworked by the in-flight `light-map-composite` change, so this change has to be archived after both.
- `level-editor`: the area tool — placing, dragging, editing and removing a lit area, and its panel. Added requirements only. The capability is introduced by the in-flight `add-editor-textures-and-rect-editing` change and extended by `add-level-lights` and `light-map-composite`; this change has to be archived after all three.

## Impact

- `level_lights.go`: a `LitArea` type beside `LevelLight` and `Ambient`, the `"litAreas"` key in `parseLevel` and `marshalFile`, and the areas' draw into `LightAccum` in `computeLights` — additively, in world space, in the same pass the lights are added in, so the order of a light and an area never matters.
- The feathered edge is nine draws - a solid core and four fading sides as raylib's own gradient rectangles, and four corners from one 2x2 texture - rather than a new shader or a new render texture: additive blending already scales what is added by alpha, so a fading alpha is a fading lift. An area off screen costs nothing.
- `level_editor.go`: a fourth tool in the Q rotation, the areas' selection and drag state on `LevelEditor`, their panel and their overlay. The handle hit-testing and dragging in `rect_edit.go` are reused as they are — they already work on a plain world-space `rl.Rectangle`.
- `level_1.json` may gain a `"litAreas"` array once areas are authored into it; its tiles, lights and ambient are untouched, and so is its fingerprint.
- The dedicated server draws nothing, so lit areas only ride along with the level it loads, as the lights and the ambient do.
- A build of an older revision reading a level saved with lit areas ignores the key it does not know and lights the level as it always did.
