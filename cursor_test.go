package resistor

import (
	"math"
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// A full pool follows exactly as the cross did at 60 fps; an empty one lags,
// and the share only grows with energy.
func TestAttackFollowByEnergy(t *testing.T) {
	if got := attackFollow(1, 1.0/60); math.Abs(float64(got)-0.5) > 1e-5 {
		t.Errorf("full pool at 60 fps closes %v of the gap, want 0.5", got)
	}
	if full, empty := attackFollow(1, 1.0/60), attackFollow(0, 1.0/60); empty >= full {
		t.Errorf("empty pool closes %v, full %v: want empty slower", empty, full)
	}
	prev := float32(-1)
	for i := 0; i <= 20; i++ {
		got := attackFollow(float32(i)/20, 1.0/60)
		if got < prev {
			t.Errorf("share falls from %v to %v as energy rises to %v", prev, got, float32(i)/20)
		}
		prev = got
	}
}

// Two half steps leave the same gap as one whole step, so the cross settles as
// fast at any frame rate.
func TestAttackFollowFrameRateIndependent(t *testing.T) {
	for _, e := range []float32{0, 0.5, 1} {
		dt := float32(1.0 / 30)
		whole := 1 - attackFollow(e, dt)
		half := 1 - attackFollow(e, dt/2)
		if math.Abs(float64(whole-half*half)) > 1e-5 {
			t.Errorf("energy %v: one step leaves %v, two half steps %v", e, whole, half*half)
		}
	}
}

func TestAttackFollowClamps(t *testing.T) {
	dt := float32(1.0 / 60)
	if attackFollow(2, dt) != attackFollow(1, dt) {
		t.Error("energy above full is not treated as full")
	}
	if attackFollow(-1, dt) != attackFollow(0, dt) {
		t.Error("energy below empty is not treated as empty")
	}
}

func TestCrosshairColor(t *testing.T) {
	if got := crosshairColor(1); got != rl.Green {
		t.Errorf("full pool: %v, want %v", got, rl.Green)
	}
	if got := crosshairColor(0); got != crosshairTired {
		t.Errorf("empty pool: %v, want %v", got, crosshairTired)
	}
	half := crosshairColor(0.5)
	for _, c := range []struct {
		name             string
		got, tired, full uint8
	}{
		{"R", half.R, crosshairTired.R, rl.Green.R},
		{"G", half.G, crosshairTired.G, rl.Green.G},
		{"B", half.B, crosshairTired.B, rl.Green.B},
	} {
		want := (float64(c.tired) + float64(c.full)) / 2
		if math.Abs(float64(c.got)-want) > 0.5 {
			t.Errorf("half pool %s = %d, want about %v", c.name, c.got, want)
		}
	}
	for _, e := range []float32{-1, 0, 0.3, 0.5, 1, 2} {
		if a := crosshairColor(e).A; a != 255 {
			t.Errorf("energy %v: alpha %d, want 255", e, a)
		}
	}
	if crosshairColor(2) != rl.Green || crosshairColor(-1) != crosshairTired {
		t.Error("energy outside 0..1 is not clamped")
	}
}

func TestDetectFlash(t *testing.T) {
	pt := func(n int32) float32 { return float32(n) / MaxEnergy }
	step := float32(1.0 / 60)
	perSecond := pt(60 / EnergyFreeRegenTicks)       // the fastest refilling
	refill60, refill10 := perSecond/60, perSecond/10 // a frame's worth at 60 and 10 fps
	for _, c := range []struct {
		name      string
		prev, now float32
		dt        float32
		live      bool
		want      flashKind
	}{
		{"a payback below full", pt(100), pt(100 + EnergyBlockRestore), step, true, flashBlock},
		{"a payback that fills the pool", pt(MaxEnergy - 5), 1, step, true, flashFull},
		{"refilling at 60 fps", pt(100), pt(100) + refill60, step, true, flashNone},
		{"refilling at 10 fps", pt(100), pt(100) + refill10, 0.1, true, flashNone},
		{"refilling to full", pt(MaxEnergy - 1), 1, step, true, flashFull},
		{"staying full", 1, 1, step, true, flashNone},
		{"spending", pt(100), pt(100 - EnergySwing), step, true, flashNone},
		{"a payback, not live", pt(100), pt(100 + EnergyBlockRestore), step, false, flashNone},
		{"a respawn, not live", 0, 1, step, false, flashNone},
	} {
		if got := detectFlash(c.prev, c.now, c.dt, c.live); got != c.want {
			t.Errorf("%s: flash %d, want %d", c.name, got, c.want)
		}
	}
}

func TestFlashRing(t *testing.T) {
	for _, k := range []flashKind{flashBlock, flashFull} {
		s := flashStyles[k]
		inner, outer, col, ok := flashRing(k, 0)
		if !ok || outer != s.from || inner != s.from-s.thick || col.A != s.col.A {
			t.Errorf("kind %d at start: %v..%v alpha %d ok %v, want %v..%v alpha %d", k, inner, outer, col.A, ok, s.from-s.thick, s.from, s.col.A)
		}
		_, outer, _, ok = flashRing(k, s.dur*0.999)
		if !ok || math.Abs(float64(outer-s.to)) > 0.1 {
			t.Errorf("kind %d at the end: radius %v, want about %v", k, outer, s.to)
		}
		if _, _, _, ok := flashRing(k, s.dur); ok {
			t.Errorf("kind %d still showing after %v s", k, s.dur)
		}
		prevR, prevA := float32(0), 256
		for i := 0; i < 20; i++ {
			_, r, c, _ := flashRing(k, s.dur*float32(i)/20)
			if r < prevR || int(c.A) > prevA {
				t.Errorf("kind %d step %d: radius %v after %v, alpha %d after %d", k, i, r, prevR, c.A, prevA)
			}
			prevR, prevA = r, int(c.A)
		}
	}
	if _, _, _, ok := flashRing(flashNone, 0); ok {
		t.Error("no flash still draws a ring")
	}
	b, f := flashStyles[flashBlock], flashStyles[flashFull]
	if f.to <= b.to || f.dur <= b.dur {
		t.Error("the full flash is not bigger and longer than the block flash")
	}
}

// The cursor's flash through a match: a payback, refilling to full, a respawn
// and a rewind, one frame at a time.
func TestCursorTracksEnergy(t *testing.T) {
	const dt = float32(1.0 / 60)
	c := Cursor{energy: 1}
	pt := func(n int32) float32 { return float32(n) / MaxEnergy }
	showing := func() flashKind {
		if _, _, _, ok := flashRing(c.flash, c.flashAge); ok {
			return c.flash
		}
		return flashNone
	}

	c.trackEnergy(dt, pt(100), true) // swung down to 100
	if showing() != flashNone {
		t.Fatalf("spending flashed %d", showing())
	}
	c.trackEnergy(dt, pt(100+EnergyBlockRestore), true)
	if showing() != flashBlock {
		t.Fatalf("a payback shows %d, want the block flash", showing())
	}
	for range 30 { // half a second: the block flash is over
		c.trackEnergy(dt, c.energy, true)
	}
	if showing() != flashNone {
		t.Fatalf("the block flash still shows after half a second")
	}

	for e := int32(100 + EnergyBlockRestore); e < MaxEnergy; e++ {
		c.trackEnergy(dt, pt(e), true)
		if showing() != flashNone {
			t.Fatalf("refilling flashed %d at %d energy", showing(), e)
		}
	}
	c.trackEnergy(dt, 1, true)
	if showing() != flashFull {
		t.Fatalf("a full pool shows %d, want the full flash", showing())
	}
	for range 60 {
		c.trackEnergy(dt, 1, true)
	}

	c.trackEnergy(dt, 0, true) // died: the energy of a new fighter follows
	c.trackEnergy(dt, 1, false)
	if showing() != flashNone {
		t.Errorf("a respawn flashed %d", showing())
	}

	c.trackEnergy(dt, pt(50), false) // rewound to an emptier moment...
	c.trackEnergy(dt, 1, false)      // ...and forward to a full one
	c.trackEnergy(dt, 1, true)       // live again where it was
	if showing() != flashNone {
		t.Errorf("a rewind flashed %d", showing())
	}
}
