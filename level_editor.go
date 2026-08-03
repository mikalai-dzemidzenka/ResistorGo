package resistor

import rl "github.com/gen2brain/raylib-go/raylib"

type LevelEditor struct {
	obj   int // current object
	gridW int32
	gridH int32
}

func NewLevelEditor(gridW, gridH int32) LevelEditor {
	return LevelEditor{
		gridW: gridW,
		gridH: gridH,
	}
}

func (le *LevelEditor) DrawCursor(cam Camera) {
	mp := v2i(rl.GetScreenToWorld2D(rl.GetMousePosition(), cam.Camera2D))
	if mp.X < 0 {
		mp.X -= le.gridW
	}
	if mp.Y < 0 {
		mp.Y -= le.gridH
	}
	if mp.X != 0 {
		mp.X /= le.gridW
		mp.X *= le.gridW
	}

	if mp.Y != 0 {
		mp.Y /= le.gridH
		mp.Y *= le.gridH
	}
	rl.DrawTexture(E1M1Wall1, mp.X, mp.Y, rl.White)
}

func (le *LevelEditor) DrawGrid() {
	rl.PushMatrix()
	{
		a := float32(le.gridH) / float32(le.gridW)
		rl.Translatef(
			float32(le.gridW),
			float32(le.gridH+(5*le.gridH)),
			0)
		rl.Rotatef(90, 1, 0, 0)
		rl.Scalef(1, 0, a)
		rl.DrawGrid(100, 32)
	}
	rl.PopMatrix()
}
