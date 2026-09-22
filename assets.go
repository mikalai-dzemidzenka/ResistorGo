package resistor

import "embed"

// assetFS holds every asset the game loads at runtime, so a shipped binary needs
// nothing beside it. Only what is actually used is embedded: the spare music
// tracks and the source .fbx would otherwise add ~28 MB of dead weight.
//
//go:embed resources/blocks/*.png
//go:embed resources/threefour/*.png
//go:embed resources/style_dark.rgs
//go:embed resources/main_menu/resistor.png
//go:embed resources/main_menu/head.obj
//go:embed resources/main_menu/jab1.obj
//go:embed resources/main_menu/jab2.obj
//go:embed resources/main_menu/kick.obj
//go:embed resources/main_menu/martelo.obj
//go:embed "resources/music/Neon Sprint.mp3"
//go:embed "resources/music/Final Run (1).mp3"
//go:embed level_1.json
var assetFS embed.FS

// asset returns an embedded asset's bytes. A missing name is a build mistake,
// not a runtime condition, so it panics instead of drawing nothing.
func asset(name string) []byte {
	b, err := assetFS.ReadFile(name)
	if err != nil {
		panic("embedded asset missing: " + name)
	}
	return b
}
