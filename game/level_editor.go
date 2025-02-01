package game

import (
	rlg "github.com/gen2brain/raylib-go/raygui"
	rl "github.com/gen2brain/raylib-go/raylib"
	"image/color"
	"slices"
)

var le = &LevelEditor{
	Entities:            []LevelEntity{},
	EntityPositionIndex: make(map[LevelBlock]uint16),
	CurEntity: LevelEntity{
		Typ: ENTITY_SOLID,
		Tex: TEXTURE_BLOCK_1,
		Pos: rl.Vector2{},
	},
	GridSize: rl.Vector2{X: 39, Y: 83},
	Z:        -1,
	GUIZ:     1,
}

type LevelBlock = rl.Vector2

type LevelEntity struct {
	Typ EntityType
	Tex uint16
	Pos rl.Vector2
}

type LevelEditor struct {
	Entities            []LevelEntity
	EntityPositionIndex map[LevelBlock]uint16

	GridSize   rl.Vector2
	CurEntity  LevelEntity
	ViewOpened bool
	Z          int8
	GUIZ       int8
}

func (le *LevelEditor) NoSerialize() {}

var _ NoSerialize = (*LevelEditor)(nil)
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

	le.CurEntity.Pos = rl.GetScreenToWorld2D(rl.GetMousePosition(), cam.Cam)
	if le.GridSize != (rl.Vector2{}) {
		le.CurEntity.Pos = pos2block(le.CurEntity.Pos)
	}
	if rl.IsMouseButtonPressed(rl.MouseButtonLeft) {
		le.addEntity()
	}
}

func (le *LevelEditor) Draw() {
	le.drawEntities()
	le.drawCurEntity()
	le.drawGrid()

	rl.DrawRectangle(int32(le.CurEntity.Pos.X), int32(le.CurEntity.Pos.Y), 2, 2, rl.Red)
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
	i, ok := le.EntityPositionIndex[le.CurEntity.Pos]
	if ok {
		le.Entities[i] = le.CurEntity
	} else {
		le.Entities = append(le.Entities, le.CurEntity)
	}

	slices.SortFunc(le.Entities, func(a, b LevelEntity) int {
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

	clear(le.EntityPositionIndex)
	for i, e := range le.Entities {
		le.EntityPositionIndex[e.Pos] = uint16(i)
	}
}

func (le *LevelEditor) drawEntities() {
	for _, ent := range le.Entities {
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
	pos.X = float32(int32(pos.X) - int32(pos.X)%int32(le.GridSize.X))
	pos.Y = float32(int32(pos.Y) - int32(pos.Y)%int32(le.GridSize.Y))
	return pos
}
