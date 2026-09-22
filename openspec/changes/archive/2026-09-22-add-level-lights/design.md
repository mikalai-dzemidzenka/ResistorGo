## Context

See `proposal.md` — Why. What shapes the approach:

- **Lights today.** `game.go` holds three package-level `*ZimaLight`s, created in `runLevel`'s first-frame init as `newZimaLight(512, r, rl.White, 1.0)` with r = 256, 128, 128. Each frame the render loop computes each at a hardcoded world position against `LE.LightOccluders()` — called three times — and later, inside the scene's texture pass, after a translucent black overlay, draws each result additively with `addLight(tex, pos, radius)` in camera space.
- **What a light is.** `ZimaLight` has `Size` (render-texture resolution), `Radius` (world size), `Color`, `Intensity`, and unused `Direction`/`ConeAngle`. `Compute` sets its own camera from `lightPos` and `Radius` (zoom = Size / 2·Radius), runs caster → distance → distort → a reduction chain of log2(Size) passes → shadow, and returns the shadow texture. The shadow shader already takes `lightColor` built from `Color` × `Intensity`; `addLight` then draws with a white tint. So color and intensity already work per light — they only need setting per light.
- **What a light costs.** Each `ZimaLight` owns about a dozen render textures at 512 (the four full-size ones plus the reduction chain), roughly 5 MB, and a dozen passes a frame. A `ZimaLight` per placed light would make memory grow with the level.
- **The level file.** `level_1.json` is a JSON object keyed by layer number, unmarshalled straight into `map[int]LevelLayer`. `marshal()` serialises `le.objects` and is used both for saving and for `Hash()`; the fingerprint is compared in the client's hello and the server's handshake. The dedicated server loads the same file through `LE.LoadLevel` and never renders.
- **The editor.** Two tools (`toolPlace`, `toolRect`) switched by `Q` as `1 - le.tool`; a panel drawn by `drawPanel` whose numeric fields go through `drawRectFields` and `nextEditField`, both sized for exactly four fields; `grabTolerance(zoom)`; one `dirty` flag and `Save()`.

## Goals / Non-Goals

**Goals:**

- Lights are level data, loaded everywhere the level is, and today's three become level data without the level looking or fingerprinting any differently.
- Memory for lighting stays constant in the number of lights; lights wholly off screen cost nothing.
- The light tool reuses the editor's machinery — tolerance, dirty flag, save, panel fields — rather than growing a parallel copy.

**Non-Goals:**

- Spotlights, animated lights, lights attached to tiles, copy and paste of lights.
- A cap on visible lights, or a cheaper light technique.
- Changing the shadow algorithm, its resolution, or the look of a light.

## Decisions

### Lights are one extra key in the level file; the fingerprint stays over the tiles

`LevelLight` is `{X, Y int32; Radius int32; Color [3]uint8; Intensity float32}`, saved as `{"x":…, "y":…, "radius":…, "color":[r,g,b], "intensity":…}` — whole pixels, as tiles are. `LevelEditor` gains `lights []LevelLight`.

The file stays the layer-keyed object it is, with one more key, `"lights"`, holding an array. Loading reads the object as `map[string]json.RawMessage`: `"lights"` goes to `le.lights`, every key that parses as an integer goes to that layer, anything else is logged and skipped. Saving writes the layers and `"lights"`. `marshal()` — the fingerprint's input — is left as it is, over the layers alone.

*Alternative considered:* wrapping the file as `{"layers": {...}, "lights": [...]}`. Rejected: it rewrites every line of every level for no gain, and the fingerprint would then need its own serialisation to stay unchanged. *Alternative considered:* including lights in the fingerprint. Rejected: the fingerprint exists to stop two peers simulating different geometry, and a light is not geometry. A light rectangle is also only visual, but it lives on the tile and moving it out of the fingerprint would change today's fingerprint, which this change must not do.

### One shared set of shadow textures, and an accumulation buffer

`runLevel`'s init creates one `ZimaLight` at 512 and one `LightAccum` render texture the size of `ScreenFullTexture`. Per frame:

1. Compute the occluders once.
2. Clear `LightAccum` to black.
3. For each light that passes the visibility test: set the shared light's `Radius`, `Color` and `Intensity` from it, `Compute`, then draw the result into `LightAccum` with additive blending inside `Cam.View()` — the same `addLight` call as today, into a different target. This has to happen per light, straight after its `Compute`, because the next `Compute` overwrites the shared textures.
4. In the scene pass, where the three `addLight` calls are now, draw `LightAccum` once, additively, in screen space.

Additive blending into an 8-bit target saturates, and it does so identically either way: adding each light to the scene gives min(s + l₁ + l₂, 1), and adding the clamped sum gives min(s + min(l₁ + l₂, 1), 1), which are equal. So the result matches drawing each light straight onto the scene, while memory is one light's textures plus one screen-sized buffer however many lights the level has. `Size` stays 512 for every light: the migrated lights look exactly as they did, including the two radius-128 lights that have always had 512 texels.

*Alternative considered:* a pool of `ZimaLight`s grown to the number of visible lights. Rejected: about 5 MB each, for no visual difference.

### Culling against the camera's world rectangle

`lightVisible(l LevelLight, view rl.Rectangle) bool` tests the light's square (center ± radius) against the world rectangle the camera shows: the corners of the 1280×720 scene target through `rl.GetScreenToWorld2D` with `Cam.View()`. It is a pure function of the two rectangles and is unit tested. Partly visible lights are drawn whole, so nothing pops at the edge of the screen.

### The three built-in lights move into `level_1.json`

The change adds to `level_1.json` a `"lights"` array holding (40, 30) r 256, (120, 120) r 128 and (−40, 30) r 128, each white at intensity 1, and removes `light1..3`, `pos1..3` and their `newZimaLight` calls. Since the embedded level and the disk level are the same file, the release build and the dedicated server get them too. Tiles are left byte-for-byte alone, so the fingerprint is unchanged — pinned by a test comparing `Hash()` with the fingerprint of the file with its `"lights"` key removed.

### A third tool, `toolLight`

`Q` cycles place → rect → light → place, replacing `1 - le.tool`. The light tool's state is its own: `lightSel int` (−1 for none) and a `lightDrag` holding the mode (move or radius) and the grab offset. It never reads or writes the tile selection, and the other two tools never read or write it.

Pure functions carry the testable part:

- `lightAt(p, tol)` — a click on the radius handle (the point at center + (radius, 0)) is tried first, then the center within `max(tol·2, 2)`.
- `placeLight(p)` — inherits the selected light's radius, color and intensity, else radius 128, white, intensity 1.
- `dragLight(to)` — move, or radius = max(1, round(to.X − center.X)).
- `deleteLight(i)`.

All of them set `dirty`. Clicks go through `overGui` like the other tools. A left click on nothing places; a left click on a light selects and starts a drag; a right click on a light deletes it; Delete deletes the selection. `Save()` already writes the level file, which now carries the lights.

### The panel generalises to N fields

`drawRectFields` and `nextEditField` become `drawIntFields(rows, fields)` over a slice of `{label, value, min, max}` and `nextEditField(editing int, toggled []bool)`. The rect tool uses them with four fields and the light tool with seven: X, Y, radius, R, G, B, and intensity in percent. Integer boxes are what raygui's `ValueBox` offers, and a percentage keeps intensity exact. `nextEditField`'s existing tests are moved to the slice form.

### Markers

`LE.DrawOverlay`, in edit mode with the light tool, draws each light's circle at its radius, a small cross at its center, and the radius handle. The selected light is drawn in white, the others in their own color at half alpha.

Two things changed from the first plan once the markers were rendered:

- **They are drawn over the lighting, not under it, and with a dark edge.** They are drawn after the lighting, alongside the attack outline and the structure bars. A light is brightest exactly at the center its cross marks: saturated white for a white light. Drawn under the lighting, each light's own glow washed its cross out. Even on top, a white cross vanished into it. So the cross and the radius handle each sit on a thicker, mostly opaque black backing that shows on the glow and on the dark alike.
- **They are filled shapes a level pixel wide, not GL lines.** The circle is `DrawRing` and the cross `DrawLineEx`, rather than `DrawCircleLinesV`/`DrawLineV`. A GL line is one screen pixel wide at any zoom, and the palette pass samples the frame in blocks of the zoom, so it kept only scattered dots of one. The tiles' debug rectangles use `DrawRectangleLinesEx` for the same reason.

## Risks / Trade-offs

- **A crowded view is slow.** Each visible light is a dozen full-resolution passes. → Culling handles the common case of a level spread out beyond the screen. There is no cap; if a scene ever needs many overlapping lights, lowering `Size` for small radii is the next lever, and it is deliberately not taken here because it would change how the migrated lights look.
- **The accumulation buffer changes the pipeline.** → The saturation argument above makes it equivalent, but it is still the one place the look could shift, so it is checked by eye against a screenshot taken before the change, from the same spot.
- **Older development builds cannot read a level with lights.** → Unmarshalling a non-integer key fails in the old loader. Every build carries its own level, so this only bites when checking out an older revision with a newer level saved on disk; the old build logs the error and loads nothing rather than crashing.
- **Lights stay out of the fingerprint, so two peers can see different lighting.** → Intended: it only happens between development builds with different unsaved or uncommitted light edits, and never changes play.
- **This change restates a dev-tools requirement the in-flight editor change introduces.** → It must be archived after `add-editor-textures-and-rect-editing`, as `reuse-object-rects` must be.

## Migration Plan

The change edits `level_1.json` to add the three lights and removes the hardcoded ones in the same commit, so no revision has both or neither. The fingerprint is unchanged, so builds before and after this change can still play together. Rollback is reverting the commit, which also removes the `"lights"` key the older loader could not read.
