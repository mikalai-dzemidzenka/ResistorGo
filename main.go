package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"resistor/gamepkg"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type GameStateEnum int

const (
	GameMode GameStateEnum = iota
	EditorMode
)

type GameState struct {
	// Ресурсы
	TextureManager   *gamepkg.TextureManager
	RenderTex        rl.RenderTexture2D
	TimeRewindShader *gamepkg.TimeRewindShader

	// Игровые объекты
	Player       gamepkg.Actor
	Enemies      []gamepkg.Actor
	TimeRewind   *gamepkg.TimeRewindSystem
	EnemyRewinds []*gamepkg.TimeRewindSystem

	// Редактор
	Editor gamepkg.LevelEditor

	// Камера и рендер
	Camera       rl.Camera2D
	Scale        float32
	TargetScale  float32
	CurrentScale float32
	ZoomSpeed    float32
	ShaderTime   float32

	// Размеры
	SceneWidth   int32
	SceneHeight  int32
	WindowWidth  int32
	WindowHeight int32

	// Переменные для рендера
	ScaledWidth  float32
	ScaledHeight float32
	PosX         float32
	PosY         float32

	// Состояние
	State GameStateEnum

	// Переменные для обработки мыши
	MousePos      rl.Vector2
	SceneMousePos rl.Vector2
}

func NewGameState() *GameState {
	sceneWidth := int32(gamepkg.ScreenWidth)
	sceneHeight := int32(gamepkg.ScreenHeight)
	windowWidth := int32(1280)
	windowHeight := int32(720)

	ts := &GameState{
		TextureManager: gamepkg.NewTextureManager(),
		SceneWidth:     sceneWidth,
		SceneHeight:    sceneHeight,
		WindowWidth:    windowWidth,
		WindowHeight:   windowHeight,
		Scale:          2.0,
		TargetScale:    5.0,
		CurrentScale:   5.0,
		ZoomSpeed:      0.15,
		State:          GameMode,
	}

	ts.TimeRewind = gamepkg.NewTimeRewindSystem(300)
	ts.EnemyRewinds = []*gamepkg.TimeRewindSystem{}
	ts.Editor = gamepkg.NewLevelEditor()
	ts.Player = gamepkg.NewActor(rl.NewVector2(200, float32(sceneHeight)-90*gamepkg.SpriteScale), gamepkg.PlayerActor, ts.TextureManager)
	ts.Enemies = []gamepkg.Actor{}
	ts.Camera = rl.Camera2D{
		Offset:   rl.NewVector2(float32(sceneWidth)/2, float32(sceneHeight)/2),
		Target:   rl.NewVector2(ts.Player.Position.X+ts.Player.Size.X/2, ts.Player.Position.Y+ts.Player.Size.Y/2),
		Rotation: 0,
		Zoom:     1.0,
	}

	return ts
}

func Lerp(start, end, t float32) float32 {
	return start + t*(end-start)
}

func (gs *GameState) Update() {
	gs.ShaderTime += rl.GetFrameTime()

	// Переключение режимов
	if rl.IsKeyPressed(rl.KeyOne) {
		if gs.State == EditorMode {
			gs.Editor.SaveCurrentLevel()
		}
		gs.State = GameMode
		gs.Player = gamepkg.NewActor(rl.NewVector2(200, float32(gs.SceneHeight)-90*gamepkg.SpriteScale), gamepkg.PlayerActor, gs.TextureManager)
		gs.TimeRewind.ClearHistory()
		currentLevel := gs.Editor.LevelManager.GetCurrentLevel()
		if currentLevel != nil && len(currentLevel.Enemies) > 0 {
			gs.Enemies = make([]gamepkg.Actor, len(currentLevel.Enemies))
			for i, enemyData := range currentLevel.Enemies {
				gs.Enemies[i] = gamepkg.NewActor(enemyData.Position, enemyData.Type, gs.TextureManager)
				gs.Enemies[i].Health = 100
				gs.Enemies[i].IsAttacking = false
				gs.Enemies[i].IsBackAttacking = false
				gs.Enemies[i].IsHit = false
				gs.Enemies[i].HasPendingAttack = false
				gs.Enemies[i].AttackTimer = 0
				gs.Enemies[i].BackAttackTimer = 0
				gs.Enemies[i].HitTimer = 0
				gs.Enemies[i].LastAttackTime = 0
				gs.Enemies[i].AttackCooldown = 0
			}
			gs.EnemyRewinds = make([]*gamepkg.TimeRewindSystem, len(gs.Enemies))
			for i := range gs.Enemies {
				gs.EnemyRewinds[i] = gamepkg.NewTimeRewindSystem(300)
			}
		} else {
			gs.Enemies = []gamepkg.Actor{
				gamepkg.NewActor(rl.NewVector2(600, float32(gs.SceneHeight)-90*gamepkg.SpriteScale), gamepkg.AIActor, gs.TextureManager),
				gamepkg.NewActor(rl.NewVector2(700, float32(gs.SceneHeight)-90*gamepkg.SpriteScale), gamepkg.AIActor, gs.TextureManager),
				gamepkg.NewActor(rl.NewVector2(500, float32(gs.SceneHeight)-90*gamepkg.SpriteScale), gamepkg.AIActor, gs.TextureManager),
			}
			gs.EnemyRewinds = make([]*gamepkg.TimeRewindSystem, len(gs.Enemies))
			for i := range gs.Enemies {
				gs.EnemyRewinds[i] = gamepkg.NewTimeRewindSystem(300)
			}
		}
	}
	if rl.IsKeyPressed(rl.KeyTwo) {
		if gs.State == GameMode {
			enemyDataList := make([]gamepkg.EnemyData, len(gs.Enemies))
			for i, enemy := range gs.Enemies {
				enemyDataList[i] = gamepkg.EnemyData{
					Position: enemy.Position,
					Type:     enemy.Type,
				}
			}
			currentLevel := gs.Editor.LevelManager.GetCurrentLevel()
			if currentLevel != nil {
				currentLevel.Enemies = enemyDataList
				gs.Editor.SaveCurrentLevel()
			}
		}
		gs.State = EditorMode
	}

	mouseWheel := rl.GetMouseWheelMove()
	if mouseWheel != 0 {
		zoomChange := float32(mouseWheel) * 0.2
		gs.TargetScale += zoomChange
		if gs.TargetScale < 0.5 {
			gs.TargetScale = 0.5
		} else if gs.TargetScale > 8.0 {
			gs.TargetScale = 8.0
		}
	}
	gs.CurrentScale = Lerp(gs.CurrentScale, gs.TargetScale, gs.ZoomSpeed)
	gs.ScaledWidth = float32(gs.SceneWidth) * gs.CurrentScale
	gs.ScaledHeight = float32(gs.SceneHeight) * gs.CurrentScale
	gs.PosX = (float32(gs.WindowWidth) - gs.ScaledWidth) / 2
	gs.PosY = (float32(gs.WindowHeight) - gs.ScaledHeight) / 2

	gs.MousePos = rl.GetMousePosition()
	if gs.MousePos.X >= gs.PosX && gs.MousePos.X <= gs.PosX+gs.ScaledWidth &&
		gs.MousePos.Y >= gs.PosY && gs.MousePos.Y <= gs.PosY+gs.ScaledHeight {
		normalizedX := (gs.MousePos.X - gs.PosX) / gs.ScaledWidth
		normalizedY := (gs.MousePos.Y - gs.PosY) / gs.ScaledHeight
		sceneX := normalizedX * float32(gs.SceneWidth)
		sceneY := normalizedY * float32(gs.SceneHeight)
		gs.SceneMousePos = rl.GetScreenToWorld2D(rl.NewVector2(sceneX, sceneY), gs.Camera)
	} else {
		gs.SceneMousePos = rl.NewVector2(-1000, -1000)
	}

	// Обработка перемотки времени
	if rl.IsKeyPressed(rl.KeyR) {
		if gs.TimeRewind.IsRewinding {
			// Если уже перематываем - останавливаем
			gs.TimeRewind.StopRewind()
			for _, er := range gs.EnemyRewinds {
				er.StopRewind()
			}
			fmt.Printf("STOPPED REWIND: States %d, Current %d\n", len(gs.TimeRewind.States), gs.TimeRewind.CurrentIndex)
		} else {
			// Если не перематываем - начинаем
			gs.TimeRewind.StartRewind()
			for _, er := range gs.EnemyRewinds {
				er.StartRewind()
			}
			fmt.Printf("STARTED REWIND: States %d, Current %d\n", len(gs.TimeRewind.States), gs.TimeRewind.CurrentIndex)
		}
	}

	// Если перематываем время - управляем направлением по положению мыши на экране
	if gs.TimeRewind.IsRewinding {
		// Делим экран на 3 части по Y
		if gs.MousePos.Y < float32(gs.WindowHeight)/3 {
			t := 1.0 - (gs.MousePos.Y / (float32(gs.WindowHeight) / 3))
			speed := 1.0 + 2.0*t // от 1x до 3x
			gs.TimeRewind.RewindSpeed = speed
			for _, er := range gs.EnemyRewinds {
				er.RewindSpeed = speed
			}
		} else if gs.MousePos.Y > 2*float32(gs.WindowHeight)/3 {
			t := (gs.MousePos.Y - 2*float32(gs.WindowHeight)/3) / (float32(gs.WindowHeight) / 3)
			speed := -1.0 - 2.0*t // от -1x до -3x
			gs.TimeRewind.RewindSpeed = speed
			for _, er := range gs.EnemyRewinds {
				er.RewindSpeed = speed
			}
		} else {
			// Средняя треть — пауза
			gs.TimeRewind.RewindSpeed = 0.0
			for _, er := range gs.EnemyRewinds {
				er.RewindSpeed = 0.0
			}
		}
	}

	// Обновляем перемотку времени
	gs.TimeRewind.UpdateRewind(&gs.Player)

	// Обновляем перемотку врагов только если игрок в режиме перемотки
	if gs.TimeRewind.IsRewinding {
		for i := range gs.Enemies {
			gs.EnemyRewinds[i].UpdateRewind(&gs.Enemies[i])
		}
	}

	// Сохраняем состояние игрока (только если не перематываем)
	if !gs.TimeRewind.IsRewinding {
		if gs.TimeRewind.CurrentIndex < len(gs.TimeRewind.States)-1 {
			gs.TimeRewind.CutHistoryToCurrent()
		}
		gs.TimeRewind.SaveState(&gs.Player)
		for i := range gs.Enemies {
			if gs.EnemyRewinds[i].CurrentIndex < len(gs.EnemyRewinds[i].States)-1 {
				gs.EnemyRewinds[i].CutHistoryToCurrent()
			}
			gs.EnemyRewinds[i].SaveState(&gs.Enemies[i])
		}
	}

	// Камера плавно следует за игроком
	targetX := gs.Player.Position.X + gs.Player.Size.X/2
	targetY := gs.Player.Position.Y + gs.Player.Size.Y/2

	// Плавное следование камеры
	gs.Camera.Target.X = Lerp(gs.Camera.Target.X, targetX, 0.1)
	gs.Camera.Target.Y = Lerp(gs.Camera.Target.Y, targetY, 0.1)

	// Находим ближайшего врага для игрока
	var nearestEnemy *gamepkg.Actor
	if len(gs.Enemies) > 0 {
		nearestEnemy = &gs.Enemies[0]
		minDist := float32(999999)
		for i := range gs.Enemies {
			dist := rl.Vector2Distance(gs.Player.Position, gs.Enemies[i].Position)
			if dist < minDist {
				minDist = dist
				nearestEnemy = &gs.Enemies[i]
			}
		}
	}

	// Обновляем игрока и врагов только если не перематываем время
	if !gs.TimeRewind.IsRewinding {
		gamepkg.UpdateActor(&gs.Player, nearestEnemy, &gs.Camera, gs.SceneMousePos)
		gamepkg.UpdateActorAnimation(&gs.Player)

		// Обновляем всех врагов с новой системой ИИ
		for i := range gs.Enemies {
			// Используем новую систему позиционирования для врагов
			gamepkg.UpdateEnemyAI(&gs.Enemies[i], &gs.Player, gs.Enemies, i)
			gamepkg.UpdateActorAnimation(&gs.Enemies[i])
			// Обновляем хитбоксы врагов
			gamepkg.UpdateActorHitboxes(&gs.Enemies[i])
		}

		// Проверяем атаки игрока против всех врагов
		for i := range gs.Enemies {
			gamepkg.CheckActorAttack(&gs.Player, &gs.Enemies[i])
		}

		// Проверяем атаки всех врагов против игрока
		for i := range gs.Enemies {
			gamepkg.CheckActorAttack(&gs.Enemies[i], &gs.Player)
		}

		// Перераспределяем очередь врагов, если нужно
		gamepkg.ReorderEnemyQueue(gs.Enemies)
	} else {
		// Если перематываем время - только обновляем анимации
		gamepkg.UpdateActorAnimation(&gs.Player)
		for i := range gs.Enemies {
			gamepkg.UpdateActorAnimation(&gs.Enemies[i])
		}
	}
}

func (gs *GameState) Render() {
	rl.BeginTextureMode(gs.RenderTex)
	rl.ClearBackground(rl.RayWhite)

	// Применяем шейдер перемотки времени, если активна перемотка
	if gs.TimeRewind.IsRewinding && gs.TimeRewindShader != nil {
		direction := float32(0.0)
		if gs.TimeRewind.RewindSpeed > 0 {
			direction = 1.0
		} else if gs.TimeRewind.RewindSpeed < 0 {
			direction = -1.0
		}
		gs.TimeRewindShader.BeginTimeRewindMode(gs.ShaderTime, gs.TimeRewind.RewindSpeed, direction)
	}

	rl.BeginMode2D(gs.Camera)
	if gs.State == GameMode {
		currentLevel := gs.Editor.LevelManager.GetCurrentLevel()
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
					parallaxOffset = gs.Camera.Target.X * 0.1 // Медленный параллакс
				case gamepkg.StoneBlock:
					color = rl.DarkGray
					parallaxOffset = gs.Camera.Target.X * 0.05 // Очень медленный параллакс
				case gamepkg.WoodBlock:
					color = rl.Brown
					parallaxOffset = gs.Camera.Target.X * 0.15 // Средний параллакс
				case gamepkg.MetalBlock:
					color = rl.LightGray
				case gamepkg.BackgroundBlock:
					color = rl.Fade(rl.Blue, 0.3)             // Полупрозрачный
					parallaxOffset = gs.Camera.Target.X * 0.3 // Быстрый параллакс
				}

				// Применяем параллакс
				parallaxPos := rl.NewVector2(block.Position.X+parallaxOffset, block.Position.Y)
				rl.DrawRectangleV(parallaxPos, block.Size, color)
			}
		}

		// Находим ближайшего врага для отрисовки
		var nearestEnemy *gamepkg.Actor
		if len(gs.Enemies) > 0 {
			nearestEnemy = &gs.Enemies[0]
			minDist := float32(999999)
			for i := range gs.Enemies {
				dist := rl.Vector2Distance(gs.Player.Position, gs.Enemies[i].Position)
				if dist < minDist {
					minDist = dist
					nearestEnemy = &gs.Enemies[i]
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
		if gs.TimeRewind.IsRewinding {
			// Эффект перемотки времени - игрок становится более прозрачным и меняет цвет
			alpha := uint8(128 + 127*gs.TimeRewind.GetRewindProgress()) // От 128 до 255
			playerColor = rl.NewColor(255, 0, 0, alpha)
		}
		actorsToDraw = append(actorsToDraw, ActorDrawInfo{
			Actor:    &gs.Player,
			Color:    playerColor,
			IsPlayer: true,
		})

		// Добавляем врагов
		for i := range gs.Enemies {
			actorsToDraw = append(actorsToDraw, ActorDrawInfo{
				Actor:    &gs.Enemies[i],
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
				gamepkg.DrawActor(drawInfo.Actor, nearestEnemy, drawInfo.Color, &gs.Camera)
			} else {
				gamepkg.DrawActor(drawInfo.Actor, nil, drawInfo.Color, &gs.Camera)
			}
		}

		// Рисуем красный квадратик на месте курсора в сцене
		if gs.SceneMousePos.X >= 0 && gs.SceneMousePos.Y >= 0 {
			cursorSize := float32(8)
			rl.DrawRectangleV(
				rl.NewVector2(gs.SceneMousePos.X-cursorSize/2, gs.SceneMousePos.Y-cursorSize/2),
				rl.NewVector2(cursorSize, cursorSize),
				rl.Red,
			)
		}
	} else {
		// Рисуем только сцену в редакторе (без UI)
		gs.Editor.DrawScene(&gs.Camera)
	}
	rl.EndMode2D()

	// Завершаем шейдер перемотки времени
	if gs.TimeRewind.IsRewinding && gs.TimeRewindShader != nil {
		gs.TimeRewindShader.EndTimeRewindMode()
	}

	// HUD рисуем после масштабирования!
	rl.EndTextureMode()

	rl.BeginDrawing()
	rl.ClearBackground(rl.RayWhite)

	rl.DrawTexturePro(
		gs.RenderTex.Texture,
		rl.NewRectangle(0, 0, float32(gs.SceneWidth), -float32(gs.SceneHeight)), // -height для flip по Y
		rl.NewRectangle(gs.PosX, gs.PosY, gs.ScaledWidth, gs.ScaledHeight),
		rl.NewVector2(0, 0),
		0,
		rl.White,
	)

	// HUD (рисуем поверх, чтобы не масштабировался)
	if gs.State == GameMode {
		rl.DrawText("Player HP: "+strconv.Itoa(gs.Player.Health), 10, 10, 20, rl.Black)
		rl.DrawText(fmt.Sprintf("Enemies: %d", len(gs.Enemies)), gs.WindowWidth-150, 10, 20, rl.Black)
		rl.DrawText(fmt.Sprintf("Zoom: %.1fx (Mouse Wheel)", gs.CurrentScale), 10, 40, 20, rl.Black)
		rl.DrawText(fmt.Sprintf("Camera: %.0f, %.0f", gs.Camera.Target.X, gs.Camera.Target.Y), 10, 70, 20, rl.Black)
		rl.DrawText(fmt.Sprintf("Raw Mouse: %.0f, %.0f", gs.MousePos.X, gs.MousePos.Y), 10, 100, 20, rl.Black)
		rl.DrawText(fmt.Sprintf("Scene Mouse: %.0f, %.0f", gs.SceneMousePos.X, gs.SceneMousePos.Y), 10, 130, 20, rl.Black)
		mouseDelta := rl.GetMouseDelta()
		rl.DrawText(fmt.Sprintf("Mouse Delta: %.1f, %.1f", mouseDelta.X, mouseDelta.Y), 10, 280, 20, rl.Black)
		rl.DrawText(fmt.Sprintf("Pos: %.0f, %.0f", gs.PosX, gs.PosY), 10, 160, 20, rl.Black)
		rl.DrawText(fmt.Sprintf("Player Attack: %v, Pending: %v, Attack2: %v", gs.Player.IsAttacking, gs.Player.HasPendingAttack, gs.Player.IsCurrentAttack2), 10, 190, 20, rl.Black)
		if len(gs.Enemies) > 0 {
			rl.DrawText(fmt.Sprintf("Enemy Attack: %v, Pending: %v, Attack2: %v", gs.Enemies[0].IsAttacking, gs.Enemies[0].HasPendingAttack, gs.Enemies[0].IsCurrentAttack2), 10, 220, 20, rl.Black)
		}
		rl.DrawText(fmt.Sprintf("Rewind: %v, Speed: %.1f, States: %d/%d", gs.TimeRewind.IsRewinding, gs.TimeRewind.RewindSpeed, gs.TimeRewind.CurrentIndex, len(gs.TimeRewind.States)), 10, 250, 20, rl.Black)
		rl.DrawText(fmt.Sprintf("Can Rewind: %v, Timer: %.3f", gs.TimeRewind.CanRewind(), gs.TimeRewind.RewindTimer), 10, 310, 20, rl.Black)
		rl.DrawText(fmt.Sprintf("Player Transition: %v, From: %d, To: %d", gs.Player.IsTransitioning, gs.Player.TransitionFrom, gs.Player.TransitionTo), 10, 340, 20, rl.Black)
		rl.DrawText(fmt.Sprintf("Player TargetPart: %d", gs.Player.TargetPart), 10, 370, 20, rl.Black)

		// Информация о текущем уровне
		currentLevel := gs.Editor.LevelManager.GetCurrentLevel()
		if currentLevel != nil {
			rl.DrawText(fmt.Sprintf("Level: %s (%d/%d)", currentLevel.Name, gs.Editor.LevelManager.CurrentLevel+1, len(gs.Editor.LevelManager.Levels)), 10, 400, 20, rl.Black)
		}

		// Индикатор перемотки времени
		if gs.TimeRewind.IsRewinding {
			var speedText string
			if gs.TimeRewind.RewindSpeed > 0 {
				speedText = "TIME FORWARD"
				if gs.TimeRewind.RewindSpeed > 1.0 {
					speedText = fmt.Sprintf("TIME FORWARD (%dx)", int(gs.TimeRewind.RewindSpeed))
				}
			} else if gs.TimeRewind.RewindSpeed < 0 {
				speedText = "TIME BACKWARD"
				if gs.TimeRewind.RewindSpeed < -1.0 {
					speedText = fmt.Sprintf("TIME BACKWARD (%dx)", int(-gs.TimeRewind.RewindSpeed))
				}
			} else {
				speedText = "TIME PAUSED"
			}
			rl.DrawText(speedText, gs.WindowWidth-200, 10, 20, rl.Red)

			progress := gs.TimeRewind.GetRewindProgress()
			barWidth := float32(150)
			barHeight := float32(10)
			barX := float32(gs.WindowWidth) - barWidth - 25
			barY := float32(35)

			// Фон индикатора
			rl.DrawRectangle(int32(barX), int32(barY), int32(barWidth), int32(barHeight), rl.Gray)
			// Прогресс перемотки
			rl.DrawRectangle(int32(barX), int32(barY), int32(barWidth*progress), int32(barHeight), rl.Red)
			// Рамка
			rl.DrawRectangleLines(int32(barX), int32(barY), int32(barWidth), int32(barHeight), rl.Black)

			rl.DrawText(fmt.Sprintf("Progress: %.1f%%", progress*100), gs.WindowWidth-200, 50, 15, rl.Black)
			rl.DrawText("Mouse Y: Forward/Backward", gs.WindowWidth-200, 70, 12, rl.DarkGray)
			rl.DrawText("Hold Shift for 3x speed", gs.WindowWidth-200, 85, 12, rl.DarkGray)
		} else if gs.TimeRewind.CanRewind() {
			rl.DrawText("Press R to Pause Time", gs.WindowWidth-200, 10, 15, rl.Blue)
			rl.DrawText("Mouse Y: Forward/Backward", gs.WindowWidth-200, 30, 12, rl.DarkGray)
			rl.DrawText("Hold Shift for 3x speed", gs.WindowWidth-200, 45, 12, rl.DarkGray)
		}

		rl.DrawText("Game Mode: Press 2 for Editor", 10, gs.WindowHeight-30, 20, rl.Black)
	} else {
		// UI редактора (рисуем поверх масштабированного изображения)
		gs.Editor.DrawUI(gs.CurrentScale, gs.PosX, gs.PosY)
		rl.DrawText(fmt.Sprintf("Zoom: %.1fx (Mouse Wheel)", gs.CurrentScale), 10, 40, 20, rl.Black)
		rl.DrawText(fmt.Sprintf("Editor Enemies: %d", len(gs.Editor.LevelManager.GetCurrentLevel().Enemies)), 10, 70, 20, rl.Black)
		rl.DrawText("Editor Mode: Press 1 for Game", 10, gs.WindowHeight-30, 20, rl.Black)
	}
	rl.EndDrawing()
}

func (gs *GameState) Cleanup() {
	gs.Editor.Cleanup()
	gamepkg.UnloadActorTextures(&gs.Player)
	for i := range gs.Enemies {
		gamepkg.UnloadActorTextures(&gs.Enemies[i])
	}
	for i := range gs.Editor.DisplayEnemies {
		gamepkg.UnloadActorTextures(&gs.Editor.DisplayEnemies[i])
	}
	gs.TextureManager.UnloadAll()
	rl.UnloadRenderTexture(gs.RenderTex)
	if gs.TimeRewindShader != nil {
		gs.TimeRewindShader.Unload()
	}
}

func main() {
	rl.InitWindow(1280, 720, "Fist Fight with Level Editor")
	rl.SetTargetFPS(60)
	rand.Seed(time.Now().UnixNano())

	gs := NewGameState()
	gs.RenderTex = rl.LoadRenderTexture(gs.SceneWidth, gs.SceneHeight)
	gs.TimeRewindShader = gamepkg.NewTimeRewindShader()

	fmt.Printf("Loaded %d levels\n", len(gs.Editor.LevelManager.Levels))

	for !rl.WindowShouldClose() {
		gs.Update()
		gs.Render()
	}

	// Сохранение уровня при выходе
	fmt.Printf("Exiting game, state: %v\n", gs.State)
	if gs.State == GameMode {
		fmt.Printf("Game mode exit: %d enemies to save\n", len(gs.Enemies))
		enemyDataList := make([]gamepkg.EnemyData, len(gs.Enemies))
		for i, enemy := range gs.Enemies {
			enemyDataList[i] = gamepkg.EnemyData{
				Position: enemy.Position,
				Type:     enemy.Type,
			}
		}
		currentLevel := gs.Editor.LevelManager.GetCurrentLevel()
		if currentLevel != nil {
			currentLevel.Enemies = enemyDataList
			gs.Editor.SaveCurrentLevel()
		}
	} else {
		fmt.Printf("Editor mode exit: %d enemies in editor\n", len(gs.Editor.LevelManager.GetCurrentLevel().Enemies))
	}
	gs.Editor.SaveCurrentLevel()
	gs.Cleanup()
	rl.CloseWindow()
}
