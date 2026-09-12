package resistor

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	MaxZoom = 5
	MinZoom = 1
)

type Camera struct {
	rl.Camera2D
	Target    rl.Vector2
	ZoomSpeed float32
}

func NewCamera(zoom float32, offset rl.Vector2) Camera {
	cam := rl.NewCamera2D(offset, ZeroVec2, 0, zoom)
	return Camera{Camera2D: cam, ZoomSpeed: 0}
}

func (c *Camera) SetTarget(coord rl.Vector2) {
	c.Target = coord
}

// ZoomStep zooms by whole steps, one per wheel notch, at once. Between the
// whole numbers there is no zoom level worth showing - the view is quantised to
// them anyway (see PixelSize) - so the wheel snaps straight to the next one
// instead of feeding momentum. It also cancels any zoom effect still running:
// what the player asks for beats what a punch left behind.
func (c *Camera) ZoomStep(notches float32) {
	if notches == 0 {
		return
	}
	step := roundf(notches)
	if step == 0 { // a trackpad's fraction of a notch is still one step
		step = 1
		if notches < 0 {
			step = -1
		}
	}
	c.Zoom = clamp(roundf(c.Zoom)+step, MinZoom, MaxZoom)
	c.ZoomSpeed = 0
}

func (c *Camera) Update(dt float32) {
	c.ZoomStep(rl.GetMouseWheelMove())
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

// PixelSize is how many screen pixels one texel covers. It is a whole number:
// the palette shader quantises the screen into blocks of this size, and a
// fractional one makes those blocks alternate between two widths.
func (c *Camera) PixelSize() float32 {
	return clamp(roundf(c.Zoom), MinZoom, MaxZoom)
}

// View is the camera to draw and pick with: a whole number zoom, a target on
// whole world pixels and an offset on whole blocks. Any fraction left in those
// puts the texel grid out of phase with the shader's block grid, so texel rows
// get doubled or dropped at random and the art crawls while the camera zooms
// or moves. The camera keeps its smooth values; only the view is snapped.
func (c *Camera) View() rl.Camera2D {
	v := c.Camera2D
	v.Zoom = c.PixelSize()
	v.Target = rl.Vector2{X: roundf(v.Target.X), Y: roundf(v.Target.Y)}
	v.Offset = rl.Vector2{X: roundf(v.Offset.X/v.Zoom) * v.Zoom, Y: roundf(v.Offset.Y/v.Zoom) * v.Zoom}
	return v
}
