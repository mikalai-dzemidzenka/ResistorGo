package resistor

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// menuRedIndexes are the Endesga 64 entries the menu scene may be painted from:
// the palette's reds, every one of them within 20 degrees of pure red. Nothing
// in the scene takes a colour from anywhere else, so it cannot drift off-red.
var menuRedIndexes = []int{17, 63, 62, 27, 61, 26, 60, 59, 58, 0}

// The colours the scene is painted with, each an entry of menuRedIndexes.
var (
	menuSkyBands   = []int{17, 63, 62, 27, 61, 60} // the top of the scene down to the horizon
	menuSunColors  = []int{58, 59, 60, 61}         // the top of the sun down
	menuSilhouette = EndesgaPalette[17]            // the city, and the ground under it
	menuGround     = EndesgaPalette[17]
	menuModelHead  = EndesgaPalette[60]
	menuModelJab   = EndesgaPalette[59]
	menuModelKick  = EndesgaPalette[58]
)

// neonRole is what a cycling colour is used for.
type neonRole uint8

const (
	neonGrid neonRole = iota
	neonWindow
	neonRoleCount
)

// neonRamps are the palette entries each role steps through. Every entry is an
// index into menuRedIndexes' colours, so a cycling colour is always a palette
// red: it never blends through anything in between.
var neonRamps = [neonRoleCount][]int{
	neonGrid:   {62, 61, 60, 59, 60, 61},
	neonWindow: {59, 58, 60, 61},
}

// neonStepSec is how long one colour of a ramp is held.
const neonStepSec = 0.9

// neonColor is the colour of a role at time t: one step of its ramp.
func neonColor(role neonRole, t float64) rl.Color {
	ramp := neonRamps[role]
	i := int(math.Floor(t/neonStepSec)) % len(ramp)
	if i < 0 {
		i += len(ramp)
	}
	return EndesgaPalette[ramp[i]]
}

// drawMenuBackground paints the scene's fixed layers: the dusk sky and its sun,
// the Neo-Minsk skyline, and the neon grid running to the horizon. Only the
// grid's scroll and the cycling colours change from frame to frame; nothing
// here moves as a whole.
func drawMenuBackground(t float64, grid, window rl.Color) {
	rl.BeginTextureMode(MenuBackground)
	rl.ClearBackground(EndesgaPalette[menuSkyBands[0]])

	// the sky, in bands from the top of the scene down to the horizon
	band := float32(cityHorizonY) / float32(len(menuSkyBands))
	for i, c := range menuSkyBands {
		y := float32(i) * band
		rl.DrawRectangleRec(rl.Rectangle{Y: y, Width: menuSceneW, Height: band + 1}, EndesgaPalette[c])
	}

	// the sun, banded like a synthwave sun and cut off by the skyline
	cx := float32(menuSceneW) / 2
	sunY, sunR := float32(cityHorizonY)-16, float32(21)
	for i, c := range menuSunColors {
		h := 2 * sunR / float32(len(menuSunColors))
		top := sunY - sunR + float32(i)*h
		rl.BeginScissorMode(0, int32(top), menuSceneW, int32(h))
		rl.DrawCircle(int32(cx), int32(sunY), sunR, EndesgaPalette[c])
		rl.EndScissorMode()
	}
	for y := sunY; y < sunY+sunR; y += 4 { // the sun's scanline gaps
		rl.DrawRectangleRec(rl.Rectangle{X: cx - sunR, Y: y, Width: 2 * sunR, Height: 2}, EndesgaPalette[menuSkyBands[len(menuSkyBands)-1]])
	}

	drawCity(menuCity, menuSilhouette, window)
	drawNeonGrid(t, grid)

	rl.EndTextureMode()
}

// gridSpeed is how fast the grid scrolls toward the viewer, in rows per second.
const gridSpeed = 0.45

// drawNeonGrid draws the synthwave floor below the horizon: lines converging on
// the vanishing point, crossed by lines that scroll toward the viewer. The cross
// lines are placed by 1/z, so they spread apart as they come nearer, and the
// scroll repeats seamlessly because row k at the end of a step is exactly where
// row k-1 started.
func drawNeonGrid(t float64, col rl.Color) {
	rl.DrawRectangleRec(rl.Rectangle{Y: cityHorizonY, Width: menuSceneW, Height: menuSceneH - cityHorizonY}, menuGround)
	cx := float32(menuSceneW) / 2
	for i := -14; i <= 14; i++ {
		end := rl.Vector2{X: cx + float32(i)*26, Y: menuSceneH}
		rl.DrawLineEx(rl.Vector2{X: cx, Y: cityHorizonY}, end, 1, col)
	}
	const depth = float64(menuSceneH - cityHorizonY)
	frac := t*gridSpeed - math.Floor(t*gridSpeed)
	for k := 1; k <= 18; k++ {
		z := float64(k) - frac
		if z < 0.35 {
			continue
		}
		y := float32(cityHorizonY + depth/z)
		if y > menuSceneH {
			continue
		}
		rl.DrawLineEx(rl.Vector2{Y: y}, rl.Vector2{X: menuSceneW, Y: y}, 1, col)
	}
}

// drawSceneBackground puts the finished background into the scene. The scene
// texture is shown unflipped (see runMainMenu), so the background is drawn
// upside down here and that final draw turns it back.
func drawSceneBackground() {
	rl.DrawTexturePro(MenuBackground.Texture,
		rl.Rectangle{Width: menuSceneW, Height: menuSceneH},
		rl.Rectangle{Width: menuSceneW, Height: menuSceneH}, rl.Vector2{}, 0, rl.White)
}
