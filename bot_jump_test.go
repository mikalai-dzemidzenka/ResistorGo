package resistor

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// A floor long enough that nothing under test runs off the end of it.
func longFloor() []Vector2i { return row(-480, 48, 80) }

// arc flies one jump and reports how far the feet rose, how many ticks the
// fighter was off the ground, and how many it spent winding up before it left.
func arc(t *testing.T, runJump bool) (rise float32, air, windup int) {
	t.Helper()
	bot := loneBot(t, longFloor(), rl.Vector2{})
	if runJump {
		for range 120 { // up to a run before pressing it
			bot.Input = Input{Left: true, Updown: bot.Updown}
			bot.Update(float32(FixedTimeStep))
		}
	}
	startY, peak, launch := bot.Pos.Y, bot.Pos.Y, -1
	for i := range 300 {
		in := Input{Left: runJump, Updown: bot.Updown}
		if i == 0 {
			in.Jump = true
		}
		bot.Input = in
		bot.Update(float32(FixedTimeStep))
		if launch < 0 && !bot.OnGround {
			launch = i
		}
		if bot.Pos.Y < peak {
			peak = bot.Pos.Y
		}
		if launch >= 0 && i > launch+2 && bot.OnGround {
			return startY - peak, i - launch, launch
		}
	}
	t.Fatalf("the jump never finished (runJump=%t)", runJump)
	return
}

// What the bot believes about its jumps has to stay true of the jumps it
// actually makes: the beliefs are derived from the movement constants, and this
// is what notices when retuning those leaves the two disagreeing.
func TestJumpBeliefsMatchTheArcsFlown(t *testing.T) {
	cases := []struct {
		name                   string
		runJump                bool
		rise, airTicks, windup float32
		riseSlack, tickSlack   float32
	}{
		{"standing jump", false, JumpRise, JumpAirTicks, JumpWindupTicks, 5, 2},
		{"run-jump", true, RunJumpRise, RunJumpAirTicks, RunJumpWindupTicks, 10, 2},
	}
	for _, c := range cases {
		rise, air, windup := arc(t, c.runJump)
		// The belief must never promise more than the jump delivers: a bot that
		// thinks it rises higher than it does walks at walls it cannot clear.
		if c.rise > rise {
			t.Errorf("%s: believes it rises %v, actually rises %v - the belief must not be the taller",
				c.name, c.rise, rise)
		}
		if rise-c.rise > c.riseSlack {
			t.Errorf("%s: believes it rises %v, actually rises %v - %v short, more than the %v allowed",
				c.name, c.rise, rise, rise-c.rise, c.riseSlack)
		}
		if abs(c.airTicks-float32(air)) > c.tickSlack {
			t.Errorf("%s: believes it is airborne %v ticks, actually %d", c.name, c.airTicks, air)
		}
		if abs(c.windup-float32(windup)) > c.tickSlack {
			t.Errorf("%s: believes it winds up for %v ticks, actually %d", c.name, c.windup, windup)
		}
	}
}

// The standing jump has a long wind-up, and it has to survive being walked into
// or a bot could never use it to get anywhere.
func TestAStandingJumpSurvivesWalkingIntoIt(t *testing.T) {
	bot := loneBot(t, longFloor(), rl.Vector2{})
	bot.enterJumpStartState()
	for i := range 120 {
		bot.Input = Input{Left: true, Updown: bot.Updown}
		bot.Update(float32(FixedTimeStep))
		if bot.State == StateJump {
			return
		}
		if bot.State != StateJumpStart {
			t.Fatalf("the jump was lost at tick %d: state %d, not the jump", i, bot.State)
		}
	}
	t.Error("the jump never left the ground")
}

func TestWallAhead(t *testing.T) {
	const wallX = 240
	cases := []struct {
		name    string
		level   []Vector2i
		want    bool
		wantTop float32
	}{
		{"nothing ahead", longFloor(), false, 0},
		{"a one-tile wall", append(longFloor(), Vector2i{X: wallX, Y: 0}), true, 0},
		{"a wall of stacked tiles reports the topmost", append(longFloor(),
			Vector2i{X: wallX, Y: 0}, Vector2i{X: wallX, Y: -48}, Vector2i{X: wallX, Y: -96}), true, -96},
	}
	for _, c := range cases {
		bot := loneBot(t, c.level, rl.Vector2{X: 160})
		top, _, ok := bot.wallAhead(1, 120)
		if ok != c.want {
			t.Errorf("%s: wallAhead found=%t, want %t", c.name, ok, c.want)
			continue
		}
		if ok && top != c.wantTop {
			t.Errorf("%s: wall top %v, want %v", c.name, top, c.wantTop)
		}
	}
}

// A rise no taller than a step is not a wall: the fighter walks onto it, and
// the bot must not see it as something to jump.
func TestWallAheadIgnoresAStep(t *testing.T) {
	// a tile whose top stands StepHeight above the floor's top
	level := append(longFloor(), Vector2i{X: 240, Y: 48 - StepHeight})
	bot := loneBot(t, level, rl.Vector2{X: 160})
	if _, _, ok := bot.wallAhead(1, 120); ok {
		t.Error("a step-height rise was read as a wall")
	}
}

// Nothing behind the bot is in its way.
func TestWallAheadLooksOnlyAhead(t *testing.T) {
	level := append(longFloor(), Vector2i{X: 96, Y: 0})
	bot := loneBot(t, level, rl.Vector2{X: 240}) // the wall is well behind it
	if _, _, ok := bot.wallAhead(1, 240); ok {
		t.Error("a wall behind the bot was read as being in its way")
	}
}

func TestGroundAt(t *testing.T) {
	// a floor with one tile missing at x=240..288
	level := append(row(-480, 48, 15), row(288, 48, 20)...)
	tiles(level...)
	const feet, drop = 48, 64
	cases := []struct {
		name string
		x    float32
		want bool
	}{
		{"over the floor", 100, true},
		{"over the hole", 260, false},
		{"on the far side of the hole", 300, true},
		{"exactly on a tile seam", 96, true},
		{"past the end of the floor", 2000, false},
	}
	for _, c := range cases {
		if _, ok := groundAt(c.x, feet, drop); ok != c.want {
			t.Errorf("%s: groundAt(%v) found=%t, want %t", c.name, c.x, ok, c.want)
		}
	}
}

func TestEdgeAndLandingAcrossAGap(t *testing.T) {
	// floor up to x=288, a one-tile hole, then floor again from x=336
	level := append(row(-480, 48, 16), row(336, 48, 20)...)
	bot := loneBot(t, level, rl.Vector2{X: 240})
	bot.Speed.X = MaxSpeed
	edge, ok := bot.edgeAhead(1, 120, 64)
	if !ok {
		t.Fatal("the hole in the floor was not found")
	}
	if edge < 280 || edge > 292 {
		t.Errorf("the edge is at x=%v, want it at the end of the floor near 288", edge)
	}
	if _, ok := bot.landingBeyond(1, edge, bot.reach(RunJumpAirTicks), 64); !ok {
		t.Error("no landing was found across a gap a run-jump crosses")
	}
	if _, ok := bot.landingBeyond(1, edge, 8, 64); ok {
		t.Error("a landing was found across a gap far wider than the reach given")
	}
}

// The floor simply ending is a gap with nothing on the far side.
func TestNoLandingPastTheEndOfTheLevel(t *testing.T) {
	bot := loneBot(t, row(-480, 48, 16), rl.Vector2{X: 240})
	bot.Speed.X = MaxSpeed
	edge, ok := bot.edgeAhead(1, 120, 64)
	if !ok {
		t.Fatal("the end of the floor was not found")
	}
	if _, ok := bot.landingBeyond(1, edge, bot.reach(RunJumpAirTicks), 64); ok {
		t.Error("found somewhere to land past the end of the level")
	}
}
