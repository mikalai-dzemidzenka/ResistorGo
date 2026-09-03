package resistor

import rl "github.com/gen2brain/raylib-go/raylib"

var (
	resources = make(map[string]rl.Texture2D)
	models    = make(map[string]rl.Model)
)

func LoadTextures() {
	loadTex("resources/blocks/e1m1_wall1.png")
	loadTex("resources/blocks/e1m1_platform1.png")
	loadTex("resources/blocks/e1m1_door1.png")
	loadTex("resources/threefour/run.png")
	loadTex("resources/threefour/idle.png")
	loadTex("resources/threefour/jump-start.png")
	loadTex("resources/threefour/jump-climb.png")
	loadTex("resources/threefour/updown.png")
	loadTex("resources/threefour/updown-0-1.png")
	loadTex("resources/threefour/updown-0-2.png")
	loadTex("resources/threefour/updown-1-0.png")
	loadTex("resources/threefour/updown-1-2.png")
	loadTex("resources/threefour/updown-2-0.png")
	loadTex("resources/threefour/updown-2-1.png")
	loadTex("resources/threefour/attack-high.png")
	loadTex("resources/threefour/attack-mid.png")
	loadTex("resources/threefour/attack-mid-2.png")
	loadTex("resources/threefour/attack-low.png")
	loadTex("resources/threefour/hit-top.png")
	loadTex("resources/threefour/back-attack-low-1.png")
	loadTex("resources/threefour/back-attack-mid-1.png")
	loadTex("resources/threefour/back-attack-high-1.png")

	loadTex("resources/threefour/block-low-start.png")
	loadTex("resources/threefour/block-low-1.png")

	loadTex("resources/threefour/block-mid-start.png")
	loadTex("resources/threefour/block-mid-1.png")
	loadTex("resources/threefour/block-mid-2.png")
	loadTex("resources/threefour/walk.png")
	loadTex("resources/threefour/turnover.png")

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

func loadTex(path string) {
	resources[path] = rl.LoadTexture(path)
}

func loadModel(path string) {
	models[path] = rl.LoadModel(path)
}
