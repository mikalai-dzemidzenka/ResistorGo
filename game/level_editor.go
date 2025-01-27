package game

import (
	rlg "github.com/gen2brain/raylib-go/raygui"
	rl "github.com/gen2brain/raylib-go/raylib"
	"image/color"
	"slices"
)

var le = &LevelEditor{
	entities: []LevelEntity{},
	idx:      make(map[LevelBlock]uint16),
	curEntity: LevelEntity{
		typ: ENTITY_SOLID,
		tex: TEXTURE_BLOCK_1,
		pos: rl.Vector2{},
	},
	gridSize: rl.Vector2{X: 39, Y: 83},
	z:        -1,
	guiZ:     1,
}

type LevelBlock = rl.Vector2

// TODO must be serializable
type LevelEntity struct {
	typ EntityType
	tex uint16
	pos rl.Vector2
}

type LevelEditor struct {
	entities []LevelEntity
	idx      map[LevelBlock]uint16

	gridSize      rl.Vector2
	curEntity     LevelEntity
	entViewOpened bool
	z             int8
	guiZ          int8
}

var _ Updatable = (*LevelEditor)(nil)
var _ GUIDrawable = (*LevelEditor)(nil)
var _ Drawable = (*LevelEditor)(nil)

func (le *LevelEditor) Z() int8 {
	return le.z
}

func (le *LevelEditor) GuiZ() int8 {
	return le.guiZ
}
func (le *LevelEditor) Update(dt float32) {
	if rl.IsKeyPressed(rl.KeyTab) {
		le.entViewOpened = !le.entViewOpened
	}
	if le.entViewOpened {
		return
	}

	le.curEntity.pos = rl.GetScreenToWorld2D(rl.GetMousePosition(), cam.cam)
	if le.gridSize != (rl.Vector2{}) {
		le.curEntity.pos = pos2block(le.curEntity.pos)
	}
	if rl.IsMouseButtonPressed(rl.MouseButtonLeft) {
		le.addEntity()
	}
}

func (le *LevelEditor) Draw() {
	le.drawEntities()
	le.drawCurEntity()
	le.drawGrid()
	rl.DrawRectangle(int32(le.curEntity.pos.X), int32(le.curEntity.pos.Y), 2, 2, rl.Red)

}

func (le *LevelEditor) DrawGui() {
	if le.entViewOpened {
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
				le.curEntity.tex = uint16(i)
			}
		}

	}
}

func (le *LevelEditor) addEntity() {
	idx, ok := le.idx[le.curEntity.pos]
	if ok {
		le.entities[idx] = le.curEntity
	} else {
		le.entities = append(le.entities, le.curEntity)
		idx = uint16(len(le.entities) - 1)
		le.idx[le.curEntity.pos] = idx
	}

	slices.SortFunc(le.entities, func(a, b LevelEntity) int {
		switch {
		case a.pos.Y > b.pos.Y:
			return -1
		case a.pos.Y < b.pos.Y:
			return 1
		case a.pos.X > b.pos.X:
			return -1
		case a.pos.X < b.pos.X:
			return 1
		default:
			return 0
		}
	})
}

func (le *LevelEditor) drawEntities() {
	for _, ent := range le.entities {
		tex := Textures[ent.tex]
		rl.DrawTextureRec(
			tex,
			rl.Rectangle{
				X:      0,
				Y:      0,
				Width:  float32(tex.Width),
				Height: float32(tex.Height),
			},
			ent.pos,
			rl.White,
		)
	}
}

func (le *LevelEditor) drawCurEntity() {
	tex := Textures[le.curEntity.tex]
	rl.DrawTextureRec(
		tex,
		rl.Rectangle{
			X:      0,
			Y:      0,
			Width:  float32(tex.Width),
			Height: float32(tex.Height),
		},
		le.curEntity.pos,
		color.RGBA{255, 255, 255, 128},
	)
}

func (le *LevelEditor) drawGrid() {
	rl.PushMatrix()
	{
		a := le.gridSize.Y / le.gridSize.X
		rl.Translatef(
			float32(int32(le.gridSize.X)+1),
			float32(int32(le.gridSize.Y)+(5*int32(le.gridSize.Y))),
			0)
		rl.Rotatef(90, 1, 0, 0)
		rl.Scalef(1, 0, a)
		rl.DrawGrid(100, le.gridSize.X)
	}
	rl.PopMatrix()
}

func pos2block(pos rl.Vector2) LevelBlock {
	pos.X = float32(int32(pos.X) - int32(pos.X)%int32(le.gridSize.X))
	pos.Y = float32(int32(pos.Y) - int32(pos.Y)%int32(le.gridSize.Y))
	return pos
}
