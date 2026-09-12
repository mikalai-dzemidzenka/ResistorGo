package resistor

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"testing"
)

func historyWorld() (*World, *Actor, *Actor) {
	w := NewWorld(true)
	w.History = NewLagHistory()
	att := w.Spawn(rl.Vector2{}, true)
	def := w.Spawn(rl.Vector2{X: 100}, true)
	for tick := uint32(1); tick <= 20; tick++ {
		w.Tick = tick
		def.Pos.X = 100 + float32(tick)*2
		def.Updown = uint8(tick % 3)
		w.History.Record(w)
	}
	w.Tick = 21
	def.Pos.X = 200
	def.Updown = 2
	return w, att, def
}

func TestLagCompRewindRestore(t *testing.T) {
	w, att, def := historyWorld()
	attPos := att.Pos

	saves := w.History.rewindOthers(w, att, 10)
	if def.Pos.X != 120 || def.Updown != 1 {
		t.Errorf("rewound to tick 10: x=%v updown=%d, want 120 1", def.Pos.X, def.Updown)
	}
	if att.Pos != attPos {
		t.Error("attacker was rewound")
	}
	restore(saves)
	if def.Pos.X != 200 || def.Updown != 2 {
		t.Errorf("after restore: x=%v updown=%d, want 200 2", def.Pos.X, def.Updown)
	}

	saves = w.History.rewindOthers(w, att, 10.5)
	if def.Pos.X != 121 {
		t.Errorf("fractional rewind: x=%v, want 121", def.Pos.X)
	}
	restore(saves)

	if saves := w.History.rewindOthers(w, att, float64(w.Tick)); saves != nil || def.Pos.X != 200 {
		t.Error("rewind to the current tick changed state")
	}
}

func TestClampViewTime(t *testing.T) {
	cases := []struct {
		view float64
		now  uint32
		want float64
	}{
		{50, 100, 85},     // 600+ ms behind: clamped to the 250 ms window
		{120, 100, 100},   // forged future tick: current state
		{95.5, 100, 95.5}, // inside the window
		{0, 5, 0},         // window before tick 0
	}
	for _, c := range cases {
		if got := clampViewTime(c.view, c.now, DefaultMaxUnlag); got != c.want {
			t.Errorf("clamp(%v, %d) = %v, want %v", c.view, c.now, got, c.want)
		}
	}
}

// An attacker swings at where it saw the defender: in range and with a low guard.
// On the server the defender has since stepped out of range and raised its guard.
func lagCompSwing(t *testing.T, lagComp bool) *Actor {
	t.Helper()
	w := NewWorld(true)
	w.History = NewLagHistory()
	att := w.Spawn(rl.Vector2{}, true)
	def := w.Spawn(rl.Vector2{X: 20}, true)
	def.Updown = 0
	for tick := uint32(1); tick <= 20; tick++ {
		w.Tick = tick
		w.History.Record(w)
	}
	w.Tick = 21
	def.Pos.X = 200 // stepped back
	def.Updown = 1  // same guard as the attack: would block

	att.Updown = 1
	att.Orientation = 1
	att.Ticks = 100
	att.SetAnim("resources/threefour/attack-mid.png", AnimOptions{animDuration: 0.3, totalFrames: 8, once: true})
	att.State = StateAttack
	att.Ticks = 114 // frame 6, past the hit frame
	att.advanceAnim()
	att.Input = Input{FaceLeft: false}

	s := &Server{Cfg: ServerConfig{LagComp: lagComp, MaxUnlag: DefaultMaxUnlag}, W: w}
	s.runCmd(&netPlayer{ID: 1, Actor: att}, Input{Updown: 1, ViewTick: 10})
	if def.Pos.X != 200 || def.Updown != 1 {
		t.Errorf("defender not restored: x=%v updown=%d", def.Pos.X, def.Updown)
	}
	return def
}

func TestLagCompHitsWhatAttackerSaw(t *testing.T) {
	if def := lagCompSwing(t, true); def.Structure != MaxStructure-HitStructure {
		t.Errorf("with lag compensation the defender is at %v structure, want the %v a hit leaves", def.Structure, float32(MaxStructure-HitStructure))
	}
}

// A client's predicted world animates swings but never decides hits.
func TestNonAuthorityNeverHits(t *testing.T) {
	for _, auth := range []bool{true, false} {
		w := NewWorld(auth)
		att := w.Spawn(rl.Vector2{}, true)
		def := w.Spawn(rl.Vector2{X: 20}, true)
		def.Updown = 0
		att.Updown = 1
		att.Orientation = 1
		att.Ticks = 100
		att.SetAnim("resources/threefour/attack-mid.png", AnimOptions{animDuration: 0.3, totalFrames: 8, once: true})
		att.State = StateAttack
		att.Ticks = 114
		att.advanceAnim()
		att.Input = Input{Updown: 1}
		att.Update(float32(FixedTimeStep))
		if hit := def.Structure < MaxStructure; hit != auth {
			t.Errorf("authority %t: defender hit = %t", auth, hit)
		}
		if att.AnimCurrentFrame == 0 {
			t.Errorf("authority %t: attack animation did not advance", auth)
		}
	}
}

func TestLagCompDisabledMisses(t *testing.T) {
	if def := lagCompSwing(t, false); def.State != StateIdle {
		t.Errorf("without lag compensation the defender is in state %d, want idle", def.State)
	}
}
