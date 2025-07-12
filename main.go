package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"resistor/gamepkg"

	rl "github.com/gen2brain/raylib-go/raylib"
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
	// Размер сцены в пикселях (оригинальный pixel art размер)
	sceneWidth := int32(gamepkg.ScreenWidth)
	sceneHeight := int32(gamepkg.ScreenHeight)

	// Фиксированный размер окна
	windowWidth := int32(1280)
	windowHeight := int32(720)
	rl.InitWindow(windowWidth, windowHeight, "Fist Fight with Level Editor")
	rl.SetTargetFPS(60)
	rand.Seed(time.Now().UnixNano())

	// Инициализация render texture для pixel-perfect рендера
	renderTex := rl.LoadRenderTexture(sceneWidth, sceneHeight)
	defer rl.UnloadRenderTexture(renderTex)

	// Инициализация игрока и ИИ
	player := gamepkg.NewPlayer(rl.NewVector2(200, float32(sceneHeight)-90*gamepkg.SpriteScale), true)
	ai := gamepkg.NewPlayer(rl.NewVector2(600, float32(sceneHeight)-90*gamepkg.SpriteScale), false)
	aiAttackInterval := float32(1.5)

	// Инициализация редактора уровней
	editor := gamepkg.NewLevelEditor()
	editor.LoadLevel("level.json") // Загружаем уровень при старте

	// Инициализация камеры (без зума, всегда 1.0)
	camera := rl.Camera2D{
		Offset:   rl.NewVector2(float32(sceneWidth)/2, float32(sceneHeight)/2),
		Target:   rl.NewVector2(player.Position.X+player.Size.X/2, player.Position.Y+player.Size.Y/2),
		Rotation: 0,
		Zoom:     1.0,
	}

	state := GameMode // Начальный режим — игра

	// Масштаб вывода (только целые значения)
	scale := float32(2) // Можно сделать настраиваемым (например, 2x, 3x)

	// Переменные для рендеринга
	var scaledWidth, scaledHeight, posX, posY float32

	for !rl.WindowShouldClose() {
		// Переключение режимов
		if rl.IsKeyPressed(rl.KeyOne) {
			if state == EditorMode {
				editor.SaveLevel("level.json")
			}
			state = GameMode
			player = gamepkg.NewPlayer(rl.NewVector2(200, float32(sceneHeight)-90*gamepkg.SpriteScale), true)
			ai = gamepkg.NewPlayer(rl.NewVector2(600, float32(sceneHeight)-90*gamepkg.SpriteScale), false)
		}
		if rl.IsKeyPressed(rl.KeyTwo) {
			if state == GameMode {
				editor.Level = gamepkg.Level{
					Blocks:  editor.Level.Blocks,
					Enemies: []gamepkg.Player{ai},
				}
				editor.SaveLevel("level.json")
			}
			state = EditorMode
		}

		// Изменение integer scale клавишами +/-
		if rl.IsKeyPressed(rl.KeyEqual) || rl.IsKeyPressed(rl.KeyKpAdd) {
			if scale < 6 {
				scale += 1
			}
		}
		if rl.IsKeyPressed(rl.KeyMinus) || rl.IsKeyPressed(rl.KeyKpSubtract) {
			if scale > 1 {
				scale -= 1
			}
		}

		// Вычисляем размер масштабированного изображения и позицию для центрирования
		scaledWidth = float32(sceneWidth) * scale
		scaledHeight = float32(sceneHeight) * scale
		posX = (float32(windowWidth) - scaledWidth) / 2
		posY = (float32(windowHeight) - scaledHeight) / 2

		// Преобразуем координаты мыши с экрана в координаты сцены
		mousePos := rl.GetMousePosition()

		// Используем raylib функции для преобразования координат
		var sceneMousePos rl.Vector2

		// Проверяем, находится ли мышь в пределах масштабированного изображения
		if mousePos.X >= posX && mousePos.X <= posX+scaledWidth &&
			mousePos.Y >= posY && mousePos.Y <= posY+scaledHeight {

			// Нормализуем координаты мыши относительно масштабированного изображения
			normalizedX := (mousePos.X - posX) / scaledWidth
			normalizedY := (mousePos.Y - posY) / scaledHeight

			// Преобразуем в координаты сцены (без учета камеры)
			sceneX := normalizedX * float32(sceneWidth)
			sceneY := normalizedY * float32(sceneHeight)

			// Применяем обратное преобразование камеры
			sceneMousePos = rl.GetScreenToWorld2D(rl.NewVector2(sceneX, sceneY), camera)
		} else {
			sceneMousePos = rl.NewVector2(-1000, -1000) // Вне сцены
		}

		// Обновление в зависимости от режима
		if state == GameMode {
			// Камера плавно следует за игроком
			targetX := player.Position.X + player.Size.X/2
			targetY := player.Position.Y + player.Size.Y/2

			// Плавное следование камеры
			camera.Target.X = Lerp(camera.Target.X, targetX, 0.1)
			camera.Target.Y = Lerp(camera.Target.Y, targetY, 0.1)

			gamepkg.UpdatePlayer(&player, &ai, true, &camera, sceneMousePos)
			gamepkg.UpdateAI(&ai, &player, aiAttackInterval)
			gamepkg.UpdateAnimation(&player)
			gamepkg.UpdateAnimation(&ai)
			gamepkg.CheckAttack(&player, &ai, true)
			gamepkg.CheckAttack(&ai, &player, false)
		} else {
			editor.Update(&camera)
		}

		// --- Рисуем всю сцену в renderTex ---
		rl.BeginTextureMode(renderTex)
		rl.ClearBackground(rl.RayWhite)
		rl.BeginMode2D(camera)
		if state == GameMode {
			for _, block := range editor.Level.Blocks {
				rl.DrawRectangleV(block.Position, block.Size, rl.Gray)
			}
			gamepkg.DrawPlayer(&player, &ai, rl.Red, true, &camera)
			gamepkg.DrawPlayer(&ai, nil, rl.Blue, false, &camera)

			// Рисуем красный квадратик на месте курсора в сцене
			if sceneMousePos.X >= 0 && sceneMousePos.Y >= 0 {
				cursorSize := float32(8)
				rl.DrawRectangleV(
					rl.NewVector2(sceneMousePos.X-cursorSize/2, sceneMousePos.Y-cursorSize/2),
					rl.NewVector2(cursorSize, cursorSize),
					rl.Red,
				)
			}
		} else {
			editor.Draw(&camera)
		}
		rl.EndMode2D()
		// HUD рисуем после масштабирования!
		rl.EndTextureMode()

		// --- Теперь растягиваем renderTex на экран с центрированием ---
		rl.BeginDrawing()
		rl.ClearBackground(rl.RayWhite)

		rl.DrawTexturePro(
			renderTex.Texture,
			rl.NewRectangle(0, 0, float32(sceneWidth), -float32(sceneHeight)), // -height для flip по Y
			rl.NewRectangle(posX, posY, scaledWidth, scaledHeight),
			rl.NewVector2(0, 0),
			0,
			rl.White,
		)

		// HUD (рисуем поверх, чтобы не масштабировался)
		if state == GameMode {

			rl.DrawText("Player HP: "+strconv.Itoa(player.Health), 10, 10, 20, rl.Black)
			rl.DrawText("AI HP: "+strconv.Itoa(ai.Health), windowWidth-150, 10, 20, rl.Black)
			rl.DrawText(fmt.Sprintf("Scale: x%.0f (Press +/-)", scale), 10, 40, 20, rl.Black)
			rl.DrawText(fmt.Sprintf("Camera: %.0f, %.0f", camera.Target.X, camera.Target.Y), 10, 70, 20, rl.Black)
			rl.DrawText(fmt.Sprintf("Raw Mouse: %.0f, %.0f", mousePos.X, mousePos.Y), 10, 100, 20, rl.Black)
			rl.DrawText(fmt.Sprintf("Scene Mouse: %.0f, %.0f", sceneMousePos.X, sceneMousePos.Y), 10, 130, 20, rl.Black)
			rl.DrawText(fmt.Sprintf("Pos: %.0f, %.0f", posX, posY), 10, 160, 20, rl.Black)
			rl.DrawText("Game Mode: Press 2 for Editor", 10, windowHeight-30, 20, rl.Black)
		} else {
			rl.DrawText(fmt.Sprintf("Scale: x%.0f (Press +/-)", scale), 10, 40, 20, rl.Black)
			rl.DrawText("Editor Mode: Press 1 for Game", 10, windowHeight-30, 20, rl.Black)
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
