package game

import (
	"encoding/json"
	"fmt"
	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	MinZoom = 4
	MaxZoom = 6
)

type Camera struct {
	Cam       rl.Camera2D `json:"cam"`
	ZoomSpeed float32     `json:"zoomSpeed,omitempty"`
}

var cam = &Camera{
	Cam: rl.Camera2D{
		Offset:   rl.Vector2{X: SW / 2, Y: SH / 2},
		Target:   p.Pos,
		Rotation: 0,
		Zoom:     MinZoom,
	},
}

func (c *Camera) MarshalJSON() ([]byte, error) {
	type C Camera
	return json.Marshal(struct {
		Type EntityType `json:"typ"`
		*C
	}{
		Type: ENTITY_CAMERA,
		C:    (*C)(c),
	})
}

func (c *Camera) Update(dt float32) {
	c.ZoomSpeed += rl.GetMouseWheelMove() * dt * 5
	c.Cam.Zoom += c.ZoomSpeed
	c.ZoomSpeed = moveTo0(c.ZoomSpeed, 1.2*dt)
	if !Debug {
		c.Cam.Zoom = clamp(c.Cam.Zoom, MinZoom, MaxZoom)
	} else {
		c.Cam.Zoom = clamp(c.Cam.Zoom, 0.5, MaxZoom)
	}
	c.Cam.Target = rl.Vector2Add(p.Pos, rl.Vector2{X: 50})
}

func (c *Camera) Debug() string {
	return fmt.Sprintf("zoom: %f\nzoom_speed: %f", c.Cam.Zoom, c.ZoomSpeed)
}
