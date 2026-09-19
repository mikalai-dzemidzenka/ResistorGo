package resistor

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"math"
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

const fsWhite = `
#version 330

in vec2 fragTexCoord;
in vec4 fragColor;

uniform sampler2D texture0;
uniform float time; // Передаем время из raylib (GetTime())

out vec4 finalColor;

void main()
{
    vec4 texel = texture(texture0, fragTexCoord) * fragColor;
    
    // sin(time) возвращает значения от -1.0 до 1.0.
    // abs() превращает их в диапазон от 0.0 до 1.0 (плавное мигание).
    // Умножение time * 10.0 ускоряет частоту мигания. Примените нужную вам скорость.
    float flashIntensity = abs(sin(time * 20.0));
    
    // Плавно смешиваем оригинальный цвет пикселя с чисто белым (vec3(1.0))
    vec3 mixedColor = mix(texel.rgb, vec3(1.0), flashIntensity);
    
    finalColor = vec4(mixedColor, texel.a);
}
`

// fsOutline traces the silhouette of a sprite frame, clipped to a band of it.
//
// It is an OUTER outline: it paints the transparent pixels that touch the body,
// so the fighter's own pixels are never recoloured and the sprite underneath
// stays exactly as it was drawn. Sampling is clamped to the frame's own
// rectangle in the sheet, or a neighbouring frame would bleed in and outline a
// pose the fighter is not in.
const fsOutline = `
#version 330

in vec2 fragTexCoord;
in vec4 fragColor;

uniform sampler2D texture0;
uniform vec2 texelSize;    // one texel, in texture coordinates
uniform vec4 frameBounds;  // u0, v0, u1, v1 of this frame within the sheet
uniform vec2 band;         // the v range the swing lands on
uniform vec4 outlineColor;

out vec4 finalColor;

// alphaAt reads the sprite's coverage, treating anything outside this frame as
// empty so the outline stops at the frame's edge instead of sampling the next.
float alphaAt(vec2 uv)
{
    if (uv.x < frameBounds.x || uv.x > frameBounds.z ||
        uv.y < frameBounds.y || uv.y > frameBounds.w) {
        return 0.0;
    }
    return texture(texture0, uv).a;
}

void main()
{
    vec2 uv = fragTexCoord;

    // outside the band the swing would land on: nothing is marked
    if (uv.y < band.x || uv.y > band.y) {
        discard;
    }
    // inside the body: leave the fighter's own pixels alone
    if (alphaAt(uv) > 0.5) {
        discard;
    }
    // a transparent pixel is part of the outline only if the body touches it
    float neighbours = alphaAt(uv + vec2(texelSize.x, 0.0))
                     + alphaAt(uv - vec2(texelSize.x, 0.0))
                     + alphaAt(uv + vec2(0.0, texelSize.y))
                     + alphaAt(uv - vec2(0.0, texelSize.y));
    if (neighbours < 0.5) {
        discard;
    }

    // Solid along the middle of the band - the line the swing actually lands
    // on - fading to nothing at its top and bottom. The mark has no hard edge
    // where the band stops, and the brightest row is the one that would be hit.
    float middle = (band.x + band.y) * 0.5;
    float reach = max((band.y - band.x) * 0.5, 1e-6);
    float fade = clamp(1.0 - abs(uv.y - middle) / reach, 0.0, 1.0);

    finalColor = vec4(outlineColor.rgb, outlineColor.a * fade);
}
`

// fsRim lights the ring of empty pixels around a sprite frame on the sides that
// face each given light.
//
// It is an OUTER rim, like fsOutline: it paints the transparent texels that
// touch the fighter, so no pixel the fighter draws is recoloured. Which way a
// rim texel faces comes from where the body is next to it, and it is lit by how
// squarely the opposite direction points at the light. Everything else is
// discarded, and the pass is drawn additively, so a side facing away is left as
// the scene drew it.
//
// Work is done per texel, not per screen pixel: every screen pixel of one
// sprite texel gets the same answer at any zoom. The quad is drawn a texel
// larger than the frame all round, so the ring is not cut off where the sprite
// reaches the frame's edge; sampling stays clamped to the frame itself.
//
// A fighter drawn in two layers is one silhouette here: coverage is the greater
// of the body sprite's alpha and the top layer's, so the rim runs around the
// two together and never along the seam between them.
//
// Light positions are in the frame's own texels (already mirrored for a fighter
// facing left); their colours carry their strength.
const fsRim = `
#version 330

in vec2 fragTexCoord;
in vec4 fragColor;

uniform sampler2D texture0;   // the body sprite's sheet
uniform sampler2D topTexture; // the top layer's sheet, if there is one
uniform vec2 texelSize;       // one texel of texture0, in texture coordinates
uniform vec4 frameBounds;     // u0, v0, u1, v1 of this frame within the sheet
uniform vec2 topTexelSize;    // one texel of topTexture
uniform vec4 topBounds;       // this frame within the top layer's sheet
uniform float hasTop;         // 1 when the fighter has a top layer
uniform float lightCount;     // how many of the entries below are in use
uniform vec2 lightPos[4];     // in the frame's texels
uniform vec3 lightColor[4];
uniform float rimCap;         // the brightest a rim may add, as a fraction of white

out vec4 finalColor;

// alphaAt is how covered the fighter is at a pixel of the frame, counting both
// layers. Anything outside the frame is empty, so the rim cannot sample the
// pose next to it.
float alphaAt(vec2 pixel)
{
    vec2 uv = frameBounds.xy + pixel * texelSize;
    if (uv.x < frameBounds.x || uv.x > frameBounds.z ||
        uv.y < frameBounds.y || uv.y > frameBounds.w) {
        return 0.0;
    }
    float a = texture(texture0, uv).a;
    if (hasTop > 0.5) {
        a = max(a, texture(topTexture, topBounds.xy + pixel * topTexelSize).a);
    }
    return a;
}

void main()
{
    // the texel this fragment belongs to, in the frame's own grid
    vec2 texel = floor((fragTexCoord - frameBounds.xy) / texelSize);
    vec2 centre = texel + 0.5;

    if (alphaAt(centre) >= 0.5) {
        discard; // the fighter's own pixel: leave it exactly as it was drawn
    }

    // towards the body: the neighbours the fighter covers
    vec2 body = vec2(0.0);
    body += vec2( 1.0,  0.0) * alphaAt(centre + vec2(1.0, 0.0));
    body += vec2(-1.0,  0.0) * alphaAt(centre - vec2(1.0, 0.0));
    body += vec2( 0.0,  1.0) * alphaAt(centre + vec2(0.0, 1.0));
    body += vec2( 0.0, -1.0) * alphaAt(centre - vec2(0.0, 1.0));
    if (dot(body, body) < 0.01) {
        discard; // nothing adjacent, so not part of the ring
    }
    vec2 outward = -normalize(body);

    vec3 sum = vec3(0.0);
    int count = int(lightCount + 0.5);
    for (int i = 0; i < 4; i++) {
        if (i >= count) {
            break;
        }
        vec2 toLight = lightPos[i] - centre;
        if (dot(toLight, toLight) < 1e-4) {
            continue;
        }
        float f = max(dot(outward, normalize(toLight)), 0.0);
        sum += lightColor[i] * f;
    }
    float peak = max(sum.r, max(sum.g, sum.b));
    if (peak <= 0.0) {
        discard;
    }
    // a rim is added on top of a scene the same lights have already lit, so
    // near a lamp - or inside several - it would add up past white into a flat
    // sticker border; scaled as a whole, it keeps its colour under the ceiling
    if (peak > rimCap) {
        sum *= rimCap / peak;
    }
    finalColor = vec4(sum, 1.0);
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

uniform float lightDirection;
uniform float coneAngle;
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

	float angle = atan(p.y, p.x);
	
	float diff = abs(atan(
		sin(angle - lightDirection),
		cos(angle - lightDirection)
	));
	
	float cone = step(diff, coneAngle * 0.5);


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

	float lit = (dist <= shadowMapDistance) ? 1.0 : 0.0;
	
	lit *= cone;


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

	Direction float32 // радианы
	ConeAngle float32 // радианы
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
		Direction: 0,
		ConeAngle: 2 * math.Pi,
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
	distance                rl.Shader
	distort                 rl.Shader
	reduce                  rl.Shader
	shadow                  rl.Shader
	locReduceSrcWidth       int32
	locShadowColor          int32
	locShadowIntensity      int32
	locShadowlightDirection int32
	locConeAngle            int32
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
	s.locShadowlightDirection = rl.GetShaderLocation(s.shadow, "lightDirection")
	s.locConeAngle = rl.GetShaderLocation(s.shadow, "coneAngle")

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

	rl.SetShaderValue(
		sh.shadow,
		sh.locShadowlightDirection,
		[]float32{z.Direction},
		rl.ShaderUniformFloat,
	)

	rl.SetShaderValue(
		sh.shadow,
		sh.locConeAngle,
		[]float32{z.ConeAngle},
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
