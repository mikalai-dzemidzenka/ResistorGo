package main

import rl "github.com/gen2brain/raylib-go/raylib"

const (
	MinZoom = 4
	MaxZoom = 6
)

type Camera struct {
	cam rl.Camera2D
}

var c = Camera{
	cam: rl.Camera2D{
		Offset:   rl.Vector2{X: SW / 2, Y: SH / 2},
		Target:   p.pos,
		Rotation: 0,
		Zoom:     MinZoom,
	},
}

func (c *Camera) Update(dt float32) {
	c.cam.Zoom += rl.GetMouseWheelMove() * dt * 5 // TODO lerp? zoom speed?
	c.cam.Zoom = clamp(c.cam.Zoom, MinZoom, MaxZoom)
	c.cam.Target = p.pos
}
