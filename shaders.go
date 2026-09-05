package resistor

import rl "github.com/gen2brain/raylib-go/raylib"

// ============================================================
// SHADERS
// ============================================================

const fsDistance = `
#version 330

in vec2 fragTexCoord;
out vec4 finalColor;

uniform sampler2D texture0;

void main()
{
    vec4 c = texture(texture0, fragTexCoord);

    float d =
        (c.a > 0.3)
        ? distance(fragTexCoord, vec2(0.5))  * 2.0
        : 1.0;

    finalColor = vec4(d, 0.0, 0.0, 1.0);
}
`

// ============================================================
// DISTORT
// ============================================================

const fsDistort = `
#version 330

in vec2 fragTexCoord;
out vec4 finalColor;

uniform sampler2D texture0;

void main()
{
    float u0 =
        fragTexCoord.x * 2.0 - 1.0;

    float v0 =
        fragTexCoord.y * 2.0 - 1.0;


    // --------------------------------------------------------
    // Catalin Zima distortion
    // --------------------------------------------------------

    v0 =
        v0 * abs(u0);

    v0 =
        (v0 + 1.0) * 0.5;


    vec2 newCoords =
        vec2(
            fragTexCoord.x,
            v0
        );


    float horizontal =
        texture(
            texture0,
            newCoords
        ).r;


    float vertical =
        texture(
            texture0,
            newCoords.yx
        ).r;


    finalColor =
        vec4(
            horizontal,
            vertical,
            0.0,
            1.0
        );
}
`

// ============================================================
// REDUCTION
// ============================================================

const fsReduce = `
#version 330

in vec2 fragTexCoord;
out vec4 finalColor;

uniform sampler2D texture0;

// Ширина входной текстуры.
// Например:
// 256 -> 128
// 128 -> 64
// 64  -> 32
// ...
uniform float srcWidth;

void main()
{
    // --------------------------------------------------------
    // Точный индекс выходного пикселя.
    //
    // Не используем fragTexCoord.x для вычисления пары,
    // поскольку при нескольких reduction passes ошибка
    // начинает накапливаться.
    // --------------------------------------------------------

    float outX =
        floor(gl_FragCoord.x - 0.5);


    float srcX0 =
        outX * 2.0;


    float srcX1 =
        srcX0 + 1.0;


    vec2 uv0 =
        vec2(
            (srcX0 + 0.5) / srcWidth,
            fragTexCoord.y
        );


    vec2 uv1 =
        vec2(
            (srcX1 + 0.5) / srcWidth,
            fragTexCoord.y
        );


    vec2 c0 =
        texture(
            texture0,
            uv0
        ).rg;


    vec2 c1 =
        texture(
            texture0,
            uv1
        ).rg;


    finalColor =
        vec4(
            min(c0, c1),
            0.0,
            1.0
        );
}
`

// ============================================================
// SHADOW
//
// Финальная reduction texture:
//
//     width  = 2
//     height = Size
//
// Поэтому shadow lookup должен быть дискретным.
// Используем texelFetch.
//
// НИКАКОГО hardcoded 256.
// ============================================================

const fsShadow = `
#version 330

in vec2 fragTexCoord;
out vec4 finalColor;

uniform sampler2D texture0;
uniform vec4 lightColor;
uniform float lightIntensity;


// ============================================================
// HORIZONTAL
// ============================================================

float shadowDistanceH(vec2 uv)
{
    float u =
        uv.x * 2.0 - 1.0;

    float v =
        uv.y * 2.0 - 1.0;

    // Catalin Zima:
    // as U approaches zero, V approaches zero
    v /= max(abs(u), 0.000001);

    v =
        (v + 1.0) * 0.5;

    ivec2 size =
        textureSize(texture0, 0);

    int y =
        int(
            clamp(
                floor(v * float(size.y)),
                0.0,
                float(size.y - 1)
            )
        );

    int x =
        (uv.x < 0.5)
        ? 0
        : 1;

    return texelFetch(
        texture0,
        ivec2(x, y),
        0
    ).r;
}


// ============================================================
// VERTICAL
// ============================================================

float shadowDistanceV(vec2 uv)
{
    float u =
        uv.y * 2.0 - 1.0;

    float v =
        uv.x * 2.0 - 1.0;

    v /= max(abs(u), 0.000001);

    v =
        (v + 1.0) * 0.5;

    ivec2 size =
        textureSize(texture0, 0);

    int y =
        int(
            clamp(
                floor(v * float(size.y)),
                0.0,
                float(size.y - 1)
            )
        );

    int x =
        (uv.y < 0.5)
        ? 0
        : 1;

    return texelFetch(
        texture0,
        ivec2(x, y),
        0
    ).g;
}


// ============================================================
// MAIN
// ============================================================

void main()
{
    vec2 p =
        fragTexCoord - vec2(0.5);


    // ========================================================
    // КРИТИЧЕСКОЕ ИСПРАВЛЕНИЕ
    //
    // Расстояние должно быть:
    //
    // center = 0
    // edge   = 1
    //
    // Поэтому * 2.0.
    //
    // В предыдущей версии здесь отсутствовал * 2.0.
    // ========================================================

    float dist =
        length(p) * 2.0;


    // Небольшой bias против self-shadowing.

    dist -= 0.006;


    float du =
        abs(p.x);

    float dv =
        abs(p.y);


    // ========================================================
    // Выбор горизонтальной / вертикальной карты
    // ========================================================

    float shadowMapDistance;

    if (du > dv)
    {
        shadowMapDistance =
            shadowDistanceH(
                fragTexCoord
            );
    }
    else
    {
        shadowMapDistance =
            shadowDistanceV(
                fragTexCoord
            );
    }


    // ========================================================
    // Shadow test
    // ========================================================

    float lit =
        (dist <= shadowMapDistance)
        ? 1.0
        : 0.0;


    // ========================================================
    // Radial attenuation
    //
    // dist:
    //
    // 0 = центр
    // 1 = край
    // ========================================================

    float atten =
        clamp(
            1.0 - dist,
            0.0,
            1.0
        );


    atten *= atten;


    float value =
        lit *
        atten *
        lightIntensity;


    finalColor =
        vec4(
            lightColor.rgb * value,
            value
        );
}
`

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

// ============================================================
// ZIMA LIGHT
// ============================================================

type ZimaLight struct {

	// --------------------------------------------------------
	// Разрешение внутренних render textures.
	//
	// Это НЕ радиус.
	//
	// Например:
	//
	// Size = 128
	// Size = 256
	// Size = 512
	// --------------------------------------------------------

	Size int32

	// --------------------------------------------------------
	// Реальный размер света в world coordinates.
	//
	// Свет занимает:
	//
	// X = Position.X - Radius ... Position.X + Radius
	// Y = Position.Y - Radius ... Position.Y + Radius
	// --------------------------------------------------------

	Radius float32

	// --------------------------------------------------------
	// Интенсивность
	// --------------------------------------------------------

	Intensity float32

	// --------------------------------------------------------
	// Цвет
	// --------------------------------------------------------

	Color rl.Color

	// --------------------------------------------------------
	// Render textures
	// --------------------------------------------------------

	caster rl.RenderTexture2D

	dist rl.RenderTexture2D

	distort rl.RenderTexture2D

	reduce []rl.RenderTexture2D

	shadow rl.RenderTexture2D
}

// ============================================================
// CREATE ZIMA LIGHT
// ============================================================

func newZimaLight(
	size int32,
	radius float32,
	color rl.Color,
	intensity float32,
) *ZimaLight {

	if size < 2 {
		size = 2
	}

	z := &ZimaLight{

		Size: size,

		Radius: radius,

		Color: color,

		Intensity: intensity,

		caster: rl.LoadRenderTexture(
			size,
			size,
		),

		dist: rl.LoadRenderTexture(
			size,
			size,
		),

		distort: rl.LoadRenderTexture(
			size,
			size,
		),

		shadow: rl.LoadRenderTexture(
			size,
			size,
		),
	}

	// ========================================================
	// REDUCTION CHAIN
	//
	// Size = 256:
	//
	// 256 -> 128 -> 64 -> 32 -> 16 -> 8 -> 4 -> 2
	//
	// Size = 128:
	//
	// 128 -> 64 -> 32 -> 16 -> 8 -> 4 -> 2
	//
	// ========================================================

	for s := size / 2; s >= 2; s /= 2 {

		z.reduce =
			append(
				z.reduce,

				rl.LoadRenderTexture(
					s,
					size,
				),
			)
	}

	// ========================================================
	// FILTERING
	//
	// Shadow-map pipeline должен быть point sampled.
	// ========================================================

	rl.SetTextureFilter(
		z.caster.Texture,
		rl.FilterPoint,
	)

	rl.SetTextureFilter(
		z.dist.Texture,
		rl.FilterPoint,
	)

	rl.SetTextureFilter(
		z.distort.Texture,
		rl.FilterPoint,
	)

	for _, rt := range z.reduce {

		rl.SetTextureFilter(
			rt.Texture,
			rl.FilterPoint,
		)
	}

	rl.SetTextureFilter(
		z.shadow.Texture,
		rl.FilterPoint,
	)

	return z
}

// ============================================================
// UNLOAD
// ============================================================

func (z *ZimaLight) Unload() {

	rl.UnloadRenderTexture(
		z.caster,
	)

	rl.UnloadRenderTexture(
		z.dist,
	)

	rl.UnloadRenderTexture(
		z.distort,
	)

	for _, rt := range z.reduce {

		rl.UnloadRenderTexture(
			rt,
		)
	}

	rl.UnloadRenderTexture(
		z.shadow,
	)
}

// ============================================================
// GENERIC SHADER PASS
// ============================================================

func runPass(
	output rl.RenderTexture2D,
	outW int32,
	outH int32,
	input rl.Texture2D,
	shader rl.Shader,
) {

	rl.BeginTextureMode(
		output,
	)

	rl.ClearBackground(
		rl.Blank,
	)

	rl.BeginShaderMode(
		shader,
	)

	// ========================================================
	// Raylib render texture Y-flip.
	//
	// Это обязательно для переходов:
	//
	// caster -> distance
	// distance -> distort
	// distort -> reduce
	// reduce -> shadow
	// shadow -> blur
	// ========================================================

	src :=
		rl.NewRectangle(
			0,
			0,
			float32(input.Width),
			-float32(input.Height),
		)

	dst :=
		rl.NewRectangle(
			0,
			0,
			float32(outW),
			float32(outH),
		)

	rl.DrawTexturePro(
		input,
		src,
		dst,
		rl.Vector2{},
		0,
		rl.White,
	)

	rl.EndShaderMode()

	rl.EndTextureMode()
}

// ============================================================
// SHADERS
// ============================================================

type Shaders struct {
	distance           rl.Shader
	distort            rl.Shader
	reduce             rl.Shader
	shadow             rl.Shader
	locReduceSrcWidth  int32
	locShadowColor     int32
	locShadowIntensity int32
}

// ============================================================
// LOAD SHADERS
// ============================================================

func loadShaders() *Shaders {

	s := &Shaders{

		distance: rl.LoadShaderFromMemory(
			"",
			fsDistance,
		),

		distort: rl.LoadShaderFromMemory(
			"",
			fsDistort,
		),

		reduce: rl.LoadShaderFromMemory(
			"",
			fsReduce,
		),

		shadow: rl.LoadShaderFromMemory(
			"",
			fsShadow,
		),
	}

	s.locReduceSrcWidth =
		rl.GetShaderLocation(
			s.reduce,
			"srcWidth",
		)

	s.locShadowColor =
		rl.GetShaderLocation(
			s.shadow,
			"lightColor",
		)

	s.locShadowIntensity =
		rl.GetShaderLocation(
			s.shadow,
			"lightIntensity",
		)

	return s
}

// ============================================================
// COMPUTE ZIMA LIGHT
// ============================================================

func (z *ZimaLight) Compute(
	sh *Shaders,
	lightPos rl.Vector2,
	occluders []rl.Rectangle,
) rl.Texture2D {

	size :=
		z.Size

	// ========================================================
	// 1. CASTER
	// ========================================================

	rl.BeginTextureMode(
		z.caster,
	)

	rl.ClearBackground(
		rl.Blank,
	)

	// ========================================================
	// КЛЮЧЕВОЕ ИЗМЕНЕНИЕ
	//
	// Size = resolution
	//
	// Radius = world-space размер.
	//
	// Благодаря Zoom одна и та же геометрия света
	// корректно отображается для 128 / 256 / 512.
	// ========================================================

	cam :=
		rl.Camera2D{

			Target: lightPos,

			Offset: rl.Vector2{
				X: float32(size) * 0.5,
				Y: float32(size) * 0.5,
			},

			Zoom: float32(size) /
				(z.Radius * 2.0),
		}

	rl.BeginMode2D(
		cam,
	)

	for _, o := range occluders {
		rl.DrawRectangleRec(
			o,
			rl.White,
		)
	}

	rl.EndMode2D()

	rl.EndTextureMode()

	// ========================================================
	// 2. DISTANCE
	// ========================================================

	runPass(
		z.dist,

		size,
		size,

		z.caster.Texture,

		sh.distance,
	)

	// ========================================================
	// 3. DISTORT
	// ========================================================

	runPass(
		z.distort,

		size,
		size,

		z.dist.Texture,

		sh.distort,
	)

	// ========================================================
	// 4. REDUCTION
	// ========================================================

	srcTex :=
		z.distort.Texture

	srcW :=
		size

	for _, rt := range z.reduce {

		rl.SetShaderValue(
			sh.reduce,

			sh.locReduceSrcWidth,

			[]float32{
				float32(srcW),
			},

			rl.ShaderUniformFloat,
		)

		runPass(
			rt,

			rt.Texture.Width,
			size,

			srcTex,

			sh.reduce,
		)

		srcTex =
			rt.Texture

		srcW =
			rt.Texture.Width
	}

	// Финальная карта:
	//
	// 2 x Size

	shadowMap :=
		srcTex

	// ========================================================
	// 5. SHADOW
	// ========================================================

	colorData :=
		[]float32{

			float32(z.Color.R) /
				255.0,

			float32(z.Color.G) /
				255.0,

			float32(z.Color.B) /
				255.0,

			1.0,
		}

	rl.SetShaderValue(
		sh.shadow,

		sh.locShadowColor,

		colorData,

		rl.ShaderUniformVec4,
	)

	rl.SetShaderValue(
		sh.shadow,

		sh.locShadowIntensity,

		[]float32{
			z.Intensity,
		},

		rl.ShaderUniformFloat,
	)

	runPass(
		z.shadow,

		size,
		size,

		shadowMap,

		sh.shadow,
	)

	return z.shadow.Texture
}

// ============================================================
// LIGHT BUFFER
// ============================================================
//
// ВСЕ источники сначала складываются сюда.
//
// Потом LightBuffer один раз рисуется на сцену.
//
// Поэтому:
//
//     Shadow(Light A)
//
// не может удалить:
//
//     Light B
//
// ============================================================

type LightBuffer struct {
	rt rl.RenderTexture2D
}

// ============================================================
// CREATE LIGHT BUFFER
// ============================================================

func newLightBuffer() *LightBuffer {

	l :=
		&LightBuffer{

			rt: rl.LoadRenderTexture(
				SW,
				SH,
			),
		}

	rl.SetTextureFilter(
		l.rt.Texture,
		rl.FilterBilinear,
	)

	return l
}

// ============================================================
// BEGIN
// ============================================================

func (l *LightBuffer) Begin() {

	rl.BeginTextureMode(
		l.rt,
	)

	// Никакого ambient здесь нет.
	// LightBuffer начинается с чёрного.

	rl.ClearBackground(
		rl.Black,
	)
}

// ============================================================
// END
// ============================================================

func (l *LightBuffer) End() {

	rl.EndTextureMode()
}

// ============================================================
// UNLOAD
// ============================================================

func (l *LightBuffer) Unload() {

	rl.UnloadRenderTexture(
		l.rt,
	)
}

// ============================================================
// ADD LIGHT
// ============================================================

func addLight(
	texture rl.Texture2D,
	position rl.Vector2,
	radius float32,
) {

	diameter :=
		radius * 2.0

	// RenderTexture Y flip.

	src :=
		rl.NewRectangle(
			0,
			0,
			float32(texture.Width),
			-float32(texture.Height),
		)

	// ========================================================
	// WORLD SPACE
	//
	// Именно здесь Radius используется как физический размер.
	// ========================================================

	dst :=
		rl.NewRectangle(

			position.X-radius,

			position.Y-radius,

			diameter,

			diameter,
		)

	rl.DrawTexturePro(
		texture,
		src,
		dst,
		rl.Vector2{},
		0,
		rl.White,
	)
}

// ============================================================
// DRAW LIGHT BUFFER
// ============================================================

func drawLightBuffer(
	buffer *LightBuffer,
) {

	src :=
		rl.NewRectangle(
			0,
			0,
			float32(SW),
			-float32(SH),
		)

	dst :=
		rl.NewRectangle(
			0,
			0,
			float32(SW),
			float32(SH),
		)

	rl.DrawTexturePro(
		buffer.rt.Texture,
		src,
		dst,
		rl.Vector2{},
		0,
		rl.White,
	)
}
