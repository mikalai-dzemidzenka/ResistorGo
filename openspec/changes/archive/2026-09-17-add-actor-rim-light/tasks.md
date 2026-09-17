## 1. Which lights rim a fighter

- [x] 1.1 In a new `rim_light.go`, add `rimStrengthAt(l LevelLight, p rl.Vector2) float32` that returns `intensity × (1 − d/R)²`, clamped at 0, which is the light shader's falloff. Verify with a test in `rim_light_test.go`: full intensity at the centre, a quarter at half the radius, zero at and beyond the radius, and zero for zero intensity.
- [x] 1.2 Add `segmentHitsRect(a, b rl.Vector2, r rl.Rectangle) bool` (slab test). Verify with tests: a segment crossing a rectangle hits, one passing beside it or stopping short does not, one starting inside hits, and axis-parallel segments work (no division by zero).
- [x] 1.3 Add `rimLights(lights []LevelLight, occluders []rl.Rectangle, body rl.Rectangle) []rimLight`, which measures at the body centre, drops lights below `rimMinStrength` or blocked by any occluder, and keeps the `rimMaxLights` (4) strongest in order of strength. Verify with tests covering the spec scenarios "Out of reach", "Behind a wall", "Clear line", "Too many lights" (six lights in reach give the four strongest), and that a zero-intensity light is dropped.

## 2. Lights in the sprite's own coordinates

- [x] 2.1 Add `rimLocal(light, spritePos rl.Vector2, frameWidth float32, facingLeft bool) rl.Vector2`, which returns `light − spritePos`, with x mirrored to `frameWidth − x` when the fighter faces left. Verify with tests: a light left of a right-facing fighter has negative local x, and the same light left of a left-facing fighter has local x greater than `frameWidth`, so it lies on the frame's right in sheet space ("Turning around").
- [x] 2.2 Move `Draw`'s time-history rectangle into `spriteDraw`, so that while `TimeShowsHistory()` it returns the recorded animation, the recorded frame, height 48 and `drawPos()`. Make `Draw`'s history branch use `spriteDraw`. Verify `go test ./...` still passes, the attack-outline tests included, and check by reading that the history draw is unchanged.

## 3. The shader

- [x] 3.1 Add `fsRim` to `shaders.go`. It snaps to the texel centre in frame space, discards non-body texels, and reads the four neighbours through a frame-clamped `alphaAt`. It forms the outward normal and discards interior texels. It sums `lightColor[i] × max(0, dot(n, normalize(lightPos[i] − texel)))` over `lightCount` (at most `rimMaxLights`) and outputs `vec4(sum, 1)`. Verify `go build ./...` succeeds and the shader compiles at startup with no raylib shader error in the log.
- [x] 3.2 Load it as `RimShader` in `runLevel` next to `OutlineShader`, and look up its uniform locations once into package vars. Verify the game starts and a level loads with no warning about missing uniform locations.

## 4. Drawing it

- [x] 4.1 Add `drawRim(a *Actor, lights []LevelLight, occluders []rl.Rectangle)`. It skips `deathHidden()` fighters and empty textures, calls `rimLights` with the body from `outlineRect()`, and returns early when no lights are left. It converts each light with `rimLocal` using the sprite position and frame from `spriteDraw`, multiplies the colour by strength × `rimStrength`, sets `texelSize`, `frameBounds`, `lightCount`, `lightPos` and `lightColor` (arrays via `SetShaderValueV`), and draws the sprite through `RimShader`. Verify `go build ./...` succeeds.
- [x] 4.2 In `runLevel`, after `compositeLights()` and `BeginMode2D`, and before the attack-outline loop, draw every fighter in `drawList` with `drawRim` inside `rl.BeginBlendMode(rl.BlendAdditive)`. Compute `LE.LightOccluders()` once and share it with `computeLights`. Verify `go build ./...` succeeds and, by reading, that the outlines, bars and editor overlay are still drawn after it.
- [x] 4.3 Tune `rimStrength` and `rimMinStrength` by eye in `level_1` so the rim is clearly visible near a lamp, fades toward the radius, and does not blow out to white under a bright light. Verify with before and after screenshots of the same spot.

## 6. The rim moves outside the sprite, around body and top layer

- [x] 6.1 Turn `fsRim` into an outer ring: keep the texels the fighter does not draw, discard the ones it does, take the direction towards the body from the neighbours' alpha and light the texel by how squarely the opposite direction faces each light. Verify `go build ./...` succeeds and the shader compiles at startup.
- [x] 6.2 Draw the rim on a quad one texel larger than the frame on every side, source rectangle grown with it and the left-facing mirror preserved, so the ring is not cut off where the sprite meets the frame's edge. Verify by eye against a pose whose head or feet touch the frame edge ("Against the frame's edge").
- [x] 6.3 Take the union of the body sprite and `AnimTop` as the silhouette: a second sampler with its own bounds and texel size, switched off when there is no top layer, and `alphaAt` returning the greater of the two. Verify `go build ./...` succeeds, the shader compiles with the sampler unused, and the rim on a fighter with no top layer is unchanged.
- [x] 6.4 Add `Actor.topDraw()` returning the top layer's texture and source rectangle for this frame, or nothing when `AnimTop` is nil, and use it in `drawRim`. Verify with a test that it reports nothing when `AnimTop` is nil and the matching frame when it is set.
- [x] 6.6 Find where a fighter is really drawn in two layers: `DrawAnim` draws the walk's arms from `walk-bot`/`walk-mid`/`walk-top` by guard height, which the rim was ignoring. Put the rule in `Actor.topAnim` (`AnimTop` first, then the walk), have `topDraw` build the draw from it, and have `DrawAnim` draw its second layer through it so the two cannot drift. Verify with a test over the stances, block-all and `AnimTop` winning, and that `go test ./...` still passes.
- [x] 6.7 Re-render a walking fighter at each guard height and confirm the rim follows the arms it is showing rather than the body alone ("Walking with a guard up").
- [x] 6.5 Re-check the rendered scenarios after the move outside: one lamp, two coloured lamps, overhead, turned around, shadowed, lit area, rewind, the attack outline over a rimmed fighter, and a hidden dead fighter. Verify each still behaves as the spec says.

## 5. Whole-change verification

- [ ] 5.1 Run `go test ./...` and `go test -tags release ./...` and confirm both pass.
- [x] 5.2 Check `level_1`'s lights against its tiles' light rectangles: list any light whose centre is inside a light rectangle (it will never rim). Confirm with the user that this is intended, or move the light. Verify the list is recorded in this task when it is checked off. **Checked: no lamp in `level_1` has its centre inside a light rectangle, so the list is empty and nothing had to move.**
- [ ] 5.3 Play a local match (`go run ./cmd`) and confirm by eye: the rim is on the lamp side only ("Light from one side"); a lamp overhead lights head and shoulders ("Light from above"); the rim stays on the lamp side when the fighter turns; it traces each animation frame; it fades walking away and is gone outside the radius; it disappears behind a wall that casts a shadow; two differently coloured lamps rim two sides; a lit area alone gives no rim; the silhouette does not grow.
- [ ] 5.4 Confirm in play that the rim is readable in the dark at the edge of a light's radius, that the red attack outline and the structure bars look unchanged on a rimmed fighter, that the rim follows the recorded pose while rewinding (hold R), and that it is not drawn while a dead fighter is flickered out.
- [ ] 5.5 Play a networked match against the dedicated server (`go run ./cmd -dedicated` plus a client, and a spectator). Confirm both fighters are rimmed for both players and for the spectator, and that the dedicated server still runs headless without loading the shader.
- [x] 5.6 Rebuild the deliverable with `build.ps1` (UPX included, output not truncated). Confirm `resistor.exe` is replaced and starts.
