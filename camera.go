package resistor

import rl "github.com/gen2brain/raylib-go/raylib"

const (
	MaxZoom = 5
	MinZoom = 1
)

type Camera struct {
	rl.Camera2D
	Target    rl.Vector2
	ZoomSpeed float32
}

func NewCamera(zoom float32) Camera {
	offset := rl.Vector2{SW / 2, SH / 2}
	cam := rl.NewCamera2D(offset, ZeroVec2, 0, zoom)
	return Camera{Camera2D: cam, ZoomSpeed: 0}
}

func (c *Camera) SetTarget(coord rl.Vector2) {
	c.Target = coord
}

func (c *Camera) Update(dt float32) {
	c.ZoomSpeed += rl.GetMouseWheelMove() * 0.2
	c.Zoom += c.ZoomSpeed * dt

	c.ZoomSpeed = moveTo0(c.ZoomSpeed, 0.025*dt)

	c.Camera2D.Target.X = moveTo(c.Camera2D.Target.X, 1000*dt, c.Target.X)
	c.Camera2D.Target.Y = moveTo(c.Camera2D.Target.Y, 1000*dt, c.Target.Y)

	c.Zoom = clamp(c.Zoom, MinZoom, MaxZoom)
	c.ZoomSpeed = clamp(c.ZoomSpeed, -2, 2)
	if c.Zoom == MaxZoom || c.Zoom == MinZoom {
		c.ZoomSpeed = 0
	}
}
