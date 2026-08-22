package main

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	screenWidth  = 960
	screenHeight = 640

	tileSize = 16

	mapWidth  = 60
	mapHeight = 40

	lightLevels = 6

	ambientLight = 0.08
)

type Light struct {
	X, Y      float64
	Radius    float64
	Intensity float64
}

// ------------------------------------------------------------
// FIXED PALETTE
// ------------------------------------------------------------

var palette = []rl.Color{
	{8, 10, 16, 255},
	{24, 28, 38, 255},
	{42, 48, 62, 255},
	{65, 72, 88, 255},
	{91, 98, 112, 255},
	{122, 126, 134, 255},
	{157, 157, 151, 255},
	{190, 183, 161, 255},
	{222, 201, 157, 255},
	{244, 215, 143, 255},
	{222, 139, 92, 255},
	{185, 78, 74, 255},
	{117, 61, 70, 255},
	{72, 68, 92, 255},
	{55, 90, 94, 255},
	{91, 127, 91, 255},
}

// ------------------------------------------------------------
// WORLD
// ------------------------------------------------------------

// 0 = empty
// 1 = wall
var world [mapHeight][mapWidth]uint8

func buildWorld() {
	for y := 0; y < mapHeight; y++ {
		for x := 0; x < mapWidth; x++ {
			if x == 0 ||
				y == 0 ||
				x == mapWidth-1 ||
				y == mapHeight-1 {
				world[y][x] = 1
			}
		}
	}

	// Горизонтальные стены.
	for x := 6; x < 50; x++ {
		world[10][x] = 1
	}

	for x := 15; x < 45; x++ {
		world[28][x] = 1
	}

	// Вертикальные стены.
	for y := 10; y < 29; y++ {
		world[y][20] = 1
	}

	for y := 5; y < 28; y++ {
		world[y][42] = 1
	}

	// Двери.
	world[10][13] = 0
	world[10][35] = 0
	world[20][20] = 0
	world[24][42] = 0
	world[28][32] = 0
}

// ------------------------------------------------------------
// DDA
// ------------------------------------------------------------
//
// Луч от (ox, oy) в направлении (dx, dy).
//
// Возвращает расстояние до первого solid tile.
// Если препятствия нет — maxDistance.
//
// ВАЖНО:
// ox/oy — координаты мира в tile-space.
// То есть клетка 10,5 находится около x=10.5 y=5.5.
// ------------------------------------------------------------

func castRay(
	ox, oy float64,
	dx, dy float64,
	maxDistance float64,
) float64 {
	mapX := int(math.Floor(ox))
	mapY := int(math.Floor(oy))

	if mapX < 0 ||
		mapX >= mapWidth ||
		mapY < 0 ||
		mapY >= mapHeight {
		return 0
	}

	const epsilon = 0.0000001

	var deltaX, deltaY float64

	if math.Abs(dx) < epsilon {
		deltaX = math.Inf(1)
	} else {
		deltaX = math.Abs(1.0 / dx)
	}

	if math.Abs(dy) < epsilon {
		deltaY = math.Inf(1)
	} else {
		deltaY = math.Abs(1.0 / dy)
	}

	var stepX, stepY int
	var sideX, sideY float64

	if dx < 0 {
		stepX = -1
		sideX = (ox - float64(mapX)) * deltaX
	} else {
		stepX = 1
		sideX = (float64(mapX) + 1.0 - ox) * deltaX
	}

	if dy < 0 {
		stepY = -1
		sideY = (oy - float64(mapY)) * deltaY
	} else {
		stepY = 1
		sideY = (float64(mapY) + 1.0 - oy) * deltaY
	}

	for {
		var distance float64

		if sideX < sideY {
			sideX += deltaX
			mapX += stepX

			distance = sideX - deltaX
		} else {
			sideY += deltaY
			mapY += stepY

			distance = sideY - deltaY
		}

		if distance >= maxDistance {
			return maxDistance
		}

		if mapX < 0 ||
			mapX >= mapWidth ||
			mapY < 0 ||
			mapY >= mapHeight {
			return maxDistance
		}

		if world[mapY][mapX] != 0 {
			return distance
		}
	}
}

// ------------------------------------------------------------
// SHADOW TEST
// ------------------------------------------------------------

func visibleFromLight(
	x, y float64,
	light Light,
) bool {
	dx := light.X - x
	dy := light.Y - y

	distance := math.Hypot(dx, dy)

	if distance < 0.0001 {
		return true
	}

	dx /= distance
	dy /= distance

	hitDistance := castRay(
		x,
		y,
		dx,
		dy,
		distance,
	)

	// Маленький epsilon нужен, чтобы собственная клетка
	// источника/объекта не считалась стеной.
	return hitDistance >= distance-0.001
}

// ------------------------------------------------------------
// LIGHT ATTENUATION
// ------------------------------------------------------------

func pointLight(
	x, y float64,
	light Light,
) float64 {
	dx := light.X - x
	dy := light.Y - y

	distance := math.Hypot(dx, dy)

	if distance >= light.Radius {
		return 0
	}

	// 0..1
	t := 1.0 - distance/light.Radius

	// Smooth falloff.
	t = t * t * (3.0 - 2.0*t)

	return t * light.Intensity
}

// ------------------------------------------------------------
// LIGHT
// ------------------------------------------------------------

func calculateLight(
	x, y float64,
	lights []Light,
) float64 {
	result := ambientLight

	for _, light := range lights {
		dx := light.X - x
		dy := light.Y - y

		distance := math.Hypot(dx, dy)

		if distance >= light.Radius {
			continue
		}

		// Есть ли прямая видимость источника?
		if !visibleFromLight(x, y, light) {
			continue
		}

		value := pointLight(x, y, light)

		if value > result {
			result = value
		}
	}

	if result > 1 {
		result = 1
	}

	return result
}

// ------------------------------------------------------------
// PALETTE
// ------------------------------------------------------------
//
// Освещение НЕ создаёт новый RGB.
//
// Мы получаем уровень:
//
// 0 1 2 3 4 5
//
// и выбираем ближайший цвет из фиксированной палитры.
// ------------------------------------------------------------

func paletteColor(
	base rl.Color,
	light float64,
) rl.Color {
	level := int(math.Round(
		light * float64(lightLevels-1),
	))

	if level < 0 {
		level = 0
	}

	if level >= lightLevels {
		level = lightLevels - 1
	}

	// Вычисляем целевой RGB только для поиска.
	// Сам итоговый цвет всё равно берётся ИЗ palette.
	targetR := float64(base.R) * float64(level) / float64(lightLevels-1)
	targetG := float64(base.G) * float64(level) / float64(lightLevels-1)
	targetB := float64(base.B) * float64(level) / float64(lightLevels-1)

	best := palette[0]
	bestDistance := math.Inf(1)

	for _, c := range palette {
		dr := float64(c.R) - targetR
		dg := float64(c.G) - targetG
		db := float64(c.B) - targetB

		distance := dr*dr + dg*dg + db*db

		if distance < bestDistance {
			bestDistance = distance
			best = c
		}
	}

	return best
}

// ------------------------------------------------------------
// WORLD RENDER
// ------------------------------------------------------------

func drawWorld(
	lights []Light,
) {
	for y := 0; y < mapHeight; y++ {
		for x := 0; x < mapWidth; x++ {

			var base rl.Color

			if world[y][x] == 0 {
				// Пол.
				base = palette[5]
			} else {
				// Стена.
				base = palette[7]
			}

			// Сэмплируем ЦЕНТР клетки.
			//
			// Поэтому свет может находиться в любой
			// float-позиции, а не только в координатах tile.
			wx := float64(x) + 0.5
			wy := float64(y) + 0.5

			light := calculateLight(
				wx,
				wy,
				lights,
			)

			color := paletteColor(
				base,
				light,
			)

			rl.DrawRectangle(
				int32(x*tileSize),
				int32(y*tileSize),
				tileSize,
				tileSize,
				color,
			)
		}
	}
}

// ------------------------------------------------------------
// MAIN
// ------------------------------------------------------------

func main() {
	rl.InitWindow(
		screenWidth,
		screenHeight,
		"raylib-go - 2D lighting",
	)

	defer rl.CloseWindow()

	rl.SetTargetFPS(60)

	buildWorld()

	lights := []Light{
		{
			X:         10.5,
			Y:         5.5,
			Radius:    14,
			Intensity: 1,
		},
		{
			X:         35.5,
			Y:         18.5,
			Radius:    12,
			Intensity: 1,
		},
		{
			X:         50.5,
			Y:         34.5,
			Radius:    14,
			Intensity: 1,
		},
	}

	// --------------------------------------------------------
	// CAMERA
	// --------------------------------------------------------

	camera := rl.Camera2D{
		Offset: rl.Vector2{
			X: float32(screenWidth) / 2,
			Y: float32(screenHeight) / 2,
		},

		Target: rl.Vector2{
			X: float32(screenWidth) / 2,
			Y: float32(screenHeight) / 2,
		},

		Rotation: 0,
		Zoom:     1,
	}

	for !rl.WindowShouldClose() {

		// ------------------------------------------------------
		// MOVE FIRST LIGHT WITH MOUSE
		// ------------------------------------------------------

		mouse := rl.GetMousePosition()

		lights[0].X =
			float64(mouse.X) / tileSize

		lights[0].Y =
			float64(mouse.Y) / tileSize

		// ------------------------------------------------------
		// ZOOM
		// ------------------------------------------------------

		wheel := rl.GetMouseWheelMove()

		if wheel != 0 {
			camera.Zoom += wheel

			if camera.Zoom < 1 {
				camera.Zoom = 1
			}

			if camera.Zoom > 8 {
				camera.Zoom = 8
			}
		}

		// ------------------------------------------------------
		// DRAW
		// ------------------------------------------------------

		rl.BeginDrawing()

		rl.ClearBackground(
			palette[0],
		)

		rl.BeginMode2D(camera)

		drawWorld(lights)

		// Источники.
		for _, light := range lights {
			rl.DrawCircle(
				int32(light.X*tileSize),
				int32(light.Y*tileSize),
				5,
				palette[9],
			)
		}

		rl.EndMode2D()

		rl.DrawText(
			"Mouse = light",
			10,
			10,
			20,
			palette[9],
		)

		rl.DrawText(
			"Wheel = zoom",
			10,
			35,
			20,
			palette[8],
		)

		rl.EndDrawing()
	}
}
