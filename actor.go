package resistor

import (
	"fmt"
	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	DefaultAcc      = 4
	DefaultFriction = 0.7
	MaxSpeed        = 7
	Gravity         = 0.3
	MaxJumpSpeed    = -10
	MaxFallSpeed    = 4
)

type ActorState uint8

const (
	Idle ActorState = iota
	Run
)

type Actor struct {
	Pos          rl.Vector2
	Speed        rl.Vector2
	Acc          float32
	Friction     float32
	Orientation  int8
	State        ActorState
	RewindData   [RewindFrames]ActorRewindFrame
	CollisionBox rl.Rectangle
	OnGround     bool
	IsPlayer     bool
	Input        Input
}
type Input struct {
	Left   bool
	Right  bool
	Attack bool
}

func NewActor(pos rl.Vector2, isPlayer bool) *Actor {
	return &Actor{
		Pos:         pos,
		Speed:       rl.Vector2{},
		Acc:         DefaultAcc,
		Friction:    DefaultFriction,
		Orientation: 1,
		CollisionBox: rl.Rectangle{
			X:      11,
			Y:      6,
			Width:  17,
			Height: 42,
		},
		IsPlayer: isPlayer,
	}
}

func (a *Actor) Update(dt float32) {
	a.updateRewind()

	if IsRewinding {
		return
	}
	if a.IsPlayer {
		a.updatePlayerInput()
	} else {
		a.updateBotInput()
	}
	a.updateHorSpeed(dt)
	a.updateVertSpeed(dt)
	a.checkCollision()

	a.Pos.X += a.Speed.X
	a.Pos.Y += a.Speed.Y
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
func (a *Actor) UpdatePhysics() {}

func (a *Actor) Draw() {
	if a.State == Idle {
		rl.DrawTextureRec(resources["resources/threefour/idle.png"], rl.Rectangle{
			X:      0,
			Y:      0,
			Width:  48 * float32(a.Orientation),
			Height: 48,
		}, a.Pos, rl.White)
	} else {
		absSpeed := abs(a.Speed.X)
		var fps float32 = 24
		if absSpeed < MaxSpeed {
			fps = 12
		}
		a.PlayAnim(resources["resources/threefour/run.png"], fps, 10)
	}
	rl.DrawRectangleLinesEx(rl.Rectangle{
		X:      a.Pos.X,
		Y:      a.Pos.Y,
		Width:  48,
		Height: 48,
	}, 1, rl.Blue)
	rl.DrawRectangleLinesEx(rl.Rectangle{
		X:      a.Pos.X + a.CollisionBox.X,
		Y:      a.Pos.Y + a.CollisionBox.Y,
		Width:  a.CollisionBox.Width,
		Height: a.CollisionBox.Height,
	}, 1, rl.SkyBlue)
	rl.DrawText(fmt.Sprintf("RewindIndex: %d", RewindIdx), int32(a.Pos.X), int32(a.Pos.Y), 2, rl.White)
	rl.DrawText(fmt.Sprintf("RewindLeft: %d", RewindLeft), int32(a.Pos.X), int32(a.Pos.Y)-8, 2, rl.White)
	px := a.Pos.X + a.CollisionBox.X + a.CollisionBox.Width/2
	rl.DrawPixel(int32(px), int32(a.Pos.Y+a.CollisionBox.Y+a.CollisionBox.Height), rl.Red)
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

func (a *Actor) updateHorSpeed(dt float32) {
	var movePressed bool
	if a.Input.Right {
		movePressed = true
		a.Speed.X -= a.Acc * dt
		a.Orientation = -1
	}
	if a.Input.Left {
		movePressed = true
		a.Speed.X += a.Acc * dt
		a.Orientation = 1
	}
	if !movePressed {
		a.Speed.X = moveTo0(a.Speed.X, a.Friction*dt) // friction
	}
	a.Speed.X = clamp(a.Speed.X, -MaxSpeed, MaxSpeed)

	absSpeed := abs(a.Speed.X)
	if absSpeed < 0.001 {
		a.State = Idle
		a.Speed.X = 0
	} else {
		a.State = Run
	}
}

func (a *Actor) updateVertSpeed(dt float32) {
	if !a.OnGround {
		a.Speed.Y += Gravity * dt
	}
	a.Speed.Y = clamp(a.Speed.Y, MaxJumpSpeed, MaxFallSpeed)
}

func (a *Actor) checkCollision() {
	for _, obj := range LE.objectsIndexed {
		oColY := float32(obj.Y + int32(obj.CollisionRect.Y))
		oColX := float32(obj.X + int32(obj.CollisionRect.X))

		pColY := a.Pos.Y + a.CollisionBox.Y + a.CollisionBox.Height
		pColX := a.Pos.X + a.CollisionBox.X + a.CollisionBox.Width/2

		if oColX <= pColX && oColX+obj.CollisionRect.Width >= pColX &&
			pColY >= oColY && pColY <= oColY+obj.CollisionRect.Height {
			a.OnGround = true
			a.Speed.Y = 0
			a.Pos.Y = oColY - (a.CollisionBox.Y + a.CollisionBox.Height)
			return
		}
	}
	a.OnGround = false
}

func (a *Actor) updateBotInput() {
	const near = 50
	a.Input = Input{}
	if Pl.Pos.X < a.Pos.X {
		if a.Pos.X-Pl.Pos.X > near {
			a.Input.Right = true
		}
	}
	if Pl.Pos.X > a.Pos.X {
		if Pl.Pos.X-a.Pos.X > near {
			a.Input.Left = true
		}
	}
}

func (a *Actor) updatePlayerInput() {
	a.Input = Input{}
	if rl.IsKeyDown(rl.KeyA) {
		a.Input.Right = true
	}
	if rl.IsKeyDown(rl.KeyD) {
		a.Input.Left = true
	}
}
