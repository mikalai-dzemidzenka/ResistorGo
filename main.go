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

type GameState int

const (
	GameMode GameState = iota
	EditorMode
)

func Lerp(start, end, t float32) float32 {
	return start + t*(end-start)
}

func main() {
	rl.InitWindow(gamepkg.ScreenWidth, gamepkg.ScreenHeight, "Fist Fight with Level Editor")
	rl.SetTargetFPS(60)
	rand.Seed(time.Now().UnixNano())

	// Инициализация игрока и ИИ
	player := gamepkg.NewPlayer(rl.NewVector2(200, gamepkg.ScreenHeight-90*gamepkg.SpriteScale), true)
	ai := gamepkg.NewPlayer(rl.NewVector2(600, gamepkg.ScreenHeight-90*gamepkg.SpriteScale), false)
	aiAttackInterval := float32(1.5)

	// Инициализация редактора уровней
	editor := gamepkg.NewLevelEditor()
	editor.LoadLevel("level.json") // Загружаем уровень при старте

	// Инициализация камеры
	camera := rl.Camera2D{
		Offset:   rl.NewVector2(float32(gamepkg.ScreenWidth)/2, float32(gamepkg.ScreenHeight)/2),
		Target:   rl.NewVector2(player.Position.X+player.Size.X/2, player.Position.Y+player.Size.Y/2),
		Rotation: 0,
		Zoom:     2.0,
	}
	targetZoom := float32(2.0)

	state := GameMode // Начальный режим — игра

	for !rl.WindowShouldClose() {
		// Переключение режимов
		if rl.IsKeyPressed(rl.KeyOne) {
			if state == EditorMode {
				editor.SaveLevel("level.json") // Сохраняем уровень при переключении из редактора
			}
			state = GameMode
			// Сбрасываем позиции игрока и ИИ
			player = gamepkg.NewPlayer(rl.NewVector2(200, gamepkg.ScreenHeight-90*gamepkg.SpriteScale), true)
			ai = gamepkg.NewPlayer(rl.NewVector2(600, gamepkg.ScreenHeight-90*gamepkg.SpriteScale), false)
		}
		if rl.IsKeyPressed(rl.KeyTwo) {
			if state == GameMode {
				// Сохраняем текущий уровень (игрок + ИИ как враги)
				editor.Level = gamepkg.Level{
					Blocks:  editor.Level.Blocks,  // Сохраняем существующие блоки
					Enemies: []gamepkg.Player{ai}, // Сохраняем ИИ как врага
				}
				editor.SaveLevel("level.json")
			}
			state = EditorMode
		}

		// Обработка масштабирования камеры
		wheelMove := rl.GetMouseWheelMove()
		if wheelMove > 0 {
			targetZoom += gamepkg.ZoomSpeed
		} else if wheelMove < 0 {
			targetZoom -= gamepkg.ZoomSpeed
		}
		targetZoom = float32(math.Max(float64(gamepkg.MinZoom), math.Min(float64(gamepkg.MaxZoom), float64(targetZoom))))
		camera.Zoom = Lerp(camera.Zoom, targetZoom, 0.1)

		// Обновление в зависимости от режима
		if state == GameMode {
			// Обновляем позицию камеры
			camera.Target = rl.NewVector2(player.Position.X+player.Size.X/2, player.Position.Y+player.Size.Y/2)

			gamepkg.UpdatePlayer(&player, &ai, true, &camera)
			gamepkg.UpdateAI(&ai, &player, aiAttackInterval)

			gamepkg.UpdateAnimation(&player)
			gamepkg.UpdateAnimation(&ai)

			gamepkg.CheckAttack(&player, &ai, true)
			gamepkg.CheckAttack(&ai, &player, false)
		} else {
			// В редакторе камера статична (или следует за последней позицией игрока)
			editor.Update(&camera)
		}

		rl.BeginDrawing()
		rl.ClearBackground(rl.RayWhite)

		rl.BeginMode2D(camera)
		if state == GameMode {
			// Отрисовка блоков из редактора
			for _, block := range editor.Level.Blocks {
				rl.DrawRectangleV(block.Position, block.Size, rl.Gray)
			}
			// Отрисовка игрока и ИИ
			gamepkg.DrawPlayer(&player, &ai, rl.Red, true, &camera)
			gamepkg.DrawPlayer(&ai, nil, rl.Blue, false, &camera)
		} else {
			editor.Draw(&camera)
		}
		rl.EndMode2D()

		// Отрисовка HUD
		if state == GameMode {
			rl.DrawText("Player HP: "+strconv.Itoa(player.Health), 10, 10, 20, rl.Black)
			rl.DrawText("AI HP: "+strconv.Itoa(ai.Health), gamepkg.ScreenWidth-150, 10, 20, rl.Black)
			rl.DrawText(fmt.Sprintf("Zoom: x%.2f", camera.Zoom), 10, 40, 20, rl.Black)
			rl.DrawText("Game Mode: Press 2 for Editor", 10, gamepkg.ScreenHeight-30, 20, rl.Black)
		} else {
			rl.DrawText(fmt.Sprintf("Zoom: x%.2f", camera.Zoom), 10, 40, 20, rl.Black)
			rl.DrawText("Editor Mode: Press 1 for Game", 10, gamepkg.ScreenHeight-30, 20, rl.Black)
		}

		rl.EndDrawing()
	}

	// Сохранение уровня при выходе
	if state == GameMode {
		editor.Level = gamepkg.Level{
			Blocks:  editor.Level.Blocks,
			Enemies: []gamepkg.Player{ai},
		}
	}
	editor.SaveLevel("level.json")

	gamepkg.UnloadTextures([]*gamepkg.Player{&player, &ai})
	for i := range editor.Level.Enemies {
		gamepkg.UnloadTextures([]*gamepkg.Player{&editor.Level.Enemies[i]})
	}
	rl.CloseWindow()
}
