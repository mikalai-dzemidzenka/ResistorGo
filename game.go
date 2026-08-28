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
	SW          = 160
	SH          = 90
	WindowScale = 8
)

const (
	StageMainMenu = iota
	StageLevel
)

var (
	Debug         = true
	GameTimeSec   float64
	Accumulator   float64
	PhysicsTime   float64
	Ticks         int
	FixedTimeStep float64 = 1.0 / 60.0
	MouseWorldPos Vector2i
	MousePos      rl.Vector2
	Cam           Camera
	LE            LevelEditor
	Pl            *Actor
	Actors        []*Actor
	Stage         int = StageLevel
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
	rl.InitWindow(SW*WindowScale, SH*WindowScale, "Resistor")
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

	ScreenTexture = rl.LoadRenderTexture(SW, SH)
	rl.SetTextureFilter(ScreenTexture.Texture, rl.FilterPoint)
	Cam.Camera2D.Offset = rl.Vector2{SW * WindowScale * 0.9 / 2, SH * WindowScale / 2}
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

func runLevel(dt float32) {
	Accumulator += float64(dt)

	for Accumulator > FixedTimeStep {
		PhysicsTime += FixedTimeStep
		for _, actor := range Actors {
			actor.Update(float32(FixedTimeStep))
		}
		Accumulator -= FixedTimeStep
	}

	alpha := float32(Accumulator / FixedTimeStep)

	for _, actor := range Actors {
		actor.Interp(alpha)
	}

	if Debug {
		LE.Update(dt)
	}
	UpdateRewind(dt)
	Cam.SetTarget(Pl.Pos)
	Cam.Update(dt)
	rl.BeginTextureMode(ScreenTexture)
	rl.ClearBackground(rl.Color{255, 255, 255, 0})
	oldOffset := Cam.Camera2D.Offset
	Cam.Camera2D.Offset = rl.Vector2{SW * 0.9 / 2, SH / 2}
	Cam.Camera2D.Zoom /= WindowScale
	pl := rl.GetWorldToScreen2D(Pl.Pos, Cam.Camera2D)
	type world struct {
		Coord rl.Vector2
		Col   rl.Vector2
	}
	objCoords := make([]world, 0, len(LE.objectsIndexed))
	for _, obj := range LE.objectsIndexed {
		worldCoord := rl.GetWorldToScreen2D(rl.Vector2{
			X: float32(obj.X) + obj.CollisionRect.X,
			Y: float32(obj.Y) + obj.CollisionRect.Y,
		}, Cam.Camera2D)
		objCoords = append(objCoords, world{worldCoord, rl.Vector2{obj.CollisionRect.Width, obj.CollisionRect.Height}})
	}
	for y := -SH; y < SH; y++ {
		for x := -SW; x < SW; x++ {
			coord := rl.GetWorldToScreen2D(rl.Vector2{X: float32(x) + Pl.Pos.X, Y: float32(y) + Pl.Pos.Y}, Cam.Camera2D)
			//coord := rl.GetScreenToWorld2D(rl.Vector2{float32(x), float32(y)}, Cam.Camera2D)
			for _, worldCoord := range objCoords {
				if segmentIntersectsRect(coord, pl, rl.Rectangle{
					X:      worldCoord.Coord.X,
					Y:      worldCoord.Coord.Y,
					Width:  worldCoord.Col.X * Cam.Zoom,
					Height: worldCoord.Col.Y * Cam.Zoom,
				}) {
					rl.DrawPixelV(coord, rl.Color{0, 0, 0, 255})
				}
			}
		}
	}
	//for _, obj := range LE.objectsIndexed {
	//	worldCoord := rl.GetWorldToScreen2D(rl.Vector2{
	//		X: float32(obj.X) + obj.CollisionRect.X,
	//		Y: float32(obj.Y) + obj.CollisionRect.Y,
	//	}, Cam.Camera2D)
	//
	//	rl.DrawRectangleRec(rl.Rectangle{
	//		X:      worldCoord.X,
	//		Y:      worldCoord.Y,
	//		Width:  obj.CollisionRect.Width * Cam.Zoom,
	//		Height: obj.CollisionRect.Height * Cam.Zoom,
	//	}, rl.Green)
	//}
	Cam.Camera2D.Zoom *= WindowScale
	Cam.Camera2D.Offset = oldOffset
	rl.EndTextureMode()

	rl.BeginDrawing()
	{
		rl.ClearBackground(rl.Black)
		rl.BeginMode2D(Cam.Camera2D)
		{
			LE.Draw() // TODO sort objects with actors

			slices.SortFunc(Actors, func(a, b *Actor) int {
				return int(b.Pos.X - a.Pos.X)
			})
			for _, a := range Actors {
				a.Draw()
			}
		}
		rl.EndMode2D()

		rl.DrawTexturePro(ScreenTexture.Texture,
			rl.Rectangle{
				X:      0,
				Y:      0,
				Width:  float32(ScreenTexture.Texture.Width),
				Height: -float32(ScreenTexture.Texture.Height),
			},
			rl.Rectangle{
				X:      0,
				Y:      0,
				Width:  SW * WindowScale,
				Height: SH * WindowScale,
			},
			rl.Vector2{}, 0, rl.White)

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

func segmentIntersectsRect(
	start rl.Vector2,
	end rl.Vector2,
	rect rl.Rectangle,
) bool {

	dx := end.X - start.X
	dy := end.Y - start.Y

	tMin := float32(0.0)
	tMax := float32(1.0)

	// --------------------------------------------------------
	// X
	// --------------------------------------------------------

	if math.Abs(float64(dx)) < 0.000001 {

		if start.X < rect.X ||
			start.X > rect.X+rect.Width {
			return false
		}

	} else {

		tx1 := (rect.X - start.X) / dx
		tx2 := (rect.X + rect.Width - start.X) / dx

		if tx1 > tx2 {
			tx1, tx2 = tx2, tx1
		}

		if tx1 > tMin {
			tMin = tx1
		}

		if tx2 < tMax {
			tMax = tx2
		}

		if tMin > tMax {
			return false
		}
	}

	// --------------------------------------------------------
	// Y
	// --------------------------------------------------------

	if math.Abs(float64(dy)) < 0.000001 {

		if start.Y < rect.Y ||
			start.Y > rect.Y+rect.Height {
			return false
		}

	} else {

		ty1 := (rect.Y - start.Y) / dy
		ty2 := (rect.Y + rect.Height - start.Y) / dy

		if ty1 > ty2 {
			ty1, ty2 = ty2, ty1
		}

		if ty1 > tMin {
			tMin = ty1
		}

		if ty2 < tMax {
			tMax = ty2
		}

		if tMin > tMax {
			return false
		}
	}

	return true
}
