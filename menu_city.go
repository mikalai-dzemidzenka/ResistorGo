package resistor

import rl "github.com/gen2brain/raylib-go/raylib"

// The menu scene is drawn at this size in scene pixels; menuPixel screen pixels
// show one scene pixel.
const (
	menuSceneW = SW * WindowScale / menuPixel
	menuSceneH = SH * WindowScale / menuPixel

	cityHorizonY  = 104 // the skyline stands on this line; the grid runs below it
	citySkyMargin = 22  // nothing reaches above this, so the dusk sky still shows
	cityWindow    = 2   // a lit window is this many scene pixels square
	citySeed      = 0x4e656f4d696e736b
)

// cityShape is how a building's silhouette is drawn.
type cityShape uint8

const (
	shapeBlock   cityShape = iota // a plain slab: the filler city and the avenue
	shapeStepped                  // a stepped crown: the Gates of Minsk towers
	shapeDiamond                  // a cut gem on a plinth: the National Library
	shapeObelisk                  // a needle on a stepped base: the Hero City obelisk
	shapeMast                     // a shaft with an observation ring: the TV tower
)

// cityBuilding is one building of the skyline, in scene pixels. Rect.Y is its
// top, and its base always sits on the horizon.
type cityBuilding struct {
	Name    string // the landmark it is, or "" for a filler building
	Rect    rl.Rectangle
	Shape   cityShape
	Windows []rl.Vector2 // top-left of each lit window
}

// cityLandmarks are the five silhouettes that make the skyline Minsk. Their
// positions are fixed, so they never move and never cover one another.
var cityLandmarks = []cityBuilding{
	{Name: "avenue", Shape: shapeBlock, Rect: rl.Rectangle{X: 6, Y: 62, Width: 58, Height: cityHorizonY - 62}},
	{Name: "gates", Shape: shapeStepped, Rect: rl.Rectangle{X: 74, Y: 40, Width: 46, Height: cityHorizonY - 40}},
	{Name: "library", Shape: shapeDiamond, Rect: rl.Rectangle{X: 130, Y: 44, Width: 46, Height: cityHorizonY - 44}},
	{Name: "obelisk", Shape: shapeObelisk, Rect: rl.Rectangle{X: 192, Y: 28, Width: 26, Height: cityHorizonY - 28}},
	{Name: "tvtower", Shape: shapeMast, Rect: rl.Rectangle{X: 234, Y: 24, Width: 24, Height: cityHorizonY - 24}},
}

// cityRand is a small deterministic generator. The city must look the same on
// every run, so it does not use math/rand, whose stream is free to change.
type cityRand uint64

func (r *cityRand) next() uint64 {
	*r = cityRand(uint64(*r)*6364136223846793005 + 1442695040888963407)
	return uint64(*r) >> 33
}

func (r *cityRand) intn(n int) int { return int(r.next() % uint64(n)) }

// cityLayout builds the skyline: the five landmarks at their fixed positions,
// with filler buildings in the gaps between them. The same seed always gives
// the same city.
func cityLayout(seed int64) []cityBuilding {
	r := cityRand(seed)
	out := make([]cityBuilding, 0, 24)
	x := float32(0)
	for _, lm := range cityLandmarks {
		out = append(out, cityFill(&r, x, lm.Rect.X)...)
		b := lm
		b.Windows = cityWindows(&r, b)
		out = append(out, b)
		x = lm.Rect.X + lm.Rect.Width
	}
	return append(out, cityFill(&r, x, menuSceneW)...)
}

// cityFill puts plain blocks in the gap between two landmarks. They are always
// shorter than the landmarks around them, so nothing important is hidden.
func cityFill(r *cityRand, x0, x1 float32) []cityBuilding {
	var out []cityBuilding
	for x := x0 + 1; x+6 <= x1-1; {
		w := float32(6 + r.intn(10))
		if x+w > x1-1 {
			w = x1 - 1 - x
		}
		if w < 6 {
			break
		}
		top := float32(68 + r.intn(24))
		b := cityBuilding{Shape: shapeBlock, Rect: rl.Rectangle{X: x, Y: top, Width: w, Height: cityHorizonY - top}}
		b.Windows = cityWindows(r, b)
		out = append(out, b)
		x += w + float32(2+r.intn(3))
	}
	return out
}

// cityBody is the solid part of a building, the part that can carry windows. A
// diamond, a needle or a mast is only solid near its base.
func cityBody(b cityBuilding) rl.Rectangle {
	r := b.Rect
	switch b.Shape {
	case shapeDiamond:
		h := r.Height * 0.45
		return rl.Rectangle{X: r.X, Y: cityHorizonY - h, Width: r.Width, Height: h}
	case shapeObelisk:
		h := r.Height * 0.22
		return rl.Rectangle{X: r.X + r.Width*0.2, Y: cityHorizonY - h, Width: r.Width * 0.6, Height: h}
	case shapeMast:
		h := r.Height * 0.18
		return rl.Rectangle{X: r.X + r.Width*0.25, Y: cityHorizonY - h, Width: r.Width * 0.5, Height: h}
	}
	return r
}

// cityWindows lights a fixed scatter of windows over a building's solid body.
func cityWindows(r *cityRand, b cityBuilding) []rl.Vector2 {
	body := cityBody(b)
	var out []rl.Vector2
	for y := body.Y + 3; y <= cityHorizonY-1-cityWindow; y += 2 * cityWindow {
		for x := body.X + 2; x <= body.X+body.Width-1-cityWindow; x += 2 * cityWindow {
			if r.intn(10) < 4 {
				out = append(out, rl.Vector2{X: x, Y: y})
			}
		}
	}
	return out
}

// drawCity draws the skyline into the background texture: flat silhouettes,
// lit only by their windows. The buildings carry no outline.
func drawCity(bs []cityBuilding, silhouette, window rl.Color) {
	for _, b := range bs {
		switch b.Shape {
		case shapeStepped:
			drawGates(b, silhouette)
		case shapeDiamond:
			drawLibrary(b, silhouette)
		case shapeObelisk:
			drawObelisk(b, silhouette)
		case shapeMast:
			drawTVTower(b, silhouette)
		default:
			rl.DrawRectangleRec(b.Rect, silhouette)
		}
		for _, w := range b.Windows {
			rl.DrawRectangle(int32(w.X), int32(w.Y), cityWindow, cityWindow, window)
		}
	}
}

// drawGates is the twin towers of the Gates of Minsk, each with a stepped crown.
func drawGates(b cityBuilding, fill rl.Color) {
	r := b.Rect
	tw := (r.Width - 6) / 2
	for _, x := range []float32{r.X, r.X + tw + 6} {
		rl.DrawRectangleRec(rl.Rectangle{X: x, Y: r.Y + 6, Width: tw, Height: cityHorizonY - r.Y - 6}, fill)
		rl.DrawRectangleRec(rl.Rectangle{X: x + 3, Y: r.Y, Width: tw - 6, Height: 6}, fill)
	}
}

// fillDiamond draws a rhombus as one run of pixels per row. Drawn as a triangle
// fan it would be at the mercy of raylib's winding rule; rows always draw, and
// they land on the scene's pixel grid.
func fillDiamond(cx, top, bot, hw float32, c rl.Color) {
	mid := (top + bot) / 2
	for y := top; y < bot; y++ {
		f := (y - top) / (mid - top)
		if y >= mid {
			f = (bot - y) / (bot - mid)
		}
		w := hw * f
		rl.DrawRectangleRec(rl.Rectangle{X: cx - w, Y: y, Width: 2*w + 1, Height: 1}, c)
	}
}

// fillTaper draws a needle: a point at the top widening to a base.
func fillTaper(cx, top, bot, halfBase float32, c rl.Color) {
	for y := top; y < bot; y++ {
		w := halfBase * (y - top) / (bot - top)
		rl.DrawRectangleRec(rl.Rectangle{X: cx - w, Y: y, Width: 2*w + 1, Height: 1}, c)
	}
}

// drawLibrary is the National Library: a cut gem standing on a plinth.
func drawLibrary(b cityBuilding, fill rl.Color) {
	body := cityBody(b)
	rl.DrawRectangleRec(rl.Rectangle{X: body.X + body.Width/3, Y: body.Y, Width: body.Width / 3, Height: body.Height}, fill)
	cx := b.Rect.X + b.Rect.Width/2
	fillDiamond(cx, b.Rect.Y, body.Y, b.Rect.Width/2, fill)
}

// drawObelisk is the Hero City obelisk: a needle on a stepped base.
func drawObelisk(b cityBuilding, fill rl.Color) {
	body := cityBody(b)
	rl.DrawRectangleRec(body, fill)
	cx := b.Rect.X + b.Rect.Width/2
	fillTaper(cx, b.Rect.Y, body.Y, 3, fill)
}

// drawTVTower is the TV tower: a thin shaft with an observation ring and a mast.
func drawTVTower(b cityBuilding, fill rl.Color) {
	body := cityBody(b)
	rl.DrawRectangleRec(body, fill)
	cx := b.Rect.X + b.Rect.Width/2
	ringY := b.Rect.Y + b.Rect.Height*0.35
	rl.DrawRectangleRec(rl.Rectangle{X: cx - 2, Y: ringY, Width: 4, Height: body.Y - ringY}, fill)
	rl.DrawRectangleRec(rl.Rectangle{X: cx - 6, Y: ringY - 3, Width: 12, Height: 6}, fill)
	rl.DrawRectangleRec(rl.Rectangle{X: cx - 1, Y: b.Rect.Y, Width: 2, Height: ringY - 3 - b.Rect.Y}, fill)
}
