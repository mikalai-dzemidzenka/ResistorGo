package resistor

import (
	"path"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var (
	resources = make(map[string]rl.Texture2D)
	models    = make(map[string]rl.Model)
)

// animTable lists every actor animation. The index is the animation's network id,
// so the order must be identical on client and server (append only).
var animTable = []string{
	"resources/threefour/run.png",
	"resources/threefour/idle.png",
	"resources/threefour/jump-start.png",
	"resources/threefour/jump-climb.png",
	"resources/threefour/updown.png",
	"resources/threefour/updown-0-1.png",
	"resources/threefour/updown-0-2.png",
	"resources/threefour/updown-1-0.png",
	"resources/threefour/updown-1-2.png",
	"resources/threefour/updown-2-0.png",
	"resources/threefour/updown-2-1.png",
	"resources/threefour/updown-atk-2-1.png",
	"resources/threefour/updown-atk-2-0.png",
	"resources/threefour/updown-atk-1-2.png",
	"resources/threefour/updown-atk-1-0.png",
	"resources/threefour/updown-atk-0-1.png",
	"resources/threefour/updown-atk-0-2.png",
	"resources/threefour/attack-high.png",
	"resources/threefour/attack-mid.png",
	"resources/threefour/attack-mid-2.png",
	"resources/threefour/attack-low-2.png",
	"resources/threefour/attack-high-2.png",
	"resources/threefour/attack-high-3.png",
	"resources/threefour/attack-low.png",
	"resources/threefour/hit-top.png",
	"resources/threefour/death_fall.png",
	"resources/threefour/back-attack-low-1.png",
	"resources/threefour/back-attack-low-2.png",
	"resources/threefour/back-attack-mid-1.png",
	"resources/threefour/back-attack-high-1.png",
	"resources/threefour/back-attack-high-2.png",
	"resources/threefour/block-all-start.png",
	"resources/threefour/block-all-1.png",
	"resources/threefour/block-low-start.png",
	"resources/threefour/block-low-1.png",
	"resources/threefour/block-mid-start.png",
	"resources/threefour/block-mid-1.png",
	"resources/threefour/block-mid-2.png",
	"resources/threefour/block-mid-v2-start.png",
	"resources/threefour/block-mid-v2-1.png",
	"resources/threefour/block.png",
	"resources/threefour/walk.png",
	"resources/threefour/walk-bot.png",
	"resources/threefour/walk-mid.png",
	"resources/threefour/walk-top.png",
	"resources/threefour/turnover.png",
}

var animIDs = func() map[string]uint8 {
	m := make(map[string]uint8, len(animTable))
	for i, p := range animTable {
		m[p] = uint8(i)
	}
	return m
}()

func animID(path string) uint8 {
	id, ok := animIDs[path]
	if !ok {
		panic("animation not in animTable: " + path)
	}
	return id
}

func animPath(id uint8) string {
	if int(id) < len(animTable) {
		return animTable[id]
	}
	return animTable[animID("resources/threefour/updown.png")]
}

func LoadTextures() {
	loadTex("resources/blocks/e1m1_wall1.png")
	loadTex("resources/blocks/e1m1_platform1.png")
	loadTex("resources/blocks/e1m1_door1.png")
	for _, p := range animTable {
		loadTex(p)
	}

	loadTex("resources/main_menu/resistor.png")
	loadModel("resources/main_menu/head.obj")
	loadModel("resources/main_menu/jab1.obj")
	loadModel("resources/main_menu/jab2.obj")
	loadModel("resources/main_menu/kick.obj")
	loadModel("resources/main_menu/martelo.obj")
}

func LoadLevelObjects() []LevelObject {
	return []LevelObject{
		{
			Resource: "resources/blocks/e1m1_wall1.png",
			Layer:    0,
			X:        0,
			Y:        0,
			CollisionRect: &rl.Rectangle{
				X:      15,
				Y:      86,
				Width:  40,
				Height: 20,
			},
			LightRect: &rl.Rectangle{
				X:      22,
				Y:      90,
				Width:  40,
				Height: 15,
			},
		}, {
			Resource: "resources/blocks/e1m1_door1.png",
			Layer:    0,
			X:        0,
			Y:        0,
			CollisionRect: &rl.Rectangle{
				X:      0,
				Y:      86,
				Width:  40,
				Height: 20,
			},
		}, {
			Resource: "resources/blocks/e1m1_platform1.png",
			Layer:    0,
			X:        0,
			Y:        0,
			CollisionRect: &rl.Rectangle{
				X:      0,
				Y:      86,
				Width:  40,
				Height: 20,
			},
		},
	}
}

// loadTex decodes an embedded image straight into a GPU texture, so nothing is
// read from disk at runtime.
func loadTex(name string) {
	data := asset(name)
	img := rl.LoadImageFromMemory(path.Ext(name), data, int32(len(data)))
	resources[name] = rl.LoadTextureFromImage(img)
	rl.UnloadImage(img)
}

func loadModel(name string) {
	models[name] = loadOBJ(name, asset(name))
}
