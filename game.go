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
	SW          = 160
	SH          = 90
	WindowScale = 8
)

const (
	StageMainMenu = iota
	StageLevel
)

var (
	Debug                = true
	GameTimeSec          float64
	Accumulator          float64
	PhysicsTime          float64
	Ticks                int
	FixedTimeStep        float64 = 1.0 / 60.0
	MouseWorldPos        Vector2i
	MousePos             rl.Vector2
	Cam                  Camera
	LE                   LevelEditor
	Pl                   *Actor
	Actors               []*Actor
	Stage                int = StageLevel
	ScreenTexture        rl.RenderTexture2D
	ScreenFullTexture    rl.RenderTexture2D
	ScreenPaletteTexture rl.Texture2D
	PaletteShader        rl.Shader
	Black                = rl.Color{39, 35, 36, 255}
	Sh                   *Shaders
	light1               *ZimaLight
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

	Sh = loadShaders()
	light1 = newZimaLight()

	PaletteShader = rl.LoadShaderFromMemory("", fsPalette)

	palette := []rl.Color{
		rl.NewColor(57, 31, 33, 255),
		rl.NewColor(39, 35, 36, 255),
		rl.NewColor(60, 51, 53, 255),
		rl.NewColor(87, 28, 39, 255),
		rl.NewColor(85, 70, 78, 255),
		rl.NewColor(91, 74, 100, 255),
		rl.NewColor(117, 102, 108, 255),
		rl.NewColor(137, 30, 43, 255),
		rl.NewColor(196, 36, 48, 255),
		rl.NewColor(234, 50, 60, 255),
		rl.NewColor(138, 72, 54, 255),
		rl.NewColor(191, 111, 74, 255),
		rl.NewColor(230, 156, 105, 255),
		rl.NewColor(156, 123, 132, 255),
		rl.NewColor(173, 153, 140, 255),
		rl.NewColor(186, 177, 155, 255),
		rl.NewColor(212, 207, 107, 255),
	}

	ScreenPaletteTexture = makePaletteTexture(palette)
	ScreenFullTexture = rl.LoadRenderTexture(SW*WindowScale, SH*WindowScale)

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

	tex1 := light1.Compute(Sh, Pl.Pos, rl.NewColor(255, 225, 170, 255), LE.LightOccluders())
	rl.BeginTextureMode(ScreenFullTexture)
	{
		rl.ClearBackground(Black)
		rl.BeginMode2D(Cam.Camera2D)

		LE.Draw() // TODO sort objects with actors

		slices.SortFunc(Actors, func(a, b *Actor) int {
			return int(b.Pos.X - a.Pos.X)
		})
		for _, a := range Actors {
			a.Draw()
		}

		rl.BeginBlendMode(rl.BlendAdditive)
		for _, lp := range []struct {
			pos rl.Vector2
			tex rl.Texture2D
		}{{Pl.Pos, tex1}} {
			src := rl.NewRectangle(0, 0, LA, -LA) // флип по Y для render-текстуры
			dst := rl.NewRectangle(lp.pos.X-LA/2, lp.pos.Y-LA/2, LA, LA)
			rl.DrawTexturePro(lp.tex, src, dst, rl.Vector2{}, 0, rl.White)
		}
		rl.EndBlendMode()

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
	rl.EndTextureMode()

	rl.BeginDrawing()
	rl.ClearBackground(rl.Black)

	locTexture1 := rl.GetShaderLocation(PaletteShader, "texture1")
	locPaletteSize := rl.GetShaderLocation(PaletteShader, "paletteSize")
	locDither := rl.GetShaderLocation(PaletteShader, "ditherStrength")
	rl.BeginShaderMode(PaletteShader)
	rl.SetShaderValueTexture(PaletteShader, locTexture1, ScreenPaletteTexture)
	palette := []rl.Color{
		rl.NewColor(57, 31, 33, 255),
		rl.NewColor(39, 35, 36, 255),
		rl.NewColor(60, 51, 53, 255),
		rl.NewColor(87, 28, 39, 255),
		rl.NewColor(85, 70, 78, 255),
		rl.NewColor(91, 74, 100, 255),
		rl.NewColor(117, 102, 108, 255),
		rl.NewColor(137, 30, 43, 255),
		rl.NewColor(196, 36, 48, 255),
		rl.NewColor(234, 50, 60, 255),
		rl.NewColor(138, 72, 54, 255),
		rl.NewColor(191, 111, 74, 255),
		rl.NewColor(230, 156, 105, 255),
		rl.NewColor(156, 123, 132, 255),
		rl.NewColor(173, 153, 140, 255),
		rl.NewColor(186, 177, 155, 255),
		rl.NewColor(212, 207, 107, 255),
	}
	rl.SetShaderValue(PaletteShader, locPaletteSize, []float32{float32(len(palette))}, rl.ShaderUniformFloat)
	rl.SetShaderValue(PaletteShader, locDither, []float32{1}, rl.ShaderUniformFloat)

	src := rl.NewRectangle(0, 0, SW*WindowScale, -SH*WindowScale) // флип render-текстуры
	dst := rl.NewRectangle(0, 0, SW*WindowScale, SH*WindowScale)
	rl.DrawTexturePro(ScreenFullTexture.Texture, src, dst, rl.Vector2{}, 0, rl.White)
	rl.EndShaderMode()
	rl.EndDrawing()
}
