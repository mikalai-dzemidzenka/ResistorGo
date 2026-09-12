## Context

See proposal.md (Why) for the bug. The pipeline in `main_menu.go` before this change:

1. `MainMenuTexture` is a 160x90 (`SW`x`SH`) render texture with point filtering.
2. The background fan is drawn with a perspective camera (fovy 120). The wireframes are drawn with an orthographic camera at (0,5,10), fovy 10, so 10 scene units span the texture height.
3. The texture is drawn to the 1280x720 window with `DrawTextureEx(..., scale = WindowScale)`, which is the ×8 upscale.
4. The logo (209x28 px) is drawn straight to the window at `scale = WindowScale`.
5. raygui widgets are drawn straight to the window in window pixels: a 300x44 row, a 10 px gap, and the first row at y = 360 + 60 = 420, horizontally centred.

Both 3D projections are resolution-independent. Rendering them into a smaller texture changes only pixel density, not composition. Multiplying scene sizes by `WindowScale` in step 2 therefore enlarges the scene 8×, on top of the ×8 upscale in step 3.

Measured numbers:
- Model bounds: the head is about 3.1 units tall; the jab, kick and martelo models are about 5.2 to 5.7 units tall, with a horizontal (XZ) reach of up to about 3.7 units.
- Current sizes: the head scale is 0.75×8 = 6, so it is 18.4 units tall, almost twice the 10-unit view.
- Before the texture change, the menu rendered into a 1280x720 texture with the unscaled constants (head 0.75, models 0.25). That is 72 screen px per unit, the same as 9 texture px per unit × 8 now.

How the layout was settled during implementation (user review, with screenshots):
1. Putting the original-size models back at their original positions hid them behind the logo and the buttons. The buttons were added after that composition was designed.
2. Moving the models into the free space around a centred logo kept them visible but too small, even at the largest sizes that space allows. The band above the logo is only 248 px tall.
3. The user chose to move the logo and buttons to the left and give the models the right side of the window.

## Goals / Non-Goals

**Goals:**
- The scene, logo and widgets each scaled exactly once.
- A menu panel on the left and a model stage on the right, with everything fully visible (see specs/main-menu/spec.md).
- The layout math is pure and unit-tested, so future rows or assets that no longer fit are caught.

**Non-Goals:**
- Resizable or fullscreen windows, and dynamic `WindowScale`. The window stays fixed at 1280x720, and the panel is laid out in window pixels.
- High-DPI handling. On the 125% Windows display the window is 1280x720 physical pixels; that affects the whole game, not just the menu.
- The render-texture Y-flip. The menu draws the texture unflipped and compensates with negative model Y scale; the look is unchanged and is left as-is.
- Moving the background fan's vanishing point to the stage. The fan stays centred in the window.
- Restyling the menu: colours, fonts, animation timing, raygui theme.
- Replacing the logo asset or putting it through the palette shader.

## Decisions

### 1. Remove `WindowScale` from all 3D scene constants
- Background fan: endpoints `X: ±15`, `Y: i*3`. These are the original constants.
- Model scales and positions: no `WindowScale` factor. The values are set in Decision 4.

The upscale belongs only to the texture blit. The scene is then defined in scene units, as the spec requires.

*Alternative:* keep the scaled constants and render into a full-size `SW*WindowScale` texture. That fixes the size but loses the pixel-art look the texture change was meant to add, so it is rejected.

*Alternative:* keep the background fan's current ×8 endpoints. Rejected for consistency: the fan would still change whenever `WindowScale` changes. If the user prefers the denser fan, it is a two-constant change (see Risks).

### 2. Split the window: menu panel left, model stage right
The menu panel is x 0–560 and the stage is x 560–1280, a 720x720 square.

The panel width is the minimum that holds two things:
- the Host rows: a 160 px label, a 10 px gap and a 300 px widget, plus margins;
- the ×2 logo (418 px), centred over the widget column.

Any wider only shrinks the stage.

*Alternative:* keep the logo centred at ×3 and use full-height side columns for the models. The user was offered this and turned it down: it gains only 30–60% in model size.

*Alternative:* put the menu on the right. That gives the same space; the left matches the chosen mock-up.

### 3. Menu panel layout
Constants sit next to the pure helpers `menuRow(i)` and `logoScale(texW)`:
- `menuPanelW = 560`.
- Rows: `menuRowX = 190`, `menuRowY = 360`, 300x40, 8 px gap. The six Host rows end at y 640, 80 px above the window bottom.
- Labels: `menuLabelW = 170` px left of the row, so x 20–180.
- Logo: `logoScale(texW)` is the largest integer scale at which the logo, centred over the widget column (x 340), stays in the panel. That is `floor(2·(menuPanelW − 340) / texW) = floor(440 / texW)`, clamped to `[1, WindowScale]`. For the 209 px asset it is ×2: 418x56 at x 131–549, with its bottom at `menuLogoBottom = menuRowY − 64` (y 296).
- Error line: at y = first row − 34 (326–346), starting at the label column (x 20). Long messages then have about 540 px before they reach the stage.

*Trade-off:* the logo is smaller than in the centred layout (×2 instead of ×4), and its pixels are a quarter of the scene's 8 px blocks. A ×3 logo (627 px) cannot fit a panel that still leaves the models a useful stage.

### 4. Model placement on the stage
How a scene position maps to the screen:
- The orthographic camera is at (0,5,10) looking at the origin, at 72 px per unit.
- The texture is drawn unflipped, so scene −Y is up on screen.
- Screen x = 640 + 72·X, and screen y ≈ 360 + 64.4·Y − 32.2·Z.
- A model with negative Y scale extends upward on screen from its position, by its scaled height.

The top half of the stage holds jab1, the head and jab2. The bottom half holds the martelo and the kick. Each model is as large as its cell allows while its rotating silhouette stays inside the cell.

| Model | Scale | Position (X, Y) | Approx. screen box, including rotation |
|---|---|---|---|
| jab1 | 0.55 | (0.7, −1.25) | x 584–796, y 44–326 |
| head | 1.3 | (3.9, −2.7) | x 811–1029, y 28–342 |
| jab2 | 0.55 | (7.1, −1.25) | x 1044–1256, y 44–326 |
| martelo | 0.7 | (1.4, 4.7) | x 583–899, y 374–704 |
| kick | 0.7 | (6.4, 4.7) | x 973–1229, y 375–705 |

The boxes are estimates from the OBJ bounds, reduced by about 20%. Earlier screenshots showed that the raw worst-case bounds overestimate the silhouettes. They must be confirmed with screenshots over several rotations (task 4.1).

### 5. Scene pixel size is a separate setting from the window scale
`menuPixel` is the on-screen size of one scene pixel. The scene renders into a `1280/menuPixel` x `720/menuPixel` texture and is drawn back up by `menuPixel`. It replaces the old arrangement, where the texture was `SW`x`SH` and the blit scale was `WindowScale`, which tied the menu's pixel density to the level's 160x90 grid for no reason: the menu has no tiles or sprites on that grid.

`menuPixel = 4` (a 320x180 scene). At the old value of 8 the user found the wireframes too coarse to read — at 8 px per pixel the head's edges merge into blocks. Screenshots at 4 and 2 were compared: 4 keeps hard pixel edges while separating the edges of the head; 2 is nearly smooth and no longer reads as pixel art.

Because both 3D projections are resolution-independent, this changes nothing about composition: no model scale or position needs adjusting. The one visible side effect is that `DrawLine3D` edges, always one texture pixel wide, are thinner on screen (4 px instead of 8). That is the point of the setting.

*Alternative:* keep 8 and enlarge the models further. Rejected: the stage has no room left, and coarse pixels are what hides the detail.

*Alternative:* drop the render texture and draw the scene straight to the window. Rejected: that removes the pixel-art look entirely.

### 6. Unit test for layout
`main_menu_test.go` needs only pure functions, no window. It checks that:
- `menuPixel` divides the window evenly, so the scene texture has whole dimensions;
- all six rows end at least 16 px above the window bottom;
- each row, with its label, is inside x 16–`menuPanelW`;
- the error line is at or below `menuLogoBottom`;
- `logoScale(209)` is 2;
- for a range of widths, the scale is an integer in `[1, WindowScale]`, and when it is above 1 the logo fits the 440 px available.

The visual requirements (model boxes, pixel blocks) are checked manually by running the game.

## Risks / Trade-offs

- [The user may be used to the denser background fan] → The fan endpoints are two constants. Mention the change when handing it over, and restore the ×8 endpoints if the user asks.
- [The background fan converges at the window centre (x 640), not the stage centre] → Left as a non-goal. Moving the vanishing point is a small follow-up if it looks off.
- [Mixed pixel density: the logo's pixels versus the scene's 4 px blocks] → Accepted as the cost of the side-by-side layout.
- [A smaller `menuPixel` means a bigger render texture and more fragments] → At 320x180 the cost is negligible, and the menu is not a hot path. Anything below 2 would be worth measuring.
- [raygui text in 40 px rows] → The text size is 20, so a 40 px row still leaves 10 px of padding.
- [Model positions are tuned to the panel width and the stage, so changing the panel, logo or rows could overlap them again] → The Decision 4 table records each model's box. Re-check with a screenshot whenever the panel changes.
- [Long error messages (over about 50 characters at size 20) could run into the stage, because DrawText does not wrap] → Current messages are shorter. Wrap or truncate if a longer one is added.
- [Clamping the logo scale to 1 when an asset is too wide] → A logo wider than 440 px at ×1 would still draw and could clip. That can't happen with the current asset; the unit test documents the limit.
