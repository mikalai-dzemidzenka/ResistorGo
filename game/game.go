package game

import (
	"cmp"
	"encoding/json"
	"flag"
	rlg "github.com/gen2brain/raylib-go/raygui"
	rl "github.com/gen2brain/raylib-go/raylib"
	"os"
	"slices"
	"strconv"
)

const (
	SW = 1280
	SH = 720
)

const (
	LEVEL_1 = iota
	LEVEL_EDITOR
	LEVEL_EDITOR_PLAY
	LEVELS
)

const (
	ID_CURSOR = iota
	ID_PLAYER
	ID_LEVEL_EDITOR
	ID_SOLID_BASE = 100
)

var g struct {
	Level       int
	Player      *Player
	LevelEditor *LevelEditor
	Cursor      *Cursor
	Camera      *rl.Camera2D
	Entities    RuntimeEntities
	Solids      []*Solid
	Drawables   []Drawable
	Updaters    []Updatable
	DrawGUIs    []GUIDrawable
}

func LoadLevel(level int) {
	prevLevel := g.Level

	g.Level = level
	g.Entities = make(RuntimeEntities)
	if g.Solids != nil {
		g.Solids = g.Solids[:0]
	}
	g.Cursor = nil

	if prevLevel != level && prevLevel == LEVEL_EDITOR {
		g.LevelEditor.SaveLevel("levels/level.json")
	}

	switch g.Level {
	case LEVEL_EDITOR:
		g.LevelEditor = NewLevelEditor()
		g.LevelEditor.LoadLevel("levels/level.json")
		g.Camera = &g.LevelEditor.Camera
		g.Entities[ID_LEVEL_EDITOR] = g.LevelEditor
	case LEVEL_EDITOR_PLAY:
		path := "levels/level.json"
		// func
		b, err := os.ReadFile(path)
		if err != nil {
			panic(err)
		}
		var layers []*LevelLayer
		err = json.Unmarshal(b, &layers)
		if err != nil {
			panic(err)
		}

		g.Cursor = &Cursor{}
		g.Entities[ID_CURSOR] = g.Cursor

		g.Player = NewPlayer()
		g.Camera = &g.Player.Camera
		g.Entities[ID_PLAYER] = g.Player

		solidID := ID_SOLID_BASE
		for _, layer := range layers {
			for _, e := range layer.Entities {
				switch e.Type {
				case ENTITY_SOLID:
					g.Solids = append(g.Solids, &Solid{
						Z:   layer.Z,
						Tex: e.Tex,
						Pos: e.Pos,
						Col: rl.Rectangle{},
					})
					solidID += 1
				}
			}
		}
	default:
		g.Cursor = &Cursor{}
		g.Entities[ID_CURSOR] = g.Cursor

		g.Player = NewPlayer()
		g.Camera = &g.Player.Camera
		g.Entities[ID_PLAYER] = g.Player

		{
			g.Solids = append(g.Solids, NewSolid(rl.Vector2{0, -30}))
		}

	}

	EvalEntities()
}

func Update(dt float32) {
	if rl.IsKeyPressed(rl.KeyOne) {
		LoadLevel(LEVEL_EDITOR)
	}
	if rl.IsKeyPressed(rl.KeyTwo) {
		LoadLevel(LEVEL_EDITOR_PLAY)
	}
	for i := range g.Updaters {
		g.Updaters[i].Update(dt)
	}
}

func Draw() {
	for i := range g.Drawables {
		g.Drawables[i].Draw()
	}
}

func DrawGUI() {
	for i := range g.DrawGUIs {
		g.DrawGUIs[i].DrawGUI()
	}
}

func EvalEntities() {
	if g.Updaters != nil {
		g.Updaters = g.Updaters[:0]
	}
	if g.Drawables != nil {
		g.Drawables = g.Drawables[:0]
	}
	if g.DrawGUIs != nil {
		g.DrawGUIs = g.DrawGUIs[:0]
	}

	for i := range g.Entities {
		{
			u, ok := g.Entities[i].(Updatable)
			if ok {
				g.Updaters = append(g.Updaters, u)
			}
		}
		{
			d, ok := g.Entities[i].(Drawable)
			if ok {
				g.Drawables = append(g.Drawables, d)
			}
		}
		{
			dg, ok := g.Entities[i].(GUIDrawable)
			if ok {
				g.DrawGUIs = append(g.DrawGUIs, dg)
			}
		}
	}
	for i := range g.Solids {
		g.Drawables = append(g.Drawables, g.Solids[i])
	}

	slices.SortFunc(g.Drawables, func(a, b Drawable) int {
		return cmp.Compare(a.GetZ(), b.GetZ())
	})

	slices.SortFunc(g.DrawGUIs, func(a, b GUIDrawable) int {
		return cmp.Compare(a.GetGUIZ(), b.GetGUIZ())
	})
}

func Run() {
	rl.InitWindow(SW, SH, "Resistor")
	defer rl.CloseWindow()

	flag.IntVar(&g.Level, "l", LEVEL_EDITOR_PLAY, "level to load")

	rl.DisableCursor()

	LoadTextures()
	rlg.LoadStyle("resources/style_dark.rgs")

	rl.SetTargetFPS(60)

	LoadLevel(g.Level)

	for !rl.WindowShouldClose() {
		if rl.IsKeyPressed(rl.KeyGrave) {
			Debug = !Debug
		}

		dt := rl.GetFrameTime()

		Update(dt)

		rl.BeginDrawing()
		{
			rl.ClearBackground(rl.Black)
			rl.BeginMode2D(*g.Camera)
			{
				Draw()
			}
			rl.EndMode2D()

			DrawGUI()

			if Debug {
				fps := strconv.Itoa(int(rl.GetFPS()))
				rl.DrawText("fps:"+fps, 20, 20, 2, rl.White)
				rl.DrawText(g.Player.Debug(), 40, 40, 1, rl.Red)
				rl.DrawText(g.Cursor.Debug(), 40, 116, 1, rl.Blue)
			}
		}
		rl.EndDrawing()
	}
	if g.LevelEditor != nil {
		g.LevelEditor.SaveLevel("levels/level.json")
	}
}
