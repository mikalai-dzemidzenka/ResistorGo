package game

import rl "github.com/gen2brain/raylib-go/raylib"

type Enemy struct {
	Bound    rl.Rectangle `json:"bound"`
	Pos      rl.Vector2   `json:"pos"`
	Speed    rl.Vector2   `json:"speed"`
	Collider rl.Rectangle
	Gravity  float32
	Acc      float32
	OnGround bool
	Friction float32 `json:"friction,omitempty"`
	Anim     int32   `json:"anim,omitempty"`
	// Orient 1 right, -1 left
	Orient int8 `json:"orient,omitempty"`
	Z      int8 `json:"z,omitempty"`
}
