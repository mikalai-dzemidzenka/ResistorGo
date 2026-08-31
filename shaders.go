package resistor

import rl "github.com/gen2brain/raylib-go/raylib"

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
    v0 = clamp((v0 + 1.0) / 2.0, 0.0, 1.0);
    return texture(texture0, vec2(uv.x, v0)).r;
}
 
float shadowDistanceV(vec2 uv)
{
    float u = abs(uv.y - 0.5) * 2.0;
    float v = uv.x * 2.0 - 1.0;
    float v0 = v / max(u, 0.0001);
    v0 = clamp((v0 + 1.0) / 2.0, 0.0, 1.0);
    return texture(texture0, vec2(uv.y, v0)).g;
}
 
void main()
{
    float dist = distance(fragTexCoord, vec2(0.5));
    dist -= 0.006; // небольшой bias, чтобы убрать самозатенение
 
    float du = abs(fragTexCoord.x - 0.5);
    float dv = abs(fragTexCoord.y - 0.5);
 
    // Обе карты (H и V) считаются всегда и плавно смешиваются около
    // диагоналей (du ~ dv), где они сходятся. Жёсткое переключение
    // (du > dv ? H : V) даёт видимый шов на диагоналях из-за
    // дискретизации - H и V карты в этой точке не идеально совпадают.
    float distH = shadowDistanceH(fragTexCoord);
    float distV = shadowDistanceV(fragTexCoord);
    float w = smoothstep(-0.0001, 0.0001, du - dv); // 0 -> берём V, 1 -> берём H
    float shadowMapDistance = mix(distV, distH, w);
 
    float lit = (dist < shadowMapDistance) ? 1.0 : 0.0;
    float atten = clamp(1.0 - dist / 0.5, 0.0, 1.0);
    atten *= atten;
 
    finalColor = vec4(lightColor.rgb * lit * atten, 1.0);
}`

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

// ZimaLight содержит все промежуточные render-текстуры пайплайна для
// одного источника света (переиспользуются между кадрами).
type ZimaLight struct {
	LA      int32
	caster  rl.RenderTexture2D
	dist    rl.RenderTexture2D
	distort rl.RenderTexture2D
	reduce  []rl.RenderTexture2D // LA/2 ... 2, высота LA
	shadow  rl.RenderTexture2D
}

func newZimaLight(LA int32) *ZimaLight {
	z := &ZimaLight{
		LA:      LA,
		caster:  rl.LoadRenderTexture(LA, LA),
		dist:    rl.LoadRenderTexture(LA, LA),
		distort: rl.LoadRenderTexture(LA, LA),
		shadow:  rl.LoadRenderTexture(LA, LA),
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
	distance, distort, reduce, shadow rl.Shader
	locReduceSrcWidth                 int32
	locShadowColor                    int32
}

func loadShaders() *Shaders {
	s := &Shaders{
		distance: rl.LoadShaderFromMemory("", fsDistance),
		distort:  rl.LoadShaderFromMemory("", fsDistort),
		reduce:   rl.LoadShaderFromMemory("", fsReduce),
		shadow:   rl.LoadShaderFromMemory("", fsShadow),
	}
	s.locReduceSrcWidth = rl.GetShaderLocation(s.reduce, "srcWidth")
	s.locShadowColor = rl.GetShaderLocation(s.shadow, "lightColor")
	return s
}

// Compute прогоняет весь пайплайн Zima и возвращает финальную LA x LA
// текстуру со светом+тенями для наложения на сцену (аддитивно).
func (z *ZimaLight) Compute(sh *Shaders, lightPos rl.Vector2, color rl.Color, occluders []rl.Rectangle) rl.Texture2D {
	// 1) caster map: рисуем окклюдеры в систему координат, центрированную на свете
	rl.BeginTextureMode(z.caster)
	rl.ClearBackground(rl.Blank)
	cam := rl.Camera2D{Target: lightPos, Offset: rl.Vector2{X: float32(z.LA) / 2, Y: float32(z.LA) / 2}, Zoom: 1}
	rl.BeginMode2D(cam)
	for _, o := range occluders {
		rl.DrawRectangleRec(o, rl.White)
	}
	rl.EndMode2D()
	rl.EndTextureMode()

	// 2) distance pass
	runPass(z.dist, z.LA, z.LA, z.caster.Texture, sh.distance)

	// 3) distort pass
	runPass(z.distort, z.LA, z.LA, z.dist.Texture, sh.distort)

	// 4) каскад min-редукции: LA/2 -> ... -> 2 (ширина), высота остаётся LA
	srcTex := z.distort.Texture
	srcW := z.LA
	for _, rt := range z.reduce {
		rl.SetShaderValue(sh.reduce, sh.locReduceSrcWidth, []float32{float32(srcW)}, rl.ShaderUniformFloat)
		runPass(rt, rt.Texture.Width, z.LA, srcTex, sh.reduce)
		srcTex = rt.Texture
		srcW = rt.Texture.Width
	}
	shadowMap := srcTex // финальная 2xLA "радиальная shadow map"

	// 5) shadow pass: сравниваем расстояние пикселя с shadow map по его лучу
	col := []float32{float32(color.R) / 255, float32(color.G) / 255, float32(color.B) / 255, 1}
	rl.SetShaderValue(sh.shadow, sh.locShadowColor, col, rl.ShaderUniformVec4)
	runPass(z.shadow, z.LA, z.LA, shadowMap, sh.shadow)

	return z.shadow.Texture
}
