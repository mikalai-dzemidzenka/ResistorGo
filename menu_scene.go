package resistor

import (
	"math"
	"math/rand"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// The menu scene is drawn at this size in scene pixels; menuPixel screen pixels
// show one scene pixel.
const (
	menuSceneW = SW * WindowScale / menuPixel
	menuSceneH = SH * WindowScale / menuPixel

	// The scene is composed for a menuDesignH-tall scene and every length in it
	// is scaled from there (see menuScaled), so menuPixel decides how finely the
	// scene is drawn, not where anything in it sits.
	menuDesignH = 180

	menuHorizonY = 104 * menuSceneH / menuDesignH // the sky stands above this line; the grid runs below it
)

// menuScaled is a length of the menuDesignH-tall composition in scene pixels.
// Line widths are not scaled: a neon line is one scene pixel whatever the scene
// is drawn at, so a finer scene draws a finer line.
func menuScaled(v float32) float32 { return v * menuSceneH / menuDesignH }

// menuRedIndexes are the Endesga 64 entries the menu may be painted from: the
// palette's reds, every one of them within 20 degrees of pure red. Nothing in
// the scene or in the terminal chrome around it takes a colour from anywhere
// else, so the menu cannot drift off-red.
var menuRedIndexes = []int{17, 63, 62, 27, 61, 26, 60, 59, 58, 0}

// The colours the scene is painted with, each an entry of menuRedIndexes.
var (
	menuSky       = EndesgaPalette[17]    // the flat sky, and the dark the CRT gaps are cut in
	menuSkyTop    = EndesgaPalette[63]    // the sky lifts to this at the top of the window
	menuSunColors = []int{58, 59, 60, 61} // the top of the sun down
	menuGround    = EndesgaPalette[17]
	menuRain      = EndesgaPalette[61] // the streaks falling over the scene
	menuModelHead = EndesgaPalette[60]
	menuModelJab  = EndesgaPalette[59]
	menuModelKick = EndesgaPalette[58]
)

// The colours of the terminal the menu is framed as: its rules, its panel and
// the rows drawn on it. They come from the same list as the scene, so the
// chrome reads as part of the picture rather than as something laid over it.
var (
	menuNeon      = EndesgaPalette[60] // the rules, and the border of the highlighted row
	menuAccent    = EndesgaPalette[0]  // the horizon, the row marker, the footer
	menuBright    = EndesgaPalette[58] // row labels, headings, widget text
	menuMuted     = EndesgaPalette[26] // the small readouts
	menuDim       = EndesgaPalette[62] // the border of a row that is not highlighted
	menuPanelCol  = EndesgaPalette[17] // the panel the menu column sits on, darkening the scene behind it
	menuRowFill   = EndesgaPalette[63] // the inside of a row, a step up from the panel
	menuRowFillOn = EndesgaPalette[62] // the inside of the highlighted row
)

// neonRole is what a cycling colour is used for.
type neonRole uint8

const (
	neonGrid neonRole = iota
	neonRoleCount
)

// neonRamps are the palette entries each role steps through. Every entry is an
// index into menuRedIndexes' colours, so a cycling colour is always a palette
// red: it never blends through anything in between.
var neonRamps = [neonRoleCount][]int{
	neonGrid: {62, 61, 60, 59, 60, 61},
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

// menuDrop is one streak of the rain falling over the scene. Its position is
// normalized to the scene, so the rain is laid out the same whatever the scene
// is drawn at.
type menuDrop struct{ x, y, speed, length float32 }

// menuDropCount is how many streaks fall at once.
const menuDropCount = 90

var (
	// menuDrops is the rain, laid out once from a fixed seed so every run gets
	// the same shower rather than a different one each time.
	menuDrops = makeMenuDrops(7)
	// the column a streak comes back in is drawn from this rather than the
	// global source, so nothing else the game does shifts the rain
	menuRainRand = rand.New(rand.NewSource(7))
)

// makeMenuDrops lays out the rain: a normalized position, a fall speed in scene
// heights per second, and a streak length in scene pixels.
func makeMenuDrops(seed int64) []menuDrop {
	r := rand.New(rand.NewSource(seed))
	drops := make([]menuDrop, menuDropCount)
	for i := range drops {
		drops[i] = menuDrop{
			x:      r.Float32(),
			y:      r.Float32(),
			speed:  0.15 + r.Float32()*0.35,
			length: menuScaled(5 + r.Float32()*18),
		}
	}
	return drops
}

// drawMenuBackground paints the scene's layers: the sky and its sun, the neon
// grid running to the horizon, the rain, and the CRT over all of it. Only the
// grid's scroll, the rain, the sweeping band and the cycling colours change from
// frame to frame; nothing here moves as a whole.
func drawMenuBackground(t float64, dt float32, grid rl.Color) {
	rl.BeginTextureMode(MenuBackground)
	rl.ClearBackground(menuSky)
	rl.DrawRectangleGradientV(0, 0, menuSceneW, menuHorizonY, menuSkyTop, menuSky)

	// the sun, banded like a synthwave sun and cut off by the horizon
	cx := float32(menuSceneW) / 2
	sunY, sunR := menuHorizonY-menuScaled(16), menuScaled(21)
	for i, c := range menuSunColors {
		h := 2 * sunR / float32(len(menuSunColors))
		top := sunY - sunR + float32(i)*h
		rl.BeginScissorMode(0, int32(top), menuSceneW, int32(h))
		rl.DrawCircle(int32(cx), int32(sunY), sunR, EndesgaPalette[c])
		rl.EndScissorMode()
	}
	for y := sunY; y < sunY+sunR; y += menuScaled(4) { // the sun's scanline gaps
		rl.DrawRectangleRec(rl.Rectangle{X: cx - sunR, Y: y, Width: 2 * sunR, Height: menuScaled(2)}, menuSky)
	}

	drawNeonGrid(t, grid)
	// the horizon itself, the one line the grid converges on
	rl.DrawLineEx(rl.Vector2{Y: menuHorizonY}, rl.Vector2{X: menuSceneW, Y: menuHorizonY}, 1, rl.Fade(menuAccent, 0.75))

	drawMenuRain(dt)
	drawMenuScan(t)

	rl.EndTextureMode()
}

// gridSpeed is how fast the grid scrolls toward the viewer, in rows per second.
const gridSpeed = 1.25

// drawNeonGrid draws the synthwave floor below the horizon: lines converging on
// the vanishing point, crossed by lines that scroll toward the viewer. The cross
// lines are placed by 1/z, so they spread apart as they come nearer, and the
// scroll repeats seamlessly because row k at the end of a step is exactly where
// row k-1 started.
func drawNeonGrid(t float64, col rl.Color) {
	rl.DrawRectangleRec(rl.Rectangle{Y: menuHorizonY, Width: menuSceneW, Height: menuSceneH - menuHorizonY}, menuGround)
	cx := float32(menuSceneW) / 2
	for i := -14; i <= 14; i++ {
		end := rl.Vector2{X: cx + float32(i)*menuScaled(26), Y: menuSceneH}
		rl.DrawLineEx(rl.Vector2{X: cx, Y: menuHorizonY}, end, 1, col)
	}
	depth := float64(menuSceneH - menuHorizonY)
	frac := t*gridSpeed - math.Floor(t*gridSpeed)
	for k := 1; k <= 18; k++ {
		z := float64(k) - frac
		if z < 0.35 {
			continue
		}
		y := float32(menuHorizonY + depth/z)
		if y > menuSceneH {
			continue
		}
		rl.DrawLineEx(rl.Vector2{Y: y}, rl.Vector2{X: menuSceneW, Y: y}, 1, col)
	}
}

// drawMenuRain advances the rain by dt and draws it: a dim streak with a bright
// tip, the way rain catches the neon. A streak that has fallen past the bottom
// comes back above the top in a new column.
func drawMenuRain(dt float32) {
	tip := menuScaled(2)
	for i := range menuDrops {
		d := &menuDrops[i]
		d.y += d.speed * dt
		if d.y > 1.15 {
			d.y = -0.15
			d.x = menuRainRand.Float32()
		}
		x, y := d.x*menuSceneW, d.y*menuSceneH
		menuVLine(x, y, d.length, rl.Fade(menuRain, 0.45))
		menuVLine(x, y+d.length-tip, tip, menuRain)
	}
}

// menuVLine draws a one pixel wide vertical line h pixels long.
func menuVLine(x, y, h float32, col rl.Color) {
	rl.DrawLineEx(rl.Vector2{X: x, Y: y}, rl.Vector2{X: x, Y: y + h}, 1, col)
}

// menuScanSpeed is how fast the scanning band crosses the scene, in scene
// pixels per second.
const menuScanSpeed = 32

// drawMenuScan lays the CRT over the finished scene: a band sweeping down it,
// then the scanline gaps. The gaps are painted in the scene's darkest red
// rather than black, so the scene keeps to its reds; they cut the grid, the sun
// and the rain, and leave the flat sky as it is.
func drawMenuScan(t float64) {
	band := menuScaled(40)
	top := float32(math.Mod(t*menuScanSpeed, menuSceneH+float64(band))) - band
	rl.DrawRectangleGradientV(0, int32(top), menuSceneW, int32(band), rl.Fade(menuNeon, 0), rl.Fade(menuNeon, 0.12))
	for y := int32(0); y < menuSceneH; y += 2 {
		rl.DrawLine(0, y, menuSceneW, y, rl.Fade(menuSky, 0.5))
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
