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
	PauseTime            float64
	Ticks                int
	FixedTimeStep        float64 = 1.0 / 60.0
	MouseWorldPos        Vector2i
	MousePos             rl.Vector2
	Cam                  Camera
	LE                   LevelEditor
	Pl                   *Actor
	Actors               []*Actor
	Stage                int = 1
	MainMenuTexture      rl.RenderTexture2D
	ScreenFullTexture    rl.RenderTexture2D
	ScreenPaletteTexture rl.Texture2D
	PaletteShader        rl.Shader
	Sh                   *Shaders
	light1               *ZimaLight
	light2               *ZimaLight
	light3               *ZimaLight
	Palette              []rl.Color
	Cur                  *Cursor
	WhiteShader          rl.Shader
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

	WhiteShader = rl.LoadShaderFromMemory("", fsWhite)

	rl.DisableCursor()
	Cam = NewCamera(3)
	Cur = &Cursor{
		AtkSpeed: CursorAttackSpeed,
		AtkPos:   rl.Vector2{SW * WindowScale / 2, SH * WindowScale / 2},
		GUIZ:     2,
	}
	Pl = NewActor(rl.Vector2{}, true)
	enemy := NewActor(rl.Vector2{X: 50}, false)
	LE = NewLevelEditor(39, 83)
	LE.LoadLevel("level_1.json")

	Actors = []*Actor{enemy, Pl}

	MainMenuTexture = rl.LoadRenderTexture(SW, SH)
	rl.SetTextureFilter(MainMenuTexture.Texture, rl.FilterPoint)
	Cam.Camera2D.Offset = rl.Vector2{SW * WindowScale * 0.9 / 2, SH * WindowScale / 2}

	Sh = loadShaders()
	const intense = 1.0
	light1 = newZimaLight(512, 256, rl.White, intense)
	light2 = newZimaLight(512, 512, rl.Red, intense)
	light3 = newZimaLight(512, 512, rl.Green, intense)

	PaletteShader = rl.LoadShaderFromMemory("", fsPalette)

	Palette = []rl.Color{
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
		rl.NewColor(234, 240, 147, 255),
		rl.NewColor(242, 221, 161, 255),
		rl.NewColor(246, 245, 192, 255),
		rl.SkyBlue,
		rl.White,
		rl.Red,
		rl.Green,
		rl.NewColor(0, 200, 48, 255),
		rl.NewColor(10, 250, 50, 255),
		rl.NewColor(10, 255, 60, 255),
	}

	ScreenPaletteTexture = makePaletteTexture(Palette)
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
		if PauseTime > 0 {
			PauseTime -= FixedTimeStep
		} else {
			PhysicsTime += FixedTimeStep
			for _, actor := range Actors {
				actor.Update(float32(FixedTimeStep))
			}
		}

		Accumulator -= FixedTimeStep
	}

	alpha := float32(Accumulator / FixedTimeStep)

	Cur.Update(dt)
	for _, actor := range Actors {
		actor.Interp(alpha)
	}

	if Debug {
		LE.Update(dt)
	}
	UpdateRewind(dt)
	Cam.SetTarget(rl.Vector2{Pl.Pos.X + 24, Pl.Pos.Y + 24})
	Cam.Update(dt)

	pos1 := rl.Vector2{40, 30}
	pos2 := rl.Vector2{120, 120}
	pos3 := rl.Vector2{-40, 30}
	tex1 := light1.Compute(Sh, pos1, LE.LightOccluders())
	tex2 := light2.Compute(Sh, pos2, LE.LightOccluders())
	tex3 := light3.Compute(Sh, pos3, LE.LightOccluders())
	rl.BeginTextureMode(ScreenFullTexture)
	{
		rl.ClearBackground(rl.Black)
		rl.BeginMode2D(Cam.Camera2D)

		LE.Draw() // TODO sort objects with actors

		slices.SortFunc(Actors, func(a, b *Actor) int {
			return int(b.Pos.X - a.Pos.X)
		})
		locTime := rl.GetShaderLocation(WhiteShader, "time")
		for _, a := range Actors {
			if a.AnimWhite && a.AnimCurrentFrame < 2 {
				rl.BeginShaderMode(WhiteShader)
				rl.SetShaderValue(WhiteShader, locTime, []float32{float32(rl.GetTime())}, rl.ShaderUniformFloat)
				a.Draw()
				rl.EndShaderMode()
			} else {
				a.Draw()
			}
		}

		rl.EndMode2D()
		rl.DrawRectangle(0, 0, SW*WindowScale, SH*WindowScale, rl.NewColor(0, 0, 0, 50))
		rl.BeginMode2D(Cam.Camera2D)

		rl.BeginBlendMode(rl.BlendAdditive)
		addLight(
			tex1,
			pos1,
			light1.Radius,
		)

		addLight(
			tex2,
			pos2,
			light2.Radius,
		)

		addLight(
			tex3,
			pos3,
			light3.Radius,
		)
		rl.EndBlendMode()

		rl.EndMode2D()
	}
	rl.EndTextureMode()

	rl.BeginDrawing()
	rl.ClearBackground(rl.Black)

	locTexture1 := rl.GetShaderLocation(PaletteShader, "texture1")
	locPaletteSize := rl.GetShaderLocation(PaletteShader, "paletteSize")
	locDither := rl.GetShaderLocation(PaletteShader, "ditherStrength")
	locResolution := rl.GetShaderLocation(PaletteShader, "resolution")
	locPixelSize := rl.GetShaderLocation(PaletteShader, "pixelSize")
	rl.BeginShaderMode(PaletteShader)
	rl.SetShaderValueTexture(PaletteShader, locTexture1, ScreenPaletteTexture)
	rl.SetShaderValue(PaletteShader, locResolution, []float32{SW * WindowScale, SH * WindowScale}, rl.ShaderUniformVec2)
	rl.SetShaderValue(PaletteShader, locPixelSize, []float32{Cam.Zoom}, rl.ShaderUniformFloat)
	rl.SetShaderValue(PaletteShader, locPaletteSize, []float32{float32(len(Palette))}, rl.ShaderUniformFloat)
	rl.SetShaderValue(PaletteShader, locDither, []float32{1}, rl.ShaderUniformFloat)

	src := rl.NewRectangle(0, 0, SW*WindowScale, -SH*WindowScale) // флип render-текстуры
	dst := rl.NewRectangle(0, 0, SW*WindowScale, SH*WindowScale)
	rl.DrawTexturePro(ScreenFullTexture.Texture, src, dst, rl.Vector2{}, 0, rl.White)
	rl.EndShaderMode()
	if Debug {
		LE.DrawGui()
	}
	Cur.DrawGUI()

	if Debug {
		fps := strconv.Itoa(int(rl.GetFPS()))
		rl.DrawText("fps:"+fps, 20, 20, 2, rl.White)
		rl.DrawText(fmt.Sprintf("gameTime:%f", GameTimeSec), 20, 30, 2, rl.White)
		rl.DrawText(fmt.Sprintf("physTime:%f", PhysicsTime), 20, 40, 2, rl.White)
		rl.DrawText(fmt.Sprintf("camera zoom speed %f", Cam.ZoomSpeed), 20, 50, 2, rl.White)
	}
	rl.EndDrawing()
}
