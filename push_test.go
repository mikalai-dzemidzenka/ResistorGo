package resistor

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// shoveTicks is how long the shove takes to land: the frame both arms come out,
// at the rate the animation runs, and a tick either side of it.
const shoveTicks = int(60*PushSeconds/PushFrames*PushFrame) + 2

// pushTicks is the whole shove, and a tick past its end.
const pushTicks = int(60*PushSeconds) + 2

// crowd puts two fighters on a floor, settled on their feet: a facing o, and b
// dx behind or in front of it, facing a. The facing is held the way a player
// holds it, by where the mouse is.
func crowd(t *testing.T, dx float32, o int8) (*World, *Actor, *Actor) {
	t.Helper()
	floor()
	w := NewWorld(true)
	a := w.Spawn(rl.Vector2{}, true)
	b := w.Spawn(rl.Vector2{X: dx}, true)
	a.Orientation, a.Input.FaceLeft = o, o == -1
	bo := facing(-dx)
	b.Orientation, b.Input.FaceLeft = bo, bo == -1
	for _, f := range []*Actor{a, b} {
		f.checkCollision()
		if !f.OnGround {
			t.Fatalf("the fighter at %v is not standing on the floor", f.Pos.X)
		}
	}
	return w, a, b
}

// overlaps is every offset at which two fighters' collision boxes overlap: in
// front, behind, and on the same column.
func overlaps() []float32 {
	var dxs []float32
	for dx := float32(-13); dx <= 13; dx++ {
		dxs = append(dxs, dx)
	}
	return dxs
}

// away is the way a shove from a fighter facing o sends one dx from it.
func away(dx float32, o int8) int8 {
	if dx == 0 {
		return o
	}
	return facing(dx)
}

// Any overlap is crowding, in front or behind, and nothing short of one is.
func TestAnyOverlapIsCrowding(t *testing.T) {
	for _, o := range []int8{1, -1} {
		for dx := float32(-20); dx <= 20; dx++ {
			_, a, b := crowd(t, dx, o)
			want := abs(dx) < 14
			if got := a.crowding() == b; got != want {
				t.Errorf("facing %d, dx=%v: crowding %t, want %t", o, dx, got, want)
			}
		}
	}
}

// PushDistance is bounded both ways: from anywhere two fighters overlap, the
// throw clears the overlap and stays inside the reach of a swing.
func TestPushDistanceLandsAtArmsLength(t *testing.T) {
	for _, o := range []int8{1, -1} {
		for _, dx := range overlaps() {
			d := away(dx, o)
			a := NewActor(rl.Vector2{}, true)
			b := NewActor(rl.Vector2{X: dx + float32(d)*PushDistance}, true)
			if aabb(a.collisionRect(), b.collisionRect()) {
				t.Errorf("facing %d, from dx=%v: still overlapping at %v", o, dx, b.Pos.X)
			}
			if !a.swingReachFacing(b, d) {
				t.Errorf("facing %d, from dx=%v: thrown to %v, out of reach", o, dx, b.Pos.X)
			}
		}
	}
}

// Two fighters overlapping, in front or behind: the first free to act turns to
// the other and shoves it, and the other - stunned by it - does not shove back.
func TestEveryOverlapIsShoved(t *testing.T) {
	for _, o := range []int8{1, -1} {
		for _, dx := range overlaps() {
			w, a, b := crowd(t, dx, o)
			run(w, 1)
			if a.State != StatePush {
				t.Errorf("facing %d, dx=%v: the crowded fighter is in state %d, want the shove %d", o, dx, a.State, StatePush)
				continue
			}
			if want := away(dx, o); a.Orientation != want {
				t.Errorf("facing %d, dx=%v: the shover faces %d, want it facing the one it shoves, %d", o, dx, a.Orientation, want)
			}
			if a.Anim != "resources/threefour/push.png" {
				t.Errorf("the shove plays %q, want push.png", a.Anim)
			}
			if b.State == StatePush {
				t.Errorf("facing %d, dx=%v: the other shoved back", o, dx)
			}
		}
	}
}

// The shove lands: the crowder goes away from the shover, clear of it and
// inside its reach - from anywhere two fighters overlap, both ways.
func TestAShoveEndsAtArmsLength(t *testing.T) {
	for _, o := range []int8{1, -1} {
		for _, dx := range overlaps() {
			w, a, b := crowd(t, dx, o)
			run(w, pushTicks)
			got := b.Pos.X - a.Pos.X
			if facing(got) != away(dx, o) || abs(got) <= abs(dx) {
				t.Errorf("facing %d, dx=%v: the crowder ended at %v, not further off on its own side", o, dx, got)
			}
			if aabb(a.collisionRect(), b.collisionRect()) {
				t.Errorf("facing %d, dx=%v: still overlapping after the shove, at %v", o, dx, got)
			}
			if !a.swingReachFacing(b, facing(got)) {
				t.Errorf("facing %d, dx=%v: shoved to %v, out of reach", o, dx, got)
			}
			if a.State == StatePush {
				t.Errorf("facing %d, dx=%v: the shove never ended", o, dx)
			}
		}
	}
}

// A fighter on its feet and free shoves: standing, walking or running.
func TestAFreeFighterShoves(t *testing.T) {
	for _, c := range []struct {
		name string
		set  func(a *Actor)
	}{
		{"standing", func(a *Actor) {}},
		{"walking", func(a *Actor) { a.Input.Left = true }}, // Left moves right
		{"running", func(a *Actor) { a.State, a.Speed.X = StateRun, RunReach+10 }},
	} {
		w, a, _ := crowd(t, -8, 1)
		c.set(a)
		run(w, 1)
		if a.State != StatePush {
			t.Errorf("%s: the fighter is in state %d, want the shove", c.name, a.State)
		}
	}
}

// A fighter already committed finishes what it is doing first - a swing, a
// block, a stun - and one in the air, or guarding everything, does not shove.
func TestABusyFighterDoesNotShove(t *testing.T) {
	for _, c := range []struct {
		name string
		set  func(a, b *Actor)
	}{
		{"mid-swing", func(a, b *Actor) {
			a.Updown, a.Input.Updown = 1, 1
			a.State = StateAttack
			a.SetAnim("resources/threefour/attack-mid.png", AnimOptions{animDuration: 5, totalFrames: 8, once: true})
		}},
		{"blocking", func(a, b *Actor) { a.enterBlockState(b) }},
		{"stunned", func(a, b *Actor) { a.enterShovedState(b) }},
		{"flinching", func(a, b *Actor) { a.FlinchTicks = FlinchHoldTicks }},
		{"in the air", func(a, b *Actor) { a.OnGround, a.Speed.Y, a.State = false, -10, StateJump }},
		{"guarding everything", func(a, b *Actor) { a.Updown, a.PrevUpdown, a.Input.Updown = 3, 3, 3 }},
	} {
		w, a, b := crowd(t, -8, 1)
		b.Updown, b.PrevUpdown, b.Input.Updown = 3, 3, 3 // so only a could shove
		c.set(a, b)
		run(w, 1)
		if a.State == StatePush {
			t.Errorf("%s: the fighter shoved", c.name)
		}
	}
}

// A fighter in the air is not shoved: it lands first. Shoving a climber out of
// the air knocked it off the same ledge every time it jumped for it.
func TestAFighterInTheAirIsNotShoved(t *testing.T) {
	w, a, b := crowd(t, -8, 1)
	b.OnGround, b.Speed.Y, b.State = false, -200, StateJump
	run(w, 1)
	if a.State == StatePush {
		t.Error("a fighter shoved one in the air")
	}
}

// The shove stuns the crowder and does it no damage: the lockout and the
// animation of a stun, and no structure, no energy, no hit counted, no flash
// and no hit-stop.
func TestAShoveStunsAndDoesNoDamage(t *testing.T) {
	w, a, b := crowd(t, -8, 1)
	structure, energy := b.Structure, b.Energy
	run(w, 1)
	if a.State != StatePush {
		t.Fatalf("the crowded fighter is in state %d, want the shove", a.State)
	}
	if b.State != StateAttacked || b.Anim != "resources/threefour/hit-top.png" {
		t.Errorf("the crowder is in state %d playing %q, want stunned", b.State, b.Anim)
	}
	if b.Attacker != a {
		t.Error("the stun does not name the shover as what stunned it")
	}
	if b.FreezeTicks != 0 || a.FreezeTicks != 0 {
		t.Errorf("the shove froze the fight: shover %d, crowder %d ticks of hit-stop", a.FreezeTicks, b.FreezeTicks)
	}
	if b.Events&EvHit != 0 || b.AnimWhite {
		t.Error("the shove was shown as a hit")
	}
	run(w, pushTicks)
	if b.Structure != structure {
		t.Errorf("the shove cost the fighter shoved %v structure, want nothing", structure-b.Structure)
	}
	if b.Energy != energy {
		t.Errorf("the shove moved the shoved fighter's energy by %v, want nothing", b.Energy-energy)
	}
	if b.HitCount != 0 || b.HitRun != 0 || b.FlinchTicks != 0 {
		t.Error("the shove was counted as a hit")
	}
	if b.State == StateAttacked {
		t.Error("the stun never ended")
	}
}

// A crowder mid-swing is stunned out of it the moment the shove starts, so its
// swing cannot land on the shover's guard during the wind-up and cut the shove
// off before it throws anyone - the way a bot, whose lunges are how it walks
// in, was cutting off nearly every one.
func TestASwingingCrowderCannotCutTheShoveOff(t *testing.T) {
	w, a, b := crowd(t, -8, 1)
	// the same guard height, so a swing that reached would put the shover on guard
	a.Updown, a.PrevUpdown, a.Input.Updown = 1, 1, 1
	b.Updown, b.Input.Updown = 1, 1
	b.State = StateAttack
	b.SetAnim("resources/threefour/attack-mid.png", AnimOptions{animDuration: 0.3, totalFrames: 8, once: true})
	b.AnimStartTick = b.Ticks - 12 // just short of its hit frame
	from := b.Pos.X

	run(w, 1)
	if a.State != StatePush {
		t.Fatalf("the crowded fighter is in state %d, want the shove", a.State)
	}
	for range pushTicks {
		if a.State != StatePush {
			break
		}
		run(w, 1)
	}
	if a.State != StateIdle || !a.PushDone {
		t.Errorf("the shove was cut off: the shover ended in state %d, landed=%t", a.State, a.PushDone)
	}
	if b.Pos.X >= from {
		t.Errorf("the swinging crowder was not moved away: %v from %v", b.Pos.X, from)
	}
	if a.Structure != MaxStructure {
		t.Errorf("the crowder's swing landed through the shove: the shover lost %v", MaxStructure-a.Structure)
	}
}

// A client moves nobody and stuns nobody: both come from the snapshot.
func TestAClientDoesNotStun(t *testing.T) {
	floor()
	w := NewWorld(false)
	a := w.Spawn(rl.Vector2{}, true)
	b := w.Spawn(rl.Vector2{X: -8}, true)
	a.Orientation = 1
	b.Orientation = 1
	for _, f := range []*Actor{a, b} {
		f.checkCollision()
	}
	run(w, shoveTicks)
	if b.State == StateAttacked {
		t.Error("a client stunned a fighter with a shove of its own")
	}
}

// The dead are out of the fight, and are not shoved out of the way either.
func TestTheDeadAreNotShoved(t *testing.T) {
	w, a, b := crowd(t, -8, 1)
	b.State = StateDeath
	b.SetAnim("resources/threefour/death_fall.png", AnimOptions{animDuration: 60, totalFrames: 9, once: true})
	from := b.Pos.X
	run(w, shoveTicks)
	if a.State == StatePush {
		t.Error("a fighter shoved a body")
	}
	if b.Pos.X != from {
		t.Errorf("the body was moved to %v from %v", b.Pos.X, from)
	}
}

// The shove cannot be steered, and ends on its own animation.
func TestAShoveCannotBeSteered(t *testing.T) {
	w, a, _ := crowd(t, -8, 1)
	run(w, 1)
	if a.State != StatePush {
		t.Fatalf("the crowded fighter is in state %d, want the shove", a.State)
	}
	a.Input.Right = true
	from, facingNow := a.Pos.X, a.Orientation
	a.Input.FaceLeft = !a.Input.FaceLeft // the mouse flips sides mid-shove
	for range pushTicks - 3 {
		run(w, 1)
		if a.State != StatePush {
			break
		}
		if a.Pos.X != from {
			t.Fatalf("the shover moved under input, %v to %v", from, a.Pos.X)
		}
		if a.Orientation != facingNow {
			t.Fatalf("the shover turned under input mid-shove")
		}
	}
	a.Input.Right = false
	run(w, 3)
	if a.State != StateIdle {
		t.Errorf("the shove ended in state %d, want standing", a.State)
	}
}

// A shove moves each fighter at most once.
func TestAShoveLandsOnce(t *testing.T) {
	w, a, b := crowd(t, -8, 1)
	moved := 0
	last := b.Pos.X
	for range pushTicks {
		run(w, 1)
		if b.Pos.X != last && a.State == StatePush {
			moved++
		}
		last = b.Pos.X
	}
	if moved != 1 {
		t.Errorf("the shove moved the crowder %d times, want once", moved)
	}
}

// The shove is the authority's: a client is shown the ground it changed by the
// snapshot, and moves nobody itself.
func TestAClientDoesNotShove(t *testing.T) {
	floor()
	w := NewWorld(false) // a predicting client's world
	a := w.Spawn(rl.Vector2{}, true)
	b := w.Spawn(rl.Vector2{X: -8}, true)
	a.Orientation = 1
	b.Orientation, b.Input.FaceLeft = 1, false
	for _, f := range []*Actor{a, b} {
		f.checkCollision()
	}
	from := b.Pos.X
	shoves, was := 0, a.State
	for range 3 * pushTicks {
		run(w, 1)
		if a.State == StatePush && was != StatePush {
			shoves++
		}
		was = a.State
	}
	if b.Pos.X != from {
		t.Errorf("a client moved a fighter with a shove of its own: %v to %v", from, b.Pos.X)
	}
	// nobody moved, so the two still overlap: the snapshot is what parts them,
	// and until it does the client does not throw the shove again
	if shoves != 1 {
		t.Errorf("the client threw %d shoves at a fighter it cannot move, want 1", shoves)
	}
}

// A hit before the shove lands answers it: the shover reacts to the hit, and
// the shove is not thrown - a stun and a flinch alike.
func TestAHitInterruptsTheShove(t *testing.T) {
	for _, stun := range []bool{false, true} {
		w, a, b := crowd(t, -8, 1)
		a.Updown, a.PrevUpdown, a.Input.Updown = 0, 0, 0
		run(w, 1)
		if a.State != StatePush {
			t.Fatalf("the crowded fighter is in state %d, want the shove", a.State)
		}
		if stun {
			a.HitCount = StunEveryHits - 1
		}
		a.Input.FaceLeft = true // the player looks where the shove turned it
		b.Updown = 1
		a.takeHit(b)
		from := b.Pos.X
		run(w, shoveTicks)
		if a.State == StatePush {
			t.Errorf("stun=%t: still shoving after the hit", stun)
		}
		if stun && a.State != StateAttacked {
			t.Errorf("stun=%t: the shover is in state %d, want stunned", stun, a.State)
		}
		if b.Pos.X != from {
			t.Errorf("stun=%t: the shove landed after the hit, moving the crowder %v to %v", stun, from, b.Pos.X)
		}
	}
}

// walls() raises a wall a fighter can be walked into and stopped by, flush.
func TestAFighterWalkedIntoAWallRestsFlush(t *testing.T) {
	walls()
	defer floor()
	w := NewWorld(true)
	a := w.Spawn(rl.Vector2{X: 0}, true)
	a.Input.Right, a.Input.FaceLeft = true, true // Right walks left
	run(w, 120)
	if left := a.Pos.X + a.CollisionBox.X; left != wallX {
		t.Errorf("the fighter came to rest with its left edge at %v, want flush on the wall at %v", left, wallX)
	}
	if !a.OnGround {
		t.Error("the fighter walked into the wall is not standing")
	}
}

// A crowder flush against a wall cannot be thrown: the shove moves it nowhere.
// It must not be thrown again from the same spot, and the shover must get its
// input back - a bot, which faces whatever it turned to, and a player whose
// mouse still points away from the crowder.
func TestAWallStopsTheThrow(t *testing.T) {
	for _, player := range []bool{false, true} {
		walls()
		w := NewWorld(true)
		b := w.Spawn(rl.Vector2{}, true)
		b.Pos.X = wallX - b.CollisionBox.X // flush on the wall
		b.Orientation, b.Input.FaceLeft = 1, false
		// guarding everything, so it never shoves back and parts the two itself:
		// what is measured is that the shover does not throw the same shove again
		b.Updown, b.PrevUpdown, b.Input.Updown = 3, 3, 3
		a := w.Spawn(rl.Vector2{X: b.Pos.X + 8}, player)
		a.Orientation, a.Input.FaceLeft = 1, false // looking away from the crowder
		for _, f := range []*Actor{a, b} {
			f.checkCollision()
		}
		if a.crowding() != b {
			t.Fatalf("player=%t: the fighter at the wall is not crowding", player)
		}

		shoves, was := 0, a.State
		for range 3 * pushTicks {
			run(w, 1)
			if a.State == StatePush && was != StatePush {
				shoves++
			}
			was = a.State
		}
		if shoves != 1 {
			t.Errorf("player=%t: %d shoves from the same spot, want 1", player, shoves)
		}
		if b.Speed.X < 0 {
			t.Errorf("player=%t: the crowder is left with speed %v into the wall", player, b.Speed.X)
		}

		a.Input.Attack = true
		run(w, 2)
		if a.State != StateAttack {
			t.Errorf("player=%t: the shover did not answer the attack button, state %d", player, a.State)
		}
	}
	floor()
}
