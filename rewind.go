package resistor

import rl "github.com/gen2brain/raylib-go/raylib"

const (
	RewindFrames = 256
)

var (
	IsRewinding = false
	RewindIdx   = 0
	RewindLeft  = 0
)

type ActorRewindFrame struct {
	Pos         rl.Vector2
	Speed       rl.Vector2
	Orientation int8
}
