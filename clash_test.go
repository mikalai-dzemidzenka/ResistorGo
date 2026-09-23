package resistor

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// duelGap is how far apart duel stands its two fighters: each inside the
// other's swing, and not so close that either shoves the other.
const duelGap = 26

// duel stands two opposing fighters on the floor facing each other at arm's
// length, guarding low, the left one at x and the right one duelGap further on.
func duel(t *testing.T, x float32) (w *World, l, r *Actor) {
	t.Helper()
	floor()
	w = NewWorld(true)
	l = w.Spawn(rl.Vector2{X: x}, true)
	r = w.Spawn(rl.Vector2{X: x + duelGap}, true)
	l.Orientation, r.Orientation = 1, -1
	for range 10 {
		duelStep(w, l, Input{}, r, Input{})
	}
	if !l.OnGround || !r.OnGround || l.State != StateIdle || r.State != StateIdle {
		t.Fatalf("the duel did not settle: %d %d", l.State, r.State)
	}
	if !l.inSwingRange(r) || !r.inSwingRange(l) {
		t.Fatalf("the duel is not at arm's length")
	}
	return w, l, r
}

// duelStep runs one tick with the two fighters given these commands. The right
// fighter always faces left; hit-stop is skipped, as run skips it.
func duelStep(w *World, l *Actor, lin Input, r *Actor, rin Input) {
	rin.FaceLeft = true
	l.Input, r.Input = lin, rin
	run(w, 1)
}

var swingLow = Input{Attack: true}

// clashed runs the duel on with no more input until both are clashing, or
// gives up.
func clashed(w *World, l, r *Actor) bool {
	for range 20 {
		if l.State == StateClash && r.State == StateClash {
			return true
		}
		duelStep(w, l, Input{}, r, Input{})
	}
	return false
}

// openLow puts f straight into the first blow of a low swing opened now.
func openLow(w *World, f *Actor) {
	f.Updown = 0
	f.State = StateAttack
	f.SetAnim("resources/threefour/attack-low.png", AnimOptions{animDuration: 0.3, totalFrames: 8, once: true, reset: true})
	f.SwingTick = w.Tick
}

// An opening swing stamps the world's tick; a chained blow leaves the stamp.
func TestAnOpeningSwingStampsTheTick(t *testing.T) {
	w, l, r := duel(t, 0)
	r.Updown = 1 // guarding mid: the low swing lands, and nothing clashes
	w.Tick = 500
	duelStep(w, l, swingLow, r, Input{Updown: 1})
	if l.State != StateAttack || l.SwingTick != 501 {
		t.Fatalf("state %d, swing tick %d, want an attack stamped 501", l.State, l.SwingTick)
	}
	for range 60 {
		duelStep(w, l, swingLow, r, Input{Updown: 1})
		if l.Anim == "resources/threefour/attack-low-2.png" {
			break
		}
	}
	if l.Anim != "resources/threefour/attack-low-2.png" {
		t.Fatal("the swing never chained")
	}
	if l.SwingTick != 501 {
		t.Errorf("a chained blow restamped the swing tick to %d", l.SwingTick)
	}
}

// A clash is played out, not steered, and asks for nothing after it. The
// fighters are pressed at for most of it and then left alone, so what they do
// once it ends can only have been carried over.
func TestAClashIsNotSteered(t *testing.T) {
	w, l, r := duel(t, 0)
	openLow(w, l)
	openLow(w, r)
	if !tryClash(l, r) {
		t.Fatal("no clash")
	}
	lx, rx := l.Pos.X, r.Pos.X
	busy := Input{Right: true, Attack: true, Jump: true, FaceLeft: true}
	for tick := range 120 {
		if l.State != StateClash {
			break
		}
		in := busy
		if tick >= 30 { // ClashSeconds is 55 ticks
			in = Input{}
		}
		duelStep(w, l, in, r, in)
		if l.State == StateClash && (l.Pos.X != lx || r.Pos.X != rx || l.Orientation != 1 || r.Orientation != -1) {
			t.Fatalf("a clashing fighter moved or turned: %v %v facing %d %d", l.Pos.X, r.Pos.X, l.Orientation, r.Orientation)
		}
	}
	if l.State != StateIdle || r.State != StateIdle || l.AttackQueued || r.AttackQueued {
		t.Fatalf("after the clash: states %d %d, queued %v %v", l.State, r.State, l.AttackQueued, r.AttackQueued)
	}
	duelStep(w, l, Input{}, r, Input{})
	if l.State == StateAttack || r.State == StateAttack {
		t.Error("a press made during the clash was carried out after it")
	}
}

// Both pay half a hit and one swing, and neither is hit.
func TestAClashCostsBoth(t *testing.T) {
	w, l, r := duel(t, 0)
	l.HitCount, l.HitRun = 2, 2
	duelStep(w, l, swingLow, r, swingLow)
	if !clashed(w, l, r) {
		t.Fatal("no clash")
	}
	for _, f := range []*Actor{l, r} {
		if f.Structure != MaxStructure-ClashStructure {
			t.Errorf("fighter %d has %v structure, want %v", f.ID, f.Structure, float32(MaxStructure-ClashStructure))
		}
		if f.Energy != MaxEnergy-EnergySwing {
			t.Errorf("fighter %d has %d energy, want one swing spent", f.ID, f.Energy)
		}
		if f.Flickering() || f.FlinchTicks != 0 {
			t.Errorf("fighter %d flashes", f.ID)
		}
	}
	if l.HitCount != 2 || l.HitRun != 2 {
		t.Errorf("the clash counted as a hit: count %d run %d", l.HitCount, l.HitRun)
	}
}

// Structure the clash breaks kills, the other takes the kill, and there is no
// clash to play.
func TestAClashCanKill(t *testing.T) {
	w, l, r := duel(t, 0)
	l.Structure = ClashStructure
	openLow(w, l)
	openLow(w, r)
	if !tryClash(l, r) {
		t.Fatal("no clash")
	}
	if l.State != StateDeath {
		t.Errorf("the broken fighter is in state %d, want death", l.State)
	}
	if r.State != StateIdle {
		t.Errorf("the survivor is in state %d, want standing", r.State)
	}
	if r.Energy != MaxEnergy {
		t.Errorf("the survivor was not credited with the kill: %d energy", r.Energy)
	}
}

// The fighter on the left plays the left sheet facing right, the one on the
// right the right sheet facing left, the gap between them what the sheets are
// drawn for - whichever of the two opened first.
func TestSidesFollowPosition(t *testing.T) {
	for _, leftFirst := range []bool{true, false} {
		w, l, r := duel(t, 0)
		openLow(w, l)
		openLow(w, r)
		a, b := l, r
		if !leftFirst {
			a, b = r, l
		}
		if !tryClash(a, b) {
			t.Fatal("no clash")
		}
		if l.Anim != clashLeftAnim || l.Orientation != 1 || r.Anim != clashRightAnim || r.Orientation != -1 {
			t.Errorf("left plays %q facing %d, right plays %q facing %d", l.Anim, l.Orientation, r.Anim, r.Orientation)
		}
		if gap := r.Pos.X - l.Pos.X; gap != ClashGap {
			t.Errorf("they stand %v apart, want %d", gap, ClashGap)
		}
	}
}

// A clash beside a wall does not put the fighter into it.
func TestAClashStopsAtAWall(t *testing.T) {
	walls()
	w := NewWorld(true)
	l := w.Spawn(rl.Vector2{}, true)
	l.Pos.X = wallX - l.CollisionBox.X // flush against the wall
	r := w.Spawn(rl.Vector2{X: l.Pos.X + 20}, true)
	l.Orientation, r.Orientation = 1, -1
	for _, f := range []*Actor{l, r} {
		f.OnGround = true
		openLow(w, f)
	}
	if !tryClash(l, r) {
		t.Fatal("no clash")
	}
	if left := l.Pos.X + l.CollisionBox.X; left < wallX {
		t.Errorf("the clash put the fighter %v into the wall", wallX-left)
	}
}

// Every condition, one at a time: each case starts from two fighters that
// would clash and changes one thing.
func TestTryClashConditions(t *testing.T) {
	for _, c := range []struct {
		name string
		set  func(w *World, a, b *Actor)
		want bool
	}{
		{"both open on the same tick", func(w *World, a, b *Actor) {}, true},
		{"five ticks apart", func(w *World, a, b *Actor) { b.SwingTick -= 5 }, true},
		{"six ticks apart", func(w *World, a, b *Actor) { a.SwingTick += 6 }, true},
		{"seven ticks apart", func(w *World, a, b *Actor) { b.SwingTick -= 7 }, false},
		{"allies", func(w *World, a, b *Actor) { a.Team, b.Team = BotTeam, BotTeam }, false},
		{"back to back", func(w *World, a, b *Actor) { b.Orientation = 1 }, false},
		{"out of reach", func(w *World, a, b *Actor) { b.Pos.X += 60 }, false},
		{"the other swings mid", func(w *World, a, b *Actor) {
			b.SetAnim("resources/threefour/attack-mid.png", AnimOptions{animDuration: 0.3, totalFrames: 8, once: true})
		}, false},
		{"the other is a chained blow", func(w *World, a, b *Actor) {
			b.SetAnim("resources/threefour/attack-low-2.png", AnimOptions{animDuration: 0.2, totalFrames: 6, once: true})
		}, false},
		{"this one is a chained blow", func(w *World, a, b *Actor) {
			a.SetAnim("resources/threefour/attack-low-2.png", AnimOptions{animDuration: 0.2, totalFrames: 6, once: true})
		}, false},
		{"already clashing", func(w *World, a, b *Actor) { b.State = StateClash }, false},
		{"stunned", func(w *World, a, b *Actor) { b.State = StateAttacked }, false},
		{"dead", func(w *World, a, b *Actor) { b.State = StateDeath }, false},
		{"in the air", func(w *World, a, b *Actor) { b.OnGround = false }, false},
		{"answers from the block", func(w *World, a, b *Actor) { blockAgainst(b, a, true) }, true},
		{"blocking and not asking", func(w *World, a, b *Actor) { blockAgainst(b, a, false) }, false},
		{"answers from the block too late", func(w *World, a, b *Actor) {
			blockAgainst(b, a, true)
			w.Tick += 7
		}, false},
		{"answers from the block at mid", func(w *World, a, b *Actor) {
			blockAgainst(b, a, true)
			b.Updown = 1
		}, false},
		{"cannot pay from the block", func(w *World, a, b *Actor) {
			blockAgainst(b, a, true)
			b.Energy = EnergySwing - 1
		}, false},
		{"blocking someone else", func(w *World, a, b *Actor) {
			blockAgainst(b, a, true)
			b.Attacker = b
		}, false},
	} {
		t.Run(c.name, func(t *testing.T) {
			w, a, b := duel(t, 0)
			w.Tick = 100
			openLow(w, a)
			openLow(w, b)
			c.set(w, a, b)
			if got := tryClash(a, b); got != c.want {
				t.Errorf("tryClash = %v, want %v", got, c.want)
			}
		})
	}
}

// blockAgainst puts b in the low block a's swing throws it into, asking for a
// swing or not.
func blockAgainst(b, a *Actor, asking bool) {
	b.Updown = 0
	b.enterBlockState(a)
	b.Input = Input{Attack: asking, FaceLeft: b.Orientation < 0}
}

// A world that is not the authority never starts a clash.
func TestOnlyTheAuthorityClashes(t *testing.T) {
	w, a, b := duel(t, 0)
	w.Authority = false
	openLow(w, a)
	openLow(w, b)
	if tryClash(a, b) {
		t.Error("a client world started a clash")
	}
}

// Same tick, in either update order, through the real updates.
func TestSameTickClashes(t *testing.T) {
	for _, swap := range []bool{false, true} {
		w, l, r := duel(t, 0)
		if swap {
			w.Actors[0], w.Actors[1] = w.Actors[1], w.Actors[0]
		}
		duelStep(w, l, swingLow, r, swingLow)
		if !clashed(w, l, r) {
			t.Errorf("swapped %v: no clash, states %d %d", swap, l.State, r.State)
		}
	}
}

// The answer comes a few ticks late: the first swing has already thrown the
// other into a low block, and the press from it is caught there.
func TestCaughtFromTheBlock(t *testing.T) {
	for _, late := range []int{3, 5} {
		for _, swap := range []bool{false, true} {
			w, l, r := duel(t, 0)
			if swap {
				w.Actors[0], w.Actors[1] = w.Actors[1], w.Actors[0]
			}
			duelStep(w, l, swingLow, r, Input{})
			for range late - 1 {
				duelStep(w, l, Input{}, r, Input{})
			}
			if r.State != StateBlock {
				t.Fatalf("the answering fighter is in state %d, want the low block", r.State)
			}
			duelStep(w, l, Input{}, r, swingLow)
			if !clashed(w, l, r) {
				t.Errorf("%d ticks late, swapped %v: no clash, states %d %d", late, swap, l.State, r.State)
				continue
			}
			// the answer never paid for its swing as a swing: the clash charges it
			if r.Energy != MaxEnergy-EnergySwing {
				t.Errorf("the fighter answering from the block has %d energy, want one swing spent", r.Energy)
			}
		}
	}
}

// Too late, the block plays out exactly as if the press had never been made.
func TestOutsideTheWindowTheBlockPlaysOut(t *testing.T) {
	play := func(press bool) (*Actor, *Actor) {
		w, l, r := duel(t, 0)
		duelStep(w, l, swingLow, r, Input{})
		for tick := 2; tick <= 60; tick++ {
			rin := Input{}
			if press && tick == 10 {
				rin = swingLow
			}
			duelStep(w, l, Input{}, r, rin)
			if l.State == StateClash || r.State == StateClash {
				t.Fatalf("a clash %d ticks late", tick-1)
			}
		}
		return l, r
	}
	l1, r1 := play(true)
	l2, r2 := play(false)
	if l1.Structure != l2.Structure || r1.Structure != r2.Structure || l1.Energy != l2.Energy || r1.Energy != r2.Energy || r1.State != r2.State {
		t.Errorf("a late press changed the outcome: %v/%v %v/%v", r1.Structure, r2.Structure, r1.State, r2.State)
	}
}

// Mid against mid, low against high, a chained blow, back to back: no clash.
func TestOnlyOpeningLowSwingsFacingClash(t *testing.T) {
	t.Run("mid", func(t *testing.T) {
		w, l, r := duel(t, 0)
		l.Updown, r.Updown = 1, 1
		duelStep(w, l, Input{Attack: true, Updown: 1}, r, Input{Attack: true, Updown: 1})
		if clashed(w, l, r) {
			t.Error("mid swings clashed")
		}
	})
	t.Run("low and high", func(t *testing.T) {
		w, l, r := duel(t, 0)
		r.Updown = 2
		duelStep(w, l, swingLow, r, Input{Attack: true, Updown: 2})
		if clashed(w, l, r) {
			t.Error("a low and a high swing clashed")
		}
	})
	t.Run("chained", func(t *testing.T) {
		w, l, r := duel(t, 0)
		r.Updown = 1
		for range 60 {
			duelStep(w, l, swingLow, r, Input{Updown: 1})
			if l.Anim == "resources/threefour/attack-low-2.png" {
				break
			}
		}
		r.Updown = 0
		duelStep(w, l, swingLow, r, swingLow)
		if clashed(w, l, r) {
			t.Error("a chained blow clashed")
		}
	})
	t.Run("back to back", func(t *testing.T) {
		w, l, r := duel(t, 0)
		l.Orientation = -1
		l.Input.FaceLeft = true
		lin := Input{Attack: true, FaceLeft: true}
		duelStep(w, l, lin, r, swingLow)
		for range 20 {
			if l.State == StateClash || r.State == StateClash {
				break
			}
			duelStep(w, l, Input{FaceLeft: true}, r, Input{})
		}
		if l.State == StateClash || r.State == StateClash {
			t.Error("fighters back to back clashed")
		}
	})
}

// A third fighter's swing passes through a clash; the clash ends and the next
// one lands.
func TestAClashIsOutOfEveryoneElsesReach(t *testing.T) {
	w, l, r := duel(t, 0)
	c := w.Spawn(rl.Vector2{X: -duelGap}, true)
	c.Orientation = 1
	for range 10 {
		c.Input = Input{}
		duelStep(w, l, Input{}, r, Input{})
	}
	openLow(w, l)
	openLow(w, r)
	if !tryClash(l, r) {
		t.Fatal("no clash")
	}
	structure := l.Structure

	// c swings low at l, which guards low: a swing that would throw it into a
	// block at once, and hit it if it did not
	c.Input = Input{Attack: true}
	for range 20 {
		run(w, 1)
		c.Input = Input{}
		if l.State != StateClash {
			break
		}
		if l.Structure != structure {
			t.Fatal("a swing landed on a clashing fighter")
		}
	}
	if l.State == StateBlock {
		t.Fatal("a swing threw a clashing fighter into a block")
	}

	// and once the clash is over, a swing lands again
	for l.State == StateClash {
		run(w, 1)
	}
	if l.State != StateIdle || r.State != StateIdle {
		t.Fatalf("after the clash: states %d %d", l.State, r.State)
	}
	for range 30 {
		c.Input = Input{Attack: true}
		l.Input = Input{Updown: 1} // guarding mid now: a low swing lands
		run(w, 1)
		if l.Structure < structure {
			return
		}
	}
	t.Error("a swing after the clash did not land")
}

// A fighter standing in a clashing one neither shoves it nor is shoved by it.
func TestAClashIsNotShoved(t *testing.T) {
	w, l, r := duel(t, 0)
	c := w.Spawn(l.Pos, true)
	c.Orientation, c.OnGround = 1, true
	openLow(w, l)
	openLow(w, r)
	if !tryClash(l, r) {
		t.Fatal("no clash")
	}
	c.Pos = l.Pos // standing in it
	c.Update(float32(FixedTimeStep))
	if c.State == StatePush || l.State != StateClash {
		t.Errorf("the bystander shoved a clashing fighter: states %d %d", c.State, l.State)
	}
	l.Update(float32(FixedTimeStep))
	if l.State != StateClash {
		t.Errorf("the clashing fighter left the clash for state %d", l.State)
	}
}

// Both leave the clash on the same tick, whichever is updated first.
func TestAClashEndsTogether(t *testing.T) {
	for _, swap := range []bool{false, true} {
		w, l, r := duel(t, 0)
		if swap {
			w.Actors[0], w.Actors[1] = w.Actors[1], w.Actors[0]
		}
		duelStep(w, l, swingLow, r, swingLow)
		if !clashed(w, l, r) {
			t.Fatal("no clash")
		}
		for range 120 {
			duelStep(w, l, Input{}, r, Input{})
			if (l.State == StateClash) != (r.State == StateClash) {
				t.Fatalf("swapped %v: one left the clash before the other", swap)
			}
			if l.State != StateClash {
				break
			}
		}
		if l.State == StateClash {
			t.Fatal("the clash never ended")
		}
	}
}

// Online: the server puts the player's fighter into the clash, and the client
// shows it once the snapshot arrives - the server's structure and energy, and
// no swing out of it from the commands replayed on top.
func TestAClashReachesTheClient(t *testing.T) {
	Stage = StageLevel
	defer func() { Stage = StageMainMenu }()
	w, l, r := duel(t, 0)

	c := &Client{W: NewWorld(false), ActorID: l.ID}
	c.reconcile(recOf(l), 0)
	local := c.W.Local

	// the client predicts its own swing; the server runs it with the answer
	cmd := Input{Seq: 1, Attack: true}
	c.Seq = 1
	c.Ring[1%CmdRing] = cmd
	local.Input = cmd
	local.Update(float32(FixedTimeStep))
	if local.State != StateAttack {
		t.Fatalf("the client did not predict its swing")
	}
	duelStep(w, l, swingLow, r, swingLow)
	if !clashed(w, l, r) {
		t.Fatal("no clash on the server")
	}

	// two more commands the server has not seen, both asking to swing
	for seq := uint32(2); seq <= 3; seq++ {
		cmd := Input{Seq: seq, Attack: true}
		c.Seq = seq
		c.Ring[seq%CmdRing] = cmd
		local.Input = cmd
		local.Update(float32(FixedTimeStep))
	}
	c.reconcile(recOf(l), 1)
	if local.State != StateClash {
		t.Fatalf("the client's fighter is in state %d, want the clash", local.State)
	}
	if local.Structure != l.Structure || local.Energy != l.Energy {
		t.Errorf("client has %v structure %d energy, server %v %d", local.Structure, local.Energy, l.Structure, l.Energy)
	}
}
