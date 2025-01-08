package main

import (
	"flag"
	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	SW = 1024
	SH = 768
)

var l int

const (
	LEVEL_1 = iota
	LEVEL_EDITOR
	LEVELS
)

func main() {
	rl.InitWindow(SW, SH, "Resistor")
	defer rl.CloseWindow()

	flag.IntVar(&l, "l", LEVEL_EDITOR, "level to load")

	rl.DisableCursor()

	LoadTextures()

	rl.SetTargetFPS(60)

	for !rl.WindowShouldClose() {
		if rl.IsKeyPressed(rl.KeyGrave) {
			Debug = !Debug
		}

		dt := rl.GetFrameTime()

		switch l {
		case LEVEL_EDITOR:
			lvlEditLoop(dt)
		default:
			lvl1Loop(dt)
		}
	}
}
