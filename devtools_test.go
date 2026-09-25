package resistor

import (
	"os"
	"testing"
)

// Where the level comes from, in both builds: a modified level_1.json in the
// working directory is played over the embedded one, so a saved edit takes
// effect. The dedicated server loads the same way.
func TestLoadLevelSource(t *testing.T) {
	embedded := func() *LevelEditor {
		le := NewLevelEditor(39, 83)
		b, err := assetFS.ReadFile("level_1.json")
		if err != nil {
			t.Fatal(err)
		}
		objects, _, _, _, err := parseLevel(b)
		if err != nil {
			t.Fatal(err)
		}
		le.objects = objects
		return &le
	}
	want := embedded().Hash()

	mod := embedded()
	if mod.objects[0] == nil {
		mod.objects[0] = make(LevelLayer)
	}
	mod.objects[0][Vector2i{X: 9999, Y: 9999}] = LevelObject{X: 9999, Y: 9999}
	b, err := mod.marshal()
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if err := os.WriteFile(dir+string(os.PathSeparator)+"level_1.json", b, 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)

	le := NewLevelEditor(39, 83)
	le.LoadLevel("level_1.json")
	got := le.Hash()
	if got == want {
		t.Errorf("the level file on disk was ignored")
	}
}
