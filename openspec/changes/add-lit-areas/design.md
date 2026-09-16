## Context

See proposal.md — Why. What matters for the approach is the shape of the lighting as `light-map-composite` leaves it:

1. `computeLights(LE.lights, LE.ambient, LE.LightOccluders(), Cam.View())` clears `LightAccum` to `ambientClear(ambient)` and then, for each light whose area is in view, computes its shadow texture and adds it into the map additively, inside `BeginMode2D(cam)` — so the map is built in world space but lands where the camera shows it;
2. the scene is drawn into `ScreenFullTexture`, and `compositeLights()` then does `frame = scene × map` (`rl.BlendMultiplied`) followed by `frame += map × lightGlow` (`rl.BlendAdditive`, `lightGlow = 0.2`);
3. the attack outlines, the bars and the editor's overlay are drawn after the composite, so nothing in the map touches them;
4. the frame goes through `PaletteShader`, which snaps every pixel to its nearest Endesga 64 entry.

The level file (`parseLevel` / `marshalFile` in `level_lights.go`) is a JSON object keyed by layer number plus `"lights"` and `"ambient"`; `marshal()` — the fingerprint's input — is the tiles alone, which is how lights and ambient stay out of it. The editor has three tools on Q (`toolPlace`, `toolRect`, `toolLight`), and `rect_edit.go` already holds world-space rectangle editing that knows nothing about objects: `rectHandleAt`, `handleAnchor`, `dragRect`.

## Goals / Non-Goals

**Goals:**

- A lit area is a term in the light map, not an exception threaded through the lighting code: it is added into the same map in the same pass as a light, so nothing downstream has to know it exists.
- No new render target, no new shader, no extra full-screen pass (one 2×2 texture is all the lit areas add).
- The geometry of the feather is pure arithmetic, so it is tested without a GL context; only the draw calls need a window.
- The editor reuses `rect_edit.go` rather than growing a second rectangle editor.

**Non-Goals:**

- Changing how a light is computed, how the composite works, or what `lightGlow` is.
- Making a lit area interact with the shadow passes (it never becomes an occluder, and it never removes one).
- Per-side feathers, or a choice of falloff curve: the sides ramp linearly and the corners fade as the product of their two distances, and neither is authorable.

## Decisions

### A lit area is white added into the light map, feathered by alpha

`rl.BlendAdditive` is `glBlendFunc(GL_SRC_ALPHA, GL_ONE)`: the source colour is scaled by its own alpha before being added. So a white rectangle drawn additively at alpha `a` lifts the map by `a`, and a rectangle whose alpha ramps from 1 to 0 across its last pixels is exactly the feather the spec asks for. The areas are drawn in `computeLights`, inside the same `BeginTextureMode(LightAccum)` / `BeginMode2D(cam)` / `BlendAdditive` shape the lights use, after the ambient clear.

Because the lift is additive:

- an area can only brighten — a light already at full brightness inside one stays exactly as bright, which is what the spec requires;
- the order of lights and areas in the pass does not matter, so areas can be drawn in one batch before or after the lights;
- an area over a shadow raises that shadow to white, which is the whole point: the shadow is in the map, and the area overwrites it by saturation rather than by being subtracted out anywhere.

Alternatives considered. *Drawing the area with `rl.BlendAlpha`* would set the map to white and lerp at the feather — but it would also **replace** a brighter light inside the area with plain white, dimming it; rejected on that alone. *Marking a rectangle as "no shadow" inside the shadow passes* — skipping occluders whose rectangle falls in an area — was the other reading of the request; it was put to the user and they chose full brightness instead, and it would also cost per-light work and still leave the ambient dark in place.

### The feather is nine disjoint pieces: five gradient rectangles and four corners from one 2×2 texture

An area with effective feather `f` is drawn as: the core inset by `f` at full white; four side strips ramping from transparent at the outer edge to white at the core; four corner squares of side `f`, lit at the inner corner and transparent at the other three. The core and the sides are `rl.DrawRectangleGradientEx(rec, topLeft, bottomLeft, bottomRight, topRight)`.

The corners cannot be. A gradient quad is rasterised as two triangles split from its top left to its bottom right, and each interpolates its own three vertices: a corner lit at the top left or the bottom right — on that diagonal — comes out as `1 − max(x, y)`, but a corner lit at the top right or the bottom left leaves one triangle with three dark vertices, so half the corner goes black along the diagonal. Rendered, that is a hard wedge across two of every area's four corners. (This was found by rendering the composite during apply; the design had assumed all four corners behaved like the two on the diagonal.)

The corners are therefore drawn from `cornerFade`, a 2×2 texture lit at one texel and clear at the other three, sampled from that texel's centre to the far one's (source `{0.5, 0.5, ±1, ±1}`, the sign putting the lit texel at the corner the piece carries its light in) and filtered bilinearly, so the fade is computed per pixel rather than per triangle. Its clear texels are white at zero alpha, not blank: filtering a blank texel would drag the colour towards black as well as the alpha towards nothing, and additive blending multiplies the two, squaring the falloff. The corner's fade is the product of the two distances rather than the smaller of them — a rounded pool of light rather than a square one — which meets each side's ramp exactly along the axis where they join.

Alternatives: a fragment shader computing the distance to the rectangle would give a choice of curve, at the cost of a shader, a uniform set and a state change per area; an n-patch cannot follow a per-area feather, because raylib's n-patch borders are fixed in source pixels. One 2×2 texture created beside `LightAccum` costs neither.

The pieces are produced by a pure function over the area — `areaPieces(LitArea) []gradRect` — so the geometry, the inset and the feather clamp are unit-tested headlessly, and the draw is a loop that sends a corner to the texture and everything else to a gradient rectangle.

### The feather is clamped to half the shorter side

A feather wider than half the area would otherwise leave the core inverted and the middle dim. The effective feather is `min(Feather, min(Width, Height)/2)`, computed at draw time rather than clamped in the data: a small area nudged larger gets its authored feather back instead of having lost it.

### Lit areas cost nothing off screen, exactly as lights do

Each area is tested against `viewRect(cam, ...)` with `aabb`, the same test `lightVisible` does, and skipped when it is out of view. Because the rectangle bounds the feather (the fade is inside the area), the area's own rectangle is the whole test.

### The data is a rectangle plus a feather, in whole level pixels

```jsonc
"litAreas": [ { "x": 320, "y": 96, "w": 128, "h": 64, "feather": 16 } ]
```

Positions and sizes are `int32` level pixels, like the tiles and the lights, so an area lands on the pixel grid the game draws on. The key is written only when the level has areas, so a level with none keeps the file it has today; `parseLevel` ignores an unknown key already, so an older build reading a newer file is unaffected.

`marshal()` — the fingerprint — stays the tiles alone, so lit areas are out of it for free, like the lights and the ambient. `parseLevel` grows a fourth return value; `LoadLevel` and the existing level tests are the call sites.

No strength and no colour: a lit area either lifts the map to white or is not there. A dimmer or tinted lift is what a light is for, and one number that is always 1 in the file is one number to maintain. (Recorded in proposal.md as an assumption.)

### The area tool is the fourth tool, and reuses the rect tool's dragging

`toolArea` goes after `toolLight` in the Q rotation, so the two lighting tools are neighbours and `nextTool` stays `(t + 1) % toolCount`. The editor gains `areas []LitArea`, `areaSel int` and an `areaDrag rectDrag` — the same `rectDrag` the rect tool uses, because `rectHandleAt` / `handleAnchor` / `dragRect` work on a bare world-space `rl.Rectangle` and enforce the "never thinner than a pixel, never inside out" rule already. Hit testing walks the areas back to front, so the area drawn on top is the one a click takes.

The tool's mouse handling mirrors `updateLightTool`, which already has the shape the spec describes (left click takes hold or places, right click removes or deselects, Delete removes the selected one, a new one inherits the selected one's parameters). The panel mirrors `drawLightPanel`: `drawIntFields` for X, Y, W, H and Feather when an area is selected, a count and a hint when none is.

### The markers are drawn with the editor's overlay

`DrawOverlay` already draws the light markers over the lit scene, and for the same reason the area markers belong there: an area is full brightness, and an outline drawn under it would be washed out. The markers are the area's rectangle and a second rectangle inset by the effective feather, so the extent of the fade is visible, drawn as filled rectangles a level pixel wide (`DrawRectangleLinesEx`) rather than GL lines — a GL line is one screen pixel at any zoom and the palette pass, sampling in blocks, would keep only scattered dots of it. Selected white, the rest faded, as the lights are.

## Risks / Trade-offs

- **The glow pass brightens a lit area by about 12%** → `compositeLights` adds `map × 0.2` after the multiply, so where the map is white the frame gets `scene + 0.2` instead of `scene + 0.2 × ambient`. An area therefore reads as slightly hazy, exactly as standing under a very bright light does — consistent with the rest of the lighting rather than a bug, and the areas are authored against what is on screen. If it turns out to be too much, the fix is a per-area strength below 1, which the file format can grow without breaking anything.
- **One filtered texture in a point-filtered game** → `cornerFade` is the only bilinear texture the game draws. It is 2×2 and never shows a texel of its own — only a ramp of alpha — so nothing about the pixel-perfect look changes; but a future pass that reaches for "set every texture to point" has to leave it alone.
- **Banding in the feather** → the palette pass quantizes to 64 colours and `ditherStrength` is 0, so a wide feather over a flat surface may step. Mitigation is authorial: narrower feathers, or an area placed where the falloff crosses texture detail. Dithering the gradient is out of scope here for the same reason it was in `light-map-composite`.
- **An area is a blunt instrument** → it lights everything in its rectangle, including a tile that was meant to stay dark and a fighter passing behind it. That is the feature as asked for; the mitigation is that areas are cheap to move and shape in the editor.
- **A fourth tool on one key** → Q now cycles through four tools, which is one more press to get back to placing tiles. Keeping the lighting tools adjacent means the two lighting tasks are one press apart, which is where the authoring happens.
- **`parseLevel`'s signature grows again** → a fourth return value is one more thing to thread through; it stays a plain function with the tests it already has, and the alternative (a level struct) is a refactor this change does not need to carry.

## Migration Plan

No migration. A level file with no `"litAreas"` key loads with none; the key is written only once a level has areas; the fingerprint is unchanged either way, so a build with areas and a server without them still agree on the level. Rolling the change back leaves the key in the file, which an older build skips with a trace warning, exactly as it does for any key it does not know.

## Open Questions

- The default size and feather of a freshly placed area (a rectangle around 64×64 with a feather around 16) are a starting point to settle by eye during apply, like `defaultLight`'s radius.
