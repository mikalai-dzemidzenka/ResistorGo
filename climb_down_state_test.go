package resistor

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// climbDown presses S once for a, then runs it until it leaves the climb-down
// or n ticks pass, holding in on every tick after the press.
func climbDown(a *Actor, in Input, n int) {
	step := float32(FixedTimeStep)
	press := in
	press.Down = true
	a.Input = press
	a.Update(step)
	for range n {
		if a.State != StateClimbDown {
			return
		}
		a.Input = in
		a.Update(step)
	}
}

// settle runs a until it stands idle on something, or n ticks pass.
func settle(a *Actor, n int) {
	for range n {
		if a.OnGround && a.State == StateIdle {
			return
		}
		a.Update(float32(FixedTimeStep))
	}
}

func TestSClimbsDownFromAStand(t *testing.T) {
	defer dropLevel()()
	w := NewWorld(true)
	a := standOn(w, 100, 150)
	a.Input.Down = true
	a.Update(float32(FixedTimeStep))
	if a.State != StateClimbDown {
		t.Fatalf("S on the platform left state %d, want the climb-down %d", a.State, StateClimbDown)
	}
	if a.Anim != "resources/threefour/climb-down.png" || a.AnimCurrentFrame != 0 {
		t.Errorf("the climb-down plays %q from frame %d", a.Anim, a.AnimCurrentFrame)
	}
	if feet(a) != 150 {
		t.Errorf("the climb-down started with the feet at %v, want on the platform at 150", feet(a))
	}
}

func TestSOnSolidFloorDoesNothing(t *testing.T) {
	defer dropLevel()()
	w := NewWorld(true)
	a := standOn(w, 300, 300)
	a.Input.Down = true
	a.Update(float32(FixedTimeStep))
	if a.State != StateIdle || feet(a) != 300 {
		t.Errorf("S on solid floor left state %d, feet %v", a.State, feet(a))
	}
}

func TestSClimbsDownFromARun(t *testing.T) {
	defer dropLevel()()
	w := NewWorld(true)
	a := standOn(w, 60, 150)
	a.State = StateRun
	a.Speed.X = RunReach + 20
	a.Input.Left = true // running right
	a.Input.Down = true
	a.Update(float32(FixedTimeStep))
	if a.State != StateClimbDown {
		t.Fatalf("S at a run left state %d, want the climb-down", a.State)
	}
	if a.Speed.X != 0 {
		t.Errorf("the climb-down started at speed %v, want stopped", a.Speed.X)
	}
}

func TestSHeldDoesNotClimbAgain(t *testing.T) {
	defer dropLevel()()
	w := NewWorld(true)
	a := standOn(w, 100, 150)
	a.Input.Down = true
	var start uint32
	for i := range 20 {
		a.Update(float32(FixedTimeStep))
		if i == 0 {
			start = a.AnimStartTick
		}
	}
	if a.State != StateClimbDown || a.AnimStartTick != start {
		t.Errorf("S held restarted the climb-down (state %d, started %d, first %d)", a.State, a.AnimStartTick, start)
	}
}

func TestSWhileSwingingIsNotKept(t *testing.T) {
	defer dropLevel()()
	w := NewWorld(true)
	a := standOn(w, 100, 150)
	a.Updown = 1
	a.Input.Updown = 1
	a.Input.Attack = true
	a.Update(float32(FixedTimeStep))
	if a.State != StateAttack {
		t.Fatalf("the swing did not start: state %d", a.State)
	}
	a.Input = Input{Updown: 1, Down: true}
	a.Update(float32(FixedTimeStep))
	a.Input = Input{Updown: 1}
	for range 120 {
		if a.State == StateClimbDown {
			t.Fatal("S during a swing climbed down")
		}
		a.Update(float32(FixedTimeStep))
	}
	if a.State != StateIdle || feet(a) != 150 {
		t.Errorf("after the swing: state %d, feet %v; want standing on the platform", a.State, feet(a))
	}
}

func TestClimbDownPlaysInPlace(t *testing.T) {
	defer dropLevel()()
	w := NewWorld(true)
	a := standOn(w, 100, 150)
	x, o := a.Pos.X, a.Orientation
	step := float32(FixedTimeStep)
	a.Input.Down = true
	a.Update(step)
	var last uint8
	for a.State == StateClimbDown {
		last = a.AnimCurrentFrame
		if a.Pos.X != x || feet(a) != 150 || a.Orientation != o {
			t.Fatalf("mid climb-down the fighter moved to x %v feet %v facing %d", a.Pos.X, feet(a), a.Orientation)
		}
		a.Input = Input{Right: true, Attack: true, Jump: true, FaceLeft: true, Updown: 1}
		a.Update(step)
		if a.Ticks > 200 {
			t.Fatal("the climb-down never ended")
		}
	}
	if last != ClimbDownFrames-1 {
		t.Errorf("the climb-down ended on frame %d, want its last %d", last, ClimbDownFrames-1)
	}
	if a.State != StateFall {
		t.Errorf("after the climb-down the state is %d, want the fall %d", a.State, StateFall)
	}
	if feet(a) <= 150 {
		t.Errorf("after the climb-down the feet are at %v, not below the platform's top", feet(a))
	}
}

func TestClimbDownLandsBelow(t *testing.T) {
	defer dropLevel()()
	w := NewWorld(true)
	a := standOn(w, 100, 150)
	climbDown(a, Input{}, 100)
	if a.State != StateFall || feet(a) != 150+ClimbDownDrop {
		t.Fatalf("handed over in state %d with the feet at %v, want falling from %v", a.State, feet(a), 150+ClimbDownDrop)
	}
	for range 120 {
		if a.OnGround && a.State == StateIdle {
			break
		}
		if a.CanClimbLeft || a.CanClimbRight || a.State == StateClimb {
			t.Fatal("falling from the platform caught its ledge")
		}
		a.Update(float32(FixedTimeStep))
	}
	if feet(a) != 300 || a.State != StateIdle {
		t.Errorf("the fall ended with the feet at %v in state %d, want standing on the floor at 300", feet(a), a.State)
	}
}

func TestClimbDownOntoANearFloor(t *testing.T) {
	defer dropLevel()()
	LE.objects[0][Vector2i{X: 0, Y: 170}] = LevelObject{X: 0, Y: 170, CollisionRects: []rl.Rectangle{rect(0, 0, 200, 20)}}
	LE.buildIndex()
	w := NewWorld(true)
	a := standOn(w, 100, 150)
	climbDown(a, Input{}, 100)
	if feet(a) != 170 {
		t.Fatalf("handed over with the feet at %v, want on the floor 20 px below at 170", feet(a))
	}
	for range 5 {
		a.Update(float32(FixedTimeStep))
	}
	if feet(a) != 170 || !a.OnGround || a.State != StateIdle {
		t.Errorf("after the climb-down: feet %v, on ground %t, state %d; want standing at 170", feet(a), a.OnGround, a.State)
	}
}

func TestClimbDownThroughAStack(t *testing.T) {
	defer dropLevel()()
	LE.objects[0][Vector2i{X: 0, Y: 230}] = LevelObject{X: 0, Y: 230, DropRects: []rl.Rectangle{rect(0, 0, 200, 16)}}
	LE.buildIndex()
	w := NewWorld(true)
	a := standOn(w, 100, 150)
	climbDown(a, Input{}, 100)
	settle(a, 120)
	if feet(a) != 230 {
		t.Fatalf("the first climb-down landed the feet at %v, want on the second platform at 230", feet(a))
	}
	climbDown(a, Input{}, 100)
	settle(a, 120)
	if feet(a) != 300 {
		t.Errorf("the second climb-down landed the feet at %v, want on the floor at 300", feet(a))
	}
}

// midClimb is a fighter pressed S on the platform and ticks into its climb-down.
func midClimb(t *testing.T, ticks int) *Actor {
	w := NewWorld(true)
	a := standOn(w, 100, 150)
	a.Input.Down = true
	a.Update(float32(FixedTimeStep))
	a.Input = Input{}
	for range ticks {
		a.Update(float32(FixedTimeStep))
	}
	if a.State != StateClimbDown {
		t.Fatalf("not climbing down after %d ticks: state %d", ticks, a.State)
	}
	return a
}

func TestStunInterruptsClimbDown(t *testing.T) {
	defer dropLevel()()
	for _, name := range []string{"stun", "shove"} {
		a := midClimb(t, 12)
		b := NewActor(rl.Vector2{X: a.Pos.X + 10, Y: a.Pos.Y}, false)
		if name == "stun" {
			a.enterAttackedState(b)
		} else {
			a.enterShovedState(b)
		}
		for range 60 {
			a.Update(float32(FixedTimeStep))
		}
		if feet(a) != 150 || a.State == StateFall || a.State == StateClimbDown {
			t.Errorf("%s: after it the feet are at %v in state %d; want on top of the platform", name, feet(a), a.State)
		}
	}
}

func TestFlinchDoesNotInterruptClimbDown(t *testing.T) {
	defer dropLevel()()
	a := midClimb(t, 12)
	b := NewActor(rl.Vector2{X: a.Pos.X + 10, Y: a.Pos.Y}, false)
	a.enterFlinchState(b)
	x := a.Pos.X
	for range 60 {
		if a.State != StateClimbDown {
			break
		}
		if a.Pos.X != x {
			t.Fatalf("the flinch moved a fighter climbing down: x %v, was %v", a.Pos.X, x)
		}
		a.Update(float32(FixedTimeStep))
	}
	if a.State != StateFall || feet(a) <= 150 {
		t.Errorf("after a flinch: state %d feet %v; want the climb-down to end in the fall", a.State, feet(a))
	}
}

func TestDeathOnTheClimbDown(t *testing.T) {
	defer dropLevel()()
	a := midClimb(t, 12)
	a.enterDeathState(nil)
	for range 60 {
		a.Update(float32(FixedTimeStep))
	}
	if feet(a) != 150 {
		t.Errorf("a fighter killed climbing down ended with the feet at %v, want on top at 150", feet(a))
	}
}
