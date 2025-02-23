package game

import (
	"fmt"
	rl "github.com/gen2brain/raylib-go/raylib"
	"image/color"
)

const (
	CursorLen             = 6
	CursorThick           = 2
	CursorAttackSpeed     = 0.05
	CursorFastAttackSpeed = 0.1
)

type Cursor struct {
	Pos      rl.Vector2
	AtkPos   rl.Vector2
	AtkSpeed float32
	GUIZ     int8
}

var _ Updatable = (*Cursor)(nil)
var _ GUIDrawable = (*Cursor)(nil)

func (c *Cursor) GetGUIZ() int8 {
	return c.GUIZ
}

func (c *Cursor) Update(dt float32) {
	if rl.IsMouseButtonDown(rl.MouseButtonRight) {
		c.AtkSpeed = CursorFastAttackSpeed
	} else {
		c.AtkSpeed = CursorAttackSpeed
	}

	c.Pos = rl.GetMousePosition()
	c.AtkPos = rl.Vector2Lerp(c.AtkPos, c.Pos, c.AtkSpeed)
}

func (c *Cursor) DrawGUI() {
	DrawCross(c.Pos, rl.Red)
	DrawCross(g.Cursor.AtkPos, rl.Green)
	if Debug {
		a := rl.GetWorldToScreen2D(rl.Vector2{
			X: g.Player.Pos.X + g.Player.Bound.Width/2,
			Y: g.Player.Pos.Y + g.Player.Bound.Height*1/4,
		}, *g.Camera)
		b := g.Cursor.AtkPos
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
	return fmt.Sprintf("pos: %v\natk_pos:%v\n", g.Cursor.Pos, g.Cursor.AtkPos)
}
