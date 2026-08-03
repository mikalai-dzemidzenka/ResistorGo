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
	GameTimeSec   float32
	PhysicsTime   float32
	FixedTimeStep float32 = 1.0 / 30.0
)

var (
	E1M1Wall1 rl.Texture2D
	TexTFRun  rl.Texture2D
	TexTFIdle rl.Texture2D
)

func LoadTextures() {
	E1M1Wall1 = rl.LoadTexture("resources/blocks/e1m1_wall1.png")
	TexTFRun = rl.LoadTexture("resources/threefour/run.png")
	TexTFIdle = rl.LoadTexture("resources/threefour/idle.png")
}

func DrawGUI() {
	rlg.Button(rl.Rectangle{0, 0, 200, 200}, "Press me")
}

type Vector2i struct {
	X int32
	Y int32
}

func v2i(v rl.Vector2) Vector2i {
	return Vector2i{
		X: int32(v.X),
		Y: int32(v.Y),
	}
}

func Game() {
	rl.InitWindow(SW, SH, "Resistor")
	defer rl.CloseWindow()

	//rl.DisableCursor()

	rlg.LoadStyle("resources/style_dark.rgs")

	rl.SetTargetFPS(60)

	LoadTextures()

	cam := NewCamera(3)
	pl := NewActor()
	le := NewLevelEditor(32, 83)

	for !rl.WindowShouldClose() {
		if rl.IsKeyPressed(rl.KeyGrave) {
			Debug = !Debug
		}

		dt := rl.GetFrameTime()
		if dt > 0.25 {
			dt = 0.25
		}

		GameTimeSec += dt

		for PhysicsTime < GameTimeSec {
			PhysicsTime += FixedTimeStep
			pl.UpdatePhysics()
		}

		dt = (PhysicsTime - GameTimeSec) / FixedTimeStep
		pl.Update(dt)

		rl.BeginDrawing()
		{
			rl.ClearBackground(rl.Black)
			rl.BeginMode2D(cam.Camera2D)
			{
				le.DrawCursor(cam)
				cam.SetTarget(pl.Pos)
				cam.Update(dt)
				pl.Draw()
				le.DrawGrid()
			}
			rl.EndMode2D()

			DrawGUI()

			if Debug {
				fps := strconv.Itoa(int(rl.GetFPS()))
				rl.DrawText("fps:"+fps, 20, 20, 2, rl.White)
				rl.DrawText(fmt.Sprintf("gameTime:%f", GameTimeSec), 20, 30, 2, rl.White)
				rl.DrawText(fmt.Sprintf("physTime:%f", PhysicsTime), 20, 40, 2, rl.White)
				rl.DrawText(fmt.Sprintf("camera zoom speed %f", cam.ZoomSpeed), 20, 50, 2, rl.White)
			}
		}
		rl.EndDrawing()
	}
}
