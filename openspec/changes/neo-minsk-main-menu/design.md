## Context

See proposal.md — Why. How the menu is drawn today, in `runMainMenu`:

```
MainMenuTexture (320x180, point filter)                    scene pixel = 4 screen px
  +-- perspective camera: the red line fan (the "background")
  +-- orthographic camera, 18 scene px / unit: head, jab1, jab2, kick, martelo
        red wireframes at x 0.7 .. 7.1 units, all right of the 560 px panel
VirtualScreen (1280x720)
  +-- MainMenuTexture upscaled x4, tinted rl.Red
  +-- logo (resistor.png), drawMenuUI() raygui widgets
present() --> window (letterboxed)
```

- Nothing in the menu is palette-limited. The level's palette pass — `PaletteShader` with `ScreenPaletteTexture` built from the 27-colour `Palette` — runs only inside `runLevel`. Both are loaded lazily in `runLevel`'s `levelInit` block, so they do not exist yet when the menu first draws.
- `fsPalette` quantizes every pixel to the nearest entry of an N×1 palette texture, N ≤ 256, in blocks of `pixelSize` screen pixels. It takes any palette, so it can serve the menu as-is.
- `drawMenuUI` draws raygui widgets into whichever render texture is active. Mouse coordinates are already mapped to the 1280×720 virtual screen by `letterbox()`, so the draw target does not change hit-testing.
- `MainMenuTexture` is shown with `DrawTextureEx`, which does not undo a render texture's vertical flip. The existing 3D model pass compensates with a negative Y scale. Anything else drawn into that texture has to account for the same flip.
- The still-active `add-menu-music-and-options` change has an unsynced delta to `main-menu` that modifies *Pixel-art menu scene*. Its code has shipped. This change's delta is written against that version.

## Goals / Non-Goals

**Goals:**

- A layered scene in which the background is painted once and the foreground is drawn over it, so the existing layout guarantees for the models, logo and widgets hold without being re-proved.
- One mechanism that guarantees "every menu pixel is Endesga 64", instead of hand-picking colours and hoping fades and blends stay in palette.
- A single list that decides which palette entries the scene may use, so "the scene is red" is one testable fact rather than a habit spread over the drawing code.
- Pure, window-free functions for everything that can be tested: the palette table, the red set, the skyline layout and the colour cycle.

**Non-Goals:**

- The level's palette or look (proposal — Non-goals).
- Any change to the layout constants, the widgets or the menu screens.

## Decisions

### Three render targets for the menu

```
MenuBackground (320x180, point filter)
  sky bands + sun --> Neo-Minsk skyline --> perspective neon grid
        |
        v
MainMenuTexture (320x180)
  background --> models (orthographic 3D)
        |  upscaled x menuPixel
        v
MenuFrame (1280x720)
  scene --> logo --> drawMenuUI() widgets
        |  drawn through PaletteShader with the Endesga 64 table, pixelSize 1
        v
VirtualScreen --> present()
```

The background is painted into its own texture rather than straight into the scene, because it is drawn with ordinary top-down 2D coordinates while `MainMenuTexture` is displayed flipped. Drawing the finished background into the scene with a positive source height applies that flip once, in one place, instead of spreading it through the sky, city and grid code.

*Revised during implementation — the background first rolled and swayed, Hotline Miami style:* removed at the user's request. With it went the overscanned background texture, the roll and sway curves, and the margin calculation that kept a rotated texture covering the screen. The background is now exactly scene-sized, and `MenuBackground` is drawn at the origin with no rotation.

### The palette guarantee is one shader pass

`MenuFrame` is drawn into `VirtualScreen` through the existing `fsPalette`, with `pixelSize` 1 and dithering off, using an N×1 texture of the Endesga 64 colours. Every pixel of the frame, fades and text anti-aliasing included, is snapped to its nearest palette colour. The scene still draws with deliberate palette colours, so the snapping is a safety net and not the look.

`PaletteShader` and a new `MenuPaletteTexture` are loaded once at start-up (in `initPresent`), not in `levelInit`. The level's own palette texture is untouched, so the level keeps its colours.

*Alternative — only ever draw with palette colours:* rejected as the sole mechanism. Model fade-in alpha, line anti-aliasing and raygui's text rendering would all produce off-palette pixels, and the spec says every pixel.

### Endesga 64 as a table in `palette.go`, and one red set beside it

`EndesgaPalette []rl.Color` sits beside `Palette`. Its 64 values come from the published Endesga 64 list on Lospec, copied in from the downloaded file rather than typed. A unit test pins that there are exactly 64 colours, all distinct, and that the first and last entries match the published list.

`menuRedIndexes` then names the entries the scene may use. Every colour the scene paints with — sky bands, sun, silhouette, ground, both colour ramps, and the model colours — is taken from that one list. A test checks each of those is in the list, and that every entry in the list has a hue within 20 degrees of pure red, which is the spec's rule. The scene cannot drift off-red without failing a test.

### A procedural, deterministic skyline

`menu_city.go` builds the skyline from a fixed layout: the five landmarks at chosen x positions across the 320-px width, with filler buildings between them from a seeded generator. It returns plain rectangles and window positions, and a separate function draws them. The layout is pure and tested for four things:
- every landmark is present, inside the viewport and not covered by another building;
- building tops stay below a sky margin, so the dusk sky and sun show;
- the same seed gives the same city;
- the base sits on the horizon.

Landmarks as silhouettes, in scene pixels:

```
National Library        TV tower          Hero City obelisk
     /\                     |                    *
    /  \                   (o)                   |
   <    >   cut-gem        |                    /|\
    \  /    on a plinth    |  mast on a        / | \   needle on
     \/                    |  thin shaft          a stepped base
   [____]                 _|_
Gates of Minsk (twin towers, stepped tops)      Independence Ave blocks
   _^_   _^_                                    _______   _______
  |   | |   |                                  |  |||  | | ||| |  long
  |   | |   |                                  |  |||  | | ||| |  colonnaded
```

*Revised during implementation — the silhouettes first carried neon outlines:* removed at the user's request. The buildings are filled shapes only, and the lit windows are the sole light on them. The outline colour role went with them, leaving the grid and the windows as the two cycling roles.

### Colour cycling steps along red ramps

The two roles are the grid and the lit windows. Each takes its colour from a short list of `menuRedIndexes` entries, advancing one entry every fraction of a second. A role's colour is always an entry of the list, so cycling never produces an off-palette or off-red colour even before the shader pass. The cycle step is a pure function of time, and a test checks it only ever returns palette reds.

### The grid floor is drawn in 2D, by 1/z

Cross lines sit at `y = horizon + depth/(k - frac)`, where `frac` is the fractional part of `t · speed`. Placing them by `1/z` is what makes them spread apart as they approach the viewer, and the scroll repeats seamlessly because row `k` at the end of a step lands exactly where row `k-1` started. Depth lines fan out from the vanishing point on the horizon. Drawing it in 2D alongside the sky and the city keeps the whole background in one coordinate system.

## Risks / Trade-offs

- **Widget legibility over a busy city.** raygui buttons are opaque, but labels sit on the scene, and the city runs behind the panel. → Seen in play: the Options screen's labels are readable but low-contrast against the city. If it bothers, a dim band behind the panel is a small addition within the same layout.
- **raygui's dark style shifts slightly** when its colours snap to Endesga 64. → Accepted in the proposal. The palette has a full grey ramp, so text stays readable.
- **A red-only scene has less separation than a multi-hue one.** The city, the grid and the models all sit in the same hue family, so they are told apart by brightness alone. → The silhouette is near-black maroon, the windows and grid are bright crimson; checked in play.
- **The Endesga 64 values are copied in by hand.** → The count, uniqueness and endpoint test catches a bad paste; the values themselves come from the downloaded palette file, not from memory.

## Migration Plan

Nothing persists, and the level is untouched. The change reverts cleanly.

## Open Questions

- The cycle speed and the exact ramp entries are tuning. They are constants to settle by eye in play and do not change the approach or the tasks.
