package game

import (
	"flag"
	rlg "github.com/gen2brain/raylib-go/raygui"
	rl "github.com/gen2brain/raylib-go/raylib"
	"strconv"
)

const (
	SW = 1280
	SH = 720
)

var CurrentLevel int

const (
	LEVEL_1 = iota
	LEVEL_EDITOR
	LEVEL_EDITOR_PLAY
	LEVELS
)

func Run() {
	rl.InitWindow(SW, SH, "Resistor")
	defer rl.CloseWindow()

	flag.IntVar(&CurrentLevel, "l", LEVEL_EDITOR, "level to load")

	rl.DisableCursor()

	LoadTextures()
	rlg.LoadStyle("resources/style_dark.rgs")

	rl.SetTargetFPS(60)

	Entities = []any{cur, p, cam, le}
	SortEntities()

	for !rl.WindowShouldClose() {
		if rl.IsKeyPressed(rl.KeyGrave) {
			Debug = !Debug
		}

		dt := rl.GetFrameTime()

		for i := range Updaters {
			Updaters[i](dt)
		}

		rl.BeginDrawing()
		{
			rl.ClearBackground(rl.Black)
			rl.BeginMode2D(cam.cam)
			{
				for i := range Drawables {
					Drawables[i].draw()
				}
			}
			rl.EndMode2D()

			for i := range DrawGUIs {
				DrawGUIs[i].draw()
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
