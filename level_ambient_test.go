package resistor

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// The clear colour of the light map is the ambient's colour weighted by its
// strength, saturating rather than wrapping, and always opaque.
func TestAmbientClear(t *testing.T) {
	for _, c := range []struct {
		name string
		a    Ambient
		want rl.Color
	}{
		{"a quarter of a cool blue", Ambient{Color: [3]uint8{120, 140, 255}, Strength: 0.25}, rl.Color{R: 30, G: 35, B: 64, A: 255}},
		{"full strength is the colour itself", Ambient{Color: [3]uint8{10, 20, 30}, Strength: 1}, rl.Color{R: 10, G: 20, B: 30, A: 255}},
		{"no strength is black", Ambient{Color: [3]uint8{120, 140, 255}}, rl.Color{A: 255}},
		{"over full strength saturates", Ambient{Color: [3]uint8{120, 140, 255}, Strength: 4}, rl.Color{R: 255, G: 255, B: 255, A: 255}},
		{"a negative strength is black", Ambient{Color: [3]uint8{120, 140, 255}, Strength: -2}, rl.Color{A: 255}},
	} {
		if got := ambientClear(c.a); got != c.want {
			t.Errorf("%s: ambientClear(%+v) = %+v, want %+v", c.name, c.a, got, c.want)
		}
	}
}

// A level's ambient is saved with it and comes back as it was.
func TestLevelFileAmbientRoundTrip(t *testing.T) {
	le := levelWithLights(threeLights)
	le.ambient = Ambient{Color: [3]uint8{255, 120, 40}, Strength: 0.4}
	b, err := le.marshalFile()
	if err != nil {
		t.Fatal(err)
	}
	_, _, _, ambient, err := parseLevel(b)
	if err != nil {
		t.Fatal(err)
	}
	if ambient != le.ambient {
		t.Errorf("ambient = %+v, want %+v", ambient, le.ambient)
	}
}

// A level file with no ambient in it is lit by the default one, and every level
// saved carries its ambient - the one it was authored with, not whatever the
// default has become since.
func TestLevelFileWithoutAmbient(t *testing.T) {
	_, _, _, ambient, err := parseLevel([]byte(`{"0":{"39,0":{"resource":"a","x":39}}}`))
	if err != nil {
		t.Fatal(err)
	}
	if ambient != defaultAmbient {
		t.Errorf("a file with no ambient loaded %+v, want the default %+v", ambient, defaultAmbient)
	}
	le := levelWithLights(nil)
	b, _ := le.marshalFile()
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(b, &raw); err != nil {
		t.Fatal(err)
	}
	if _, ok := raw[ambientKey]; !ok {
		t.Errorf("a level lit by the default ambient was saved without an ambient key: %s", b)
	}
	if _, _, _, back, err := parseLevel(b); err != nil || back != le.ambient {
		t.Errorf("the saved ambient came back as %+v (%v), want %+v", back, err, le.ambient)
	}
}

// withoutAmbient is a saved level with its ambient key taken back out, for the
// tests that compare one against the bytes its tiles alone give. The key sorts
// wherever it sorts among the layer numbers, so either the comma before it or
// the one after it goes with it.
func withoutAmbient(b []byte, a Ambient) string {
	m, err := json.Marshal(map[string]Ambient{ambientKey: a})
	if err != nil {
		panic(err)
	}
	key := string(m[1 : len(m)-1])
	s := strings.Replace(string(b), ","+key, "", 1)
	return strings.Replace(s, key+",", "", 1)
}

// A level that carries an ambient gets it as written - a strength of zero is a
// level that is black outside its lights, not a level with no ambient at all.
func TestLevelFileAmbientZeroStrength(t *testing.T) {
	_, _, _, ambient, err := parseLevel([]byte(`{"0":null,"ambient":{"color":[10,20,30],"strength":0}}`))
	if err != nil {
		t.Fatal(err)
	}
	want := Ambient{Color: [3]uint8{10, 20, 30}}
	if ambient != want {
		t.Errorf("ambient = %+v, want %+v", ambient, want)
	}
	if ambientClear(ambient) != (rl.Color{A: 255}) {
		t.Errorf("a level at strength zero is not black outside its lights")
	}
}

// The ambient stays out of the fingerprint, as the lights do: how dark a level
// is must never decide who can play together.
func TestAmbientDoesNotChangeTheFingerprint(t *testing.T) {
	plain := levelWithLights(threeLights)
	lit := levelWithLights(threeLights)
	lit.ambient = Ambient{Color: [3]uint8{255, 120, 40}, Strength: 0.4}
	dark := levelWithLights(threeLights)
	dark.ambient = Ambient{Color: [3]uint8{0, 0, 0}}
	if plain.Hash() != lit.Hash() || plain.Hash() != dark.Hash() {
		t.Errorf("the ambient changed the fingerprint: %d, %d, %d", plain.Hash(), lit.Hash(), dark.Hash())
	}
}

// The load path the dedicated server uses keeps the ambient along with the
// lights and the tiles.
func TestLoadLevelKeepsTheAmbient(t *testing.T) {
	if !DevTools {
		t.Skip("a release build loads the level it was built with, not one written here")
	}
	le := levelWithLights(threeLights)
	le.ambient = Ambient{Color: [3]uint8{80, 90, 200}, Strength: 0.35}
	b, err := le.marshalFile()
	if err != nil {
		t.Fatal(err)
	}
	path := t.TempDir() + "/level.json"
	if err := os.WriteFile(path, b, 0o644); err != nil {
		t.Fatal(err)
	}
	back := NewLevelEditor(39, 83)
	back.LoadLevel(path)
	if back.ambient != le.ambient {
		t.Errorf("loaded ambient %+v, want %+v", back.ambient, le.ambient)
	}
	if len(back.lights) != len(threeLights) || len(back.objects[0]) != 2 {
		t.Errorf("loaded %d lights and %d tiles with it", len(back.lights), len(back.objects[0]))
	}
}

// A server and a client loading the same level agree on its ambient, as they do
// on its lights and its fingerprint.
func TestServerAndClientLoadTheSameAmbient(t *testing.T) {
	server, client := NewLevelEditor(39, 83), NewLevelEditor(39, 83)
	server.LoadLevel("level_1.json")
	client.LoadLevel("level_1.json")
	if server.ambient != client.ambient {
		t.Errorf("server loaded ambient %+v, client %+v", server.ambient, client.ambient)
	}
	if server.Hash() != client.Hash() {
		t.Errorf("server and client disagree on the fingerprint")
	}
}

// Setting the ambient is an edit like any other: it takes effect and leaves the
// level unsaved, and setting the ambient it already has changes nothing.
func TestSetAmbient(t *testing.T) {
	le := levelWithLights(threeLights)
	if le.dirty {
		t.Fatal("a freshly built level is already unsaved")
	}
	le.setAmbient(le.ambient)
	if le.dirty {
		t.Errorf("setting the ambient it already has marked the level unsaved")
	}
	want := Ambient{Color: [3]uint8{255, 120, 40}, Strength: 0.4}
	le.setAmbient(want)
	if le.ambient != want {
		t.Errorf("ambient = %+v, want %+v", le.ambient, want)
	}
	if !le.dirty {
		t.Errorf("changing the ambient left the level saved")
	}
}

// Letting go of a light is what brings the ambient fields back, and it is not
// an edit: a right click on empty space deselects rather than placing.
func TestDeselectLight(t *testing.T) {
	le := levelWithLights(threeLights)
	le.lightSel = 2
	le.lightDrag = lightDrag{mode: lightDragMove}
	le.editField = 1
	le.deselectLight()
	if le.lightSel != -1 || le.lightDrag.mode != lightDragNone || le.editField != -1 {
		t.Errorf("after letting go: selection %d, drag %d, field %d", le.lightSel, le.lightDrag.mode, le.editField)
	}
	if len(le.lights) != len(threeLights) {
		t.Errorf("letting go of a light removed one: %d left", len(le.lights))
	}
	if le.dirty {
		t.Errorf("letting go of a light marked the level unsaved")
	}
}

// The darkness slider reads the ambient strength upside down, clamped to what a
// slider can show, and sets the strength that leaves the level as dark as it
// was dragged to.
func TestAmbientDarkness(t *testing.T) {
	for _, c := range []struct {
		s    float32
		want int32
	}{{0.4, 60}, {0, 100}, {1, 0}, {4, 0}, {-2, 100}, {0.255, 75}} {
		if got := ambientDarkness(c.s); got != c.want {
			t.Errorf("ambientDarkness(%v) = %d, want %d", c.s, got, c.want)
		}
	}
	for d := int32(0); d <= 100; d++ {
		if got := ambientDarkness(darknessStrength(d)); got != d {
			t.Errorf("the slider at %d%% reads back as %d%%", d, got)
		}
	}
	if darknessStrength(-5) != 1 || darknessStrength(150) != 0 {
		t.Errorf("a darkness out of range was not clamped: %v, %v", darknessStrength(-5), darknessStrength(150))
	}
}
