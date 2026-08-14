package resistor

import (
	"fmt"
	rl "github.com/gen2brain/raylib-go/raylib"
	"math/rand"
	"strings"
)

const (
	DefaultAcc      = 1
	DefaultFriction = 2
	MaxSpeed        = 7
	Gravity         = 0.3
	MaxJumpSpeed    = -10
	MaxFallSpeed    = 4
)

type ActorState uint8

const (
	StateIdle ActorState = iota
	StateRun
	StateAttack
	StateAttacked
)

type Actor struct {
	Pos               rl.Vector2
	Speed             rl.Vector2
	Acc               float32
	Friction          float32
	Orientation       int8
	State             ActorState
	RewindData        [RewindFrames]ActorRewindFrame
	CollisionBox      rl.Rectangle
	AttackBoxHigh     rl.Rectangle
	AttackBoxMid      rl.Rectangle
	AttackBoxLow      rl.Rectangle
	OnGround          bool
	IsPlayer          bool
	Input             Input
	PrevUpdown        uint8
	Updown            uint8
	AnimGameTimeStart float32
	Anim              string
	AnimDuration      float32
	AnimOffset        uint8
	AnimTotalFrames   uint8
	AnimFrame         uint8
	AnimOnce          bool
	AnimFinished      bool
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
		State:       StateIdle,
		Orientation: 1,
		CollisionBox: rl.Rectangle{
			X:      11,
			Y:      6,
			Width:  17,
			Height: 42,
		},
		AttackBoxHigh: rl.Rectangle{
			12, 5, 24, 6,
		},
		AttackBoxMid: rl.Rectangle{
			12, 14, 24, 6,
		},
		AttackBoxLow: rl.Rectangle{
			12, 24, 24, 6,
		},
		IsPlayer:        isPlayer,
		Anim:            "resources/threefour/updown.png",
		AnimTotalFrames: 1,
		AnimDuration:    1,
		AnimFinished:    false,
		PrevUpdown:      2,
		Updown:          0,
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
	a.checkCollision()

	switch a.State {
	case StateIdle:
		a.updateUpdown(dt)
		a.updateHorSpeed(dt)
		a.updateVertSpeed(dt)
		if a.Input.Attack {
			a.enterAttackState()
		}
	case StateRun:
		a.updateHorSpeed(dt)
		a.updateVertSpeed(dt)
		if a.Input.Attack {
			a.enterAttackState()
		}
	case StateAttack:
		a.updateAttack()
	case StateAttacked:
		a.updateAttacked()
	}

	a.Pos.X += a.Speed.X
	a.Pos.Y += a.Speed.Y
}

func (a *Actor) updateRewind() {
	if IsRewinding {
		data := a.RewindData[RewindIdx%RewindFrames]
		a.Pos = data.Pos
		a.Speed = data.Speed
		a.Orientation = data.Orientation
	} else {
		a.RewindData[RewindIdx%RewindFrames] = ActorRewindFrame{
			Pos:         a.Pos,
			Speed:       a.Speed,
			Orientation: a.Orientation,
			Anim:        a.Anim,
			AnimFrame:   a.AnimFrame,
		}
	}
}
func (a *Actor) UpdatePhysics() {}

func (a *Actor) Draw() {
	if IsRewinding {
		data := a.RewindData[RewindIdx%RewindFrames]
		rl.DrawTextureRec(resources[data.Anim], rl.Rectangle{
			X:      float32(data.AnimFrame) * 48,
			Y:      0,
			Width:  48 * float32(a.Orientation),
			Height: 48,
		}, a.Pos, rl.White)
	} else {
		if a.AnimOnce {
			a.AnimFinished = a.DrawAnimOnce()
		} else {
			a.DrawAnim()
		}
	}

	if Debug {
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
		rl.DrawRectangleLinesEx(rl.Rectangle{
			X:      12 + a.Pos.X + a.AttackBoxHigh.X*float32(a.Orientation),
			Y:      a.Pos.Y + a.AttackBoxHigh.Y,
			Width:  a.AttackBoxHigh.Width,
			Height: a.AttackBoxHigh.Height,
		}, 1, rl.Red)
		rl.DrawRectangleLinesEx(rl.Rectangle{
			X:      12 + a.Pos.X + a.AttackBoxMid.X*float32(a.Orientation),
			Y:      a.Pos.Y + a.AttackBoxMid.Y,
			Width:  a.AttackBoxMid.Width,
			Height: a.AttackBoxMid.Height,
		}, 1, rl.Red)
		rl.DrawRectangleLinesEx(rl.Rectangle{
			X:      12 + a.Pos.X + a.AttackBoxLow.X*float32(a.Orientation),
			Y:      a.Pos.Y + a.AttackBoxLow.Y,
			Width:  a.AttackBoxLow.Width,
			Height: a.AttackBoxLow.Height,
		}, 1, rl.Red)
		if a.IsPlayer {
			rl.DrawText(fmt.Sprintf("RewindIndex: %d", RewindIdx), int32(a.Pos.X), int32(a.Pos.Y), 2, rl.White)
			rl.DrawText(fmt.Sprintf("RewindLeft: %d", RewindLeft), int32(a.Pos.X), int32(a.Pos.Y)-8, 2, rl.White)
			rl.DrawText(fmt.Sprintf("Mouse X, Y: %d %d", MouseWorldPos.X, MouseWorldPos.Y), int32(a.Pos.X), int32(a.Pos.Y)-16, 2, rl.White)
			rl.DrawText(fmt.Sprintf("PrevUpdown, Updown: %d %d", a.PrevUpdown, a.Updown), int32(a.Pos.X), int32(a.Pos.Y)-24, 2, rl.White)
			rl.DrawText(fmt.Sprintf("Frame, Finished: %d %t", a.AnimFrame, a.AnimFinished), int32(a.Pos.X), int32(a.Pos.Y)-32, 2, rl.White)
		}
		px := a.Pos.X + a.CollisionBox.X + a.CollisionBox.Width/2
		rl.DrawPixel(int32(px), int32(a.Pos.Y+a.CollisionBox.Y+a.CollisionBox.Height), rl.Red)
	}

}

func (a *Actor) SetAnimFrame(anim string, offset uint8) {
	a.AnimGameTimeStart = GameTimeSec
	a.Anim = anim
	a.AnimDuration = 1
	a.AnimTotalFrames = 1
	a.AnimOnce = false
	a.AnimOffset = offset
}

func (a *Actor) SetAnim(anim string, animDuration float32, totalFrames uint8, once bool) {
	a.AnimGameTimeStart = GameTimeSec
	a.Anim = anim
	a.AnimDuration = animDuration
	a.AnimTotalFrames = totalFrames
	a.AnimOnce = once
	a.AnimOffset = 0
}

func (a *Actor) DrawAnimOnce() bool {
	tex, ok := resources[a.Anim]
	if !ok {
		panic(a.Anim)
	}

	frameTimeMs := a.AnimDuration / float32(a.AnimTotalFrames)
	frame := int((GameTimeSec - a.AnimGameTimeStart) / frameTimeMs)
	finished := frame >= int(a.AnimTotalFrames)
	frame = clamp(frame, 0, int(a.AnimTotalFrames)-1)

	a.AnimFrame = uint8(frame)
	rl.DrawTextureRec(tex, rl.Rectangle{
		X:      float32(a.AnimFrame) * 48,
		Y:      0,
		Width:  48 * float32(a.Orientation),
		Height: 48,
	}, a.Pos, rl.White)
	return finished
}

func (a *Actor) DrawAnim() {
	tex, ok := resources[a.Anim]
	if !ok {
		panic(a.Anim)
	}
	frameTimeMs := a.AnimDuration / float32(a.AnimTotalFrames)
	if GameTimeSec-a.AnimGameTimeStart > a.AnimDuration {
		a.AnimGameTimeStart += a.AnimDuration
	}
	frame := int((GameTimeSec - a.AnimGameTimeStart) / frameTimeMs)
	frame = clamp(frame, 0, int(a.AnimTotalFrames)-1)
	a.AnimFrame = a.AnimOffset + uint8(frame)
	rl.DrawTextureRec(tex, rl.Rectangle{
		X:      float32(a.AnimFrame) * 48,
		Y:      0,
		Width:  48 * float32(a.Orientation),
		Height: 48,
	}, a.Pos, rl.White)
}

func (a *Actor) updateHorSpeed(dt float32) {
	if a.State != StateIdle && a.State != StateRun {
		return
	}
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
		a.enterIdleState()
	} else {
		a.enterRunState()
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
		} else {
			if rand.Int()%61 == 3 {
				a.Input.Attack = true
			}
		}
		a.Orientation = -1
	}
	if Pl.Pos.X > a.Pos.X {
		if Pl.Pos.X-a.Pos.X > near {
			a.Input.Left = true
		} else {
			if rand.Int()%61 == 3 {
				a.Input.Attack = true
			}
		}
		a.Orientation = 1
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
	if rl.IsMouseButtonPressed(rl.MouseButtonLeft) {
		a.Input.Attack = true
	}
}

func (a *Actor) updateUpdown(dt float32) {
	if a.State != StateIdle || !a.AnimFinished && strings.Contains(a.Anim, "updown-") {
		return
	}
	var updown uint8
	switch {
	case float32(MouseWorldPos.Y) <= a.Pos.Y+14:
		updown = 2
	case float32(MouseWorldPos.Y) <= a.Pos.Y+32:
		updown = 1
	default:
		updown = 0
	}
	if updown != a.Updown {
		a.PrevUpdown = a.Updown
		anim := fmt.Sprintf("resources/threefour/updown-%d-%d.png", a.PrevUpdown, updown)
		a.SetAnim(anim, 0.1, 4, true)
	} else if a.AnimFinished && strings.Contains(a.Anim, "updown-") {
		a.SetAnimFrame("resources/threefour/updown.png", a.Updown)
	}
	a.Updown = updown
}

func (a *Actor) enterAttackState() {
	a.State = StateAttack
	switch a.Updown {
	case 0:
		a.SetAnim("resources/threefour/attack-low.png", 0.4, 8, true)
	case 1:
		a.SetAnim("resources/threefour/attack-mid.png", 0.4, 8, true)
	case 2:
		a.SetAnim("resources/threefour/attack-high.png", 0.4, 8, true)
	}
}

func (a *Actor) enterAttackedState() {
	a.State = StateAttacked
	a.SetAnim("resources/threefour/hit-top.png", 0.5, 4, true)
}

func (a *Actor) enterIdleState() {
	a.State = StateIdle
	a.Speed.X = 0
	if !strings.Contains(a.Anim, "updown") {
		a.SetAnimFrame("resources/threefour/updown.png", a.Updown)
	}
}

func (a *Actor) enterRunState() {
	a.State = StateRun
	if a.Anim != "resources/threefour/run.png" {
		a.SetAnim("resources/threefour/run.png", 0.4, 10, false)
	}
}

func (a *Actor) updateAttacked() {
	if a.State != StateAttacked {
		return
	}
	if a.AnimFinished {
		a.enterIdleState()
	}
}

func (a *Actor) updateAttack() {
	if a.State != StateAttack {
		return
	}
	var r1 rl.Rectangle
	switch a.Updown {
	case 2:
		r1 = rl.Rectangle{
			X:      12 + a.Pos.X + a.AttackBoxHigh.X*float32(a.Orientation),
			Y:      a.Pos.Y + a.AttackBoxHigh.Y,
			Width:  a.AttackBoxHigh.Width,
			Height: a.AttackBoxHigh.Height,
		}
	case 1:
		r1 = rl.Rectangle{
			X:      12 + a.Pos.X + a.AttackBoxMid.X*float32(a.Orientation),
			Y:      a.Pos.Y + a.AttackBoxMid.Y,
			Width:  a.AttackBoxMid.Width,
			Height: a.AttackBoxMid.Height,
		}
	case 0:
		r1 = rl.Rectangle{
			X:      12 + a.Pos.X + a.AttackBoxLow.X*float32(a.Orientation),
			Y:      a.Pos.Y + a.AttackBoxLow.Y,
			Width:  a.AttackBoxLow.Width,
			Height: a.AttackBoxLow.Height,
		}
	}

	for _, actor := range Actors {
		if a == actor {
			continue
		}
		r2 := rl.Rectangle{
			X:      actor.Pos.X + a.CollisionBox.X,
			Y:      actor.Pos.Y + a.CollisionBox.Y,
			Width:  a.CollisionBox.Width,
			Height: a.CollisionBox.Height,
		}
		if aabb(r1, r2) && actor.State != StateAttacked {
			actor.enterAttackedState()
		}
	}

	if a.AnimFinished {
		a.State = StateIdle
		return
	}
}
