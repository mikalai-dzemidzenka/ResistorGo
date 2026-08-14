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
	MaxAttackSpeed  = 2
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
	StateBlock
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
	AnimPhysTimeStart float64
	Anim              string
	AnimDuration      float64
	AnimOffset        uint8
	AnimTotalFrames   uint8
	AnimFrame         uint8
	AnimOnce          bool
	AnimFinished      bool
	Attacker          *Actor
	AttackQueued      bool
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
			X:      14,
			Y:      6,
			Width:  14,
			Height: 42,
		},
		AttackBoxHigh: rl.Rectangle{
			0, 5, 24, 6,
		},
		AttackBoxMid: rl.Rectangle{
			0, 14, 24, 6,
		},
		AttackBoxLow: rl.Rectangle{
			0, 24, 24, 6,
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
	a.updateHorSpeed(dt)
	a.updateVertSpeed(dt)

	switch a.State {
	case StateIdle:
		a.updateUpdown()
		absSpeed := abs(a.Speed.X)
		if absSpeed < 0.001 {
			a.enterIdleState()
		} else {
			a.enterRunState()
		}
		if a.Input.Attack {
			a.enterAttackState()
		}
	case StateRun:
		absSpeed := abs(a.Speed.X)
		if absSpeed < 0.001 {
			a.enterIdleState()
		} else {
			a.enterRunState()
		}
		if a.Input.Attack {
			a.enterAttackState()
		}
	case StateAttack:
		a.Speed.X = clamp(a.Speed.X, -MaxAttackSpeed, MaxAttackSpeed)
		a.updateAttack()
	case StateAttacked:
		a.Speed.X = clamp(a.Speed.X, -MaxAttackSpeed, MaxAttackSpeed)
		a.updateAttacked()
	case StateBlock:
		a.Speed.X = clamp(a.Speed.X, -MaxAttackSpeed, MaxAttackSpeed)
		a.updateBlock()
	}

	a.Pos.X += a.Speed.X
	a.Pos.Y += a.Speed.Y
	a.Pos.X = float32(int(a.Pos.X))
	a.Pos.Y = float32(int(a.Pos.Y))
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
			X:      a.Pos.X + a.AttackBoxHigh.X + 12*float32(a.Orientation+1),
			Y:      a.Pos.Y + a.AttackBoxHigh.Y,
			Width:  a.AttackBoxHigh.Width,
			Height: a.AttackBoxHigh.Height,
		}, 1, rl.Red)
		rl.DrawRectangleLinesEx(rl.Rectangle{
			X:      a.Pos.X + a.AttackBoxMid.X + 12*float32(a.Orientation+1),
			Y:      a.Pos.Y + a.AttackBoxMid.Y,
			Width:  a.AttackBoxMid.Width,
			Height: a.AttackBoxMid.Height,
		}, 1, rl.Red)
		rl.DrawRectangleLinesEx(rl.Rectangle{
			X:      a.Pos.X + a.AttackBoxLow.X + 12*float32(a.Orientation+1),
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
		rl.DrawText(fmt.Sprintf("State: %d", a.State), int32(a.Pos.X), int32(a.Pos.Y+a.CollisionBox.Y+a.CollisionBox.Height), 2, rl.White)
	}

}

func (a *Actor) SetAnimFrame(anim string, offset uint8) {
	a.AnimPhysTimeStart = PhysicsTime
	a.Anim = anim
	a.AnimDuration = 1
	a.AnimTotalFrames = 1
	a.AnimFinished = false
	a.AnimOnce = false
	a.AnimOffset = offset
}

func (a *Actor) SetAnim(anim string, animDuration float64, totalFrames uint8, once bool) {
	a.AnimPhysTimeStart = PhysicsTime
	a.Anim = anim
	a.AnimFinished = false
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

	frameTimeMs := a.AnimDuration / float64(a.AnimTotalFrames)
	diff := PhysicsTime - a.AnimPhysTimeStart
	frame := int(diff / frameTimeMs)
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
	frameTimeMs := a.AnimDuration / float64(a.AnimTotalFrames)
	if PhysicsTime-a.AnimPhysTimeStart > a.AnimDuration {
		a.AnimPhysTimeStart += a.AnimDuration
	}
	frame := int((PhysicsTime - a.AnimPhysTimeStart) / frameTimeMs)
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

func (a *Actor) updateUpdown() {
	if a.State != StateIdle || !a.AnimFinished && strings.Contains(a.Anim, "updown-") {
		return
	}
	var updown uint8
	if a.IsPlayer {
		switch {
		case float32(MouseWorldPos.Y) <= a.Pos.Y+14:
			updown = 2
		case float32(MouseWorldPos.Y) <= a.Pos.Y+32:
			updown = 1
		default:
			updown = 0
		}
	} else {
		updown = a.Updown
		if rand.Int()%61 == 0 {
			updown = uint8(rand.Int() % 3)
		}
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
	a.SetAnim("resources/threefour/hit-top.png", 0.35, 4, true)
}

func (a *Actor) enterBlockState() {
	a.State = StateBlock
	switch a.Updown {
	case 0:
		a.SetAnim("resources/threefour/block-low-start.png", 0.1, 5, true)
	case 1, 2:
		a.SetAnim("resources/threefour/block-mid-start.png", 0.1, 5, true)
	}
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
	if a.Input.Attack {
		a.AttackQueued = true
	}
	var r1 rl.Rectangle
	switch a.Updown {
	case 2:
		r1 = rl.Rectangle{
			X:      a.Pos.X + a.AttackBoxHigh.X + 12*float32(a.Orientation+1),
			Y:      a.Pos.Y + a.AttackBoxHigh.Y,
			Width:  a.AttackBoxHigh.Width,
			Height: a.AttackBoxHigh.Height,
		}
	case 1:
		r1 = rl.Rectangle{
			X:      a.Pos.X + a.AttackBoxMid.X + 12*float32(a.Orientation+1),
			Y:      a.Pos.Y + a.AttackBoxMid.Y,
			Width:  a.AttackBoxMid.Width,
			Height: a.AttackBoxMid.Height,
		}
	case 0:
		r1 = rl.Rectangle{
			X:      a.Pos.X + a.AttackBoxLow.X + 12*float32(a.Orientation+1),
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
			X:      actor.Pos.X + actor.CollisionBox.X,
			Y:      actor.Pos.Y + actor.CollisionBox.Y,
			Width:  actor.CollisionBox.Width,
			Height: actor.CollisionBox.Height,
		}
		if aabb(r1, r2) && actor.State != StateAttacked && actor.State != StateBlock {
			actor.Attacker = a
			if a.Updown == actor.Updown {
				actor.enterBlockState()
			} else if a.AnimFrame > 4 {
				actor.enterAttackedState()
			}
		}
	}

	if a.AnimFinished {
		if strings.Contains(a.Anim, "back") {
			a.enterIdleState()
			return
		}

		if !a.AttackQueued || strings.Contains(a.Anim, "mid-2") {
			switch a.Updown {
			case 0:
				a.SetAnim("resources/threefour/back-attack-low-1.png", 0.2, 4, true)
			case 1:
				a.SetAnim("resources/threefour/back-attack-mid-1.png", 0.2, 4, true)
			case 2:
				a.SetAnim("resources/threefour/back-attack-high-1.png", 0.2, 4, true)
			}
			return
		}

		a.AttackQueued = false

		switch a.Updown {
		case 0:
			a.SetAnim("resources/threefour/attack-mid-2.png", 0.2, 6, true)
		case 1:
			a.SetAnim("resources/threefour/attack-mid-2.png", 0.2, 6, true)
		case 2:
			a.SetAnim("resources/threefour/attack-mid-2.png", 0.2, 6, true)
		}
	}
}

func (a *Actor) updateBlock() {
	if a.Attacker.State != StateAttack && a.AnimFinished {
		a.enterIdleState()
		return
	}
	if strings.HasSuffix(a.Anim, "start.png") && a.AnimFinished && a.Attacker.AnimFrame > 4 {
		switch a.Updown {
		case 0:
			a.SetAnim("resources/threefour/block-low-1.png", 0.2, 3, true)
		case 1, 2:
			a.SetAnim("resources/threefour/block-mid-1.png", 0.2, 3, true)
		}
	} else if a.AnimFinished && a.Attacker.AnimFrame > 4 {
		switch a.Updown {
		case 0:
			a.SetAnim("resources/threefour/block-low-1.png", 0.2, 3, true)
		case 1, 2:
			a.SetAnim("resources/threefour/block-mid-2.png", 0.2, 3, true)
		}
	}
}
