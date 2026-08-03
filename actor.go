package resistor

import (
	"fmt"
	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	DefaultAcc      = 4
	DefaultFriction = 0.7
	MaxSpeed        = 7
)

type ActorState uint8

const (
	Idle ActorState = iota
	Run
)

type Actor struct {
	Pos         rl.Vector2
	Speed       rl.Vector2
	Acc         float32
	Friction    float32
	Orientation int8
	State       ActorState
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

func (a *Actor) updateRewind() {
	if rl.IsKeyPressed(rl.KeyR) {
		RewindIdx--
		RewindLeft--
		data := a.RewindData[RewindIdx%RewindFrames]
		a.Pos = data.Pos
		a.Speed = data.Speed
		a.Orientation = data.Orientation
	}
	if rl.IsKeyDown(rl.KeyR) && RewindLeft > 0 {
		IsRewinding = true
		RewindIdx--
		RewindLeft--
		data := a.RewindData[RewindIdx%RewindFrames]
		a.Pos = data.Pos
		a.Speed = data.Speed
		a.Orientation = data.Orientation
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

func (a *Actor) Update(dt float32) {
	a.updateRewind()

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
func (a *Actor) UpdatePhysics() {}

func (a *Actor) Draw() {
	if a.State == Idle {
		rl.DrawTextureRec(TexTFIdle, rl.Rectangle{
			X:      0,
			Y:      0,
			Width:  48 * float32(a.Orientation),
			Height: 48,
		}, a.Pos, rl.White)
	} else {
		a.PlayAnim(TexTFRun, 24, 10)
	}
	rl.DrawRectangleLinesEx(rl.Rectangle{
		X:      a.Pos.X,
		Y:      a.Pos.Y,
		Width:  48,
		Height: 48,
	}, 1, rl.Blue)
	rl.DrawText(fmt.Sprintf("RewindIndex: %d", RewindIdx), int32(a.Pos.X), int32(a.Pos.Y), 2, rl.White)
	rl.DrawText(fmt.Sprintf("RewindLeft: %d", RewindLeft), int32(a.Pos.X), int32(a.Pos.Y)-8, 2, rl.White)
}

func (a *Actor) PlayAnim(tex rl.Texture2D, framesPerSecond float32, totalFrames int) {
	frameTimeMs := 1000 / framesPerSecond
	rl.DrawTextureRec(tex, rl.Rectangle{
		X:      float32(int(GameTimeSec*1000/frameTimeMs)%totalFrames) * 48,
		Y:      0,
		Width:  48 * float32(a.Orientation),
		Height: 48,
	}, a.Pos, rl.White)
}
