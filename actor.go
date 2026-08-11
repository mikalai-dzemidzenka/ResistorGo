package resistor

import (
	"fmt"
	rl "github.com/gen2brain/raylib-go/raylib"
	"math/rand"
	"strings"
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
	StateIdle ActorState = iota
	StateRun
	StateAttack
)

type Actor struct {
	Pos             rl.Vector2
	Speed           rl.Vector2
	Acc             float32
	Friction        float32
	Orientation     int8
	State           ActorState
	RewindData      [RewindFrames]ActorRewindFrame
	CollisionBox    rl.Rectangle
	OnGround        bool
	IsPlayer        bool
	Input           Input
	PrevUpdown      uint8
	Updown          uint8
	AnimTime        float32
	Anim            string
	AnimFPS         uint8
	AnimOffset      uint8
	AnimTotalFrames uint8
	AnimFrame       uint8
	AnimOnce        bool
	AnimFinished    bool
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
		IsPlayer:        isPlayer,
		Anim:            "resources/threefour/updown.png",
		AnimTotalFrames: 1,
		AnimFPS:         1,
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
	a.updateHorSpeed(dt)
	a.updateVertSpeed(dt)
	a.updateUpdown(dt)
	a.checkCollision()
	a.updateAttack()

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

func (a *Actor) Draw(dt float32) {
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
			a.AnimFinished = a.DrawAnimOnce(dt)
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
	a.AnimTime = 0
	a.Anim = anim
	a.AnimFPS = 1
	a.AnimTotalFrames = 1
	a.AnimOnce = false
	a.AnimOffset = offset
}

func (a *Actor) SetAnim(anim string, fps uint8, totalFrames uint8, once bool) {
	a.AnimTime = 0
	a.Anim = anim
	a.AnimFPS = fps
	a.AnimTotalFrames = totalFrames
	a.AnimOnce = once
	a.AnimOffset = 0
}

func (a *Actor) DrawAnimOnce(dt float32) bool {
	frameTimeMs := 1000 / float32(a.AnimFPS)
	a.AnimTime += dt
	a.AnimTime = clamp(a.AnimTime, 0, frameTimeMs*float32(a.AnimTotalFrames))
	frame := int(a.AnimTime*1000/frameTimeMs) / int(a.AnimTotalFrames)
	finished := frame >= int(a.AnimTotalFrames)
	frame = clamp(frame, 0, int(a.AnimTotalFrames)-1)
	tex, ok := resources[a.Anim]
	if !ok {
		panic(a.Anim)
	}
	a.AnimFrame = uint8(frame)
	rl.DrawTextureRec(tex, rl.Rectangle{
		X:      float32(frame) * 48,
		Y:      0,
		Width:  48 * float32(a.Orientation),
		Height: 48,
	}, a.Pos, rl.White)
	return finished
}

func (a *Actor) DrawAnim() {
	frameTimeMs := 1000 / float32(a.AnimFPS)
	tex, ok := resources[a.Anim]
	if !ok {
		panic(a.Anim)
	}
	frame := float32(a.AnimOffset) + float32(int(GameTimeSec*1000/frameTimeMs)%int(a.AnimTotalFrames+1))
	a.AnimFrame = uint8(frame)
	rl.DrawTextureRec(tex, rl.Rectangle{
		X:      frame * 48,
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
		a.State = StateIdle
		a.Speed.X = 0
		if !strings.Contains(a.Anim, "updown") {
			a.SetAnimFrame("resources/threefour/updown.png", a.Updown)
		}
	} else {
		a.State = StateRun
		var fps uint8 = 24
		if absSpeed < MaxSpeed {
			fps = 12
		}
		a.SetAnim("resources/threefour/run.png", fps, 8, false)
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
		a.SetAnim(anim, 8, 4, true)
	} else if a.AnimFinished && strings.Contains(a.Anim, "updown-") {
		a.SetAnimFrame("resources/threefour/updown.png", a.Updown)
	}
	a.Updown = updown
}

func (a *Actor) updateAttack() {
	if a.Input.Attack && a.State != StateAttack {
		a.State = StateAttack
		switch a.Updown {
		case 0:
			a.SetAnim("resources/threefour/attack-low.png", 4, 8, true)
		case 1:
			a.SetAnim("resources/threefour/attack-mid.png", 4, 8, true)
		case 2:
			a.SetAnim("resources/threefour/attack-high.png", 4, 8, true)
		}
		return
	}
	if a.State != StateAttack {
		return
	}

	if a.AnimFinished {
		a.State = StateIdle
		return
	}
}
