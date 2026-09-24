package resistor

import (
	"encoding/json"
	"os"
	"slices"
	"strings"
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var twoAreas = []LitArea{
	{X: 320, Y: 96, Width: 128, Height: 64, Feather: 16},
	{X: -40, Y: -8, Width: 32, Height: 200, Feather: 0},
}

// An area is five numbers in the file, under the short keys the tiles and the
// lights use, and it comes back exactly as it went in.
func TestLitAreaJSON(t *testing.T) {
	a := LitArea{X: 320, Y: -96, Width: 128, Height: 64, Feather: 16}
	b, err := json.Marshal(a)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(b), `{"x":320,"y":-96,"w":128,"h":64,"feather":16}`; got != want {
		t.Errorf("marshalled %s, want %s", got, want)
	}
	var back LitArea
	if err := json.Unmarshal(b, &back); err != nil {
		t.Fatal(err)
	}
	if back != a {
		t.Errorf("unmarshalled %+v, want %+v", back, a)
	}
}

// levelWithAreas is a level with tiles, lights and the given lit areas.
func levelWithAreas(areas []LitArea) *LevelEditor {
	le := levelWithLights(threeLights)
	le.areas = slices.Clone(areas) // tests move areas; twoAreas must stay as it is
	return le
}

// A level's lit areas are saved with it and come back as they were.
func TestLevelFileAreasRoundTrip(t *testing.T) {
	le := levelWithAreas(twoAreas)
	b, err := le.marshalFile()
	if err != nil {
		t.Fatal(err)
	}
	objects, lights, areas, _, err := parseLevel(b)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(areas, le.areas) {
		t.Errorf("areas = %+v, want %+v", areas, le.areas)
	}
	if !slices.Equal(lights, le.lights) || len(objects[0]) != 2 {
		t.Errorf("the areas cost the level %d lights and %d tiles", len(lights), len(objects[0]))
	}
}

// A level with no lit areas is exactly the file it was before they existed.
func TestLevelFileWithoutAreas(t *testing.T) {
	le := levelWithAreas(nil)
	b, _ := le.marshalFile()
	if strings.Contains(string(b), areasKey) {
		t.Errorf("a level without lit areas was saved with an areas key: %s", b)
	}
	_, _, areas, _, err := parseLevel(b)
	if err != nil || len(areas) != 0 {
		t.Errorf("loaded %d areas, err %v", len(areas), err)
	}
	// a file written before lit areas existed loads with none
	_, _, areas, _, err = parseLevel([]byte(`{"0":{"39,0":{"resource":"a","x":39}}}`))
	if err != nil || len(areas) != 0 {
		t.Errorf("a file with no areas key loaded %d areas, err %v", len(areas), err)
	}
}

// An unknown key beside the areas is still skipped rather than failing the level.
func TestLevelFileAreasBesideAnUnknownKey(t *testing.T) {
	objects, _, areas, _, err := parseLevel([]byte(
		`{"0":{"39,0":{"resource":"a","x":39}},"litAreas":[{"x":8,"y":16,"w":32,"h":64,"feather":4}],"notes":"hi"}`))
	if err != nil {
		t.Fatalf("an unknown key failed the level: %v", err)
	}
	want := []LitArea{{X: 8, Y: 16, Width: 32, Height: 64, Feather: 4}}
	if !slices.Equal(areas, want) || len(objects[0]) != 1 {
		t.Errorf("loaded %+v and %v", areas, objects)
	}
}

// The lit areas stay out of the fingerprint, as the lights and the ambient do:
// how bright a corner of a level is must never decide who can play together.
func TestAreasDoNotChangeTheFingerprint(t *testing.T) {
	plain := levelWithAreas(nil)
	lit := levelWithAreas(twoAreas)
	other := levelWithAreas([]LitArea{{X: -300, Y: 40, Width: 8, Height: 8, Feather: 2}})
	if plain.Hash() != lit.Hash() || plain.Hash() != other.Hash() {
		t.Errorf("the lit areas changed the fingerprint: %d, %d, %d", plain.Hash(), lit.Hash(), other.Hash())
	}
	// and a saved level's tiles are the ones the fingerprint is taken over
	if got := savedTilesHash(t, lit); got != plain.Hash() {
		t.Errorf("the areas changed the tiles in the file: %d, want %d", got, plain.Hash())
	}
}

// savedTilesHash is the fingerprint of the tiles a level saves, read back.
func savedTilesHash(t *testing.T, le *LevelEditor) uint64 {
	t.Helper()
	objects, _, _, _, err := parseLevel(mustMarshalFile(t, le))
	if err != nil {
		t.Fatal(err)
	}
	return (&LevelEditor{objects: objects}).Hash()
}

func mustMarshalFile(t *testing.T, le *LevelEditor) []byte {
	t.Helper()
	b, err := le.marshalFile()
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// The load path the dedicated server uses keeps the lit areas along with the
// lights, the ambient and the tiles.
func TestLoadLevelKeepsTheAreas(t *testing.T) {
	if !DevTools {
		t.Skip("a release build loads the level it was built with, not one written here")
	}
	le := levelWithAreas(twoAreas)
	le.ambient = Ambient{Color: [3]uint8{80, 90, 200}, Strength: 0.35}
	path := t.TempDir() + "/level.json"
	if err := os.WriteFile(path, mustMarshalFile(t, le), 0o644); err != nil {
		t.Fatal(err)
	}
	back := NewLevelEditor(39, 83)
	back.LoadLevel(path)
	if !slices.Equal(back.areas, le.areas) {
		t.Errorf("loaded areas %+v, want %+v", back.areas, le.areas)
	}
	if back.areaSel != -1 {
		t.Errorf("a freshly loaded level has area %d selected", back.areaSel)
	}
	if back.ambient != le.ambient || len(back.lights) != len(threeLights) || len(back.objects[0]) != 2 {
		t.Errorf("loaded %+v, %d lights and %d tiles with them", back.ambient, len(back.lights), len(back.objects[0]))
	}
}

// A server and a client loading a level that differs only in its lit areas
// agree on the fingerprint: an area is never a level mismatch.
func TestAreasAreNotALevelMismatch(t *testing.T) {
	if !DevTools {
		t.Skip("a release build loads the level it was built with, not one written here")
	}
	server, client := NewLevelEditor(39, 83), NewLevelEditor(39, 83)
	dir := t.TempDir()
	withAreas, without := dir+"/with.json", dir+"/without.json"
	if err := os.WriteFile(withAreas, mustMarshalFile(t, levelWithAreas(twoAreas)), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(without, mustMarshalFile(t, levelWithAreas(nil)), 0o644); err != nil {
		t.Fatal(err)
	}
	client.LoadLevel(withAreas)
	server.LoadLevel(without)
	if server.Hash() != client.Hash() {
		t.Errorf("a client with lit areas and a server without them disagree on the fingerprint")
	}
}

// The feather is never wider than half the area's shorter side, and never
// negative, whatever the file or the panel puts in it.
func TestAreaFeather(t *testing.T) {
	for _, c := range []struct {
		name string
		a    LitArea
		want int32
	}{
		{"as authored", LitArea{Width: 128, Height: 64, Feather: 16}, 16},
		{"a hard edge", LitArea{Width: 128, Height: 64}, 0},
		{"wider than the area", LitArea{Width: 16, Height: 16, Feather: 64}, 8},
		{"half of the shorter side", LitArea{Width: 200, Height: 16, Feather: 64}, 8},
		{"negative", LitArea{Width: 128, Height: 64, Feather: -8}, 0},
		{"nothing to feather", LitArea{}, 0},
	} {
		if got := c.a.feather(); got != c.want {
			t.Errorf("%s: %+v.feather() = %d, want %d", c.name, c.a, got, c.want)
		}
	}
}

// The pieces of an area tile it exactly: every piece inside the area, none
// overlapping another, and their areas adding up to the whole - so no pixel is
// left dark and none is lit twice.
func TestAreaPiecesTileTheArea(t *testing.T) {
	for _, a := range []LitArea{
		{X: 320, Y: 96, Width: 128, Height: 64, Feather: 16},
		{X: -40, Y: -8, Width: 32, Height: 200, Feather: 0},
		{X: 0, Y: 0, Width: 16, Height: 16, Feather: 64}, // feathered to its own half
		{X: 7, Y: -3, Width: 9, Height: 5, Feather: 2},   // odd sizes
	} {
		pieces := areaPieces(a)
		r := a.rect()
		var total float32
		for i, p := range pieces {
			if p.rect.Width <= 0 || p.rect.Height <= 0 {
				t.Errorf("%+v: piece %d has nothing to it: %+v", a, i, p.rect)
			}
			if p.rect.X < r.X || p.rect.Y < r.Y ||
				p.rect.X+p.rect.Width > r.X+r.Width || p.rect.Y+p.rect.Height > r.Y+r.Height {
				t.Errorf("%+v: piece %d %+v is outside the area", a, i, p.rect)
			}
			total += p.rect.Width * p.rect.Height
			for j := i + 1; j < len(pieces); j++ {
				if aabb(p.rect, pieces[j].rect) {
					t.Errorf("%+v: pieces %d %+v and %d %+v overlap", a, i, p.rect, j, pieces[j].rect)
				}
			}
		}
		if want := r.Width * r.Height; total != want {
			t.Errorf("%+v: the pieces cover %g of %g", a, total, want)
		}
	}
}

// An area with no feather is its core alone, at full light: a hard edge.
func TestAreaPiecesHardEdge(t *testing.T) {
	a := LitArea{X: 10, Y: 20, Width: 30, Height: 40}
	pieces := areaPieces(a)
	if len(pieces) != 1 {
		t.Fatalf("a hard-edged area is %d pieces, want 1: %+v", len(pieces), pieces)
	}
	p := pieces[0]
	if p.rect != a.rect() {
		t.Errorf("the piece is %+v, want the area itself %+v", p.rect, a.rect())
	}
	if p.topLeft != 1 || p.bottomLeft != 1 || p.bottomRight != 1 || p.topRight != 1 {
		t.Errorf("a hard-edged area is not at full light: %+v", p)
	}
}

// An area with nothing to it is drawn as nothing at all.
func TestAreaPiecesOfNothing(t *testing.T) {
	for _, a := range []LitArea{{}, {Width: 10}, {Height: 10}, {Width: -4, Height: -4, Feather: 2}} {
		if got := areaPieces(a); got != nil {
			t.Errorf("%+v is drawn as %+v, want nothing", a, got)
		}
	}
}

// The light at a point of an area, as a piece really lands on the screen. The
// core and the sides are axis-aligned ramps, and a corner is a filtered texture
// stretched over it, so bilinear interpolation of the four corner values is
// what all of them come to.
func lightAtPoint(pieces []gradRect, x, y float32) (float32, bool) {
	for _, p := range pieces {
		if x < p.rect.X || x > p.rect.X+p.rect.Width || y < p.rect.Y || y > p.rect.Y+p.rect.Height {
			continue
		}
		u, v := float32(0), float32(0)
		if p.rect.Width > 0 {
			u = (x - p.rect.X) / p.rect.Width
		}
		if p.rect.Height > 0 {
			v = (y - p.rect.Y) / p.rect.Height
		}
		top := p.topLeft + (p.topRight-p.topLeft)*u
		bottom := p.bottomLeft + (p.bottomRight-p.bottomLeft)*u
		return top + (bottom-top)*v, true
	}
	return 0, false
}

// The ramps meet: the light is continuous across every join between a side and
// the core, a side and a corner, and the two halves of a corner, and it is full
// inside the core and out at the bounds.
func TestAreaPiecesRampsMeet(t *testing.T) {
	a := LitArea{X: 0, Y: 0, Width: 128, Height: 64, Feather: 16}
	pieces := areaPieces(a)
	at := func(x, y float32) float32 {
		got, ok := lightAtPoint(pieces, x, y)
		if !ok {
			t.Fatalf("(%g,%g) is in none of the pieces", x, y)
		}
		return got
	}
	for _, c := range []struct {
		name string
		x, y float32
		want float32
	}{
		{"the middle", 64, 32, 1},
		{"the core's left edge", 16, 32, 1},
		{"the core's top left corner", 16, 16, 1},
		{"half way along the left side", 8, 32, 0.5},
		{"half way along the top", 64, 8, 0.5},
		{"the left bound", 0, 32, 0},
		{"the top bound", 64, 0, 0},
		{"the outer corner", 0, 0, 0},
		{"half way through a corner, along its diagonal", 8, 8, 0.25},
		{"in a corner, nearer one bound", 8, 4, 0.125},
	} {
		if got := at(c.x, c.y); abs(got-c.want) > 1e-5 {
			t.Errorf("%s: light at (%g,%g) = %g, want %g", c.name, c.x, c.y, got, c.want)
		}
	}
	// the join a side shares with the corner beside it carries the same light
	// from both sides of the line
	for _, x := range []float32{0, 4, 8, 12, 16} {
		corner, side := at(x, 16-0.0001), at(x, 16+0.0001)
		if abs(corner-side) > 1e-3 {
			t.Errorf("at x=%g the corner gives %g and the side below it %g", x, corner, side)
		}
	}
}

// An area is drawn when any of it is in view, and skipped when none of it is.
func TestAreaVisible(t *testing.T) {
	view := rl.Rectangle{X: 0, Y: 0, Width: 160, Height: 90}
	for _, c := range []struct {
		name string
		a    LitArea
		want bool
	}{
		{"inside", LitArea{X: 10, Y: 10, Width: 20, Height: 20}, true},
		{"across the edge", LitArea{X: 150, Y: 10, Width: 40, Height: 20}, true},
		{"around the view", LitArea{X: -100, Y: -100, Width: 400, Height: 400}, true},
		{"off to the right", LitArea{X: 200, Y: 10, Width: 20, Height: 20}, false},
		{"above", LitArea{X: 10, Y: -60, Width: 20, Height: 20}, false},
	} {
		if got := areaVisible(c.a, view); got != c.want {
			t.Errorf("%s: areaVisible(%+v) = %v, want %v", c.name, c.a, got, c.want)
		}
	}
}

// Switching tools lets go of whatever was being dragged, the area tool's drag
// included, and changes no area. Which tool Q reaches is TestToolCycle's.
func TestSetToolLetsGoOfAnAreaDrag(t *testing.T) {
	le := levelWithAreas(twoAreas)
	le.areaDrag = rectDrag{active: true, h: handleBody}
	le.lightDrag = lightDrag{mode: lightDragMove}
	le.editField = 2
	le.setTool(toolArea)
	if le.areaDrag.active || le.lightDrag.mode != lightDragNone || le.editField != -1 {
		t.Errorf("switching tools kept a drag or the keyboard: %+v %+v %d", le.areaDrag, le.lightDrag, le.editField)
	}
	if !slices.Equal(le.areas, twoAreas) || le.dirty {
		t.Errorf("switching tools changed the areas")
	}
}

// A click on empty space places an area centred on it, taking the size and the
// feather of the area selected before it.
func TestPlaceArea(t *testing.T) {
	le := levelWithAreas(nil)
	le.placeArea(rl.Vector2{X: 100, Y: 50})
	if len(le.areas) != 1 || le.areaSel != 0 {
		t.Fatalf("placed %d areas, selected %d", len(le.areas), le.areaSel)
	}
	want := LitArea{X: 100 - defaultArea.Width/2, Y: 50 - defaultArea.Height/2,
		Width: defaultArea.Width, Height: defaultArea.Height, Feather: defaultArea.Feather}
	if le.areas[0] != want {
		t.Errorf("the first area is %+v, want %+v", le.areas[0], want)
	}
	if !le.dirty {
		t.Errorf("placing an area left the level saved")
	}
	// the next one matches the selected one
	le.setArea(0, LitArea{X: 0, Y: 0, Width: 30, Height: 10, Feather: 3})
	le.placeArea(rl.Vector2{X: -20, Y: -4})
	got := le.areas[1]
	if got.Width != 30 || got.Height != 10 || got.Feather != 3 {
		t.Errorf("the second area is %+v, want the first one's size and feather", got)
	}
	if got.X != -20-15 || got.Y != -4-5 {
		t.Errorf("the second area is at %d,%d, want it centred on the click", got.X, got.Y)
	}
}

// An area is moved by its body and resized by an edge or a corner, in whole
// pixels, and never inside out.
func TestDragArea(t *testing.T) {
	le := levelWithAreas([]LitArea{{X: 100, Y: 100, Width: 40, Height: 20, Feather: 4}})
	tol := float32(2)
	if !le.beginAreaDrag(rl.Vector2{X: 120, Y: 110}, tol) { // the body
		t.Fatal("a click inside the area took hold of nothing")
	}
	if le.areaSel != 0 || le.areaDrag.h != handleBody {
		t.Fatalf("took hold of area %d by %v", le.areaSel, le.areaDrag.h)
	}
	le.dragArea(rl.Vector2{X: 130.4, Y: 119.6})
	if got, want := le.areas[0], (LitArea{X: 110, Y: 110, Width: 40, Height: 20, Feather: 4}); got != want {
		t.Errorf("after moving: %+v, want %+v", got, want)
	}
	// a corner resizes
	if !le.beginAreaDrag(rl.Vector2{X: 150, Y: 130}, tol) {
		t.Fatal("a click on the bottom right corner took hold of nothing")
	}
	if le.areaDrag.h != handleBottomRight {
		t.Fatalf("took hold by %v, want the bottom right corner", le.areaDrag.h)
	}
	le.dragArea(rl.Vector2{X: 200, Y: 160})
	if got, want := le.areas[0], (LitArea{X: 110, Y: 110, Width: 90, Height: 50, Feather: 4}); got != want {
		t.Errorf("after resizing: %+v, want %+v", got, want)
	}
	// dragging the right edge past the left one stops at a pixel wide
	if !le.beginAreaDrag(rl.Vector2{X: 200, Y: 130}, tol) {
		t.Fatal("a click on the right edge took hold of nothing")
	}
	le.dragArea(rl.Vector2{X: 20, Y: 130})
	if got := le.areas[0]; got.Width != 1 || got.X != 110 {
		t.Errorf("dragging the right edge past the left one gave %+v", got)
	}
}

// A drag with nothing selected, or with the selection gone, takes nothing with
// it and lets go.
func TestDragAreaWithNothingHeld(t *testing.T) {
	le := levelWithAreas(twoAreas)
	le.areaSel, le.areaDrag = -1, rectDrag{active: true, h: handleBody}
	le.dragArea(rl.Vector2{X: 10, Y: 10})
	if le.areaDrag.active || !slices.Equal(le.areas, twoAreas) || le.dirty {
		t.Errorf("a drag with nothing selected moved something: %+v", le.areas)
	}
}

// Setting an area is an edit; setting the area it already is, is not.
func TestSetArea(t *testing.T) {
	le := levelWithAreas(twoAreas)
	le.setArea(0, le.areas[0])
	if le.dirty {
		t.Errorf("setting the area it already is marked the level unsaved")
	}
	want := LitArea{X: 1, Y: 2, Width: 3, Height: 4, Feather: 1}
	le.setArea(0, want)
	if le.areas[0] != want || !le.dirty {
		t.Errorf("area 0 is %+v, dirty %v", le.areas[0], le.dirty)
	}
	le.dirty = false
	le.setArea(99, want) // out of range
	le.setArea(-1, want)
	if le.dirty {
		t.Errorf("setting an area that is not there marked the level unsaved")
	}
}

// Removing an area keeps the selection on the area it was on.
func TestDeleteArea(t *testing.T) {
	le := levelWithAreas([]LitArea{{X: 1}, {X: 2}, {X: 3}})
	le.areaSel = 2
	le.deleteArea(0)
	if len(le.areas) != 2 || le.areas[0].X != 2 || le.areaSel != 1 {
		t.Errorf("after removing the first: %+v, selection %d", le.areas, le.areaSel)
	}
	if !le.dirty {
		t.Errorf("removing an area left the level saved")
	}
	le.deleteArea(1) // the selected one
	if le.areaSel != -1 {
		t.Errorf("removing the selected area left %d selected", le.areaSel)
	}
	le.dirty = false
	le.deleteArea(5)
	if len(le.areas) != 1 || le.dirty {
		t.Errorf("removing an area that is not there changed the level")
	}
}

// Letting go of an area is not an edit: nothing is removed and the level is as
// saved as it was.
func TestDeselectArea(t *testing.T) {
	le := levelWithAreas(twoAreas)
	le.areaSel = 1
	le.areaDrag = rectDrag{active: true, h: handleBody}
	le.editField = 3
	le.deselectArea()
	if le.areaSel != -1 || le.areaDrag.active || le.editField != -1 {
		t.Errorf("after letting go: selection %d, drag %+v, field %d", le.areaSel, le.areaDrag, le.editField)
	}
	if !slices.Equal(le.areas, twoAreas) || le.dirty {
		t.Errorf("letting go of an area changed the level: %+v, dirty %v", le.areas, le.dirty)
	}
}

// A click takes hold of the area drawn on top: the later one of two that
// overlap, and nothing where there is no area.
func TestAreaAt(t *testing.T) {
	le := levelWithAreas([]LitArea{{X: 0, Y: 0, Width: 100, Height: 100}, {X: 50, Y: 50, Width: 100, Height: 100}})
	if i, h := le.areaAt(rl.Vector2{X: 75, Y: 75}, 2); i != 1 || h != handleBody {
		t.Errorf("a click where both areas lie took area %d by %v", i, h)
	}
	if i, _ := le.areaAt(rl.Vector2{X: 10, Y: 10}, 2); i != 0 {
		t.Errorf("a click on the first area alone took area %d", i)
	}
	if i, h := le.areaAt(rl.Vector2{X: 500, Y: 500}, 2); i != -1 || h != handleNone {
		t.Errorf("a click on empty space took area %d by %v", i, h)
	}
}
