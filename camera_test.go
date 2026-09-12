package resistor

import (
	"math"
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// screenX is where a world pixel lands on screen under a Camera2D.
func screenX(v rl.Camera2D, worldX float32) float32 {
	return (worldX-v.Target.X)*v.Zoom + v.Offset.X
}

// Whatever the camera's smooth values are, the view it is drawn with must put
// every whole world pixel on a whole block of the palette shader's grid.
func TestCameraViewSnapsToPixelGrid(t *testing.T) {
	cam := NewCamera(3.37, rl.Vector2{X: 578, Y: 362}) // mid-zoom, shaken offset
	cam.Camera2D.Target = rl.Vector2{X: 100.62, Y: -40.25}

	v := cam.View()
	if v.Zoom != 3 {
		t.Errorf("view zoom %v, want a whole number of pixels per texel", v.Zoom)
	}
	if v.Target.X != 101 || v.Target.Y != -40 {
		t.Errorf("view target %+v, want whole world pixels", v.Target)
	}
	if math.Mod(float64(v.Offset.X), float64(v.Zoom)) != 0 || math.Mod(float64(v.Offset.Y), float64(v.Zoom)) != 0 {
		t.Errorf("view offset %+v is not a whole number of %v pixel blocks", v.Offset, v.Zoom)
	}
	for _, worldX := range []float32{-97, 0, 13, 512} {
		if got := screenX(v, worldX); math.Mod(float64(got), float64(v.Zoom)) != 0 {
			t.Errorf("world x=%v draws at screen x=%v, off the %v pixel grid", worldX, got, v.Zoom)
		}
	}
	if cam.Zoom != 3.37 || cam.Camera2D.Target.X != 100.62 {
		t.Errorf("snapping the view disturbed the camera itself: zoom %v target %+v", cam.Zoom, cam.Camera2D.Target)
	}
}

func TestCameraPixelSize(t *testing.T) {
	for _, c := range []struct{ zoom, want float32 }{
		{1, 1}, {1.4, 1}, {1.5, 2}, {3.37, 3}, {4.9, 5}, {5, 5},
		{0.2, MinZoom}, {9, MaxZoom}, // out of range, e.g. mid-clamp
	} {
		cam := NewCamera(c.zoom, ZeroVec2)
		if got := cam.PixelSize(); got != c.want {
			t.Errorf("zoom %v: pixel size %v, want %v", c.zoom, got, c.want)
		}
	}
}

// The wheel zooms by whole steps, at once.
func TestZoomStep(t *testing.T) {
	for _, c := range []struct {
		name          string
		from, notches float32
		want          float32
	}{
		{"one notch in", 3, 1, 4},
		{"one notch out", 3, -1, 2},
		{"several notches at once", 1, 3, 4},
		{"clamped at the top", MaxZoom, 1, MaxZoom},
		{"clamped at the bottom", MinZoom, -1, MinZoom},
		{"from a zoom an effect left fractional", 3.4, 1, 4},
		{"a trackpad fraction still steps", 3, 0.2, 4},
		{"a trackpad fraction out still steps", 3, -0.2, 2},
		{"no wheel, no change", 3, 0, 3},
	} {
		cam := NewCamera(c.from, ZeroVec2)
		cam.ZoomStep(c.notches)
		if cam.Zoom != c.want {
			t.Errorf("%s: zoom %v + %v notches = %v, want %v", c.name, c.from, c.notches, cam.Zoom, c.want)
		}
	}
}

// A zoom effect (the attack punch) still runs on momentum, and the wheel
// overrides it: the step lands whole and the leftover momentum is dropped.
func TestZoomStepCancelsEffect(t *testing.T) {
	cam := NewCamera(3, ZeroVec2)
	cam.ZoomSpeed = AttackZoomSpeed
	cam.Update(1.0 / 60)
	if cam.Zoom <= 3 {
		t.Errorf("the attack zoom effect did not move the camera: %v", cam.Zoom)
	}
	cam.ZoomStep(-1)
	if cam.Zoom != 2 || cam.ZoomSpeed != 0 {
		t.Errorf("after a notch out: zoom %v speed %v, want zoom 2 and no momentum", cam.Zoom, cam.ZoomSpeed)
	}
}
