## Context

See proposal.md — Why. What shapes the approach is how a level frame is drawn today, in `runLevel`:

1. Into `ScreenFullTexture`, inside `BeginMode2D(Cam.View())`: the level tiles, then the draw list of actors sorted by x.
2. Still inside that texture: a full-screen black rectangle at alpha 50 that darkens everything, then the three lights in additive blend.
3. `ScreenFullTexture` is drawn into `VirtualScreen` through the palette shader, which snaps every pixel to the nearest `Palette` colour in blocks of `Cam.PixelSize()` screen pixels.
4. After the shader: `drawSpectatorHUD` and `Cur.DrawGUI`, at screen resolution, unquantized.

So anything drawn with the actors in step 1 is darkened by step 2 and brightened by whichever light it stands in. Anything drawn in step 4 is off the pixel grid and in unrestricted colour.

Two more facts:

- `Actor.drawPos()` rounds the interpolated position to whole world pixels. It is what the sprite is drawn at, for local, bot and remote fighters alike.
- `HitCount` is simulation-only. `ActorRec` carries `Structure` but not the hit count. `ActorRec.Flags` is a `uint16` with 8 of its bits in use.

## Goals / Non-Goals

**Goals:**

- A bar that looks like part of the pixel-art scene — on the grid, zooming with it — without taking on the scene's lighting.
- The visibility rules as one small function that tests can call without a window.
- Pips that work for remote fighters without growing the snapshot.

**Non-Goals:**

- Any screen-space HUD, and any animation of the bar (proposal — Non-goals).
- Changing the draw order of anything already drawn.

## Decisions

### Draw the bars in their own pass, after the lights, still inside the scene texture

After the additive light pass and before `EndTextureMode`, open `BeginMode2D(Cam.View())` again and walk the draw list once more, drawing each fighter's bar. The pass is still inside `ScreenFullTexture`, so bars go through the palette shader and the camera's pixel snapping like the sprites do. It comes after the darkening rectangle and the lights, so lighting never touches it.

*Alternative — draw the bar inside `Actor.Draw`:* rejected. Lighting would then change the bar's colour with position, and a neighbouring fighter drawn later in the list could cover it.

*Alternative — draw in screen space after the palette shader, like the spectator HUD:* rejected. The bar would sit off the pixel grid at every non-integer position, and in colours that are not in the palette — in a scene that is otherwise rigorously both.

### Geometry in whole world pixels, symmetric fill

The bar is 16 × 2 world pixels, centred on the fighter's collision box horizontally (`drawPos().X + CollisionBox.X + CollisionBox.Width/2`), with its bottom edge 3 world pixels above the top of the collision box. The fill is drawn as two halves growing out from the centre. Each half is `round(8 × lost/12)` pixels wide, so the fill is always an even width and symmetric, and never lands on half a pixel.

Everything is computed in whole world pixels from `drawPos()`, so the camera's integer zoom turns every edge into a whole number of screen pixels. The pixel-snapping work already guarantees that for anything drawn at integer world coordinates.

### Colours from the palette

Track `(39,35,36)`, fill `(234,50,60)`, pips `(242,221,161)` — all three are exact `Palette` entries. The palette shader maps a colour to its nearest palette entry, so an exact entry comes through unchanged, and the scenario "identical colours under a lamp and in shadow" holds by construction rather than by luck.

### Pips as 2 × 2 squares in a row centred above the bar

Up to two, 2 pixels apart, in a row centred on the bar with its bottom 1 pixel above the bar's top. Squares rather than a second bar, so they read as a count rather than a quantity. The gap is 2 rather than 1 so that both possible rows — 2 and 6 pixels wide — centre on a whole pixel over a 16-pixel bar.

*Revised during implementation — first placed to the right of the bar:* seen in play and rejected. That row reached up to 23 world pixels right of the fighter's centre, while fighters trading blows stand about 20–24 pixels apart, so one fighter's pips landed on the other fighter's bar at ordinary fighting range. Above the bar and never wider than it, the pips cannot reach a neighbour's bar before the bars themselves would overlap.

### Visibility is a pure predicate

`showStructureBar(a)` returns whether to draw: structure below `MaxStructure`, state not `StateDeath`, and time manipulation not showing a recorded frame (`TimeShowsHistory()`). The draw pass calls it, and unit tests call it directly. Pip count is `a.HitCount`, clamped to `StunEveryHits-1` so a stale or corrupt value cannot draw a third pip.

### Hit count in two spare flag bits

The count is 0–2 and needs two bits. It goes in bits 8–9 of `ActorRec.Flags` — the next free bits after `fClimbLeft` — set in `recOf` and read back in `apply`. The record stays 42 bytes. The protocol version still goes to 4: a version-3 peer would otherwise silently ignore the bits and draw no pips, and the version bump turns that into the clear "version mismatch" the handshake already gives.

*Alternative — a new `uint8` field:* workable, but it grows every actor record by a byte to carry two bits, when the flag word already has eight spare ones.

## Risks / Trade-offs

- **Overlapping fighters overlap their bars.** Two fighters standing in one spot draw two bars in the same place. → Bars are drawn in draw-list order, so the nearer fighter's bar is on top, which is the one the sprite order would suggest. Stacking them is a follow-up if it shows up in play.
- **At zoom 1 the bar is 16 × 2 screen-virtual pixels.** It is small, but it is the same size relative to the fighter at every zoom, and the default zoom is 3. → Judge it in play; the size is two constants.
- **Pips lag a round-trip on remote fighters.** They come from the snapshot like everything else a client shows. → Consistent with structure and hits, which already behave this way.
- **The debug overlay's structure text sits under the fighter, and the bar sits over its head.** They do not collide, but they show the same thing twice while debug is on. → Harmless; the debug line also shows the exact value.

## Migration Plan

Nothing is persisted. Protocol version 4 means client and server builds move together, and mismatched builds are rejected with "version mismatch". Reverting is a single revert.
