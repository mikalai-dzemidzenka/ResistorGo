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
)

var g struct {
	Level       int
	Player      *Player
	LevelEditor *LevelEditor
	Cursor      *Cursor
	Camera      *rl.Camera2D
	Entities    RuntimeEntities
	Drawables   []Drawable
	Updaters    []Updatable
	DrawGUIs    []GUIDrawable
}

func InitGameState() {
	g.Entities = make(RuntimeEntities)

	switch g.Level {
	case LEVEL_EDITOR:
		g.LevelEditor = &LevelEditor{
			LevelLayers: []*LevelLayer{
				{
					Entities:            []LevelEntity{},
					EntityPositionIndex: make(map[LevelBlock]uint16),
				},
			},
			Camera: rl.NewCamera2D(rl.Vector2{}, rl.Vector2{}, 0, 1),
			CurEntity: LevelEntity{
				Type: ENTITY_SOLID,
				Tex:  TEXTURE_BLOCK_1,
			},
			GridSize: rl.Vector2{X: 39, Y: 83},
			Z:        -1,
			GUIZ:     1,
		}
		g.LevelEditor.LoadLevel("levels/level.json")
		g.Camera = &g.LevelEditor.Camera
		g.Entities[ID_LEVEL_EDITOR] = g.LevelEditor
	default:
		g.Cursor = &Cursor{
			Pos:    rl.Vector2{},
			atkPos: rl.Vector2{},
			GUIZ:   0,
		}
		g.Entities[ID_CURSOR] = g.Cursor

		g.Player = &Player{
			Camera:   rl.NewCamera2D(rl.Vector2{X: SW / 2, Y: SH / 2}, rl.Vector2{}, 0, 1),
			Bound:    rl.Rectangle{Width: 39, Height: 44},
			Pos:      rl.Vector2{0, 0},
			Speed:    rl.Vector2{},
			Friction: DefaultFriction,
			Anim:     TEXTURE_PLAYER_IDLE,
			Orient:   1,
			Z:        0,
		}

		g.Camera = &g.Player.Camera
		g.Entities[ID_PLAYER] = g.Player
	}

	EvalEntities()
}

func Update(dt float32) {
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

	InitGameState()

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
		g.LevelEditor.SaveEntities("levels/level.json")
	}
}
