## Context

See proposal.md (Why). The approach depends on how the scene pass in `runLevel` works today:

1. `computeLights(LE.lights, LE.areas, LE.ambient, LE.LightOccluders(), Cam.View())` builds `LightAccum`, the frame's light map. A level light is a `LevelLight` (`X`, `Y`, `Radius`, `Color`, `Intensity`). Its contribution is `color × intensity × (1 − d/R)²`, with the Zima shadow test on top (`shaders.go`, "Radial attenuation"). Its occluders are the tiles' light rectangles in world space.
2. Tiles and fighters are drawn into `ScreenFullTexture` in `Cam.View()`, back to front from `drawList`. A flickering fighter is drawn through `WhiteShader`.
3. `compositeLights()` multiplies the scene by the map and adds back `map × lightGlow`.
4. In `Cam.View()` again, draw `drawAttackOutline` for each fighter, then the structure bars, then `LE.DrawOverlay()`. Anything drawn here is untouched by the lighting.
5. The palette pass snaps every pixel to Endesga 64.

`attack_outline.go` already draws a fighter's sprite a second time through a shader (`fsOutline`). It gets the sprite from `Actor.spriteDraw()` and clamps sampling to the frame's rectangle in the sheet (`frameBounds`) so a neighbouring frame cannot bleed in. It draws nothing while `TimeShowsHistory()`, so `spriteDraw` has never had to handle the recorded frame that `Draw` shows during rewind.

Sprites are drawn unscaled, so one sprite texel is one world pixel. A left-facing fighter is drawn with a negative source width, which reverses the texture coordinates across the quad.

## Goals / Non-Goals

**Goals:**

- One extra sprite draw per fighter, through one new shader, added to the existing pipeline. No new render target and no change to the light map or the composite.
- The choices a unit test can pin are pure Go functions, tested without a GL context: which lights rim a fighter, how strong each is, whether one is shadowed, and where each is in the sprite's own coordinates.
- Reuse the outline pass's approach (`spriteDraw`, `frameBounds`) rather than inventing a second way to redraw a sprite.

**Non-Goals:**

- Normal maps or any per-sprite authored lighting data. The edge direction comes from the silhouette alone.
- Rims wider than one pixel, soft or blurred rims, and bloom.
- Rim from ambient light or lit areas. See the spec.
- Rims on level objects or tiles. Only fighters get them.
- Per-pixel shadowing. A light is either shadowed from the whole fighter or not.
- Any editor control or level data for the rim. A light rims fighters because it is a light.

## Decisions

### The rim is added after the composite, not drawn into the scene

The rim pass runs right after `compositeLights()` and before the attack outlines, in additive blending (`rl.BlendAdditive`). Each fighter's sprite is drawn once more through the rim shader. The shader discards every pixel that is not a lit edge pixel and outputs the rim colour for the rest.

Why here:

- **Not dimmed by the ambient.** Drawing the rim into the scene before the composite would multiply it by the light map. A rim at the edge of a light's radius sits where the map is darkest, so it would almost disappear exactly where it matters most. The spec requires the rim to keep its strength in the dark.
- **Additive can only brighten.** An edge facing away gets zero and is left exactly as the composite made it, which is what the spec asks for. Two lights on the same pixel add up without extra work.
- **The overlays stay above.** The attack outline, the bars and the editor overlay are drawn after this pass, as the spec requires. The outline paints the same ring of pixels the rim does, and being drawn later it takes precedence there.

Alternative considered: adding the rim into `LightAccum` so the composite applies it. The map is per screen pixel, not per sprite, so the shader would need the fighter's silhouette in the map pass. It would also be multiplied by the scene colour, so a rim on a dark cloak would stay dark. Rejected.

### The rim is an outer ring, drawn on a quad one pixel larger than the frame

The rim paints the transparent texels that touch the fighter, not the fighter's own texels. An inner rim was tried first and read badly: it eats a pixel of a 48-pixel sprite all round, so the fighter looks thinner and its own colours are replaced exactly where the silhouette is already hardest to read.

For the pixel being drawn, the shader works in the frame's own texel grid. It floors `(uv - frameBounds.xy) / texelSize` to a texel and uses that texel's centre, so every screen pixel of one sprite texel gets the same answer at any zoom. Then it:

1. discards the texel if its own alpha is at or above 0.5 (that is the fighter, and it must stay untouched);
2. reads the four neighbours through a frame-clamped `alphaAt`, where anything outside the frame counts as transparent;
3. forms `body = sum of dir * alpha(neighbour)`, which points from this texel towards the fighter, and discards the texel when it is zero - nothing adjacent, so not a rim pixel at all;
4. takes the outward direction as `-normalize(body)`;
5. for each light `i`: `L = normalize(lightPos[i] - texelCentre)` and `f = max(0, dot(outward, L))`, then adds `lightColor[i] * f`.

The output is `vec4(sum, 1)`. `BlendAdditive` scales by alpha, so alpha is 1 and the strength is carried in the colour.

Four neighbours rather than eight keeps the ring one pixel wide and matches `fsOutline`. The per-texel direction matters because lights sit close to fighters: a lamp overhead should light the outside of the head and the shoulders, and those face slightly different ways.

The quad is drawn one texel larger than the sprite frame on every side, with the source rectangle grown to match. A fragment shader only runs where something is drawn, so on the frame's own quad the ring would be cut off wherever the fighter reaches the frame's edge - which the sprites do at the feet and often at the top of the head. Growing the quad gives the ring somewhere to land; `alphaAt` still clamps sampling to the frame's own rectangle, so no neighbouring pose can bleed in. The negative source width that mirrors a left-facing fighter is preserved, and the grown rectangle is mirrored with it.

### A fighter drawn in two layers is one silhouette to the shader

A walking fighter is already drawn in two layers: `DrawAnim` draws the body's walk and then, for a guard height of low, mid or high, the matching arms sheet over it (block-all has none). `Actor.AnimTop` is the general form of the same thing, a second sprite drawn over the body. The rim must run around the two together, with no line along the seam where the upper sprite overlaps the body, so the shader takes the union of their coverage: `alphaAt` returns the greater of the body's alpha and the top layer's alpha at the same frame pixel. The top layer is a second sampler with its own bounds; `hasTop` switches it off when there is none, and the union is then the body alone.

Both sprites are drawn from the same frame index at the same position, so one pixel of the frame maps to the same pixel of either sheet, and the shader converts a frame pixel into each sheet's texture coordinates with that sheet's own texel size and frame origin.

Which sprite is on top is answered in one place, `Actor.topAnim`: `AnimTop` when the state machine has set one, otherwise the walk's own rule. `Actor.topDraw` turns that into a texture, a source rectangle and a position, and both `DrawAnim` and the rim draw from it - the same reason `spriteDraw` exists, so the rim cannot trace a silhouette the fighter is not showing. The walk's rule stays a `TODO` where it already was: it belongs with the animation, not in the drawing, and moving it there later leaves `AnimTop` as the only path.

### Strength, choice and shadowing are decided once per fighter, on the CPU

`rimLights(lights []LevelLight, occluders []rl.Rectangle, body rl.Rectangle) []rimLight` is a pure function. `body` is the fighter's collision box at its draw position, the same rectangle `outlineRect()` gives. For each light it:

- measures the distance `d` from the light's centre to the centre of `body`;
- computes the strength `intensity × (1 − d/R)²`, the formula the light shader uses, clamped at 0;
- drops the light if its strength is below a small threshold (`rimMinStrength`), which covers "outside the radius" and "zero intensity";
- drops the light if the segment from its centre to the body centre crosses any occluder (`segmentHitsRect`, a slab test);
- keeps the `rimMaxLights` (4) strongest, sorted by strength.

The strength is multiplied into the colour on the CPU (`color × strength × rimStrength`), so the shader receives only a position and a colour per light.

Measuring once, at the body centre, rather than per pixel: a sprite is 48 pixels and the default radius is 128, so the falloff barely varies across one fighter. A single value also makes the choice and the shadow test agree with the strength, and all three can be tested without a window. The shadow test is a line to the body centre, not the Zima shadow map. The map is a screen-space texture that is overwritten light by light, so reading it back per fighter would cost a GPU readback per light. A segment test against the same rectangles gives the same answer for a fighter that is plainly in or out of a shadow.

`rimMaxLights` is fixed because GLSL 330 uniform arrays need a constant size. The shader gets `lightCount` and loops up to it.

### Light positions are converted into the frame's texel space on the CPU, mirrored for a left-facing fighter

`rimLocal(light rl.Vector2, spritePos rl.Vector2, frameWidth float32, facingLeft bool) rl.Vector2` returns `light − spritePos`, with `x` replaced by `frameWidth − x` when the fighter faces left. The shader then adds the frame's own texel origin, so the texture coordinates and the light are in the same unflipped sheet space.

The mirror is needed because a negative source width reverses the texture coordinates across the quad. The texel the shader reads at the left of the screen is the right-hand texel of the frame. Without the mirror a left-facing fighter lit from the left would be rimmed on its right. This is pure arithmetic, so a test pins it, and the spec's "Turning around" scenario is the visual check.

### `spriteDraw` becomes the one source for both live and recorded frames

`Draw` currently builds its history rectangle inline. That branch moves into `spriteDraw`: while `TimeShowsHistory()` it returns `resources[data.Anim]`, frame `data.AnimFrame`, height 48 and `drawPos()`, exactly what `Draw` draws today. `Draw`'s history branch then calls `spriteDraw` too, so the two cannot drift. `drawAttackOutline` already returns early in history, so its behaviour does not change. The rim pass skips a fighter when `deathHidden()` is true, as `Draw` does.

### Shader locations are looked up once

The rim pass is called once per fighter per frame with several uniforms. Its locations (`texelSize`, `frameBounds`, `lightCount`, `lightPos`, `lightColor`) are looked up once when the shader loads in `runLevel`, next to `OutlineShader`, and kept in package vars. The arrays are set with `rl.SetShaderValueV` and a count. Uniform arrays are looked up by their bare name.

## Risks / Trade-offs

- **The palette can swallow a dim rim.** A weak rim added to a dark pixel may snap back to the same Endesga entry, so rims fade out in steps and vanish before the radius. → `rimStrength` is a constant tuned by eye in `level_1`. The spec only requires the rim to fade and be gone beyond the radius, which snapping does not break.
- **A light inside an occluder never rims.** The segment starts inside the rectangle, so the test reports a hit. The Zima pass also lights nothing from inside an occluder, so this matches what the map shows. → A task checks `level_1`'s lights against its light rectangles, so an intended rim is not lost to a lamp sitting in a wall's light rect.
- **Shadowing is all or nothing at the body centre.** A fighter half in shadow gets the full rim or none, and the rim can switch on across one pixel of movement. → Accepted. If it looks bad, testing the top and bottom of the body and averaging is a local change to `rimLights`.
- **Thin features rim on both sides.** A one-pixel-wide sword or strand has empty pixels either side, and each of them sees the body in the opposite direction, so the one facing the light is rimmed. That is correct, but it doubles a thin feature's apparent width where the light falls. → Accepted as how an outer rim works.
- **The union depends on the two sheets lining up.** The top layer is sampled at the same frame pixel as the body, so a top sheet framed differently from the body's, or drawn at its own offset, would be rimmed in the wrong place. The walk's four sheets are all 432x48 and drawn at the same position, and `topDraw` takes its rectangle and position from the body's, so they cannot drift. → A sheet that ever needs its own offset has to pass it to the shader too.
- **The attack outline and the rim share pixels.** Both paint the ring just outside the fighter. The outline is drawn after the rim and blends over it, so a marked fighter's mark stays legible and the rim survives on the sides the band does not cover. → Recorded in the spec as precedence rather than as independence.
- **The hit flash already whitens the sprite.** The rim adds on top and saturates. → Harmless. The flash lasts a few frames.
- **The silhouette grows by a pixel.** An outer rim adds a lit pixel around the fighter, so a rimmed fighter reads very slightly larger. → Intended: it is what separates the fighter from the background, and it never moves or hides a pixel the fighter draws, so nothing about reading a pose or a hitbox changes.
- **Cost.** One extra draw and a few uniform sets per fighter. Matches have a handful of fighters, and the per-fighter CPU work is a loop over the level's lights (tens) and occluders (hundreds). → If a large level makes that loop noticeable, `rimLights` can pre-filter occluders by the segment's bounds. It is not needed for `level_1`.

## Migration Plan

No data migration: nothing is added to the level file, the fingerprint, or the wire. Rollback is a revert of the change. An older build and a newer one can play together because the rim is purely local drawing.
