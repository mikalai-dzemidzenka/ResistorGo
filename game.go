package resistor

import (
	"fmt"
	rlg "github.com/gen2brain/raylib-go/raygui"
	rl "github.com/gen2brain/raylib-go/raylib"
	"math"
	"slices"
	"strconv"
	"strings"
)

const (
	SW = 1280
	SH = 720
)

const (
	StageMainMenu = iota
	StageLevel
)

var (
	Debug           = true
	GameTimeSec     float64
	PhysicsTime     float64
	Ticks           int
	FixedTimeStep   float64 = 1.0 / 30.0
	MouseWorldPos   Vector2i
	MousePos        rl.Vector2
	Cam             Camera
	LE              LevelEditor
	Pl              *Actor
	Actors          []*Actor
	Stage           int
	MainMenuTexture rl.RenderTexture2D
)

type Vector2i struct {
	X int32
	Y int32
}

func (v Vector2i) String() string {
	return fmt.Sprintf("%d,%d", v.X, v.Y)
}

func (v Vector2i) MarshalJSON() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *Vector2i) UnmarshalJSON(b []byte) error {
	tokens := strings.Split(string(b), ",")
	if len(tokens) != 2 {
		return fmt.Errorf("malformed input: %s", b)
	}
	x, _ := strconv.Atoi(tokens[0])
	y, _ := strconv.Atoi(tokens[0])
	v.X = int32(x)
	v.Y = int32(y)
	return nil
}

func v2i(v rl.Vector2) Vector2i {
	return Vector2i{
		X: int32(v.X),
		Y: int32(v.Y),
	}
}

func v2f(v Vector2i) rl.Vector2 {
	return rl.Vector2{
		X: float32(v.X),
		Y: float32(v.Y),
	}
}

func Game() {
	rl.InitWindow(SW, SH, "Resistor")
	defer rl.CloseWindow()

	//rl.DisableCursor()

	rlg.LoadStyle("resources/style_dark.rgs")

	rl.SetTargetFPS(60)

	LoadTextures()

	Cam = NewCamera(3)
	Pl = NewActor(rl.Vector2{}, true)
	enemy := NewActor(rl.Vector2{X: 50}, false)
	LE = NewLevelEditor(39, 83)
	LE.LoadLevel("level_1.json")

	Actors = []*Actor{Pl, enemy}

	MainMenuTexture = rl.LoadRenderTexture(SW, SH)
	for !rl.WindowShouldClose() {
		if rl.IsKeyPressed(rl.KeyGrave) {
			Debug = !Debug
		}
		dt := rl.GetFrameTime()
		if dt > 0.25 {
			dt = 0.25
		}

		GameTimeSec += float64(dt)
		MousePos = rl.GetMousePosition()
		MouseWorldPos = v2i(rl.GetScreenToWorld2D(MousePos, Cam.Camera2D))
		switch Stage {
		case StageMainMenu:
			runMainMenu(dt)
		case StageLevel:
			runLevel(dt)
		}
	}
	LE.SaveLevel("level_1.json")
}

func runMainMenu(dt float32) {
	Cam.Camera2D.Zoom = 4
	rl.BeginDrawing()
	{
		rl.BeginTextureMode(MainMenuTexture)
		rl.ClearBackground(rl.Color{0, 0, 0, 0})
		rl.BeginMode3D(rl.NewCamera3D(rl.Vector3{0, -1.5, 10}, rl.Vector3{0, 0, 0}, rl.Vector3{0, 1, 0}, 120, rl.CameraPerspective))
		var linesColor = rl.Red
		linesColor.A = 200
		for i := float32(-20); i <= 20; i++ {
			rl.DrawLine3D(rl.Vector3{X: i}, rl.Vector3{X: i * 1.5, Z: 10}, linesColor)
		}
		for i := float32(-20); i <= 20; i++ {
			rl.DrawLine3D(rl.Vector3{Y: float32(math.Sin(GameTimeSec)+1.2) * i}, rl.Vector3{X: -15, Y: i * 3}, linesColor)
			rl.DrawLine3D(rl.Vector3{Y: float32(math.Sin(GameTimeSec)+1.2) * i}, rl.Vector3{X: 15, Y: i * 3}, linesColor)
		}
		rl.EndMode3D()

		rl.BeginMode3D(rl.NewCamera3D(rl.Vector3{0, 5, 10}, rl.Vector3{0, 0, 0}, rl.Vector3{0, 1, 0}, 10, rl.CameraOrthographic))
		head()
		jab1()
		jab2()
		rl.EndMode3D()
		rl.EndTextureMode()
		rl.ClearBackground(rl.Black)
		rl.BeginMode2D(Cam.Camera2D)
		{
			rl.DrawTextureEx(MainMenuTexture.Texture, rl.Vector2{
				X: Pl.Pos.X - SW/2,
				Y: Pl.Pos.Y - SH/2,
			}, 0, 1, rl.Red)
			tex := resources["resources/main_menu/resistor.png"]
			col := rl.White
			if GameTimeSec > 3 {
				col.A = 255
			} else if GameTimeSec > 2 {
				col.A = uint8(math.Min((math.Sin(GameTimeSec-2))*255, 255))
			} else {
				col.A = 0
			}
			rl.DrawTexture(tex, -tex.Width/2, -20, col)
		}
		rl.EndMode2D()
	}
	rl.EndDrawing()
}

func head() {
	const headScale = 0.75
	color := rl.Red
	color.A = uint8((math.Sin(GameTimeSec) + 1) * 127)
	rl.DrawModelWiresEx(
		models["resources/main_menu/head.obj"],
		rl.Vector3{},
		rl.Vector3{0, 1, 0},
		float32(GameTimeSec*100),
		rl.Vector3{headScale, -headScale, headScale},
		color,
	)
}

func jab1() {
	const jabScale = 0.25
	colorJab := rl.Red
	if GameTimeSec > 1.5 {
		colorJab.A = uint8((math.Sin(GameTimeSec+200) + 1) * 127)
	} else {
		colorJab.A = 0
	}
	rl.DrawModelWiresEx(
		models["resources/main_menu/jab1.obj"],
		rl.Vector3{X: -1.5, Y: 1},
		rl.Vector3{0, 1, 0},
		float32(-GameTimeSec*100),
		rl.Vector3{jabScale, -jabScale, jabScale},
		colorJab,
	)
}

func jab2() {
	const jabScale = 0.25
	colorJab := rl.Red
	if GameTimeSec > 1.8 {
		colorJab.A = uint8((math.Sin(GameTimeSec+400) + 1) * 127)
	} else {
		colorJab.A = 0
	}
	rl.DrawModelWiresEx(
		models["resources/main_menu/jab2.obj"],
		rl.Vector3{X: 1.5, Y: 1},
		rl.Vector3{0, 1, 0},
		float32(-GameTimeSec*100),
		rl.Vector3{jabScale, -jabScale, jabScale},
		colorJab,
	)
}

func runLevel(dt float32) {
	for PhysicsTime < GameTimeSec {
		PhysicsTime += FixedTimeStep
		Ticks++
		for _, actor := range Actors {
			actor.UpdatePhysics()
		}
	}

	dt = float32((PhysicsTime - GameTimeSec) / FixedTimeStep)
	for _, actor := range Actors {
		actor.Update(dt)
	}
	if Debug {
		LE.Update(dt)
	}
	UpdateRewind(dt)

	rl.BeginDrawing()
	{
		rl.ClearBackground(rl.Black)
		rl.BeginMode2D(Cam.Camera2D)
		{
			LE.Draw()
			Cam.SetTarget(Pl.Pos)
			Cam.Update(dt)

			slices.SortFunc(Actors, func(a, b *Actor) int {
				return int(b.Pos.X - a.Pos.X)
			})
			for _, a := range Actors {
				a.Draw()
			}
		}
		rl.EndMode2D()

		if Debug {
			LE.DrawGui()
		}

		if Debug {
			fps := strconv.Itoa(int(rl.GetFPS()))
			rl.DrawText("fps:"+fps, 20, 20, 2, rl.White)
			rl.DrawText(fmt.Sprintf("gameTime:%f", GameTimeSec), 20, 30, 2, rl.White)
			rl.DrawText(fmt.Sprintf("physTime:%f", PhysicsTime), 20, 40, 2, rl.White)
			rl.DrawText(fmt.Sprintf("camera zoom speed %f", Cam.ZoomSpeed), 20, 50, 2, rl.White)
		}
	}
	rl.EndDrawing()
}
