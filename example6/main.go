package main

import (
	"image/color"
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	// Внутреннее разрешение.
	// Вся графика считается именно здесь.
	ScreenWidth  = 160
	ScreenHeight = 90

	// Размер окна.
	WindowScale = 8

	// Количество уровней фиксированной палитры.
	PaletteSize = 16

	// Максимальное количество источников света.
	MaxLights = 8

	// Радиус игрока.
	PlayerRadius = 2.5

	// Скорость игрока.
	PlayerSpeed = 35.0
)

// ============================================================
// МАТЕМАТИКА
// ============================================================

type Vec2 struct {
	X float64
	Y float64
}

func v2(x, y float64) Vec2 {
	return Vec2{
		X: x,
		Y: y,
	}
}

func add(a, b Vec2) Vec2 {
	return Vec2{
		X: a.X + b.X,
		Y: a.Y + b.Y,
	}
}

func sub(a, b Vec2) Vec2 {
	return Vec2{
		X: a.X - b.X,
		Y: a.Y - b.Y,
	}
}

func mul(a Vec2, value float64) Vec2 {
	return Vec2{
		X: a.X * value,
		Y: a.Y * value,
	}
}

func length(v Vec2) float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y)
}

func normalize(v Vec2) Vec2 {
	l := length(v)

	if l < 0.0001 {
		return Vec2{}
	}

	return Vec2{
		X: v.X / l,
		Y: v.Y / l,
	}
}

func distance(a, b Vec2) float64 {
	return length(sub(a, b))
}

func clamp(v, min, max float64) float64 {
	if v < min {
		return min
	}

	if v > max {
		return max
	}

	return v
}

// ============================================================
// FIXED PALETTE
// ============================================================
//
// Здесь специально нет произвольных цветов освещения.
// В конце рендера цвет всегда выбирается из этой палитры.
//
// Палитру можно полностью заменить на свою.
// ============================================================

var palette = [PaletteSize]color.RGBA{
	{8, 8, 12, 255},
	{15, 15, 22, 255},
	{23, 22, 32, 255},
	{32, 29, 42, 255},
	{43, 36, 48, 255},
	{55, 43, 54, 255},
	{70, 52, 59, 255},
	{87, 62, 64, 255},
	{106, 73, 68, 255},
	{128, 87, 71, 255},
	{151, 101, 75, 255},
	{173, 116, 79, 255},
	{194, 132, 84, 255},
	{214, 149, 91, 255},
	{231, 169, 103, 255},
	{247, 191, 122, 255},
}

// ============================================================
// 4x4 BAYER DITHER
// ============================================================

var bayer4 = [4][4]int{
	{0, 8, 2, 10},
	{12, 4, 14, 6},
	{3, 11, 1, 9},
	{15, 7, 13, 5},
}

// ============================================================
// WALL
// ============================================================

type Wall struct {
	X float64
	Y float64
	W float64
	H float64
}

// ============================================================
// DYNAMIC OBJECT
// ============================================================
//
// Такой объект тоже может блокировать свет.
//
// Поэтому игрок и другие движущиеся объекты могут
// отбрасывать тени.
// ============================================================

type DynamicObject struct {
	Position Vec2
	Size     Vec2

	// Отрисовывается ли объект.
	Visible bool

	// Участвует ли объект в блокировке света.
	CastsShadow bool
}

// ============================================================
// LIGHT
// ============================================================

type Light struct {
	Position Vec2

	Radius float64

	Intensity float64

	// Цвет света.
	Color Vec3

	Enabled bool
}

type Vec3 struct {
	R float64
	G float64
	B float64
}

func vec3(r, g, b float64) Vec3 {
	return Vec3{
		R: r,
		G: g,
		B: b,
	}
}

func vec3Add(a, b Vec3) Vec3 {
	return Vec3{
		R: a.R + b.R,
		G: a.G + b.G,
		B: a.B + b.B,
	}
}

func vec3Mul(a Vec3, v float64) Vec3 {
	return Vec3{
		R: a.R * v,
		G: a.G * v,
		B: a.B * v,
	}
}

// ============================================================
// RGB -> PALETTE
// ============================================================

func paletteDistance(a, b color.RGBA) float64 {
	dr := float64(a.R) - float64(b.R)
	dg := float64(a.G) - float64(b.G)
	db := float64(a.B) - float64(b.B)

	return dr*dr + dg*dg + db*db
}

func nearestPaletteColor(c color.RGBA) color.RGBA {
	best := palette[0]
	bestDistance := math.MaxFloat64

	for _, p := range palette {
		d := paletteDistance(c, p)

		if d < bestDistance {
			bestDistance = d
			best = p
		}
	}

	return best
}

// ============================================================
// DITHER
// ============================================================
//
// Смешиваем текущий цвет с ближайшим более светлым цветом.
// ============================================================

func ditherColor(
	x,
	y int,
	c color.RGBA,
) color.RGBA {

	base := nearestPaletteColor(c)

	// Ищем ближайший цвет выше.
	baseIndex := 0

	bestDist := math.MaxFloat64

	for i, p := range palette {
		d := paletteDistance(c, p)

		if d < bestDist {
			bestDist = d
			baseIndex = i
		}
	}

	if baseIndex >= PaletteSize-1 {
		return base
	}

	next := palette[baseIndex+1]

	// Насколько исходный цвет ближе к следующему оттенку.
	baseDist := math.Sqrt(paletteDistance(c, base))
	nextDist := math.Sqrt(paletteDistance(c, next))

	total := baseDist + nextDist

	if total <= 0.001 {
		return base
	}

	fraction := baseDist / total

	threshold := float64(bayer4[y%4][x%4]) / 16.0

	if fraction > threshold {
		return next
	}

	return base
}

// ============================================================
// POINT INSIDE RECTANGLE
// ============================================================

func pointInsideWall(p Vec2, wall Wall) bool {
	return p.X >= wall.X &&
		p.X < wall.X+wall.W &&
		p.Y >= wall.Y &&
		p.Y < wall.Y+wall.H
}

func pointInsideObject(p Vec2, obj DynamicObject) bool {
	halfW := obj.Size.X / 2
	halfH := obj.Size.Y / 2

	return p.X >= obj.Position.X-halfW &&
		p.X <= obj.Position.X+halfW &&
		p.Y >= obj.Position.Y-halfH &&
		p.Y <= obj.Position.Y+halfH
}

// ============================================================
// SEGMENT VS RECTANGLE
// ============================================================
//
// Проверяет пересечение отрезка:
// start -> end
//
// с прямоугольником.
//
// Используется для raycasting теней.
// ============================================================

func segmentIntersectsRect(
	start Vec2,
	end Vec2,
	rectX float64,
	rectY float64,
	rectW float64,
	rectH float64,
) bool {

	dx := end.X - start.X
	dy := end.Y - start.Y

	tMin := 0.0
	tMax := 1.0

	// --------------------------------------------------------
	// X
	// --------------------------------------------------------

	if math.Abs(dx) < 0.000001 {

		if start.X < rectX ||
			start.X > rectX+rectW {
			return false
		}

	} else {

		tx1 := (rectX - start.X) / dx
		tx2 := (rectX + rectW - start.X) / dx

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

	if math.Abs(dy) < 0.000001 {

		if start.Y < rectY ||
			start.Y > rectY+rectH {
			return false
		}

	} else {

		ty1 := (rectY - start.Y) / dy
		ty2 := (rectY + rectH - start.Y) / dy

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

// ============================================================
// SHADOW RAYCAST
// ============================================================

func isPointShadowed(
	light Vec2,
	point Vec2,
	walls []Wall,
	objects []DynamicObject,
) bool {

	// --------------------------------------------------------
	// WALLS
	// --------------------------------------------------------

	for _, wall := range walls {

		if segmentIntersectsRect(
			light,
			point,
			wall.X,
			wall.Y,
			wall.W,
			wall.H,
		) {
			return true
		}
	}

	// --------------------------------------------------------
	// DYNAMIC OBJECTS
	// --------------------------------------------------------

	for _, object := range objects {

		if !object.CastsShadow {
			continue
		}

		halfW := object.Size.X / 2
		halfH := object.Size.Y / 2

		if segmentIntersectsRect(
			light,
			point,
			object.Position.X-halfW,
			object.Position.Y-halfH,
			object.Size.X,
			object.Size.Y,
		) {
			return true
		}
	}

	return false
}

// ============================================================
// LIGHT CALCULATION
// ============================================================

func calculateLight(
	light Light,
	point Vec2,
	walls []Wall,
	objects []DynamicObject,
) float64 {

	if !light.Enabled {
		return 0
	}

	dist := distance(light.Position, point)

	if dist >= light.Radius {
		return 0
	}

	// Нормализованное расстояние.
	t := dist / light.Radius

	// Плавное затухание.
	attenuation := math.Pow(1.0-t, 2.0)

	attenuation *= light.Intensity

	// --------------------------------------------------------
	// SHADOW
	// --------------------------------------------------------

	if isPointShadowed(
		light.Position,
		point,
		walls,
		objects,
	) {

		// Не полностью чёрная тень.
		// Это даёт немного ambient light.
		attenuation *= 0.12
	}

	return clamp(attenuation, 0, 1)
}

// ============================================================
// APPLY LIGHT
// ============================================================

func applyLight(
	base Vec3,
	light Light,
	amount float64,
) Vec3 {

	lightContribution := vec3Mul(
		light.Color,
		amount,
	)

	return vec3Add(
		base,
		lightContribution,
	)
}

// ============================================================
// RGB TO COLOR
// ============================================================

func vec3ToColor(v Vec3) color.RGBA {

	v.R = clamp(v.R, 0, 1)
	v.G = clamp(v.G, 0, 1)
	v.B = clamp(v.B, 0, 1)

	return color.RGBA{
		R: uint8(v.R * 255),
		G: uint8(v.G * 255),
		B: uint8(v.B * 255),
		A: 255,
	}
}

// ============================================================
// PLAYER COLLISION
// ============================================================

func collidesWithWall(
	position Vec2,
	radius float64,
	walls []Wall,
) bool {

	for _, wall := range walls {

		closestX := clamp(
			position.X,
			wall.X,
			wall.X+wall.W,
		)

		closestY := clamp(
			position.Y,
			wall.Y,
			wall.Y+wall.H,
		)

		dx := position.X - closestX
		dy := position.Y - closestY

		if dx*dx+dy*dy < radius*radius {
			return true
		}
	}

	return false
}

// ============================================================
// MOVE PLAYER
// ============================================================

func movePlayer(
	position Vec2,
	direction Vec2,
	speed float64,
	delta float64,
	walls []Wall,
) Vec2 {

	movement := mul(
		normalize(direction),
		speed*delta,
	)

	// X отдельно.
	nextX := add(
		position,
		v2(movement.X, 0),
	)

	if !collidesWithWall(
		nextX,
		PlayerRadius,
		walls,
	) {
		position.X = nextX.X
	}

	// Y отдельно.
	nextY := add(
		position,
		v2(0, movement.Y),
	)

	if !collidesWithWall(
		nextY,
		PlayerRadius,
		walls,
	) {
		position.Y = nextY.Y
	}

	return position
}

// ============================================================
// DRAW PIXEL ART PLAYER
// ============================================================

func drawPlayer(
	player Vec2,
	brightness float64,
) {

	// Получаем цвет игрока из палитры.
	base := paletteColorFromBrightness(
		brightness,
	)

	x := int32(player.X)
	y := int32(player.Y)

	// Небольшой пиксельный персонаж.
	rl.DrawRectangle(
		x-2,
		y-3,
		5,
		5,
		base,
	)

	// Голова.
	rl.DrawRectangle(
		x-1,
		y-4,
		3,
		2,
		palette[14],
	)
}

// ============================================================
// BRIGHTNESS -> PALETTE
// ============================================================

func paletteColorFromBrightness(
	brightness float64,
) color.RGBA {

	brightness = clamp(
		brightness,
		0,
		1,
	)

	index := int(
		brightness * float64(PaletteSize-1),
	)

	return palette[index]
}

// ============================================================
// MAIN
// ============================================================

func main() {

	rl.SetConfigFlags(
		rl.FlagWindowResizable,
	)

	rl.InitWindow(
		ScreenWidth*WindowScale,
		ScreenHeight*WindowScale,
		"Pixel Lighting / Raycast Shadows",
	)

	defer rl.CloseWindow()

	rl.SetTargetFPS(60)

	// --------------------------------------------------------
	// LOW RES RENDER TARGET
	// --------------------------------------------------------

	target := rl.LoadRenderTexture(
		ScreenWidth,
		ScreenHeight,
	)

	defer rl.UnloadRenderTexture(target)

	// Самое важное для пиксельного вида.
	rl.SetTextureFilter(
		target.Texture,
		rl.FilterPoint,
	)

	// ========================================================
	// WORLD
	// ========================================================

	walls := []Wall{

		// Outer walls.

		{
			X: 8,
			Y: 8,
			W: 144,
			H: 4,
		},

		{
			X: 8,
			Y: 78,
			W: 144,
			H: 4,
		},

		{
			X: 8,
			Y: 8,
			W: 4,
			H: 74,
		},

		{
			X: 148,
			Y: 8,
			W: 4,
			H: 74,
		},

		// ----------------------------------------------------
		// Центральные стены.
		// ----------------------------------------------------

		{
			X: 45,
			Y: 25,
			W: 6,
			H: 35,
		},

		{
			X: 95,
			Y: 18,
			W: 6,
			H: 32,
		},

		{
			X: 70,
			Y: 58,
			W: 45,
			H: 6,
		},

		{
			X: 118,
			Y: 35,
			W: 20,
			H: 6,
		},
	}

	// ========================================================
	// PLAYER
	// ========================================================

	player := Vec2{
		X: 80,
		Y: 45,
	}

	// ========================================================
	// DYNAMIC OBJECTS
	// ========================================================

	objects := []DynamicObject{

		// Игрок.
		{
			Position: player,
			Size: Vec2{
				X: 5,
				Y: 7,
			},
			Visible:     true,
			CastsShadow: true,
		},

		// Камень.
		{
			Position: Vec2{
				X: 55,
				Y: 18,
			},
			Size: Vec2{
				X: 5,
				Y: 5,
			},
			Visible:     true,
			CastsShadow: true,
		},

		// Второй объект.
		{
			Position: Vec2{
				X: 125,
				Y: 60,
			},
			Size: Vec2{
				X: 7,
				Y: 5,
			},
			Visible:     true,
			CastsShadow: true,
		},
	}

	// ========================================================
	// LIGHTS
	// ========================================================

	lights := []Light{

		// Тёплый свет слева.
		{
			Position: Vec2{
				X: 30,
				Y: 35,
			},
			Radius:    42,
			Intensity: 1.0,
			Color: vec3(
				1.0,
				0.55,
				0.25,
			),
			Enabled: true,
		},

		// Голубой свет справа.
		{
			Position: Vec2{
				X: 125,
				Y: 30,
			},
			Radius:    38,
			Intensity: 1.0,
			Color: vec3(
				0.25,
				0.55,
				1.0,
			),
			Enabled: true,
		},

		// Фиолетовый свет снизу.
		{
			Position: Vec2{
				X: 82,
				Y: 72,
			},
			Radius:    30,
			Intensity: 0.8,
			Color: vec3(
				0.65,
				0.25,
				1.0,
			),
			Enabled: true,
		},
	}

	// ========================================================
	// MAIN LOOP
	// ========================================================

	for !rl.WindowShouldClose() {

		delta := float64(
			rl.GetFrameTime(),
		)

		// ====================================================
		// INPUT
		// ====================================================

		direction := Vec2{}

		if rl.IsKeyDown(rl.KeyW) {
			direction.Y -= 1
		}

		if rl.IsKeyDown(rl.KeyS) {
			direction.Y += 1
		}

		if rl.IsKeyDown(rl.KeyA) {
			direction.X -= 1
		}

		if rl.IsKeyDown(rl.KeyD) {
			direction.X += 1
		}

		// ====================================================
		// PLAYER
		// ====================================================

		player = movePlayer(
			player,
			direction,
			PlayerSpeed,
			delta,
			walls,
		)

		// Первый dynamic object — игрок.
		objects[0].Position = player

		// ====================================================
		// ANIMATION
		// ====================================================

		time := float64(
			rl.GetTime(),
		)

		// Двигаем камень.
		objects[1].Position.X =
			55 +
				math.Sin(time*1.2)*8

		// Немного двигаем третий объект.
		objects[2].Position.Y =
			60 +
				math.Sin(time*0.8)*5

		// ====================================================
		// RENDER LOW RES
		// ====================================================

		rl.BeginTextureMode(target)

		rl.ClearBackground(
			palette[0],
		)

		// ====================================================
		// LIGHTING
		// ====================================================

		for y := 0; y < ScreenHeight; y++ {

			for x := 0; x < ScreenWidth; x++ {

				point := Vec2{
					X: float64(x) + 0.5,
					Y: float64(y) + 0.5,
				}

				// ------------------------------------------------
				// BASE AMBIENT
				// ------------------------------------------------

				// Базовый цвет сцены.
				base := vec3(
					0.035,
					0.035,
					0.055,
				)

				// ------------------------------------------------
				// LIGHTS
				// ------------------------------------------------

				for _, light := range lights {

					if !light.Enabled {
						continue
					}

					amount := calculateLight(
						light,
						point,
						walls,
						objects,
					)

					if amount <= 0 {
						continue
					}

					base = applyLight(
						base,
						light,
						amount,
					)
				}

				// ------------------------------------------------
				// COLOR RANGE
				// ------------------------------------------------

				base.R = clamp(
					base.R,
					0,
					1,
				)

				base.G = clamp(
					base.G,
					0,
					1,
				)

				base.B = clamp(
					base.B,
					0,
					1,
				)

				// ------------------------------------------------
				// PALETTE
				// ------------------------------------------------

				finalColor := vec3ToColor(
					base,
				)

				// ------------------------------------------------
				// DITHER
				// ------------------------------------------------

				finalColor = ditherColor(
					x,
					y,
					finalColor,
				)

				rl.DrawPixel(
					int32(x),
					int32(y),
					finalColor,
				)
			}
		}

		// ====================================================
		// WORLD GEOMETRY
		// ====================================================

		for _, wall := range walls {

			rl.DrawRectangle(
				int32(wall.X),
				int32(wall.Y),
				int32(wall.W),
				int32(wall.H),
				palette[3],
			)

			// Маленький верхний highlight.
			if wall.W > wall.H {

				rl.DrawRectangle(
					int32(wall.X),
					int32(wall.Y),
					int32(wall.W),
					1,
					palette[5],
				)
			}
		}

		// ====================================================
		// DYNAMIC OBJECTS
		// ====================================================

		for i, object := range objects {

			if !object.Visible {
				continue
			}

			halfW := object.Size.X / 2
			halfH := object.Size.Y / 2

			var objectColor color.RGBA

			if i == 0 {

				// Игрок.
				objectColor = palette[12]

			} else {

				// Остальные объекты.
				objectColor = palette[8]
			}

			rl.DrawRectangle(
				int32(object.Position.X-halfW),
				int32(object.Position.Y-halfH),
				int32(object.Size.X),
				int32(object.Size.Y),
				objectColor,
			)
		}

		// ====================================================
		// LIGHT SOURCES
		// ====================================================

		for _, light := range lights {

			if !light.Enabled {
				continue
			}

			// Переводим цвет света примерно в цвет палитры.
			lightColor := vec3ToColor(
				light.Color,
			)

			rl.DrawPixel(
				int32(light.Position.X),
				int32(light.Position.Y),
				lightColor,
			)

			rl.DrawPixel(
				int32(light.Position.X)+1,
				int32(light.Position.Y),
				lightColor,
			)

			rl.DrawPixel(
				int32(light.Position.X)-1,
				int32(light.Position.Y),
				lightColor,
			)

			rl.DrawPixel(
				int32(light.Position.X),
				int32(light.Position.Y)+1,
				lightColor,
			)

			rl.DrawPixel(
				int32(light.Position.X),
				int32(light.Position.Y)-1,
				lightColor,
			)
		}

		rl.EndTextureMode()

		// ====================================================
		// UPSCALE
		// ====================================================

		rl.BeginDrawing()

		rl.ClearBackground(
			color.RGBA{R: 0, G: 0, B: 0, A: 255},
		)

		screenWidth := float32(
			rl.GetScreenWidth(),
		)

		screenHeight := float32(
			rl.GetScreenHeight(),
		)

		rl.DrawTexturePro(
			target.Texture,

			rl.Rectangle{
				X:      0,
				Y:      0,
				Width:  float32(ScreenWidth),
				Height: -float32(ScreenHeight),
			},

			rl.Rectangle{
				X:      0,
				Y:      0,
				Width:  screenWidth,
				Height: screenHeight,
			},

			rl.Vector2{},

			0,

			color.RGBA{R: 255, G: 255, B: 255, A: 255},
		)

		// ====================================================
		// UI
		// ====================================================

		rl.DrawText(
			"WASD - MOVE",
			10,
			10,
			18,
			color.RGBA{R: 255, G: 255, B: 255, A: 255},
		)

		rl.EndDrawing()
	}
}
