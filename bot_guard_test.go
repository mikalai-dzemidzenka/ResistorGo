package resistor

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// hitPair builds an authoritative world with an attacker and a defender in
// reach. isBot decides which kind of fighter the defender is, which is the
// whole question here: swing() in structure_test.go spawns both as players.
func hitPair(t *testing.T, isBot bool) (*Actor, *Actor) {
	t.Helper()
	floor()
	w := NewWorld(true)
	att := w.Spawn(rl.Vector2{}, true)
	def := w.Spawn(rl.Vector2{X: 20}, !isBot)
	att.Orientation = 1
	return att, def
}

// freshHit puts the defender back to a fighter that can take another hit
// without dying or stunning, so a loop measures one hit at a time.
func freshHit(def *Actor, guard uint8) {
	def.Structure = MaxStructure
	def.HitCount, def.HitRun = 0, 0
	def.State = StateIdle
	def.Updown, def.PrevUpdown = guard, guard
}

// The odds are not seeded, so this asserts what the spec promises - that the
// guard both moves and stays, at roughly one hit in three - rather than a
// particular draw from math/rand.
func TestBotGuardAdaptsToWhereItWasHit(t *testing.T) {
	att, def := hitPair(t, true)
	att.Updown = 2 // hitting high
	const guard, n = 0, 1000

	moved := 0
	for range n {
		freshHit(def, guard)
		def.takeHit(att)
		switch def.Updown {
		case att.Updown:
			moved++
			if def.PrevUpdown != guard {
				t.Fatalf("after adapting, PrevUpdown is %d, want the guard it had (%d)", def.PrevUpdown, guard)
			}
		case guard:
			// kept its guard, which is the other two thirds
		default:
			t.Fatalf("the guard became %d, which is neither the old one (%d) nor the attacker's (%d)", def.Updown, guard, att.Updown)
		}
	}

	if moved == 0 {
		t.Errorf("the guard never moved in %d hits; a bot must sometimes answer where it was hit", n)
	}
	if moved == n {
		t.Errorf("the guard moved on every one of %d hits; a bot must stay readable", n)
	}
	// wide enough that math/rand cannot fail it, tight enough to catch the
	// odds being changed out from under the spec
	if rate := float64(moved) / n; rate < 0.27 || rate > 0.40 {
		t.Errorf("the guard moved on %.0f%% of hits, want about a third", rate*100)
	}
}

// A player's guard follows their aim and is never taken, on any tick.
func TestPlayerGuardIsNeverAdapted(t *testing.T) {
	att, def := hitPair(t, false)
	att.Updown = 2
	const guard = 0

	for i := range 1000 {
		freshHit(def, guard)
		def.takeHit(att)
		if def.Updown != guard {
			t.Fatalf("hit %d moved a player's guard from %d to %d", i+1, guard, def.Updown)
		}
	}
}

// The hit that ends a fighter teaches it nothing.
func TestKillingHitAdaptsNothing(t *testing.T) {
	att, def := hitPair(t, true)
	att.Updown = 2
	const guard = 0

	for i := range 200 {
		freshHit(def, guard)
		def.Structure = 0.5 // the next hit breaks it
		def.takeHit(att)
		if def.State != StateDeath {
			t.Fatalf("hit %d did not kill: state %d", i+1, def.State)
		}
		if def.Updown != guard {
			t.Fatalf("the killing hit moved the guard from %d to %d", guard, def.Updown)
		}
	}
}

// A stun adapts on the same terms as a flinch: the guard the bot wakes up with
// is the one the hit taught it.
func TestStunAdaptsLikeAFlinch(t *testing.T) {
	att, def := hitPair(t, true)
	att.Updown = 2
	const guard, n = 0, 1000

	moved, stuns := 0, 0
	for range n {
		freshHit(def, guard)
		def.HitCount = StunEveryHits - 1 // the next hit is the stunning one
		def.takeHit(att)
		if def.State != StateAttacked {
			t.Fatalf("the hit did not stun: state %d", def.State)
		}
		stuns++
		if def.Updown == att.Updown {
			moved++
		}
	}

	if stuns != n {
		t.Fatalf("only %d of %d hits stunned", stuns, n)
	}
	if moved == 0 {
		t.Errorf("the guard never moved on a stunning hit in %d tries", n)
	}
	if rate := float64(moved) / n; rate < 0.27 || rate > 0.40 {
		t.Errorf("a stun moved the guard on %.0f%% of hits, want the same about-a-third as a flinch", rate*100)
	}
}

// A body is not hit, so it cannot adapt its guard: a swing through a bot lying
// dead in reach lands nothing and leaves its guard and structure where they
// were. The odds are one in three, so thirty swings would move a guard that
// could be moved. The live case first, so a miss cannot pass for a pass.
func TestABodyDoesNotAdaptItsGuard(t *testing.T) {
	w, att, def := swing(t, 0)
	def.IsPlayer = false
	run(w, 1)
	if !att.SwingHit {
		t.Fatal("the swing did not land on a live bot in reach")
	}

	for range 30 {
		w, att, def := swing(t, 0)
		def.IsPlayer = false
		def.State = StateDeath
		structure := def.Structure
		run(w, 1)
		if att.SwingHit || def.Updown != 0 || def.Structure != structure {
			t.Fatalf("a swing through a body: hit %t, guard %d, structure %v; want no hit, guard 0, structure %v",
				att.SwingHit, def.Updown, def.Structure, structure)
		}
	}
}
