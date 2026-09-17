## Context

See proposal.md — Why. What matters for the approach is the shape of the scene pass in `runLevel`, which the composite has to fit into:

1. the world is drawn into `ScreenFullTexture` inside `Cam.View()` — tiles, then actors back to front;
2. `rl.EndMode2D()`, then a screen-space black rectangle at alpha 50 over the whole target;
3. `drawLightAccum()` adds `LightAccum` — every visible light, already placed in screen space by `computeLights` — additively;
4. `Cam.View()` again for the attack outlines, the structure bars and the editor's overlay, which are drawn **after** the lighting precisely so it cannot tint them;
5. `ScreenFullTexture` is drawn into `VirtualScreen` through `PaletteShader`, which snaps every pixel to its nearest Endesga 64 entry.

Constraints that shape the design: everything between (1) and (4) is 8-bit RGBA, the overlays in (4) must stay untinted, `LightAccum` is already a full-screen render texture cleared once a frame, and the palette pass at (5) quantizes whatever it is handed, so a composite that produces smooth gradients still lands on 64 colours.

## Goals / Non-Goals

**Goals:**

- One light map per frame — ambient plus every visible light — that the scene is modulated by, so darkness is multiplicative and shadows read as shadows.
- Ambient light authored per level, saved with it, and out of the fingerprint.
- No new render target, no new shader, no extra full-screen pass, and no change to what the overlays look like.
- The parts that can be tested without a GL context — the level file, the fingerprint, the clamping, the colour arithmetic — are tested.

**Non-Goals:**

- Changing how a light is computed: the Zima passes, the shared shadow textures and the per-light accumulation stay exactly as they are.
- Any tone mapping, exposure curve or bloom; the composite clamps as 8-bit blending clamps.
- Dithering the falloff. `ditherStrength` stays 0 and banding stays whatever the palette pass makes of it.
- A second ambient for a region of a level, or an ambient that changes over time.

## Decisions

### The composite is two blended draws in the scene pass, not a shader in the palette pass

The math wanted is `frame = scene × (ambient + light) + (ambient + light) × glow`. Both halves are available as blend modes over the buffer that already exists:

```
computeLights:  clear LightAccum to ambientClear(level.Ambient)   // was rl.Black
                add every visible light into it, as today

composite:      BlendMultiplied, tint white          -> frame = scene × map
                BlendAdditive,   tint {255,255,255,glow×255} -> frame += map × glow
```

`rl.BlendMultiplied` is `src × dst`, and raylib's default shader multiplies the source texel by the tint, so the tint's alpha is exactly the glow weight. The black rectangle and today's additive draw are replaced one for one; the cost per frame is unchanged.

The alternative — folding the composite into `PaletteShader` with `LightAccum` as a second sampler, which is where the math would naturally live — was rejected because of step (4) above. `PaletteShader` sees one texture holding the scene *and* the overlays, so lighting inside it would tint the structure bars, the energy bars, the attack outlines and the editor's light markers. Keeping the untinted overlays would mean splitting the scene and the overlays into two render textures and adding a pass to combine them: a new full-screen target and a third pass, to buy 8-bit rounding that a 64-colour palette then throws away. The two-draw form keeps the pipeline exactly as it is.

What is given up: no soft clip. Where `ambient + light` exceeds 1, a lit surface still clips to white as it does today. That is tuning, not structure — it moves with the lights' intensities — and a shader can be introduced later without touching the level format or the spec.

### The glow term takes the whole map, not just the lights

Adding `map × glow` adds `ambient × glow` along with the light: a flat lift of a few percent that puts a floor under the darkest pixels. Isolating the lights would need a second buffer or a subtract pass, for a difference of that few percent. The scenario in the spec that pins the visible consequence — ambient strength zero gives black outside the lights — still holds, because at zero ambient the map *is* the lights.

`glow` is a constant in code (≈0.2, settled by eye during apply), not level data. One more authored number per level buys little when the ambient already sets how dark the level is; if a level ever needs its own, it can join the ambient in the file without changing anything else.

### Ambient is `colour × strength`, stored as a colour and a percentage

```jsonc
"ambient": { "color": [120, 140, 255], "strength": 0.25 }
```

The colour is the hue of the dark at full brightness and the strength scales it, so an author picks "what colour is the night" and "how dark is it" separately instead of juggling three dim numbers. The clear colour is `color × strength` per channel, clamped to 0–255 — the one place the two combine, small enough to be a pure function (`ambientClear`) and therefore testable without a window.

Strength is edited as a percentage, 0–1000 like a light's intensity, so the light panel's existing integer fields take it unchanged; above 100 % the clear colour simply saturates.

### A missing `"ambient"` key and a black ambient are different things

`parseLevel` already walks the file's keys, so presence is known without a sentinel: the key absent means the default ambient, the key present with strength 0 means a level that is pitch black outside its lights. Both are in the spec, and the distinction is why the ambient is not decoded into a plain zero-valued struct.

Saving is not symmetric with this: every save writes the key, including one whose value matches the default. Omitting it would be lossless only until the default moved, and the default is exactly the kind of constant that gets retuned — a level authored against one default would then silently re-light when someone changed it for another level. A file without the key is therefore only ever one written before levels had an ambient at all. The cost is that a level file gains a key the first time it is saved, which the `KeepsItsBytes` test pins by stripping the key back out rather than by expecting it to be absent.

### The ambient stays out of the fingerprint for free

The fingerprint is `marshal()`, over the tile layers alone; the file is `marshalFile()`, which adds `"lights"` and now `"ambient"`. Nothing needs to be excluded — the ambient only has to be added to the second, and the existing fingerprint tests pin that.

### `level_1.json` is re-authored as part of the change

Today's lights were tuned against `scene × 0.804 + light`. Under `scene × (ambient + light)` the same numbers are darker everywhere and brighter at the centres, so the level is re-tuned — the ambient first, then any light intensity or radius that reads wrong — against a screenshot taken before the change, exactly as `add-level-lights` did. The acceptance is not "identical", which is impossible and not wanted, but "at least as readable": a fighter is never lost in the dark in a place the level expects them to fight.

## Risks / Trade-offs

- **Archive order.** Both delta specs target capabilities that exist only in in-flight changes — `lighting` in `add-level-lights`, `level-editor` in `add-editor-textures-and-rect-editing` and `add-level-lights`. `openspec validate` already reports that archiving this change first would refuse the `MODIFIED` requirement. → Archive `add-editor-textures-and-rect-editing` and `add-level-lights` first; the last task checks it before archiving.
- **The level gets darker everywhere, including where play depends on seeing.** → The before screenshot, the re-tune, and a pass through the level looking for a spot where a fighter disappears.
- **Posterization.** Multiplying by a dark ambient pushes a wide range of scene colours into a narrow band, and the palette pass then snaps that band to a handful of Endesga entries, so a gradient can step where it used to fade. → Keep the ambient light enough that the dark still spans several palette entries; dithering is the fix if it is not enough, and it is out of scope on purpose.
- **Clipping in bright spots** stays as it is: `ambient + light` over 1 clips to white. → Light intensities are re-tuned with the ambient; a soft clip is a later, self-contained change.
- **Alpha.** `BlendMultiplied` multiplies alpha as well as colour. `ScreenFullTexture` is cleared opaque and `LightAccum` is cleared to an opaque ambient, and the palette shader writes alpha 1 regardless, so nothing downstream can see it. → Keep the ambient clear opaque; a test pins the clear colour's alpha.
- **Nothing visual is covered by tests.** As everywhere in this project, the composite itself is verified by running the game; only the file format, the fingerprint, the clamping and the colour arithmetic are testable. → The tasks say which check is a test and which is by eye.

## Migration Plan

No migration. A level file without `"ambient"` loads with the default ambient, and a build from an older revision skips the key it does not know (`parseLevel` logs and skips unknown keys) and lights the level as it always did. Multiplayer is untouched: the ambient is out of the fingerprint, so a client and a server may disagree about it entirely and still play together.

Rollback is the revert: nothing is written to disk that an older build cannot read, and the only file that changes is `level_1.json`, whose tiles and fingerprint are untouched.
