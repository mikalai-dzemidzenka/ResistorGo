package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"slices"
	"strconv"
)

var le = &LevelEditor{
	entities: []LevelEntity{},
	texID:    TEXTURE_BLOCK_1,
	pos:      rl.Vector2{},
	gridSize: 39,
}

type LevelEntity struct {
	pos   rl.Vector2
	texID int32
}

type LevelEditor struct {
	entities []LevelEntity

	texID    int32
	pos      rl.Vector2
	gridSize int32
}

func (le *LevelEditor) Update(dt float32) {
	le.pos = rl.GetScreenToWorld2D(rl.GetMousePosition(), cam.cam)
	if le.gridSize > 0 {
		le.pos.X = float32(int32(le.pos.X) - int32(le.pos.X)%le.gridSize)
		le.pos.Y = float32(int32(le.pos.Y) - int32(le.pos.Y)%le.gridSize)
	}
	if rl.IsMouseButtonPressed(rl.MouseButtonLeft) {
		le.entities = append(le.entities, LevelEntity{
			pos:   le.pos,
			texID: le.texID,
		})
		slices.SortFunc(le.entities, func(a, b LevelEntity) int {
			switch {
			case a.pos.X > b.pos.X:
				return -1
			case a.pos.Y > b.pos.Y:
				return -1
			default:
				return 0
			}
		})
	}
}

func (le *LevelEditor) Draw() {
	for _, e := range le.entities {
		tex := Textures[e.texID]
		rl.DrawTextureRec(
			tex,
			rl.Rectangle{
				X:      0,
				Y:      0,
				Width:  float32(tex.Width),
				Height: float32(tex.Height),
			},
			e.pos,
			rl.White,
		)
	}
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

func lvlEditLoop(dt float32) {
	Debug = true

	cur.Update(dt)
	p.Update(dt)
	cam.Update(dt)
	le.Update(dt)

	rl.BeginDrawing()
	{
		rl.ClearBackground(rl.Black)
		rl.BeginMode2D(cam.cam)
		{

			le.Draw()
			p.Draw()
		}
		rl.EndMode2D()

		cur.DrawGui()
		if Debug {
			fps := strconv.Itoa(int(rl.GetFPS()))
			rl.DrawText("fps:"+fps, 20, 20, 2, rl.White)
			rl.DrawText(p.Debug(), 40, 40, 1, rl.Red)
			rl.DrawText(cam.Debug(), 40, 78, 1, rl.Green)
			rl.DrawText(cur.Debug(), 40, 116, 1, rl.Blue)
		}
	}
	rl.EndDrawing()
}

func pos2block(pos rl.Vector2) rl.Vector2 {
	return rl.Vector2{} // TODO
}
