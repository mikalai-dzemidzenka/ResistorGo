## Why

The level is lit now: an ambient dark, level lights with shadows, and lit areas. A fighter is lit the same way as the wall behind it, though, so under a lamp it looks flat and in the dark it blends into the tiles. Katana Zero solves this with rim light. The edge of a character that faces a light picks up that light's colour, so the silhouette separates from the background and you can tell where the light comes from. A game about reading an opponent's body needs that silhouette.

## What Changes

- A fighter near a level light SHALL get a rim light: a one-pixel line in the light's colour along the sprite edges that face that light. The edges facing away stay unlit. A lamp on the fighter's left lights its left edges, and a lamp above lights its head and shoulders.
- The rim is on the fighter's own outermost pixels. The silhouette stays the same size, and the red attack-target outline, which paints the transparent pixels just outside the body, stays separate and on top.
- The rim is as strong as the light is at that spot. It uses the light's colour and intensity and the same falloff with distance as the light map. It fades as the fighter walks away and disappears at the light's radius.
- A light that a tile's light rectangle shadows from the fighter gives no rim. The rim cannot contradict the shadows the level casts.
- Several lights can rim one fighter at once, each on its own side and in its own colour. The strongest few in reach are used.
- Ambient light and lit areas have no direction, so they give no rim.
- The rim follows the animation frame by frame, mirrors correctly when the fighter turns, and stays on during time rewind. It is not drawn while a dead fighter is flickered out.
- It is scene lighting, not a player cue, so every fighter gets it and every viewer sees it: the local player, remote players, bots and spectators.
- Presentation only. Nothing about the simulation, the network, the level file or the fingerprint changes, and there is no new asset or editor tool.

## Capabilities

### New Capabilities

- `actor-rim-light`: which lights rim a fighter, which edge pixels they light, in what colour and how strongly, and when no rim is drawn.

### Modified Capabilities

None. The rim reads the level lights and their occluders as they are. The `lighting` and `attack-target-outline` capabilities are still in flight (`add-level-lights`, `light-map-composite`, `add-lit-areas`, `outline-attack-target`), and neither needs a changed requirement. The only constraint between them, that the attack outline stays visible above the rim, is stated in this change's own spec.

## Impact

- `shaders.go`: a new fragment shader that finds a sprite frame's edge pixels and lights those that face each given light.
- A new file, `rim_light.go`, next to `attack_outline.go`. It picks the lights that reach a fighter, tests them against the occluders, converts them into the sprite frame's own coordinates (mirrored when the fighter faces left) and draws the sprite a second time through the shader.
- `game.go` (`runLevel`): loads the shader beside `OutlineShader`, and adds a rim pass after `compositeLights()` and before the attack outline.
- `actor.go`: `spriteDraw` learns to return the recorded frame while time shows history, so the rim traces what `Draw` shows during rewind. The attack outline is unaffected because it already draws nothing then.
- Tests for the pure parts (light choice, falloff, occlusion, mirroring) in `rim_light_test.go`. The look itself is checked by running the game, as with all visuals here.
- Cost: one extra sprite draw per fighter per frame, with up to four lights as uniforms. The dedicated server draws nothing.
