package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"strconv"
)

func lvl1Loop(dt float32) {
	cur.Update(dt)
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

		cur.DrawGui()
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
