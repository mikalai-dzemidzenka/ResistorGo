package resistor

import (
	"math"
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// ducked puts a standing fighter into a duck held for hold ticks.
func ducked(t *testing.T, x float32, hold int) (*World, *Actor) {
	t.Helper()
	w, a := walker(t, x)
	steps(a, 1, Input{Roll: true})
	if a.State != StateDuck {
		t.Fatalf("a standing fighter holding roll is in state %d, want the duck %d", a.State, StateDuck)
	}
	steps(a, hold-1, Input{Roll: true})
	return w, a
}

// rollOut releases a duck and runs the roll to its landing, reporting the
// peak (lowest Y) and where it touched down.
func rollOut(t *testing.T, a *Actor) (peak, landX float32) {
	t.Helper()
	peak = a.Pos.Y
	for range RollPathTicks {
		steps(a, 1, Input{})
		peak = min(peak, a.Pos.Y)
		if a.State == StateRollRecover {
			return peak, a.Pos.X
		}
		if a.State != StateRoll {
			t.Fatalf("mid-roll the fighter went to state %d", a.State)
		}
	}
	t.Fatal("the roll never landed")
	return
}

func TestRollLaunchGrowsAndCaps(t *testing.T) {
	a := &Actor{Orientation: 1, AnimStartTick: 100, Input: Input{RollAim: 255}}
	at := func(held uint32) rl.Vector2 {
		a.Ticks = a.AnimStartTick + held
		return a.rollLaunch()
	}
	speed := func(v rl.Vector2) float32 { return float32(math.Hypot(float64(v.X), float64(v.Y))) }
	if s := speed(at(0)); abs(s-RollMinSpeed) > 0.01 {
		t.Errorf("a tap launches at speed %v, want the minimum %d", s, RollMinSpeed)
	}
	short, long := at(15), at(45)
	if !(long.X > short.X && long.Y < short.Y) {
		t.Errorf("a 45-tick hold launches at %v, a 15-tick hold at %v: want the longer both farther and higher", long, short)
	}
	if capped := at(RollChargeTicks); at(180) != capped || at(600) != capped {
		t.Errorf("holds past the cap launch differently: %v %v %v", capped, at(180), at(600))
	}
	a.Orientation = -1
	if v := at(0); v.X >= 0 {
		t.Errorf("facing left the roll launches at %v, want it going left", v)
	}
}

// The same charge aimed higher goes higher and less far, and a flat launch at
// full charge is held to RollMaxX.
func TestRollAimSetsTheAngle(t *testing.T) {
	a := &Actor{Orientation: 1, AnimStartTick: 100, Ticks: 100 + RollChargeTicks}
	a.Input.RollAim = 0
	flat := a.rollLaunch()
	a.Input.RollAim = 255
	steep := a.rollLaunch()
	if !(steep.Y < flat.Y && steep.X < flat.X) {
		t.Errorf("aimed steep the roll launches at %v, flat at %v: want steep higher and shorter", steep, flat)
	}
	if flat.X > RollMaxX {
		t.Errorf("a flat launch goes %v, past RollMaxX %d", flat.X, RollMaxX)
	}
}

func TestRollAimFollowsTheCursor(t *testing.T) {
	from := rl.Vector2{X: 100, Y: 100}
	cases := []struct {
		name   string
		cursor rl.Vector2
		want   uint8
	}{
		{"level with the fighter", rl.Vector2{X: 200, Y: 100}, 0},
		{"below it", rl.Vector2{X: 200, Y: 200}, 0},
		{"straight up", rl.Vector2{X: 100, Y: 0}, 255},
		{"45 degrees, the middle of the range", rl.Vector2{X: 0, Y: 0}, 127},
	}
	for _, c := range cases {
		if got := rollAim(from, c.cursor); got != c.want {
			t.Errorf("%s: aim %d, want %d", c.name, got, c.want)
		}
	}
}

func TestDuckFromAStandAndARun(t *testing.T) {
	_, a := ducked(t, -400, 10)
	if a.Speed.X != 0 || a.Anim != "resources/threefour/duck.png" {
		t.Errorf("ducking: speed %v, anim %q", a.Speed.X, a.Anim)
	}
	x := a.Pos.X
	steps(a, 20, Input{Roll: true, Left: true})
	if a.Pos.X != x || a.State != StateDuck {
		t.Errorf("a ducking fighter holding a key moved from %v to %v (state %d)", x, a.Pos.X, a.State)
	}

	_, r := walker(t, -400)
	r.State = StateRun
	r.Speed.X = RunReach + 20
	steps(r, 1, Input{Roll: true, Left: true})
	if r.State != StateDuck || r.Speed.X != 0 {
		t.Errorf("a running fighter pressing roll is in state %d at speed %v, want ducked and stopped", r.State, r.Speed.X)
	}
}

func TestDuckFacesTheCursor(t *testing.T) {
	_, a := ducked(t, -400, 5)
	a.IsPlayer = true
	steps(a, 1, Input{Roll: true, FaceLeft: true})
	if a.Orientation != -1 {
		t.Error("the ducking fighter did not turn to the cursor on its left")
	}
	if v := a.rollLaunch(); v.X >= 0 {
		t.Errorf("turned left, the roll would launch at %v", v)
	}
}

func TestTooPoorToDuck(t *testing.T) {
	_, a := walker(t, -400)
	a.Energy = EnergyRunJump - 1
	a.Events = 0
	steps(a, 1, Input{Roll: true})
	if a.State != StateIdle || a.Energy < EnergyRunJump-1 {
		t.Errorf("a poor fighter pressing roll: state %d, energy %d", a.State, a.Energy)
	}
	if a.Events&EvNoEnergy == 0 {
		t.Error("the refused duck raised no refusal")
	}
}

func TestReleaseRollsAndPays(t *testing.T) {
	_, a := ducked(t, -400, 20)
	if a.Energy != MaxEnergy {
		t.Errorf("ducking spent energy: %d", a.Energy)
	}
	steps(a, 1, Input{})
	if a.State != StateRoll || a.OnGround || a.Anim != "resources/threefour/roll.png" {
		t.Fatalf("released: state %d, on ground %v, anim %q", a.State, a.OnGround, a.Anim)
	}
	if a.Energy != MaxEnergy-EnergyRunJump || a.EnergyCalm != 0 {
		t.Errorf("released: energy %d calm %d, want %d and a restarted wait", a.Energy, a.EnergyCalm, MaxEnergy-EnergyRunJump)
	}
}

func TestARollIsNotSteeredAndDoesNotFall(t *testing.T) {
	_, a := ducked(t, -400, 30)
	steps(a, 1, Input{})
	sx := a.Speed.X
	for a.State == StateRoll {
		steps(a, 1, Input{Right: true}) // Right moves -X: against the roll
		if a.State == StateRoll && a.Speed.X != sx {
			t.Fatalf("a steered roll changed speed from %v to %v", sx, a.Speed.X)
		}
		if a.State == StateFall {
			t.Fatal("the roll turned into a fall past its apex")
		}
	}
	if a.State != StateRollRecover || a.Anim != "resources/threefour/roll-recover.png" {
		t.Fatalf("landed a roll into state %d, anim %q", a.State, a.Anim)
	}
	// pressed through most of the recovery, which lasts 18 ticks: nothing
	// comes out of it until it is over
	steps(a, 12, Input{Attack: true, Jump: true})
	if a.State != StateRollRecover {
		t.Fatalf("attack and jump during the recovery put the fighter in state %d", a.State)
	}
	steps(a, 20, Input{})
	if a.State != StateIdle && a.State != StateRun {
		t.Errorf("after the recovery the fighter is in state %d, want it standing or running", a.State)
	}
}

func TestLongerHoldIsHigherAndFarther(t *testing.T) {
	_, a := ducked(t, -400, 15)
	start := a.Pos.X
	shortPeak, shortLand := rollOut(t, a)
	_, b := ducked(t, -400, 45)
	longPeak, longLand := rollOut(t, b)
	if !(longPeak < shortPeak && longLand-start > shortLand-start) {
		t.Errorf("15 ticks: peak %v land %v; 45 ticks: peak %v land %v", shortPeak, shortLand, longPeak, longLand)
	}
}

func TestADuckDoesNotGuard(t *testing.T) {
	floor()
	w := NewWorld(true)
	att := w.Spawn(rl.Vector2{X: -400}, false)
	def := w.Spawn(rl.Vector2{X: -380}, false)
	att.Team, def.Team = NoTeam, NoTeam
	att.Orientation, def.Orientation = 1, -1
	for range 20 {
		run(w, 1)
	}
	def.Input = Input{Roll: true}
	run(w, 1)
	if def.State != StateDuck {
		t.Fatalf("the defender is in state %d, want ducked", def.State)
	}
	att.Updown = 0
	if !att.enterAttackState() {
		t.Fatal("the attacker could not swing")
	}
	for range 30 {
		att.Input = Input{}
		def.Input = Input{Roll: true, Updown: 0} // the guard it would block this swing with
		run(w, 1)
		if def.State != StateDuck {
			break
		}
	}
	if def.State != StateAttacked && def.Structure == MaxStructure {
		t.Fatalf("the ducking defender was not hit: state %d", def.State)
	}
	if def.Energy != MaxEnergy {
		t.Errorf("the interrupted duck spent energy: %d", def.Energy)
	}
}

// The drawn arc is the roll: released on open ground, the fighter's feet land
// where the arc drawn the tick before ended.
func TestTheArcIsWhereTheRollGoes(t *testing.T) {
	_, a := ducked(t, -400, 40)
	path := a.rollPath()
	if len(path) < 10 {
		t.Fatalf("the path is %d points long", len(path))
	}
	if a.State != StateDuck || a.Energy != MaxEnergy {
		t.Fatal("drawing the path disturbed the fighter")
	}
	rollOut(t, a)
	end := path[len(path)-1]
	feetX := a.Pos.X + a.CollisionBox.X + a.CollisionBox.Width/2
	feetY := a.Pos.Y + a.CollisionBox.Y + a.CollisionBox.Height
	if abs(end.X-feetX) > 1 || abs(end.Y-feetY) > 1 {
		t.Errorf("the arc ends at %v, the roll landed at (%v, %v)", end, feetX, feetY)
	}
}

// A click is an escape: a fighter boxed in on both sides taps roll and comes
// to rest well clear of the enemy it rolled past.
func TestATapEscapes(t *testing.T) {
	w, a := walker(t, -400)
	left := w.Spawn(a.Pos, false)
	right := w.Spawn(a.Pos, false)
	left.Pos.X -= 20
	right.Pos.X += 20
	left.Team, right.Team, a.Team = NoTeam, NoTeam, NoTeam
	steps(a, 6, Input{Roll: true}) // a click: about a tenth of a second held
	rollOut(t, a)
	steps(a, 40, Input{})
	if clear := a.Pos.X - right.Pos.X; clear < 48 {
		t.Errorf("after a tapped roll the fighter rests %v px past the enemy it rolled over, want at least 48", clear)
	}
}

// A roll that carries the head up to a ledge climbs it, as a jump does; one
// that stays too low passes under it and lands.
func TestAHighRollClimbs(t *testing.T) {
	for _, c := range []struct {
		name  string
		hold  int
		aim   uint8
		climb bool
	}{
		{"full charge aimed steep", RollChargeTicks, 255, true},
		{"a flat tap", 1, 0, false},
	} {
		tiles(append(row(-480, 48, 20), Vector2i{X: 0, Y: -96})...) // a ledge hanging at y -96..-48, x 0..48
		w := NewWorld(true)
		a := w.Spawn(rl.Vector2{X: -60}, false)
		a.Orientation = 1
		steps(a, 20, Input{})
		steps(a, c.hold, Input{Roll: true, RollAim: c.aim})
		climbed := false
		for range 180 {
			steps(a, 1, Input{RollAim: c.aim})
			climbed = climbed || a.State == StateClimb
		}
		if climbed != c.climb {
			t.Errorf("%s: climbed %t, want %t (ended at %v in state %d)", c.name, climbed, c.climb, a.Pos, a.State)
		}
		if c.climb && a.Pos.Y > -143 {
			t.Errorf("%s: ended at y %v, want on top of the wall", c.name, a.Pos.Y)
		}
	}
}
