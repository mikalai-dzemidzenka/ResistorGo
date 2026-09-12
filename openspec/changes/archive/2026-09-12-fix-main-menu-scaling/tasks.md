## 1. Layout helpers and test

- [x] 1.1 In `main_menu.go`, extract the row geometry from `drawMenuUI` into a package-level `menuRow(i int) rl.Rectangle` with `w, h, gap = 300, 40, 8` and the first row at `SH*WindowScale/2 + 60`. Make `drawMenuUI` (rows, labels, error line at `y - 34`) use it. Verify `go build ./...` succeeds.
- [x] 1.2 Add `logoScale(texW int32) float32` returning `floor(0.7*SW*WindowScale/texW)` clamped to `[1, WindowScale]`. Verify `go build ./...` succeeds.
- [x] 1.3 Add `main_menu_test.go` with `TestMenuLayout`. It checks: `menuRow(5)` ends ≤ `SH*WindowScale-16`; `menuRow(0).Y-34` is at or below the logo bottom (`SH*WindowScale/2`); `logoScale(209) == 4`; and for texture widths 1..2000 the result is an integer ≥ 1, and `logoScale(w)*w` ≤ `SW*WindowScale` whenever the scale is above 1. Verify `go test -run TestMenuLayout .` passes.

## 2. Scale fixes

- [x] 2.1 In `head`, `jab1`, `jab2`, `kick` and `martelo`, remove `WindowScale` from the model scale and position: head scale 0.75; the others 0.25 at (-1.5,1), (1.5,1), (1,1), (-1,1). Verify with `grep WindowScale main_menu.go` that none of these functions still references it.
- [x] 2.2 In `runMainMenu`, change the background fan endpoints to `X: ±15, Y: i*3`, with no `WindowScale`. Verify `go build ./...` succeeds.
- [x] 2.3 In `runMainMenu`, draw the logo at `s := logoScale(tex.Width)`, positioned at `((SW*WindowScale - tex.Width*s)/2, SH*WindowScale/2 - tex.Height*s)`, keeping the fade-in alpha. Verify `go build ./...` succeeds.
- [x] 2.4 Move the models out from behind the logo and buttons (first placement around a centred logo): head at (0,-3.6), jab1 at (-3.2,-2.6), jab2 at (3.2,-2.6), martelo at (-6.8,3.7), kick at (6.8,3.7). Verify `go build ./...` succeeds.
- [x] 2.5 Enlarge the models to fill the free areas around the centred logo: head at scale 0.85 at (0,-3.65); jabs at 0.4 at (±3.4,-2.6); martelo and kick at 0.5 at (-6.9,4.4) and (6.9,4.4). Verify `go build ./...` succeeds.

## 3. Split layout: menu left, models right

- [x] 3.1 In `main_menu.go`, add the panel constants (design Decision 3): `menuPanelW = 560`, `menuRowX = 190`, `menuRowY = 280`, `menuLabelW = 170`, `menuLogoBottom = menuRowY - 64`. Make `menuRow`, the labels and the error line (at the label column) use them. Make `logoScale` fit the panel when centred over the widget column (`floor(2*(menuPanelW-menuRowX-menuRowW/2)/texW)`), and draw the logo centred over the widget column with its bottom at `menuLogoBottom`. Verify `go build ./...` succeeds.
- [x] 3.2 Update `TestMenuLayout` to the checks in design Decision 5: six rows ending ≤ `SH*WindowScale-16`; rows with labels inside x 16..`menuPanelW`; error line at or below `menuLogoBottom`; `logoScale(209) == 2`; widths 1..2000 give an integer scale in [1, WindowScale], at most 440 px wide when above 1. Verify `go test -run TestMenuLayout .` passes.
- [x] 3.3 Move the models to the stage (design Decision 4): jab1 0.55 at (0.7,-1.25); head 1.3 at (3.9,-2.7); jab2 0.55 at (7.1,-1.25); martelo 0.7 at (1.4,4.7); kick 0.7 at (6.4,4.7). Update the comment at the orthographic camera. Verify `go build ./...` succeeds.

## 4. Scene pixel size

- [x] 4.1 Add a `menuPixel` constant (design Decision 5) set to 4. Render the scene into a `SW*WindowScale/menuPixel` x `SH*WindowScale/menuPixel` texture and draw it back up at `menuPixel` instead of `WindowScale`. Verify `go build ./...` succeeds.
- [x] 4.2 Add a check to `TestMenuLayout` that `menuPixel` divides the window evenly. Compare screenshots at pixel size 4 and 2 and keep the one that stays pixelated while reading clearly. Verify `go test -run TestMenuLayout .` passes.

## 5. Manual verification

- [x] 5.1 Run the game from the repo root and wait more than 5 s on the menu. Check: the whole logo is visible at 2x, centred over the buttons on the left; the head is at the top centre of the right side with a jab on each side, and the martelo and kick are below; across several frames (different rotations) every model is fully inside the window and right of the menu panel; the scene is blocky (8x8 pixel blocks).
- [x] 5.2 Open Host Game and check that all six rows and their labels are fully visible inside the left panel, with readable text. Enter port `0` and press Start: check that "invalid port" appears between the logo and the Port row, overlapping neither. Check that the Join Game and Connecting screens keep their widgets in the panel. On every screen, no model is covered.
