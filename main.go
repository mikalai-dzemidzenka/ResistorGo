package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	SW = 1024
	SH = 768
)

func main() {
	rl.InitWindow(SW, SH, "Resistor")
	defer rl.CloseWindow()

	rl.DisableCursor()

	Textures[TEXTURE_PLAYER_IDLE] = rl.LoadTexture("resources/idle.png")
	Textures[TEXTURE_PLAYER_IDLE_1] = rl.LoadTexture("resources/idle-1.png")
	Textures[TEXTURE_PLAYER_IDLE_2] = rl.LoadTexture("resources/idle-2.png")
	Textures[TEXTURE_PLAYER_IDLE_3] = rl.LoadTexture("resources/idle-3.png")
	Textures[TEXTURE_PLAYER_IDLE_4] = rl.LoadTexture("resources/idle-4.png")

	rl.SetTargetFPS(60)

	for !rl.WindowShouldClose() {
		if rl.IsKeyPressed(rl.KeyGrave) {
			Debug = !Debug
		}

		dt := rl.GetFrameTime()

		cur.Update()
		p.Update(dt)
		cam.Update(dt)

		rl.BeginDrawing()
		{
			rl.ClearBackground(rl.Black)
			rl.BeginMode2D(cam.cam)
			{
				p.Draw()
			}
			rl.EndMode2D()

			cur.Draw()
			if Debug {
				rl.DrawText(p.Debug(), 40, 40, 1, rl.Red)
				rl.DrawText(cam.Debug(), 40, 78, 1, rl.Green)
				rl.DrawText(cur.Debug(), 40, 116, 1, rl.Blue)
			}
		}
		rl.EndDrawing()
	}
}
