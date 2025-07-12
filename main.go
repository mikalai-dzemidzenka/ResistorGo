package main

import (
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
	"resistor/gamepkg"
)

// Lerp для линейной интерполяции
func Lerp(start, end, t float32) float32 {
	return start + t*(end-start)
}

func main() {
	rl.InitWindow(gamepkg.ScreenWidth, gamepkg.ScreenHeight, "Fist Fight with Animations")
	rl.SetTargetFPS(60)
	rand.Seed(time.Now().UnixNano())

	// Инициализация игрока и ИИ
	player := gamepkg.NewPlayer(rl.NewVector2(100, gamepkg.ScreenHeight-90*gamepkg.SpriteScale), true)
	ai := gamepkg.NewPlayer(rl.NewVector2(gamepkg.ScreenWidth-140, gamepkg.ScreenHeight-90*gamepkg.SpriteScale), false)
	aiAttackInterval := float32(1.5)

	// Инициализация камеры
	camera := rl.Camera2D{
		Offset:   rl.NewVector2(float32(gamepkg.ScreenWidth)/2, float32(gamepkg.ScreenHeight)/2),
		Target:   rl.NewVector2(player.Position.X+player.Size.X/2, player.Position.Y+player.Size.Y/2), // Следуем за игроком
		Rotation: 0,
		Zoom:     2.0, // Зум по умолчанию x2
	}
	targetZoom := float32(2.0) // Целевой масштаб для интерполяции

	for !rl.WindowShouldClose() {
		// Обновляем позицию камеры, чтобы следовать за игроком
		camera.Target = rl.NewVector2(player.Position.X+player.Size.X/2, player.Position.Y+player.Size.Y/2)

		// Обработка масштабирования камеры
		wheelMove := rl.GetMouseWheelMove()
		if wheelMove > 0 {
			targetZoom += gamepkg.ZoomSpeed // Колесо вверх: увеличение масштаба
		} else if wheelMove < 0 {
			targetZoom -= gamepkg.ZoomSpeed // Колесо вниз: уменьшение масштаба
		}
		// Ограничение масштаба
		targetZoom = float32(math.Max(float64(gamepkg.MinZoom), math.Min(float64(gamepkg.MaxZoom), float64(targetZoom))))
		// Плавное изменение масштаба
		camera.Zoom = Lerp(camera.Zoom, targetZoom, 0.1)

		gamepkg.UpdatePlayer(&player, &ai, true, &camera)
		gamepkg.UpdateAI(&ai, &player, aiAttackInterval)

		gamepkg.UpdateAnimation(&player)
		gamepkg.UpdateAnimation(&ai)

		gamepkg.CheckAttack(&player, &ai, true)
		gamepkg.CheckAttack(&ai, &player, false)

		rl.BeginDrawing()
		rl.ClearBackground(rl.RayWhite)

		rl.BeginMode2D(camera)
		gamepkg.DrawPlayer(&player, &ai, rl.Red, true, &camera)
		gamepkg.DrawPlayer(&ai, nil, rl.Blue, false, &camera)
		rl.EndMode2D()

		// Отрисовка HUD
		rl.DrawText("Player HP: "+strconv.Itoa(player.Health), 10, 10, 20, rl.Black)
		rl.DrawText("AI HP: "+strconv.Itoa(ai.Health), gamepkg.ScreenWidth-150, 10, 20, rl.Black)
		rl.DrawText(fmt.Sprintf("Zoom: x%.2f", camera.Zoom), 10, 40, 20, rl.Black)

		rl.EndDrawing()
	}

	gamepkg.UnloadTextures([]*gamepkg.Player{&player, &ai})
	rl.CloseWindow()
}
