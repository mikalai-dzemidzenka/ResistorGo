package resistor

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func TestLetterbox(t *testing.T) {
	cases := []struct {
		name                   string
		winW, winH             float32
		wantX, wantY, wantW, h float32
	}{
		{"exact size", 1280, 720, 0, 0, 1280, 720},
		{"exact double", 2560, 1440, 0, 0, 2560, 1440},
		{"16:9 step", 1920, 1080, 0, 0, 1920, 1080},
		{"16:10 bars top and bottom", 1920, 1200, 0, 60, 1920, 1080},
		{"4:3 bars top and bottom", 1024, 768, 0, 96, 1024, 576},
		{"ultrawide bars at the sides", 2560, 1080, 320, 0, 1920, 1080},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := letterboxFor(c.winW, c.winH)
			if r.X != c.wantX || r.Y != c.wantY || r.Width != c.wantW || r.Height != c.h {
				t.Errorf("letterboxFor(%v,%v) = %+v, want x%v y%v w%v h%v",
					c.winW, c.winH, r, c.wantX, c.wantY, c.wantW, c.h)
			}
			// never larger than the window, always centred, aspect preserved
			if r.Width > c.winW || r.Height > c.winH {
				t.Errorf("%+v overflows the %vx%v window", r, c.winW, c.winH)
			}
			if got, want := r.X*2+r.Width, c.winW; got != want {
				t.Errorf("not centred horizontally: %v vs %v", got, want)
			}
			if got, want := r.Y*2+r.Height, c.winH; got != want {
				t.Errorf("not centred vertically: %v vs %v", got, want)
			}
			if got, want := r.Width/r.Height, float32(VW)/VH; got != want {
				t.Errorf("aspect %v, want %v", got, want)
			}
		})
	}
}

// Point sampling is exact only when a virtual pixel covers a whole number of
// window pixels; the in-between window sizes get bilinear instead.
func TestPresentFilter(t *testing.T) {
	for _, c := range []struct {
		winW, winH float32
		want       rl.TextureFilterMode
	}{
		{1280, 720, rl.FilterPoint},     // 1x
		{2560, 1440, rl.FilterPoint},    // 2x
		{1600, 900, rl.FilterBilinear},  // 1.25x
		{1920, 1080, rl.FilterBilinear}, // 1.5x
	} {
		scale := letterboxFor(c.winW, c.winH).Width / VW
		if got := presentFilter(scale); got != c.want {
			t.Errorf("%vx%v (scale %v): filter %v, want %v", c.winW, c.winH, scale, got, c.want)
		}
	}
}
