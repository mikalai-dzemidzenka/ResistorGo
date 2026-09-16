package resistor

import (
	"image/color"
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	CursorLen   = 10
	CursorThick = 3
)

// The attack cross follows the mouse at a rate set by the fighter's energy:
// CursorFollowFull closes half the gap every 1/60 s, as the cross always did at
// 60 fps, and CursorFollowTired is how sluggish an empty pool makes it.
const (
	CursorFollowFull  = 60 * math.Ln2
	CursorFollowTired = 8
)

// crosshairTired is the attack cross's colour at an empty pool.
var crosshairTired = rl.Color{R: 170, G: 200, B: 175, A: 255}

type Cursor struct {
	Pos    rl.Vector2
	AtkPos rl.Vector2
	energy float32 // the local fighter's energy, 0..1, as of the last Update
	GUIZ   int8
	once   bool
}

func (c *Cursor) GetGUIZ() int8 {
	return c.GUIZ
}

// Update moves the attack cross towards the mouse. energy is the local
// fighter's pool as a fraction of full: the emptier it is, the slower the cross
// follows, and so the slower the fighter changes height.
func (c *Cursor) Update(dt, energy float32) {
	c.energy = energy
	c.Pos = rl.GetMousePosition()
	c.AtkPos = rl.Vector2Lerp(c.AtkPos, c.Pos, attackFollow(energy, dt))
}

func (c *Cursor) DrawGUI() {
	DrawCross(c.Pos, rl.Red)
	DrawCross(c.AtkPos, crosshairColor(c.energy))
}

// attackFollow is the share of the gap to the mouse the attack cross closes in
// dt seconds. It is exponential in time, so the cross settles as fast at any
// frame rate.
func attackFollow(energy, dt float32) float32 {
	k := lerp(CursorFollowTired, CursorFollowFull, clamp(energy, 0, 1))
	return 1 - float32(math.Exp(-float64(k*dt)))
}

// crosshairColor is the attack cross's colour: green at a full pool, paling
// towards crosshairTired as it empties, and always opaque.
func crosshairColor(energy float32) rl.Color {
	e := clamp(energy, 0, 1)
	ch := func(tired, full uint8) uint8 {
		return uint8(roundf(lerp(float32(tired), float32(full), e)))
	}
	return rl.Color{
		R: ch(crosshairTired.R, rl.Green.R),
		G: ch(crosshairTired.G, rl.Green.G),
		B: ch(crosshairTired.B, rl.Green.B),
		A: 255,
	}
}

func DrawCross(pos rl.Vector2, col color.RGBA) {
	rl.DrawLineEx(rl.Vector2{X: pos.X, Y: pos.Y + 2},
		rl.Vector2{X: pos.X, Y: pos.Y + CursorLen},
		CursorThick, col)
	rl.DrawLineEx(rl.Vector2{X: pos.X, Y: pos.Y - 2},
		rl.Vector2{X: pos.X, Y: pos.Y - CursorLen},
		CursorThick, col)
	rl.DrawLineEx(rl.Vector2{X: pos.X + 2, Y: pos.Y},
		rl.Vector2{X: pos.X + CursorLen, Y: pos.Y},
		CursorThick, col)
	rl.DrawLineEx(rl.Vector2{X: pos.X, Y: pos.Y + 2},
		rl.Vector2{X: pos.X - CursorLen, Y: pos.Y},
		CursorThick, col)
}
