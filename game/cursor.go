package game

import (
	"fmt"
	rl "github.com/gen2brain/raylib-go/raylib"
	"image/color"
)

const (
	CursorLen             = 6
	CursorThick           = 2
	CursorAttackSpeed     = 0.1
	CursorFastAttackSpeed = 0.5
)

type Cursor struct {
	Pos      rl.Vector2
	atkPos   rl.Vector2
	atkSpeed float32
	GUIZ     int8
}

var _ Updatable = (*Cursor)(nil)
var _ GUIDrawable = (*Cursor)(nil)

func (c *Cursor) GetGUIZ() int8 {
	return c.GUIZ
}

func (c *Cursor) Update(dt float32) {
	if rl.IsMouseButtonDown(rl.MouseButtonRight) {
		c.atkSpeed = CursorFastAttackSpeed
	} else {
		c.atkSpeed = CursorAttackSpeed
	}

	c.Pos = rl.GetMousePosition()
	c.atkPos = rl.Vector2Lerp(c.atkPos, c.Pos, c.atkSpeed)
}

func (c *Cursor) DrawGUI() {
	DrawCross(c.Pos, rl.Red)
	DrawCross(g.Cursor.atkPos, rl.Green)
	if Debug {
		a := rl.GetWorldToScreen2D(rl.Vector2{
			X: g.Player.Pos.X + g.Player.Bound.Width/2,
			Y: g.Player.Pos.Y + g.Player.Bound.Height*1/4,
		}, *g.Camera)
		b := g.Cursor.atkPos
		rl.DrawLineV(a, b, rl.Blue)
	}
}

func DrawCross(pos rl.Vector2, col color.RGBA) {
	rl.DrawLineEx(rl.Vector2Add(pos, rl.Vector2{0, 2}),
		rl.Vector2Add(pos, rl.Vector2{0, CursorLen}),
		CursorThick, col)
	rl.DrawLineEx(rl.Vector2Add(pos, rl.Vector2{0, -2}),
		rl.Vector2Add(pos, rl.Vector2{0, -CursorLen}),
		CursorThick, col)
	rl.DrawLineEx(rl.Vector2Add(pos, rl.Vector2{2, 0}),
		rl.Vector2Add(pos, rl.Vector2{CursorLen, 0}),
		CursorThick, col)
	rl.DrawLineEx(rl.Vector2Add(pos, rl.Vector2{-2, 0}),
		rl.Vector2Add(pos, rl.Vector2{-CursorLen, 0}),
		CursorThick, col)
}

func (c *Cursor) Debug() string {
	if c == nil {
		return ""
	}
	return fmt.Sprintf("pos: %v\natk_pos:%v\n", g.Cursor.Pos, g.Cursor.atkPos)
}
