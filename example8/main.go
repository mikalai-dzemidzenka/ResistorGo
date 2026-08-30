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

// Простой 4x4 Bayer-дизер, чтобы переходы между цветами палитры
// не давали грубые полосы (banding), а выглядели немного "зашумлённо",
// как на старых консолях.
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
    return m[idx] / 16.0 - 0.5; // диапазон примерно [-0.5, 0.5]
}

void main()
{
    vec3 col = texture(texture0, fragTexCoord).rgb;

    // лёгкий дизер перед квантованием сглаживает видимые "ступеньки"
    float dith = bayer4x4(gl_FragCoord.xy) * ditherStrength * 0.06;
    col += vec3(dith);

    float bestDist = 1.0e9;
    vec3 bestColor = col;

    for (int i = 0; i < 256; i++)
    {
        if (float(i) >= paletteSize) break;
        vec2 uv = vec2((float(i) + 0.5) / paletteSize, 0.5);
        vec3 pal = texture(texture1, uv).rgb;
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

	for !rl.WindowShouldClose() {
		if rl.IsKeyPressed(rl.KeyD) {
			ditherOn = !ditherOn
		}

		// 1) рисуем обычную сцену (тут - просто демонстрационный градиент + фигуры)
		rl.BeginTextureMode(scene)
		rl.ClearBackground(rl.RayWhite)
		for x := int32(0); x < screenW; x += 4 {
			t := float32(x) / screenW
			c := rl.ColorLerp(rl.NewColor(20, 20, 60, 255), rl.NewColor(255, 200, 80, 255), t)
			rl.DrawRectangle(x, 0, 4, screenH/2, c)
		}
		rl.DrawCircleGradient(rl.Vector2{300, 420}, 140, rl.NewColor(255, 60, 120, 255), rl.Blank)
		rl.DrawCircleGradient(rl.Vector2{650, 460}, 160, rl.NewColor(40, 200, 255, 255), rl.Blank)
		rl.DrawText("PALETTE DEMO", 60, 40, 60, rl.White)
		rl.EndTextureMode()

		// 2) применяем палитровый шейдер как пост-процесс
		rl.BeginDrawing()
		rl.ClearBackground(rl.Black)

		rl.BeginShaderMode(shader)
		rl.SetShaderValueTexture(shader, locTexture1, paletteTex)
		rl.SetShaderValue(shader, locPaletteSize, []float32{float32(len(palette))}, rl.ShaderUniformFloat)
		dv := float32(0)
		if ditherOn {
			dv = 1
		}
		rl.SetShaderValue(shader, locDither, []float32{dv}, rl.ShaderUniformFloat)

		src := rl.NewRectangle(0, 0, screenW, -screenH) // флип render-текстуры
		dst := rl.NewRectangle(0, 0, screenW, screenH)
		rl.DrawTexturePro(scene.Texture, src, dst, rl.Vector2{}, 0, rl.White)
		rl.EndShaderMode()

		rl.DrawText("D - переключить дизеринг", 10, screenH-30, 18, rl.White)
		rl.DrawFPS(10, 10)
		rl.EndDrawing()
	}

	rl.UnloadTexture(paletteTex)
	rl.UnloadRenderTexture(scene)
	rl.UnloadShader(shader)
	rl.CloseWindow()
}
