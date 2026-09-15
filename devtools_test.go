package resistor

import (
	"os"
	"testing"
)

// Where the level comes from, in both builds. With a modified level_1.json in
// the working directory, a development build plays it (so a saved edit takes
// effect) and a release build plays the embedded level (so every copy of a
// release agrees on the level fingerprint). The dedicated server loads the
// same way.
func TestLoadLevelSource(t *testing.T) {
	embedded := func() *LevelEditor {
		le := NewLevelEditor(39, 83)
		b, err := assetFS.ReadFile("level_1.json")
		if err != nil {
			t.Fatal(err)
		}
		objects, _, _, err := parseLevel(b)
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
	switch {
	case DevTools && got == want:
		t.Errorf("a development build ignored the level file on disk")
	case !DevTools && got != want:
		t.Errorf("a release build played the level file on disk instead of the embedded level")
	}
}
