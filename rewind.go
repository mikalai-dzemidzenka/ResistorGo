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
	Anim        string
	AnimFrame   uint8
}

func UpdateRewind(dt float32) {
	if rl.IsKeyPressed(rl.KeyR) {
		RewindIdx--
		RewindLeft--
	}
	if rl.IsKeyDown(rl.KeyR) && RewindLeft > 0 {
		IsRewinding = true
		RewindIdx--
		RewindLeft--
	} else {
		IsRewinding = false
		RewindIdx++
		if RewindLeft < RewindFrames {
			RewindLeft++
		}
	}
}
