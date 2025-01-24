package game

import (
	"fmt"
	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	MinZoom = 4
	MaxZoom = 6
)

type Camera struct {
	cam       rl.Camera2D
	zoomSpeed float32
}

var cam = &Camera{
	cam: rl.Camera2D{
		Offset:   rl.Vector2{X: SW / 2, Y: SH / 2},
		Target:   p.pos,
		Rotation: 0,
		Zoom:     MinZoom,
	},
}

func (c *Camera) Update(dt float32) {
	c.zoomSpeed += rl.GetMouseWheelMove() * dt * 5
	c.cam.Zoom += c.zoomSpeed
	c.zoomSpeed = moveTo0(c.zoomSpeed, 1.2*dt)
	if !Debug {
		c.cam.Zoom = clamp(c.cam.Zoom, MinZoom, MaxZoom)
	} else {
		c.cam.Zoom = clamp(c.cam.Zoom, 0.5, MaxZoom)
	}
	c.cam.Target = rl.Vector2Add(p.pos, rl.Vector2{X: 50})
}

func (c *Camera) Debug() string {
	return fmt.Sprintf("zoom: %f\nzoom_speed: %f", c.cam.Zoom, c.zoomSpeed)
}
