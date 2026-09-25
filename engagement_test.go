package resistor

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// lineUp puts a player on the floor with n bots lined up to its left, every one
// of them facing it and far enough back to have to walk in.
func lineUp(w *World, n int) (*Actor, []*Actor) {
	floor()
	p := w.Spawn(rl.Vector2{}, true)
	p.OnGround = true
	bots := make([]*Actor, 0, n)
	for i := range n {
		b := w.Spawn(rl.Vector2{X: -60 - 30*float32(i)}, false)
		b.Orientation = 1 // facing +X, where the player is
		b.OnGround = true
		bots = append(bots, b)
	}
	return p, bots
}

// runFight drives the world the way runBots does, but holds p still and on its
// feet, and reports whether any bot without a place ever pressed the attack.
//
// The pinning is what makes a line something to measure. Every spot in the line
// is worked out from where the fighter stands, so a fighter being shoved up and
// down the level drags the whole line after it and none of the bots ever arrive
// anywhere. Keeping it on its feet matters for the same reason: a player that
// goes down takes the fight and every place at it out of the world.
func runFight(w *World, n int, p *Actor) (swungFromTheLine bool) {
	at := p.Pos.X
	for range n {
		w.Tick++
		p.Structure, p.HitCount = MaxStructure, 0
		p.Pos.X, p.Speed.X, p.ShoveLeft = at, 0, 0
		for _, a := range w.Actors {
			if !a.IsPlayer {
				a.Input = botInput(a)
				if a.Input.Attack && a.Bot.target == p && !a.holdsPlace() {
					swungFromTheLine = true
				}
			}
			a.Update(float32(FixedTimeStep))
		}
	}
	return swungFromTheLine
}

// engaged is the bots holding a place at t, and inReach the ones whose swing
// would land on it from where they stand.
func engaged(t *Actor) []*Actor {
	var held []*Actor
	for _, b := range t.Engaged {
		if b != nil {
			held = append(held, b)
		}
	}
	return held
}

func inReach(w *World, t *Actor) []*Actor {
	var near []*Actor
	for _, b := range w.Actors {
		if b != t && !b.IsPlayer && b.inSwingRange(t) {
			near = append(near, b)
		}
	}
	return near
}

// Which side of a fighter another one is on. Two on the same pixel column have
// no side, and there the facing decides.
func TestEngageSide(t *testing.T) {
	w := NewWorld(true)
	mid := w.Spawn(rl.Vector2{X: 100}, true)
	left := w.Spawn(rl.Vector2{X: 40}, false)
	right := w.Spawn(rl.Vector2{X: 160}, false)
	if got := left.engageSide(mid); got != -1 {
		t.Errorf("a fighter at 40 is on side %d of one at 100, want -1", got)
	}
	if got := right.engageSide(mid); got != 1 {
		t.Errorf("a fighter at 160 is on side %d of one at 100, want 1", got)
	}
	same := w.Spawn(rl.Vector2{X: 100}, false)
	for _, o := range []int8{-1, 1} {
		same.Orientation = o
		if got := same.engageSide(mid); got != o {
			t.Errorf("on the same column facing %d the side is %d, want %d", o, got, o)
		}
	}
}

// Two places at a fighter and no more: the first bot takes the side it is on,
// the second is sent round to the far side, and the third gets nothing.
func TestTwoPlacesOneASide(t *testing.T) {
	floor()
	w := NewWorld(true)
	p := w.Spawn(rl.Vector2{}, true)
	first := w.Spawn(rl.Vector2{X: -40}, false)
	second := w.Spawn(rl.Vector2{X: -80}, false)
	third := w.Spawn(rl.Vector2{X: -120}, false)

	if !first.claimPlace(p) {
		t.Fatal("the first bot got no place")
	}
	if first.Bot.side != -1 || p.Engaged[0] != first {
		t.Errorf("the first bot took side %d, want the left place", first.Bot.side)
	}
	if !second.claimPlace(p) {
		t.Fatal("the second bot got no place")
	}
	if second.Bot.side != 1 || p.Engaged[1] != second {
		t.Errorf("the second bot took side %d, want the far (right) place", second.Bot.side)
	}
	if third.claimPlace(p) {
		t.Error("a third bot took a place at a fighter that already had two")
	}
	if third.Bot.target != p || third.Bot.side != -1 {
		t.Errorf("the third bot waits at %v on side %d, want the player on its own side",
			third.Bot.target, third.Bot.side)
	}

	// Giving one up hands it to the next claimer.
	first.releasePlace()
	if p.Engaged[0] != nil {
		t.Fatal("the left place is still held after it was given up")
	}
	if !third.claimPlace(p) || p.Engaged[0] != third {
		t.Error("the waiting bot did not get the place that came free")
	}
}

// No place is ever held at an ally, and none at the dead.
func TestNoPlaceAtAnAllyOrTheDead(t *testing.T) {
	floor()
	w := NewWorld(true)
	bot := w.Spawn(rl.Vector2{X: -40}, false)
	mate := w.Spawn(rl.Vector2{}, false)
	if bot.claimPlace(mate) {
		t.Error("a bot took a place at its own teammate")
	}
	if mate.Engaged[0] != nil || mate.Engaged[1] != nil {
		t.Error("a teammate has places held at it")
	}
	corpse := w.Spawn(rl.Vector2{X: 40}, true)
	corpse.State = StateDeath
	if bot.claimPlace(corpse) {
		t.Error("a bot took a place at a dead fighter")
	}
}

// A bot that gives up its search gives up its place with it: a mind cleared
// without that would leave a place held by a bot that has forgotten it.
func TestForgettingGivesUpThePlace(t *testing.T) {
	floor()
	w := NewWorld(true)
	p := w.Spawn(rl.Vector2{}, true)
	bot := w.Spawn(rl.Vector2{X: -40}, false)
	bot.claimPlace(p)
	bot.forget()
	if p.Engaged[0] != nil || p.Engaged[1] != nil {
		t.Error("a bot that forgot everything is still holding a place")
	}
	if bot.Bot.target != nil {
		t.Error("a bot that forgot everything is still in a fight")
	}
}

// A bot dying frees the place it held, and the fighter it held it at losing its
// life leaves nobody holding one.
func TestDeathFreesPlaces(t *testing.T) {
	floor()
	w := NewWorld(true)
	p := w.Spawn(rl.Vector2{}, true)
	holder := w.Spawn(rl.Vector2{X: -40}, false)
	waiting := w.Spawn(rl.Vector2{X: -80}, false)
	holder.claimPlace(p)
	waiting.Bot.target, waiting.Bot.side = p, -1

	w.Kill(holder)
	w.updateRespawns() // reaps the dying
	if p.Engaged[placeIndex(-1)] != nil {
		t.Error("a dead bot is still holding its place")
	}
	if !waiting.claimPlace(p) {
		t.Error("the place a dead bot held did not come free")
	}

	w.Kill(p)
	w.updateRespawns()
	if waiting.Bot.target != nil {
		t.Errorf("a bot is still fighting a fighter that left the world: %v", waiting.Bot.target)
	}
}

// The front of the line stands past the reach of a swing thrown either way,
// and consecutive spots are a clear body apart.
func TestLineGeometryClearsTheFight(t *testing.T) {
	floor()
	w := NewWorld(true)
	p := w.Spawn(rl.Vector2{}, true)
	bot := w.Spawn(rl.Vector2{}, false)

	// Measured at the worst the drift allows: a bot swaying its whole
	// allowance toward the fight, and stopped at the near edge of its dead
	// zone on top of that.
	for _, side := range []int8{-1, 1} {
		bot.Pos.X = p.Pos.X + float32(side)*(bot.lineFront()-LineSway-botArrived)
		bot.Orientation = -side
		bot.checkCollision()
		p.Orientation = side
		p.checkCollision()
		if bot.inSwingRange(p) {
			t.Errorf("a bot at the front of the line on side %d is in reach of its own swing", side)
		}
		if p.inSwingRange(bot) {
			t.Errorf("a bot at the front of the line on side %d is in reach of the fighter's swing", side)
		}
	}

	// Two bots at their spots, each drifted its whole allowance and stopped as
	// far off as the dead zone allows, both leaning toward one another, still
	// do not overlap.
	gap := bot.lineSpacing() - 2*botArrived - 2*LineSway
	if gap <= bot.CollisionBox.Width {
		t.Errorf("the worst-case gap in the line is %v, want more than a body (%v)",
			gap, bot.CollisionBox.Width)
	}
}

// A bot waiting its turn shifts about rather than standing to attention, and
// stays clear of the fight while it does.
func TestAWaitingBotDoesNotStandStill(t *testing.T) {
	w := NewWorld(true)
	p, bots := lineUp(w, 3)
	runFight(w, 400, p) // let the line form

	seen := map[float32]bool{}
	for range 900 {
		w.Tick++
		p.Structure, p.HitCount = MaxStructure, 0
		p.Pos.X, p.Speed.X, p.ShoveLeft = 0, 0, 0
		for _, a := range w.Actors {
			if !a.IsPlayer {
				a.Input = botInput(a)
			}
			a.Update(float32(FixedTimeStep))
		}
		for _, b := range bots {
			if b.holdsPlace() || b.Bot.target != p {
				continue
			}
			seen[b.Pos.X] = true
			// Only once it has got where it is drifting to. A bot that has
			// just given its place up starts from inside the fight and walks
			// out of it, which is the changeover, not the drift.
			spot := b.lineSpot(p, b.Bot.side, b.lineRank(p, b.Bot.side))
			if abs(b.Pos.X-(spot.X+float32(b.Bot.sway))) > botArrived {
				continue
			}
			if b.inSwingRange(p) || p.inSwingRange(b) {
				t.Fatalf("a waiting bot drifted into the fight, at %v", b.Pos.X)
			}
		}
	}
	if len(seen) < 3 {
		t.Errorf("the waiting bots stood on %d different spots over 15 seconds, want several", len(seen))
	}
}

// The line is ordered by nearness: the bot closest to the fight stands at the
// front of it, and the spots step out one spacing at a time.
func TestLineRankAndSpot(t *testing.T) {
	floor()
	w := NewWorld(true)
	p := w.Spawn(rl.Vector2{}, true)
	near := w.Spawn(rl.Vector2{X: -60}, false)
	mid := w.Spawn(rl.Vector2{X: -100}, false)
	far := w.Spawn(rl.Vector2{X: -140}, false)
	for _, b := range []*Actor{near, mid, far} {
		b.Bot.target, b.Bot.side = p, -1
	}
	for i, b := range []*Actor{near, mid, far} {
		if got := b.lineRank(p, -1); got != i {
			t.Errorf("the bot at %v ranks %d in the line, want %d", b.Pos.X, got, i)
		}
	}
	// A bot holding a place is not in the line at all.
	near.claimPlace(p)
	if got := mid.lineRank(p, -1); got != 0 {
		t.Errorf("with the nearest bot engaged the next ranks %d, want 0", got)
	}

	for _, side := range []int8{-1, 1} {
		first := far.lineSpot(p, side, 0)
		second := far.lineSpot(p, side, 1)
		if want := p.Pos.X + float32(side)*far.lineFront(); first.X != want {
			t.Errorf("the front of the line on side %d is at %v, want %v", side, first.X, want)
		}
		if got := abs(second.X - first.X); got != far.lineSpacing() {
			t.Errorf("the spots are %v apart, want %v", got, far.lineSpacing())
		}
	}
}

// Three bots on one player: two fight it, from opposite sides, and the third
// waits out past the reach of a swing, facing the fight.
func TestThreeBotsLeaveTwoFighting(t *testing.T) {
	w := NewWorld(true)
	p, bots := lineUp(w, 3)
	swung := runFight(w, 600, p)

	if swung {
		t.Error("a bot with no place at the fight threw a swing at it")
	}
	if held := engaged(p); len(held) != 2 {
		t.Fatalf("%d bots hold a place at the player, want 2", len(held))
	}
	if p.Engaged[0] == nil || p.Engaged[1] == nil {
		t.Error("the two engaged bots are not one a side")
	}
	// Only of a bot that has got to its place. One still crossing is on the
	// wrong side by definition - walking through the player is how it gets to
	// the other side of it - and atPlace is the question the bot itself asks
	// before it will stop and swing.
	if left, right := p.Engaged[0], p.Engaged[1]; left != nil && right != nil {
		if left.atPlace(p) && left.Pos.X > p.Pos.X {
			t.Errorf("the bot at the left place stands at %v, right of the player at %v", left.Pos.X, p.Pos.X)
		}
		if right.atPlace(p) && right.Pos.X < p.Pos.X {
			t.Errorf("the bot at the right place stands at %v, left of the player at %v", right.Pos.X, p.Pos.X)
		}
	}
	if near := inReach(w, p); len(near) > 2 {
		t.Errorf("%d bots are in swing range of the player, want no more than 2", len(near))
	}

	// And the line does settle: at some point a bot with no place stands at its
	// spot, out of reach of the fight and facing it. Asked of one instant this
	// would be asking the wrong question - the line advances, and a bot that
	// has just given its place up is walking to the back of it with its back to
	// the fight, which is the line working rather than failing.
	settled := false
	for range 600 {
		w.Tick++
		p.Structure, p.HitCount = MaxStructure, 0
		p.Pos.X, p.Speed.X, p.ShoveLeft = 0, 0, 0
		for _, a := range w.Actors {
			if !a.IsPlayer {
				a.Input = botInput(a)
			}
			a.Update(float32(FixedTimeStep))
		}
		for _, b := range bots {
			if b.holdsPlace() || b.Bot.target != p {
				continue
			}
			spot := b.lineSpot(p, b.Bot.side, b.lineRank(p, b.Bot.side))
			atSpot := abs(b.Pos.X-spot.X) <= botArrived+LineSway
			clear := !b.inSwingRange(p) && !p.inSwingRange(b)
			looking := b.Orientation == facing(p.Pos.X-b.Pos.X)
			if atSpot && clear && looking {
				settled = true
			}
		}
	}
	if !settled {
		t.Error("no bot ever stood at its spot in the line, clear of the fight and facing it")
	}
}

// Bots waiting on one side stand apart rather than in a pile.
//
// Watched over a stretch of the fight rather than at one tick: the bots draw
// from the global random source, so no two runs are the same fight, and at any
// one tick most of the line may be walking - moving up a place, or to a new
// drift. Every tick is checked for a pile, and the stretch has to have held a
// line worth checking at least once.
func TestTheLineIsNotAPile(t *testing.T) {
	w := NewWorld(true)
	p, bots := lineUp(w, 5)
	runFight(w, 600, p)

	most := 0
	for range 300 {
		runFight(w, 1, p)
		// One line at a time. A bot on its way round to the far side walks the
		// length of the other side to get there, and passing through the line it
		// is leaving is not the pile this is about.
		// Bots that have arrived at the spot they are drifting to. One still
		// walking is in transit - crossing the line it is leaving, or stepping
		// back after giving its place up - and two bodies passing one another is
		// not the pile this is about.
		lines := map[int8][]*Actor{}
		for _, b := range bots {
			if b.holdsPlace() || b.Bot.target != p {
				continue
			}
			spot := b.lineSpot(p, b.Bot.side, b.lineRank(p, b.Bot.side))
			if abs(b.Pos.X-(spot.X+float32(b.Bot.sway))) > botArrived {
				continue
			}
			lines[b.Bot.side] = append(lines[b.Bot.side], b)
		}
		most = max(most, len(lines[-1])+len(lines[1]))
		for side, waiting := range lines {
			for i, a := range waiting {
				for _, b := range waiting[i+1:] {
					if aabb(a.collisionRect(), b.collisionRect()) {
						t.Fatalf("tick %d: two bots waiting on side %d stand in one another, at %v and %v",
							w.Tick, side, a.Pos.X, b.Pos.X)
					}
				}
			}
		}
	}
	if most < 2 {
		t.Fatalf("at most %d bots were waiting in line at once, want at least 2", most)
	}
}

// Waiting changes only where a bot walks and whether it swings: it is hit as
// any fighter is, and its guard goes on drifting.
func TestAWaitingBotIsStillAFighter(t *testing.T) {
	w := NewWorld(true)
	p, bots := lineUp(w, 3)
	runFight(w, 600, p)

	var waiting *Actor
	for _, b := range bots {
		if !b.holdsPlace() {
			waiting = b
		}
	}
	if waiting == nil {
		t.Fatal("no bot is waiting in line")
	}

	// The guard goes on drifting while it waits. Only the ticks it spends
	// waiting count: a bot that takes a place and starts swinging holds its
	// guard still on purpose, and that is the swing's doing, not the line's.
	//
	// The drift is a roll, a change about one waiting tick in ninety, and the
	// bot takes its turn at the front between waits; so it is watched until it
	// has waited long enough that a guard still unmoved is the line's doing
	// and not the dice's - a chance in a million, where a fixed 600 ticks left
	// it one in thirty.
	drifted, waited, ticks := false, 0, 0
	was, wasWaiting := waiting.Updown, true
	for ; ticks < 6000 && waited < 1200; ticks++ {
		runFight(w, 1, p)
		if waiting.holdsPlace() {
			wasWaiting = false
			continue
		}
		waited++
		if wasWaiting && waiting.Updown != was {
			drifted = true
		}
		was, wasWaiting = waiting.Updown, true
	}
	if waited < 100 {
		t.Fatalf("the bot only waited %d of %d ticks, too few to say anything about its guard", waited, ticks)
	}
	if !drifted {
		t.Error("a waiting bot's guard never moved")
	}

	before := waiting.Structure
	waiting.takeHit(p)
	if waiting.Structure >= before {
		t.Error("a waiting bot shrugged a hit off")
	}
}

// Walking the fight out of sight frees both places.
func TestLosingSightFreesThePlace(t *testing.T) {
	w := NewWorld(true)
	p, _ := lineUp(w, 2)
	runFight(w, 300, p)
	if len(engaged(p)) == 0 {
		t.Fatal("no bot ever took a place")
	}
	p.Pos.X = 5000 // out of range of every cone
	runFight(w, 60, p)
	if held := engaged(p); len(held) != 0 {
		t.Errorf("%d places are still held at a fighter nobody can see", len(held))
	}
}

// The places are the authority's: a predicting client claims none.
func TestAClientClaimsNoPlace(t *testing.T) {
	floor()
	cl := NewWorld(false)
	p := cl.Spawn(rl.Vector2{}, true)
	p.OnGround = true
	for i := range 3 {
		b := cl.Spawn(rl.Vector2{X: -60 - 40*float32(i)}, false)
		b.Orientation, b.OnGround = 1, true
	}
	for range 300 {
		cl.Tick++
		for _, a := range cl.Actors {
			if !a.IsPlayer {
				a.Input = botInput(a)
			}
			a.Update(float32(FixedTimeStep))
			for _, e := range a.Engaged {
				if e != nil {
					t.Fatal("a place was claimed on a predicting client")
				}
			}
		}
	}
}

// A turn at the front runs out: with somebody waiting behind it, the holder
// steps back and the bot behind it steps in.
func TestTheLineAdvances(t *testing.T) {
	floor()
	w := NewWorld(true)
	p := w.Spawn(rl.Vector2{}, true)
	holder := w.Spawn(rl.Vector2{X: -40}, false)
	waiting := w.Spawn(rl.Vector2{X: -80}, false)
	holder.claimPlace(p)
	waiting.Bot.target, waiting.Bot.side = p, -1

	for range EngageHoldTicks {
		if !holder.engage(p) {
			break
		}
	}
	if holder.holdsPlace() {
		t.Fatal("the holder kept its place past its turn with a bot waiting")
	}
	if holder.Bot.yieldWait == 0 {
		t.Error("the bot that gave its place up did not stand off")
	}
	if !waiting.claimPlace(p) {
		t.Fatal("the waiting bot did not get the place")
	}
	// And the bot that stepped back cannot take it straight back.
	if holder.claimPlace(p) {
		t.Error("the bot that just yielded took a place again at once")
	}
}

// With nobody waiting, a holder keeps its place however long the fight runs.
func TestAHolderAloneKeepsItsPlace(t *testing.T) {
	floor()
	w := NewWorld(true)
	p := w.Spawn(rl.Vector2{}, true)
	holder := w.Spawn(rl.Vector2{X: -40}, false)
	for range EngageHoldTicks * 2 {
		if !holder.engage(p) {
			t.Fatal("a holder with nobody waiting gave its place up")
		}
	}
}

// Nothing about the line goes over the wire. The record was 46 bytes before
// the line; the one byte since is the fighter's team, which scripts change.
func TestPlacesAreNotReplicated(t *testing.T) {
	if actorRecSize != 46+1 {
		t.Errorf("the replicated actor record is %d bytes, want the 46 it was before the line and the team", actorRecSize)
	}
}

// Company slows the pressure down: a bot fighting alongside another leaves a
// longer gap before it opens an attack than one fighting alone ever does.
func TestABotAlongsideAnotherWaitsLonger(t *testing.T) {
	floor()
	w := NewWorld(true)
	p := w.Spawn(rl.Vector2{}, true)
	solo := w.Spawn(rl.Vector2{X: -40}, false)

	// Alone: one place held, and nothing drawn may reach the crowded floor.
	solo.claimPlace(p)
	loneWorst := uint16(0)
	for range 500 {
		solo.Bot.swingWait = 0
		if !solo.pressAttack(p) {
			t.Fatal("a bot with a full pool refused to open an attack")
		}
		loneWorst = max(loneWorst, solo.Bot.swingWait)
	}

	// With company: the other place taken too.
	mate := w.Spawn(rl.Vector2{X: 40}, false)
	mate.claimPlace(p)
	if p.engagedBy() != 2 {
		t.Fatalf("%d bots hold a place at the player, want 2", p.engagedBy())
	}
	crowdedBest := uint16(1 << 15)
	for range 500 {
		solo.Bot.swingWait = 0
		if !solo.pressAttack(p) {
			t.Fatal("a bot with a full pool refused to open an attack")
		}
		crowdedBest = min(crowdedBest, solo.Bot.swingWait)
	}
	if crowdedBest <= loneWorst {
		t.Errorf("the shortest wait with company is %d, the longest alone %d - they overlap",
			crowdedBest, loneWorst)
	}
}

// The same thing end to end: each of two bots fighting one player opens fewer
// attacks than a bot with that player to itself.
//
// Averaged over a good many fights, because one fight says very little. A fight
// is chaotic - shoves, crossings, a bot walking back into range - and the count
// from a single run of this swings by half either way whatever the wait is set
// to. What is exact is the wait itself, which TestABotAlongsideAnotherWaitsLonger
// pins down; this is here to show it survives contact.
//
// Twenty-four of them, where six used to do. Taking turns tidied the crowd up:
// two bots that no longer swing into one another are shoved, stunned and
// knocked out of position far less, so each of a pair now gets more openings in
// than it did - around seven in ten of what one alone manages, where it was
// nearer five. The gap is still there and still the wait's doing, but it is no
// longer wide enough to read off six fights.
func TestTwoAttackersOpenLessEach(t *testing.T) {
	opensPerBot := func(bots int) float64 {
		total := 0
		const fights = 24
		for range fights {
			floor()
			w := NewWorld(true)
			p := w.Spawn(rl.Vector2{}, true)
			p.OnGround = true
			side := []float32{-40, 40}
			for i := range bots {
				b := w.Spawn(rl.Vector2{X: side[i]}, false)
				b.Orientation, b.OnGround = int8(facing(-side[i])), true
			}
			for range 1800 {
				w.Tick++
				p.Structure, p.HitCount = MaxStructure, 0
				p.Pos.X, p.Speed.X, p.ShoveLeft = 0, 0, 0
				for _, a := range w.Actors {
					if a.IsPlayer {
						a.Update(float32(FixedTimeStep))
						continue
					}
					// Openings only. A chain blow is queued mid-swing and the
					// crowd wait has no say in it.
					wasSwinging := a.State == StateAttack
					a.Input = botInput(a)
					if a.Input.Attack && !wasSwinging {
						total++
					}
					a.Update(float32(FixedTimeStep))
				}
			}
		}
		return float64(total) / float64(fights*bots)
	}
	alone, paired := opensPerBot(1), opensPerBot(2)
	if alone == 0 || paired == 0 {
		t.Fatalf("nothing was opened at all: %v alone, %v paired", alone, paired)
	}
	// A wait of 40-83 ticks against 10-23 should cut the rate hard, but travel
	// and recovery are in the cycle too and the wait runs down during them, so
	// the margin asked for here is well short of what the constants alone say.
	if paired >= 0.85*alone {
		t.Errorf("each of a pair opened %.1f attacks a fight, one alone %.1f - the pair press no less each",
			paired, alone)
	}
	t.Logf("openings a fight per bot: %.1f alone, %.1f paired", alone, paired)
}

// One attacker at a time. Opening an attack takes the fighter for the length of
// the swing and a beat after it, and the bot at the other place is refused for
// as long as that turn runs.
func TestOpeningAnAttackTakesTheTurn(t *testing.T) {
	floor()
	w := NewWorld(true)
	p := w.Spawn(rl.Vector2{}, true)
	left := w.Spawn(rl.Vector2{X: -40}, false)
	right := w.Spawn(rl.Vector2{X: 40}, false)
	left.claimPlace(p)
	right.claimPlace(p)
	if p.engagedBy() != 2 {
		t.Fatalf("%d bots hold a place at the player, want 2", p.engagedBy())
	}

	// Nothing is held against the first to press: a free turn and a full pool.
	if !left.pressAttack(p) {
		t.Fatal("a bot with a free turn and a full pool refused to open an attack")
	}
	if left.Bot.turnLeft == 0 {
		t.Fatal("opening an attack took no turn")
	}

	// The other place is refused for every tick of it. The wait is cleared each
	// time round, so the turn is the only thing that can be doing the refusing.
	for left.Bot.turnLeft > 0 {
		right.Bot.swingWait = 0
		if right.pressAttack(p) {
			t.Fatalf("the other attacker opened with %d ticks of the turn left", left.Bot.turnLeft)
		}
		if right.Bot.swingWait != 0 {
			t.Fatal("a refused turn armed the wait: the bot would start its gap over")
		}
		left.Bot.turnLeft--
	}
	right.Bot.swingWait = 0
	if !right.pressAttack(p) {
		t.Error("the other attacker did not open once the turn had run out")
	}
}

// The turn is held up while the swing plays and runs down once it stops, so a
// bot that is killed or stunned mid-attack lets the fighter go a beat later
// rather than keeping it for the rest of the match.
func TestTheTurnOutlastsTheSwingAndNoMore(t *testing.T) {
	floor()
	w := NewWorld(true)
	b := w.Spawn(rl.Vector2{X: -40}, false)

	b.State = StateAttack
	b.Bot.turnLeft = 1
	botInput(b)
	if b.Bot.turnLeft != BotTurnBeatTicks {
		t.Errorf("a swinging bot's turn is %d ticks, want it held at %d", b.Bot.turnLeft, BotTurnBeatTicks)
	}

	b.State = StateAttacked // cut down mid-swing
	for range BotTurnBeatTicks {
		botInput(b)
	}
	if b.Bot.turnLeft != 0 {
		t.Errorf("%d ticks of turn left a beat after the swing stopped, want none", b.Bot.turnLeft)
	}
}

// The same thing end to end: over a long fight two bots never have attacks in
// the air at the same time.
//
// Counted over every bot rather than over the two places, because the two can
// differ for a moment - a bot that yields its place mid-swing is still swinging
// - and the whole point is that no two blows overlap, whoever is holding what.
func TestTwoAttackersNeverSwingTogether(t *testing.T) {
	floor()
	w := NewWorld(true)
	p := w.Spawn(rl.Vector2{}, true)
	p.OnGround = true
	for _, x := range []float32{-40, 40} {
		b := w.Spawn(rl.Vector2{X: x}, false)
		b.Orientation, b.OnGround = facing(-x), true
	}

	swings := 0
	for tick := range 3000 {
		w.Tick++
		p.Structure, p.HitCount = MaxStructure, 0
		p.Pos.X, p.Speed.X, p.ShoveLeft = 0, 0, 0
		for _, a := range w.Actors {
			if !a.IsPlayer {
				a.Input = botInput(a)
			}
			a.Update(float32(FixedTimeStep))
		}
		inTheAir := 0
		for _, a := range w.Actors {
			if !a.IsPlayer && a.State == StateAttack {
				inTheAir++
			}
		}
		if inTheAir > 1 {
			t.Fatalf("tick %d: %d attacks in the air at once", tick, inTheAir)
		}
		swings += inTheAir
	}
	if swings == 0 {
		t.Fatal("nothing was swung at all over the whole fight")
	}
	t.Logf("%d ticks of swing over 3000", swings)
}

// A body that stays on the floor - single player keeps them - must not keep
// its place in the fight with it: the bots waiting behind it for a turn at the
// front would wait out the match.
func TestABodyLeftLyingHoldsNoPlace(t *testing.T) {
	floor()
	w := NewWorld(true)
	w.BotsStayDead = true
	p := w.Spawn(rl.Vector2{}, true)
	holder := w.Spawn(rl.Vector2{X: -40}, false)
	waiting := w.Spawn(rl.Vector2{X: -80}, false)
	holder.claimPlace(p)
	waiting.Bot.target, waiting.Bot.side = p, -1

	holder.Structure = HitStructure
	holder.enterDeathState(nil)
	for range DeathFlickerTicks + RespawnTicks + 60 {
		w.Tick++
		w.updateRespawns()
		holder.Update(float32(FixedTimeStep))
	}
	if w.ByID(holder.ID) != holder {
		t.Fatalf("the body left the world: nothing to test")
	}
	if p.Engaged[placeIndex(-1)] == holder {
		t.Error("the body is still holding its place at the fighter it died fighting")
	}
	if !waiting.claimPlace(p) {
		t.Error("the bot behind it never got a turn at the front")
	}
	if holder.Bot.target != nil {
		t.Errorf("the body is still fighting %v", holder.Bot.target)
	}
}

// The two places at a fighter that dies are given up as it falls, so the bots
// that were on it are free to go and fight somebody else.
func TestABodyReleasesItsOwnAttackers(t *testing.T) {
	floor()
	w := NewWorld(true)
	w.BotsStayDead = true
	victim := w.Spawn(rl.Vector2{}, false)
	att := w.Spawn(rl.Vector2{X: -40}, false)
	att.Team = NoTeam // or the two of them are allies and neither engages
	att.claimPlace(victim)
	if !att.holdsPlace() {
		t.Fatalf("the attacker never took a place")
	}

	victim.enterDeathState(att)
	if victim.engagedBy() != 0 {
		t.Errorf("%d bots are still recorded as fighting the body", victim.engagedBy())
	}
	if att.claimPlace(victim) {
		t.Error("a bot took a place at a body")
	}
	if att.Bot.target != nil {
		t.Errorf("the attacker is still fighting the body")
	}
}

// The whole point, from the crowd's side: with the bodies left lying, a line
// of bots keeps two of its own on the fighter as the ones at the front fall.
func TestTheLineFillsThePlaceOfTheFallen(t *testing.T) {
	w := NewWorld(true)
	w.BotsStayDead = true
	p, _ := lineUp(w, 3)

	runFight(w, 240, p)
	if len(engaged(p)) != 2 {
		t.Fatalf("%d bots fighting before the death, want 2", len(engaged(p)))
	}
	fallen := engaged(p)[0]
	fallen.Structure = HitStructure
	fallen.enterDeathState(p)

	runFight(w, 240, p)
	held := engaged(p)
	if len(held) != 2 {
		t.Fatalf("%d bots fighting after the death, want 2: the line did not step in", len(held))
	}
	for _, b := range held {
		if b == fallen {
			t.Error("the body is one of the two fighting")
		}
	}
}

// A closed place opens again after exactly RollClosesTicks.
func TestAClosedPlaceOpensAfterTwoSeconds(t *testing.T) {
	_, a := walker(t, -400)
	a.ClosedSide, a.ClosedTicks = 1, RollClosesTicks
	steps(a, RollClosesTicks-1, Input{})
	if a.ClosedTicks == 0 {
		t.Fatal("the place opened a tick early")
	}
	steps(a, 1, Input{})
	if a.ClosedTicks != 0 {
		t.Errorf("after %d ticks the place is still closed for %d more", RollClosesTicks, a.ClosedTicks)
	}
}

// closedScene is a player at x 0 whose right-hand place a landed roll has
// just closed, and a bot at each x given.
func closedScene(xs ...float32) (*Actor, []*Actor) {
	floor()
	w := NewWorld(true)
	p := w.Spawn(rl.Vector2{}, true)
	p.ClosedSide, p.ClosedTicks = 1, RollClosesTicks
	bots := make([]*Actor, 0, len(xs))
	for _, x := range xs {
		bots = append(bots, w.Spawn(rl.Vector2{X: x}, false))
	}
	return p, bots
}

func TestAClosedPlaceIsNotTaken(t *testing.T) {
	p, b := closedScene(40)
	if b[0].claimPlace(p) || p.Engaged != [2]*Actor{} {
		t.Errorf("a bot on the closed side took a place: %v", p.Engaged)
	}
}

func TestNoCrossingToAClosedPlace(t *testing.T) {
	p, b := closedScene(-40, -60)
	if !b[0].claimPlace(p) {
		t.Fatal("the first bot on the open side took no place")
	}
	if b[1].claimPlace(p) || p.Engaged[1] != nil {
		t.Errorf("the second bot crossed to the closed place: %v", p.Engaged)
	}
}

func TestHoldersStandBackFromAClosedPlace(t *testing.T) {
	// one holds the closed place, one holds the open place from the closed side
	p, b := closedScene(40, 20)
	p.Engaged = [2]*Actor{b[1], b[0]}
	b[0].Bot.target, b[0].Bot.side = p, 1
	b[1].Bot.target, b[1].Bot.side = p, -1
	for i, bot := range b {
		if bot.claimPlace(p) || bot.holdsPlace() {
			t.Errorf("bot %d kept its place", i)
		}
		if bot.Bot.target != p || bot.Bot.side != 1 {
			t.Errorf("bot %d is not waiting on the side it stands: target %v side %d", i, bot.Bot.target, bot.Bot.side)
		}
	}
	if p.Engaged != [2]*Actor{} {
		t.Errorf("places still held: %v", p.Engaged)
	}
}

// Driven the way play drives it: a bot coming at the fighter from the closed
// side stands in the line there for as long as the place is closed, with the
// open place free the whole time, and swings at nobody.
func TestABotAheadOfTheRollWaits(t *testing.T) {
	p, b := closedScene(120)
	b[0].Orientation = -1 // facing the player
	if swung := runFight(p.W, RollClosesTicks-10, p); swung {
		t.Error("a bot with no place swung")
	}
	if b[0].holdsPlace() || p.Engaged != [2]*Actor{} {
		t.Errorf("a bot on the closed side got a place: %v", p.Engaged)
	}
	if b[0].Bot.target != p || b[0].Pos.X <= p.Pos.X {
		t.Errorf("the bot is not waiting on the closed side: target %v x %v", b[0].Bot.target, b[0].Pos.X)
	}
}

// Once the window has run out the second attacker comes round again.
func TestTheSurroundComesBack(t *testing.T) {
	p, b := closedScene(-40, -60)
	if !b[0].claimPlace(p) || b[1].claimPlace(p) {
		t.Fatal("the scene did not start with one place held and the far one refused")
	}
	steps(p, RollClosesTicks, Input{})
	if p.ClosedTicks != 0 {
		t.Fatalf("the place is still closed for %d ticks", p.ClosedTicks)
	}
	if !b[1].claimPlace(p) || p.Engaged[1] != b[1] {
		t.Errorf("the second bot did not cross to the far place once it opened: %v", p.Engaged)
	}
}

// The whole escape, as it is played: surrounded, the player rolls out over the
// bot on its right and lands. For the next 2 seconds nobody holds the place on
// the side it rolled toward and nobody comes round to it; after that the
// surround is allowed back. The player is kept on its feet and never stunned -
// a stun would end the roll before it landed, and is its own test. Bots guard
// at random, so it is played a few times.
func TestRollingOutOfASurround(t *testing.T) {
	for run := range 10 {
		tiles(row(-960, 48, 60)...) // long enough that the roll lands on it
		w := NewWorld(true)
		p := w.Spawn(rl.Vector2{}, true)
		w.Spawn(rl.Vector2{X: -40}, false)
		right := w.Spawn(rl.Vector2{X: 40}, false)
		right.Orientation = -1
		tick := func(in Input) {
			w.Tick++
			p.Structure, p.HitCount, p.Energy = MaxStructure, 0, max(p.Energy, EnergyRunJump)
			p.Input = in
			for _, a := range w.Actors {
				if !a.IsPlayer {
					a.Input = botInput(a)
				}
				a.Update(float32(FixedTimeStep))
			}
		}
		surrounded := func() bool { return p.Engaged[0] != nil && p.Engaged[1] != nil }
		for range 300 {
			if surrounded() {
				break
			}
			tick(Input{})
		}
		if !surrounded() {
			t.Fatalf("run %d: the player was never surrounded: %v", run, p.Engaged)
		}

		p.Energy = MaxEnergy
		for range RollChargeTicks + 1 {
			tick(Input{Roll: true, RollAim: 128})
		}
		for range RollPathTicks {
			tick(Input{RollAim: 128})
			if p.State == StateRollRecover {
				break
			}
		}
		if p.State != StateRollRecover || p.ClosedSide != 1 || p.ClosedTicks == 0 {
			t.Fatalf("run %d: the roll did not land and close the right: state %d side %d ticks %d", run, p.State, p.ClosedSide, p.ClosedTicks)
		}

		fought := false
		for p.ClosedTicks > 1 {
			tick(Input{})
			fought = fought || p.Engaged[0] != nil
			if p.Engaged[1] != nil {
				t.Fatalf("run %d: %d ticks into the window a bot holds the closed place", run, RollClosesTicks-p.ClosedTicks)
			}
			if h := p.Engaged[0]; h != nil && h.engageSide(p) == p.ClosedSide {
				t.Fatalf("run %d: the holder of the open place is on the closed side", run)
			}
		}

		if !fought {
			t.Fatalf("run %d: nobody fought the player from the open side during the window", run)
		}

		came := false
		for range 600 {
			tick(Input{})
			if came = surrounded(); came {
				break
			}
		}
		if !came {
			t.Errorf("run %d: the surround never came back once the window closed: %v", run, p.Engaged)
		}
	}
}
