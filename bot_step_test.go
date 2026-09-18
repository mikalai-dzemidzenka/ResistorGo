package resistor

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// runBots drives a world for n ticks, giving every bot its own command, the way
// an offline game or a server does. Unlike run() in structure_test.go the
// fighters here are steering themselves, which is the whole point.
func runBots(w *World, n int) {
	for range n {
		w.Tick++
		for _, a := range w.Actors {
			if !a.IsPlayer {
				a.Input = botInput(a)
			}
			a.Update(float32(FixedTimeStep))
		}
	}
}

// jumped reports whether the bot asked to leave the ground at any point.
func askedToJump(a *Actor, n int) bool {
	for range n {
		in := botInput(a)
		a.Input = in
		a.Update(float32(FixedTimeStep))
		if in.Jump {
			return true
		}
	}
	return false
}

// A one-tile wall between the bot and where it is going: it gets over it.
// Spec: "A wall is jumped and climbed".
func TestBotClimbsAWallInItsWay(t *testing.T) {
	const wallX = 240
	level := append(row(-480, 48, 40), Vector2i{X: wallX, Y: 0})
	bot := loneBot(t, level, rl.Vector2{X: 150})
	goal := rl.Vector2{X: 300, Y: -48} // on top of the wall and beyond
	for range 900 {
		bot.Input = bot.walk(Input{Updown: bot.Updown}, goal)
		bot.Update(float32(FixedTimeStep))
		if bot.OnGround && bot.Pos.Y <= -46 && bot.Pos.X >= wallX-24 {
			return
		}
	}
	t.Errorf("the bot never got onto the wall: x=%v y=%v state=%d", bot.Pos.X, bot.Pos.Y, bot.State)
}

// A wall far taller than any jump stops it, and it does not walk on the spot
// pressed into the tile for ever.
// Spec: "An impassable wall stops it".
func TestBotStopsAtAWallItCannotClear(t *testing.T) {
	const wallX = 240
	level := row(-480, 48, 40)
	for y := int32(-480); y < 48; y += 48 { // a wall ten tiles high
		level = append(level, Vector2i{X: wallX, Y: y})
	}
	bot := loneBot(t, level, rl.Vector2{X: 60})
	goal := rl.Vector2{X: 600}
	for range 600 {
		bot.Input = bot.walk(Input{Updown: bot.Updown}, goal)
		bot.Update(float32(FixedTimeStep))
	}
	if bot.Pos.X > wallX {
		t.Errorf("the bot got past a wall ten tiles high, ending at x=%v", bot.Pos.X)
	}
	if s := botStep(bot, goal); !s.stop {
		t.Errorf("at the wall the bot answers %+v, want it to come to rest", s)
	}
}

// A gap with ground on the far side is jumped.
// Spec: "A crossable gap is jumped".
func TestBotJumpsACrossableGap(t *testing.T) {
	// floor to x=288, one tile of nothing, floor again from x=336
	level := append(row(-480, 48, 16), row(336, 48, 20)...)
	bot := loneBot(t, level, rl.Vector2{X: 60})
	goal := rl.Vector2{X: 600}
	for range 900 {
		bot.Input = bot.walk(Input{Updown: bot.Updown}, goal)
		bot.Update(float32(FixedTimeStep))
		if bot.Pos.X > 336 && bot.OnGround {
			return
		}
	}
	t.Errorf("the bot never crossed the gap, ending at x=%v y=%v", bot.Pos.X, bot.Pos.Y)
}

// A gap with nothing on the far side is not walked into.
// Spec: "A bottomless gap is not walked into".
func TestBotStopsAtAGapItCannotCross(t *testing.T) {
	bot := loneBot(t, row(-480, 48, 16), rl.Vector2{X: 60}) // floor ends at x=288
	goal := rl.Vector2{X: 2000}
	for range 900 {
		bot.Input = bot.walk(Input{Updown: bot.Updown}, goal)
		bot.Update(float32(FixedTimeStep))
		if bot.Pos.Y > 100 {
			t.Fatalf("the bot walked off the end of the level: x=%v y=%v", bot.Pos.X, bot.Pos.Y)
		}
	}
	if bot.Pos.X > 288 {
		t.Errorf("the bot ended past the edge at x=%v", bot.Pos.X)
	}
}

// Spec: "A jump it cannot afford is a jump it cannot make".
func TestABotWithNoEnergyDoesNotCommitToARunJump(t *testing.T) {
	// a gap only the run-jump crosses
	level := append(row(-480, 48, 16), row(432, 48, 20)...)
	bot := loneBot(t, level, rl.Vector2{X: 60})
	bot.Energy = EnergyRunJump - 1
	goal := rl.Vector2{X: 700}
	for range 900 {
		// Kept poor at the top of the tick, not the bottom: the pool refills a
		// point every tick or two, so a bot topped up after its command was
		// built could afford the jump by the time the next one was.
		bot.Energy = EnergyRunJump - 1
		bot.Input = bot.walk(Input{Updown: bot.Updown}, goal)
		bot.Update(float32(FixedTimeStep))
		if bot.Pos.Y > 100 {
			t.Fatalf("a bot that cannot pay for a run-jump went into the gap: x=%v y=%v", bot.Pos.X, bot.Pos.Y)
		}
	}
}

// Spec: "No jumping for its own sake".
func TestBotDoesNotJumpOnClearGround(t *testing.T) {
	bot, _ := botScene(t, row(-480, 48, 40), rl.Vector2{X: 60}, rl.Vector2{X: 400})
	if askedToJump(bot, 300) {
		t.Error("the bot jumped crossing clear, flat ground")
	}
}

// Spec: "A step is walked onto".
func TestBotWalksOntoAStep(t *testing.T) {
	// a shelf one step-height above the floor, from x=240 on
	level := append(row(-480, 48, 16), row(240, 48-StepHeight, 20)...)
	bot := loneBot(t, level, rl.Vector2{X: 60})
	goal := rl.Vector2{X: 600}
	jumped := false
	for range 900 {
		in := bot.walk(Input{Updown: bot.Updown}, goal)
		jumped = jumped || in.Jump
		bot.Input = in
		bot.Update(float32(FixedTimeStep))
	}
	if jumped {
		t.Error("the bot jumped at a rise it could have walked onto")
	}
	if bot.Pos.X < 240 {
		t.Errorf("the bot did not get onto the step, ending at x=%v", bot.Pos.X)
	}
}

// Spec: "Following down" - a target on ground below is dropped down to.
func TestBotDropsDownAfterATargetBelow(t *testing.T) {
	// an upper shelf to x=288, and the floor proper one tile lower beyond it
	level := append(row(-480, 0, 16), row(288, 48, 20)...)
	bot := loneBot(t, level, rl.Vector2{X: 150, Y: -48})
	goal := rl.Vector2{X: 400} // down on the lower floor
	for range 900 {
		bot.Input = bot.walk(Input{Updown: bot.Updown}, goal)
		bot.Update(float32(FixedTimeStep))
		if bot.OnGround && bot.Pos.X > 300 && bot.Pos.Y >= -2 {
			return
		}
	}
	t.Errorf("the bot never came down after its target: x=%v y=%v", bot.Pos.X, bot.Pos.Y)
}

// Spec: "No route means standing still" - the far side is out of reach and
// there is no way round that the bot can read from where it stands.
func TestBotStandsWhenThereIsNoWayToTheTarget(t *testing.T) {
	// floor ends at x=288; the next ground is far out of any jump
	level := append(row(-480, 48, 16), row(1200, 48, 10)...)
	bot := loneBot(t, level, rl.Vector2{X: 60})
	goal := rl.Vector2{X: 1300}
	for range 900 {
		bot.Input = bot.walk(Input{Updown: bot.Updown}, goal)
		bot.Update(float32(FixedTimeStep))
		if bot.Pos.Y > 100 {
			t.Fatalf("the bot went into the gap: x=%v y=%v", bot.Pos.X, bot.Pos.Y)
		}
	}
	if s := botStep(bot, goal); !s.stop {
		t.Errorf("with no way across the bot answers %+v, want it to come to rest", s)
	}
	if abs(bot.Speed.X) > 1 {
		t.Errorf("the bot is still moving at %v", bot.Speed.X)
	}
}

// And it moves again the moment the target is somewhere it can go.
func TestBotMovesAgainWhenTheTargetBecomesReachable(t *testing.T) {
	level := append(row(-480, 48, 16), row(1200, 48, 10)...)
	bot := loneBot(t, level, rl.Vector2{X: 60})
	far := rl.Vector2{X: 1300}
	for range 600 {
		bot.Input = bot.walk(Input{Updown: bot.Updown}, far)
		bot.Update(float32(FixedTimeStep))
	}
	if s := botStep(bot, rl.Vector2{X: 150}); s.stop || s.dir == 0 {
		t.Errorf("with the target back on its own floor the bot answers %+v, want it walking", s)
	}
}

// A bot that finds itself too close to a gap to get up to a run backs away from
// the edge, takes a run at it and crosses. Standing at the edge deciding it
// cannot make it - which is true, from a standstill a stride away - and then
// never moving again is the failure this guards: the answer has to be to go and
// get the room, because a gap is crossed by the running jump or not at all.
func TestBotBacksOffToTakeARunAtAGap(t *testing.T) {
	// floor to x=288, a 56px gap, floor again from x=344
	level := append(row(-480, 48, 16), row(344, 48, 20)...)
	bot := loneBot(t, level, rl.Vector2{X: 240}) // a stride from the edge, at rest
	goal := rl.Vector2{X: 420}
	bot.Update(float32(FixedTimeStep)) // one tick to put it on its feet
	if s := botStep(bot, goal); s.dir != -1 {
		t.Fatalf("from a standstill by the edge the bot answers %+v, want it backing off", s)
	}
	backedOff := false
	for range 900 {
		bot.Input = bot.walk(Input{Updown: bot.Updown}, goal)
		bot.Update(float32(FixedTimeStep))
		if bot.Pos.X < 200 {
			backedOff = true
		}
		if bot.Pos.Y > 100 {
			t.Fatalf("the bot went into the gap at x=%v", bot.Pos.X)
		}
		if bot.OnGround && bot.Pos.X > 344 {
			if !backedOff {
				t.Error("the bot crossed without ever backing off for a run-up")
			}
			return
		}
	}
	t.Errorf("the bot never crossed: x=%v y=%v", bot.Pos.X, bot.Pos.Y)
}

// It does not back off into a second hole to get a run at the first.
func TestBotDoesNotBackOffIntoAnotherGap(t *testing.T) {
	// a short island: nothing before x=240, floor 240..336, a gap, floor beyond
	level := append(row(240, 48, 2), row(440, 48, 20)...)
	bot := loneBot(t, level, rl.Vector2{X: 290})
	goal := rl.Vector2{X: 520}
	for range 600 {
		bot.Input = bot.walk(Input{Updown: bot.Updown}, goal)
		bot.Update(float32(FixedTimeStep))
		if bot.Pos.Y > 100 {
			t.Fatalf("the bot backed off the island at x=%v", bot.Pos.X)
		}
	}
}

// A bot coming down off a ledge has to read the ground it is landing on. It
// used to steer freely in the air, so one that dropped onto a short strip with
// a hole past it landed at a run with no room left to stop, and went straight in.
func TestBotLandingOnAShortStripDoesNotRunOffIt(t *testing.T) {
	// a high ledge to x=288, a strip 96px long one tile below it, then nothing
	level := append(row(-480, 0, 16), row(288, 96, 2)...)
	bot := loneBot(t, level, rl.Vector2{X: 100, Y: -48})
	goal := rl.Vector2{X: 600}
	for range 900 {
		bot.Input = bot.walk(Input{Updown: bot.Updown}, goal)
		bot.Update(float32(FixedTimeStep))
		if bot.Pos.Y > 200 {
			t.Fatalf("the bot ran off the strip it landed on: x=%v y=%v", bot.Pos.X, bot.Pos.Y)
		}
	}
}

// When there is not even room to coast to a halt, it turns and pushes the other
// way: a fighter that turns loses two thirds of its speed on the spot, which is
// the only brake strong enough to matter.
func TestBotHardBrakesWhenCoastingWouldNotStopIt(t *testing.T) {
	bot := loneBot(t, row(-480, 48, 16), rl.Vector2{X: 200}) // floor ends at 288
	bot.Update(float32(FixedTimeStep))
	bot.Speed.X = MaxSpeed // at full tilt, a stride from the edge
	if s := botStep(bot, rl.Vector2{X: 600}); s.dir != -1 {
		t.Errorf("at full speed by the edge the bot answers %+v, want it braking the other way", s)
	}
}
