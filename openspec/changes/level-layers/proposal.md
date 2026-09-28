## Why

Every tile is drawn before every fighter, so nothing in the level can stand in front of one: no pillar, railing or foreground crate. The level data already holds tiles by layer (`map[int]LevelLayer`, `LevelObject.Layer`), but nothing ever changes the layer, every tile sits on layer 0, and the draw pass ignores the layers.

## What Changes

- Layer 0 belongs to the fighters. Tiles on negative layers are drawn behind them, and tiles on positive layers are drawn in front of them. Within each group, lower layers are drawn first.
- The editor has a current layer. New tiles are placed on it and erasing removes tiles from it. It starts at -1, PageUp/PageDown step it, and stepping skips 0. The editor panel shows the current layer.
- **BREAKING** (level file): tiles saved on layer 0 are loaded onto layer -1, so the next save rewrites `level_1.json` with its tiles under `"-1"`. This changes the level's fingerprint, so client and server builds must be rebuilt together. They always are.
- Collision is unchanged: every tile collides no matter which layer it is on.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `level-editor`: new requirements for placing tiles on a chosen layer, the layer's draw order relative to fighters, and loading layer-0 tiles from older level files.

## Impact

- `level_editor.go`: `place`, `erase`, the key handling and the panel use and show `le.layer`, which starts at -1. `Draw` splits into a pass behind the fighters and a pass in front of them.
- `game.go`: the level draw pass calls the background pass before stains and fighters, and the foreground pass after puffs and before `compositeLights`.
- `level_lights.go` (`parseLevel`): moves layer-0 tiles to layer -1 on load.
- `level_1.json`: rewritten under `"-1"` on its next save.
- Known limitation: the rim light is added after lighting, so a fighter behind a foreground tile shows a rim through it. This is left for later.
