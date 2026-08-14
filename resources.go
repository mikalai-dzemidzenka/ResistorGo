package resistor

import rl "github.com/gen2brain/raylib-go/raylib"

var (
	resources = make(map[string]rl.Texture2D)
)

func LoadTextures() {
	load("resources/blocks/e1m1_wall1.png")
	load("resources/blocks/e1m1_platform1.png")
	load("resources/blocks/e1m1_door1.png")
	load("resources/threefour/run.png")
	load("resources/threefour/idle.png")
	load("resources/threefour/updown.png")
	load("resources/threefour/updown-0-1.png")
	load("resources/threefour/updown-0-2.png")
	load("resources/threefour/updown-1-0.png")
	load("resources/threefour/updown-1-2.png")
	load("resources/threefour/updown-2-0.png")
	load("resources/threefour/updown-2-1.png")
	load("resources/threefour/attack-high.png")
	load("resources/threefour/attack-mid.png")
	load("resources/threefour/attack-mid-2.png")
	load("resources/threefour/attack-low.png")
	load("resources/threefour/hit-top.png")
	load("resources/threefour/back-attack-low-1.png")
	load("resources/threefour/back-attack-mid-1.png")
	load("resources/threefour/back-attack-high-1.png")

	load("resources/threefour/block-low-start.png")
	load("resources/threefour/block-low-1.png")

	load("resources/threefour/block-mid-start.png")
	load("resources/threefour/block-mid-1.png")
	load("resources/threefour/block-mid-2.png")
}

func LoadLevelObjects() []LevelObject {
	return []LevelObject{
		{
			Resource: "resources/blocks/e1m1_wall1.png",
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

func load(path string) {
	resources[path] = rl.LoadTexture(path)
}
