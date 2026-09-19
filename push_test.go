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

// window is every offset at which a fighter facing o is crowded: behind it, on
// its body, out of reach of its swing.
func window(o int8) []float32 {
	var dxs []float32
	for dx := float32(3); dx <= 13; dx++ {
		dxs = append(dxs, -float32(o)*dx)
	}
	return dxs
}

// The crowding window is exactly the ground behind a fighter that its body
// covers and its swing does not: 3 to 13px behind, the same both ways.
func TestCrowdingIsBehindAndTheSameBothWays(t *testing.T) {
	for _, o := range []int8{1, -1} {
		a := NewActor(rl.Vector2{}, true)
		a.Orientation = o
		b := NewActor(rl.Vector2{}, true)
		var got []float32
		for dx := float32(-20); dx <= 20; dx++ {
			b.Pos.X = dx
			if aabb(a.collisionRect(), b.collisionRect()) && !a.swingReachFacing(b, o) {
				got = append(got, dx)
			}
		}
		want := window(o)
		if len(got) != len(want) {
			t.Fatalf("facing %d: crowded at %v, want %v", o, got, want)
		}
		for _, dx := range want {
			found := false
			for _, g := range got {
				found = found || g == dx
			}
			if !found {
				t.Errorf("facing %d: crowded at %v, want %v", o, got, want)
				break
			}
		}
	}
}

// PushDistance is bounded both ways: from anywhere in the window, the throw
// clears the overlap and stays inside the reach of a swing.
func TestPushDistanceLandsAtArmsLength(t *testing.T) {
	for _, o := range []int8{1, -1} {
		for _, dx := range window(o) {
			a := NewActor(rl.Vector2{}, true)
			b := NewActor(rl.Vector2{X: dx + float32(facing(dx))*PushDistance}, true)
			if aabb(a.collisionRect(), b.collisionRect()) {
				t.Errorf("facing %d, from dx=%v: still overlapping at %v", o, dx, b.Pos.X)
			}
			if !a.swingReachFacing(b, facing(dx)) {
				t.Errorf("facing %d, from dx=%v: thrown to %v, out of reach", o, dx, b.Pos.X)
			}
		}
	}
}

// A standing fighter with another on its back turns on it and shoves; the one
// behind, which can reach, does not.
func TestAFighterOnYourBackIsShoved(t *testing.T) {
	for _, o := range []int8{1, -1} {
		for _, dx := range window(o) {
			w, a, b := crowd(t, dx, o)
			if a.crowding() != b {
				t.Fatalf("facing %d, dx=%v: the fighter behind is not crowding", o, dx)
			}
			if b.crowding() != nil {
				t.Fatalf("facing %d, dx=%v: the fighter behind, which can reach, is crowded too", o, dx)
			}
			run(w, 1)
			if a.State != StatePush {
				t.Errorf("facing %d, dx=%v: the crowded fighter is in state %d, want the shove %d", o, dx, a.State, StatePush)
				continue
			}
			if a.Orientation != -o {
				t.Errorf("facing %d, dx=%v: the shover faces %d, want it turned to %d", o, dx, a.Orientation, -o)
			}
			if a.Anim != "resources/threefour/push.png" {
				t.Errorf("the shove plays %q, want push.png", a.Anim)
			}
			if b.State == StatePush {
				t.Errorf("facing %d, dx=%v: the fighter behind shoved as well", o, dx)
			}
		}
	}
}

// Overlapping in front is overlapping within reach: a fighter swings at that,
// it does not shove it.
func TestOverlappingInFrontDoesNotShove(t *testing.T) {
	for _, o := range []int8{1, -1} {
		for dx := float32(-13); dx <= 13; dx++ {
			w, a, b := crowd(t, dx, o)
			if !a.swingReachFacing(b, o) {
				continue // behind: the window
			}
			run(w, 1)
			if a.State == StatePush {
				t.Errorf("facing %d, dx=%v: shoved a fighter it could have hit", o, dx)
			}
		}
	}
}

// The shove lands: the crowder goes away from the shover, clear of it and in
// front of it, inside its reach - from anywhere in the window, both ways.
func TestAShoveEndsAtArmsLength(t *testing.T) {
	for _, o := range []int8{1, -1} {
		for _, dx := range window(o) {
			w, a, b := crowd(t, dx, o)
			run(w, pushTicks)
			got := b.Pos.X - a.Pos.X
			if facing(got) != facing(dx) || abs(got) <= abs(dx) {
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

// A running fighter has its follower behind it by definition, and a walking
// one too; only a fighter standing still shoves.
func TestAChaseDoesNotSetItOff(t *testing.T) {
	w, a, b := crowd(t, -8, 1)
	a.State = StateRun
	a.Speed.X = RunReach + 10
	run(w, 1)
	if a.State == StatePush {
		t.Error("a running fighter shoved its follower")
	}

	w, a, _ = crowd(t, -8, 1)
	a.Input.Left = true // walking: Left moves right, and the fighter faces it
	run(w, 1)
	if a.State == StatePush {
		t.Error("a walking fighter shoved its follower")
	}
	_ = b
}

// The shove comes off the feet, and a fighter guarding everything stands its
// ground.
func TestOnlyAStandingFighterShoves(t *testing.T) {
	w, a, _ := crowd(t, -8, 1)
	a.OnGround, a.Speed.Y = false, -10
	a.State = StateJump
	run(w, 1)
	if a.State == StatePush {
		t.Error("a fighter in the air threw a shove")
	}

	w, a, _ = crowd(t, -8, 1)
	a.Updown, a.PrevUpdown, a.Input.Updown = 3, 3, 3
	run(w, 1)
	if a.State == StatePush {
		t.Error("a fighter guarding everything shoved instead of standing its ground")
	}
}

// The shove costs nothing but ground: no structure, no energy, no hit.
func TestAShoveMovesGroundAndNothingElse(t *testing.T) {
	w, _, b := crowd(t, -8, 1)
	structure, energy := b.Structure, b.Energy
	run(w, pushTicks)
	if b.Structure != structure {
		t.Errorf("the shove cost the fighter shoved %v structure, want nothing", structure-b.Structure)
	}
	if b.Energy != energy {
		t.Errorf("the shove moved the shoved fighter's energy by %v, want nothing", b.Energy-energy)
	}
	if b.HitCount != 0 || b.State == StateAttacked || b.FlinchTicks != 0 {
		t.Error("the shove was counted as a hit")
	}
}

// A fighter mid-swing is moved too. Its own state holds its speed down to the
// lunge, so a shove that only pushed speed around would leave it where it was.
func TestAShoveMovesAFighterMidSwing(t *testing.T) {
	w, a, b := crowd(t, -8, 1)
	b.Updown = 1
	b.State = StateAttack
	b.SetAnim("resources/threefour/attack-mid.png", AnimOptions{animDuration: 5, totalFrames: 8, once: true})
	from := b.Pos.X
	a.Updown, a.PrevUpdown = 0, 0

	run(w, 1)
	if a.State != StatePush {
		t.Fatalf("the crowded fighter is in state %d, want the shove", a.State)
	}
	run(w, shoveTicks)
	if b.Pos.X >= from {
		t.Errorf("the fighter mid-swing was not moved away: %v from %v", b.Pos.X, from)
	}
	if b.State != StateAttack {
		t.Errorf("the shove put the swinging fighter into state %d, want its swing left alone", b.State)
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
