package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	TEXTURE_PLAYER_IDLE = iota
	TEXTURE_PLAYER_LEN
)

var (
	Debug       = false
	DebugColor1 = rl.NewColor(0, 100, 230, 50)
	DebugColor2 = rl.NewColor(0, 100, 255, 255)
)

var Textures [TEXTURE_PLAYER_LEN]rl.Texture2D

type Player struct {
	pos         rl.Vector2
	bound       rl.Rectangle
	anim        int
	orientation int8
}

var p = Player{
	pos: rl.Vector2{
		X: 200, Y: 200,
	},
	bound:       rl.Rectangle{Width: 39, Height: 44},
	anim:        TEXTURE_PLAYER_IDLE,
	orientation: 1,
}

var c = rl.Camera2D{
	Offset:   rl.Vector2{X: SW / 2, Y: SH / 2},
	Target:   p.pos,
	Rotation: 0,
	Zoom:     4,
}

func (p *Player) Update() {
	switch {
	case rl.IsKeyDown(rl.KeyLeft):
		p.pos.X -= 1
		p.orientation = -1
	case rl.IsKeyDown(rl.KeyRight):
		p.pos.X += 1
		p.orientation = 1
	}
}

func (p *Player) Draw() {
	anim := Textures[p.anim]
	if Debug {
		rl.DrawRectangle(int32(p.pos.X), int32(p.pos.Y), 39, 44, DebugColor1)
		rl.DrawRectangleLines(int32(p.pos.X), int32(p.pos.Y), 39, 44, DebugColor2)
	}
	rl.DrawTextureRec(
		anim,
		rl.Rectangle{
			X:      0,
			Y:      0,
			Width:  float32(anim.Width * int32(p.orientation)),
			Height: float32(anim.Height),
		},
		p.pos,
		rl.White,
	)
}

const (
	SW = 800
	SH = 600
)

func main() {
	rl.InitWindow(SW, SH, "raylib [core] example - basic window")
	defer rl.CloseWindow()

	Textures[TEXTURE_PLAYER_IDLE] = rl.LoadTexture("resources/idle.png")

	rl.SetTargetFPS(60)

	for !rl.WindowShouldClose() {
		if rl.IsKeyPressed(rl.KeyGrave) {
			Debug = !Debug
		}

		p.Update()

		rl.BeginDrawing()
		{
			rl.ClearBackground(rl.Black)
			rl.BeginMode2D(c)
			{
				p.Draw()
			}
			rl.EndMode2D()
		}
		rl.EndDrawing()
	}
}
