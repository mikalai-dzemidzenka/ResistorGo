package resistor

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func overlaps(a, b rl.Rectangle) bool {
	return a.X < b.X+b.Width && b.X < a.X+a.Width
}

// The five landmarks are what make the skyline Minsk: each has to be there, in
// the scene, standing on the horizon, and not covered by its neighbours.
func TestCityLandmarks(t *testing.T) {
	city := cityLayout(citySeed)
	for _, name := range []string{"avenue", "gates", "library", "obelisk", "tvtower"} {
		var found *cityBuilding
		for i := range city {
			if city[i].Name == name {
				found = &city[i]
			}
		}
		if found == nil {
			t.Errorf("the %s is missing from the skyline", name)
			continue
		}
		r := found.Rect
		if r.X < 0 || r.X+r.Width > menuSceneW {
			t.Errorf("the %s spans x %v..%v, outside the %d wide scene", name, r.X, r.X+r.Width, menuSceneW)
		}
		for _, other := range city {
			if other.Name != name && overlaps(r, other.Rect) {
				t.Errorf("the %s at x %v..%v is covered by a building at x %v..%v", name, r.X, r.X+r.Width, other.Rect.X, other.Rect.X+other.Rect.Width)
			}
		}
	}
}

// Every building stands on the horizon, and none reaches into the sky the sun
// and the dusk gradient need.
func TestCityFitsTheScene(t *testing.T) {
	for _, b := range cityLayout(citySeed) {
		name := b.Name
		if name == "" {
			name = "a filler building"
		}
		if b.Rect.Y+b.Rect.Height != cityHorizonY {
			t.Errorf("%s ends at y %v, not on the horizon at %d", name, b.Rect.Y+b.Rect.Height, cityHorizonY)
		}
		if b.Rect.Y < citySkyMargin {
			t.Errorf("%s reaches up to y %v, above the sky margin at %d", name, b.Rect.Y, citySkyMargin)
		}
		if b.Rect.X < 0 || b.Rect.X+b.Rect.Width > menuSceneW {
			t.Errorf("%s spans x %v..%v, outside the %d wide scene", name, b.Rect.X, b.Rect.X+b.Rect.Width, menuSceneW)
		}
		body := cityBody(b)
		for _, w := range b.Windows {
			if w.X < body.X || w.X+cityWindow > body.X+body.Width || w.Y < body.Y || w.Y+cityWindow > cityHorizonY {
				t.Errorf("%s has a lit window at %+v, outside its solid body %+v", name, w, body)
			}
		}
	}
}

// No filler building may hide a landmark behind it.
func TestCityBuildingsDoNotOverlap(t *testing.T) {
	city := cityLayout(citySeed)
	for i, a := range city {
		for _, b := range city[i+1:] {
			if overlaps(a.Rect, b.Rect) {
				t.Errorf("buildings overlap: x %v..%v and x %v..%v", a.Rect.X, a.Rect.X+a.Rect.Width, b.Rect.X, b.Rect.X+b.Rect.Width)
			}
		}
	}
	if len(city) <= len(cityLandmarks) {
		t.Errorf("the skyline has %d buildings, want the landmarks plus fillers", len(city))
	}
}

// The same seed gives the same city, so it looks identical on every run.
func TestCityIsDeterministic(t *testing.T) {
	a, b := cityLayout(citySeed), cityLayout(citySeed)
	if len(a) != len(b) {
		t.Fatalf("two layouts of the same seed have %d and %d buildings", len(a), len(b))
	}
	for i := range a {
		if a[i].Name != b[i].Name || a[i].Rect != b[i].Rect || a[i].Shape != b[i].Shape {
			t.Fatalf("building %d differs between two layouts of the same seed: %+v vs %+v", i, a[i], b[i])
		}
		if len(a[i].Windows) != len(b[i].Windows) {
			t.Fatalf("building %d has %d lit windows in one layout and %d in the other", i, len(a[i].Windows), len(b[i].Windows))
		}
		for j := range a[i].Windows {
			if a[i].Windows[j] != b[i].Windows[j] {
				t.Fatalf("building %d window %d differs: %+v vs %+v", i, j, a[i].Windows[j], b[i].Windows[j])
			}
		}
	}
	lit := 0
	for _, x := range a {
		lit += len(x.Windows)
	}
	if lit == 0 {
		t.Errorf("no window in the city is lit")
	}
}
