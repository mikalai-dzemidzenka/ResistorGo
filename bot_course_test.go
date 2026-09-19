package resistor

import (
	"os"
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// The test course, level_bot_test.json, is the level this change is meant to be
// played on: level_1.json carries four collision rectangles and no floor, so
// none of this can be seen in it. The course is a flat spawn, a step, a wall one
// body-height up, a gap that can be jumped, two steps to a high ledge, a drop
// back down, and a gap far too wide with a pit floor under it.
//
// It is loaded off the disk, so this covers the file itself: that it parses,
// that its rectangles come through, and that a bot can actually get along it.
func loadCourse(t *testing.T) {
	t.Helper()
	if _, err := os.Stat("level_bot_test.json"); err != nil {
		t.Fatalf("the test course is missing: %v", err)
	}
	if !DevTools {
		t.Skip("a release build reads levels from the embedded assets only")
	}
	LE = NewLevelEditor(39, 83)
	LE.LoadLevel("level_bot_test.json")
	if len(LE.solids) == 0 {
		t.Fatal("the test course loaded no collision rectangles")
	}
}

func TestTestCourseLoads(t *testing.T) {
	loadCourse(t)
	if len(LE.solids) != 33 {
		t.Errorf("the course has %d collision rectangles, want 33", len(LE.solids))
	}
	// the fighters spawn at Pos.Y 0, so the ground they land on must be at 48
	if _, ok := groundAt(0, 48, 1); !ok {
		t.Error("there is no ground under the spawn at x=0")
	}
	for _, x := range spawnXs {
		if _, ok := groundAt(x+21, 48, 1); !ok {
			t.Errorf("there is no ground under the spawn column x=%v", x)
		}
	}
}

// Every obstacle on the course, walked by a bot given somewhere to be.
func TestABotGetsAlongTheTestCourse(t *testing.T) {
	loadCourse(t)
	legs := []struct {
		name       string
		from, goal rl.Vector2
		wantX      float32
		wantY      float32
		ticks      int
	}{
		{"over the step", rl.Vector2{X: 200}, rl.Vector2{X: 420}, 400, 0, 600},
		{"up the wall", rl.Vector2{X: 380}, rl.Vector2{X: 500, Y: -48}, 456, -48, 900},
		{"across the gap", rl.Vector2{X: 540}, rl.Vector2{X: 860}, 816, 0, 900},
		{"up to the high ledge", rl.Vector2{X: 830}, rl.Vector2{X: 1060, Y: -96}, 968, -96, 1200},
		{"down off the ledge", rl.Vector2{X: 1050, Y: -96}, rl.Vector2{X: 1170}, 1120, 0, 900},
		{"along the strip", rl.Vector2{X: 1150}, rl.Vector2{X: 1380}, 1340, 0, 900},
	}
	for _, leg := range legs {
		w := NewWorld(true)
		bot := w.Spawn(leg.from, false)
		bot.Orientation = 1
		got := false
		for range leg.ticks {
			bot.Input = bot.walk(Input{Updown: bot.Updown}, leg.goal)
			bot.Update(float32(FixedTimeStep))
			if bot.OnGround && bot.Pos.X >= leg.wantX && abs(bot.Pos.Y-leg.wantY) <= 2 {
				got = true
				break
			}
		}
		if !got {
			t.Errorf("%s: the bot ended at x=%v y=%v, want it past x=%v on ground y=%v",
				leg.name, bot.Pos.X, bot.Pos.Y, leg.wantX, leg.wantY)
		}
	}
}

// And the one obstacle it is meant to lose to.
func TestABotStopsAtTheCoursesUncrossableGap(t *testing.T) {
	loadCourse(t)
	w := NewWorld(true)
	bot := w.Spawn(rl.Vector2{X: 1200}, false)
	bot.Orientation = 1
	goal := rl.Vector2{X: 1800} // the far side, well out of reach
	for range 1200 {
		bot.Input = bot.walk(Input{Updown: bot.Updown}, goal)
		bot.Update(float32(FixedTimeStep))
		if bot.Pos.Y > 100 {
			t.Fatalf("the bot went into the pit at x=%v", bot.Pos.X)
		}
	}
	if bot.Pos.X > 1424 {
		t.Errorf("the bot ended at x=%v, past the edge of the gap at 1424", bot.Pos.X)
	}
}

// The whole thing at once: a fighter walks the length of the course and a bot
// follows it the way it would in play, seeing, losing, searching, climbing the
// wall, jumping the gap, taking the stairs and coming down the drop.
//
// This is the test that found what playing it found. Driving a bot at a fixed
// goal, as the legs above do, never reproduces a bot that loses sight of
// something that moves - and losing sight at the wrong moment was the failure.
func TestABotFollowsAFighterTheLengthOfTheCourse(t *testing.T) {
	loadCourse(t)
	w := NewWorld(true)
	bot := w.Spawn(rl.Vector2{X: 200}, false)
	bot.Orientation = 1
	leader := w.Spawn(rl.Vector2{X: 330}, true)
	leader.Orientation = 1
	// Along the strip below the ledge, and well short of the pit: the point is
	// to follow the course, not to stage a duel on the brink of a hole, where
	// the lunge that comes with every swing walks a fighter over the edge.
	goal := rl.Vector2{X: 1280}

	lostAt, worst := -1, 0
	for i := range 2500 {
		leader.Input = leader.walk(Input{Updown: leader.Updown}, goal)
		bot.Input = botInput(bot)
		// The swing is held back. A bot that catches up fights, and one that
		// fights long enough kills - and then there is nothing left to follow
		// and nothing the rest of the course can say. Everything that gets the
		// bot along the course is left as it is, the jumps included; what it
		// does once it arrives has its own tests.
		bot.Input.Attack = false
		leader.Update(float32(FixedTimeStep))
		bot.Update(float32(FixedTimeStep))
		if bot.canSee(leader) {
			if lostAt >= 0 && i-lostAt > worst {
				worst = i - lostAt
			}
			lostAt = -1
		} else if lostAt < 0 {
			lostAt = i
		}
		if bot.Pos.Y > 150 {
			t.Fatalf("the bot fell into the pit at tick %d, x=%v", i, bot.Pos.X)
		}
	}
	if lostAt >= 0 {
		t.Errorf("the bot ended blind, since tick %d: x=%v y=%v", lostAt, bot.Pos.X, bot.Pos.Y)
	}
	if d := abs(bot.Pos.X - leader.Pos.X); d > 80 {
		t.Errorf("the bot ended %v from the fighter it was following", d)
	}
	// A bot cannot be blind for longer than it is willing to look: the bound is
	// its patience, not a number picked to fit. The stretch that gets near it is
	// the one where the fighter drops off the high ledge, which takes the bot a
	// couple of seconds to work out.
	if worst > SightSearchTicks {
		t.Errorf("the bot was blind for %d ticks at a stretch, longer than it will search for (%d)", worst, SightSearchTicks)
	}
}
