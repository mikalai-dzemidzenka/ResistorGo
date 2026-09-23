package resistor

import (
	"slices"
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// undoEditor is an editor with one palette entry that has one collision
// rectangle, its history started, as the first frame of edit mode starts it.
func undoEditor() *LevelEditor {
	le := newLevelEditor(39, 83, []LevelObject{{Resource: "x", CollisionRects: []rl.Rectangle{rect(0, 0, 39, 83)}}}, nil)
	le.commitEdit(false)
	return &le
}

// frame is the start of the next frame: the last one's edit, if it finished, is
// a step.
func frame(le *LevelEditor) { le.commitEdit(false) }

func TestUndoAndRedoPlacing(t *testing.T) {
	le := undoEditor()
	a, b := Vector2i{X: 39}, Vector2i{X: 78}
	le.place(a)
	frame(le)
	le.place(b)
	frame(le)

	count := func() int { return len(le.objects[0]) }
	if !le.undo() || count() != 1 {
		t.Fatalf("one undo left %d objects, want 1", count())
	}
	if _, ok := le.objects[0][a]; !ok {
		t.Errorf("the first placement was undone instead of the second")
	}
	if len(le.solids) != 1 {
		t.Errorf("after the undo the level collides with %d rectangles, want 1", len(le.solids))
	}
	if !le.undo() || count() != 0 {
		t.Fatalf("a second undo left %d objects, want 0", count())
	}
	if le.undo() {
		t.Errorf("undo went back past the start")
	}
	if !le.redo() || !le.redo() || count() != 2 {
		t.Errorf("two redos gave %d objects, want 2", count())
	}
	if le.redo() {
		t.Errorf("redo went past the last edit")
	}

	// a new edit after an undo forgets what could have been redone
	le.undo()
	le.erase(a)
	frame(le)
	if le.redo() {
		t.Errorf("redo is still offered after a new edit")
	}
}

// A drag runs over many frames and is one step.
func TestADragIsOneStep(t *testing.T) {
	le := undoEditor()
	for i := range 10 {
		le.setActive(&rl.Rectangle{Width: float32(40 + i), Height: 83})
		le.commitEdit(true) // the mouse is still held
	}
	frame(le) // let go
	if n := len(le.hist.undo); n != 1 {
		t.Fatalf("a ten-frame drag took %d steps, want 1", n)
	}
	le.undo()
	if got := le.levelObjects[0].CollisionRects; !slices.Equal(got, []rl.Rectangle{rect(0, 0, 39, 83)}) {
		t.Errorf("undoing the drag left %+v, want the rectangle as it was", got)
	}
}

// Undoing a palette edit takes back what it passed on: the defaults, the
// cursor's copy, and every tile that follows the texture.
func TestUndoAPaletteEdit(t *testing.T) {
	le := undoEditor()
	a := Vector2i{X: 39}
	le.place(a)
	frame(le)
	le.setActive(&rl.Rectangle{Width: 10, Height: 10})
	frame(le)
	if got := le.objects[0][a].CollisionRects[0]; got != rect(0, 0, 10, 10) {
		t.Fatalf("the following tile did not take the edit: %+v", got)
	}

	le.undo()
	want := rect(0, 0, 39, 83)
	if got := le.objects[0][a].CollisionRects[0]; got != want {
		t.Errorf("the following tile kept %+v after the undo", got)
	}
	if got := le.obj.CollisionRects[0]; got != want {
		t.Errorf("the cursor's copy kept %+v after the undo", got)
	}
	if got, ok := le.defaults["x"]; ok {
		t.Errorf("the defaults kept %+v after the undo, want none as before the edit", got)
	}
}

func TestUndoLightsAreasAmbient(t *testing.T) {
	le := undoEditor()
	le.placeLight(rl.Vector2{X: 10, Y: 10})
	frame(le)
	l := le.lights[0]
	l.Radius = 5
	le.setLight(0, l)
	frame(le)
	le.placeArea(rl.Vector2{X: 50, Y: 50})
	frame(le)
	le.setAmbient(Ambient{Strength: 0.9})
	frame(le)

	le.undo()
	if le.ambient != defaultAmbient {
		t.Errorf("undoing the ambient left %+v", le.ambient)
	}
	le.undo()
	if len(le.areas) != 0 || le.areaSel != -1 {
		t.Errorf("undoing the area left %d areas, selected %d", len(le.areas), le.areaSel)
	}
	le.undo()
	if le.lights[0].Radius != defaultLight.Radius {
		t.Errorf("undoing the radius left %d", le.lights[0].Radius)
	}
	le.undo()
	if len(le.lights) != 0 || le.lightSel != -1 {
		t.Errorf("undoing the light left %d lights, selected %d", len(le.lights), le.lightSel)
	}
}

// Undoing back to what was saved is saved; anything else is not.
func TestUndoKnowsWhatIsSaved(t *testing.T) {
	le := undoEditor()
	le.place(Vector2i{X: 39})
	frame(le)
	le.dirty = false // as SaveLevel leaves it
	le.saved()
	le.place(Vector2i{X: 78})
	frame(le)
	if !le.dirty {
		t.Fatal("an edit after saving is not unsaved")
	}
	le.undo()
	if le.dirty {
		t.Errorf("undoing back to the saved level still says unsaved")
	}
	le.undo()
	if !le.dirty {
		t.Errorf("undoing past the saved level says saved")
	}
	le.redo()
	if le.dirty {
		t.Errorf("redoing back to the saved level still says unsaved")
	}
}

// A selected tile an undo takes away is let go of, not left pointing at nothing.
func TestUndoLetsGoOfAVanishedSelection(t *testing.T) {
	le := undoEditor()
	a := Vector2i{X: 39}
	le.place(a)
	frame(le)
	le.selectAt(a)
	le.undo()
	if !le.sel.proto {
		t.Errorf("the selection still points at a tile that is gone: %+v", le.sel)
	}
}

// What the history keeps is its own: editing after an undo does not reach the
// step that redo would bring back.
func TestHistoryIsNotSharedWithTheLevel(t *testing.T) {
	le := undoEditor()
	a := Vector2i{X: 39}
	le.place(a)
	frame(le)
	le.undo()
	le.redo()
	le.objects[0][a] = LevelObject{Resource: "changed"}
	le.undo()
	le.redo()
	if got := le.objects[0][a].Resource; got != "x" {
		t.Errorf("an edit made outside the history leaked into it: %q", got)
	}
}

func TestUndoForgetsTheOldest(t *testing.T) {
	le := undoEditor()
	for i := range undoDepth + 5 {
		le.place(Vector2i{X: int32(i) * 39})
		frame(le)
	}
	if n := len(le.hist.undo); n != undoDepth {
		t.Errorf("the history holds %d steps, want %d", n, undoDepth)
	}
	for le.undo() {
	}
	if n := len(le.objects[0]); n != 5 {
		t.Errorf("undoing all the way back left %d objects, want the 5 older than the history", n)
	}
}
