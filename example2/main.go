package main

import (
	"math"
	"math/rand"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	screenWidth  = 960
	screenHeight = 720
	pixelScale   = 4 // масштаб пиксель-арта: мир 240x180, растянут до окна
)

func main() {
	rl.InitWindow(screenWidth, screenHeight, "Animal Well style 2D lighting — Go + Raylib")
	defer rl.CloseWindow()
	rl.SetTargetFPS(60)

	// Низкое разрешение «игровой камеры»
	gameW := int32(screenWidth / pixelScale)
	gameH := int32(screenHeight / pixelScale)

	// Две offscreen-текстуры: мир и маска света
	worldTex := rl.LoadRenderTexture(gameW, gameH)
	lightTex := rl.LoadRenderTexture(gameW, gameH)
	defer rl.UnloadRenderTexture(worldTex)
	defer rl.UnloadRenderTexture(lightTex)

	// --- Игровые объекты ---
	player := rl.Vector2{X: float32(gameW) / 2, Y: float32(gameH) / 2}

	// Статичные блоки (стены/платформы)
	walls := []rl.Rectangle{
		{X: 20, Y: 40, Width: 60, Height: 20},
		{X: 100, Y: 80, Width: 40, Height: 40},
		{X: 160, Y: 120, Width: 50, Height: 15},
		{X: 40, Y: 130, Width: 30, Height: 30},
	}

	// Статичные источники света (факелы)
	torches := []struct {
		pos rl.Vector2
		col rl.Color
	}{
		{pos: rl.Vector2{X: 50, Y: 35}, col: rl.Color{R: 255, G: 140, B: 40, A: 255}},    // оранжевый
		{pos: rl.Vector2{X: 185, Y: 115}, col: rl.Color{R: 255, G: 200, B: 120, A: 255}}, // тёплый
	}

	// Случайные декоративные камни/кирпичи
	rand.Seed(42)
	stones := make([]rl.Vector2, 30)
	for i := range stones {
		stones[i] = rl.Vector2{
			X: float32(rand.Int31n(gameW-16) + 8),
			Y: float32(rand.Int31n(gameH-16) + 8),
		}
	}

	for !rl.WindowShouldClose() {
		dt := rl.GetFrameTime()

		// --- Управление ---
		spd := float32(60) * dt
		if rl.IsKeyDown(rl.KeyW) || rl.IsKeyDown(rl.KeyUp) {
			player.Y -= spd
		}
		if rl.IsKeyDown(rl.KeyS) || rl.IsKeyDown(rl.KeyDown) {
			player.Y += spd
		}
		if rl.IsKeyDown(rl.KeyA) || rl.IsKeyDown(rl.KeyLeft) {
			player.X -= spd
		}
		if rl.IsKeyDown(rl.KeyD) || rl.IsKeyDown(rl.KeyRight) {
			player.X += spd
		}

		// --- 1. Рендер мира ---
		rl.BeginTextureMode(worldTex)
		rl.ClearBackground(rl.Color{R: 20, G: 20, B: 28, A: 255}) // тёмный фон

		// Пиксель-арт «плитка» пола
		for x := int32(0); x < gameW; x += 16 {
			for y := int32(0); y < gameH; y += 16 {
				c := rl.Color{R: 25, G: 25, B: 35, A: 255}
				if (x/16+y/16)%2 == 0 {
					c = rl.Color{R: 28, G: 28, B: 38, A: 255}
				}
				rl.DrawRectangle(x, y, 16, 16, c)
			}
		}

		// Камни
		for _, s := range stones {
			rl.DrawRectangle(int32(s.X-3), int32(s.Y-2), 6, 4, rl.Color{R: 60, G: 55, B: 50, A: 255})
		}

		// Стены
		for _, w := range walls {
			rl.DrawRectangleRec(w, rl.Color{R: 70, G: 65, B: 60, A: 255})
			rl.DrawRectangleLinesEx(w, 1, rl.Color{R: 45, G: 42, B: 38, A: 255})
		}

		// Факелы (спрайты)
		for _, t := range torches {
			rl.DrawRectangle(int32(t.pos.X-1), int32(t.pos.Y-6), 2, 6, rl.Color{R: 90, G: 60, B: 40, A: 255}) // древко
			rl.DrawCircle(int32(t.pos.X), int32(t.pos.Y-8), 2, t.col)                                         // пламя
		}

		// Игрок
		rl.DrawRectangle(int32(player.X-3), int32(player.Y-7), 6, 14, rl.Color{R: 100, G: 180, B: 255, A: 255})
		rl.DrawCircle(int32(player.X), int32(player.Y-10), 3, rl.Color{R: 100, G: 180, B: 255, A: 255})

		rl.EndTextureMode()

		// --- 2. Рендер lightmap (маски света) ---
		rl.BeginTextureMode(lightTex)
		rl.ClearBackground(rl.Black) // по умолчанию полная темнота

		// Слабое глобальное освещение (луна/ambient), чтобы силуэты стен едва проглядывали
		rl.DrawRectangle(0, 0, gameW, gameH, rl.Color{R: 40, G: 40, B: 55, A: 255})

		// Свет от факелов с мерцанием
		for _, t := range torches {
			flicker := float32(math.Sin(float64(rl.GetTime())*8+float64(t.pos.X))) * 3
			radius := float32(22 + flicker)
			drawSoftLight(int32(t.pos.X), int32(t.pos.Y-8), radius, t.col)
		}

		// Фонарик игрока (более белый/жёлтый)
		drawSoftLight(int32(player.X), int32(player.Y-6), 35, rl.Color{R: 255, G: 250, B: 220, A: 255})

		rl.EndTextureMode()

		// --- 3. Финальный композитинг ---
		rl.BeginDrawing()
		rl.ClearBackground(rl.Black)

		// UV-рект для текстур (Height отрицательный, т.к. RenderTexture перевернута по Y)
		src := rl.Rectangle{
			X: 0, Y: 0,
			Width:  float32(worldTex.Texture.Width),
			Height: -float32(worldTex.Texture.Height),
		}
		dst := rl.Rectangle{
			X: 0, Y: 0,
			Width:  screenWidth,
			Height: screenHeight,
		}

		// Сначала мир
		rl.DrawTexturePro(worldTex.Texture, src, dst, rl.Vector2{}, 0, rl.White)

		// Поверх — lightmap в режиме умножения (Multiply)
		rl.SetBlendMode(rl.BlendMultiplied)
		rl.DrawTexturePro(lightTex.Texture, src, dst, rl.Vector2{}, 0, rl.White)
		rl.SetBlendMode(rl.BlendAlpha) // вернуть обычный блендинг

		// UI
		rl.DrawText("WASD / Arrows — move", 12, 12, 20, rl.Color{R: 200, G: 200, B: 200, A: 255})
		rl.DrawText("Animal Well style lighting", 12, 36, 20, rl.Color{R: 180, G: 180, B: 180, A: 255})

		rl.EndDrawing()
	}
}

// drawSoftLight рисует мягкий круг света из нескольких градиентных кругов
func drawSoftLight(cx, cy int32, radius float32, color rl.Color) {
	// Яркое ядро
	rl.DrawCircle(cx, cy, radius*0.4, color)

	// Полутона
	rl.DrawCircle(cx, cy, radius*0.7, rl.Color{
		R: color.R, G: color.G, B: color.B, A: 140,
	})

	// Мягкие края
	rl.DrawCircle(cx, cy, radius, rl.Color{
		R: color.R, G: color.G, B: color.B, A: 60,
	})
}
