package resistor

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// tiles installs a level made of 48x48 solid tiles at the given top-left
// corners, replacing whatever the global level editor held.
func tiles(corners ...Vector2i) {
	LE = NewLevelEditor(39, 83)
	LE.objects[0] = make(LevelLayer)
	for _, c := range corners {
		LE.objects[0][c] = LevelObject{
			X: c.X, Y: c.Y,
			CollisionRect: &rl.Rectangle{Width: 48, Height: 48},
		}
	}
	LE.buildIndex()
}

// jumper is an actor rising with its head at headY and its collision box
// centered on centerX.
func jumper(centerX, headY float32) *Actor {
	a := NewActor(rl.Vector2{}, true)
	a.Pos.X = centerX - (a.CollisionBox.X + a.CollisionBox.Width/2)
	a.Pos.Y = headY - a.CollisionBox.Y
	a.Speed.Y = -5
	a.Orientation = 1
	return a
}

// Ledge climbing must not fire for a tile that is part of a ceiling: jumping
// into one used to teleport a right-facing actor on top of it, because tiles
// are visited right-to-left and the neighbour was mantled onto before the tile
// above the head was seen.
func TestNoClimbThroughCeiling(t *testing.T) {
	for _, orientation := range []int8{1, -1} {
		// A three tile ceiling row spanning x 0..144, head inside the middle
		// tile and close enough to a neighbour to trigger a climb.
		tiles(Vector2i{X: 0}, Vector2i{X: 48}, Vector2i{X: 96})
		a := jumper(90, 16)
		a.Orientation = orientation
		a.checkCollision()
		if want := float32(48); a.Pos.Y != want {
			t.Errorf("orientation %d: head stopped at y=%v, want pushed below the ceiling (%v)", orientation, a.Pos.Y, want)
		}
		if a.CanClimbRight || a.CanClimbLeft {
			t.Errorf("orientation %d: climbing a ceiling tile (right=%t left=%t)", orientation, a.CanClimbRight, a.CanClimbLeft)
		}
	}
}

// A real ledge - a tile with nothing above the actor's head - is still climbable.
func TestClimbLedge(t *testing.T) {
	tiles(Vector2i{X: 96})
	a := jumper(90, 16) // just left of the tile, head level with it
	a.checkCollision()
	if !a.CanClimbRight {
		t.Errorf("right-facing actor did not grab the ledge")
	}
	if want := float32(-47); a.Pos.Y != want {
		t.Errorf("climbed to y=%v, want %v", a.Pos.Y, want)
	}
}
