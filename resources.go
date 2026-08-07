package resistor

import rl "github.com/gen2brain/raylib-go/raylib"

var (
	resources = make(map[string]rl.Texture2D)
)

func LoadTextures() {
	load("resources/blocks/e1m1_wall1.png")
	load("resources/threefour/run.png")
	load("resources/threefour/idle.png")
	load("resources/threefour/updown.png")
	load("resources/threefour/updown-0-1.png")
	load("resources/threefour/updown-0-2.png")
	load("resources/threefour/updown-1-0.png")
	load("resources/threefour/updown-1-2.png")
	load("resources/threefour/updown-2-0.png")
	load("resources/threefour/updown-2-1.png")
}

func load(path string) {
	resources[path] = rl.LoadTexture(path)
}
