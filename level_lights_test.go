package resistor

import (
	"encoding/json"
	"os"
	"slices"
	"strings"
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var threeLights = []LevelLight{
	{X: 40, Y: 30, Radius: 256, Color: [3]uint8{255, 255, 255}, Intensity: 1},
	{X: 120, Y: 120, Radius: 128, Color: [3]uint8{255, 255, 255}, Intensity: 1},
	{X: -40, Y: 30, Radius: 128, Color: [3]uint8{255, 255, 255}, Intensity: 1},
}

func levelWithLights(lights []LevelLight) *LevelEditor {
	le := newLevelEditor(39, 83, nil, nil)
	le.objects[0] = LevelLayer{
		{X: 39}:         {Resource: "a", X: 39, CollisionRects: []rl.Rectangle{{Width: 5, Height: 5}}},
		{X: -78, Y: 83}: {Resource: "b", X: -78, Y: 83},
	}
	le.objects[-1] = nil
	le.lights = slices.Clone(lights) // tests move lights; threeLights must stay as it is
	return &le
}

func TestLevelFileRoundTrip(t *testing.T) {
	le := levelWithLights([]LevelLight{{X: -5, Y: 7, Radius: 64, Color: [3]uint8{255, 10, 0}, Intensity: 0.5}})
	b, err := le.marshalFile()
	if err != nil {
		t.Fatal(err)
	}
	objects, lights, _, _, err := parseLevel(b)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(lights, le.lights) {
		t.Errorf("lights = %+v, want %+v", lights, le.lights)
	}
	back := newLevelEditor(39, 83, nil, nil)
	back.objects = objects
	if back.Hash() != le.Hash() {
		t.Errorf("the tiles did not survive the round trip")
	}
	again, _ := back.marshal()
	orig, _ := le.marshal()
	if string(again) != string(orig) {
		t.Errorf("tiles re-saved as %s, want %s", again, orig)
	}
}

func TestLevelFileWithoutLights(t *testing.T) {
	le := levelWithLights(nil)
	b, _ := le.marshalFile()
	if strings.Contains(string(b), lightsKey) {
		t.Errorf("a level without lights was saved with a lights key: %s", b)
	}
	if got := savedTilesHash(t, le); got != le.Hash() {
		t.Errorf("a level without lights saved other tiles: %d, want %d", got, le.Hash())
	}
	objects, lights, _, _, err := parseLevel(b)
	if err != nil || len(lights) != 0 || len(objects[0]) != 2 {
		t.Errorf("loaded %d layers, %d lights, err %v", len(objects), len(lights), err)
	}
}

func TestLevelFileSkipsUnknownKeys(t *testing.T) {
	objects, lights, _, _, err := parseLevel([]byte(`{"0":{"39,0":{"resource":"a","x":39}},"notes":"hi"}`))
	if err != nil {
		t.Fatalf("an unknown key failed the level: %v", err)
	}
	if len(objects[0]) != 1 || len(lights) != 0 {
		t.Errorf("loaded %v and %v", objects, lights)
	}
}

func TestLightsDoNotChangeTheFingerprint(t *testing.T) {
	with, without := levelWithLights(threeLights), levelWithLights(nil)
	if with.Hash() != without.Hash() {
		t.Errorf("adding lights changed the fingerprint")
	}
	with.lights[0].X += 10
	if with.Hash() != without.Hash() {
		t.Errorf("moving a light changed the fingerprint")
	}
}

// level1Bytes is level_1.json as this build loads it.
func level1Bytes(t *testing.T) []byte {
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
	return b
}

// checkLevelLights checks a level file's lights: they survive a save unchanged,
// and the fingerprint is the one its tiles alone give. It
// returns that fingerprint.
func checkLevelLights(t *testing.T, name string, b []byte) uint64 {
	t.Helper()
	objects, lights, _, _, err := parseLevel(b)
	if err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	le := newLevelEditor(39, 83, nil, nil)
	le.objects, le.lights = objects, lights
	saved, err := le.marshalFile()
	if err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	if _, back, _, _, err := parseLevel(saved); err != nil || !slices.Equal(back, lights) {
		t.Errorf("%s: lights after a save = %+v (%v), want %+v", name, back, err, lights)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(b, &raw); err != nil {
		t.Fatal(err)
	}
	delete(raw, lightsKey)
	delete(raw, ambientKey)
	delete(raw, areasKey) // the same file, with only its tiles
	stripped, _ := json.Marshal(raw)
	tiles, _, _, _, err := parseLevel(stripped)
	if err != nil {
		t.Fatal(err)
	}
	if le.Hash() != (&LevelEditor{objects: tiles}).Hash() {
		t.Errorf("%s: the lights changed the level's fingerprint", name)
	}
	return le.Hash()
}

// The level has lights, they survive a save, and they stay out of its
// fingerprint - whatever lights the level is given, and wherever they are.
func TestLevel1Lights(t *testing.T) {
	b := level1Bytes(t)
	le := NewLevelEditor(39, 83)
	le.LoadLevel("level_1.json")
	if len(le.lights) == 0 {
		t.Fatalf("level_1.json loaded with no lights")
	}
	hash := checkLevelLights(t, "level_1.json", b)

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(b, &raw); err != nil {
		t.Fatal(err)
	}
	lights := slices.Clone(le.lights)
	lights[0].X += 17
	lights = append(lights, LevelLight{X: -3, Y: 9, Radius: 40, Color: [3]uint8{255, 0, 0}, Intensity: 0.5})
	raw[lightsKey], _ = json.Marshal(lights)
	edited, _ := json.Marshal(raw)
	if checkLevelLights(t, "with a light moved and one added", edited) != hash {
		t.Errorf("moving a light and adding one changed the level's fingerprint")
	}
}

// The dedicated server loads the level through LoadLevel, and never draws it:
// it gets the lights all the same, and agrees with a client on them and on the
// fingerprint.
func TestServerAndClientLoadTheSameLevel(t *testing.T) {
	server, client := NewLevelEditor(39, 83), NewLevelEditor(39, 83)
	server.LoadLevel("level_1.json")
	client.LoadLevel("level_1.json")
	if len(server.lights) == 0 || len(server.objectsIndexed) == 0 {
		t.Errorf("the server loaded %d lights and %d tiles", len(server.lights), len(server.objectsIndexed))
	}
	if !slices.Equal(server.lights, client.lights) {
		t.Errorf("server and client loaded different lights")
	}
	if server.Hash() != client.Hash() {
		t.Errorf("server and client disagree on the fingerprint")
	}
}

func TestLightVisible(t *testing.T) {
	view := rl.Rectangle{X: 0, Y: 0, Width: 100, Height: 50}
	cases := []struct {
		name string
		l    LevelLight
		want bool
	}{
		{"inside", LevelLight{X: 50, Y: 25, Radius: 5}, true},
		{"covering the view", LevelLight{X: 50, Y: 25, Radius: 500}, true},
		{"outside", LevelLight{X: 300, Y: 25, Radius: 50}, false},
		{"over the left edge", LevelLight{X: -10, Y: 25, Radius: 20}, true},
		{"only a corner in", LevelLight{X: 105, Y: 55, Radius: 10}, true},
		{"just short of the edge", LevelLight{X: -21, Y: 25, Radius: 20}, false},
	}
	for _, c := range cases {
		if got := lightVisible(c.l, view); got != c.want {
			t.Errorf("%s: lightVisible = %t, want %t", c.name, got, c.want)
		}
	}
}

func TestViewRect(t *testing.T) {
	cam := rl.Camera2D{Offset: rl.Vector2{X: 640, Y: 360}, Target: rl.Vector2{X: 100, Y: 50}, Zoom: 1}
	near := func(a, b rl.Rectangle) bool {
		return abs(a.X-b.X) < 0.01 && abs(a.Y-b.Y) < 0.01 && abs(a.Width-b.Width) < 0.01 && abs(a.Height-b.Height) < 0.01
	}
	if got, want := viewRect(cam, 1280, 720), (rl.Rectangle{X: -540, Y: -310, Width: 1280, Height: 720}); !near(got, want) {
		t.Errorf("zoom 1: view = %+v, want %+v", got, want)
	}
	cam.Zoom = 3
	if got, want := viewRect(cam, 1280, 720), (rl.Rectangle{X: 100 - 640.0/3, Y: -70, Width: 1280.0 / 3, Height: 240}); !near(got, want) {
		t.Errorf("zoom 3: view = %+v, want %+v", got, want)
	}
	cam.Zoom = 4
	if got, want := viewRect(cam, 1280, 720), (rl.Rectangle{X: -60, Y: -40, Width: 320, Height: 180}); !near(got, want) {
		t.Errorf("zoom 4: view = %+v, want %+v", got, want)
	}
}

func TestToolCycle(t *testing.T) {
	got := []int{toolPlace}
	for range 4 {
		got = append(got, nextTool(got[len(got)-1]))
	}
	if want := []int{toolPlace, toolRect, toolLight, toolArea, toolPlace}; !slices.Equal(got, want) {
		t.Errorf("Q cycles %v, want %v", got, want)
	}
	le := newLevelEditor(39, 83, nil, nil)
	le.lightDrag = lightDrag{mode: lightDragMove}
	le.setTool(toolPlace)
	if le.lightDrag.mode != lightDragNone {
		t.Errorf("switching tool left a light drag going")
	}
}

func v2(x, y float32) rl.Vector2 { return rl.Vector2{X: x, Y: y} }

func TestLightAt(t *testing.T) {
	le := newLevelEditor(39, 83, nil, nil)
	// the second light's center sits on the first light's radius handle
	le.lights = []LevelLight{{X: 0, Y: 0, Radius: 50}, {X: 50, Y: 0, Radius: 20}}
	if i, mode := le.lightAt(v2(50, 0), 1); i != 0 || mode != lightDragRadius {
		t.Errorf("on a handle that is also a center: light %d mode %d, want the handle of light 0", i, mode)
	}
	if i, mode := le.lightAt(v2(1, 1), 1); i != 0 || mode != lightDragMove {
		t.Errorf("on a center: light %d mode %d, want the center of light 0", i, mode)
	}
	if i, _ := le.lightAt(v2(25, 30), 1); i != -1 {
		t.Errorf("on nothing: light %d, want none", i)
	}
}

func TestPlaceLight(t *testing.T) {
	le := newLevelEditor(39, 83, nil, nil)
	le.placeLight(v2(10.4, -3.6))
	first := le.lights[0]
	if first.X != 10 || first.Y != -4 || first.Radius != defaultLight.Radius ||
		first.Color != defaultLight.Color || first.Intensity != defaultLight.Intensity {
		t.Errorf("first light = %+v, want the defaults at 10,-4", first)
	}
	if le.lightSel != 0 || !le.dirty {
		t.Errorf("a placed light is not selected (%d) or not an unsaved edit (%t)", le.lightSel, le.dirty)
	}

	le.lights[0].Radius, le.lights[0].Color, le.lights[0].Intensity = 64, [3]uint8{0, 0, 255}, 0.5
	for i := range 3 {
		le.placeLight(v2(float32(100+i*50), 0))
	}
	for _, l := range le.lights[1:] {
		if l.Radius != 64 || l.Color != [3]uint8{0, 0, 255} || l.Intensity != 0.5 {
			t.Errorf("a light placed after a blue radius-64 one is %+v", l)
		}
	}
}

func TestDragLight(t *testing.T) {
	le := newLevelEditor(39, 83, nil, nil)
	le.lights = []LevelLight{{X: 0, Y: 0, Radius: 50, Intensity: 1}}

	if !le.beginLightDrag(v2(1, 0), 1) || le.lightDrag.mode != lightDragMove {
		t.Fatal("did not take hold of the center")
	}
	le.dragLight(v2(11, 5))
	if l := le.lights[0]; l.X != 10 || l.Y != 5 || l.Radius != 50 {
		t.Errorf("moved light = %+v, want 10,5 radius 50", l)
	}
	if !le.dirty {
		t.Errorf("moving a light is not an unsaved edit")
	}

	le.lightDrag = lightDrag{}
	if !le.beginLightDrag(v2(60, 5), 1) || le.lightDrag.mode != lightDragRadius {
		t.Fatal("did not take hold of the radius handle")
	}
	le.dragLight(v2(90, 40))
	if l := le.lights[0]; l.X != 10 || l.Y != 5 || l.Radius != 80 {
		t.Errorf("resized light = %+v, want center 10,5 radius 80", l)
	}
	le.dragLight(v2(-100, 5))
	if r := le.lights[0].Radius; r != 1 {
		t.Errorf("dragging the handle past the center gave radius %d, want 1", r)
	}
}

func TestDeleteLight(t *testing.T) {
	le := newLevelEditor(39, 83, nil, nil)
	le.lights = []LevelLight{{X: 1}, {X: 2}, {X: 3}}
	le.lightSel = 2
	le.deleteLight(0)
	if len(le.lights) != 2 || le.lightSel != 1 || le.lights[le.lightSel].X != 3 {
		t.Errorf("after deleting an earlier light the selection is %d of %+v, want the same light", le.lightSel, le.lights)
	}
	if !le.dirty {
		t.Errorf("deleting a light is not an unsaved edit")
	}
	le.deleteLight(le.lightSel)
	if le.lightSel != -1 || len(le.lights) != 1 {
		t.Errorf("deleting the selected light left selection %d, %d lights", le.lightSel, len(le.lights))
	}
	le.deleteLight(-1) // Delete with nothing selected
	if len(le.lights) != 1 {
		t.Errorf("Delete with nothing selected removed a light")
	}
}

// Saving writes the lights; a light edit is an edit like any other.
func TestSaveWritesLights(t *testing.T) {
	if !DevTools {
		t.Skip("a release build never writes the level")
	}
	le := levelWithLights(nil)
	le.placeLight(v2(5, 5))
	path := t.TempDir() + "/level.json"
	le.SaveLevel(path)
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	_, lights, _, _, err := parseLevel(b)
	if err != nil || len(lights) != 1 || lights[0].X != 5 {
		t.Errorf("saved lights = %+v, err %v", lights, err)
	}
	if le.dirty {
		t.Errorf("still unsaved after saving")
	}
}
