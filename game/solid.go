package game

import (
	"encoding/json"
	rl "github.com/gen2brain/raylib-go/raylib"
)

type Solid struct {
	Z   int8
	Tex int32
	Pos rl.Vector2
}

func (s *Solid) GetZ() int8 {
	return s.Z
}

func (s *Solid) Draw() {
	rl.DrawTexture(Textures[s.Tex], int32(s.Pos.X), int32(s.Pos.Y), rl.White)
}

func (s *Solid) MarshalJSON() ([]byte, error) {
	type S Solid
	return json.Marshal(EntitySerde{
		Type:   ENTITY_SOLID,
		Entity: (*S)(s),
	})
}

var _ Drawable = (*Solid)(nil)
var _ Serializable = (*Solid)(nil)
