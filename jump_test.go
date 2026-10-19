package resistor

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// animTable's index is the animation's network id, so an entry may only be
// appended. These ids are spelled out: inserting anything ahead of them shifts
// every id that follows, and a client drawing from an older table would show
// the wrong animation for every fighter.
func TestAnimTableIsAppendOnly(t *testing.T) {
	for path, want := range map[string]uint8{
		"resources/threefour/run.png":                0,
		"resources/threefour/idle.png":               1,
		"resources/threefour/jump-start.png":         2,
		"resources/threefour/jump-climb.png":         3,
		"resources/threefour/updown.png":             4,
		"resources/threefour/hit-top.png":            24,
		"resources/threefour/death_fall.png":         25,
		"resources/threefour/turnover.png":           45,
		"resources/threefour/run-jump.png":           46,
		"resources/threefour/fall.png":               47,
		"resources/threefour/push.png":               48,
		"resources/threefour/battle-low-1-right.png": 50,
		"resources/threefour/climb-down.png":         51,
		"resources/threefour/duck.png":               52,
		"resources/threefour/roll.png":               53,
		"resources/threefour/roll-recover.png":       54,
		"resources/threefour/gun-bot-idle.png":       55,
	} {
		if got := animID(path); got != want {
			t.Errorf("%s has id %d, want %d - something was inserted rather than appended", path, got, want)
		}
	}

	// the newest one is on the end, and survives the round trip
	const newest = "resources/threefour/pistol-top.png"
	id := animID(newest)
	if want := uint8(len(animTable) - 1); id != want {
		t.Errorf("%s has id %d, want %d - it is not at the end of the table", newest, id, want)
	}
	if back := animPath(id); back != newest {
		t.Errorf("id %d came back as %q, want %q", id, back, newest)
	}
}

// ground puts a fighter standing on a floor of tiles, settled on it.
func ground(w *World, x float32) *Actor {
	tiles(Vector2i{X: 0, Y: 96}, Vector2i{X: 48, Y: 96}, Vector2i{X: 96, Y: 96}, Vector2i{X: 144, Y: 96})
	a := w.Spawn(rl.Vector2{X: x, Y: 96 - 48}, true)
	a.Orientation = 1
	a.OnGround = true
	return a
}

// The run-jump begins on the tick jump is pressed, and pushes off on the third
// frame of its own animation rather than in a wind-up state of its own.
func TestRunJumpPushesOffOnItsFrame(t *testing.T) {
	step := float32(FixedTimeStep)
	w := NewWorld(true)
	a := ground(w, 10)
	a.State = StateRun
	a.Speed.X = RunReach + 20
	a.Input.Jump = true

	a.Update(step)
	if a.State != StateRunJump {
		t.Fatalf("running + jump left state %d, want the run-jump %d", a.State, StateRunJump)
	}
	if a.Anim != "resources/threefour/run-jump.png" {
		t.Errorf("the run-jump plays %q, want the running-jump animation", a.Anim)
	}
	if !a.OnGround {
		t.Errorf("the run-jump left the ground on frame %d, want it to wait for frame %d", a.AnimCurrentFrame, RunJumpLaunchFrame)
	}

	// released early: the jump is committed from the tick it began
	a.Input.Jump = false
	var launchFrame uint8 = 255
	for range 40 {
		if !a.OnGround {
			break
		}
		if a.AnimCurrentFrame >= RunJumpLaunchFrame && launchFrame == 255 {
			launchFrame = a.AnimCurrentFrame
		}
		a.Update(step)
	}
	if a.OnGround {
		t.Fatalf("the run-jump never left the ground, stuck on frame %d", a.AnimCurrentFrame)
	}
	if launchFrame != RunJumpLaunchFrame {
		t.Errorf("the fighter left the ground on frame %d, want frame %d", launchFrame, RunJumpLaunchFrame)
	}
	if a.Speed.Y >= 0 {
		t.Errorf("the run-jump is not moving upward: %v", a.Speed.Y)
	}
}

// The standing jump still winds up first, and it is the shorter of the two:
// the run-jump leaves the ground harder for the run and the energy it costs.
func TestStandingJumpUnchanged(t *testing.T) {
	w := NewWorld(true)
	stand := ground(w, 10)
	stand.State = StateIdle
	stand.Input.Jump = true
	stand.Update(float32(FixedTimeStep))
	if stand.State != StateJumpStart {
		t.Fatalf("a standing jump left state %d, want the wind-up %d", stand.State, StateJumpStart)
	}
	if !stand.OnGround {
		t.Errorf("the standing jump left the ground before its wind-up finished")
	}
	if stand.Anim != "resources/threefour/jump-start.png" {
		t.Errorf("the standing jump plays %q, want the wind-up", stand.Anim)
	}

	// clear of the other: two fighters standing in one another shove each other
	// apart, and this test is about the jump
	runner := ground(w, 120)
	runner.State = StateRun
	runner.Speed.X = RunReach + 20
	runner.Input.Jump = true
	for range 40 { // run it up to its push-off frame
		runner.Update(float32(FixedTimeStep))
		runner.Input.Jump = false
		if !runner.OnGround {
			break
		}
	}
	if runner.OnGround {
		t.Fatalf("the run-jump never left the ground")
	}

	// let the wind-up play out and compare take-off speeds
	for range 60 {
		if stand.State == StateJump {
			break
		}
		stand.Update(float32(FixedTimeStep))
	}
	if stand.State != StateJump {
		t.Fatalf("the wind-up never finished: state %d", stand.State)
	}
	if runner.Speed.Y >= stand.Speed.Y {
		t.Errorf("the run-jump left at %v, no harder than the standing jump's %v", runner.Speed.Y, stand.Speed.Y)
	}
	if want := stand.Speed.Y * RunJumpLift; runner.Speed.Y != want {
		t.Errorf("the run-jump left at %v, want %v - the standing jump lifted by %v", runner.Speed.Y, want, float32(RunJumpLift))
	}
}

// The extra push-off speed is extra height: a run-jump clears more ground than
// a standing one, measured at the apex rather than at take-off.
func TestRunJumpReachesHigher(t *testing.T) {
	step := float32(FixedTimeStep)
	apex := func(start func(a *Actor)) float32 {
		w := NewWorld(true)
		a := ground(w, 10)
		start(a)
		a.Input.Jump = true
		top := a.Pos.Y
		for range 120 {
			a.Update(step)
			a.Input.Jump = false
			if a.Pos.Y < top {
				top = a.Pos.Y
			}
			if a.OnGround && a.State != StateJumpStart && a.State != StateRunJump {
				break
			}
		}
		return a.Pos.Y - top // how far up it got, from where it stood
	}
	stand := apex(func(a *Actor) { a.State = StateIdle })
	run := apex(func(a *Actor) { a.State, a.Speed.X = StateRun, RunReach })
	if stand <= 0 {
		t.Fatalf("the standing jump reached %v, want it off the ground", stand)
	}
	if run <= stand {
		t.Errorf("the run-jump topped out %v above the ground, no higher than the standing jump's %v", run, stand)
	}
}

// A run-jump carries the run's momentum: the speed it left at is the speed it
// had, and a faster run covers more ground.
func TestRunJumpKeepsMomentum(t *testing.T) {
	w := NewWorld(true)
	// the jump itself must not touch horizontal speed: a runner that jumps and
	// one that does not are moving at the same speed a tick later, whatever
	// friction took from both of them
	jumped := ground(w, 10)
	jumped.State = StateRun
	jumped.Speed.X = RunReach
	jumped.Input.Jump = true
	kept := ground(w, 120) // clear of the other, so neither shoves the other
	kept.State = StateRun
	kept.Speed.X = RunReach
	jumped.Update(float32(FixedTimeStep))
	kept.Update(float32(FixedTimeStep))
	if jumped.Speed.X != kept.Speed.X {
		t.Errorf("taking off left %v horizontal speed, want the %v a fighter that stayed on the ground kept", jumped.Speed.X, kept.Speed.X)
	}
	if jumped.Speed.X < RunReach*0.9 {
		t.Errorf("the run-jump launched at %v, far below the run's %v", jumped.Speed.X, RunReach)
	}

	distance := func(speed float32) float32 {
		w := NewWorld(true)
		f := ground(w, 10)
		f.State = StateRun
		f.Speed.X = speed
		f.Input.Jump = true
		start := f.Pos.X
		for range 40 {
			f.Input.Jump = false
			f.Update(float32(FixedTimeStep))
		}
		return f.Pos.X - start
	}
	fast, slow := distance(RunReach), distance(RunReach/2)
	if fast <= slow {
		t.Errorf("jumping at %v covered %v, no further than %v at half that speed", RunReach, fast, slow)
	}
}

// Every way of ending up airborne and going down reaches the fall.
func TestEverythingFalls(t *testing.T) {
	step := float32(FixedTimeStep)

	t.Run("over a run-jump's apex", func(t *testing.T) {
		w := NewWorld(true)
		a := ground(w, 10)
		a.State, a.Speed.X, a.Input.Jump = StateRun, RunReach, true
		a.Update(step)
		a.Input.Jump = false
		for range 120 {
			if a.State == StateFall {
				break
			}
			a.Update(step)
		}
		if a.State != StateFall {
			t.Fatalf("a run-jump reached state %d, want the fall %d", a.State, StateFall)
		}
		if a.Anim != "resources/threefour/fall.png" {
			t.Errorf("falling plays %q, want the falling animation", a.Anim)
		}
	})

	t.Run("over a standing jump's apex", func(t *testing.T) {
		w := NewWorld(true)
		a := ground(w, 10)
		a.State, a.Input.Jump = StateIdle, true
		for range 120 {
			a.Update(step)
			a.Input.Jump = false
			if a.State == StateFall {
				break
			}
		}
		if a.State != StateFall {
			t.Fatalf("a standing jump reached state %d, want the fall %d", a.State, StateFall)
		}
		if a.Anim != "resources/threefour/fall.png" {
			t.Errorf("the standing jump falls playing %q, want the falling animation", a.Anim)
		}
	})

	// walking off a ledge: the case the idle branch carried a TODO for
	t.Run("off a ledge", func(t *testing.T) {
		w := NewWorld(true)
		a := ground(w, 10)
		a.State = StateIdle
		a.Pos.X = 400 // past the end of the floor
		for range 10 {
			a.Update(step)
			if a.State == StateFall {
				break
			}
		}
		if a.State != StateFall {
			t.Errorf("walking off a ledge reached state %d, want the fall %d", a.State, StateFall)
		}
	})
}

// A fighter on the way up is not falling: the transition reads the direction of
// travel, not merely being off the ground.
func TestRisingIsNotFalling(t *testing.T) {
	w := NewWorld(true)
	a := ground(w, 10)
	a.State, a.Speed.X, a.Input.Jump = StateRun, RunReach, true
	a.Update(float32(FixedTimeStep))
	a.Input.Jump = false

	for range 120 {
		if a.Speed.Y > 0 {
			break // past the apex, the fall is correct from here
		}
		if a.State == StateFall {
			t.Fatalf("a fighter rising at %v is in the fall state", a.Speed.Y)
		}
		a.Update(float32(FixedTimeStep))
	}
}

// A landing returns the fighter to what it was doing, and does not lock it out.
func TestLanding(t *testing.T) {
	step := float32(FixedTimeStep)
	land := func(speed float32, attack bool) *Actor {
		w := NewWorld(true)
		a := ground(w, 10)
		a.State = StateFall
		a.Pos.Y -= 20
		a.OnGround, a.Speed.Y, a.Speed.X = false, 100, speed
		for range 60 {
			a.Input.Left = speed > 0 // hold the direction, as a player would
			a.Input.Attack = attack
			a.Update(step)
			if a.OnGround {
				break
			}
		}
		return a
	}

	if a := land(MaxSpeed, false); a.State != StateRun {
		t.Errorf("landing at speed left state %d, want the run %d", a.State, StateRun)
	}
	if a := land(0, false); a.State != StateIdle {
		t.Errorf("landing stopped left state %d, want idle %d", a.State, StateIdle)
	}
	if a := land(0, true); a.State != StateAttack {
		t.Errorf("attacking on the landing tick left state %d, want the swing %d", a.State, StateAttack)
	}
}

// Airborne fighters steer: in either jump and while falling, and a fighter that
// walked off a ledge steers as it falls. The movement rule already allowed
// this; these pin it so an airborne state cannot quietly lose it.
func TestSteeringInTheAir(t *testing.T) {
	step := float32(FixedTimeStep)

	// Input.Left accelerates in +x (the field names read as the button, not
	// the direction of travel)
	drift := func(state ActorState, anim bool) float32 {
		w := NewWorld(true)
		a := ground(w, 10)
		a.State = state
		a.OnGround, a.Speed.Y = false, 40
		a.Pos.Y -= 30
		if anim {
			a.SetAnim("resources/threefour/fall.png", AnimOptions{animDuration: 0.3, totalFrames: 3})
		}
		before := a.Speed.X
		for range 5 {
			a.Input.Left = true
			a.Update(step)
		}
		return a.Speed.X - before
	}

	for _, c := range []struct {
		name  string
		state ActorState
	}{
		{"a standing jump", StateJump},
		{"a run-jump", StateRunJump},
		{"a fall", StateFall},
	} {
		if d := drift(c.state, c.state == StateFall); d <= 0 {
			t.Errorf("holding a direction during %s changed speed by %v, want it to accelerate", c.name, d)
		}
	}

	// and a fighter can turn around in the air
	w := NewWorld(true)
	a := ground(w, 10)
	a.State, a.Speed.X, a.Input.Jump = StateRun, RunReach, true
	a.Update(step)
	a.Input.Jump = false
	for range 30 {
		a.Input.Right = true // the other way
		a.Update(step)
	}
	if a.Speed.X >= 0 {
		t.Errorf("holding the other direction in the air left speed at %v, want it reversed", a.Speed.X)
	}
}

// A ledge is caught rising into it, or coming down on the far side of a jump.
// A fighter that is only falling - off an edge, or knocked out of its jump -
// passes it by.
func TestOnlyAJumpCatchesALedge(t *testing.T) {
	for _, c := range []struct {
		name   string
		state  ActorState
		vy     float32
		jumped bool
		climb  bool
	}{
		{"rising", StateJump, -300, true, true},
		{"rolling up", StateRoll, -300, false, true},
		{"far side of a jump", StateFall, 5, true, true},
		{"falling off an edge", StateFall, 5, false, false},
		{"rolling off an edge", StateRoll, 5, false, false},
	} {
		tiles(Vector2i{X: 96}) // a tile at x 96..144, top at y 0
		a := jumper(90, 16)    // just left of it, head level with its top
		a.State, a.Speed.Y, a.Jumped = c.state, c.vy, c.jumped

		a.Update(float32(FixedTimeStep))
		if got := a.State == StateClimb; got != c.climb {
			t.Errorf("%s at a ledge: climbing %t, want %t (state %d)", c.name, got, c.climb, a.State)
		}
	}
}

// A climb is committed to the ledge it grabbed: holding the other direction
// does not turn a fighter halfway up, and neither does a bot spotting a target
// behind it. A climb that turned would mantle one way while facing the other.
func TestClimbRefusesToTurn(t *testing.T) {
	tiles(Vector2i{X: 96}) // a tile at x 96..144, top at y 0
	a := jumper(90, 16)    // just left of it, head level with its top
	step := float32(FixedTimeStep)

	a.Update(step)
	if a.State != StateClimb {
		t.Fatalf("the ledge was not grabbed: state %d", a.State)
	}

	a.Input = Input{Right: true} // the key that faces the other way, held throughout
	for range 90 {
		a.Update(step)
		if a.State != StateClimb {
			break
		}
		if a.Orientation != 1 {
			t.Fatalf("holding the other direction turned the climb to face %d", a.Orientation)
		}
	}

	// The rule belongs to the state, not to the player's keys: a bot that wants
	// to turn is refused the same turn. A bot only ever turns toward somewhere
	// it means to go, and the one place that can be behind it is the spot it
	// last saw a fighter at - so that is what asks for the turn here.
	floor()
	w := NewWorld(true)
	bot := w.Spawn(rl.Vector2{X: 100}, false)
	bot.Bot.mark(rl.Vector2{}) // last seen well to the bot's left
	bot.State, bot.Orientation = StateClimb, 1
	botInput(bot)
	if bot.Orientation != 1 {
		t.Errorf("a bot turned to face %d mid-climb", bot.Orientation)
	}

	// and off the ledge the same bot turns to go and look
	bot.State = StateIdle
	botInput(bot)
	if bot.Orientation != -1 {
		t.Errorf("out of the climb the bot faces %d, want it turned toward where it last saw someone", bot.Orientation)
	}
}

// A run-jump costs a third of the pool: three from full, and the fourth is
// refused.
func TestRunJumpCostsAThirdOfThePool(t *testing.T) {
	w := NewWorld(true)
	a := ground(w, 10)
	a.State, a.Speed.X = StateRun, RunReach

	jump := func() bool {
		a.State, a.OnGround, a.Speed.X, a.Speed.Y = StateRun, true, RunReach, 0
		a.Input.Jump = true
		out := a.enterRunJumpState()
		a.Input.Jump = false
		return out
	}

	if !jump() || a.Energy != MaxEnergy-EnergyRunJump {
		t.Errorf("the first run-jump left %v energy, want %v", a.Energy, MaxEnergy-EnergyRunJump)
	}
	if !jump() || a.Energy != MaxEnergy-2*EnergyRunJump {
		t.Errorf("the second run-jump left %v energy, want %v", a.Energy, MaxEnergy-2*EnergyRunJump)
	}
	if !jump() || a.Energy != 0 {
		t.Errorf("the third run-jump left %v energy, want 0", a.Energy)
	}
	if jump() {
		t.Errorf("a fourth run-jump came out at %v energy", a.Energy)
	}
	if want := int32(MaxEnergy / 3); EnergyRunJump != want {
		t.Errorf("a run-jump costs %d, want a third of the pool (%d)", int32(EnergyRunJump), want)
	}
}

// A jump that cannot be paid for does not happen: nothing is spent, the fighter
// stays on the ground and goes on running, and the refusal is raised.
func TestRunJumpRefusedWhenPoor(t *testing.T) {
	w := NewWorld(true)
	a := ground(w, 10)
	a.State, a.Speed.X = StateRun, RunReach
	a.Energy = EnergyRunJump - 1
	a.Events = 0
	a.Input.Jump = true

	a.Update(float32(FixedTimeStep))
	if a.State != StateRun {
		t.Errorf("a refused jump left state %d, want it still running %d", a.State, StateRun)
	}
	if !a.OnGround {
		t.Errorf("a refused jump left the ground")
	}
	if a.Energy != EnergyRunJump-1 {
		t.Errorf("a refused jump spent energy: %v left, want %v", a.Energy, int32(EnergyRunJump-1))
	}
	if a.Events&EvNoEnergy == 0 {
		t.Errorf("a refused jump did not raise the refusal")
	}
}

// The standing jump stays free, at any energy at all.
func TestStandingJumpIsFree(t *testing.T) {
	w := NewWorld(true)
	a := ground(w, 10)
	a.State, a.Energy, a.Input.Jump = StateIdle, 0, true

	a.Update(float32(FixedTimeStep))
	if a.State != StateJumpStart {
		t.Fatalf("a standing jump at 0 energy left state %d, want the wind-up %d", a.State, StateJumpStart)
	}
	if a.Energy != 0 {
		t.Errorf("the standing jump moved energy to %v, want it untouched at 0", a.Energy)
	}
}

// Spending on a jump holds the refill back exactly as spending on a swing does,
// and one jump still leaves enough to swing with.
func TestRunJumpDelaysTheRefill(t *testing.T) {
	w := NewWorld(true)
	a := ground(w, 10)
	a.State, a.Speed.X = StateRun, RunReach
	a.EnergyCalm = EnergyCalmTicks // already refilling

	a.enterRunJumpState()
	if a.EnergyCalm != 0 {
		t.Errorf("after a run-jump the calm wait is %d, want 0 - the refill did not start again", a.EnergyCalm)
	}
	if !a.canSwing() {
		t.Errorf("a fighter that run-jumped once from full cannot swing at %v energy", a.Energy)
	}
}

// A jump that catches a ledge climbs that ledge, not the next one up. Still
// rising after the catch, the fighter hung over the ledge through the climb,
// and the drop onto it afterwards caught the ledge above, through the solid it
// had just climbed.
func TestClimbDoesNotChainToTheLedgeAbove(t *testing.T) {
	saved := LE
	defer func() { LE = saved }()
	le := newLevelEditor(39, 83, nil, nil)
	le.objects[0] = LevelLayer{
		{}:     {CollisionRects: []rl.Rectangle{rect(0, 300, 800, 40)}},
		{X: 1}: {X: 0, Y: 234, CollisionRects: []rl.Rectangle{rect(150, 0, 400, 16)}}, // the ledge caught
		{X: 2}: {X: 0, Y: 174, CollisionRects: []rl.Rectangle{rect(170, 0, 400, 16)}}, // the one above
	}
	le.buildIndex()
	LE = le

	step := float32(FixedTimeStep)
	for start := float32(20); start < 150; start += 3 {
		for _, run := range []bool{true, false} {
			w := NewWorld(true)
			a := standOn(w, start, 300)
			if run {
				a.State = StateRun
				a.Speed.X = MaxSpeed
				a.Input.Right = true
			}
			a.Input.Jump = true
			for i := 0; i < 180; i++ {
				a.Update(step)
				a.Input.Jump = false
			}
			if feet := a.Pos.Y + 48; feet < 234 {
				t.Errorf("run=%v from x %v: ended with feet at %v, above the caught ledge's top 234", run, start, feet)
			}
		}
	}
}
