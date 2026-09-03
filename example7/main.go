package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	screenW = 960
	screenH = 640
)

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
        ? distance(fragTexCoord, vec2(0.5))
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

// ============================================================
// BLUR
// ============================================================

const fsBlur = `
#version 330

in vec2 fragTexCoord;
out vec4 finalColor;

uniform sampler2D texture0;

uniform vec2 direction;

uniform float texelSize;

void main()
{
    vec2 off1 =
        direction *
        texelSize *
        1.3846153846;


    vec2 off2 =
        direction *
        texelSize *
        3.2307692308;


    vec3 sum =
        texture(
            texture0,
            fragTexCoord
        ).rgb *
        0.2270270270;


    sum +=
        texture(
            texture0,
            fragTexCoord + off1
        ).rgb *
        0.3162162162;


    sum +=
        texture(
            texture0,
            fragTexCoord - off1
        ).rgb *
        0.3162162162;


    sum +=
        texture(
            texture0,
            fragTexCoord + off2
        ).rgb *
        0.0702702703;


    sum +=
        texture(
            texture0,
            fragTexCoord - off2
        ).rgb *
        0.0702702703;


    finalColor =
        vec4(
            sum,
            1.0
        );
}
`

// ============================================================
// TYPES
// ============================================================

type Occluder struct {
	Rect rl.Rectangle
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

	blurA rl.RenderTexture2D

	blurB rl.RenderTexture2D
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

		blurA: rl.LoadRenderTexture(
			size,
			size,
		),

		blurB: rl.LoadRenderTexture(
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

	// Blur textures могут быть bilinear.

	rl.SetTextureFilter(
		z.blurA.Texture,
		rl.FilterBilinear,
	)

	rl.SetTextureFilter(
		z.blurB.Texture,
		rl.FilterBilinear,
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

	rl.UnloadRenderTexture(
		z.blurA,
	)

	rl.UnloadRenderTexture(
		z.blurB,
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
	distance rl.Shader

	distort rl.Shader

	reduce rl.Shader

	shadow rl.Shader

	blur rl.Shader

	locReduceSrcWidth int32

	locShadowColor int32

	locShadowIntensity int32

	locBlurDir int32

	locBlurTexel int32
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

		blur: rl.LoadShaderFromMemory(
			"",
			fsBlur,
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

	s.locBlurDir =
		rl.GetShaderLocation(
			s.blur,
			"direction",
		)

	s.locBlurTexel =
		rl.GetShaderLocation(
			s.blur,
			"texelSize",
		)

	return s
}

// ============================================================
// COMPUTE ZIMA LIGHT
// ============================================================

func (z *ZimaLight) Compute(
	sh *Shaders,
	lightPos rl.Vector2,
	occluders []Occluder,
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
			o.Rect,
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

	// ========================================================
	// 6. BLUR
	// ========================================================

	// ВАЖНО:
	//
	// texelSize зависит от Size конкретного источника.
	//
	// Поэтому:
	//
	// Size 128 -> 1/128
	// Size 256 -> 1/256
	// Size 512 -> 1/512

	texelSize :=
		1.0 /
			float32(size)

	rl.SetShaderValue(
		sh.blur,

		sh.locBlurTexel,

		[]float32{
			texelSize,
		},

		rl.ShaderUniformFloat,
	)

	// Horizontal

	rl.SetShaderValue(
		sh.blur,

		sh.locBlurDir,

		[]float32{
			1,
			0,
		},

		rl.ShaderUniformVec2,
	)

	runPass(
		z.blurA,

		size,
		size,

		z.shadow.Texture,

		sh.blur,
	)

	// Vertical

	rl.SetShaderValue(
		sh.blur,

		sh.locBlurDir,

		[]float32{
			0,
			1,
		},

		rl.ShaderUniformVec2,
	)

	runPass(
		z.blurB,

		size,
		size,

		z.blurA.Texture,

		sh.blur,
	)

	return z.blurB.Texture
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
				screenW,
				screenH,
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
			float32(screenW),
			-float32(screenH),
		)

	dst :=
		rl.NewRectangle(
			0,
			0,
			float32(screenW),
			float32(screenH),
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

// ============================================================
// MAIN
// ============================================================

func main() {

	rl.SetConfigFlags(
		rl.FlagMsaa4xHint,
	)

	rl.InitWindow(
		screenW,
		screenH,
		"Catalin Zima 2D shadows — multiple lights",
	)

	rl.SetTargetFPS(
		60,
	)

	// ========================================================
	// SHADERS
	// ========================================================

	sh :=
		loadShaders()

	// ========================================================
	// LIGHTS
	//
	// Size  != Radius
	//
	// Size:
	//     качество shadow map
	//
	// Radius:
	//     физический размер света
	// ========================================================

	light1 :=
		newZimaLight(

			128,

			110,

			rl.NewColor(
				255,
				220,
				150,
				255,
			),

			1.0,
		)

	light2 :=
		newZimaLight(

			256,

			170,

			rl.NewColor(
				120,
				180,
				255,
				255,
			),

			1.0,
		)

	light3 :=
		newZimaLight(

			512,

			280,

			rl.NewColor(
				255,
				80,
				80,
				255,
			),

			0.8,
		)

	// ========================================================
	// LIGHT BUFFER
	// ========================================================

	lightBuffer :=
		newLightBuffer()

	// ========================================================
	// OCCLUDERS
	// ========================================================

	occluders :=
		[]Occluder{

			{
				Rect: rl.NewRectangle(
					200,
					150,
					120,
					80,
				),
			},

			{
				Rect: rl.NewRectangle(
					500,
					350,
					150,
					60,
				),
			},

			{
				Rect: rl.NewRectangle(
					650,
					120,
					80,
					180,
				),
			},

			{
				Rect: rl.NewRectangle(
					100,
					420,
					200,
					40,
				),
			},

			// Центральный окклюдер.

			{
				Rect: rl.NewRectangle(
					430,
					230,
					100,
					160,
				),
			},
		}

	// ========================================================
	// AMBIENT
	// ========================================================

	ambient :=
		rl.NewColor(
			14,
			14,
			20,
			255,
		)

	// ========================================================
	// MAIN LOOP
	// ========================================================

	for !rl.WindowShouldClose() {

		// ====================================================
		// LIGHT POSITIONS
		// ====================================================

		lightPos1 :=
			rl.GetMousePosition()

		lightPos2 :=
			rl.Vector2{
				X: 760,
				Y: 480,
			}

		lightPos3 :=
			rl.Vector2{
				X: 230,
				Y: 500,
			}

		// ====================================================
		// COMPUTE LIGHT 1
		// ====================================================

		tex1 :=
			light1.Compute(
				sh,
				lightPos1,
				occluders,
			)

		// ====================================================
		// COMPUTE LIGHT 2
		// ====================================================

		tex2 :=
			light2.Compute(
				sh,
				lightPos2,
				occluders,
			)

		// ====================================================
		// COMPUTE LIGHT 3
		// ====================================================

		tex3 :=
			light3.Compute(
				sh,
				lightPos3,
				occluders,
			)

		// ====================================================
		// BUILD LIGHT BUFFER
		// ====================================================

		lightBuffer.Begin()

		// ----------------------------------------------------
		// ВАЖНО:
		//
		// Все источники складываются.
		//
		//     L1 + L2 + L3
		//
		// Ни один источник не вычитает другой.
		// ----------------------------------------------------

		rl.BeginBlendMode(
			rl.BlendAdditive,
		)

		addLight(
			tex1,
			lightPos1,
			light1.Radius,
		)

		addLight(
			tex2,
			lightPos2,
			light2.Radius,
		)

		addLight(
			tex3,
			lightPos3,
			light3.Radius,
		)

		rl.EndBlendMode()

		lightBuffer.End()

		// ====================================================
		// DRAW SCENE
		// ====================================================

		rl.BeginDrawing()

		rl.ClearBackground(
			ambient,
		)

		// ----------------------------------------------------
		// Dark scene
		// ----------------------------------------------------

		for _, o := range occluders {

			rl.DrawRectangleRec(
				o.Rect,

				rl.NewColor(
					40,
					40,
					46,
					255,
				),
			)
		}

		// ====================================================
		// ADD LIGHT BUFFER
		// ====================================================

		rl.BeginBlendMode(
			rl.BlendAdditive,
		)

		drawLightBuffer(
			lightBuffer,
		)

		rl.EndBlendMode()

		// ====================================================
		// DEBUG LIGHT POSITIONS
		// ====================================================

		rl.DrawCircleV(
			lightPos1,
			5,
			rl.Yellow,
		)

		rl.DrawCircleV(
			lightPos2,
			5,
			rl.SkyBlue,
		)

		rl.DrawCircleV(
			lightPos3,
			5,
			rl.Red,
		)

		// ====================================================
		// UI
		// ====================================================

		rl.DrawFPS(
			10,
			10,
		)

		rl.DrawText(
			"Mouse = Light 1 (Size 128 / Radius 110)",
			10,
			40,
			18,
			rl.RayWhite,
		)

		rl.DrawText(
			"Light 2 = Size 256 / Radius 170",
			10,
			62,
			18,
			rl.RayWhite,
		)

		rl.DrawText(
			"Light 3 = Size 512 / Radius 280",
			10,
			84,
			18,
			rl.RayWhite,
		)

		rl.EndDrawing()
	}

	// ========================================================
	// CLEANUP
	// ========================================================

	light1.Unload()

	light2.Unload()

	light3.Unload()

	lightBuffer.Unload()

	rl.UnloadShader(
		sh.distance,
	)

	rl.UnloadShader(
		sh.distort,
	)

	rl.UnloadShader(
		sh.reduce,
	)

	rl.UnloadShader(
		sh.shadow,
	)

	rl.UnloadShader(
		sh.blur,
	)

	rl.CloseWindow()
}
