package resistor

import (
	"path/filepath"
	"slices"
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func paletteEntry(t *testing.T, palette []LevelObject, name string) LevelObject {
	t.Helper()
	for _, obj := range palette {
		if obj.Resource == name {
			return obj
		}
	}
	t.Fatalf("%s is not in the palette", name)
	return LevelObject{}
}

// A palette entry takes its rectangles from the defaults file, then from the
// rectangles built into the game, and otherwise has none.
func TestBuildPaletteResolvesRects(t *testing.T) {
	wall := blocksDir + "/e1m1_wall1.png"
	door := blocksDir + "/e1m1_door1.png"
	bg := blocksDir + "/e1m1_bg1.png"
	defaults := map[string]ObjectRects{
		door: {Collision: []rl.Rectangle{{X: 1, Y: 2, Width: 3, Height: 4}}},
	}
	palette := buildPalette(defaults)
	builtin := builtinRects()

	w := paletteEntry(t, palette, wall)
	if len(w.CollisionRects) != 1 || !slices.Equal(w.CollisionRects, builtin[wall].Collision) ||
		len(w.LightRects) != 1 || !slices.Equal(w.LightRects, builtin[wall].Light) {
		t.Errorf("wall did not keep its built-in rectangles: %+v", w)
	}
	if &w.CollisionRects[0] == &builtin[wall].Collision[0] {
		t.Errorf("the palette shares its rectangle with the built-in table")
	}
	d := paletteEntry(t, palette, door)
	if len(d.CollisionRects) != 1 || !slices.Equal(d.CollisionRects, defaults[door].Collision) {
		t.Errorf("door collision = %+v, want the defaults file's", d.CollisionRects)
	}
	if len(d.LightRects) != 0 {
		t.Errorf("door has a light rectangle the defaults file does not give it")
	}
	b := paletteEntry(t, palette, bg)
	if len(b.CollisionRects) != 0 || len(b.LightRects) != 0 {
		t.Errorf("a texture with no rectangles anywhere was given some: %+v", b)
	}
}

// An empty palette is a level with nothing to place, not a panic.
func TestEmptyPaletteIsUsable(t *testing.T) {
	le := newLevelEditor(39, 83, nil, nil)
	le.place(Vector2i{})
	if len(le.objectsIndexed) != 0 {
		t.Errorf("an empty palette placed %d objects", len(le.objectsIndexed))
	}
	if _, ok := le.selectedObject(); ok {
		t.Errorf("an empty palette reports a selection")
	}
	le.addActiveRect()
	le.clearActiveRect()
	if le.beginDrag(rl.Vector2{}, 1) {
		t.Errorf("a drag started with nothing selected")
	}
}

// Placing copies the entry's rectangles, including their absence. The door and
// platform have no light rectangle: placing one used to dereference it.
func TestPlaceCopiesRectsIncludingNone(t *testing.T) {
	entry := LevelObject{Resource: "x", CollisionRects: []rl.Rectangle{{Width: 5, Height: 5}}}
	le := newLevelEditor(39, 83, []LevelObject{entry}, nil)
	le.place(Vector2i{X: 39})
	obj := le.objects[0][Vector2i{X: 39}]
	if len(obj.LightRects) != 0 {
		t.Errorf("an entry with no light rectangle placed an object with one")
	}
	if len(obj.CollisionRects) != 1 || !slices.Equal(obj.CollisionRects, entry.CollisionRects) {
		t.Errorf("placed collision = %+v, want %+v", obj.CollisionRects, entry.CollisionRects)
	}
	if &obj.CollisionRects[0] == &le.obj.CollisionRects[0] {
		t.Errorf("the placed object shares its rectangle with the palette entry")
	}
}

// A placed object is still the selection after other edits rebuild the index.
func TestSelectionSurvivesRebuild(t *testing.T) {
	le := newLevelEditor(39, 83, []LevelObject{{Resource: "x"}}, nil)
	a := Vector2i{X: 39, Y: 83}
	le.place(a)
	if !le.selectAt(a) {
		t.Fatal("could not select the placed object")
	}
	le.place(Vector2i{X: 78, Y: 83})
	le.erase(Vector2i{X: 78, Y: 83})
	obj, ok := le.selectedObject()
	if !ok || obj.X != a.X || obj.Y != a.Y {
		t.Errorf("after the index was rebuilt the selection is %+v (ok %t), want the object at %v", obj, ok, a)
	}
}

// A click on a placed object selects it; a click on nothing leaves the
// selection as it was.
func TestSelectAt(t *testing.T) {
	le := newLevelEditor(39, 83, []LevelObject{{Resource: "x"}}, nil)
	a := Vector2i{X: 39}
	le.place(a)
	if !le.sel.proto {
		t.Fatal("the palette entry should start selected")
	}
	if le.selectAt(Vector2i{X: 390}) || !le.sel.proto {
		t.Errorf("a click on an empty cell changed the selection")
	}
	if !le.selectAt(a) || le.sel.proto || le.sel.pos != a {
		t.Errorf("a click on a placed object did not select it: %+v", le.sel)
	}
	if le.selectAt(Vector2i{X: 390}) || le.sel.pos != a {
		t.Errorf("a click on an empty cell dropped the selected tile")
	}
	le.selectEntry(0)
	if !le.sel.proto {
		t.Errorf("picking a palette entry did not select it")
	}
}

// A new rectangle is one grid cell at the object's origin; clearing leaves
// none - an object that is not stood on, or casts no shadow.
func TestAddAndClearRect(t *testing.T) {
	le := newLevelEditor(39, 83, []LevelObject{{Resource: "x"}}, nil)
	a := Vector2i{X: 39}
	le.place(a)
	le.selectAt(a)
	le.setActiveRect(rectLight)
	le.dirty = false

	le.addActiveRect()
	obj, _ := le.selectedObject()
	if len(obj.LightRects) != 1 || obj.LightRects[0] != (rl.Rectangle{Width: 39, Height: 83}) {
		t.Fatalf("added light rectangles = %+v, want one 39x83 cell at the origin", obj.LightRects)
	}
	if !le.dirty {
		t.Errorf("adding a rectangle is not an unsaved edit")
	}
	if len(le.LightOccluders()) != 1 {
		t.Errorf("the new light rectangle casts no shadow")
	}

	le.clearActiveRect()
	obj, _ = le.selectedObject()
	if len(obj.LightRects) != 0 {
		t.Errorf("cleared light rectangle is still %+v", obj.LightRects)
	}
	if len(le.LightOccluders()) != 0 {
		t.Errorf("a tile with no light rectangle still casts a shadow")
	}
}

// Every rectangle edit is an unsaved edit, and one that changes nothing is not.
func TestRectEditMarksDirty(t *testing.T) {
	le := newLevelEditor(39, 83, []LevelObject{{Resource: "x", CollisionRects: []rl.Rectangle{{Width: 10, Height: 10}}}}, nil)
	le.setActive(&rl.Rectangle{Width: 10, Height: 10})
	if le.dirty {
		t.Errorf("setting a rectangle to what it already was marked the level unsaved")
	}
	le.setActive(&rl.Rectangle{Width: 11, Height: 10})
	if !le.dirty {
		t.Errorf("reshaping a rectangle did not mark the level unsaved")
	}
}

// A palette default reaches the tiles that follow it as well as every tile
// placed after it; a tile's own edit applies to that tile only.
func TestDefaultsAndTilesAreEditedSeparately(t *testing.T) {
	orig := rl.Rectangle{Width: 10, Height: 10}
	le := newLevelEditor(39, 83, []LevelObject{{Resource: "x", CollisionRects: []rl.Rectangle{orig}}}, nil)
	a, b, c := Vector2i{X: 39}, Vector2i{X: 78}, Vector2i{X: 117}
	le.place(a)
	le.place(b)

	changed := rl.Rectangle{Width: 20, Height: 20}
	le.setActive(&changed) // the palette entry is selected
	le.place(c)
	for _, p := range []Vector2i{a, b, c} {
		if got := le.objects[0][p].CollisionRects[0]; got != changed {
			t.Errorf("the tile at %v has %+v after its texture changed, want %+v", p, got, changed)
		}
	}
	if d, ok := le.defaults["x"]; !ok || d.Collision[0] != changed {
		t.Errorf("the changed default is not in the defaults to be saved: %+v", le.defaults)
	}

	le.selectAt(a)
	one := rl.Rectangle{Width: 5, Height: 5}
	le.setActive(&one)
	if got := le.objects[0][b].CollisionRects[0]; got != changed {
		t.Errorf("editing one tile changed another of the same texture: %+v", got)
	}
	if got := le.levelObjects[0].CollisionRects[0]; got != changed {
		t.Errorf("editing a tile changed the palette entry: %+v", got)
	}
}

// The go:embed pattern is *.png: only names it would carry are imported.
func TestDroppedTextureName(t *testing.T) {
	dir := filepath.Join("C:", "art")
	cases := []struct {
		in   string
		want string
		ok   bool
	}{
		{filepath.Join(dir, "wall.png"), blocksDir + "/wall.png", true},
		{filepath.Join(dir, "Wall.PNG"), blocksDir + "/Wall.png", true},
		{filepath.Join(dir, "notes.txt"), "", false},
		{filepath.Join(dir, "noext"), "", false},
		{filepath.Join(dir, "_hidden.png"), "", false},
		{filepath.Join(dir, ".png"), "", false},
	}
	for _, c := range cases {
		got, ok := droppedTextureName(c.in)
		if got != c.want || ok != c.ok {
			t.Errorf("droppedTextureName(%q) = %q, %t; want %q, %t", c.in, got, ok, c.want, c.ok)
		}
	}
}

// A click on the editor's own panel or picker belongs to the GUI, never the level.
func TestOverGui(t *testing.T) {
	le := newLevelEditor(39, 83, []LevelObject{{Resource: "x"}}, nil)
	p := editorPanel()
	if !le.overGui(rl.Vector2{X: p.X + 5, Y: p.Y + 5}) {
		t.Errorf("a point on the panel is not over the GUI")
	}
	mid := rl.Vector2{X: VW / 2, Y: VH / 2}
	if le.overGui(mid) {
		t.Errorf("the middle of the screen is over the GUI")
	}
	cell := pickerCellRect(0)
	onPicker := rl.Vector2{X: cell.X + 5, Y: cell.Y + 5}
	if le.overGui(onPicker) {
		t.Errorf("the closed picker takes clicks")
	}
	le.OpenObjects()
	if !le.overGui(onPicker) {
		t.Errorf("the open picker does not take clicks")
	}
	if rl.CheckCollisionRecs(editorPanel(), pickerCellRect(0)) {
		t.Errorf("the panel overlaps the picker")
	}
}

// Clicking from one numeric field to another moves typing there, whichever
// order raygui reports the two clicks in.
func TestNextEditField(t *testing.T) {
	cases := []struct {
		editing int
		toggled []bool
		want    int
	}{
		{-1, make([]bool, 4), -1},
		{-1, []bool{false, true, false, false}, 1},
		{1, []bool{false, true, false, false}, -1},
		{2, []bool{true, false, true, false}, 0},
		{0, []bool{true, false, true, false}, 2},
		{3, make([]bool, 4), 3},
	}
	for _, c := range cases {
		if got := nextEditField(c.editing, c.toggled); got != c.want {
			t.Errorf("nextEditField(%d, %v) = %d, want %d", c.editing, c.toggled, got, c.want)
		}
	}
}

// With no collision rectangle, a tile is not stood on.
func TestNoCollisionRectIsNotStoodOn(t *testing.T) {
	saved := LE
	defer func() { LE = saved }()
	LE = newLevelEditor(39, 83, nil, nil)
	LE.objects[0] = LevelLayer{{}: {Resource: "x"}}
	LE.buildIndex()
	a := &Actor{}
	a.checkCollision() // used to dereference the nil rectangle
	if a.OnGround {
		t.Errorf("a fighter stood on a tile with no collision rectangle")
	}
}
