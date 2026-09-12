package resistor

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"testing"
)

func fighters(ids ...uint16) []*Actor {
	var out []*Actor
	for _, id := range ids {
		a := NewActor(rl.Vector2{X: float32(id) * 100}, false)
		a.ID = id
		out = append(out, a)
	}
	return out
}

func TestNextFighter(t *testing.T) {
	ids := []uint16{2, 5, 9}
	cur := uint16(2)
	var seen []uint16
	for range 3 {
		cur, _ = nextFighter(ids, cur, 1)
		seen = append(seen, cur)
	}
	if seen[0] != 5 || seen[1] != 9 || seen[2] != 2 {
		t.Errorf("next cycle %v, want [5 9 2]", seen)
	}
	if id, _ := nextFighter(ids, 2, -1); id != 9 {
		t.Errorf("prev from first = %d, want 9 (wrap)", id)
	}
	if id, _ := nextFighter(ids, 6, 1); id != 9 {
		t.Errorf("next after missing 6 = %d, want 9", id)
	}
	if id, _ := nextFighter(ids, 0, -1); id != 9 {
		t.Errorf("prev with none followed = %d, want 9", id)
	}
	if _, ok := nextFighter(nil, 1, 1); ok {
		t.Error("next in an empty list")
	}
}

func TestPanSameOnScreenAtAnyZoom(t *testing.T) {
	right := rl.Vector2{X: 1}
	a := panStep(rl.Vector2{}, right, MinZoom, 1)
	b := panStep(rl.Vector2{}, right, MaxZoom, 1)
	if a.X*MinZoom != SpecPanSpeed || b.X*MaxZoom != SpecPanSpeed {
		t.Errorf("on-screen distance %v at zoom %d and %v at zoom %d, want %d", a.X*MinZoom, MinZoom, b.X*MaxZoom, MaxZoom, SpecPanSpeed)
	}
}

func TestDragMovesLevelWithMouse(t *testing.T) {
	const zoom = 4
	p := dragStep(rl.Vector2{}, rl.Vector2{X: -100}, zoom)
	// the camera moves right by 100 screen px worth, so the level moves 100 px left on screen
	if p.X*zoom != 100 || p.Y != 0 {
		t.Errorf("drag: camera moved %v screen px, want 100", p.X*zoom)
	}
}

func TestSpecCamModes(t *testing.T) {
	const dt, zoom = float32(1.0 / 60), float32(3)
	c := SpecCam{Follow: true}

	// waits for the first fighter, then follows it
	c.Update(SpecInput{}, nil, zoom, dt, rl.Vector2{})
	if !c.Follow || c.FollowID != 0 {
		t.Fatalf("no fighters: %+v", c)
	}
	fs := fighters(3, 1, 2)
	c.Update(SpecInput{}, fs, zoom, dt, c.Pos)
	if c.FollowID != 1 || c.Pos != followPoint(fs[1]) {
		t.Fatalf("first visible fighter: %+v", c)
	}

	// E three times cycles every fighter and comes back
	var seen []uint16
	for range 3 {
		c.Update(SpecInput{Next: true}, fs, zoom, dt, c.Pos)
		seen = append(seen, c.FollowID)
	}
	if seen[0] != 2 || seen[1] != 3 || seen[2] != 1 {
		t.Errorf("E cycle %v, want [2 3 1]", seen)
	}

	// Space: free mode, the view does not move
	view := rl.Vector2{X: 77, Y: 33}
	c.Update(SpecInput{Free: true}, fs, zoom, dt, view)
	if c.Follow || c.Pos != view {
		t.Errorf("space: %+v, want free at %v", c, view)
	}

	// pan input while following: free mode, panning from where the view was
	c.Update(SpecInput{Next: true}, fs, zoom, dt, c.Pos)
	view = rl.Vector2{X: 500, Y: 20}
	c.Update(SpecInput{Pan: rl.Vector2{X: -1}}, fs, zoom, dt, view)
	if c.Follow || c.Pos.Y != view.Y || c.Pos.X >= view.X {
		t.Errorf("pan left from follow: %+v, want free just left of %v", c, view)
	}

	// Q from free mode enters follow mode again
	c.Update(SpecInput{Prev: true}, fs, zoom, dt, c.Pos)
	if !c.Follow {
		t.Error("Q did not enter follow mode")
	}

	// the followed fighter leaves: the next one is followed; when none are left, free mode
	c.FollowID = 2
	c.Update(SpecInput{}, fighters(1, 3), zoom, dt, c.Pos)
	if !c.Follow || c.FollowID != 3 {
		t.Errorf("after fighter 2 left: %+v, want following 3", c)
	}
	pos := c.Pos
	c.Update(SpecInput{}, nil, zoom, dt, pos)
	if c.Follow || c.Pos != pos {
		t.Errorf("after everyone left: %+v, want free at %v", c, pos)
	}
}
