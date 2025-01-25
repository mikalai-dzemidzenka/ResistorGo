package game

import (
	"fmt"
	rl "github.com/gen2brain/raylib-go/raylib"
	"image/color"
	"math"
)

const (
	CursorLen             = 6
	CursorThick           = 2
	CursorAttackSpeed     = 0.1
	CursorFastAttackSpeed = 0.5
)

var cur = &Cursor{
	atkSpeed: CursorAttackSpeed,
	z:        0,
}

type Cursor struct {
	pos      rl.Vector2
	atkPos   rl.Vector2
	atkSpeed float32
	atkSt    int32
	z        int8
	guiZ     int8
}

func (c *Cursor) GuiZ() int8 {
	return c.guiZ
}

func (c *Cursor) Z() int8 {
	return c.z
}

func (c *Cursor) Update(dt float32) {
	if rl.IsMouseButtonDown(rl.MouseButtonRight) {
		c.atkSpeed = CursorFastAttackSpeed
	} else {
		c.atkSpeed = CursorAttackSpeed
	}

	c.pos = rl.GetMousePosition()
	c.atkPos = rl.Vector2Lerp(c.atkPos, c.pos, c.atkSpeed)

	a := rl.GetWorldToScreen2D(rl.Vector2{X: p.pos.X + p.bound.Width/2, Y: p.pos.Y + p.bound.Height*1/4}, cam.cam)
	b := cur.atkPos
	angle := rl.Vector2LineAngle(a, b) // [-pi/2;pi/2]
	angle += math.Pi / 2               // [0, pi]
	c.atkSt = int32(angle / (math.Pi / 4))
	c.atkSt = clamp(c.atkSt, 0, 3)
}

func (c *Cursor) DrawGui() {
	DrawCross(c.pos, rl.Red)
	DrawCross(cur.atkPos, rl.Green)
	if Debug {
		a := rl.GetWorldToScreen2D(rl.Vector2{X: p.pos.X + p.bound.Width/2, Y: p.pos.Y + p.bound.Height*1/4}, cam.cam)
		b := cur.atkPos
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
	return fmt.Sprintf("pos: %v\natk_pos:%v\natk_state:%d", cur.pos, cur.atkPos, cur.atkSt)
}
