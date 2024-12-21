package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	SW = 800
	SH = 600
)

func main() {
	rl.InitWindow(SW, SH, "Resistor")
	defer rl.CloseWindow()

	Textures[TEXTURE_PLAYER_IDLE] = rl.LoadTexture("resources/idle.png")

	rl.SetTargetFPS(60)

	for !rl.WindowShouldClose() {
		if rl.IsKeyPressed(rl.KeyGrave) {
			Debug = !Debug
		}

		dt := rl.GetFrameTime()

		p.Update(dt)
		c.Update(dt)

		rl.BeginDrawing()
		{
			rl.ClearBackground(rl.Black)
			rl.BeginMode2D(c.cam)
			{
				p.Draw()
			}
			rl.EndMode2D()
		}
		rl.EndDrawing()
	}
}
