package resistor

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// The roll's launch speed grows with how long the duck was held, from
// RollMinSpeed on the first tick to RollMaxSpeed after RollChargeTicks, and no
// further. The cursor picks the angle it leaves at, from RollFlatDeg to
// RollSteepDeg above the ground. Speeds in px/s; to be tuned by eye in play.
// RollMaxX stays under MaxSpeed, which would cut a flat launch short anyway.
const (
	RollChargeTicks = 60
	RollMinSpeed    = 300
	RollMaxSpeed    = 520
	RollMaxX        = 300
	RollFlatDeg     = 15
	RollSteepDeg    = 75
	RollPathTicks   = 300 // the drawn arc gives up on a roll that has not landed by then
)

// rolling reports whether the fighter is anywhere in a roll: ducked, in the
// air, or recovering. None of it guards.
func (a *Actor) rolling() bool {
	return a.State == StateDuck || a.State == StateRoll || a.State == StateRollRecover
}

// rollLaunch is the speed a duck released now would roll off with. The charge
// is read from the duck's animation start, which is replicated and recorded
// already, and the angle from the command's RollAim, so every machine arrives
// at the same one.
func (a *Actor) rollLaunch() rl.Vector2 {
	t := float32(min(a.Ticks-a.AnimStartTick, RollChargeTicks)) / RollChargeTicks
	speed := lerp(RollMinSpeed, RollMaxSpeed, t)
	angle := float64(lerp(RollFlatDeg, RollSteepDeg, float32(a.Input.RollAim)/255)) * math.Pi / 180
	return rl.Vector2{
		X: min(speed*float32(math.Cos(angle)), RollMaxX) * float32(a.Orientation),
		Y: -speed * float32(math.Sin(angle)),
	}
}

// rollAim turns the cursor's direction from the fighter into a RollAim: 0 at
// RollFlatDeg or lower, 255 at RollSteepDeg or higher, so the roll leaves at
// the angle the cursor points along.
func rollAim(from, cursor rl.Vector2) uint8 {
	deg := math.Atan2(float64(from.Y-cursor.Y), math.Abs(float64(cursor.X-from.X))) * 180 / math.Pi
	return uint8(clamp((deg-RollFlatDeg)/(RollSteepDeg-RollFlatDeg), 0, 1) * 255)
}

// enterDuckState crouches to charge a roll, if the fighter can pay for one.
// Nothing is spent until it rolls, so a duck that ends another way is free.
func (a *Actor) enterDuckState() bool {
	if !a.canRunJump() {
		a.Events |= EvNoEnergy
		return false
	}
	a.State = StateDuck
	a.Speed.X = 0
	a.SetAnim("resources/threefour/duck.png", AnimOptions{animDuration: 0.2, totalFrames: 3, once: true, reset: true})
	return true
}

func (a *Actor) launchRoll() {
	a.spendEnergy(EnergyRunJump)
	a.Speed = a.rollLaunch()
	a.OnGround = false
	a.State = StateRoll
	a.SetAnim("resources/threefour/roll.png", AnimOptions{animDuration: 0.4, totalFrames: 5})
}

func (a *Actor) enterRollRecoverState() {
	a.State = StateRollRecover
	a.SetAnim("resources/threefour/roll-recover.png", AnimOptions{animDuration: 0.3, totalFrames: 4, once: true})
}

// updateRoll is the state switch's arm for all three stages of a roll.
func (a *Actor) updateRoll() {
	switch a.State {
	case StateDuck:
		// no return after the launch: the take-off tick must move the fighter,
		// or the next collision check stands it back on the floor
		if !a.Input.Roll {
			a.launchRoll()
		}
	case StateRoll:
		// a roll that carries the head up to a ledge takes hold of it, as a
		// jump does
		if (a.CanClimbRight && a.Orientation == 1) || (a.CanClimbLeft && a.Orientation == -1) {
			a.exitJumpState()
			a.enterClimbState()
		} else if a.OnGround {
			// the place ahead closes, so the crowd rolled out of comes at
			// the fighter from one side only for a while; see claimPlace
			a.ClosedSide, a.ClosedTicks = a.Orientation, RollClosesTicks
			a.enterRollRecoverState()
		}
	case StateRollRecover:
		if a.AnimFinished {
			a.land()
		}
	}
}

// rollPath is where the fighter's feet would go if the duck were released now:
// a copy of it is released and run through the real fixed step until it lands.
// Running the simulation itself, rather than a second copy of its maths, is
// what keeps the drawn arc from lying. A copy is safe to step - an Actor is
// recorded and restored by value - and a roll touches no other fighter.
func (a *Actor) rollPath() []rl.Vector2 {
	b := *a
	b.Input = Input{FaceLeft: a.Input.FaceLeft, Updown: a.Input.Updown, RollAim: a.Input.RollAim}
	b.FreezeTicks = 0
	feet := func() rl.Vector2 {
		return rl.Vector2{X: b.Pos.X + b.CollisionBox.X + b.CollisionBox.Width/2, Y: b.Pos.Y + b.CollisionBox.Y + b.CollisionBox.Height}
	}
	path := []rl.Vector2{feet()}
	for range RollPathTicks {
		b.Update(float32(FixedTimeStep))
		path = append(path, feet())
		if b.State != StateRoll {
			break
		}
	}
	return path
}

// drawRollArc dots the path a ducking fighter would roll along.
func drawRollArc(a *Actor) {
	if a == nil || a.State != StateDuck {
		return
	}
	path := a.rollPath()
	for i := 0; i < len(path); i += 2 {
		rl.DrawPixelV(rl.Vector2{X: roundf(path[i].X), Y: roundf(path[i].Y)}, rl.RayWhite)
	}
}
