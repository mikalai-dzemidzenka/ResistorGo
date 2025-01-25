package game

import (
	rlg "github.com/gen2brain/raylib-go/raygui"
	rl "github.com/gen2brain/raylib-go/raylib"
	"slices"
)

var le = &LevelEditor{
	entities: []LevelEntity{},
	texID:    TEXTURE_BLOCK_1,
	pos:      rl.Vector2{},
	gridSize: rl.Vector2{X: 39, Y: 83},
	z:        -1,
	guiZ:     1,
}

type LevelEntity struct {
	pos   rl.Vector2
	texID int32
}

type LevelEditor struct {
	entities []LevelEntity
	pos      rl.Vector2
	gridSize rl.Vector2
	texID    int32
	z        int8
	guiZ     int8
}

func (le *LevelEditor) GuiZ() int8 {
	return le.guiZ
}

func (le *LevelEditor) Z() int8 {
	return le.z
}

func (le *LevelEditor) Update(dt float32) {
	le.pos = rl.GetScreenToWorld2D(rl.GetMousePosition(), cam.cam)
	if le.gridSize != (rl.Vector2{}) {
		le.pos.X = float32(int32(le.pos.X) - int32(le.pos.X)%int32(le.gridSize.X))
		le.pos.Y = float32(int32(le.pos.Y) - int32(le.pos.Y)%int32(le.gridSize.Y))
	}
	if rl.IsMouseButtonPressed(rl.MouseButtonLeft) {
		le.entities = append(le.entities, LevelEntity{
			pos:   le.pos,
			texID: le.texID,
		})
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
}

func (le *LevelEditor) drawEntities() {
	for i := range le.entities {
		tex := Textures[le.entities[i].texID]
		rl.DrawTextureRec(
			tex,
			rl.Rectangle{
				X:      0,
				Y:      0,
				Width:  float32(tex.Width),
				Height: float32(tex.Height),
			},
			le.entities[i].pos,
			rl.White,
		)
	}
}

func (le *LevelEditor) drawCurEntity() {
	tex := Textures[le.texID]
	rl.DrawTextureRec(
		tex,
		rl.Rectangle{
			X:      0,
			Y:      0,
			Width:  float32(tex.Width),
			Height: float32(tex.Height),
		},
		le.pos,
		rl.White,
	)
}

func (le *LevelEditor) drawGrid() {
	rl.PushMatrix()
	{
		a := le.gridSize.Y / le.gridSize.X
		rl.Translatef(
			float32(int32(le.pos.X)%int32(le.gridSize.X)+1),
			float32(int32(le.pos.Y)%int32(le.gridSize.Y)+(5*int32(le.gridSize.Y))),
			0)
		rl.Rotatef(90, 1, 0, 0)
		rl.Scalef(1, 0, a)
		rl.DrawGrid(100, le.gridSize.X)
	}
	rl.PopMatrix()
}

func (le *LevelEditor) Draw() {
	le.drawEntities()
	le.drawCurEntity()
	le.drawGrid()
}

func (le *LevelEditor) DrawGui() {
	col := rl.GetColor(uint(rlg.GetStyle(rlg.DEFAULT, rlg.BACKGROUND_COLOR)))
	rl.DrawRectangle(SW-200, 0, 200, SH, col)
	rlg.ListView(rl.Rectangle{SW - 200, 0, 200, 200}, "1;2;3", new(int32), 0)
}

func pos2block(pos rl.Vector2) rl.Vector2 {
	return rl.Vector2{} // TODO
}
