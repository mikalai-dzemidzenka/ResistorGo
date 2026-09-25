package resistor

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// blocksIn makes a textures folder in a fresh working directory, the way a
// development build finds one beside it.
func blocksIn(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Chdir(dir)
	if err := os.MkdirAll(blocksDir, 0o755); err != nil {
		t.Fatal(err)
	}
	return dir
}

func realPNG(t *testing.T) []byte {
	t.Helper()
	b, err := assetFS.ReadFile(blocksDir + "/e1m1_wall1.png")
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestObjectDefaultsRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "level_objects.json")
	want := map[string]ObjectRects{
		"both":  {Collision: []rl.Rectangle{{X: 1, Y: 2, Width: 3, Height: 4}}, Light: []rl.Rectangle{{X: 5, Y: 6, Width: 7, Height: 8}}},
		"no-lt": {Collision: []rl.Rectangle{{X: -1, Width: 1, Height: 1}}},
		"three": {Collision: []rl.Rectangle{{Width: 1, Height: 1}, {X: 2, Width: 1, Height: 1}, {X: 4, Width: 1, Height: 1}}},
		"none":  {},
	}
	if !saveObjectDefaults(path, want) {
		t.Fatal("saving the defaults failed")
	}
	got := loadObjectDefaults(path)
	if len(got) != len(want) {
		t.Fatalf("loaded %d entries, want %d: %+v", len(got), len(want), got)
	}
	for k, w := range want {
		g, ok := got[k]
		if !ok {
			t.Errorf("%s was lost", k)
			continue
		}
		if !sameRects(g, w) {
			t.Errorf("%s = %+v, want %+v", k, g, w)
		}
	}
}

// Nothing edited writes nothing.
func TestSaveObjectDefaultsWritesNothingWhenEmpty(t *testing.T) {
	path := filepath.Join(t.TempDir(), "level_objects.json")
	if !saveObjectDefaults(path, nil) {
		t.Errorf("saving nothing reported a failure")
	}
	if _, err := os.Stat(path); err == nil {
		t.Errorf("saving nothing wrote a file")
	}
}

// A corrupt defaults file must not stop the editor: it opens on the built-in
// rectangles.
func TestMalformedDefaultsFallBack(t *testing.T) {
	blocksIn(t)
	if err := os.WriteFile(objectDefaultsPath, []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := loadObjectDefaults(objectDefaultsPath); got != nil {
		t.Errorf("malformed defaults loaded as %+v, want the fallback", got)
	}
	le := NewLevelEditor(39, 83)
	wall := blocksDir + "/e1m1_wall1.png"
	w := paletteEntry(t, le.levelObjects, wall)
	if len(w.CollisionRects) != 1 || !slices.Equal(w.CollisionRects, builtinRects()[wall].Collision) {
		t.Errorf("with a corrupt defaults file the wall has %+v, want its built-in rectangle", w.CollisionRects)
	}
}

// A texture that is only on disk is placeable but marked: a build made now
// would not carry it.
func TestDiskOnlyTextureIsMarked(t *testing.T) {
	blocksIn(t)
	fresh := blocksDir + "/fresh.png"
	if err := os.WriteFile(fresh, realPNG(t), 0o644); err != nil {
		t.Fatal(err)
	}
	le := NewLevelEditor(39, 83)
	paletteEntry(t, le.levelObjects, fresh)
	if inBuild(fresh) {
		t.Errorf("a texture only on disk is reported as in the build")
	}
	if !inBuild(blocksDir + "/e1m1_wall1.png") {
		t.Errorf("an embedded texture is reported as not in the build")
	}
}

// A dropped file whose name is taken never overwrites the one there; that
// entry is selected instead.
func TestImportNameTakenSelectsExisting(t *testing.T) {
	dir := blocksIn(t)
	taken := blocksDir + "/taken.png"
	if err := os.WriteFile(taken, []byte("original"), 0o644); err != nil {
		t.Fatal(err)
	}
	src := filepath.Join(dir, "incoming", "taken.png")
	os.MkdirAll(filepath.Dir(src), 0o755)
	if err := os.WriteFile(src, realPNG(t), 0o644); err != nil {
		t.Fatal(err)
	}
	le := NewLevelEditor(39, 83)
	if !le.importTexture(src) {
		t.Fatalf("importing a taken name did not select the existing entry")
	}
	if b, _ := os.ReadFile(taken); string(b) != "original" {
		t.Errorf("the existing texture was overwritten")
	}
	if le.obj.Resource != taken || !le.sel.proto {
		t.Errorf("selected %q, want the existing %q", le.obj.Resource, taken)
	}
}

// A file that does not decode is not copied, so it cannot poison the next
// run's palette, and the selection is left alone.
func TestImportUndecodableCopiesNothing(t *testing.T) {
	dir := blocksIn(t)
	src := filepath.Join(dir, "garbage.png")
	if err := os.WriteFile(src, []byte("this is not a png"), 0o644); err != nil {
		t.Fatal(err)
	}
	txt := filepath.Join(dir, "notes.txt")
	os.WriteFile(txt, []byte("hello"), 0o644)

	le := NewLevelEditor(39, 83)
	before, n := le.obj.Resource, len(le.levelObjects)
	for _, f := range []string{src, txt} {
		if le.importTexture(f) {
			t.Errorf("imported %s", f)
		}
	}
	if _, err := os.Stat(blocksDir + "/garbage.png"); err == nil {
		t.Errorf("a file that does not decode was copied in")
	}
	if le.obj.Resource != before || len(le.levelObjects) != n {
		t.Errorf("a failed import changed the palette or the selection")
	}
}

// Ctrl+S writes the level and the palette defaults, and clears the mark.
func TestSaveWritesLevelAndDefaults(t *testing.T) {
	blocksIn(t)
	le := NewLevelEditor(39, 83)
	le.setActive(&rl.Rectangle{Width: 7, Height: 7})
	if !le.dirty {
		t.Fatal("a default edit did not mark the editor unsaved")
	}
	le.usePreset(2, true)
	le.Save()
	for _, p := range []string{"level_1.json", objectDefaultsPath, presetsPath} {
		if b, err := os.ReadFile(p); err != nil || len(b) == 0 {
			t.Errorf("saving did not write %s: %v", p, err)
		}
	}
	if le.dirty {
		t.Errorf("still marked unsaved after saving")
	}
	got := loadObjectDefaults(objectDefaultsPath)[le.obj.Resource]
	if len(got.Collision) != 1 || got.Collision[0] != (rl.Rectangle{Width: 7, Height: 7}) {
		t.Errorf("saved default = %+v, want the edited rectangle", got.Collision)
	}
	if p := NewLevelEditor(39, 83).presets[2]; p == nil || len(p.Collision) != 1 || p.Collision[0] != (rl.Rectangle{Width: 7, Height: 7}) {
		t.Errorf("reopened preset 3 = %+v, want the edited rectangle", p)
	}
}

func TestPresetsRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "level_presets.json")
	var want Presets
	want[0] = &ObjectRects{Collision: coll(3, 4)}
	want[8] = &ObjectRects{Light: coll(5, 6), Drop: coll(7, 8)}
	if !savePresets(path, want) {
		t.Fatal("saving the presets failed")
	}
	got := loadPresets(path)
	for i := range got {
		if (got[i] == nil) != (want[i] == nil) || got[i] != nil && !sameRects(*got[i], *want[i]) {
			t.Errorf("slot %d = %+v, want %+v", i+1, got[i], want[i])
		}
	}
}

// The save dialog takes a bare file name, adds .json, and refuses paths and the
// editor's and game's own files.
func TestLevelFileName(t *testing.T) {
	for in, want := range map[string]string{
		"arena":         "arena.json",
		" arena.json ":  "arena.json",
		"Arena.JSON":    "Arena.JSON",
		"level_1.json":  "level_1.json",
		"":              "",
		"../x":          "",
		`dir\x.json`:    "",
		"C:x.json":      "",
		"settings.json": "",
		"level_objects": "",
		"LEVEL_PRESETS": "",
	} {
		got, err := levelFileName(in)
		if got != want || (want == "") != (err != nil) {
			t.Errorf("levelFileName(%q) = %q, %v; want %q", in, got, err, want)
		}
	}
}

// Saving from the dialog writes the named file and makes it the level Ctrl+S
// writes; a refused name changes nothing and keeps the dialog up.
func TestSaveDialogSavesAs(t *testing.T) {
	t.Chdir(t.TempDir())
	prev := LevelPath
	defer func() { LevelPath = prev }()
	le := NewLevelEditor(39, 83)
	le.dirty, le.saveOpen = true, true

	le.saveTo("../escape")
	if LevelPath != prev || !le.saveOpen || le.saveErr == "" {
		t.Fatalf("a path was accepted: LevelPath %q, open %t, err %q", LevelPath, le.saveOpen, le.saveErr)
	}
	le.saveTo("arena")
	if _, err := os.Stat("arena.json"); err != nil {
		t.Fatalf("arena.json not written: %v", err)
	}
	if LevelPath != "arena.json" || le.saveOpen || le.dirty {
		t.Errorf("after saving: LevelPath %q, open %t, dirty %t", LevelPath, le.saveOpen, le.dirty)
	}
}
