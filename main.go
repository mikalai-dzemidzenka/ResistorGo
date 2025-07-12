package main

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
	ZeroVec2      = rl.Vector2{}
	Camera        = rl.NewCamera2D(ZeroVec2, ZeroVec2, 0, 1.0)
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

func Update(dt float32) {

}

func UpdatePhysics() {

}

func Draw() {
	PlayAnim(TexTFRun, 24, 10)
}

func PlayAnim(tex rl.Texture2D, frameTime float32, totalFrames int) {
	frameTimeMs := 1000 / frameTime
	rl.DrawTextureRec(TexTFRun, rl.Rectangle{
		X:      float32(int(GameTime*1000/frameTimeMs)%totalFrames) * 48,
		Y:      0,
		Width:  48,
		Height: 48,
	}, rl.Vector2{SW / 2, SH / 2}, rl.White)
}

func DrawGUI() {
}

func main() {
	rl.InitWindow(SW, SH, "Resistor")
	defer rl.CloseWindow()

	rl.DisableCursor()

	rlg.LoadStyle("resources/style_dark.rgs")

	rl.SetTargetFPS(60)

	LoadTextures()

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
			UpdatePhysics()
		}

		dt = (PhysicsTime - GameTime) / FixedTimeStep
		Update(dt)

		rl.BeginDrawing()
		{
			rl.ClearBackground(rl.Black)
			rl.BeginMode2D(Camera)
			{
				Draw()
			}
			rl.EndMode2D()

			DrawGUI()

			if Debug {
				fps := strconv.Itoa(int(rl.GetFPS()))
				rl.DrawText("fps:"+fps, 20, 20, 2, rl.White)
				rl.DrawText(fmt.Sprintf("gameTime:%f", GameTime), 20, 30, 2, rl.White)
				rl.DrawText(fmt.Sprintf("physTime:%f", PhysicsTime), 20, 40, 2, rl.White)
			}
		}
		rl.EndDrawing()
	}
}
