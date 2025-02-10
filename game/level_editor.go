package game

import (
	"fmt"
	rlg "github.com/gen2brain/raylib-go/raygui"
	rl "github.com/gen2brain/raylib-go/raylib"
	"image/color"
	"slices"
)

type LevelBlock = rl.Vector2

type LevelEntity struct {
	Type EntityType `json:"type"`
	Tex  uint16     `json:"tex"`
	Pos  rl.Vector2 `json:"pos"`
}

type LevelLayer struct {
	Z                   int
	Entities            []LevelEntity
	EntityPositionIndex map[LevelBlock]uint16 `json:"-"`
}

type LevelEditor struct {
	LevelLayers  []*LevelLayer
	CurrentLayer int

	GridSize   rl.Vector2
	CurEntity  LevelEntity
	ViewOpened bool
	Z          int8
	GUIZ       int8
}

var _ Updatable = (*LevelEditor)(nil)
var _ GUIDrawable = (*LevelEditor)(nil)
var _ Drawable = (*LevelEditor)(nil)

func (le *LevelEditor) GetZ() int8 {
	return le.Z
}

func (le *LevelEditor) GetGUIZ() int8 {
	return le.GUIZ
}
func (le *LevelEditor) Update(dt float32) {
	if rl.IsKeyPressed(rl.KeyTab) {
		le.ViewOpened = !le.ViewOpened
	}
	if le.ViewOpened {
		return
	}

	le.CurEntity.Pos = rl.GetScreenToWorld2D(g.Cursor.Pos, g.Camera.Cam)
	if le.GridSize != (rl.Vector2{}) {
		le.CurEntity.Pos = pos2block(le.CurEntity.Pos)
	}
	switch {
	case rl.IsMouseButtonPressed(rl.MouseButtonLeft):
		le.addEntity()
	case rl.IsMouseButtonPressed(rl.MouseButtonRight):
		le.delEntity()
	}
}

func (le *LevelEditor) Draw() {
	le.drawEntities()
	le.drawCurEntity()
	le.drawGrid()

	rl.DrawRectangle(int32(le.CurEntity.Pos.X), int32(le.CurEntity.Pos.Y), 2, 2, rl.Red)
	rl.DrawText(fmt.Sprintf("%v", le.CurEntity.Pos), int32(le.CurEntity.Pos.X), int32(le.CurEntity.Pos.Y), 2, rl.White)
}

func (le *LevelEditor) DrawGUI() {
	if le.ViewOpened {
		col := rl.GetColor(uint(rlg.GetStyle(rlg.DEFAULT, rlg.BACKGROUND_COLOR)))
		col.A = 127
		rl.DrawRectangle(25, 25, SW-50, SH-50, col)
		for i := 0; i < TEXTURE_LEN; i++ { // TODO editor entities
			tex := Textures[i]
			rl.DrawTexturePro(tex,
				rl.Rectangle{0, 0, float32(tex.Width), float32(tex.Height)},
				rl.Rectangle{float32((i * 100) + 25), 35, 100, 100},
				rl.Vector2{0, 0}, 0, rl.White)
			rlg.SetStyle(rlg.BUTTON, rlg.BASE_COLOR_NORMAL, int64(rl.ColorToInt(color.RGBA{0, 0, 0, 0})))
			rlg.SetStyle(rlg.BUTTON, rlg.BASE_COLOR_FOCUSED, int64(rl.ColorToInt(color.RGBA{0, 0, 0, 25})))

			if rlg.Button(rl.Rectangle{float32((i * 100) + 25), 35, 100, 100}, "") {
				le.CurEntity.Tex = uint16(i)
			}
		}

	}
}

func (le *LevelEditor) addEntity() {
	la := le.LevelLayers[le.CurrentLayer]
	i, ok := la.EntityPositionIndex[le.CurEntity.Pos]
	if ok {
		la.Entities[i] = le.CurEntity
	} else {
		la.Entities = append(la.Entities, le.CurEntity)
		le.reindex()
	}
}

func (le *LevelEditor) delEntity() {
	la := le.LevelLayers[le.CurrentLayer]
	i, ok := la.EntityPositionIndex[le.CurEntity.Pos]
	if ok {
		la.Entities[i] = la.Entities[len(la.Entities)-1]
		la.Entities = la.Entities[:len(la.Entities)-1]

		le.reindex()
	}
}

func (le *LevelEditor) reindex() {
	la := le.LevelLayers[le.CurrentLayer]

	slices.SortFunc(la.Entities, func(a, b LevelEntity) int {
		switch {
		case a.Pos.Y > b.Pos.Y:
			return -1
		case a.Pos.Y < b.Pos.Y:
			return 1
		case a.Pos.X > b.Pos.X:
			return -1
		case a.Pos.X < b.Pos.X:
			return 1
		default:
			return 0
		}
	})

	if la.EntityPositionIndex == nil {
		la.EntityPositionIndex = make(map[LevelBlock]uint16)
	}
	clear(la.EntityPositionIndex)
	for i, e := range la.Entities {
		la.EntityPositionIndex[e.Pos] = uint16(i)
	}
}

func (le *LevelEditor) drawEntities() {
	for _, la := range le.LevelLayers {
		for _, ent := range la.Entities {
			tex := Textures[ent.Tex]
			rl.DrawTextureRec(
				tex,
				rl.Rectangle{
					X:      0,
					Y:      0,
					Width:  float32(tex.Width),
					Height: float32(tex.Height),
				},
				ent.Pos,
				rl.White,
			)
		}
	}

}

func (le *LevelEditor) drawCurEntity() {
	tex := Textures[le.CurEntity.Tex]
	rl.DrawTextureRec(
		tex,
		rl.Rectangle{
			X:      0,
			Y:      0,
			Width:  float32(tex.Width),
			Height: float32(tex.Height),
		},
		le.CurEntity.Pos,
		color.RGBA{255, 255, 255, 128},
	)
}

func (le *LevelEditor) drawGrid() {
	rl.PushMatrix()
	{
		a := le.GridSize.Y / le.GridSize.X
		rl.Translatef(
			float32(int32(le.GridSize.X)+1),
			float32(int32(le.GridSize.Y)+(5*int32(le.GridSize.Y))),
			0)
		rl.Rotatef(90, 1, 0, 0)
		rl.Scalef(1, 0, a)
		rl.DrawGrid(100, le.GridSize.X)
	}
	rl.PopMatrix()
}

func pos2block(pos rl.Vector2) LevelBlock {
	pos.X = float32(int32(pos.X) - int32(pos.X)%int32(g.LevelEditor.GridSize.X)) // TODO sign error on negative pos here
	pos.Y = float32(int32(pos.Y) - int32(pos.Y)%int32(g.LevelEditor.GridSize.Y))
	return pos
}
