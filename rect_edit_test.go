package resistor

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// Corners win over edges, edges over the body, and nothing is grabbed from
// further away than the tolerance.
func TestRectHandleAt(t *testing.T) {
	r := rl.Rectangle{X: 10, Y: 20, Width: 30, Height: 40} // right 40, bottom 60
	const tol = 2
	cases := []struct {
		name string
		p    rl.Vector2
		want handle
	}{
		{"top-left", rl.Vector2{X: 10, Y: 20}, handleTopLeft},
		{"top-right", rl.Vector2{X: 40, Y: 20}, handleTopRight},
		{"bottom-left", rl.Vector2{X: 10, Y: 60}, handleBottomLeft},
		{"bottom-right", rl.Vector2{X: 41, Y: 61}, handleBottomRight},
		{"left", rl.Vector2{X: 10, Y: 40}, handleLeft},
		{"right", rl.Vector2{X: 40, Y: 40}, handleRight},
		{"top", rl.Vector2{X: 25, Y: 20}, handleTop},
		{"bottom", rl.Vector2{X: 25, Y: 60}, handleBottom},
		{"body", rl.Vector2{X: 25, Y: 40}, handleBody},
		{"within tolerance", rl.Vector2{X: 8.5, Y: 40}, handleLeft},
		{"just outside tolerance", rl.Vector2{X: 7.9, Y: 40}, handleNone},
		{"far away", rl.Vector2{X: 0, Y: 0}, handleNone},
	}
	for _, c := range cases {
		if got := rectHandleAt(r, c.p, tol); got != c.want {
			t.Errorf("%s: handle at %v = %d, want %d", c.name, c.p, got, c.want)
		}
	}
}

func TestDragRect(t *testing.T) {
	r := rl.Rectangle{X: 10, Y: 20, Width: 30, Height: 40}
	cases := []struct {
		name string
		h    handle
		to   rl.Vector2
		want rl.Rectangle
	}{
		{"move the body", handleBody, rl.Vector2{X: 0, Y: 5}, rl.Rectangle{X: 0, Y: 5, Width: 30, Height: 40}},
		{"resize by a corner", handleBottomRight, rl.Vector2{X: 42, Y: 60}, rl.Rectangle{X: 10, Y: 20, Width: 32, Height: 40}},
		{"opposite corner stays", handleTopLeft, rl.Vector2{X: 5, Y: 15}, rl.Rectangle{X: 5, Y: 15, Width: 35, Height: 45}},
		{"one edge only", handleLeft, rl.Vector2{X: 15, Y: 999}, rl.Rectangle{X: 15, Y: 20, Width: 25, Height: 40}},
		{"right past left", handleRight, rl.Vector2{X: 0, Y: 0}, rl.Rectangle{X: 10, Y: 20, Width: 1, Height: 40}},
		{"left past right", handleLeft, rl.Vector2{X: 100, Y: 0}, rl.Rectangle{X: 39, Y: 20, Width: 1, Height: 40}},
		{"top past bottom", handleTop, rl.Vector2{X: 0, Y: 100}, rl.Rectangle{X: 10, Y: 59, Width: 30, Height: 1}},
		{"bottom past top", handleBottom, rl.Vector2{X: 0, Y: -100}, rl.Rectangle{X: 10, Y: 20, Width: 30, Height: 1}},
	}
	for _, c := range cases {
		if got := dragRect(r, c.h, c.to); got != c.want {
			t.Errorf("%s: dragRect = %+v, want %+v", c.name, got, c.want)
		}
	}
}

// The grab offset keeps a rectangle from jumping to the cursor on the first
// frame of a drag.
func TestHandleAnchorKeepsGrabOffset(t *testing.T) {
	r := rl.Rectangle{X: 10, Y: 20, Width: 30, Height: 40}
	p := rl.Vector2{X: 25, Y: 40}
	if a := handleAnchor(r, handleBody, p); a != (rl.Vector2{X: 10, Y: 20}) {
		t.Errorf("body anchor = %v, want the top-left corner", a)
	}
	// dragging the body back to where it was grabbed must not move it
	off := sub(handleAnchor(r, handleBody, p), p)
	if got := dragRect(r, handleBody, rl.Vector2{X: p.X + off.X, Y: p.Y + off.Y}); got != r {
		t.Errorf("a body drag that has not moved the cursor moved the rectangle to %+v", got)
	}
}

// Rectangles are stored relative to their object and dragged in the level;
// the round trip has to be exact on the negative side of the origin too.
func TestRectLocalWorldRoundTrip(t *testing.T) {
	origin := rl.Vector2{X: -117, Y: 83}
	local := rl.Rectangle{X: 15, Y: 86, Width: 40, Height: 20}
	world := rectToWorld(local, origin)
	if world != (rl.Rectangle{X: -102, Y: 169, Width: 40, Height: 20}) {
		t.Errorf("rectToWorld = %+v", world)
	}
	if back := rectToLocal(world, origin); back != local {
		t.Errorf("rectToLocal(rectToWorld(r)) = %+v, want %+v", back, local)
	}
}

// A drag left of the origin must round to the nearest pixel. Truncating - what
// the place tool's cursor does - would land a pixel short every time.
func TestDragAtNegativeOriginRounds(t *testing.T) {
	le := newLevelEditor(39, 83, []LevelObject{{Resource: "x"}}, nil)
	le.objects[0] = LevelLayer{
		{X: -117, Y: 83}: {Resource: "x", X: -117, Y: 83, CollisionRect: &rl.Rectangle{X: 15, Y: 86, Width: 40, Height: 20}},
	}
	le.buildIndex()
	if !le.selectAt(Vector2i{X: -117, Y: 83}) {
		t.Fatal("could not select the tile")
	}
	// world rectangle is X -102..-62, Y 169..189; grab the body, off its corner
	if !le.beginDrag(rl.Vector2{X: -100, Y: 175}, 1) {
		t.Fatal("did not take hold of the body")
	}
	le.dragTo(rl.Vector2{X: -100.6, Y: 175})
	obj, _ := le.selectedObject()
	if obj.CollisionRect.X != 14 {
		t.Errorf("after dragging 0.6px left the rectangle is at local X %v, want 14 (rounded, not truncated)", obj.CollisionRect.X)
	}
	if obj.CollisionRect.Width != 40 || obj.CollisionRect.Height != 20 {
		t.Errorf("a body drag resized the rectangle to %vx%v", obj.CollisionRect.Width, obj.CollisionRect.Height)
	}
}

func TestGrabToleranceStaysOnScreenSize(t *testing.T) {
	if got := grabTolerance(1); got != 4 {
		t.Errorf("tolerance at zoom 1 = %v, want 4", got)
	}
	if got := grabTolerance(5); got != 1 {
		t.Errorf("tolerance at zoom 5 = %v, want the 1px floor", got)
	}
}
