package game

import (
	"cmp"
	"flag"
	rlg "github.com/gen2brain/raylib-go/raygui"
	rl "github.com/gen2brain/raylib-go/raylib"
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
	default:
		g.Cursor = &Cursor{}
		g.Entities[ID_CURSOR] = g.Cursor

		g.Player = &Player{
			Camera:   rl.NewCamera2D(rl.Vector2{X: SW / 2, Y: SH / 2}, rl.Vector2{}, 0, MinZoom),
			Bound:    rl.Rectangle{Width: 39, Height: 44},
			Pos:      rl.Vector2{0, 0},
			Speed:    rl.Vector2{},
			Acc:      DefaultAcc,
			Friction: DefaultFriction,
			Anim:     TEXTURE_PLAYER_IDLE,
			Orient:   1,
			Z:        0,
			Gravity:  0.2,
		}

		g.Camera = &g.Player.Camera
		g.Entities[ID_PLAYER] = g.Player

		{
			tex := Textures[TEXTURE_BLOCK_1]
			p := rl.Vector2{0, -30}
			var c int32 = 87
			g.Solids = append(g.Solids, &Solid{
				Z:   -1,
				Tex: TEXTURE_BLOCK_1,
				Pos: p,
				Col: rl.Rectangle{
					X:      p.X,
					Y:      p.Y + float32(c),
					Width:  float32(tex.Width),
					Height: float32(tex.Height - c),
				},
			})
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
