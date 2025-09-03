package resistor

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	DefaultAcc      = 4
	DefaultFriction = 0.7
	MaxSpeed        = 7
)

type Actor struct {
	Pos         rl.Vector2
	Speed       rl.Vector2
	Acc         float32
	Friction    float32
	Orientation int8
	State       int8 // todo enum
	RewindData  [RewindFrames]ActorRewindFrame
}

func NewActor() *Actor {
	return &Actor{
		Pos:         rl.Vector2{},
		Speed:       rl.Vector2{},
		Acc:         DefaultAcc,
		Friction:    DefaultFriction,
		Orientation: 1,
	}
}

func (a *Actor) Update(dt float32) {
	if IsRewinding {
		return
	}
	if rl.IsKeyDown(rl.KeyA) {
		a.Speed.X -= a.Acc * dt
		a.Orientation = -1
	}
	if rl.IsKeyDown(rl.KeyD) {
		a.Speed.X += a.Acc * dt
		a.Orientation = 1
	}
	a.Speed.X = clamp(a.Speed.X, -MaxSpeed, MaxSpeed)

	a.Speed.X = moveTo0(a.Speed.X, a.Friction*dt) // friction

	if abs(a.Speed.X) < 0.001 {
		a.Speed.X = 0
	}
	a.Pos.X += a.Speed.X
}

func (a *Actor) UpdatePhysics() {
	if rl.IsKeyDown(rl.KeyR) && RewindLeft > 0 {
		IsRewinding = true
		data := a.RewindData[RewindIdx%RewindFrames]
		a.Pos = data.Pos
		a.Speed = data.Speed
		a.Orientation = data.Orientation
		RewindIdx--
		RewindLeft--
	} else {
		IsRewinding = false
		a.RewindData[RewindIdx%RewindFrames] = ActorRewindFrame{
			Pos:         a.Pos,
			Speed:       a.Speed,
			Orientation: a.Orientation,
		}
		RewindIdx++
		if RewindLeft < RewindFrames {
			RewindLeft++
		}
	}
}

func (a *Actor) Draw() {
	a.PlayAnim(TexTFRun, 24, 10)
}

func (a *Actor) PlayAnim(tex rl.Texture2D, frameTime float32, totalFrames int) {
	frameTimeMs := 1000 / frameTime
	rl.DrawTextureRec(tex, rl.Rectangle{
		X:      float32(int(GameTime*1000/frameTimeMs)%totalFrames) * 48,
		Y:      0,
		Width:  48 * float32(a.Orientation),
		Height: 48,
	}, a.Pos, rl.White)
}
