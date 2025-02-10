package game

import (
	"encoding/json"
	"fmt"
	rl "github.com/gen2brain/raylib-go/raylib"
	"math"
)

const (
	DefaultAcc      = 0.3
	MaxSpeed        = 2.5
	DefaultFriction = 0.4
)

type Player struct {
	Bound    rl.Rectangle `json:"bound"`
	Pos      rl.Vector2   `json:"pos"`
	Speed    rl.Vector2   `json:"speed"`
	Friction float32      `json:"friction,omitempty"`
	Anim     int32        `json:"anim,omitempty"`
	Orient   int8         `json:"orient,omitempty"`
	Z        int8         `json:"z,omitempty"`
	Paused   bool         `json:"paused,omitempty"`
}

func (p *Player) MarshalJSON() ([]byte, error) {
	type P Player
	return json.Marshal(EntitySerde{
		Type:   ENTITY_PLAYER,
		Entity: (*P)(p),
	})
}

func (p *Player) GetZ() int8 {
	return p.Z
}

var _ Serializable = (*Player)(nil)
var _ Updatable = (*Player)(nil)
var _ Drawable = (*Player)(nil)

func (p *Player) Update(dt float32) {
	var acc rl.Vector2

	var noInput bool
	switch {
	case rl.IsKeyDown(rl.KeyLeft), rl.IsKeyDown(rl.KeyA):
		acc.X -= DefaultAcc
		p.Orient = -1
	case rl.IsKeyDown(rl.KeyRight), rl.IsKeyDown(rl.KeyD):
		acc.X += DefaultAcc
		p.Orient = 1
	default:
		noInput = true
	}

	if noInput {
		switch {
		case p.Speed.X > 0:
			p.Speed.X -= p.Friction
		case p.Speed.X < 0:
			p.Speed.X += p.Friction
		}
		if abs(p.Speed.X) < 0.2 {
			p.Speed.X = 0
		}
	} else {
		p.Speed = rl.Vector2Add(p.Speed, acc)
	}

	p.Speed.X = clamp(p.Speed.X, -MaxSpeed, MaxSpeed)

	p.Anim = TEXTURE_PLAYER_IDLE_1 + g.Cursor.AtkSt

	p.Pos = rl.Vector2Add(p.Pos, p.Speed)
	p.Pos.Y = float32(math.Round(float64(p.Pos.Y)))
}

func (p *Player) Draw() {
	anim := Textures[p.Anim]
	if Debug {
		rl.DrawRectangle(int32(p.Pos.X), int32(p.Pos.Y), 39, 44, DebugColor1)
		rl.DrawRectangleLines(int32(p.Pos.X), int32(p.Pos.Y), 39, 44, DebugColor2)
	}
	rl.DrawTextureRec(
		anim,
		rl.Rectangle{
			X:      0,
			Y:      0,
			Width:  float32(anim.Width * int32(p.Orient)),
			Height: float32(anim.Height),
		},
		p.Pos,
		rl.White,
	)
	if Debug {
		rl.DrawCircle(int32(p.Pos.X+p.Bound.Width/2), int32(p.Pos.Y+p.Bound.Height/2), 2, rl.Red)
		rl.DrawCircle(int32(p.Pos.X+p.Bound.Width/2), int32(p.Pos.Y+p.Bound.Height*1/4), 1, rl.Green)
	}
}

func (p *Player) Debug() string {
	return fmt.Sprintf("speed: %v\npos: %v\n", p.Speed, p.Pos)
}
