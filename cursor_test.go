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
