package resistor

import "testing"

// Everything the game loads at runtime has to be in the binary: a name that is
// only on disk works in the checkout and fails for anyone who gets the exe alone.
func TestEveryLoadedAssetIsEmbedded(t *testing.T) {
	names := []string{
		"resources/blocks/e1m1_wall1.png",
		"resources/blocks/e1m1_platform1.png",
		"resources/blocks/e1m1_door1.png",
		"resources/main_menu/resistor.png",
		"resources/main_menu/head.obj",
		"resources/main_menu/jab1.obj",
		"resources/main_menu/jab2.obj",
		"resources/main_menu/kick.obj",
		"resources/main_menu/martelo.obj",
		"resources/style_dark.rgs",
		menuMusicPath,
		"level_1.json",
	}
	names = append(names, animTable...)
	for _, n := range names {
		if _, err := assetFS.ReadFile(n); err != nil {
			t.Errorf("not embedded: %s", n)
		}
	}
}

// The unused music tracks are 28 MB; embedding the directory wholesale would
// quadruple the binary for nothing.
func TestOnlyTheMenuTrackIsEmbedded(t *testing.T) {
	entries, err := assetFS.ReadDir("resources/music")
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	if len(entries) != 1 || "resources/music/"+entries[0].Name() != menuMusicPath {
		t.Errorf("embedded music = %v, want only %s", entries, menuMusicPath)
	}
}
