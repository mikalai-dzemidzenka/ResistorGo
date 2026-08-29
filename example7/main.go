package main

// Точная (насколько это возможно на Go/raylib) реализация техники
// Catalin Zima "My technique for the shader-based dynamic 2D shadows":
//   http://www.catalinzima.com/2010/07/my-technique-for-the-shader-based-dynamic-2d-shadows/
//
// В отличие от geometry-based shadow volumes (extrusion), здесь ВСЁ
// считается пиксельными шейдерами вокруг источника света:
//
//   1) caster map    — окклюдеры рисуются в текстуру вокруг света
//   2) distance pass — для каждого пикселя: расстояние до центра, если там окклюдер
//   3) distort pass  — "распрямление" 4 квадрантов, чтобы лучи стали параллельны
//   4) reduction     — каскад min-редукции по горизонтали (256->128->...->2),
//                       получаем аналог 1D shadow map по углу
//   5) shadow pass   — сравниваем расстояние пикселя с shadow-картой по его лучу
//   6) blur          — сглаживание краёв тени
//   7) composite     — тёмная сцена + аддитивное наложение получившихся "лайт-текстур"
//
// go get github.com/gen2brain/raylib-go/raylib
// go run zima_light_raylib.go

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	screenW = 960
	screenH = 640
	LA      = 256 // размер квадратной области света (степень двойки)
)

// ---------------------------------------------------------------------
// Шейдеры (GLSL 330, соответствуют HLSL-коду из статьи Zima практически
// один в один, только переписаны под raylib/OpenGL).
// ---------------------------------------------------------------------

const fsDistance = `
#version 330
in vec2 fragTexCoord;
out vec4 finalColor;
uniform sampler2D texture0;

void main()
{
    vec4 c = texture(texture0, fragTexCoord);
    // если пиксель непрозрачный (окклюдер) - расстояние до центра, иначе "бесконечность"
    float d = (c.a > 0.3) ? distance(fragTexCoord, vec2(0.5)) : 1.0;
    finalColor = vec4(d, 0.0, 0.0, 1.0);
}`

const fsDistort = `
#version 330
in vec2 fragTexCoord;
out vec4 finalColor;
uniform sampler2D texture0;

void main()
{
    float u0 = fragTexCoord.x * 2.0 - 1.0;
    float v0 = fragTexCoord.y * 2.0 - 1.0;

    // "разворачиваем" квадрант - лучи из центра становятся параллельными
    v0 = v0 * abs(u0);
    v0 = (v0 + 1.0) / 2.0;

    vec2 newCoords = vec2(fragTexCoord.x, v0);

    float horizontal = texture(texture0, newCoords).r;
    float vertical   = texture(texture0, newCoords.yx).r;

    finalColor = vec4(horizontal, vertical, 0.0, 1.0);
}`

const fsReduce = `
#version 330
in vec2 fragTexCoord;
out vec4 finalColor;
uniform sampler2D texture0;
uniform float srcWidth; // ширина ВХОДНОЙ текстуры (не выходной)

void main()
{
    // gl_FragCoord.x - точная позиция текущего выходного пикселя.
    // Явно вычисляем координаты пары входных текселей (2*i, 2*i+1),
    // вместо того чтобы приближать её через fragTexCoord выходного
    // пикселя (это давало накапливающуюся погрешность по X на
    // каждом из 7 проходов редукции).
    float outX  = floor(gl_FragCoord.x - 0.5);
    float srcX0 = outX * 2.0;
    float srcX1 = srcX0 + 1.0;

    vec2 uv0 = vec2((srcX0 + 0.5) / srcWidth, fragTexCoord.y);
    vec2 uv1 = vec2((srcX1 + 0.5) / srcWidth, fragTexCoord.y);

    vec2 c0 = texture(texture0, uv0).rg;
    vec2 c1 = texture(texture0, uv1).rg;
    finalColor = vec4(min(c0, c1), 0.0, 1.0);
}`

const fsShadow = `
#version 330
in vec2 fragTexCoord;
out vec4 finalColor;
uniform sampler2D texture0; // финальная reduced shadow-map (ширина 2)
uniform vec4 lightColor;

float shadowDistanceH(vec2 uv)
{
    float u = abs(uv.x - 0.5) * 2.0;
    float v = uv.y * 2.0 - 1.0;
    float v0 = v / max(u, 0.0001);
    v0 = (v0 + 1.0) / 2.0;
    return texture(texture0, vec2(uv.x, v0)).r;
}

float shadowDistanceV(vec2 uv)
{
    float u = abs(uv.y - 0.5) * 2.0;
    float v = uv.x * 2.0 - 1.0;
    float v0 = v / max(u, 0.0001);
    v0 = (v0 + 1.0) / 2.0;
    return texture(texture0, vec2(uv.y, v0)).g;
}

void main()
{
    float dist = distance(fragTexCoord, vec2(0.5));
    dist -= 0.006; // небольшой bias, чтобы убрать самозатенение

    float du = abs(fragTexCoord.x - 0.5);
    float dv = abs(fragTexCoord.y - 0.5);

    // выбираем H- или V- карту в зависимости от доминирующего направления луча
    float shadowMapDistance = (du > dv) ? shadowDistanceH(fragTexCoord) : shadowDistanceV(fragTexCoord);

    float lit = (dist < shadowMapDistance) ? 1.0 : 0.0;
    float atten = clamp(1.0 - dist / 0.5, 0.0, 1.0);
    atten *= atten;

    finalColor = vec4(lightColor.rgb * lit * atten, 1.0);
}`

const fsBlur = `
#version 330
in vec2 fragTexCoord;
out vec4 finalColor;
uniform sampler2D texture0;
uniform vec2 direction;   // (1,0) или (0,1)
uniform float texelSize;  // 1.0 / LA

void main()
{
    vec2 off1 = direction * texelSize * 1.3846153846;
    vec2 off2 = direction * texelSize * 3.2307692308;

    vec3 sum = texture(texture0, fragTexCoord).rgb * 0.2270270270;
    sum += texture(texture0, fragTexCoord + off1).rgb * 0.3162162162;
    sum += texture(texture0, fragTexCoord - off1).rgb * 0.3162162162;
    sum += texture(texture0, fragTexCoord + off2).rgb * 0.0702702703;
    sum += texture(texture0, fragTexCoord - off2).rgb * 0.0702702703;

    finalColor = vec4(sum, 1.0);
}`

// ---------------------------------------------------------------------

type Occluder struct {
	Rect rl.Rectangle
}

// ZimaLight содержит все промежуточные render-текстуры пайплайна для
// одного источника света (переиспользуются между кадрами).
type ZimaLight struct {
	caster  rl.RenderTexture2D
	dist    rl.RenderTexture2D
	distort rl.RenderTexture2D
	reduce  []rl.RenderTexture2D // LA/2 ... 2, высота LA
	shadow  rl.RenderTexture2D
	blurA   rl.RenderTexture2D
	blurB   rl.RenderTexture2D
}

func newZimaLight() *ZimaLight {
	z := &ZimaLight{
		caster:  rl.LoadRenderTexture(LA, LA),
		dist:    rl.LoadRenderTexture(LA, LA),
		distort: rl.LoadRenderTexture(LA, LA),
		shadow:  rl.LoadRenderTexture(LA, LA),
		blurA:   rl.LoadRenderTexture(LA, LA),
		blurB:   rl.LoadRenderTexture(LA, LA),
	}
	for s := int32(LA / 2); s >= 2; s /= 2 {
		z.reduce = append(z.reduce, rl.LoadRenderTexture(s, LA))
	}

	// Точечная фильтрация: reduce/shadow-map строятся из отдельных дискретных
	// "лучей" по строкам текстуры, билинейная фильтрация между соседними
	// текселями (особенно на финальной 2-пиксельной карте) даёт неверные
	// промежуточные значения.
	rl.SetTextureFilter(z.dist.Texture, rl.FilterPoint)
	rl.SetTextureFilter(z.distort.Texture, rl.FilterPoint)
	for _, rt := range z.reduce {
		rl.SetTextureFilter(rt.Texture, rl.FilterPoint)
	}
	rl.SetTextureFilter(z.shadow.Texture, rl.FilterPoint)

	return z
}

func runPass(output rl.RenderTexture2D, outW, outH int32, input rl.Texture2D, shader rl.Shader) {
	rl.BeginTextureMode(output)
	rl.ClearBackground(rl.Blank)
	rl.BeginShaderMode(shader)
	// render-текстуры raylib физически хранятся перевёрнутыми по Y (OpenGL FBO),
	// поэтому КАЖДОЕ чтение входной RT-текстуры нужно "переворачивать" обратно
	// через отрицательную высоту src-rect - иначе к следующему шейдеру попадают
	// Y-зеркалированные данные, а distort-шейдер несимметричен по Y в G-канале
	// (переставляет местами x/y), из-за чего вертикальные тени считаются неверно.
	src := rl.NewRectangle(0, 0, float32(input.Width), -float32(input.Height))
	dst := rl.NewRectangle(0, 0, float32(outW), float32(outH))
	rl.DrawTexturePro(input, src, dst, rl.Vector2{}, 0, rl.White)
	rl.EndShaderMode()
	rl.EndTextureMode()
}

type Shaders struct {
	distance, distort, reduce, shadow, blur rl.Shader
	locReduceSrcWidth                       int32
	locShadowColor                          int32
	locBlurDir, locBlurTexel                int32
}

func loadShaders() *Shaders {
	s := &Shaders{
		distance: rl.LoadShaderFromMemory("", fsDistance),
		distort:  rl.LoadShaderFromMemory("", fsDistort),
		reduce:   rl.LoadShaderFromMemory("", fsReduce),
		shadow:   rl.LoadShaderFromMemory("", fsShadow),
		blur:     rl.LoadShaderFromMemory("", fsBlur),
	}
	s.locReduceSrcWidth = rl.GetShaderLocation(s.reduce, "srcWidth")
	s.locShadowColor = rl.GetShaderLocation(s.shadow, "lightColor")
	s.locBlurDir = rl.GetShaderLocation(s.blur, "direction")
	s.locBlurTexel = rl.GetShaderLocation(s.blur, "texelSize")
	return s
}

// Compute прогоняет весь пайплайн Zima и возвращает финальную LA x LA
// текстуру со светом+тенями для наложения на сцену (аддитивно).
func (z *ZimaLight) Compute(sh *Shaders, lightPos rl.Vector2, color rl.Color, occluders []Occluder) rl.Texture2D {
	// 1) caster map: рисуем окклюдеры в систему координат, центрированную на свете
	rl.BeginTextureMode(z.caster)
	rl.ClearBackground(rl.Blank)
	cam := rl.Camera2D{Target: lightPos, Offset: rl.Vector2{X: LA / 2, Y: LA / 2}, Zoom: 1}
	rl.BeginMode2D(cam)
	for _, o := range occluders {
		rl.DrawRectangleRec(o.Rect, rl.White)
	}
	rl.EndMode2D()
	rl.EndTextureMode()

	// 2) distance pass
	runPass(z.dist, LA, LA, z.caster.Texture, sh.distance)

	// 3) distort pass
	runPass(z.distort, LA, LA, z.dist.Texture, sh.distort)

	// 4) каскад min-редукции: LA/2 -> ... -> 2 (ширина), высота остаётся LA
	srcTex := z.distort.Texture
	srcW := int32(LA)
	for _, rt := range z.reduce {
		rl.SetShaderValue(sh.reduce, sh.locReduceSrcWidth, []float32{float32(srcW)}, rl.ShaderUniformFloat)
		runPass(rt, rt.Texture.Width, LA, srcTex, sh.reduce)
		srcTex = rt.Texture
		srcW = rt.Texture.Width
	}
	shadowMap := srcTex // финальная 2xLA "радиальная shadow map"

	// 5) shadow pass: сравниваем расстояние пикселя с shadow map по его лучу
	col := []float32{float32(color.R) / 255, float32(color.G) / 255, float32(color.B) / 255, 1}
	rl.SetShaderValue(sh.shadow, sh.locShadowColor, col, rl.ShaderUniformVec4)
	runPass(z.shadow, LA, LA, shadowMap, sh.shadow)

	// 6) блюр (горизонтальный + вертикальный проход)
	rl.SetShaderValue(sh.blur, sh.locBlurTexel, []float32{1.0 / float32(LA)}, rl.ShaderUniformFloat)
	rl.SetShaderValue(sh.blur, sh.locBlurDir, []float32{1, 0}, rl.ShaderUniformVec2)
	runPass(z.blurA, LA, LA, z.shadow.Texture, sh.blur)
	rl.SetShaderValue(sh.blur, sh.locBlurDir, []float32{0, 1}, rl.ShaderUniformVec2)
	runPass(z.blurB, LA, LA, z.blurA.Texture, sh.blur)

	return z.blurB.Texture
}

func main() {
	rl.SetConfigFlags(rl.FlagMsaa4xHint)
	rl.InitWindow(screenW, screenH, "Catalin Zima 2D shadows — raylib-go")
	rl.SetTargetFPS(60)

	sh := loadShaders()
	light1 := newZimaLight()
	light2 := newZimaLight()
	light3 := newZimaLight()

	occluders := []Occluder{
		{Rect: rl.NewRectangle(200, 150, 120, 80)},
		{Rect: rl.NewRectangle(500, 350, 150, 60)},
		{Rect: rl.NewRectangle(650, 120, 80, 180)},
		{Rect: rl.NewRectangle(100, 420, 200, 40)},
	}

	ambient := rl.NewColor(14, 14, 20, 255)

	for !rl.WindowShouldClose() {
		lightPos1 := rl.GetMousePosition()
		lightPos2 := rl.Vector2{X: 750, Y: 500}
		lightPos3 := rl.Vector2{X: 250, Y: 500}

		tex1 := light1.Compute(sh, lightPos1, rl.NewColor(255, 225, 170, 255), occluders)
		tex2 := light2.Compute(sh, lightPos2, rl.NewColor(140, 190, 255, 255), occluders)
		tex3 := light3.Compute(sh, lightPos3, rl.Red, occluders)

		rl.BeginDrawing()
		rl.ClearBackground(ambient)

		// тёмная "земля" со слегка видимыми силуэтами окклюдеров
		for _, o := range occluders {
			rl.DrawRectangleRec(o.Rect, rl.NewColor(40, 40, 46, 255))
		}

		// аддитивно накладываем результат пайплайна Zima для каждого света
		rl.BeginBlendMode(rl.BlendAdditive)
		for _, lp := range []struct {
			pos rl.Vector2
			tex rl.Texture2D
		}{{lightPos1, tex1}, {lightPos2, tex2}, {lightPos3, tex3}} {
			src := rl.NewRectangle(0, 0, LA, -LA) // флип по Y для render-текстуры
			dst := rl.NewRectangle(lp.pos.X-LA/2, lp.pos.Y-LA/2, LA, LA)
			rl.DrawTexturePro(lp.tex, src, dst, rl.Vector2{}, 0, rl.White)
		}
		rl.EndBlendMode()

		rl.DrawCircleV(lightPos1, 5, rl.Yellow)
		rl.DrawCircleV(lightPos2, 5, rl.SkyBlue)
		rl.DrawCircleV(lightPos3, 5, rl.Red)
		rl.DrawFPS(10, 10)
		rl.DrawText("Мышь — свет 1. Второй свет фиксирован.", 10, 40, 18, rl.RayWhite)

		rl.EndDrawing()
	}

	rl.CloseWindow()
}
