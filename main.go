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

	// Инициализация системы перемотки времени (5 секунд истории при 60 FPS)
	timeRewind := gamepkg.NewTimeRewindSystem(300) // 5 секунд * 60 FPS
	enemyRewinds := []*gamepkg.TimeRewindSystem{}
	if len(enemies) > 0 {
		enemyRewinds = make([]*gamepkg.TimeRewindSystem, len(enemies))
		for i := range enemies {
			enemyRewinds[i] = gamepkg.NewTimeRewindSystem(300)
		}
	}

	// Инициализация шейдера перемотки времени
	timeRewindShader := gamepkg.NewTimeRewindShader()
	defer timeRewindShader.Unload()

	// Переменная для времени (используется в шейдере)
	shaderTime := float32(0.0)

	// Инициализация редактора уровней
	editor := gamepkg.NewLevelEditor()
	// Загружаем уровни при старте (автоматически создаются демо-уровни если нет файлов)
	fmt.Printf("Loaded %d levels\n", len(editor.LevelManager.Levels))

	state := GameMode // Начальный режим — игра

	// Если стартуем сразу в GameMode — инициализируем врагов из редактора
	if state == GameMode {
		currentLevel := editor.LevelManager.GetCurrentLevel()
		if currentLevel != nil && len(currentLevel.Enemies) > 0 {
			enemies = make([]gamepkg.Actor, len(currentLevel.Enemies))
			for i, enemyData := range currentLevel.Enemies {
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
			// Инициализация rewind для врагов
			enemyRewinds = make([]*gamepkg.TimeRewindSystem, len(enemies))
			for i := range enemies {
				enemyRewinds[i] = gamepkg.NewTimeRewindSystem(300)
			}
		} else {
			// Создаем нескольких врагов по умолчанию для тестирования позиционирования
			enemies = []gamepkg.Actor{
				gamepkg.NewActor(rl.NewVector2(600, float32(sceneHeight)-90*gamepkg.SpriteScale), gamepkg.AIActor),
				gamepkg.NewActor(rl.NewVector2(700, float32(sceneHeight)-90*gamepkg.SpriteScale), gamepkg.AIActor),
				gamepkg.NewActor(rl.NewVector2(500, float32(sceneHeight)-90*gamepkg.SpriteScale), gamepkg.AIActor),
			}
			// Инициализация rewind для врагов по умолчанию
			enemyRewinds = make([]*gamepkg.TimeRewindSystem, len(enemies))
			for i := range enemies {
				enemyRewinds[i] = gamepkg.NewTimeRewindSystem(300)
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

	// Переменные для плавного зума
	targetScale := float32(2.0)  // Целевой зум
	currentScale := float32(2.0) // Текущий зум (для плавного перехода)
	zoomSpeed := float32(0.15)   // Скорость интерполяции зума

	// Переменные для рендеринга
	var scaledWidth, scaledHeight, posX, posY float32

	for !rl.WindowShouldClose() {
		// Обновляем время для шейдера
		shaderTime += rl.GetFrameTime()

		// Переключение режимов
		if rl.IsKeyPressed(rl.KeyOne) {
			if state == EditorMode {
				editor.SaveCurrentLevel()
			}
			state = GameMode
			player = gamepkg.NewActor(rl.NewVector2(200, float32(sceneHeight)-90*gamepkg.SpriteScale), gamepkg.PlayerActor)
			timeRewind.ClearHistory() // Очищаем историю при переключении в игровой режим
			// Загружаем врагов из редактора или создаем одного по умолчанию
			currentLevel := editor.LevelManager.GetCurrentLevel()
			if currentLevel != nil && len(currentLevel.Enemies) > 0 {
				enemies = make([]gamepkg.Actor, len(currentLevel.Enemies))
				for i, enemyData := range currentLevel.Enemies {
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
				// Инициализация rewind для врагов
				enemyRewinds = make([]*gamepkg.TimeRewindSystem, len(enemies))
				for i := range enemies {
					enemyRewinds[i] = gamepkg.NewTimeRewindSystem(300)
				}
			} else {
				// Создаем нескольких врагов по умолчанию для тестирования позиционирования
				enemies = []gamepkg.Actor{
					gamepkg.NewActor(rl.NewVector2(600, float32(sceneHeight)-90*gamepkg.SpriteScale), gamepkg.AIActor),
					gamepkg.NewActor(rl.NewVector2(700, float32(sceneHeight)-90*gamepkg.SpriteScale), gamepkg.AIActor),
					gamepkg.NewActor(rl.NewVector2(500, float32(sceneHeight)-90*gamepkg.SpriteScale), gamepkg.AIActor),
				}
				// Инициализация rewind для врагов по умолчанию
				enemyRewinds = make([]*gamepkg.TimeRewindSystem, len(enemies))
				for i := range enemies {
					enemyRewinds[i] = gamepkg.NewTimeRewindSystem(300)
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
				currentLevel := editor.LevelManager.GetCurrentLevel()
				if currentLevel != nil {
					currentLevel.Enemies = enemyDataList
					editor.SaveCurrentLevel()
				}
			}
			state = EditorMode
		}

		// Изменение плавного зума колесиком мыши
		mouseWheel := rl.GetMouseWheelMove()
		if mouseWheel != 0 {
			// Изменяем целевой зум на основе движения колесика
			zoomChange := float32(mouseWheel) * 0.2 // 0.2x за одно движение колесика
			targetScale += zoomChange

			// Ограничиваем зум в разумных пределах
			if targetScale < 0.5 {
				targetScale = 0.5
			} else if targetScale > 8.0 {
				targetScale = 8.0
			}
		}

		// Плавная интерполяция зума
		currentScale = Lerp(currentScale, targetScale, zoomSpeed)

		// Вычисляем размер масштабированного изображения и позицию для центрирования
		scaledWidth = float32(sceneWidth) * currentScale
		scaledHeight = float32(sceneHeight) * currentScale
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
			// Обработка перемотки времени
			if rl.IsKeyPressed(rl.KeyR) {
				if timeRewind.IsRewinding {
					// Если уже перематываем - останавливаем
					timeRewind.StopRewind()
					for _, er := range enemyRewinds {
						er.StopRewind()
					}
					fmt.Printf("STOPPED REWIND: States %d, Current %d\n", len(timeRewind.States), timeRewind.CurrentIndex)
				} else {
					// Если не перематываем - начинаем
					timeRewind.StartRewind()
					for _, er := range enemyRewinds {
						er.StartRewind()
					}
					fmt.Printf("STARTED REWIND: States %d, Current %d\n", len(timeRewind.States), timeRewind.CurrentIndex)
				}
			}

			// Если перематываем время - управляем направлением по положению мыши на экране
			if timeRewind.IsRewinding {
				// Делим экран на 3 части по Y
				if mousePos.Y < float32(windowHeight)/3 {
					t := 1.0 - (mousePos.Y / (float32(windowHeight) / 3))
					speed := 1.0 + 2.0*t // от 1x до 3x
					timeRewind.RewindSpeed = speed
					for _, er := range enemyRewinds {
						er.RewindSpeed = speed
					}
				} else if mousePos.Y > 2*float32(windowHeight)/3 {
					t := (mousePos.Y - 2*float32(windowHeight)/3) / (float32(windowHeight) / 3)
					speed := -1.0 - 2.0*t // от -1x до -3x
					timeRewind.RewindSpeed = speed
					for _, er := range enemyRewinds {
						er.RewindSpeed = speed
					}
				} else {
					// Средняя треть — пауза
					timeRewind.RewindSpeed = 0.0
					for _, er := range enemyRewinds {
						er.RewindSpeed = 0.0
					}
				}
			}

			// Обновляем перемотку времени
			timeRewind.UpdateRewind(&player)

			// Обновляем перемотку врагов только если игрок в режиме перемотки
			if timeRewind.IsRewinding {
				for i := range enemies {
					enemyRewinds[i].UpdateRewind(&enemies[i])
				}
			}

			// Сохраняем состояние игрока (только если не перематываем)
			if !timeRewind.IsRewinding {
				if timeRewind.CurrentIndex < len(timeRewind.States)-1 {
					timeRewind.CutHistoryToCurrent()
				}
				timeRewind.SaveState(&player)
				for i := range enemies {
					if enemyRewinds[i].CurrentIndex < len(enemyRewinds[i].States)-1 {
						enemyRewinds[i].CutHistoryToCurrent()
					}
					enemyRewinds[i].SaveState(&enemies[i])
				}
			}

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

			// Обновляем игрока и врагов только если не перематываем время
			if !timeRewind.IsRewinding {
				gamepkg.UpdateActor(&player, nearestEnemy, &camera, sceneMousePos)
				gamepkg.UpdateActorAnimation(&player)

				// Обновляем всех врагов с новой системой ИИ
				for i := range enemies {
					// Используем новую систему позиционирования для врагов
					gamepkg.UpdateEnemyAI(&enemies[i], &player, enemies, i)
					gamepkg.UpdateActorAnimation(&enemies[i])
					// Обновляем хитбоксы врагов
					gamepkg.UpdateActorHitboxes(&enemies[i])
				}

				// Проверяем атаки игрока против всех врагов
				for i := range enemies {
					gamepkg.CheckActorAttack(&player, &enemies[i])
				}

				// Проверяем атаки всех врагов против игрока
				for i := range enemies {
					gamepkg.CheckActorAttack(&enemies[i], &player)
				}

				// Перераспределяем очередь врагов, если нужно
				gamepkg.ReorderEnemyQueue(enemies)
			} else {
				// Если перематываем время - только обновляем анимации
				gamepkg.UpdateActorAnimation(&player)
				for i := range enemies {
					gamepkg.UpdateActorAnimation(&enemies[i])
				}
			}
		} else {
			editor.Update(&camera, scale, posX, posY)
		}

		// --- Рисуем всю сцену в renderTex ---
		rl.BeginTextureMode(renderTex)
		rl.ClearBackground(rl.RayWhite)

		// Применяем шейдер перемотки времени, если активна перемотка
		if timeRewind.IsRewinding && timeRewindShader != nil {
			direction := float32(0.0)
			if timeRewind.RewindSpeed > 0 {
				direction = 1.0
			} else if timeRewind.RewindSpeed < 0 {
				direction = -1.0
			}
			timeRewindShader.BeginTimeRewindMode(shaderTime, timeRewind.RewindSpeed, direction)
		}

		rl.BeginMode2D(camera)
		if state == GameMode {
			currentLevel := editor.LevelManager.GetCurrentLevel()
			if currentLevel != nil {
				// Отрисовка блоков с параллаксом и разными цветами
				for _, block := range currentLevel.Blocks {
					var color rl.Color
					var parallaxOffset float32 = 0

					switch block.Type {
					case gamepkg.NormalBlock:
						color = rl.Gray
					case gamepkg.GrassBlock:
						color = rl.Green
						parallaxOffset = camera.Target.X * 0.1 // Медленный параллакс
					case gamepkg.StoneBlock:
						color = rl.DarkGray
						parallaxOffset = camera.Target.X * 0.05 // Очень медленный параллакс
					case gamepkg.WoodBlock:
						color = rl.Brown
						parallaxOffset = camera.Target.X * 0.15 // Средний параллакс
					case gamepkg.MetalBlock:
						color = rl.LightGray
					case gamepkg.BackgroundBlock:
						color = rl.Fade(rl.Blue, 0.3)          // Полупрозрачный
						parallaxOffset = camera.Target.X * 0.3 // Быстрый параллакс
					}

					// Применяем параллакс
					parallaxPos := rl.NewVector2(block.Position.X+parallaxOffset, block.Position.Y)
					rl.DrawRectangleV(parallaxPos, block.Size, color)
				}
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

			// Y-sort: собираем всех акторов и сортируем по Y позиции
			type ActorDrawInfo struct {
				Actor    *gamepkg.Actor
				Color    rl.Color
				IsPlayer bool
			}

			var actorsToDraw []ActorDrawInfo

			// Добавляем игрока
			playerColor := rl.Red
			if timeRewind.IsRewinding {
				// Эффект перемотки времени - игрок становится более прозрачным и меняет цвет
				alpha := uint8(128 + 127*timeRewind.GetRewindProgress()) // От 128 до 255
				playerColor = rl.NewColor(255, 0, 0, alpha)
			}
			actorsToDraw = append(actorsToDraw, ActorDrawInfo{
				Actor:    &player,
				Color:    playerColor,
				IsPlayer: true,
			})

			// Добавляем врагов
			for i := range enemies {
				actorsToDraw = append(actorsToDraw, ActorDrawInfo{
					Actor:    &enemies[i],
					Color:    rl.Blue,
					IsPlayer: false,
				})
			}

			// Сортируем по Y позиции (больший Y рисуется поверх)
			for i := 0; i < len(actorsToDraw)-1; i++ {
				for j := i + 1; j < len(actorsToDraw); j++ {
					if actorsToDraw[i].Actor.Position.Y > actorsToDraw[j].Actor.Position.Y {
						actorsToDraw[i], actorsToDraw[j] = actorsToDraw[j], actorsToDraw[i]
					}
				}
			}

			// Рисуем всех акторов в правильном порядке
			for _, drawInfo := range actorsToDraw {
				if drawInfo.IsPlayer {
					gamepkg.DrawActor(drawInfo.Actor, nearestEnemy, drawInfo.Color, &camera)
				} else {
					gamepkg.DrawActor(drawInfo.Actor, nil, drawInfo.Color, &camera)
				}
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

		// Завершаем шейдер перемотки времени
		if timeRewind.IsRewinding && timeRewindShader != nil {
			timeRewindShader.EndTimeRewindMode()
		}

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
			rl.DrawText(fmt.Sprintf("Zoom: %.1fx (Mouse Wheel)", currentScale), 10, 40, 20, rl.Black)
			rl.DrawText(fmt.Sprintf("Camera: %.0f, %.0f", camera.Target.X, camera.Target.Y), 10, 70, 20, rl.Black)
			rl.DrawText(fmt.Sprintf("Raw Mouse: %.0f, %.0f", mousePos.X, mousePos.Y), 10, 100, 20, rl.Black)
			rl.DrawText(fmt.Sprintf("Scene Mouse: %.0f, %.0f", sceneMousePos.X, sceneMousePos.Y), 10, 130, 20, rl.Black)
			mouseDelta := rl.GetMouseDelta()
			rl.DrawText(fmt.Sprintf("Mouse Delta: %.1f, %.1f", mouseDelta.X, mouseDelta.Y), 10, 280, 20, rl.Black)
			rl.DrawText(fmt.Sprintf("Pos: %.0f, %.0f", posX, posY), 10, 160, 20, rl.Black)
			rl.DrawText(fmt.Sprintf("Player Attack: %v, Pending: %v, Attack2: %v", player.IsAttacking, player.HasPendingAttack, player.IsCurrentAttack2), 10, 190, 20, rl.Black)
			if len(enemies) > 0 {
				rl.DrawText(fmt.Sprintf("Enemy Attack: %v, Pending: %v, Attack2: %v", enemies[0].IsAttacking, enemies[0].HasPendingAttack, enemies[0].IsCurrentAttack2), 10, 220, 20, rl.Black)
			}
			rl.DrawText(fmt.Sprintf("Rewind: %v, Speed: %.1f, States: %d/%d", timeRewind.IsRewinding, timeRewind.RewindSpeed, timeRewind.CurrentIndex, len(timeRewind.States)), 10, 250, 20, rl.Black)
			rl.DrawText(fmt.Sprintf("Can Rewind: %v, Timer: %.3f", timeRewind.CanRewind(), timeRewind.RewindTimer), 10, 310, 20, rl.Black)
			rl.DrawText(fmt.Sprintf("Player Transition: %v, From: %d, To: %d", player.IsTransitioning, player.TransitionFrom, player.TransitionTo), 10, 340, 20, rl.Black)
			rl.DrawText(fmt.Sprintf("Player TargetPart: %d", player.TargetPart), 10, 370, 20, rl.Black)

			// Информация о текущем уровне
			currentLevel := editor.LevelManager.GetCurrentLevel()
			if currentLevel != nil {
				rl.DrawText(fmt.Sprintf("Level: %s (%d/%d)", currentLevel.Name, editor.LevelManager.CurrentLevel+1, len(editor.LevelManager.Levels)), 10, 400, 20, rl.Black)
			}

			// Индикатор перемотки времени
			if timeRewind.IsRewinding {
				var speedText string
				if timeRewind.RewindSpeed > 0 {
					speedText = "TIME FORWARD"
					if timeRewind.RewindSpeed > 1.0 {
						speedText = fmt.Sprintf("TIME FORWARD (%dx)", int(timeRewind.RewindSpeed))
					}
				} else if timeRewind.RewindSpeed < 0 {
					speedText = "TIME BACKWARD"
					if timeRewind.RewindSpeed < -1.0 {
						speedText = fmt.Sprintf("TIME BACKWARD (%dx)", int(-timeRewind.RewindSpeed))
					}
				} else {
					speedText = "TIME PAUSED"
				}
				rl.DrawText(speedText, windowWidth-200, 10, 20, rl.Red)

				progress := timeRewind.GetRewindProgress()
				barWidth := float32(150)
				barHeight := float32(10)
				barX := float32(windowWidth) - barWidth - 25
				barY := float32(35)

				// Фон индикатора
				rl.DrawRectangle(int32(barX), int32(barY), int32(barWidth), int32(barHeight), rl.Gray)
				// Прогресс перемотки
				rl.DrawRectangle(int32(barX), int32(barY), int32(barWidth*progress), int32(barHeight), rl.Red)
				// Рамка
				rl.DrawRectangleLines(int32(barX), int32(barY), int32(barWidth), int32(barHeight), rl.Black)

				rl.DrawText(fmt.Sprintf("Progress: %.1f%%", progress*100), windowWidth-200, 50, 15, rl.Black)
				rl.DrawText("Mouse Y: Forward/Backward", windowWidth-200, 70, 12, rl.DarkGray)
				rl.DrawText("Hold Shift for 3x speed", windowWidth-200, 85, 12, rl.DarkGray)
			} else if timeRewind.CanRewind() {
				rl.DrawText("Press R to Pause Time", windowWidth-200, 10, 15, rl.Blue)
				rl.DrawText("Mouse Y: Forward/Backward", windowWidth-200, 30, 12, rl.DarkGray)
				rl.DrawText("Hold Shift for 3x speed", windowWidth-200, 45, 12, rl.DarkGray)
			}

			rl.DrawText("Game Mode: Press 2 for Editor", 10, windowHeight-30, 20, rl.Black)
		} else {
			// UI редактора (рисуем поверх масштабированного изображения)
			editor.DrawUI(currentScale, posX, posY)
			rl.DrawText(fmt.Sprintf("Zoom: %.1fx (Mouse Wheel)", currentScale), 10, 40, 20, rl.Black)
			rl.DrawText(fmt.Sprintf("Editor Enemies: %d", len(editor.LevelManager.GetCurrentLevel().Enemies)), 10, 70, 20, rl.Black)
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
		currentLevel := editor.LevelManager.GetCurrentLevel()
		if currentLevel != nil {
			currentLevel.Enemies = enemyDataList
			editor.SaveCurrentLevel()
		}
	} else {
		fmt.Printf("Editor mode exit: %d enemies in editor\n", len(editor.LevelManager.GetCurrentLevel().Enemies))
	}
	// Всегда сохраняем уровень при выходе (независимо от режима)
	editor.SaveCurrentLevel()

	// Очищаем ресурсы редактора
	editor.Cleanup()

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
