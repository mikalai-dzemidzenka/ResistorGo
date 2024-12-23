package main

import (
	"fmt"
	rl "github.com/gen2brain/raylib-go/raylib"
)

type Player struct {
	pos         rl.Vector2
	bound       rl.Rectangle
	speed       rl.Vector2
	friction    float32
	anim        int
	orientation int8
}

const (
	DefaultAcc      = 0.3
	MaxSpeed        = 2.5
	DefaultFriction = 0.4
)

var p = Player{
	pos: rl.Vector2{
		X: 200, Y: 200,
	},
	bound:       rl.Rectangle{Width: 39, Height: 44},
	speed:       rl.Vector2{0, 0},
	friction:    DefaultFriction,
	anim:        TEXTURE_PLAYER_IDLE,
	orientation: 1,
}

func (p *Player) Update(dt float32) {
	var acc rl.Vector2

	var noInput bool
	switch {
	case rl.IsKeyDown(rl.KeyLeft), rl.IsKeyDown(rl.KeyA):
		acc.X -= DefaultAcc
		p.orientation = -1
	case rl.IsKeyDown(rl.KeyRight), rl.IsKeyDown(rl.KeyD):
		acc.X += DefaultAcc
		p.orientation = 1
	default:
		noInput = true
	}

	if noInput {
		switch {
		case p.speed.X > 0:
			p.speed.X -= p.friction
		case p.speed.X < 0:
			p.speed.X += p.friction
		}
		if abs(p.speed.X) < 0.2 {
			p.speed.X = 0
		}
	} else {
		p.speed = rl.Vector2Add(p.speed, acc)
	}

	p.speed.X = clamp(p.speed.X, -MaxSpeed, MaxSpeed)

	p.pos = rl.Vector2Add(p.pos, p.speed)
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
	if Debug {
		rl.DrawCircle(int32(p.pos.X+p.bound.Width/2), int32(p.pos.Y+p.bound.Height/2), 2, rl.Red)
		rl.DrawCircle(int32(p.pos.X+p.bound.Width/2), int32(p.pos.Y+p.bound.Height*1/4), 1, rl.Green)
	}
}

func (p *Player) Debug() string {
	return fmt.Sprintf("speed: %v\npos: %v\n", p.speed, p.pos)
}
