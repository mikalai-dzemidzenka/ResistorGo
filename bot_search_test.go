package resistor

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// hide takes the target somewhere the bot cannot see it without moving either
// of them: straight up, out of the cone.
func hide(target *Actor) { target.Pos = rl.Vector2{X: target.Pos.X, Y: -600} }

// Spec: "Walking to the last known spot".
func TestBotWalksToWhereItLastSawSomeone(t *testing.T) {
	bot, target := botScene(t, row(-480, 48, 40), rl.Vector2{X: 0}, rl.Vector2{X: 200})
	bot.Input = botInput(bot) // sees it
	if !bot.Bot.hasLast {
		t.Fatal("the bot did not note where the fighter was")
	}
	spot := target.Pos
	hide(target)
	for range 400 {
		bot.Input = botInput(bot)
		bot.Update(float32(FixedTimeStep))
		if abs(bot.Pos.X-spot.X) <= botArrived {
			return
		}
	}
	t.Errorf("the bot got to x=%v, want it at the spot it last saw the fighter, x=%v", bot.Pos.X, spot.X)
}

// Spec: "Giving up" - it waits a moment at the spot and then stops searching.
func TestBotGivesUpAfterReachingTheSpot(t *testing.T) {
	bot, target := botScene(t, row(-480, 48, 40), rl.Vector2{X: 0}, rl.Vector2{X: 200})
	bot.Input = botInput(bot)
	hide(target)
	arrived := -1
	for i := range 900 {
		bot.Input = botInput(bot)
		bot.Update(float32(FixedTimeStep))
		if arrived < 0 && abs(bot.Pos.X-200) <= botArrived {
			arrived = i
		}
		if arrived >= 0 && !bot.Bot.searching {
			if i-arrived > SightLingerTicks+2 {
				t.Errorf("it stood at the spot for %d ticks, want about %d", i-arrived, SightLingerTicks)
			}
			return
		}
	}
	t.Error("the bot never gave the search up")
}

// Spec: "Patience runs out short of the spot".
func TestBotGivesUpWhenTheSpotCannotBeReached(t *testing.T) {
	// the fighter is seen across a gap the bot has no way over
	level := append(row(-480, 48, 16), row(1200, 48, 10)...)
	bot, target := botScene(t, level, rl.Vector2{X: 100}, rl.Vector2{X: 220})
	bot.Input = botInput(bot)
	if !bot.Bot.hasLast {
		t.Fatal("the bot did not see the fighter to begin with")
	}
	target.Pos = rl.Vector2{X: 1250} // out of range, on the far side
	for i := range SightSearchTicks + 120 {
		bot.Input = botInput(bot)
		bot.Update(float32(FixedTimeStep))
		if bot.Pos.Y > 100 {
			t.Fatalf("the bot searched its way into the gap at tick %d", i)
		}
		if !bot.Bot.searching && i > 2 {
			return
		}
	}
	t.Error("the bot searched for ever")
}

// Spec: "A searching bot does not swing".
func TestASearchingBotDoesNotSwing(t *testing.T) {
	bot, target := botScene(t, row(-480, 48, 40), rl.Vector2{X: 0}, rl.Vector2{X: 200})
	bot.Input = botInput(bot)
	hide(target)
	for range SightSearchTicks {
		in := botInput(bot)
		if in.Attack {
			t.Fatal("a searching bot swung")
		}
		bot.Input = in
		bot.Update(float32(FixedTimeStep))
	}
}

// Spec: "Sight cancels the search".
func TestSightCancelsTheSearch(t *testing.T) {
	bot, target := botScene(t, row(-480, 48, 40), rl.Vector2{X: 0}, rl.Vector2{X: 200})
	bot.Input = botInput(bot)
	hide(target)
	for range 30 {
		bot.Input = botInput(bot)
		bot.Update(float32(FixedTimeStep))
	}
	if !bot.Bot.searching {
		t.Fatal("the bot is not searching, so there is nothing for sight to cancel")
	}
	target.Pos = rl.Vector2{X: bot.Pos.X + 120} // steps back into view
	bot.Input = botInput(bot)
	if bot.Bot.searching {
		t.Error("the bot went on searching with the fighter in plain view")
	}
	if bot.Bot.lastSeen != target.Pos {
		t.Errorf("the bot remembers %v, want where the fighter is now, %v", bot.Bot.lastSeen, target.Pos)
	}
}

// Spec: a respawned bot starts with a clean mind.
func TestARespawnedBotRemembersNothing(t *testing.T) {
	bot, target := botScene(t, row(-480, 48, 40), rl.Vector2{X: 0}, rl.Vector2{X: 200})
	bot.Input = botInput(bot)
	hide(target)
	for range 30 {
		bot.Input = botInput(bot)
		bot.Update(float32(FixedTimeStep))
	}
	if !bot.Bot.searching {
		t.Fatal("the bot is not searching")
	}
	w := bot.W
	w.Kill(bot)
	for range RespawnTicks + 2 {
		w.Tick++
		w.updateRespawns()
	}
	for _, a := range w.Actors {
		if a.IsPlayer || a == bot {
			continue
		}
		if a.Bot.searching || a.Bot.hasLast {
			t.Errorf("the bot came back still searching: %+v", a.Bot)
		}
		return
	}
	t.Fatal("the bot never came back")
}

// The search is what carries a bot over a wall its target went behind: sight is
// lost at the foot of the wall, and the spot it was last seen at is up on top.
func TestABotSearchesUpOntoALedge(t *testing.T) {
	const wallX = 240
	level := append(row(-480, 48, 40), Vector2i{X: wallX, Y: 0})
	bot, target := botScene(t, level, rl.Vector2{X: 150}, rl.Vector2{X: 250, Y: -48})
	if !bot.canSee(target) {
		t.Fatal("the bot cannot see the fighter standing on the wall")
	}
	for range 900 {
		bot.Input = botInput(bot)
		bot.Update(float32(FixedTimeStep))
		if bot.OnGround && bot.Pos.Y <= -46 && bot.Pos.X >= wallX-24 {
			return
		}
	}
	t.Errorf("the bot never got up onto the wall: x=%v y=%v searching=%t", bot.Pos.X, bot.Pos.Y, bot.Bot.searching)
}
