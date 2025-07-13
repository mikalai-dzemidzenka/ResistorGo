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

	// Инициализация игрока и врагов как акторов
	player := gamepkg.NewActor(rl.NewVector2(200, float32(sceneHeight)-90*gamepkg.SpriteScale), gamepkg.PlayerActor)
	enemies := []gamepkg.Actor{} // Массив врагов

	// Инициализация редактора уровней
	editor := gamepkg.NewLevelEditor()
	editor.LoadLevel("level.json") // Загружаем уровень при старте
	fmt.Printf("Loaded level: %d blocks, %d enemies\n", len(editor.Level.Blocks), len(editor.Level.Enemies))

	state := GameMode // Начальный режим — игра

	// Если стартуем сразу в GameMode — инициализируем врагов из редактора
	if state == GameMode {
		if len(editor.Level.Enemies) > 0 {
			enemies = make([]gamepkg.Actor, len(editor.Level.Enemies))
			for i, enemyData := range editor.Level.Enemies {
				enemies[i] = gamepkg.NewActor(enemyData.Position, enemyData.Type)
				// Сброс состояния врага
				enemies[i].Health = 100
				enemies[i].IsAttacking = false
				enemies[i].IsBackAttacking = false
				enemies[i].IsHit = false
				enemies[i].HasPendingAttack = false
				enemies[i].AttackTimer = 0
				enemies[i].BackAttackTimer = 0
				enemies[i].HitTimer = 0
				enemies[i].LastAttackTime = 0
				enemies[i].AttackCooldown = 0
			}
		}
	}

	// Инициализация камеры (без зума, всегда 1.0)
	camera := rl.Camera2D{
		Offset:   rl.NewVector2(float32(sceneWidth)/2, float32(sceneHeight)/2),
		Target:   rl.NewVector2(player.Position.X+player.Size.X/2, player.Position.Y+player.Size.Y/2),
		Rotation: 0,
		Zoom:     1.0,
	}

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
			player = gamepkg.NewActor(rl.NewVector2(200, float32(sceneHeight)-90*gamepkg.SpriteScale), gamepkg.PlayerActor)
			// Загружаем врагов из редактора или создаем одного по умолчанию
			if len(editor.Level.Enemies) > 0 {
				enemies = make([]gamepkg.Actor, len(editor.Level.Enemies))
				for i, enemyData := range editor.Level.Enemies {
					enemies[i] = gamepkg.NewActor(enemyData.Position, enemyData.Type)
					// Сбрасываем состояние врага
					enemies[i].Health = 100
					enemies[i].IsAttacking = false
					enemies[i].IsBackAttacking = false
					enemies[i].IsHit = false
					enemies[i].HasPendingAttack = false
					enemies[i].IsPendingAttack2 = false
					enemies[i].IsCurrentAttack2 = false
					enemies[i].AttackTimer = 0
					enemies[i].BackAttackTimer = 0
					enemies[i].HitTimer = 0
					enemies[i].LastAttackTime = 0
					enemies[i].AttackCooldown = 0
				}
			} else {
				// Создаем одного врага по умолчанию
				enemies = []gamepkg.Actor{
					gamepkg.NewActor(rl.NewVector2(600, float32(sceneHeight)-90*gamepkg.SpriteScale), gamepkg.AIActor),
				}
			}
		}
		if rl.IsKeyPressed(rl.KeyTwo) {
			if state == GameMode {
				// Конвертируем врагов в EnemyData для сохранения
				enemyDataList := make([]gamepkg.EnemyData, len(enemies))
				for i, enemy := range enemies {
					enemyDataList[i] = gamepkg.EnemyData{
						Position: enemy.Position,
						Type:     enemy.Type,
					}
				}
				// Сохраняем всех врагов в редакторе
				editor.Level = gamepkg.Level{
					Blocks:  editor.Level.Blocks,
					Enemies: enemyDataList,
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

			// Находим ближайшего врага для игрока
			var nearestEnemy *gamepkg.Actor
			if len(enemies) > 0 {
				nearestEnemy = &enemies[0]
				minDist := float32(999999)
				for i := range enemies {
					dist := rl.Vector2Distance(player.Position, enemies[i].Position)
					if dist < minDist {
						minDist = dist
						nearestEnemy = &enemies[i]
					}
				}
			}

			// Обновляем игрока с ближайшим врагом
			gamepkg.UpdateActor(&player, nearestEnemy, &camera, sceneMousePos)
			gamepkg.UpdateActorAnimation(&player)

			// Обновляем всех врагов
			for i := range enemies {
				gamepkg.UpdateActor(&enemies[i], &player, &camera, sceneMousePos)
				gamepkg.UpdateActorAnimation(&enemies[i])
			}

			// Проверяем атаки игрока против всех врагов
			for i := range enemies {
				gamepkg.CheckActorAttack(&player, &enemies[i])
			}

			// Проверяем атаки всех врагов против игрока
			for i := range enemies {
				gamepkg.CheckActorAttack(&enemies[i], &player)
			}
		} else {
			editor.Update(&camera, scale, posX, posY)
		}

		// --- Рисуем всю сцену в renderTex ---
		rl.BeginTextureMode(renderTex)
		rl.ClearBackground(rl.RayWhite)
		rl.BeginMode2D(camera)
		if state == GameMode {
			for _, block := range editor.Level.Blocks {
				rl.DrawRectangleV(block.Position, block.Size, rl.Gray)
			}

			// Находим ближайшего врага для отрисовки
			var nearestEnemy *gamepkg.Actor
			if len(enemies) > 0 {
				nearestEnemy = &enemies[0]
				minDist := float32(999999)
				for i := range enemies {
					dist := rl.Vector2Distance(player.Position, enemies[i].Position)
					if dist < minDist {
						minDist = dist
						nearestEnemy = &enemies[i]
					}
				}
			}

			// Рисуем игрока
			gamepkg.DrawActor(&player, nearestEnemy, rl.Red, &camera)

			// Рисуем всех врагов
			for i := range enemies {
				gamepkg.DrawActor(&enemies[i], nil, rl.Blue, &camera)
			}

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
			// Рисуем только сцену в редакторе (без UI)
			editor.DrawScene(&camera)
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
			rl.DrawText(fmt.Sprintf("Enemies: %d", len(enemies)), windowWidth-150, 10, 20, rl.Black)
			rl.DrawText(fmt.Sprintf("Scale: x%.0f (Press +/-)", scale), 10, 40, 20, rl.Black)
			rl.DrawText(fmt.Sprintf("Camera: %.0f, %.0f", camera.Target.X, camera.Target.Y), 10, 70, 20, rl.Black)
			rl.DrawText(fmt.Sprintf("Raw Mouse: %.0f, %.0f", mousePos.X, mousePos.Y), 10, 100, 20, rl.Black)
			rl.DrawText(fmt.Sprintf("Scene Mouse: %.0f, %.0f", sceneMousePos.X, sceneMousePos.Y), 10, 130, 20, rl.Black)
			rl.DrawText(fmt.Sprintf("Pos: %.0f, %.0f", posX, posY), 10, 160, 20, rl.Black)
			rl.DrawText(fmt.Sprintf("Player Attack: %v, Pending: %v, Attack2: %v", player.IsAttacking, player.HasPendingAttack, player.IsCurrentAttack2), 10, 190, 20, rl.Black)
			if len(enemies) > 0 {
				rl.DrawText(fmt.Sprintf("Enemy Attack: %v, Pending: %v, Attack2: %v", enemies[0].IsAttacking, enemies[0].HasPendingAttack, enemies[0].IsCurrentAttack2), 10, 220, 20, rl.Black)
			}
			rl.DrawText("Game Mode: Press 2 for Editor", 10, windowHeight-30, 20, rl.Black)
		} else {
			// UI редактора (рисуем поверх масштабированного изображения)
			editor.DrawUI(scale, posX, posY)
			rl.DrawText(fmt.Sprintf("Scale: x%.0f (Press +/-)", scale), 10, 40, 20, rl.Black)
			rl.DrawText(fmt.Sprintf("Editor Enemies: %d", len(editor.Level.Enemies)), 10, 70, 20, rl.Black)
			rl.DrawText("Editor Mode: Press 1 for Game", 10, windowHeight-30, 20, rl.Black)
		}
		rl.EndDrawing()
	}

	// Сохранение уровня при выходе
	fmt.Printf("Exiting game, state: %v\n", state)
	if state == GameMode {
		fmt.Printf("Game mode exit: %d enemies to save\n", len(enemies))
		// Конвертируем врагов в EnemyData для сохранения
		enemyDataList := make([]gamepkg.EnemyData, len(enemies))
		for i, enemy := range enemies {
			enemyDataList[i] = gamepkg.EnemyData{
				Position: enemy.Position,
				Type:     enemy.Type,
			}
		}
		// Сохраняем всех врагов из игрового режима в редакторе
		editor.Level = gamepkg.Level{
			Blocks:  editor.Level.Blocks,
			Enemies: enemyDataList,
		}
	} else {
		fmt.Printf("Editor mode exit: %d enemies in editor\n", len(editor.Level.Enemies))
	}
	// Всегда сохраняем уровень при выходе (независимо от режима)
	editor.SaveLevel("level.json")

	gamepkg.UnloadActorTextures(&player)
	// Освобождаем текстуры врагов из игры
	for i := range enemies {
		gamepkg.UnloadActorTextures(&enemies[i])
	}
	// Освобождаем текстуры врагов из редактора (если есть DisplayEnemies)
	for i := range editor.DisplayEnemies {
		gamepkg.UnloadActorTextures(&editor.DisplayEnemies[i])
	}
	rl.CloseWindow()
}
