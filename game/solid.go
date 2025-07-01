package game

import (
	"encoding/json"
	rl "github.com/gen2brain/raylib-go/raylib"
)

func NewSolid(p rl.Vector2) *Solid {
	tex := Textures[TEXTURE_BLOCK_1]
	var c int32 = 87
	return &Solid{
		Z:   -1,
		Tex: TEXTURE_BLOCK_1,
		Pos: p,
		Col: rl.Rectangle{
			X:      p.X,
			Y:      p.Y + float32(c),
			Width:  float32(tex.Width),
			Height: float32(tex.Height - c),
		},
	}
}

type Solid struct {
	Z   int8
	Tex int32
	Pos rl.Vector2
	Col rl.Rectangle
}

func (s *Solid) GetZ() int8 {
	return s.Z
}

func (s *Solid) Draw() {
	tex := Textures[s.Tex]
	rl.DrawTexture(tex, int32(s.Pos.X), int32(s.Pos.Y), rl.White)
	if Debug {
		var c int32 = 87
		rl.DrawRectangle(int32(s.Pos.X), int32(s.Pos.Y)+c, tex.Width, tex.Height-c, DebugColor1)
		rl.DrawRectangleLinesEx(s.Col, 1, DebugColor2)
	}
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
