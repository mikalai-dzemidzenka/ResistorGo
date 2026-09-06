package resistor

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"image/color"
)

const (
	CursorLen         = 10
	CursorThick       = 3
	CursorAttackSpeed = 0.5
)

type Cursor struct {
	Pos      rl.Vector2
	AtkPos   rl.Vector2
	AtkSpeed float32
	GUIZ     int8
	once     bool
}

func (c *Cursor) GetGUIZ() int8 {
	return c.GUIZ
}

func (c *Cursor) Update(dt float32) {
	if rl.IsMouseButtonDown(rl.MouseButtonRight) {
		c.AtkSpeed = CursorAttackSpeed
	} else {
		c.AtkSpeed = 0
	}

	c.Pos = rl.GetMousePosition()
	c.AtkPos = rl.Vector2Lerp(c.AtkPos, c.Pos, c.AtkSpeed)
}
func (c *Cursor) DrawGUI() {
	DrawCross(c.Pos, rl.Red)
	DrawCross(Cur.AtkPos, rl.Green)
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
