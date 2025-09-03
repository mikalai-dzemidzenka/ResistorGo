package resistor

import (
	"fmt"
	rlg "github.com/gen2brain/raylib-go/raygui"
	rl "github.com/gen2brain/raylib-go/raylib"
	"strconv"
)

const (
	SW = 1280
	SH = 720
)

var (
	Debug         = true
	GameTime      float32
	PhysicsTime   float32
	FixedTimeStep float32 = 1.0 / 30.0
)

var (
	TexTFRun rl.Texture2D
)

func LoadTextures() {
	TexTFRun = rl.LoadTexture("resources/threefour/run.png")
}

func DrawGUI() {
}

func Game() {
	rl.InitWindow(SW, SH, "Resistor")
	defer rl.CloseWindow()

	rl.DisableCursor()

	rlg.LoadStyle("resources/style_dark.rgs")

	rl.SetTargetFPS(60)

	LoadTextures()

	cam := NewCamera()
	pl := NewActor()

	for !rl.WindowShouldClose() {
		if rl.IsKeyPressed(rl.KeyGrave) {
			Debug = !Debug
		}

		dt := rl.GetFrameTime()
		if dt > 0.25 {
			dt = 0.25
		}

		GameTime += dt

		for PhysicsTime < GameTime {
			PhysicsTime += FixedTimeStep
			pl.UpdatePhysics()
		}

		dt = (PhysicsTime - GameTime) / FixedTimeStep
		pl.Update(dt)

		rl.BeginDrawing()
		{
			rl.ClearBackground(rl.Black)
			rl.BeginMode2D(cam.Camera2D)
			{
				//cam.SetTarget(pl.Pos)
				cam.Update(dt)
				pl.Draw()
			}
			rl.EndMode2D()

			DrawGUI()

			if Debug {
				fps := strconv.Itoa(int(rl.GetFPS()))
				rl.DrawText("fps:"+fps, 20, 20, 2, rl.White)
				rl.DrawText(fmt.Sprintf("gameTime:%f", GameTime), 20, 30, 2, rl.White)
				rl.DrawText(fmt.Sprintf("physTime:%f", PhysicsTime), 20, 40, 2, rl.White)
				rl.DrawText(fmt.Sprintf("camera zoom speed %f", cam.ZoomSpeed), 20, 50, 2, rl.White)
			}
		}
		rl.EndDrawing()
	}
}
