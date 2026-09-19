package resistor

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// reachSpan sweeps a target past a fighter facing o and returns the range of
// offsets a swing covers, as the half-open interval [lo, hi). It asks the same
// rectangles the hit test asks, so what it measures is what reaches.
func reachSpan(t *testing.T, a *Actor, updown uint8, o int8) (lo, hi float32, found bool) {
	t.Helper()
	box, ok := a.attackBoxFacing(updown, a.Pos, o)
	if !ok {
		t.Fatalf("guard height %d reports no attack box", updown)
	}
	box.X += float32(o) * AttackDash
	target := NewActor(rl.Vector2{}, true)
	for dx := float32(-80); dx <= 80; dx++ {
		target.Pos = rl.Vector2{X: a.Pos.X + dx, Y: a.Pos.Y}
		if !aabb(box, target.collisionRect()) {
			continue
		}
		if !found {
			lo, found = dx, true
		}
		hi = dx + 1
	}
	return lo, hi, found
}

// A swing covers the same ground whichever way a fighter faces.
//
// It did not. The attack box is the front half of the 48px sprite frame,
// mirrored about the frame's centre at 24; the collision box every reach test
// measures it against is centred at 21. The 3px offset changed sign on
// turnaround, so a fighter reached 38px facing right and 32px facing left.
func TestSwingReachesTheSameBothWays(t *testing.T) {
	a := NewActor(rl.Vector2{}, true)
	for updown := uint8(0); updown <= 2; updown++ {
		rLo, rHi, ok := reachSpan(t, a, updown, 1)
		if !ok {
			t.Fatalf("guard height %d reaches nothing facing right", updown)
		}
		lLo, lHi, ok := reachSpan(t, a, updown, -1)
		if !ok {
			t.Fatalf("guard height %d reaches nothing facing left", updown)
		}
		// mirroring [lLo, lHi) about the fighter gives (-lHi, -lLo]
		if rLo != -lHi+1 || rHi != -lLo+1 {
			t.Errorf("guard height %d reaches [%v,%v) facing right and [%v,%v) facing left - not mirrored",
				updown, rLo, rHi, lLo, lHi)
		}
		if got := rHi - rLo; got != lHi-lLo {
			t.Errorf("guard height %d covers %v px facing right and %v px facing left",
				updown, got, lHi-lLo)
		}
	}
}

// A fighter standing on another can be hit. Two fighters on the same pixel is a
// position the fight has to be able to answer; it was reachable facing left and
// not facing right, purely because the right-facing box began exactly on the
// collision box's far edge, which aabb treats as clear.
func TestAFighterStandingOnAnotherIsInReach(t *testing.T) {
	a := NewActor(rl.Vector2{}, true)
	b := NewActor(rl.Vector2{}, true)
	for _, o := range []int8{1, -1} {
		box, ok := a.attackBoxFacing(1, a.Pos, o)
		if !ok {
			t.Fatal("mid guard reports no attack box")
		}
		box.X += float32(o) * AttackDash
		if !aabb(box, b.collisionRect()) {
			t.Errorf("facing %d, a fighter on the same spot is out of reach", o)
		}
	}
}

// A fighter you are standing in is one you can hit.
//
// This is what the correction buys the bot. inSwingRange measures the swing the
// bot would throw toward its target, and the target's own side decides which
// way that is; with the two facings reaching the same ground, every overlap is
// now inside one of them. tooClose - the bot's answer to being under a fighter
// rather than in front of it - therefore cannot fire any more. It was one pixel
// wide, at dx exactly 0, and that pixel was only unreachable because the
// right-facing box began on the collision box's far edge.
func TestOverlappingIsAlwaysInReach(t *testing.T) {
	floor()
	w := NewWorld(true)
	a := w.Spawn(ZeroVec2, true)
	b := w.Spawn(ZeroVec2, true)
	for _, o := range []int8{1, -1} {
		a.Orientation = o
		for dx := float32(-13); dx <= 13; dx++ {
			b.Pos = rl.Vector2{X: a.Pos.X + dx, Y: a.Pos.Y}
			if !aabb(a.collisionRect(), b.collisionRect()) {
				t.Fatalf("facing %d, dx=%v: the two are meant to overlap", o, dx)
			}
			if !a.inSwingRange(b) {
				t.Errorf("facing %d, dx=%v: a fighter inside the body is out of swing range", o, dx)
			}
			if a.tooClose(b) {
				t.Errorf("facing %d, dx=%v: tooClose still fires", o, dx)
			}
		}
	}
}
