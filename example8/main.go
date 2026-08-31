package main

// Пост-процесс шейдер, ограничивающий цвет кадра заданной палитрой
// (ретро/pixel-art эффект): для каждого пикселя ищем ближайший по
// RGB-расстоянию цвет из палитры и заменяем им.
//
// Палитра передаётся как обычная Nx1-текстура с point-фильтрацией -
// так её удобно менять на лету (загрузить свою .png палитру, или
// сгенерировать программно), не пересобирая шейдер.
//
// go get github.com/gen2brain/raylib-go/raylib
// go run palette_shader_raylib.go

import (
	"fmt"
	rl "github.com/gen2brain/raylib-go/raylib"
)

const screenW = 960
const screenH = 640

const fsPalette = `
#version 330
in vec2 fragTexCoord;
out vec4 finalColor;

uniform sampler2D texture0;   // сцена
uniform sampler2D texture1;   // палитра, Nx1 текстура
uniform float paletteSize;    // сколько цветов реально используется в палитре
uniform float ditherStrength; // 0 = выключено, >0 - сила дизеринга (0..1)
uniform vec2  resolution;     // размер экрана в пикселях
uniform float pixelSize;      // размер "большого пикселя" в экранных пикселях (>=1)

// Простой 4x4 Bayer-дизер. Принимает координату УЖЕ В БЛОКАХ (не в
// реальных экранных пикселях) - иначе при крупных pixelSize узор
// дизера останется мелким и "поедет" относительно крупных пикселей.
float bayer4x4(vec2 p)
{
    int x = int(mod(p.x, 4.0));
    int y = int(mod(p.y, 4.0));
    int idx = y * 4 + x;
    float m[16] = float[16](
         0.0,  8.0,  2.0, 10.0,
        12.0,  4.0, 14.0,  6.0,
         3.0, 11.0,  1.0,  9.0,
        15.0,  7.0, 13.0,  5.0
    );
    return m[idx] / 16.0 - 0.5;
}

void main()
{
    float ps = max(pixelSize, 1.0);

    // координата текущего "большого пикселя" в блоках сетки
    vec2 blockIdx = floor(gl_FragCoord.xy / ps);

    // центр этого блока в экранных пикселях -> в UV сцены
    vec2 blockCenterPx = (blockIdx + 0.5) * ps;
    vec2 uv = blockCenterPx / resolution;

    vec3 col = texture(texture0, uv).rgb;

    // дизер по индексу блока, не по реальному пикселю экрана
    float dith = bayer4x4(blockIdx) * ditherStrength * 0.06;
    col += vec3(dith);

    float bestDist = 1.0e9;
    vec3 bestColor = col;

    for (int i = 0; i < 256; i++)
    {
        if (float(i) >= paletteSize) break;
        vec2 puv = vec2((float(i) + 0.5) / paletteSize, 0.5);
        vec3 pal = texture(texture1, puv).rgb;
        float d = distance(col, pal);
        if (d < bestDist)
        {
            bestDist = d;
            bestColor = pal;
        }
    }

    finalColor = vec4(bestColor, 1.0);
}`

// makePaletteTexture строит Nx1 текстуру из списка цветов палитры.
func makePaletteTexture(colors []rl.Color) rl.Texture2D {
	img := rl.GenImageColor(len(colors), 1, rl.Blank)
	for i, c := range colors {
		rl.ImageDrawPixel(img, int32(i), 0, c)
	}
	tex := rl.LoadTextureFromImage(img)
	rl.UnloadImage(img)
	rl.SetTextureFilter(tex, rl.FilterPoint)
	rl.SetTextureWrap(tex, rl.WrapClamp)
	return tex
}

func main() {
	rl.InitWindow(screenW, screenH, "Palette-limit shader — raylib-go")
	rl.SetTargetFPS(60)

	shader := rl.LoadShaderFromMemory("", fsPalette)
	locTexture1 := rl.GetShaderLocation(shader, "texture1")
	locPaletteSize := rl.GetShaderLocation(shader, "paletteSize")
	locDither := rl.GetShaderLocation(shader, "ditherStrength")
	locResolution := rl.GetShaderLocation(shader, "resolution")
	locPixelSize := rl.GetShaderLocation(shader, "pixelSize")

	// Пример палитры (можно заменить своей, хоть из 4, хоть из 64 цветов) -
	// адаптация классической PICO-8-подобной палитры из 16 цветов.
	palette := []rl.Color{
		rl.NewColor(0, 0, 0, 255),
		rl.NewColor(29, 43, 83, 255),
		rl.NewColor(126, 37, 83, 255),
		rl.NewColor(0, 135, 81, 255),
		rl.NewColor(171, 82, 54, 255),
		rl.NewColor(95, 87, 79, 255),
		rl.NewColor(194, 195, 199, 255),
		rl.NewColor(255, 241, 232, 255),
		rl.NewColor(255, 0, 77, 255),
		rl.NewColor(255, 163, 0, 255),
		rl.NewColor(255, 236, 39, 255),
		rl.NewColor(0, 228, 54, 255),
		rl.NewColor(41, 173, 255, 255),
		rl.NewColor(131, 118, 156, 255),
		rl.NewColor(255, 119, 168, 255),
		rl.NewColor(255, 204, 170, 255),
	}
	paletteTex := makePaletteTexture(palette)

	scene := rl.LoadRenderTexture(screenW, screenH)
	ditherOn := true

	zoom := float32(1.0)
	const basePixelSize = 1.0 // размер "большого пикселя" при zoom=1 (в экранных px)

	for !rl.WindowShouldClose() {
		if rl.IsKeyPressed(rl.KeyD) {
			ditherOn = !ditherOn
		}

		wheel := rl.GetMouseWheelMove()
		zoom += wheel * 0.3
		if zoom < 1.0 {
			zoom = 1.0
		}
		if zoom > 12.0 {
			zoom = 12.0
		}

		camera := rl.Camera2D{
			Target: rl.Vector2{X: screenW / 2, Y: screenH / 2},
			Offset: rl.Vector2{X: screenW / 2, Y: screenH / 2},
			Zoom:   zoom,
		}

		// 1) рисуем обычную сцену через камеру (масштаб влияет на то, что видно,
		//    а НЕ на pixelSize напрямую - но мы отдельно тоже привяжем pixelSize к zoom)
		rl.BeginTextureMode(scene)
		rl.ClearBackground(rl.RayWhite)
		rl.BeginMode2D(camera)
		for x := int32(0); x < screenW; x += 4 {
			t := float32(x) / screenW
			c := rl.ColorLerp(rl.NewColor(20, 20, 60, 255), rl.NewColor(255, 200, 80, 255), t)
			rl.DrawRectangle(x, 0, 4, screenH/2, c)
		}
		rl.DrawCircleGradient(rl.Vector2{300, 420}, 140, rl.NewColor(255, 60, 120, 255), rl.Blank)
		rl.DrawCircleGradient(rl.Vector2{650, 460}, 160, rl.NewColor(40, 200, 255, 255), rl.Blank)
		rl.DrawText("PALETTE DEMO", 60, 40, 60, rl.White)
		rl.EndMode2D()
		rl.EndTextureMode()

		// 2) применяем палитровый шейдер как пост-процесс;
		//    pixelSize растёт вместе с zoom -> крупные пиксели при большом zoom
		pixelSize := basePixelSize * zoom

		rl.BeginDrawing()
		rl.ClearBackground(rl.Black)

		rl.BeginShaderMode(shader)
		rl.SetShaderValueTexture(shader, locTexture1, paletteTex)
		rl.SetShaderValue(shader, locPaletteSize, []float32{float32(len(palette))}, rl.ShaderUniformFloat)
		rl.SetShaderValue(shader, locResolution, []float32{screenW, screenH}, rl.ShaderUniformVec2)
		rl.SetShaderValue(shader, locPixelSize, []float32{pixelSize}, rl.ShaderUniformFloat)
		dv := float32(0)
		if ditherOn {
			dv = 1
		}
		rl.SetShaderValue(shader, locDither, []float32{dv}, rl.ShaderUniformFloat)

		src := rl.NewRectangle(0, 0, screenW, -screenH) // флип render-текстуры
		dst := rl.NewRectangle(0, 0, screenW, screenH)
		rl.DrawTexturePro(scene.Texture, src, dst, rl.Vector2{}, 0, rl.White)
		rl.EndShaderMode()

		rl.DrawText("D - дизеринг | колесо мыши - zoom камеры (влияет на pixelSize)", 10, screenH-30, 18, rl.White)
		rl.DrawText(fmt.Sprintf("zoom: %.1f  pixelSize: %.1f", zoom, pixelSize), 10, screenH-55, 18, rl.White)
		rl.DrawFPS(10, 10)
		rl.EndDrawing()
	}

	rl.UnloadTexture(paletteTex)
	rl.UnloadRenderTexture(scene)
	rl.UnloadShader(shader)
	rl.CloseWindow()
}
