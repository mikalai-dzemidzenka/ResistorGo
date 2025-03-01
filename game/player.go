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
	Camera          rl.Camera2D
	Bound           rl.Rectangle `json:"bound"`
	Pos             rl.Vector2   `json:"pos"`
	Speed           rl.Vector2   `json:"speed"`
	topRay          rl.Vector2
	botRay          rl.Vector2
	Collider        rl.Rectangle
	Gravity         float32
	Acc             float32
	OnGround        bool
	CameraZoomSpeed float32
	Friction        float32 `json:"friction,omitempty"`
	Anim            int32   `json:"anim,omitempty"`
	// Orient 1 right, -1 left
	Orient int8 `json:"orient,omitempty"`
	Z      int8 `json:"z,omitempty"`
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

const (
	MinZoom = 4
	MaxZoom = 6
)

func (p *Player) Update(dt float32) {
	in := p.handleInput()
	var acc rl.Vector2
	if in.jumpPressed && p.OnGround {
		p.Speed.Y -= 0.5
		p.OnGround = false
	}

	acc.X += p.Acc * in.moveX

	if in.moveX != 0 { // move
		p.Orient = int8(in.moveX)
		p.Speed = rl.Vector2Add(p.Speed, acc)
	} else { // slowdown
		switch {
		case abs(p.Speed.X) < 0.2:
			p.Speed.X = 0
		case p.Speed.X > 0:
			p.Speed.X -= p.Friction
		case p.Speed.X < 0:
			p.Speed.X += p.Friction
		}

	}

	p.Speed.X = clamp(p.Speed.X, -MaxSpeed, MaxSpeed)

	s := g.Solids[0]
	if p.OnGround {
		p.Speed.Y = 0

		p.Collider = rl.GetCollisionRec(p.Bound, s.Col)
	} else {
		if !DebugDisableGravity {
			p.Speed.Y += p.Gravity
		}
		p.Collider = rl.Rectangle{}
	}

	p.Pos = rl.Vector2Add(p.Pos, p.Speed)
	p.Pos.Y = float32(math.Round(float64(p.Pos.Y)))

	p.Bound.X = p.Pos.X
	p.Bound.Y = p.Pos.Y

	p.topRay = rl.Vector2{p.Pos.X + p.Bound.Width/2, p.Pos.Y}
	p.botRay = rl.Vector2{p.Pos.X + p.Bound.Width/2, p.Pos.Y + p.Bound.Height}
	i := (p.botRay.Y - p.topRay.Y) / AttackIdleLen
	h := rl.GetScreenToWorld2D(g.Cursor.AtkPos, *g.Camera).Y - p.topRay.Y

	idx := clamp(h/i, TEXTURE_PLAYER_IDLE_1, TEXTURE_PLAYER_IDLE_4)
	p.Anim = int32(idx)

	p.UpdateCamera(dt)
}

func (p *Player) handleInput() (result inputs) {
	switch {
	case rl.IsKeyDown(rl.KeySpace):
		result.jumpPressed = true
	}
	switch {
	case rl.IsKeyDown(rl.KeyLeft), rl.IsKeyDown(rl.KeyA):
		result.moveX -= 1
	}
	switch {
	case rl.IsKeyDown(rl.KeyRight), rl.IsKeyDown(rl.KeyD):
		result.moveX += 1
	}
	return result
}

type inputs struct {
	moveX       float32
	jumpPressed bool
}

func (p *Player) UpdateCamera(dt float32) {
	p.CameraZoomSpeed += rl.GetMouseWheelMove() * dt * 5
	p.Camera.Zoom += p.CameraZoomSpeed
	p.CameraZoomSpeed = moveTo0(p.CameraZoomSpeed, 1.2*dt)
	if !Debug {
		p.Camera.Zoom = clamp(p.Camera.Zoom, MinZoom, MaxZoom)
	} else {
		p.Camera.Zoom = clamp(p.Camera.Zoom, 0.5, MaxZoom)
	}
	p.Camera.Target = rl.Vector2Add(g.Player.Pos, rl.Vector2{X: 50})
}

func (p *Player) Draw() {
	anim := Textures[p.Anim]

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
		// collision
		rl.DrawRectangle(int32(p.Pos.X), int32(p.Pos.Y), 39, 44, DebugColor1)
		rl.DrawRectangleLinesEx(p.Bound, 1, DebugColor2)
		if p.Collider != (rl.Rectangle{}) {
			rl.DrawRectangleLinesEx(p.Collider, 1, rl.Red)
		}

		// player centers
		rl.DrawCircle(int32(p.Pos.X+p.Bound.Width/2), int32(p.Pos.Y+p.Bound.Height/2), 2, rl.Red)

		// top and bottom attack ray bound
		rl.DrawCircle(int32(p.topRay.X), int32(p.topRay.Y), 0.5, rl.SkyBlue)
		rl.DrawCircle(int32(p.botRay.X), int32(p.botRay.Y), 0.5, rl.SkyBlue)
	}
}

func (p *Player) Debug() string {
	if p == nil {
		return ""
	}
	return fmt.Sprintf("speed: %v\npos: %v\n", p.Speed, p.Pos)
}
