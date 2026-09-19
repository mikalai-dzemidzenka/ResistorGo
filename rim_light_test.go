package resistor

import (
	"math"
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func near(a, b float32) bool {
	return math.Abs(float64(a-b)) < 1e-4
}

// The rim fades as the light map does: full at the centre, (1 - d/R)² away from
// it, nothing at the radius and beyond, and nothing from a light that is off.
func TestRimStrengthAt(t *testing.T) {
	l := LevelLight{X: 10, Y: 20, Radius: 100, Color: [3]uint8{255, 255, 255}, Intensity: 2}
	for _, c := range []struct {
		name string
		p    rl.Vector2
		want float32
	}{
		{"centre", rl.Vector2{X: 10, Y: 20}, 2},
		{"half the radius", rl.Vector2{X: 60, Y: 20}, 0.5},
		{"at the radius", rl.Vector2{X: 10, Y: 120}, 0},
		{"beyond the radius", rl.Vector2{X: 300, Y: 20}, 0},
	} {
		if got := rimStrengthAt(l, c.p); !near(got, c.want) {
			t.Errorf("%s: strength %v, want %v", c.name, got, c.want)
		}
	}
	// stepping away weakens it steadily
	prev := rimStrengthAt(l, l.center())
	for x := float32(11); x < 110; x += 7 {
		s := rimStrengthAt(l, rl.Vector2{X: x, Y: 20})
		if s >= prev {
			t.Fatalf("strength at x=%v is %v, not below %v", x, s, prev)
		}
		prev = s
	}
	l.Intensity = 0
	if got := rimStrengthAt(l, l.center()); got != 0 {
		t.Errorf("a light at zero intensity has strength %v", got)
	}
}

func TestSegmentHitsRect(t *testing.T) {
	r := rl.Rectangle{X: 10, Y: 10, Width: 10, Height: 10}
	v := func(x, y float32) rl.Vector2 { return rl.Vector2{X: x, Y: y} }
	for _, c := range []struct {
		name string
		a, b rl.Vector2
		want bool
	}{
		{"crossing diagonally", v(0, 0), v(30, 30), true},
		{"crossing horizontally", v(0, 15), v(30, 15), true},
		{"crossing vertically", v(15, 0), v(15, 30), true},
		{"passing beside, horizontal", v(0, 5), v(30, 5), false},
		{"passing beside, vertical", v(25, 0), v(25, 30), false},
		{"passing beside, diagonal", v(0, 10), v(10, 0), false},
		{"stopping short", v(0, 15), v(9, 15), false},
		{"starting inside", v(15, 15), v(40, 40), true},
		{"wholly inside", v(12, 12), v(18, 18), true},
		{"a point outside", v(5, 5), v(5, 5), false},
	} {
		if got := segmentHitsRect(c.a, c.b, r); got != c.want {
			t.Errorf("%s: hit %v, want %v", c.name, got, c.want)
		}
		if got := segmentHitsRect(c.b, c.a, r); got != c.want {
			t.Errorf("%s, reversed: hit %v, want %v", c.name, got, c.want)
		}
	}
}

func rimTestLight(x, y int32, intensity float32) LevelLight {
	return LevelLight{X: x, Y: y, Radius: 100, Color: [3]uint8{0, 255, 255}, Intensity: intensity}
}

// A fighter's body, centred on (0, 0).
var rimTestBody = rl.Rectangle{X: -8, Y: -16, Width: 16, Height: 32}

func TestRimLightsReach(t *testing.T) {
	// Out of reach
	if got := rimLights([]LevelLight{rimTestLight(150, 0, 1)}, nil, rimTestBody); len(got) != 0 {
		t.Errorf("a light beyond its radius rims the fighter: %+v", got)
	}
	// a light that is off
	if got := rimLights([]LevelLight{rimTestLight(20, 0, 0)}, nil, rimTestBody); len(got) != 0 {
		t.Errorf("a light at zero intensity rims the fighter: %+v", got)
	}
	// Clear line: colour, position and strength come through
	got := rimLights([]LevelLight{rimTestLight(-50, 0, 1)}, nil, rimTestBody)
	if len(got) != 1 {
		t.Fatalf("a light in reach gives %d rims, want 1", len(got))
	}
	want := rimLight{pos: rl.Vector2{X: -50, Y: 0}, color: rl.Color{R: 0, G: 255, B: 255, A: 255}, strength: 0.25}
	if got[0].pos != want.pos || got[0].color != want.color || !near(got[0].strength, want.strength) {
		t.Errorf("rim %+v, want %+v", got[0], want)
	}
}

func TestRimLightsShadow(t *testing.T) {
	lights := []LevelLight{rimTestLight(-50, 0, 1)}
	// Behind a wall
	wall := rl.Rectangle{X: -30, Y: -40, Width: 4, Height: 80}
	if got := rimLights(lights, []rl.Rectangle{wall}, rimTestBody); len(got) != 0 {
		t.Errorf("a light behind a wall rims the fighter: %+v", got)
	}
	// a wall elsewhere shadows nothing
	aside := rl.Rectangle{X: -30, Y: 20, Width: 4, Height: 80}
	if got := rimLights(lights, []rl.Rectangle{aside}, rimTestBody); len(got) != 1 {
		t.Errorf("a wall off the line shadows the light: %+v", got)
	}
}

func TestRimLightsStrongestFirst(t *testing.T) {
	// Too many lights: six in reach, the four nearest are kept, nearest first
	var lights []LevelLight
	for _, d := range []int32{60, 10, 50, 30, 20, 40} {
		lights = append(lights, rimTestLight(d, 0, 1))
	}
	got := rimLights(lights, nil, rimTestBody)
	if len(got) != rimMaxLights {
		t.Fatalf("%d rims, want %d", len(got), rimMaxLights)
	}
	for i, x := range []float32{10, 20, 30, 40} {
		if got[i].pos.X != x {
			t.Errorf("rim %d is the light at x=%v, want x=%v", i, got[i].pos.X, x)
		}
	}
}

// Turning around: the frame is mirrored across the quad when the fighter faces
// left, so a light on the fighter's left lands on the frame's right.
func TestRimLocal(t *testing.T) {
	spritePos := rl.Vector2{X: 100, Y: 50}
	light := rl.Vector2{X: 80, Y: 60}

	right := rimLocal(light, spritePos, 48, false)
	if right != (rl.Vector2{X: -20, Y: 10}) {
		t.Errorf("facing right: %+v, want {-20 10}", right)
	}
	left := rimLocal(light, spritePos, 48, true)
	if left != (rl.Vector2{X: 68, Y: 10}) {
		t.Errorf("facing left: %+v, want {68 10}", left)
	}
	if left.X <= 48 {
		t.Errorf("facing left, a light on the fighter's left is not on the frame's right: %v", left.X)
	}
}

// The rim's quad is a pixel larger than the sprite's on every side, so the ring
// of empty pixels around the fighter has somewhere to land, and a left-facing
// fighter's mirror is kept.
func TestRimQuad(t *testing.T) {
	pos := rl.Vector2{X: 100, Y: 50}
	src := rl.Rectangle{X: 96, Y: 0, Width: 48, Height: 48}
	grown, dst := rimQuad(src, pos)
	if want := (rl.Rectangle{X: 95, Y: -1, Width: 50, Height: 50}); grown != want {
		t.Errorf("source %+v, want %+v", grown, want)
	}
	if want := (rl.Rectangle{X: 99, Y: 49, Width: 50, Height: 50}); dst != want {
		t.Errorf("destination %+v, want %+v", dst, want)
	}
	// facing left: the same region, still mirrored
	flipped, dstL := rimQuad(rl.Rectangle{X: 96, Y: 0, Width: -48, Height: 48}, pos)
	if flipped.Width >= 0 {
		t.Errorf("a left-facing fighter lost its mirror: width %v", flipped.Width)
	}
	if flipped.X != grown.X || flipped.Y != grown.Y || -flipped.Width != grown.Width {
		t.Errorf("mirrored source %+v does not cover the same region as %+v", flipped, grown)
	}
	if dstL != dst {
		t.Errorf("mirrored destination %+v, want %+v", dstL, dst)
	}
}

// The top layer is taken from the same frame at the same place as the body, and
// there is none unless the fighter has one.
func TestTopDraw(t *testing.T) {
	a := NewActor(rl.Vector2{X: 10, Y: 20}, true)
	a.Anim = "resources/threefour/idle.png"
	a.AnimCurrentFrame = 3
	if _, _, _, ok := a.topDraw(); ok {
		t.Errorf("a fighter with no top layer reports one")
	}
	missing := "resources/threefour/not-loaded.png"
	a.AnimTop = &missing
	if _, _, _, ok := a.topDraw(); ok {
		t.Errorf("a top layer that is not a loaded texture reports one")
	}

	top := "resources/threefour/top-test.png"
	resources[top] = rl.Texture2D{ID: 1, Width: 480, Height: 48}
	defer delete(resources, top)
	a.AnimTop = &top
	_, src, _, ok := a.topDraw()
	if !ok {
		t.Fatalf("a fighter with a loaded top layer reports none")
	}
	_, body, _ := a.spriteDraw()
	if src != body {
		t.Errorf("top layer frame %+v, want the body's %+v", src, body)
	}
}

// The walk is drawn in two layers: the body walks and the arms are a second
// sheet chosen by the stance. The rim has to see both, so both come from one
// place.
func TestTopAnimWalk(t *testing.T) {
	a := NewActor(rl.Vector2{}, true)
	a.Anim = "resources/threefour/idle.png"
	if _, ok := a.topAnim(); ok {
		t.Errorf("a fighter that is not walking reports a top layer")
	}

	a.Anim = "resources/threefour/walk.png"
	for updown, want := range map[uint8]string{
		0: "resources/threefour/walk-bot.png",
		1: "resources/threefour/walk-mid.png",
		2: "resources/threefour/walk-top.png",
	} {
		a.Updown = updown
		got, ok := a.topAnim()
		if !ok || got != want {
			t.Errorf("walking at stance %d: %q (%v), want %q", updown, got, ok, want)
		}
	}
	a.Updown = 3 // block-all walks with no arms of its own
	if _, ok := a.topAnim(); ok {
		t.Errorf("block-all reports a top layer")
	}

	// an explicit top layer wins over the walk's rule
	a.Updown = 1
	top := "resources/threefour/attack-high.png"
	a.AnimTop = &top
	if got, ok := a.topAnim(); !ok || got != top {
		t.Errorf("AnimTop gives %q (%v), want %q", got, ok, top)
	}
}
