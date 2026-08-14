package resistor

import (
	"fmt"
	rlg "github.com/gen2brain/raylib-go/raygui"
	rl "github.com/gen2brain/raylib-go/raylib"
	"slices"
	"strconv"
	"strings"
)

const (
	SW = 1280
	SH = 720
)

var (
	Debug         = true
	GameTimeSec   float64
	PhysicsTime   float64
	Ticks         int
	FixedTimeStep float64 = 1.0 / 30.0
	MouseWorldPos Vector2i
	MousePos      rl.Vector2
	Cam           Camera
	LE            LevelEditor
	Pl            *Actor
	Actors        []*Actor
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

		for PhysicsTime < GameTimeSec {
			PhysicsTime += FixedTimeStep
			Ticks++
			Pl.UpdatePhysics()
			enemy.UpdatePhysics()
		}

		dt = float32((PhysicsTime - GameTimeSec) / FixedTimeStep)
		Pl.Update(dt)
		enemy.Update(dt)
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
	LE.SaveLevel("level_1.json")
}
