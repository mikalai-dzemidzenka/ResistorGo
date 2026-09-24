package resistor

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// Drop-through rectangles go to disk the way the other kinds do, under keys of
// their own: one as "dropRect", several as "dropRects", none as neither.
func TestDropRectsOnDisk(t *testing.T) {
	two := []rl.Rectangle{rect(1, 2, 3, 4), rect(5, 6, 7, 8)}
	cases := []struct {
		name      string
		rects     []rl.Rectangle
		one, many bool
	}{
		{"none", nil, false, false},
		{"one", two[:1], true, false},
		{"two", two, false, true},
	}
	solid := []rl.Rectangle{rect(0, 0, 9, 9)}
	for _, c := range cases {
		obj := LevelObject{Resource: "r", X: 39, CollisionRects: solid, DropRects: c.rects}
		b, err := json.Marshal(obj)
		if err != nil {
			t.Fatal(err)
		}
		if got := strings.Contains(string(b), `"dropRect":`); got != c.one {
			t.Errorf("%s: dropRect written %t, want %t: %s", c.name, got, c.one, b)
		}
		if got := strings.Contains(string(b), `"dropRects":`); got != c.many {
			t.Errorf("%s: dropRects written %t, want %t: %s", c.name, got, c.many, b)
		}
		if !strings.Contains(string(b), `"collisionRect":{"X":0,"Y":0,"Width":9,"Height":9}`) {
			t.Errorf("%s: the collision rectangle is not written as before: %s", c.name, b)
		}
		var back LevelObject
		if err := json.Unmarshal(b, &back); err != nil {
			t.Fatal(err)
		}
		if !slices.Equal(back.DropRects, c.rects) || !slices.Equal(back.CollisionRects, solid) {
			t.Errorf("%s: round trip gave %+v", c.name, back)
		}

		r := ObjectRects{Collision: solid, Drop: c.rects}
		b, _ = json.Marshal(r)
		if got := strings.Contains(string(b), `"dropRects":`); got != c.many {
			t.Errorf("%s: defaults dropRects written %t, want %t: %s", c.name, got, c.many, b)
		}
		var rb ObjectRects
		if err := json.Unmarshal(b, &rb); err != nil || !sameRects(r, rb) {
			t.Errorf("%s: defaults round trip gave %+v (%v) from %s", c.name, rb, err, b)
		}
	}
}

// A drop-through rectangle is part of the level, so client and server that
// disagree about one have different fingerprints.
func TestDropRectChangesTheFingerprint(t *testing.T) {
	plain := newLevelEditor(39, 83, nil, nil)
	plain.objects[0] = LevelLayer{{}: {Resource: "a", CollisionRects: []rl.Rectangle{rect(0, 0, 39, 83)}}}
	with := newLevelEditor(39, 83, nil, nil)
	with.objects[0] = LevelLayer{{}: {Resource: "a", CollisionRects: []rl.Rectangle{rect(0, 0, 39, 83)}, DropRects: []rl.Rectangle{rect(0, 0, 39, 16)}}}
	if plain.Hash() == with.Hash() {
		t.Error("a drop-through rectangle left the fingerprint unchanged")
	}
}

// P picks the drop-through rectangles and steps through them; N adds one, a
// grid cell at the origin, and Delete takes the active one away.
func TestStepAddRemoveDropRects(t *testing.T) {
	le := newLevelEditor(39, 83, []LevelObject{{Resource: "x"}}, nil)
	pos := Vector2i{X: 39}
	le.objects[0] = LevelLayer{pos: {Resource: "x", X: 39,
		CollisionRects: []rl.Rectangle{rect(0, 40, 10, 10)},
		DropRects:      []rl.Rectangle{rect(0, 0, 10, 10), rect(20, 0, 10, 10)},
	}}
	le.buildIndex()
	le.selectAt(pos)

	le.pickKind(rectDrop)
	if le.activeRect != rectDrop || le.activeIdx != 0 {
		t.Fatalf("P from collision gave %d/%d", le.activeRect, le.activeIdx)
	}
	var steps []int
	for range 2 {
		le.pickKind(rectDrop)
		steps = append(steps, le.activeIdx)
	}
	if !slices.Equal(steps, []int{1, 0}) {
		t.Errorf("P stepped through %v, want 1 0", steps)
	}

	le.addActiveRect()
	obj := le.objects[0][pos]
	if len(obj.DropRects) != 3 || obj.DropRects[2] != rect(0, 0, 39, 83) || le.activeIdx != 2 {
		t.Errorf("adding one gave %+v, active %d", obj.DropRects, le.activeIdx)
	}
	if !slices.Equal(obj.CollisionRects, []rl.Rectangle{rect(0, 40, 10, 10)}) {
		t.Errorf("adding a drop rectangle changed the collision ones: %+v", obj.CollisionRects)
	}
	if n := slices.Index(le.drop, true); n < 0 || le.solids[n] != rect(39, 0, 10, 10) {
		t.Errorf("the drop rectangles are not in the level's solids: %+v %v", le.solids, le.drop)
	}

	le.activeIdx = 0
	le.clearActiveRect()
	if got := le.objects[0][pos].DropRects; !slices.Equal(got, []rl.Rectangle{rect(20, 0, 10, 10), rect(0, 0, 39, 83)}) {
		t.Errorf("removing the first left %+v", got)
	}
	for range 2 {
		le.clearActiveRect()
	}
	if got := le.objects[0][pos].DropRects; len(got) != 0 || slices.Contains(le.drop, true) {
		t.Errorf("clearing every drop rectangle left %+v, drop %v", got, le.drop)
	}
}

// A click on a drop-through rectangle takes hold of it, and dragging moves it
// alone.
func TestClickPicksADropRect(t *testing.T) {
	le := newLevelEditor(39, 83, []LevelObject{{Resource: "x"}}, nil)
	pos := Vector2i{X: 39}
	le.objects[0] = LevelLayer{pos: {Resource: "x", X: 39,
		CollisionRects: []rl.Rectangle{rect(0, 0, 10, 10)},
		DropRects:      []rl.Rectangle{rect(20, 0, 10, 10)},
	}}
	le.buildIndex()
	le.selectAt(pos)
	if !le.beginDrag(rl.Vector2{X: 64, Y: 5}, 1) {
		t.Fatal("a click on the drop rectangle took hold of nothing")
	}
	if le.activeRect != rectDrop || le.activeIdx != 0 {
		t.Errorf("active = %d/%d, want drop 0", le.activeRect, le.activeIdx)
	}
	le.dragTo(rl.Vector2{X: 66, Y: 5})
	obj := le.objects[0][pos]
	if obj.DropRects[0] != rect(22, 0, 10, 10) || obj.CollisionRects[0] != rect(0, 0, 10, 10) {
		t.Errorf("after dragging the drop rectangle: %+v %+v", obj.DropRects, obj.CollisionRects)
	}
}

// Placing, following, copy and paste, and the level file all carry a
// texture's drop-through rectangles.
func TestEveryDropRectTravels(t *testing.T) {
	d := []rl.Rectangle{rect(0, 0, 39, 16)}
	le := newLevelEditor(39, 83, []LevelObject{{Resource: "x"}, {Resource: "y", DropRects: d}}, nil)

	// placing carries it
	le.selectEntry(1)
	b := Vector2i{X: 78}
	le.place(b)
	if got := le.objects[0][b].DropRects; !slices.Equal(got, d) {
		t.Errorf("placed with %+v, want %+v", got, d)
	}

	// following: a drop rectangle added to the texture reaches its tile
	le.selectEntry(0)
	a := Vector2i{X: 39}
	le.place(a)
	le.setActiveRect(rectDrop)
	le.addActiveRect()
	if got := le.objects[0][a].DropRects; len(got) != 1 || le.objects[0][a].own {
		t.Errorf("a following tile has %+v (own %t) after its texture gained a drop rectangle", got, le.objects[0][a].own)
	}

	// a drop rectangle of its own detaches a tile
	le.selectAt(b)
	le.setActiveRect(rectDrop)
	le.addActiveRect()
	if !le.objects[0][b].own {
		t.Errorf("a tile given a drop rectangle of its own still follows its texture")
	}

	// copy and paste carry it
	le.selectEntry(1)
	le.copySelection()
	if !slices.Equal(le.clip.Drop, d) {
		t.Errorf("copied %+v", le.clip.Drop)
	}
	le.selectEntry(0)
	le.levelObjects[0].setRects(ObjectRects{})
	le.pasteSelection()
	if got := le.levelObjects[0].DropRects; !slices.Equal(got, d) {
		t.Errorf("pasted %+v, want %+v", got, d)
	}

	// the level file carries it
	le.objects[0][b] = LevelObject{Resource: "y", X: 78, CollisionRects: []rl.Rectangle{rect(0, 20, 39, 63)}, DropRects: d}
	f, _ := le.marshalFile()
	objects, _, _, _, err := parseLevel(f)
	if err != nil || !slices.Equal(objects[0][b].DropRects, d) || len(objects[0][b].CollisionRects) != 1 {
		t.Errorf("saved and loaded %+v (%v)", objects[0][b], err)
	}
}
