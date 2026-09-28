package resistor

import (
	"encoding/json"
	"slices"
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// twoLights is an editor with a placed tile of one collision rectangle and two
// light rectangles far enough apart to click each one, the tile selected.
func twoLights() (*LevelEditor, Vector2i) {
	le := newLevelEditor(39, 83, []LevelObject{{Resource: "x"}}, nil)
	pos := Vector2i{X: 39}
	le.objects[-1] = LevelLayer{pos: {Resource: "x", X: 39,
		CollisionRects: []rl.Rectangle{rect(0, 50, 10, 10)},
		LightRects:     []rl.Rectangle{rect(0, 0, 10, 10), rect(20, 0, 10, 10)},
	}}
	le.buildIndex()
	le.selectAt(pos)
	return &le, pos
}

func TestActiveLightRectIndex(t *testing.T) {
	le, pos := twoLights()
	le.setActiveRect(rectLight)
	le.activeIdx = 1
	le.setActive(nil)
	if got := le.objects[-1][pos].LightRects; !slices.Equal(got, []rl.Rectangle{rect(0, 0, 10, 10)}) {
		t.Errorf("removing the second light rectangle left %+v", got)
	}
	if got := le.objects[-1][pos].CollisionRects; len(got) != 1 {
		t.Errorf("removing a light rectangle touched the collision ones: %+v", got)
	}

	// back to two, the second active, then collision: one rectangle, index 0
	le.objects[-1][pos] = LevelObject{Resource: "x", X: 39,
		CollisionRects: []rl.Rectangle{rect(0, 50, 10, 10)},
		LightRects:     []rl.Rectangle{rect(0, 0, 10, 10), rect(20, 0, 10, 10)},
	}
	le.activeIdx = 1
	le.setActiveRect(rectCollision)
	if le.activeIdx != 0 {
		t.Errorf("switching to collision left the index at %d", le.activeIdx)
	}

	// a selection with fewer light rectangles never points past its end
	le.setActiveRect(rectLight)
	le.activeIdx = 1
	other := Vector2i{X: 78}
	le.objects[-1][other] = LevelObject{Resource: "x", X: 78, LightRects: []rl.Rectangle{rect(0, 0, 5, 5)}}
	le.selectAt(other)
	if le.activeIdx != 0 {
		t.Errorf("a selection with one light rectangle left the index at %d", le.activeIdx)
	}
	if _, cur, _, ok := le.active(); !ok || cur == nil || *cur != rect(0, 0, 5, 5) {
		t.Errorf("the active rectangle is %+v", cur)
	}
	le.objects[-1][Vector2i{X: 117}] = LevelObject{Resource: "x", X: 117}
	le.selectAt(Vector2i{X: 117})
	if _, cur, _, ok := le.active(); !ok || cur != nil || le.activeIdx != 0 {
		t.Errorf("a selection with no light rectangle gave %+v at %d", cur, le.activeIdx)
	}
}

func TestClickPicksALightRect(t *testing.T) {
	le, pos := twoLights()
	// the second light rectangle is at world x 59..69; take hold of its body
	if !le.beginDrag(rl.Vector2{X: 64, Y: 5}, 1) {
		t.Fatal("a click on the second light rectangle took hold of nothing")
	}
	if le.activeRect != rectLight || le.activeIdx != 1 {
		t.Errorf("active = %d/%d, want light 1", le.activeRect, le.activeIdx)
	}
	le.dragTo(rl.Vector2{X: 66, Y: 5})
	obj := le.objects[-1][pos]
	if !slices.Equal(obj.LightRects, []rl.Rectangle{rect(0, 0, 10, 10), rect(22, 0, 10, 10)}) {
		t.Errorf("after dragging the second: %+v", obj.LightRects)
	}
	if obj.CollisionRects[0] != rect(0, 50, 10, 10) {
		t.Errorf("dragging a light rectangle moved the collision one: %+v", obj.CollisionRects)
	}
}

func TestStepAddRemoveLightRects(t *testing.T) {
	le, pos := twoLights()
	le.pickKind(rectLight)
	if le.activeRect != rectLight || le.activeIdx != 0 {
		t.Fatalf("L from collision gave %d/%d", le.activeRect, le.activeIdx)
	}
	var steps []int
	for range 2 {
		le.pickKind(rectLight)
		steps = append(steps, le.activeIdx)
	}
	if !slices.Equal(steps, []int{1, 0}) {
		t.Errorf("L stepped through %v, want 1 0", steps)
	}

	le.dirty = false
	le.addActiveRect()
	got := le.objects[-1][pos].LightRects
	if len(got) != 3 || got[2] != rect(0, 0, 39, 83) || got[0] != rect(0, 0, 10, 10) || le.activeIdx != 2 {
		t.Errorf("adding one gave %+v, active %d; want a grid cell at the origin, active", got, le.activeIdx)
	}
	if !le.dirty || len(le.objects[-1][pos].CollisionRects) != 1 {
		t.Errorf("adding a light rectangle is not an edit, or touched collision")
	}
	if n := len(le.LightOccluders()); n != 3 {
		t.Errorf("%d occluders, want one per light rectangle", n)
	}

	le.clearActiveRect() // the new one
	le.clearActiveRect() // the second of the two, which the first follows
	got = le.objects[-1][pos].LightRects
	if !slices.Equal(got, []rl.Rectangle{rect(0, 0, 10, 10)}) || le.activeIdx != 0 {
		t.Errorf("removing left %+v, active %d; want the first, active", got, le.activeIdx)
	}
	if occ := le.LightOccluders(); len(occ) != 1 || occ[0] != rect(39, 0, 10, 10) {
		t.Errorf("the remaining light rectangle casts %+v", occ)
	}
}

// Placing, defaults, following, copy and paste, and the level file all carry
// every light rectangle.
func TestEveryLightRectTravels(t *testing.T) {
	two := []rl.Rectangle{rect(0, 0, 10, 10), rect(20, 0, 10, 10)}
	three := append(slices.Clone(two), rect(40, 0, 10, 10))
	le := newLevelEditor(39, 83, []LevelObject{{Resource: "x", LightRects: coll(5, 5)}, {Resource: "y", LightRects: three}}, nil)
	a, a2 := Vector2i{X: 39}, Vector2i{X: 78}
	le.place(a) // both follow x
	le.place(a2)

	// following: a second light rectangle on the texture reaches its tiles
	le.setActiveRect(rectLight)
	le.addActiveRect()
	for _, p := range []Vector2i{a, a2} {
		if got := le.objects[-1][p].LightRects; len(got) != 2 || le.objects[-1][p].own {
			t.Errorf("a following tile has %+v (own %t) after its texture gained a light rectangle", got, le.objects[-1][p].own)
		}
	}
	le.markOwnRects()
	if le.objects[-1][a].own {
		t.Errorf("a tile with exactly its texture's two light rectangles does not follow it")
	}

	// no longer following: one tile gains a light rectangle of its own
	le.selectAt(a)
	le.addActiveRect()
	if got := le.objects[-1][a]; !got.own || len(got.LightRects) != 3 {
		t.Errorf("a tile given its own light rectangle has %+v (own %t)", got.LightRects, got.own)
	}
	if got := le.objects[-1][a2]; got.own || len(got.LightRects) != 2 {
		t.Errorf("the texture's other tile changed: %+v (own %t)", got.LightRects, got.own)
	}
	le.markOwnRects()
	if !le.objects[-1][a].own {
		t.Errorf("a tile with more light rectangles than its texture still follows it")
	}
	le.followTexture()
	if got := le.objects[-1][a]; got.own || !slices.Equal(got.LightRects, le.levelObjects[0].LightRects) {
		t.Errorf("F left %+v (own %t), want the texture's", got.LightRects, got.own)
	}

	// placing carries all three
	le.selectEntry(1)
	b := Vector2i{X: 117}
	le.place(b)
	if got := le.objects[-1][b].LightRects; !slices.Equal(got, three) {
		t.Errorf("placed with %+v, want %+v", got, three)
	}

	// copy and paste carry the two
	le.selectEntry(1)
	le.clip, le.hasClip = ObjectRects{Light: two}, true
	le.pasteSelection()
	if got := le.levelObjects[1].LightRects; !slices.Equal(got, two) {
		t.Errorf("pasted %+v, want %+v", got, two)
	}
	le.copySelection()
	if !slices.Equal(le.clip.Light, two) {
		t.Errorf("copied %+v", le.clip.Light)
	}

	// the texture defaults file carries all three
	defaults, _ := json.Marshal(map[string]ObjectRects{"y": {Light: three}})
	var back map[string]ObjectRects
	if err := json.Unmarshal(defaults, &back); err != nil || !slices.Equal(back["y"].Light, three) {
		t.Errorf("defaults saved and loaded %+v (%v)", back["y"].Light, err)
	}

	// the level file carries all three
	le.objects[-1][b] = LevelObject{Resource: "y", X: 117, LightRects: three}
	f, _ := le.marshalFile()
	objects, _, _, _, err := parseLevel(f)
	if err != nil || !slices.Equal(objects[-1][b].LightRects, three) {
		t.Errorf("saved and loaded %+v (%v)", objects[-1][b].LightRects, err)
	}
}

// Every light rectangle of every tile is an occluder, in the order the index
// visits the tiles and then the order of each tile's list.
func TestLightOccludersEveryRect(t *testing.T) {
	le := newLevelEditor(39, 83, nil, nil)
	le.objects[-1] = LevelLayer{
		{X: 0}:  {X: 0, LightRects: []rl.Rectangle{rect(1, 2, 3, 4)}},
		{X: 39}: {X: 39},
		{X: 78}: {X: 78, LightRects: []rl.Rectangle{rect(0, 0, 10, 20), rect(0, 40, 10, 20)}},
	}
	le.buildIndex()
	var want []rl.Rectangle
	for _, obj := range le.objectsIndexed {
		for _, l := range obj.LightRects {
			want = append(want, rl.Rectangle{X: float32(obj.X) + l.X, Y: float32(obj.Y) + l.Y, Width: l.Width, Height: l.Height})
		}
	}
	got := le.LightOccluders()
	if len(got) != 3 || !slices.Equal(got, want) {
		t.Errorf("occluders %+v, want %+v", got, want)
	}
	i := slices.Index(got, rect(78, 0, 10, 20))
	if i < 0 || i+1 >= len(got) || got[i+1] != rect(78, 40, 10, 20) {
		t.Errorf("the two-rectangle tile's occluders are not both there in list order: %+v", got)
	}

	// the level: one occluder per tile with a light rectangle, as before
	lvl := NewLevelEditor(39, 83)
	lvl.LoadLevel("level_1.json")
	n := 0
	for _, obj := range lvl.objectsIndexed {
		n += len(obj.LightRects)
		if len(obj.LightRects) > 1 {
			t.Skip("level_1.json now has a tile with several light rectangles")
		}
	}
	if len(lvl.LightOccluders()) != n || n == 0 {
		t.Errorf("level_1.json gives %d occluders for %d light rectangles", len(lvl.LightOccluders()), n)
	}
}
