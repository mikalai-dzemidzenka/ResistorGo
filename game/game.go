package game

import (
	"cmp"
	"flag"
	rl "github.com/gen2brain/raylib-go/raylib"
	"slices"
	"strconv"
)

const (
	SW = 1024
	SH = 768
)

var l int

const (
	LEVEL_1 = iota
	LEVEL_EDITOR
	LEVEL_EDITOR_PLAY
	LEVELS
)

var Entities []any

type Drawable interface {
	Z() int8
	Draw()
}

type Updatable interface {
	Update(float32)
}

type GUIDrawable interface {
	Z() int8
	DrawGui()
}

func Run() {
	rl.InitWindow(SW, SH, "Resistor")
	defer rl.CloseWindow()

	flag.IntVar(&l, "l", LEVEL_EDITOR, "level to load")

	rl.DisableCursor()

	LoadTextures()

	rl.SetTargetFPS(60)

	var drawables []func()
	var updaters []func(float32)
	var drawGuis []func()

	Entities = []any{cur, p, cam, le}
	type idx struct {
		z    int8
		draw func()
	}
	var drawablesIndexed []idx
	var drawablesIndexedGUI []idx
	for i := range Entities {
		{
			u, ok := Entities[i].(Updatable)
			if ok {
				updaters = append(updaters, u.Update)
			}
		}
		{
			d, ok := Entities[i].(Drawable)
			if ok {
				drawablesIndexed = append(drawablesIndexed, idx{d.Z(), d.Draw})
			}
		}
		{
			dg, ok := Entities[i].(GUIDrawable)
			if ok {
				drawablesIndexedGUI = append(drawablesIndexedGUI, idx{dg.Z(), dg.DrawGui})
			}
		}
	}
	{
		slices.SortFunc(drawablesIndexed, func(a, b idx) int {
			return cmp.Compare(a.z, b.z)
		})
		for i := range drawablesIndexed {
			drawables = append(drawables, drawablesIndexed[i].draw)
		}
	}
	{
		slices.SortFunc(drawablesIndexedGUI, func(a, b idx) int {
			return cmp.Compare(a.z, b.z)
		})
		for i := range drawablesIndexedGUI {
			drawGuis = append(drawGuis, drawablesIndexedGUI[i].draw)
		}
	}

	for !rl.WindowShouldClose() {
		if rl.IsKeyPressed(rl.KeyGrave) {
			Debug = !Debug
		}

		dt := rl.GetFrameTime()

		Debug = true

		for i := range updaters {
			updaters[i](dt)
		}

		rl.BeginDrawing()
		{
			rl.ClearBackground(rl.Black)
			rl.BeginMode2D(cam.cam)
			{
				for i := range drawables {
					drawables[i]()
				}
			}
			rl.EndMode2D()

			for i := range drawGuis {
				drawGuis[i]()
			}
			if Debug {
				fps := strconv.Itoa(int(rl.GetFPS()))
				rl.DrawText("fps:"+fps, 20, 20, 2, rl.White)
				rl.DrawText(p.Debug(), 40, 40, 1, rl.Red)
				rl.DrawText(cam.Debug(), 40, 78, 1, rl.Green)
				rl.DrawText(cur.Debug(), 40, 116, 1, rl.Blue)
			}
		}
		rl.EndDrawing()
	}
}
