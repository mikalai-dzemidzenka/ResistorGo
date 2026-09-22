package resistor

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// walker is a fighter standing alone on the floor at x, facing +X. It is not a
// player, so nothing turns it toward the mouse: only its own input moves it.
func walker(t *testing.T, x float32) (*World, *Actor) {
	t.Helper()
	floor()
	w := NewWorld(true)
	a := w.Spawn(rl.Vector2{X: x}, false)
	a.Orientation = 1
	for range 20 {
		a.Input = Input{}
		a.Update(float32(FixedTimeStep))
	}
	if !a.OnGround {
		t.Fatalf("the fighter did not land")
	}
	return w, a
}

// steps runs a alone for n ticks with the same input, hit-stop skipped.
func steps(a *Actor, n int, in Input) {
	for range n {
		a.Input = in
		a.FreezeTicks = 0
		a.Update(float32(FixedTimeStep))
	}
}

// A speed under a pixel a tick moves the fighter by what it describes.
func TestSlowSpeedsMove(t *testing.T) {
	_, a := walker(t, 0)
	a.Friction = 0
	x0 := a.Pos.X
	for range 60 {
		a.Speed.X = 20
		a.FlinchTicks = 2 // a speed given to it, so a released walk's stop leaves it be
		a.Input = Input{}
		a.Update(float32(FixedTimeStep))
	}
	if moved := a.Pos.X - x0; moved < 19 || moved > 21 {
		t.Errorf("a second at 20 px/s moved %v px, want 20", moved)
	}
}

// A walk covers the same ground whichever side of the origin it is on. The
// starts are chosen so neither walk leaves the test floor (x -528 to 288).
func TestTheSameWalkEverywhere(t *testing.T) {
	walk := func(x float32, in Input) float32 {
		_, a := walker(t, x)
		x0 := a.Pos.X
		steps(a, 60, in)
		return a.Pos.X - x0
	}
	for _, in := range []Input{{Left: true}, {Right: true}} {
		east, west := walk(50, in), walk(-200, in)
		if abs(east-west) > 0.01 {
			t.Errorf("input %+v: a second's walk covered %v px at x=50 and %v px at x=-200", in, east, west)
		}
	}
}

// A released walk stops once it is under WalkStopSpeed, short of where
// friction alone would have let it slide.
func TestAReleasedWalkStops(t *testing.T) {
	_, a := walker(t, -400)         // room for the walk and the slide on the floor
	steps(a, 60, Input{Left: true}) // Left moves +X in this code
	v := a.Speed.X
	if v < WalkStopSpeed {
		t.Fatalf("the walk only reached %v px/s", v)
	}
	x0 := a.Pos.X
	last := v
	for range 120 {
		steps(a, 1, Input{})
		if a.Speed.X == 0 {
			break
		}
		last = a.Speed.X
	}
	if a.Speed.X != 0 {
		t.Fatalf("the walk never stopped")
	}
	if last-a.Friction*float32(FixedTimeStep) < 0 {
		t.Errorf("the walk ran out under friction (last speed %v) instead of being stopped", last)
	}
	// the same slide under friction alone, speed applied after friction as Update does
	glide := float32(0)
	for s := v; s > 0; {
		s = max(s-a.Friction*float32(FixedTimeStep), 0)
		glide += s * float32(FixedTimeStep)
	}
	slid := a.Pos.X - x0
	if short := glide - slid; short < 5 || short > 9 {
		t.Errorf("the walk slid %v px against %v px under friction alone: %v px short, want about 7", slid, glide, short)
	}
}

// drawPos is on whole pixels whatever the position's fraction.
func TestDrawnOnWholePixels(t *testing.T) {
	_, a := walker(t, 0)
	a.InterpolatedPos = rl.Vector2{X: 10.4, Y: 3.6}
	if p := a.drawPos(); p.X != 10 || p.Y != 4 {
		t.Errorf("drawn at %v, want whole pixels (10, 4)", p)
	}
}

// knockedBack is how far a standing fighter travels from a hit given by an
// attacker facing +X, with no key held, once it has come to rest.
func knockedBack(t *testing.T, hit func(def, att *Actor)) float32 {
	t.Helper()
	w, def := walker(t, 0)
	att := w.Spawn(rl.Vector2{X: -40}, false)
	att.Orientation = 1
	x0 := def.Pos.X
	hit(def, att)
	for range 120 {
		steps(def, 1, Input{})
		if def.Speed.X == 0 && def.State != StateAttacked {
			break
		}
	}
	return def.Pos.X - x0
}

// A flinch moves the fighter, and a stun moves it clearly further. The stun's
// throw stays about what it was measured at with whole pixels.
func TestKnockbacksMove(t *testing.T) {
	flinch := knockedBack(t, func(def, att *Actor) { def.enterFlinchState(att) })
	stun := knockedBack(t, func(def, att *Actor) { def.enterAttackedState(att) })
	t.Logf("flinch %v px, stun %v px", flinch, stun)
	if flinch < 1 {
		t.Errorf("a flinch moved the fighter %v px, want it moved", flinch)
	}
	if stun < 2*flinch {
		t.Errorf("a stun threw %v px against a flinch's %v: not clearly further", stun, flinch)
	}
	if stun < 18 || stun > 22 {
		t.Errorf("a stun threw %v px, want about 20", stun)
	}
}

// swingAlone starts a low swing for a fighter with nobody in reach.
func swingAlone(t *testing.T, a *Actor) {
	t.Helper()
	a.Updown = 0
	if !a.enterAttackState() {
		t.Fatal("could not start a swing")
	}
}

// Through a whole low swing a held key moves the fighter slowly either way,
// and never turns it.
func TestASwingingFighterCreeps(t *testing.T) {
	for _, in := range []Input{{Left: true}, {Right: true}} { // Left moves +X
		_, a := walker(t, 0)
		swingAlone(t, a)
		x0, facing0 := a.Pos.X, a.Orientation
		for a.State == StateAttack {
			steps(a, 1, in)
			if a.State == StateAttack && a.Orientation != facing0 {
				t.Fatalf("input %+v turned the swinging fighter", in)
			}
			if abs(a.Speed.X) > AttackMoveSpeed {
				t.Fatalf("a swinging fighter moved at %v px/s", a.Speed.X)
			}
		}
		moved := a.Pos.X - x0
		want := float32(1)
		if in.Right {
			want = -1
		}
		if moved*want <= 0 {
			t.Errorf("input %+v moved the swinging fighter %v px, want it moved that way", in, moved)
		}
		// the swing and its recovery together are about half a second
		if abs(moved) > 12 {
			t.Errorf("input %+v moved the swinging fighter %v px through a swing, want only a few", in, moved)
		}
	}
}

// Over the low blow alone - 0.3 s - the creep is at most 6 px.
func TestTheCreepIsVerySlow(t *testing.T) {
	_, a := walker(t, 0)
	swingAlone(t, a)
	x0 := a.Pos.X
	steps(a, 18, Input{Left: true})
	if moved := a.Pos.X - x0; moved > 6 {
		t.Errorf("a low swing with a key held moved %v px, want at most 6", moved)
	}
}

// Guarding everything pins a swinging fighter, as it pins a standing one.
func TestBlockAllPinsASwing(t *testing.T) {
	_, a := walker(t, 0)
	swingAlone(t, a)
	a.Updown = 3
	x0 := a.Pos.X
	steps(a, 10, Input{Left: true, Updown: 3})
	if a.Pos.X != x0 {
		t.Errorf("a fighter guarding everything moved %v px mid-swing", a.Pos.X-x0)
	}
}

// Distance decides a swing: a fighter already swinging, guarding another
// height, 2 px inside an opponent's low reach, is missed if it backs away and
// hit if it holds no key.
func TestDistanceDecidesASwing(t *testing.T) {
	hit := func(back bool) bool {
		floor()
		w := NewWorld(true)
		att := w.Spawn(rl.Vector2{}, false)
		def := w.Spawn(rl.Vector2{X: 200}, false)
		att.Team, def.Team = NoTeam, NoTeam
		att.Orientation, def.Orientation = 1, -1
		for range 20 {
			run(w, 1)
		}
		att.Updown = 0
		if !att.enterAttackState() {
			t.Fatal("the attacker could not swing")
		}
		// the far edge of the low box, measured, then 2 px inside it
		edge := float32(0)
		for dx := float32(10); dx < 80; dx++ {
			def.Pos.X = att.Pos.X + dx
			if box, _ := att.attackBoxAt(0, att.Pos); aabb(box, def.collisionRect()) {
				edge = dx
			}
		}
		def.Pos.X = att.Pos.X + edge - 2
		// the defender's own mid swing is under way and has already landed its
		// one blow, so only the attacker's swing is in question
		def.Updown = 1
		def.State = StateAttack
		def.SetAnim("resources/threefour/attack-mid.png", AnimOptions{animDuration: 0.3, totalFrames: 8, once: true, reset: true})
		def.SwingHit = true
		for range 30 {
			att.Input = Input{}
			def.Input = Input{Updown: 1, Left: back} // Left moves +X: away from the attacker
			run(w, 1)
			if att.SwingHit || att.State != StateAttack {
				break
			}
		}
		return def.Structure < MaxStructure
	}
	if !hit(false) {
		t.Fatal("standing 2 px inside the reach, the defender was not hit: nothing to dodge")
	}
	if hit(true) {
		t.Error("the defender backed away mid-swing and was still hit")
	}
}

// A bot's steering moves it through its swing like a player's keys.
func TestABotCreepsToo(t *testing.T) {
	floor()
	w := NewWorld(true)
	p := w.Spawn(rl.Vector2{X: 80}, true)
	b := w.Spawn(rl.Vector2{}, false)
	b.Orientation = 1
	for range 20 {
		run(w, 1)
	}
	b.Bot.target, b.Bot.side = p, -1
	swingAlone(t, b)
	x0 := b.Pos.X
	steered := false
	for b.State == StateAttack {
		b.Input = botInput(b)
		steered = steered || b.Input.Left || b.Input.Right
		b.FreezeTicks = 0
		b.Update(float32(FixedTimeStep))
	}
	if !steered {
		t.Skip("the bot did not steer during its swing: nothing to measure")
	}
	if moved := b.Pos.X - x0; moved <= 0 || moved > 12 {
		t.Errorf("a bot steering toward a target out of reach moved %v px through its swing, want a slow creep toward it", moved)
	}
}
