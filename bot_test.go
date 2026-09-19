package resistor

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// botScene builds an authoritative world on a level of its own, with a bot and
// one other fighter placed in it. corners are the 48x48 tiles the level is made
// of, as tiles() takes them; no corners at all means the default floor.
//
// Every bot test starts here rather than from the world TestMain leaves behind:
// the level is global state that the tests which need a ledge or a hole replace
// wholesale, so a bot test that did not install its own would be reading
// whichever level happened to run last.
func botScene(t *testing.T, corners []Vector2i, botPos, targetPos rl.Vector2) (*Actor, *Actor) {
	t.Helper()
	if len(corners) == 0 {
		floor()
	} else {
		tiles(corners...)
	}
	w := NewWorld(true)
	bot := w.Spawn(botPos, false)
	bot.Orientation = 1
	target := w.Spawn(targetPos, true)
	return bot, target
}

// loneBot builds a world with a bot in it and nobody else.
func loneBot(t *testing.T, corners []Vector2i, pos rl.Vector2) *Actor {
	t.Helper()
	if len(corners) == 0 {
		floor()
	} else {
		tiles(corners...)
	}
	w := NewWorld(true)
	bot := w.Spawn(pos, false)
	bot.Orientation = 1
	return bot
}

// row returns the corners of an unbroken run of n tiles starting at x, y.
func row(x, y int32, n int) []Vector2i {
	corners := make([]Vector2i, 0, n)
	for i := range n {
		corners = append(corners, Vector2i{X: x + int32(i)*48, Y: y})
	}
	return corners
}

// The scene puts both fighters on their feet on the floor it installs: a bot
// left hanging in the air would fall through every test that follows.
func TestBotSceneStandsBothFightersOnTheGround(t *testing.T) {
	bot, target := botScene(t, nil, rl.Vector2{}, rl.Vector2{X: 60})
	for _, a := range []*Actor{bot, target} {
		a.checkCollision()
		if !a.OnGround {
			t.Errorf("fighter at x=%v is not on the ground", a.Pos.X)
		}
	}
	if bot.IsPlayer {
		t.Error("the bot was spawned as a player")
	}
	if !target.IsPlayer {
		t.Error("the target was spawned as a bot")
	}
	if len(bot.W.Actors) != 2 {
		t.Errorf("world holds %d actors, want 2", len(bot.W.Actors))
	}
}

// wallTile is a tile placed off the grid rows the floor helpers build, for the
// tests that need something standing between two fighters.
func wallTile(x, y int32) Vector2i { return Vector2i{X: x, Y: y} }

// addSolid puts one collision rectangle of any size into the level already
// installed, for the tests that need something thinner than a 48x48 tile -
// a pillar two fighters can stand either side of, within arm's reach.
func addSolid(x, y, w, h float32) {
	pos := Vector2i{X: int32(x), Y: int32(y)}
	LE.objects[0][pos] = LevelObject{
		X: pos.X, Y: pos.Y,
		CollisionRects: []rl.Rectangle{{Width: w, Height: h}},
	}
	LE.buildIndex()
}

func TestLineBlocked(t *testing.T) {
	// A floor under x=-48..288, and one tile standing on it at x=72..120,
	// reaching from y=0 down to the floor.
	level := append(row(-48, 48, 7), wallTile(72, 0))
	tiles(level...)

	cases := []struct {
		name string
		p, q rl.Vector2
		want bool
	}{
		{"a clear line over the floor", rl.Vector2{X: -40, Y: 20}, rl.Vector2{X: 60, Y: 20}, false},
		{"straight through the standing tile", rl.Vector2{X: -40, Y: 20}, rl.Vector2{X: 200, Y: 20}, true},
		{"just above the floor", rl.Vector2{X: -40, Y: 47}, rl.Vector2{X: 60, Y: 47}, false},
		{"just below the floor's top", rl.Vector2{X: -40, Y: 49}, rl.Vector2{X: 60, Y: 49}, true},
		{"vertical, down into the floor", rl.Vector2{X: 20, Y: 0}, rl.Vector2{X: 20, Y: 100}, true},
		{"vertical, stopping short of the floor", rl.Vector2{X: 20, Y: 0}, rl.Vector2{X: 20, Y: 40}, false},
		{"a long horizontal line meets the standing tile", rl.Vector2{X: -40, Y: 10}, rl.Vector2{X: 200, Y: 10}, true},
		{"a zero-length line in open air", rl.Vector2{X: 20, Y: 20}, rl.Vector2{X: 20, Y: 20}, false},
		{"a zero-length line inside a solid", rl.Vector2{X: 90, Y: 20}, rl.Vector2{X: 90, Y: 20}, true},
	}
	for _, c := range cases {
		if got := lineBlocked(c.p, c.q); got != c.want {
			t.Errorf("%s: lineBlocked(%v, %v) = %t, want %t", c.name, c.p, c.q, got, c.want)
		}
	}
}

// One case per way of failing to see something, from the spec's
// "A bot acts only on fighters it can see".
func TestCanSee(t *testing.T) {
	cases := []struct {
		name      string
		level     []Vector2i
		targetPos rl.Vector2
		want      bool
	}{
		{"in front, close and in the open", nil, rl.Vector2{X: 100}, true},
		{"behind the bot", nil, rl.Vector2{X: -100}, false},
		{"beyond the sight range", nil, rl.Vector2{X: SightRange + 60}, false},
		{"inside the sight range", nil, rl.Vector2{X: SightRange - 60}, true},
		{"steeply above, outside the cone", nil, rl.Vector2{X: 40, Y: -200}, false},
		{"above but inside the cone", nil, rl.Vector2{X: 140, Y: -48}, true},
		{"behind a wall", append(row(-48, 48, 7), wallTile(72, 0)), rl.Vector2{X: 140}, false},
	}
	for _, c := range cases {
		bot, target := botScene(t, c.level, rl.Vector2{}, c.targetPos)
		if got := bot.canSee(target); got != c.want {
			t.Errorf("%s: canSee = %t, want %t", c.name, got, c.want)
		}
	}
}

// The same wall stops being in the way once the fighter steps out from behind it.
func TestSightIsJudgedAfresh(t *testing.T) {
	bot, target := botScene(t, append(row(-48, 48, 7), wallTile(72, 0)), rl.Vector2{}, rl.Vector2{X: 140})
	if bot.canSee(target) {
		t.Fatal("saw a fighter through a wall")
	}
	target.Pos.X = 40 // out from behind it, on the bot's side
	if !bot.canSee(target) {
		t.Error("did not see the fighter once it stepped out from behind the wall")
	}
}

func TestSeenPicksTheNearestVisibleFighter(t *testing.T) {
	bot, near := botScene(t, nil, rl.Vector2{}, rl.Vector2{X: 100})
	far := bot.W.Spawn(rl.Vector2{X: 180}, true)
	got := bot.seen()
	if got == far {
		t.Errorf("saw the fighter at x=%v over the nearer one at x=%v", far.Pos.X, near.Pos.X)
	} else if got != near {
		t.Errorf("saw %v, want the nearer fighter at x=%v", got, near.Pos.X)
	}
}

// A fighter it cannot see does not count, however close it is.
func TestSeenIgnoresACloserFighterItCannotSee(t *testing.T) {
	bot, behind := botScene(t, nil, rl.Vector2{}, rl.Vector2{X: -40})
	ahead := bot.W.Spawn(rl.Vector2{X: 150}, true)
	if got := bot.seen(); got != ahead {
		t.Errorf("saw %v, want the visible fighter ahead at x=%v", got, ahead.Pos.X)
	}
	if bot.canSee(behind) {
		t.Error("saw the fighter standing behind it")
	}
}

func TestSeenIsNilWithNothingInView(t *testing.T) {
	bot := loneBot(t, nil, rl.Vector2{})
	if got := bot.seen(); got != nil {
		t.Errorf("a bot alone in the level saw %v", got)
	}
	bot2, _ := botScene(t, append(row(-48, 48, 7), wallTile(72, 0)), rl.Vector2{}, rl.Vector2{X: 140})
	if got := bot2.seen(); got != nil {
		t.Errorf("saw %v through a wall", got)
	}
}

// A bot with no one in view stands still, swings at nothing and keeps facing
// the way it was: spec, "A bot that sees no one does not fight".
func TestBotWithNobodyInSightDoesNothing(t *testing.T) {
	bot, _ := botScene(t, nil, rl.Vector2{}, rl.Vector2{X: -150}) // the other fighter behind it
	bot.Orientation = 1
	rolled := false
	for range 600 {
		in := botInput(bot)
		if in.Left || in.Right || in.Jump {
			t.Fatalf("a bot that can see no one asked to move: %+v", in)
		}
		if in.Attack {
			t.Fatal("a bot that can see no one asked to attack")
		}
		if bot.Orientation != 1 {
			t.Fatal("a bot that can see no one turned around")
		}
		if in.Updown != bot.Updown {
			rolled = true
		}
		bot.Updown = in.Updown
	}
	if !rolled {
		t.Error("a bot left alone never moved its guard")
	}
}

// And one that can see someone still behaves as it always did.
func TestBotApproachesAFighterItCanSee(t *testing.T) {
	bot, _ := botScene(t, nil, rl.Vector2{}, rl.Vector2{X: 150})
	in := botInput(bot)
	if !in.Left { // Left is +X: the input names are mirrored, see sampleInput
		t.Errorf("a bot with a fighter in view did not walk toward it: %+v", in)
	}
	if bot.Orientation != 1 {
		t.Errorf("bot faces %d, want it turned toward the fighter it sees", bot.Orientation)
	}
}

// Losing sight does not reach into a swing already thrown. The fighters stay
// within arm's reach throughout - a pillar goes up between their eyes, which
// breaks the line of sight and not the swing, because a hit is decided by the
// attack box alone and no wall has ever stopped one.
func TestASwingSurvivesLosingSight(t *testing.T) {
	bot, target := botScene(t, nil, rl.Vector2{}, rl.Vector2{X: 20})
	bot.Orientation, bot.Updown = 1, 2
	target.Updown = 0 // not guarding where the swing is coming in
	if !bot.canSee(target) {
		t.Fatal("the bot cannot see the fighter it is about to swing at")
	}
	if !bot.enterAttackState() {
		t.Fatal("the bot could not start a swing")
	}
	addSolid(30, 0, 2, 48) // a pillar between their eyes, mid-swing
	if bot.canSee(target) {
		t.Fatal("the pillar did not break the line of sight")
	}
	before := target.Structure
	for range 30 {
		bot.Input = botInput(bot)
		bot.Update(float32(FixedTimeStep))
		target.Update(float32(FixedTimeStep))
		if target.Structure < before {
			return
		}
	}
	t.Error("a swing started before sight was lost never landed")
}

// A hit from someone it cannot see still costs it, and still does not turn it.
func TestABotStillTakesAHitFromOutOfSight(t *testing.T) {
	bot, attacker := botScene(t, nil, rl.Vector2{}, rl.Vector2{X: -20})
	bot.Orientation = 1 // facing away from the attacker
	attacker.Orientation, attacker.Updown = 1, 2
	if bot.canSee(attacker) {
		t.Fatal("the attacker is meant to be out of the bot's sight")
	}
	before := bot.Structure
	bot.takeHit(attacker)
	if bot.Structure >= before {
		t.Errorf("structure %v, want less than %v after a hit from out of sight", bot.Structure, before)
	}
	if bot.Orientation != 1 {
		t.Error("the bot turned toward an attacker it cannot see")
	}
}
