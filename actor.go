package resistor

import (
	"fmt"
	rl "github.com/gen2brain/raylib-go/raylib"
	"math"
	"math/rand"
	"strings"
)

const (
	DefaultAcc      = 270
	DefaultFriction = 250
	RunReach        = 200
	MaxSpeed        = 320
	PunchSpeed      = 50
	AttackDash      = 4
	Gravity         = 1500
	MaxJumpSpeed    = -Gravity * 10
	MaxFallSpeed    = 900
	AttackZoomSpeed = 2.0

	BackAttackSpeed   = 0.2
	SecondAttackSpeed = 0.2
	BlockSpeed        = 0.2

	HitStopTicks = 6
)

type ActorState uint8

var attackFrames = map[string]uint8{
	"resources/threefour/attack-high.png":    5,
	"resources/threefour/attack-high-2.png":  3,
	"resources/threefour/attack-high-3.png":  3,
	"resources/threefour/attack-mid.png":     5,
	"resources/threefour/attack-mid-2.png":   3,
	"resources/threefour/attack-mid-3.png":   3,
	"resources/threefour/attack-low.png":     5,
	"resources/threefour/attack-low-2.png":   3,
	"resources/threefour/attack-low-3.png":   3,
	"resources/threefour/updown-atk-0-1.png": 4,
	"resources/threefour/updown-atk-0-2.png": 4,
	"resources/threefour/updown-atk-1-2.png": 4,
	"resources/threefour/updown-atk-1-0.png": 4,
	"resources/threefour/updown-atk-2-1.png": 4,
	"resources/threefour/updown-atk-2-0.png": 4,
}

const (
	StateIdle ActorState = iota
	StateRun
	StateAttack
	StateAttacked
	StateBlock
	StateJumpStart
	StateJump
	StateClimb
	StateTurnover
)

const (
	AttackFrame = 4
)

// presentation events, consumed (and cleared) by the renderer
const (
	EvAttackStart uint8 = 1 << iota
	EvHit
	EvBlocked
)

type Actor struct {
	ID               uint16
	W                *World
	Remote           bool // driven by snapshot interpolation, not simulated
	Pos              rl.Vector2
	PrevPos          rl.Vector2
	InterpolatedPos  rl.Vector2
	MoveToPos        rl.Vector2
	Speed            rl.Vector2
	Energy           int32
	Acc              float32
	Friction         float32
	Orientation      int8
	State            ActorState
	RewindData       [RewindFrames]ActorRewindFrame
	CollisionBox     rl.Rectangle
	AttackBoxHigh    rl.Rectangle
	AttackBoxMid     rl.Rectangle
	AttackBoxLow     rl.Rectangle
	OnGround         bool
	IsPlayer         bool
	Input            Input
	PrevUpdown       uint8
	Updown           uint8
	Ticks            uint32
	FreezeTicks      uint8
	Events           uint8
	AnimStartTick    uint32
	AnimTop          *string
	Anim             string
	AnimDuration     float64
	AnimYOffset      int8
	AnimFrame        uint8
	AnimTotalFrames  uint8
	AnimCurrentFrame uint8
	AnimOnce         bool
	AnimReverse      bool
	AnimFinished     bool
	AnimWhite        bool
	Attacker         *Actor
	AttackQueued     bool
	AttackCount      uint8
	CanClimbRight    bool
	CanClimbLeft     bool
}

// Input is one simulation tick of player intent; it is also the network command.
type Input struct {
	Seq      uint32
	Left     bool
	Right    bool
	Attack   bool
	Jump     bool
	FaceLeft bool
	Updown   uint8
	ViewTick uint32
	ViewFrac uint8
}

func NewActor(pos rl.Vector2, isPlayer bool) *Actor {
	return &Actor{
		Pos:         pos,
		Speed:       rl.Vector2{},
		Acc:         DefaultAcc,
		Friction:    DefaultFriction,
		State:       StateIdle,
		Orientation: 1,
		OnGround:    false,
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

// Update advances the actor by one fixed tick using a.Input.
func (a *Actor) Update(dt float32) {
	a.PrevPos = a.Pos
	if a.FreezeTicks > 0 {
		a.FreezeTicks--
		return
	}

	prevOrientation := a.Orientation
	a.Ticks++
	defer a.advanceAnim()

	a.checkCollision()
	a.updateHorSpeed(dt)
	a.updateVertSpeed(dt)

	switch a.State {
	case StateClimb:
		a.Speed = rl.Vector2{}
		if a.AnimFinished {
			a.enterIdleState()
		}
	case StateIdle:
		prevUpdown := a.Updown
		a.updateUpdown()
		if prevUpdown == 3 && a.Updown != 3 {
			a.SetAnimFrame("resources/threefour/updown.png", a.Updown)
		} else if a.Updown == 3 {
			a.SetAnimFrame("resources/threefour/block.png", 0)
		} else if absSpeed := abs(a.Speed.X); absSpeed > RunReach {
			a.enterRunState()
			return
		} else if absSpeed > 0.1 {
			reverse := true
			if a.Speed.X > 0 && !a.MouseOnLeft() || a.Speed.X < 0 && a.MouseOnLeft() {
				reverse = false
			}
			reset := a.AnimReverse != reverse
			a.SetAnim("resources/threefour/walk.png", AnimOptions{reset: reset, animDuration: 0.6, totalFrames: 9, reverse: reverse, startFrame: 9})
		} else if prevUpdown != a.Updown {
			anim := fmt.Sprintf("resources/threefour/updown-%d-%d.png", prevUpdown, a.Updown)
			a.SetAnim(anim, AnimOptions{animDuration: 0.1, totalFrames: 4, once: true})
		} else if a.AnimFinished && strings.Contains(a.Anim, "updown-") {
			a.SetAnimFrame("resources/threefour/updown.png", a.Updown)
		} else if absSpeed < 0.1 && a.Anim == "resources/threefour/walk.png" {
			a.SetAnimFrame("resources/threefour/updown.png", a.Updown)
		}
		if a.Input.Attack && a.Updown != 3 {
			a.enterAttackState()
			return
		}
		if a.Input.Jump {
			a.enterJumpStartState()
			return
		}
		if !a.OnGround {
			// TODO state fall
		}
	case StateRun:
		if prevOrientation != a.Orientation {
			a.enterTurnoverState()
		}
		absSpeed := abs(a.Speed.X)
		if absSpeed < 0.001 {
			a.enterIdleState()
			return
		}
		if a.Input.Attack {
			a.enterAttackState()
		}
	case StateAttack:
		a.Speed.X = clamp(a.Speed.X, -AttackDash, AttackDash)
		a.updateAttack()
	case StateAttacked:
		a.Speed.X = clamp(a.Speed.X, -AttackDash, AttackDash)
		a.updateAttacked()
	case StateBlock:
		a.updateBlock()
	case StateJumpStart:
		if a.AnimFinished {
			a.enterJumpState(dt)
		}
	case StateJump:
		if (a.CanClimbRight && a.Orientation == 1) || (a.CanClimbLeft && a.Orientation == -1) {
			a.exitJumpState()
			a.enterClimbState()
		} else if a.OnGround {
			a.exitJumpState()
			a.enterIdleState()
		}
	case StateTurnover:
		if a.AnimFinished {
			if abs(a.Speed.X) < 0.001 {
				a.enterIdleState()
			} else if a.Input.Left || a.Input.Right {
				a.Speed.X = RunReach * float32(a.Orientation)
				a.enterRunState()
			} else {
				a.enterIdleState()
			}
		}
	}

	a.Pos.X += a.Speed.X * dt
	a.Pos.Y += a.Speed.Y * dt
	a.Pos.X = float32(int(a.Pos.X))
	a.Pos.Y = float32(int(a.Pos.Y))
}

// advanceAnim derives the current animation frame from the actor's own tick counter.
func (a *Actor) advanceAnim() {
	elapsed := float64(a.Ticks-a.AnimStartTick) * FixedTimeStep
	a.AnimCurrentFrame, a.AnimFinished = animFrame(a.AnimOnce, a.AnimReverse, a.AnimFrame, a.AnimTotalFrames, a.AnimDuration, elapsed)
}

func animFrame(once, reverse bool, start, total uint8, duration, elapsed float64) (uint8, bool) {
	if total == 0 || duration <= 0 {
		return start, once
	}
	frameTime := duration / float64(total)
	if once {
		frame := int(elapsed / frameTime)
		finished := frame >= int(total)
		return uint8(clamp(frame, 0, int(total)-1)), finished
	}
	elapsed = math.Mod(max(elapsed, 0), duration)
	frame := uint8(clamp(int(elapsed/frameTime), 0, int(total)-1))
	if reverse {
		return start - frame, false
	}
	return start + frame, false
}

// recordRewind writes the actor's state into the frame the recording is open
// on. Called once per live fixed step, before the step runs.
func (a *Actor) recordRewind() {
	a.RewindData[rewindSlot(rewindHead)] = ActorRewindFrame{
		Pos:         a.Pos,
		Speed:       a.Speed,
		Orientation: a.Orientation,
		Anim:        a.Anim,
		AnimFrame:   a.AnimCurrentFrame,
	}
}

// restoreRewind puts the actor back onto the frame the playhead is on. The
// previous position goes with it: nothing is in motion on a recorded frame, so
// there is nothing to interpolate from.
func (a *Actor) restoreRewind() {
	data := a.RewindData[rewindSlot(rewindCursor)]
	a.Pos = data.Pos
	a.PrevPos = data.Pos
	a.Speed = data.Speed
	a.Orientation = data.Orientation
}
func (a *Actor) UpdatePhysics() {}

// drawPos is the sprite's position rounded to whole world pixels. Interpolation
// and the client's error correction leave sub-pixel positions, which the
// integer camera zoom turns into texel rows of uneven width - a sprite drawn at
// x.5 shimmers as it moves.
func (a *Actor) drawPos() rl.Vector2 {
	return rl.Vector2{X: roundf(a.InterpolatedPos.X), Y: roundf(a.InterpolatedPos.Y)}
}

func (a *Actor) Draw() {
	if TimeShowsHistory() {
		data := a.RewindData[rewindSlot(rewindCursor)]
		rl.DrawTextureRec(resources[data.Anim], rl.Rectangle{
			X:      float32(data.AnimFrame) * 48,
			Y:      0,
			Width:  48 * float32(a.Orientation),
			Height: 48,
		}, a.drawPos(), rl.White)
	} else {
		if a.AnimOnce {
			a.DrawAnimOnce()
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
		rl.DrawRectangleLinesEx(a.collisionRect(), 1, rl.SkyBlue)
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
			rl.DrawText(fmt.Sprintf("Time: %s", timeModeName()), int32(a.Pos.X), int32(a.Pos.Y), 2, rl.White)
			rl.DrawText(fmt.Sprintf("Frame: %d of %d", rewindCursor-rewindOldest()+1, rewindCount), int32(a.Pos.X), int32(a.Pos.Y)-8, 2, rl.White)
			rl.DrawText(fmt.Sprintf("X, Y: %d %d", int32(a.Pos.X), int32(a.Pos.Y)), int32(a.Pos.X), int32(a.Pos.Y)-16, 2, rl.White)
			rl.DrawText(fmt.Sprintf("PrevUpdown, Updown: %d %d", a.PrevUpdown, a.Updown), int32(a.Pos.X), int32(a.Pos.Y)-24, 2, rl.White)
			rl.DrawText(fmt.Sprintf("Frame, Finished: %d %t", a.AnimCurrentFrame, a.AnimFinished), int32(a.Pos.X), int32(a.Pos.Y)-32, 2, rl.White)
		}
		px := a.Pos.X + a.CollisionBox.X + a.CollisionBox.Width/2
		rl.DrawPixel(int32(px), int32(a.Pos.Y+a.CollisionBox.Y+a.CollisionBox.Height), rl.Red)
		rl.DrawText(fmt.Sprintf("State: %d", a.State), int32(a.Pos.X), int32(a.Pos.Y+a.CollisionBox.Y+a.CollisionBox.Height), 2, rl.White)
	}

}

func (a *Actor) collisionRect() rl.Rectangle {
	return rl.Rectangle{
		X:      a.Pos.X + a.CollisionBox.X,
		Y:      a.Pos.Y + a.CollisionBox.Y,
		Width:  a.CollisionBox.Width,
		Height: a.CollisionBox.Height,
	}
}

func (a *Actor) SetAnimFrame(anim string, offset uint8) {
	a.AnimStartTick = a.Ticks
	a.Anim = anim
	a.AnimDuration = 1
	a.AnimTotalFrames = 1
	a.AnimFinished = false
	a.AnimOnce = false
	a.AnimFrame = offset
}

type AnimOptions struct {
	animDuration float64
	totalFrames  uint8
	animYOffset  int8
	startFrame   uint8
	once         bool
	reset        bool
	reverse      bool
}

func (a *Actor) SetAnim(anim string, o AnimOptions) {
	if a.Anim == anim && !o.reset {
		return
	}
	a.AnimStartTick = a.Ticks
	a.Anim = anim
	a.AnimFinished = false
	a.AnimDuration = o.animDuration
	a.AnimTotalFrames = o.totalFrames
	a.AnimOnce = o.once
	a.AnimReverse = o.reverse
	a.AnimFrame = o.startFrame
	a.AnimYOffset = o.animYOffset
}

func (a *Actor) DrawFrame() {
	tex, ok := resources[a.Anim]
	if !ok {
		panic(a.Anim)
	}

	rl.DrawTextureRec(tex, rl.Rectangle{
		X:      float32(a.AnimCurrentFrame) * 48,
		Y:      0,
		Width:  48 * float32(a.Orientation),
		Height: float32(tex.Height),
	}, rl.Vector2{
		X: a.drawPos().X,
		Y: a.drawPos().Y + float32(a.AnimYOffset),
	}, rl.White)
}

func (a *Actor) DrawAnimOnce() {
	tex, ok := resources[a.Anim]
	if !ok {
		panic(a.Anim)
	}

	rl.DrawTextureRec(tex, rl.Rectangle{
		X:      float32(a.AnimCurrentFrame) * 48,
		Y:      0,
		Width:  48 * float32(a.Orientation),
		Height: float32(tex.Height),
	}, rl.Vector2{
		X: a.drawPos().X,
		Y: a.drawPos().Y + float32(a.AnimYOffset),
	}, rl.White)
}

func (a *Actor) DrawAnim() {
	tex, ok := resources[a.Anim]
	if !ok {
		panic(a.Anim)
	}

	rl.DrawTextureRec(tex, rl.Rectangle{
		X:      float32(a.AnimCurrentFrame) * 48,
		Y:      0,
		Width:  48 * float32(a.Orientation),
		Height: 48,
	}, a.drawPos(), rl.White)

	// TODO
	if strings.HasSuffix(a.Anim, "walk.png") {
		switch a.Updown {
		case 0:
			rl.DrawTextureRec(resources["resources/threefour/walk-bot.png"], rl.Rectangle{
				X:      float32(a.AnimCurrentFrame) * 48,
				Y:      0,
				Width:  48 * float32(a.Orientation),
				Height: float32(tex.Height),
			}, rl.Vector2{
				X: a.drawPos().X,
				Y: a.drawPos().Y + float32(a.AnimYOffset),
			}, rl.White)
		case 1:
			rl.DrawTextureRec(resources["resources/threefour/walk-mid.png"], rl.Rectangle{
				X:      float32(a.AnimCurrentFrame) * 48,
				Y:      0,
				Width:  48 * float32(a.Orientation),
				Height: float32(tex.Height),
			}, rl.Vector2{
				X: a.drawPos().X,
				Y: a.drawPos().Y + float32(a.AnimYOffset),
			}, rl.White)
		case 2:
			rl.DrawTextureRec(resources["resources/threefour/walk-top.png"], rl.Rectangle{
				X:      float32(a.AnimCurrentFrame) * 48,
				Y:      0,
				Width:  48 * float32(a.Orientation),
				Height: float32(tex.Height),
			}, rl.Vector2{
				X: a.drawPos().X,
				Y: a.drawPos().Y + float32(a.AnimYOffset),
			}, rl.White)
		}
	}
}

func (a *Actor) updateHorSpeed(dt float32) {
	var movePressed bool
	prevOrientation := a.Orientation
	canMove := a.State != StateAttack && a.State != StateBlock && a.State != StateAttacked && a.Updown != 3
	if a.Input.Right && !(a.State == StateTurnover && prevOrientation == 1) && canMove {
		movePressed = true
		if abs(a.Speed.X) < 0.001 {
			a.Speed.X -= 12 * a.Acc * dt
		} else {
			a.Speed.X -= a.Acc * dt
		}
		a.Orientation = -1
	}
	if a.Input.Left && !(a.State == StateTurnover && prevOrientation == -1) && canMove {
		movePressed = true
		if abs(a.Speed.X) < 0.001 {
			a.Speed.X += 12 * a.Acc * dt
		} else {
			a.Speed.X += a.Acc * dt
		}
		a.Orientation = 1
	}
	if a.State == StateIdle || a.State == StateAttack && a.AnimCurrentFrame == 0 {
		if a.IsPlayer {
			if a.MouseOnLeft() {
				a.Orientation = -1
			} else {
				a.Orientation = 1
			}
		}
	}

	if prevOrientation != a.Orientation {
		a.Speed.X /= 3
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

// headBlocked reports whether an object currently contains the actor's head.
// The climb checks in checkCollision need this up front: objectsIndexed is
// sorted by descending X, so a neighbouring tile of a ceiling row is visited
// before the tile the head is actually inside of.
func (a *Actor) headBlocked(pColCenterX float32) bool {
	pColUpY := a.Pos.Y + a.CollisionBox.Y
	for _, obj := range LE.objectsIndexed {
		if obj.CollisionRect == nil {
			continue
		}
		oColY := float32(obj.Y + int32(obj.CollisionRect.Y))
		oColX := float32(obj.X + int32(obj.CollisionRect.X))
		if oColX <= pColCenterX && oColX+obj.CollisionRect.Width >= pColCenterX &&
			pColUpY >= oColY && pColUpY <= oColY+obj.CollisionRect.Height {
			return true
		}
	}
	return false
}

func (a *Actor) checkCollision() {
	a.OnGround = false
	pColDownY := a.Pos.Y + a.CollisionBox.Y + a.CollisionBox.Height
	pColCenterX := a.Pos.X + a.CollisionBox.X + a.CollisionBox.Width/2
	pPosX := a.Pos.X
	hitTop := a.headBlocked(pColCenterX)
	for _, obj := range LE.objectsIndexed {
		oColY := float32(obj.Y + int32(obj.CollisionRect.Y))
		oColHeight := obj.CollisionRect.Height
		oColX := float32(obj.X + int32(obj.CollisionRect.X))
		oColWidth := obj.CollisionRect.Width

		if pColCenterX > oColX && pColCenterX < oColX+oColWidth &&
			pColDownY+1 > oColY && pColDownY < oColY+oColHeight {
			a.OnGround = true
			a.Speed.Y = 0
			a.Pos.Y = float32(int32(math.Ceil(float64(oColY)) - 48))
			continue
		}
		pColUpY := a.Pos.Y + a.CollisionBox.Y
		jump := a.Speed.Y < 0
		const climbOffset = 32
		if oColX <= pColCenterX && oColX+oColWidth >= pColCenterX &&
			pColUpY >= oColY && pColUpY <= oColY+oColHeight {
			hitTop = true
			a.CanClimbRight = false
			a.CanClimbLeft = false
			a.Speed.Y = 0
			a.Pos.X = pPosX
			a.Pos.Y = float32(int32(oColY + oColHeight))
			continue
		} else if pColCenterX+climbOffset >= oColX && oColX >= pColCenterX &&
			pColUpY >= oColY && pColUpY <= oColY+oColHeight && a.Orientation == 1 && jump && !hitTop {
			a.CanClimbRight = true
			a.Pos.X = float32(int32(oColX) - 20)
			a.Pos.Y = float32(int32(math.Ceil(float64(oColY)) - 48 + 1))
			continue
		} else if pColCenterX >= oColX+oColWidth && oColX+oColWidth >= pColCenterX-climbOffset &&
			pColUpY >= oColY && pColUpY <= oColY+oColHeight && a.Orientation == -1 && jump && !hitTop {
			a.CanClimbLeft = true
			a.Pos.X = oColX + oColWidth - 24
			a.Pos.Y = float32(int32(math.Ceil(float64(oColY)) - 48 + 1))
			continue
		}
	}
}

// nearest returns the closest other actor in the same world.
func (a *Actor) nearest() *Actor {
	var best *Actor
	bestD := float32(math.MaxFloat32)
	for _, b := range a.W.Actors {
		if b == a {
			continue
		}
		if d := abs(b.Pos.X-a.Pos.X) + abs(b.Pos.Y-a.Pos.Y); d < bestD {
			best, bestD = b, d
		}
	}
	return best
}

// botInput produces the bot's command for this tick (authoritative worlds only).
func botInput(a *Actor) Input {
	if a.State == StateAttacked {
		return a.Input
	}
	const near = 50
	in := Input{Updown: a.Updown}
	if rand.Int()%61 == 0 {
		in.Updown = uint8(rand.Int() % 3)
	}
	t := a.nearest()
	if t == nil {
		return in
	}
	if t.Pos.X < a.Pos.X {
		if a.Pos.X-t.Pos.X > near {
			in.Right = true
		} else {
			if rand.Int()%61 == 3 {
				in.Attack = true
			}
		}
		a.Orientation = -1
	}
	if t.Pos.X > a.Pos.X {
		if t.Pos.X-a.Pos.X > near {
			in.Left = true
		} else {
			if rand.Int()%61 == 3 {
				in.Attack = true
			}
		}
		a.Orientation = 1
	}
	return in
}

// edge-triggered buttons are latched per rendered frame and consumed by the next tick
var pressedAttack, pressedJump bool

func latchInput() {
	if rl.IsMouseButtonPressed(rl.MouseButtonLeft) {
		pressedAttack = true
	}
	if rl.IsKeyPressed(rl.KeyW) {
		pressedJump = true
	}
}

// sampleInput reads the local player's devices into a command for actor a.
func sampleInput(a *Actor) Input {
	in := Input{
		Right:  rl.IsKeyDown(rl.KeyA),
		Left:   rl.IsKeyDown(rl.KeyD),
		Attack: pressedAttack,
		Jump:   pressedJump,
	}
	pressedAttack, pressedJump = false, false

	curPos := rl.GetScreenToWorld2D(Cur.AtkPos, Cam.View())
	switch {
	case float32(curPos.Y) <= a.Pos.Y+a.CollisionBox.Y:
		in.Updown = 3
	case float32(curPos.Y) <= a.Pos.Y+16:
		in.Updown = 2
	case float32(curPos.Y) <= a.Pos.Y+24:
		in.Updown = 1
	default:
		in.Updown = 0
	}
	in.FaceLeft = float32(MouseWorldPos.X) < a.Pos.X+24
	return in
}

func (a *Actor) updateUpdown() {
	prevUpdown := a.Updown
	a.Updown = a.Input.Updown
	if a.Updown != prevUpdown {
		a.PrevUpdown = prevUpdown
	}
}

func (a *Actor) enterAttackState() {
	a.State = StateAttack
	a.Events |= EvAttackStart
	a.Pos.X += float32(a.Orientation) * AttackDash
	switch a.Updown {
	case 0:
		a.SetAnim("resources/threefour/attack-low.png", AnimOptions{animDuration: 0.3, totalFrames: 8, once: true})
	case 1:
		a.SetAnim("resources/threefour/attack-mid.png", AnimOptions{animDuration: 0.3, totalFrames: 8, once: true})
	case 2, 3:
		a.SetAnim("resources/threefour/attack-high.png", AnimOptions{animDuration: 0.3, totalFrames: 8, once: true})
	}
}

func (a *Actor) enterAttackedState(attacker *Actor) {
	a.State = StateAttacked
	a.Attacker = attacker
	a.AnimWhite = true
	a.Events |= EvHit
	a.FreezeTicks = HitStopTicks
	attacker.FreezeTicks = HitStopTicks
	a.Speed.X = float32(attacker.Orientation) * PunchSpeed
	a.SetAnim("resources/threefour/hit-top.png", AnimOptions{animDuration: 0.3, totalFrames: 4, once: true})
}

func (a *Actor) enterBlockState(attacker *Actor) {
	a.State = StateBlock
	a.Attacker = attacker
	a.Events |= EvBlocked
	a.Speed.X = float32(attacker.Orientation) * PunchSpeed
	switch a.Updown {
	case 0:
		a.SetAnim("resources/threefour/block-low-start.png", AnimOptions{animDuration: 0.1, totalFrames: 5, once: true})
	case 1:
		a.SetAnim("resources/threefour/block-mid-start.png", AnimOptions{animDuration: 0.1, totalFrames: 5, once: true})
	case 2:
		a.SetAnim("resources/threefour/block-mid-v2-start.png", AnimOptions{animDuration: 0.1, totalFrames: 5, once: true})
	case 3:
		a.SetAnim("resources/threefour/block-all-start.png", AnimOptions{animDuration: 0.1, totalFrames: 5, once: true})
	}
}

func (a *Actor) enterIdleState() {
	a.State = StateIdle
	a.Speed.X = 0
	a.Speed.Y = 0
	if !strings.Contains(a.Anim, "updown") {
		a.SetAnimFrame("resources/threefour/updown.png", a.Updown)
	}
}

func (a *Actor) enterRunState() {
	a.State = StateRun
	a.SetAnim("resources/threefour/run.png", AnimOptions{animDuration: 0.4, totalFrames: 10})
}

func (a *Actor) updateAttacked() {
	if a.AnimFinished {
		a.exitAttackedState()
		a.enterIdleState()
	}
}

func (a *Actor) updateAttack() {
	if a.Input.Attack {
		a.Events |= EvAttackStart
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

	// only the authoritative world decides hits and blocks
	if a.W.Authority {
		for _, actor := range a.W.Actors {
			if a == actor {
				continue
			}
			r2 := actor.collisionRect()
			if aabb(r1, r2) && actor.State != StateAttacked && actor.State != StateBlock {
				if a.Updown == actor.Updown || actor.Updown == 3 {
					actor.enterBlockState(a)
				} else if a.AnimCurrentFrame > attackFrames[a.Anim] {
					actor.enterAttackedState(a)
				}
			}
		}
	}

	if a.AnimFinished {
		if strings.Contains(a.Anim, "back") {
			a.exitAttackState()
			a.enterIdleState()
			return
		}

		if !a.AttackQueued {
			switch a.Updown {
			case 0:
				if strings.HasSuffix(a.Anim, "attack-low-2.png") {
					a.SetAnim("resources/threefour/back-attack-low-2.png", AnimOptions{animDuration: BackAttackSpeed, totalFrames: 4, once: true})
				} else {
					a.SetAnim("resources/threefour/back-attack-low-1.png", AnimOptions{animDuration: BackAttackSpeed, totalFrames: 4, once: true})
				}
			case 1:
				a.SetAnim("resources/threefour/back-attack-mid-1.png", AnimOptions{animDuration: BackAttackSpeed, totalFrames: 4, once: true})
			case 2, 3:
				if strings.HasSuffix(a.Anim, "attack-high-2.png") {
					a.SetAnim("resources/threefour/back-attack-high-2.png", AnimOptions{animDuration: BackAttackSpeed, totalFrames: 4, once: true})
				} else {
					a.SetAnim("resources/threefour/back-attack-high-1.png", AnimOptions{animDuration: BackAttackSpeed, totalFrames: 4, once: true})
				}
			}
			return
		}

		a.AttackQueued = false
		a.AttackCount++
		a.Speed.X = float32(a.Orientation) * AttackDash

		prevUpdown := a.Updown
		a.updateUpdown()
		if prevUpdown == 3 || a.Updown == 3 {
			a.enterIdleState()
			return
		}
		if a.Updown != prevUpdown {
			anim := fmt.Sprintf("resources/threefour/updown-atk-%d-%d.png", prevUpdown, a.Updown)
			a.SetAnim(anim, AnimOptions{animDuration: 0.4, totalFrames: 7, once: true})
			return
		}

		switch a.Updown {
		case 0:
			if a.Anim == "resources/threefour/attack-low.png" {
				a.SetAnim("resources/threefour/attack-low-2.png", AnimOptions{animDuration: SecondAttackSpeed, totalFrames: 6, once: true})
			} else if a.Anim == "resources/threefour/attack-low-2.png" {
				a.SetAnim("resources/threefour/attack-low-2.png", AnimOptions{animDuration: SecondAttackSpeed, totalFrames: 6, once: true})
			}
		case 1:
			a.SetAnim("resources/threefour/attack-mid-2.png", AnimOptions{animDuration: SecondAttackSpeed, totalFrames: 6, once: true})
		case 2:
			if a.Anim == "resources/threefour/attack-high.png" {
				a.SetAnim("resources/threefour/attack-high-2.png", AnimOptions{animDuration: SecondAttackSpeed, totalFrames: 6, once: true})
			} else if a.Anim == "resources/threefour/attack-high-2.png" {
				a.SetAnim("resources/threefour/attack-high-3.png", AnimOptions{animDuration: SecondAttackSpeed, totalFrames: 6, once: true})
			} else if a.Anim == "resources/threefour/attack-high-3.png" {
				a.SetAnim("resources/threefour/attack-high-2.png", AnimOptions{animDuration: SecondAttackSpeed, totalFrames: 6, once: true, startFrame: 3})
			}
		}
	}
}

func (a *Actor) updateBlock() {
	if a.Attacker == nil {
		a.exitBlockState()
		a.enterIdleState()
		return
	}
	if a.Attacker.State != StateAttack && a.AnimFinished {
		a.exitBlockState()
		a.enterIdleState()
		return
	}
	if a.AnimFinished && a.Attacker.AnimCurrentFrame >= attackFrames[a.Attacker.Anim] {
		if a.W.Authority && a.Updown != a.Attacker.Updown && a.Updown != 3 {
			attacker := a.Attacker
			a.exitBlockState()
			a.enterAttackedState(attacker)
			return
		}
		switch a.Anim {
		case "resources/threefour/block-all-start.png", "resources/threefour/block-all-1.png":
			a.SetAnim("resources/threefour/block-all-1.png", AnimOptions{animDuration: BlockSpeed, totalFrames: 3, once: true, reset: true})
		case "resources/threefour/block-low-start.png", "resources/threefour/block-low-1.png":
			a.SetAnim("resources/threefour/block-low-1.png", AnimOptions{animDuration: BlockSpeed, totalFrames: 3, once: true, reset: true})
		case "resources/threefour/block-mid-v2-start.png", "resources/threefour/block-mid-v2-1.png":
			a.SetAnim("resources/threefour/block-mid-v2-1.png", AnimOptions{animDuration: BlockSpeed, totalFrames: 3, once: true, reset: true})
		case "resources/threefour/block-mid-start.png", "resources/threefour/block-mid-2.png":
			a.SetAnim("resources/threefour/block-mid-1.png", AnimOptions{animDuration: BlockSpeed, totalFrames: 3, once: true, reset: true})
		case "resources/threefour/block-mid-1.png":
			a.SetAnim("resources/threefour/block-mid-2.png", AnimOptions{animDuration: BlockSpeed, totalFrames: 3, once: true, reset: true})
		}
	}
}
func (a *Actor) enterJumpState(dt float32) {
	a.OnGround = false
	a.Speed.Y = MaxJumpSpeed * dt
	a.State = StateJump
}

func (a *Actor) enterJumpStartState() {
	a.State = StateJumpStart
	a.SetAnim("resources/threefour/jump-start.png", AnimOptions{animDuration: 0.7, totalFrames: 13, once: true, animYOffset: -48})
}

func (a *Actor) Interp(alpha float32) {
	a.InterpolatedPos.X = a.PrevPos.X*(1-alpha) + a.Pos.X*alpha
	a.InterpolatedPos.Y = a.PrevPos.Y*(1-alpha) + a.Pos.Y*alpha
}

func (a *Actor) enterClimbState() {
	a.State = StateClimb
	a.SetAnim("resources/threefour/jump-climb.png", AnimOptions{animDuration: 1.2, totalFrames: 16, once: true})
}

func (a *Actor) exitJumpState() {
	a.CanClimbRight = false
	a.CanClimbLeft = false
}

func (a *Actor) enterTurnoverState() {
	a.State = StateTurnover
	a.SetAnim("resources/threefour/turnover.png", AnimOptions{animDuration: 0.9, totalFrames: 14, once: true, reset: true})
}

func (a *Actor) exitAttackState() {
	a.AttackCount = 0
}

func (a *Actor) exitBlockState() {
	a.Attacker = nil
}

func (a *Actor) exitAttackedState() {
	a.AnimWhite = false
}

func (a *Actor) MouseOnLeft() bool {
	return a.Input.FaceLeft
}
