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

// flashKind is the ring flashing out from the attack cross when energy is won
// back: a small one for a block's payback, a big one for a full pool.
type flashKind uint8

const (
	flashNone flashKind = iota
	flashBlock
	flashFull
)

// flashStyle is how one kind of flash plays out: how long it lasts, the radius
// it grows from and to, the ring's thickness and its colour at the start.
type flashStyle struct {
	dur, from, to, thick float32
	col                  rl.Color
}

var flashStyles = [...]flashStyle{
	flashBlock: {dur: 0.25, from: 10, to: 26, thick: 2, col: rl.White},
	flashFull:  {dur: 0.5, from: 10, to: 48, thick: 3, col: rl.Color{R: 180, G: 255, B: 190, A: 255}},
}

type Cursor struct {
	Pos      rl.Vector2
	AtkPos   rl.Vector2
	energy   float32   // the local fighter's energy, 0..1, as of the last Update
	flash    flashKind // the ring showing, if any
	flashAge float32   // seconds since it started
	GUIZ     int8
	once     bool
}

func (c *Cursor) GetGUIZ() int8 {
	return c.GUIZ
}

// Update moves the attack cross towards the mouse. energy is the local
// fighter's pool as a fraction of full: the emptier it is, the slower the cross
// follows, and so the slower the fighter changes height. live says whether
// energy belongs to the same fighter as last frame and is not a rewound one,
// so that a change in it is something the player earned.
func (c *Cursor) Update(dt, energy float32, live bool) {
	c.trackEnergy(dt, energy, live)
	c.Pos = rl.GetMousePosition()
	c.AtkPos = rl.Vector2Lerp(c.AtkPos, c.Pos, attackFollow(energy, dt))
}

// trackEnergy takes this frame's energy, starting a flash when it was won back.
func (c *Cursor) trackEnergy(dt, energy float32, live bool) {
	if k := detectFlash(c.energy, energy, dt, live); k != flashNone {
		c.flash, c.flashAge = k, 0
	} else {
		c.flashAge += dt
	}
	c.energy = energy
}

func (c *Cursor) DrawGUI() {
	if inner, outer, col, ok := flashRing(c.flash, c.flashAge); ok {
		rl.DrawRing(c.AtkPos, inner, outer, 0, 360, 36, col)
	}
	DrawCross(c.Pos, rl.Red)
	DrawCross(c.AtkPos, crosshairColor(c.energy))
}

// detectFlash tells from two frames' energy, as fractions of full, whether a
// flash starts: the big one when the pool has just become full, the small one
// when it rose by more than refilling could have managed in dt - which only a
// block's payback does. Nothing flashes when the change is not live.
func detectFlash(prev, now, dt float32, live bool) flashKind {
	if !live || now <= prev {
		return flashNone
	}
	if now >= 1 {
		return flashFull
	}
	refill := float32(60/EnergyFreeRegenTicks) * dt // the most refilling adds in a frame
	if (now-prev)*MaxEnergy > refill+float32(EnergyBlockRestore)/2 {
		return flashBlock
	}
	return flashNone
}

// flashRing is the flash's ring age seconds in: its inner and outer radius and
// its colour, or ok false once it is over. The ring grows fast and settles, and
// fades out as it goes.
func flashRing(k flashKind, age float32) (inner, outer float32, col rl.Color, ok bool) {
	if k == flashNone || int(k) >= len(flashStyles) {
		return 0, 0, rl.Color{}, false
	}
	s := flashStyles[k]
	if age < 0 || age >= s.dur {
		return 0, 0, rl.Color{}, false
	}
	t := age / s.dur
	outer = lerp(s.from, s.to, 1-(1-t)*(1-t))
	col = s.col
	col.A = uint8(roundf(float32(s.col.A) * (1 - t)))
	return outer - s.thick, outer, col, true
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
