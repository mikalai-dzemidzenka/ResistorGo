package resistor

import (
	"encoding/json"
	"os"
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func rectp(w, h float32) *rl.Rectangle { return &rl.Rectangle{Width: w, Height: h} }

// coll is a list of one w x h collision rectangle at the origin.
func coll(w, h float32) []rl.Rectangle { return []rl.Rectangle{{Width: w, Height: h}} }

func TestSameRects(t *testing.T) {
	a, a2, b := rectp(1, 2), rectp(1, 2), rectp(3, 4)
	cases := []struct {
		name string
		x, y ObjectRects
		want bool
	}{
		{"both empty", ObjectRects{}, ObjectRects{}, true},
		{"one missing", ObjectRects{Collision: []rl.Rectangle{*a}}, ObjectRects{}, false},
		{"equal by value", ObjectRects{Collision: []rl.Rectangle{*a}, Light: b}, ObjectRects{Collision: []rl.Rectangle{*a2}, Light: b}, true},
		{"differing", ObjectRects{Collision: []rl.Rectangle{*a}}, ObjectRects{Collision: []rl.Rectangle{*b}}, false},
		{"light missing on one", ObjectRects{Collision: []rl.Rectangle{*a}, Light: b}, ObjectRects{Collision: []rl.Rectangle{*a2}}, false},
		{"the same list", ObjectRects{Collision: []rl.Rectangle{*a, *b}}, ObjectRects{Collision: []rl.Rectangle{*a2, *b}}, true},
		{"order matters", ObjectRects{Collision: []rl.Rectangle{*a, *b}}, ObjectRects{Collision: []rl.Rectangle{*b, *a}}, false},
		{"one more", ObjectRects{Collision: []rl.Rectangle{*a}}, ObjectRects{Collision: []rl.Rectangle{*a, *b}}, false},
	}
	for _, c := range cases {
		if got := sameRects(c.x, c.y); got != c.want {
			t.Errorf("%s: sameRects = %t, want %t", c.name, got, c.want)
		}
	}
}

// Whether a tile follows its texture is never written to the level file.
func TestOwnIsNeverSaved(t *testing.T) {
	obj := LevelObject{Resource: "x", X: 39, CollisionRects: coll(5, 5)}
	plain, _ := json.Marshal(obj)
	obj.own = true
	flagged, _ := json.Marshal(obj)
	if string(plain) != string(flagged) {
		t.Errorf("the own flag reached the level file: %s vs %s", flagged, plain)
	}
}

// Loading classifies tiles - matching its texture follows, differing or of an
// unknown texture keeps its own - and changes no rectangle.
func TestMarkOwnRects(t *testing.T) {
	le := newLevelEditor(39, 83, []LevelObject{{Resource: "x", CollisionRects: coll(10, 10)}}, nil)
	match, differ, unknown := Vector2i{X: 39}, Vector2i{X: 78}, Vector2i{X: 117}
	le.objects[0] = LevelLayer{
		match:   {Resource: "x", X: 39, CollisionRects: coll(10, 10)},
		differ:  {Resource: "x", X: 78, CollisionRects: coll(11, 10)},
		unknown: {Resource: "y", X: 117, CollisionRects: coll(10, 10)},
	}
	before, _ := le.marshal()
	le.markOwnRects()
	if le.objects[0][match].own {
		t.Errorf("a tile matching its texture does not follow it")
	}
	if !le.objects[0][differ].own {
		t.Errorf("a tile differing from its texture follows it")
	}
	if !le.objects[0][unknown].own {
		t.Errorf("a tile of a texture with no palette entry follows it")
	}
	if after, _ := le.marshal(); string(after) != string(before) {
		t.Errorf("classifying the tiles changed the level")
	}
}

// Loading the real level gives the fingerprint of its bytes as they are, and
// its tiles - all laid from the wall's built-in rectangles - follow the wall.
func TestLoadLevelKeepsFingerprint(t *testing.T) {
	var b []byte
	var err error
	if DevTools {
		b, err = os.ReadFile("level_1.json")
	} else {
		b, err = assetFS.ReadFile("level_1.json")
	}
	if err != nil {
		t.Fatal(err)
	}
	plain := newLevelEditor(39, 83, nil, nil)
	objects, _, err := parseLevel(b)
	if err != nil {
		t.Fatal(err)
	}
	plain.objects = objects

	le := NewLevelEditor(39, 83)
	le.LoadLevel("level_1.json")
	if le.Hash() != plain.Hash() {
		t.Errorf("loading changed the level's fingerprint")
	}
	follows := 0
	for _, obj := range le.objectsIndexed {
		if !obj.own {
			follows++
		}
	}
	if follows == 0 {
		t.Errorf("no tile of the level follows its texture")
	}
}

// A placed tile follows until it is edited by hand - any edit.
func TestPlacedTileFollowsUntilEdited(t *testing.T) {
	le := newLevelEditor(39, 83, []LevelObject{{Resource: "x", CollisionRects: coll(10, 10)}}, nil)
	edits := map[string]func(){
		"reshape": func() { le.setActive(rectp(12, 10)) },
		"add":     func() { le.setActiveRect(rectLight); le.addActiveRect(); le.setActiveRect(rectCollision) },
		"clear":   func() { le.clearActiveRect() },
		"paste":   func() { le.clip, le.hasClip = ObjectRects{Collision: coll(3, 3)}, true; le.pasteSelection() },
	}
	x := int32(0)
	for name, edit := range edits {
		x += 39
		pos := Vector2i{X: x}
		le.selectEntry(0)
		le.place(pos)
		if le.objects[0][pos].own {
			t.Fatalf("a newly placed tile does not follow its texture")
		}
		le.selectAt(pos)
		edit()
		if !le.objects[0][pos].own {
			t.Errorf("%s: a tile edited by hand still follows its texture", name)
		}
	}
}

// A texture's change reaches its following tiles, and nothing else.
func TestTextureChangeReachesFollowingTiles(t *testing.T) {
	le := newLevelEditor(39, 83, []LevelObject{
		{Resource: "x", CollisionRects: coll(10, 10)},
		{Resource: "y", CollisionRects: coll(10, 10)},
	}, nil)
	follow, own, other := Vector2i{X: 39}, Vector2i{X: 78}, Vector2i{X: 117}
	follow2 := Vector2i{X: 156}
	le.place(follow)
	le.place(own)
	le.place(follow2)
	le.selectEntry(1)
	le.place(other)
	le.selectAt(own)
	le.setActive(rectp(5, 5))

	le.selectEntry(0)
	le.setActive(rectp(20, 20))
	for _, p := range []Vector2i{follow, follow2} {
		if got := le.objects[0][p].CollisionRects[0]; got != *rectp(20, 20) {
			t.Errorf("a following tile has %+v after its texture changed", got)
		}
	}
	if got := le.objects[0][own].CollisionRects[0]; got != *rectp(5, 5) {
		t.Errorf("a hand-edited tile took its texture's change: %+v", got)
	}
	if got := le.objects[0][other].CollisionRects[0]; got != *rectp(10, 10) {
		t.Errorf("a tile of another texture changed: %+v", got)
	}
	a, b := &le.objects[0][follow].CollisionRects[0], &le.objects[0][follow2].CollisionRects[0]
	if a == b || a == &le.levelObjects[0].CollisionRects[0] {
		t.Errorf("following tiles share a rectangle with each other or with the entry")
	}
	if len(le.LightOccluders()) != 0 {
		t.Errorf("tiles grew light rectangles nobody gave them")
	}
}

func TestFollowTexture(t *testing.T) {
	le := newLevelEditor(39, 83, []LevelObject{{Resource: "x", CollisionRects: coll(10, 10)}}, nil)
	pos := Vector2i{X: 39}
	le.place(pos)
	le.selectAt(pos)
	le.setActive(rectp(5, 5))
	le.dirty = false

	le.followTexture()
	obj := le.objects[0][pos]
	if obj.own || obj.CollisionRects[0] != *rectp(10, 10) {
		t.Fatalf("after F the tile has %+v, own %t; want its texture's rectangle, following", obj.CollisionRects, obj.own)
	}
	if !le.dirty {
		t.Errorf("going back to the texture's rectangles is not an unsaved edit")
	}
	le.selectEntry(0)
	le.setActive(rectp(30, 30))
	if got := le.objects[0][pos].CollisionRects[0]; got != *rectp(30, 30) {
		t.Errorf("after F the tile did not take its texture's next change: %+v", got)
	}

	// already following, a palette selection, an unknown texture: nothing
	le.selectAt(pos)
	le.dirty = false
	le.followTexture()
	le.selectEntry(0)
	le.followTexture()
	stray := Vector2i{X: 390}
	le.objects[0][stray] = LevelObject{Resource: "gone", X: 390, CollisionRects: coll(1, 1), own: true}
	le.selectAt(stray)
	le.followTexture()
	if le.dirty || !le.objects[0][stray].own {
		t.Errorf("F acted where there was nothing for it to do")
	}
}

func TestCopyPaste(t *testing.T) {
	src := LevelObject{Resource: "a", CollisionRects: coll(10, 10)} // no light
	dst := LevelObject{Resource: "b", CollisionRects: coll(1, 1), LightRect: rectp(2, 2)}
	le := newLevelEditor(39, 83, []LevelObject{src, dst}, nil)

	le.pasteSelection()
	if le.dirty {
		t.Errorf("a paste with nothing copied changed something")
	}

	le.selectEntry(1)
	tile := Vector2i{X: 39}
	le.place(tile)

	le.selectEntry(0)
	le.copySelection()
	le.selectEntry(1)
	le.pasteSelection()
	b := le.levelObjects[1]
	if len(b.CollisionRects) != 1 || b.CollisionRects[0] != *rectp(10, 10) {
		t.Errorf("pasted collision = %+v, want the source's", b.CollisionRects)
	}
	if b.LightRect != nil {
		t.Errorf("the source had no light rectangle, but the target kept one")
	}
	if got := le.objects[0][tile]; got.own || got.CollisionRects[0] != *rectp(10, 10) || got.LightRect != nil {
		t.Errorf("a following tile of the pasted-onto texture did not take the paste: %+v", got)
	}

	le.selectEntry(0)
	le.setActive(rectp(40, 40))
	if got := le.levelObjects[1].CollisionRects[0]; got != *rectp(10, 10) {
		t.Errorf("changing the source after pasting changed the target: %+v", got)
	}

	le.dirty = false
	le.selectEntry(1)
	le.pasteSelection() // the same rectangles again
	if le.dirty {
		t.Errorf("a paste that changed nothing is an unsaved edit")
	}

	le.selectAt(tile)
	le.clip = ObjectRects{Collision: coll(7, 7)}
	le.pasteSelection()
	if got := le.objects[0][tile]; !got.own || got.CollisionRects[0] != *rectp(7, 7) {
		t.Errorf("pasting onto a tile did not give it the rectangles and detach it: %+v", got)
	}
	le.selectEntry(1)
	le.setActive(rectp(50, 50))
	if got := le.objects[0][tile].CollisionRects[0]; got != *rectp(7, 7) {
		t.Errorf("a tile pasted onto took its texture's next change: %+v", got)
	}
}
