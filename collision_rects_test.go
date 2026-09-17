package resistor

import (
	"encoding/json"
	"slices"
	"strconv"
	"strings"
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func rect(x, y, w, h float32) rl.Rectangle { return rl.Rectangle{X: x, Y: y, Width: w, Height: h} }

// legacyLevelObject is LevelObject as it was when an object had at most one
// collision rectangle. A level that uses no more than that must still be
// written exactly as this writes it.
type legacyLevelObject struct {
	Resource      string        `json:"resource,omitempty"`
	Layer         int           `json:"layer,omitempty"`
	X             int32         `json:"x,omitempty"`
	Y             int32         `json:"y,omitempty"`
	CollisionRect *rl.Rectangle `json:"collisionRect,omitempty"`
	LightRect     *rl.Rectangle `json:"lightRect,omitempty"`
}

func TestCollisionRectsOnDisk(t *testing.T) {
	three := []rl.Rectangle{rect(1, 2, 3, 4), rect(5, 6, 7, 8), rect(-9, 0, 1, 1)}
	cases := []struct {
		name      string
		rects     []rl.Rectangle
		one, many bool
	}{
		{"none", nil, false, false},
		{"one", three[:1], true, false},
		{"three", three, false, true},
	}
	for _, c := range cases {
		obj := LevelObject{Resource: "r", X: 39, CollisionRects: c.rects, LightRects: []rl.Rectangle{{Width: 2, Height: 2}}}
		b, err := json.Marshal(obj)
		if err != nil {
			t.Fatal(err)
		}
		if got := strings.Contains(string(b), `"collisionRect":`); got != c.one {
			t.Errorf("%s: collisionRect written %t, want %t: %s", c.name, got, c.one, b)
		}
		if got := strings.Contains(string(b), `"collisionRects":`); got != c.many {
			t.Errorf("%s: collisionRects written %t, want %t: %s", c.name, got, c.many, b)
		}
		var back LevelObject
		if err := json.Unmarshal(b, &back); err != nil {
			t.Fatal(err)
		}
		if !slices.Equal(back.CollisionRects, c.rects) || back.X != 39 || len(back.LightRects) != 1 {
			t.Errorf("%s: round trip gave %+v", c.name, back)
		}

		r := ObjectRects{Collision: c.rects}
		b, _ = json.Marshal(r)
		var rb ObjectRects
		if err := json.Unmarshal(b, &rb); err != nil || !sameRects(r, rb) {
			t.Errorf("%s: defaults round trip gave %+v (%v) from %s", c.name, rb, err, b)
		}
	}

	var both LevelObject
	if err := json.Unmarshal([]byte(`{"collisionRect":{"X":1,"Y":1,"Width":1,"Height":1},"collisionRects":[{"X":2,"Y":2,"Width":2,"Height":2},{"X":3,"Y":3,"Width":3,"Height":3}]}`), &both); err != nil {
		t.Fatal(err)
	}
	if len(both.CollisionRects) != 2 || both.CollisionRects[0].X != 2 {
		t.Errorf("with both keys the list should win, got %+v", both.CollisionRects)
	}
}

// Light rectangles go to disk the way collision rectangles do: one as
// "lightRect", several as "lightRects", none as neither.
func TestLightRectsOnDisk(t *testing.T) {
	three := []rl.Rectangle{rect(1, 2, 3, 4), rect(5, 6, 7, 8), rect(-9, 0, 1, 1)}
	cases := []struct {
		name      string
		rects     []rl.Rectangle
		one, many bool
	}{
		{"none", nil, false, false},
		{"one", three[:1], true, false},
		{"three", three, false, true},
	}
	for _, c := range cases {
		obj := LevelObject{Resource: "r", X: 39, CollisionRects: three[:1], LightRects: c.rects}
		b, err := json.Marshal(obj)
		if err != nil {
			t.Fatal(err)
		}
		if got := strings.Contains(string(b), `"lightRect":`); got != c.one {
			t.Errorf("%s: lightRect written %t, want %t: %s", c.name, got, c.one, b)
		}
		if got := strings.Contains(string(b), `"lightRects":`); got != c.many {
			t.Errorf("%s: lightRects written %t, want %t: %s", c.name, got, c.many, b)
		}
		var back LevelObject
		if err := json.Unmarshal(b, &back); err != nil {
			t.Fatal(err)
		}
		if !slices.Equal(back.LightRects, c.rects) || !slices.Equal(back.CollisionRects, three[:1]) {
			t.Errorf("%s: round trip gave %+v", c.name, back)
		}

		r := ObjectRects{Collision: three[:1], Light: c.rects}
		b, _ = json.Marshal(r)
		if got := strings.Contains(string(b), `"lightRects":`); got != c.many {
			t.Errorf("%s: defaults lightRects written %t, want %t: %s", c.name, got, c.many, b)
		}
		var rb ObjectRects
		if err := json.Unmarshal(b, &rb); err != nil || !sameRects(r, rb) {
			t.Errorf("%s: defaults round trip gave %+v (%v) from %s", c.name, rb, err, b)
		}
	}

	const bothKeys = `{"lightRect":{"X":1,"Y":1,"Width":1,"Height":1},"lightRects":[{"X":2,"Y":2,"Width":2,"Height":2},{"X":3,"Y":3,"Width":3,"Height":3}]}`
	var both LevelObject
	if err := json.Unmarshal([]byte(bothKeys), &both); err != nil {
		t.Fatal(err)
	}
	if len(both.LightRects) != 2 || both.LightRects[0].X != 2 {
		t.Errorf("with both keys the list should win, got %+v", both.LightRects)
	}
	var bothDefaults ObjectRects
	if err := json.Unmarshal([]byte(bothKeys), &bothDefaults); err != nil {
		t.Fatal(err)
	}
	if len(bothDefaults.Light) != 2 || bothDefaults.Light[0].X != 2 {
		t.Errorf("with both keys the defaults' list should win, got %+v", bothDefaults.Light)
	}
}

// Only the kind with several rectangles is written as a list.
func TestOnlyTheLightRectsBecomeAList(t *testing.T) {
	obj := LevelObject{Resource: "a", CollisionRects: []rl.Rectangle{rect(1, 1, 1, 1)}, LightRects: []rl.Rectangle{rect(2, 2, 2, 2), rect(3, 3, 3, 3)}}
	b, _ := json.Marshal(obj)
	want := `{"resource":"a","collisionRect":{"X":1,"Y":1,"Width":1,"Height":1},"lightRects":[{"X":2,"Y":2,"Width":2,"Height":2},{"X":3,"Y":3,"Width":3,"Height":3}]}`
	if string(b) != want {
		t.Errorf("written as %s, want %s", b, want)
	}
}

// One rectangle is written exactly as before objects could have several.
func TestOneCollisionRectKeepsItsBytes(t *testing.T) {
	r, l := rect(15, 86, 40, 20), rect(22, 90, 40, 15)
	cases := []struct {
		obj LevelObject
		old legacyLevelObject
	}{
		{LevelObject{Resource: "a", X: -117, Y: 83, CollisionRects: []rl.Rectangle{r}, LightRects: []rl.Rectangle{l}},
			legacyLevelObject{Resource: "a", X: -117, Y: 83, CollisionRect: &r, LightRect: &l}},
		{LevelObject{Resource: "b", Layer: 2}, legacyLevelObject{Resource: "b", Layer: 2}},
	}
	for _, c := range cases {
		got, _ := json.Marshal(c.obj)
		want, _ := json.Marshal(c.old)
		if string(got) != string(want) {
			t.Errorf("written as %s, before as %s", got, want)
		}
	}
}

// The level loads and saves to its own bytes and, while its tiles have no more
// than one collision rectangle each, keeps the fingerprint the old format gave it.
func TestLevel1KeepsItsBytes(t *testing.T) {
	b := level1Bytes(t)
	le := NewLevelEditor(39, 83)
	le.LoadLevel("level_1.json")
	got, err := le.marshalFile()
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(b) {
		t.Errorf("saving the level unedited changed it (%d bytes, was %d)", len(got), len(b))
	}
	for _, layer := range le.objects {
		for _, obj := range layer {
			if len(obj.CollisionRects) > 1 || len(obj.LightRects) > 1 {
				return // the old format cannot hold this level
			}
		}
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(b, &raw); err != nil {
		t.Fatal(err)
	}
	legacy := make(map[int]map[string]legacyLevelObject)
	for k, v := range raw {
		n, err := strconv.Atoi(k)
		if err != nil {
			continue // the lights
		}
		var layer map[string]legacyLevelObject
		if err := json.Unmarshal(v, &layer); err != nil {
			t.Fatal(err)
		}
		legacy[n] = layer
	}
	lb, _ := json.Marshal(legacy)
	if levelHash(lb) != le.Hash() {
		t.Errorf("the level's fingerprint is not the one the old format gave it")
	}
}

func TestSolidsFollowTheCollisionOrder(t *testing.T) {
	le := newLevelEditor(39, 83, nil, nil)
	one := []rl.Rectangle{rect(15, 86, 40, 20)}
	le.objects[0] = LevelLayer{
		{X: 0}:         {X: 0, CollisionRects: one},
		{X: 39}:        {X: 39, CollisionRects: one},
		{X: 78, Y: 83}: {X: 78, Y: 83, CollisionRects: []rl.Rectangle{rect(0, 0, 5, 5), rect(10, 0, 5, 5), rect(20, 0, 5, 5)}},
		{X: 117}:       {X: 117}, // no collision
	}
	le.buildIndex()
	// objectsIndexed: Y descending, then X descending
	want := []rl.Rectangle{
		rect(78, 83, 5, 5), rect(88, 83, 5, 5), rect(98, 83, 5, 5),
		rect(54, 86, 40, 20),
		rect(15, 86, 40, 20),
	}
	if !slices.Equal(le.solids, want) {
		t.Errorf("solids = %+v\nwant     %+v", le.solids, want)
	}
}

// gapLevel installs a tile with two collision rectangles, x 0..40 and 80..120,
// both y 100..120, and returns a function that restores the level before it.
func gapLevel() func() {
	saved := LE
	LE = newLevelEditor(39, 83, nil, nil)
	LE.objects[0] = LevelLayer{{}: {CollisionRects: []rl.Rectangle{rect(0, 100, 40, 20), rect(80, 100, 40, 20)}}}
	LE.buildIndex()
	return func() { LE = saved }
}

// standsAt reports whether a fighter centered on x, its feet just above y,
// is on the ground.
func standsAt(x, y float32) bool {
	a := NewActor(rl.Vector2{}, true)
	a.Pos.X = x - (a.CollisionBox.X + a.CollisionBox.Width/2)
	a.Pos.Y = y - (a.CollisionBox.Y + a.CollisionBox.Height) - 0.5
	a.checkCollision()
	return a.OnGround
}

func TestStandOnEitherCollisionRect(t *testing.T) {
	defer gapLevel()()
	if !standsAt(20, 100) {
		t.Errorf("a fighter over the first rectangle is not standing on it")
	}
	if standsAt(60, 100) {
		t.Errorf("a fighter over the gap is standing on something")
	}
	if !standsAt(100, 100) {
		t.Errorf("a fighter over the second rectangle is not standing on it")
	}
}

func TestLedgeOnTheSecondCollisionRect(t *testing.T) {
	defer gapLevel()()
	a := jumper(70, 102) // rising, head level with the second rectangle, just short of its edge
	a.checkCollision()
	if !a.CanClimbRight {
		t.Errorf("a fighter jumping at the second rectangle's edge did not catch the ledge")
	}
}

// threeRects is an editor with a placed tile of three collision rectangles
// far enough apart to click each one, the tile selected.
func threeRects() (*LevelEditor, Vector2i) {
	le := newLevelEditor(39, 83, []LevelObject{{Resource: "x"}}, nil)
	pos := Vector2i{X: 39}
	le.objects[0] = LevelLayer{pos: {Resource: "x", X: 39, CollisionRects: []rl.Rectangle{
		rect(0, 0, 10, 10), rect(20, 0, 10, 10), rect(40, 0, 10, 10),
	}}}
	le.buildIndex()
	le.selectAt(pos)
	return &le, pos
}

func TestActiveCollisionRectIndex(t *testing.T) {
	le, pos := threeRects()
	le.activeIdx = 1
	le.setActive(nil)
	if got := le.objects[0][pos].CollisionRects; !slices.Equal(got, []rl.Rectangle{rect(0, 0, 10, 10), rect(40, 0, 10, 10)}) {
		t.Errorf("removing the second left %+v", got)
	}

	le.activeIdx = 2
	le.objects[0][Vector2i{X: 78}] = LevelObject{Resource: "x", X: 78, CollisionRects: []rl.Rectangle{rect(0, 0, 5, 5)}}
	le.selectAt(Vector2i{X: 78})
	if le.activeIdx != 0 {
		t.Errorf("a selection with one rectangle left the index at %d", le.activeIdx)
	}
	if _, cur, _, ok := le.active(); !ok || cur == nil || *cur != rect(0, 0, 5, 5) {
		t.Errorf("the active rectangle is %+v", cur)
	}
}

func TestClickPicksACollisionRect(t *testing.T) {
	le, pos := threeRects()
	// the third is at world x 79..89; take hold of its body
	if !le.beginDrag(rl.Vector2{X: 84, Y: 5}, 1) {
		t.Fatal("a click on the third rectangle took hold of nothing")
	}
	if le.activeRect != rectCollision || le.activeIdx != 2 {
		t.Errorf("active = %d/%d, want collision 2", le.activeRect, le.activeIdx)
	}
	le.dragTo(rl.Vector2{X: 86, Y: 5})
	got := le.objects[0][pos].CollisionRects
	if got[2] != rect(42, 0, 10, 10) || got[0] != rect(0, 0, 10, 10) || got[1] != rect(20, 0, 10, 10) {
		t.Errorf("after dragging the third: %+v", got)
	}

	// two rectangles on the same spot: the one already active wins the click
	le.objects[0][pos] = LevelObject{Resource: "x", X: 39, CollisionRects: []rl.Rectangle{rect(0, 0, 10, 10), rect(0, 0, 10, 10)}}
	le.drag, le.activeIdx = rectDrag{}, 1
	if !le.beginDrag(rl.Vector2{X: 44, Y: 5}, 1) || le.activeIdx != 1 {
		t.Errorf("on an overlap the click went to %d, want the active 1", le.activeIdx)
	}
}

func TestStepAddRemoveCollisionRects(t *testing.T) {
	le, pos := threeRects()
	le.setActiveRect(rectLight)
	le.pickKind(rectCollision)
	if le.activeRect != rectCollision || le.activeIdx != 0 {
		t.Fatalf("C from light gave %d/%d", le.activeRect, le.activeIdx)
	}
	var steps []int
	for range 3 {
		le.pickKind(rectCollision)
		steps = append(steps, le.activeIdx)
	}
	if !slices.Equal(steps, []int{1, 2, 0}) {
		t.Errorf("C stepped through %v, want 1 2 0", steps)
	}

	le.activeIdx = 1
	le.clearActiveRect()
	got := le.objects[0][pos].CollisionRects
	if !slices.Equal(got, []rl.Rectangle{rect(0, 0, 10, 10), rect(40, 0, 10, 10)}) || le.activeIdx != 0 {
		t.Errorf("removing the second of three left %+v, active %d", got, le.activeIdx)
	}

	le.addActiveRect()
	got = le.objects[0][pos].CollisionRects
	if len(got) != 3 || got[2] != rect(0, 0, 39, 83) || got[0] != rect(0, 0, 10, 10) || le.activeIdx != 2 {
		t.Errorf("adding one gave %+v, active %d; want a grid cell at the origin, active", got, le.activeIdx)
	}
	if !le.objects[0][pos].own || !le.dirty {
		t.Errorf("adding a rectangle to a tile is not a hand edit")
	}
}

// Placing, defaults, following, copy and paste, and the level file all carry
// every collision rectangle.
func TestEveryCollisionRectTravels(t *testing.T) {
	two := []rl.Rectangle{rect(0, 0, 10, 10), rect(20, 0, 10, 10)}
	three := append(slices.Clone(two), rect(40, 0, 10, 10))
	le := newLevelEditor(39, 83, []LevelObject{{Resource: "x", CollisionRects: coll(5, 5)}, {Resource: "y", CollisionRects: three}}, nil)
	a := Vector2i{X: 39}
	le.place(a) // follows x

	// following: a second rectangle on the texture reaches its tile
	le.addActiveRect()
	if got := le.objects[0][a].CollisionRects; len(got) != 2 || le.objects[0][a].own {
		t.Errorf("a following tile has %+v (own %t) after its texture gained a rectangle", got, le.objects[0][a].own)
	}
	le.markOwnRects()
	if le.objects[0][a].own {
		t.Errorf("a tile with exactly its texture's two rectangles does not follow it")
	}

	// placing carries all three
	le.selectEntry(1)
	b := Vector2i{X: 78}
	le.place(b)
	if got := le.objects[0][b].CollisionRects; !slices.Equal(got, three) {
		t.Errorf("placed with %+v, want %+v", got, three)
	}

	// copy and paste carry the two
	le.selectEntry(1)
	le.clip, le.hasClip = ObjectRects{Collision: two}, true
	le.pasteSelection()
	if got := le.levelObjects[1].CollisionRects; !slices.Equal(got, two) {
		t.Errorf("pasted %+v, want %+v", got, two)
	}
	le.copySelection()
	if !slices.Equal(le.clip.Collision, two) {
		t.Errorf("copied %+v", le.clip.Collision)
	}

	// the level file carries all three
	le.objects[0][b] = LevelObject{Resource: "y", X: 78, CollisionRects: three}
	f, _ := le.marshalFile()
	objects, _, _, _, err := parseLevel(f)
	if err != nil || !slices.Equal(objects[0][b].CollisionRects, three) {
		t.Errorf("saved and loaded %+v (%v)", objects[0][b].CollisionRects, err)
	}
}
