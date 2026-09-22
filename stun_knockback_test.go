package resistor

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// knocked builds an authoritative world and returns an attacker facing the
// given way plus a bot defender. The defender is a bot on purpose: the StateIdle
// arm turns an IsPlayer fighter to face the mouse, and a turn divides Speed.X by
// three, which would quietly eat the very thing these tests measure.
func knocked(t *testing.T, facing int8) (*World, *Actor) {
	t.Helper()
	floor()
	w := NewWorld(true)
	att := w.Spawn(rl.Vector2{}, true)
	att.Orientation = facing
	return w, att
}

// travel runs a fighter n ticks and reports how far it moved.
func travel(a *Actor, n int) float32 {
	start := a.Pos.X
	for range n {
		a.Update(float32(FixedTimeStep))
	}
	return a.Pos.X - start
}

// The spec has always said a stun throws a fighter back; the AttackDash cap on
// the StateAttacked arm meant it never moved at all.
//
// The distance is absolute rather than measured against a flinch. A flinch
// once travelled 0 px - PunchSpeed is 50, which is under a pixel a tick, and
// position used to be cut to whole pixels every tick - so "further than a
// flinch" would have passed on a one-pixel throw. This test must not be able
// to pass because of a broken flinch.
func TestStunCarriesTheFighterARealDistance(t *testing.T) {
	w, att := knocked(t, 1)
	stunned := w.Spawn(rl.Vector2{X: 200}, false)
	stunned.enterAttackedState(att)

	// well past the point friction drops the throw under the floor
	dist := abs(travel(stunned, HitStopTicks+60))
	if dist < 10 {
		t.Errorf("a stun carried the fighter %v px, want at least 10 - it should read as a throw, not a jitter", dist)
	}
}

// Measured as position, not speed: a speed assertion passes on a fighter that
// is given a knockback and then never moves, which is exactly the bug.
func TestStunActuallyMovesTheFighter(t *testing.T) {
	w, att := knocked(t, 1)
	def := w.Spawn(rl.Vector2{X: 200}, false)
	def.enterAttackedState(att)

	// through the freeze and a few live ticks
	moved := abs(travel(def, HitStopTicks+4))
	if moved <= 1 {
		t.Errorf("a stunned fighter moved %v px, which whole-pixel truncation cannot tell from standing still", moved)
	}
}

func TestStunThrowFollowsTheAttackersFacing(t *testing.T) {
	for _, c := range []struct {
		name   string
		facing int8
	}{{"right", 1}, {"left", -1}} {
		w, att := knocked(t, c.facing)
		def := w.Spawn(rl.Vector2{X: 200}, false)
		def.enterAttackedState(att)

		d := travel(def, HitStopTicks+4)
		if c.facing == 1 && d <= 0 {
			t.Errorf("%s: the fighter moved %v, want it carried the way the attacker faced", c.name, d)
		}
		if c.facing == -1 && d >= 0 {
			t.Errorf("%s: the fighter moved %v, want it carried the way the attacker faced", c.name, d)
		}
	}
}

// The other half: the fighter doing the swinging stays capped, now at the slow
// pace it may move while swinging. This is what says the right of the two
// identical lines was removed.
func TestAttackerIsStillCappedWhileSwinging(t *testing.T) {
	floor()
	w := NewWorld(true)
	a := w.Spawn(rl.Vector2{}, false) // alone: nothing for the swing to hit
	a.Updown = 1
	if !a.enterAttackState() {
		t.Fatalf("could not start a swing")
	}

	a.Speed.X = MaxSpeed
	a.Update(float32(FixedTimeStep))
	if a.State != StateAttack {
		t.Fatalf("the fighter left StateAttack: %d", a.State)
	}
	if abs(a.Speed.X) > AttackMoveSpeed {
		t.Errorf("an attacking fighter's speed is %v, want it held to the swing's cap %v", a.Speed.X, float32(AttackMoveSpeed))
	}
}
