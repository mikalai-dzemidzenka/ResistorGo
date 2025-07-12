package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

// Player представляет игрока
type Player struct {
	Position   rl.Vector2
	Velocity   rl.Vector2
	Width      float32
	Height     float32
	OnGround   bool
	OnWall     bool
	WallNormal float32 // -1 для левой стены, 1 для правой, 0 если не на стене
}

// Platform представляет платформу
type Platform struct {
	Rect rl.Rectangle
}

// Wall представляет стену
type Wall struct {
	Rect rl.Rectangle
}

// HackableObject представляет взламываемый объект
type HackableObject struct {
	Rect   rl.Rectangle
	Hacked bool
}

// GameState хранит состояние игры для перемотки времени
type GameState struct {
	PlayerPos      rl.Vector2
	PlayerVel      rl.Vector2
	PlayerOnGround bool
	PlayerOnWall   bool
	WallNormal     float32
	Hackables      []bool
}

func main() {
	// Инициализация окна
	rl.InitWindow(800, 600, "Gunpoint Clone")
	rl.SetTargetFPS(60)

	// Создание игрока
	player := Player{
		Position:   rl.NewVector2(100, 500),
		Velocity:   rl.NewVector2(0, 0),
		Width:      20,
		Height:     40,
		OnGround:   false,
		OnWall:     false,
		WallNormal: 0,
	}

	// Создание платформ
	platforms := []Platform{
		{Rect: rl.NewRectangle(0, 550, 800, 50)},   // Пол
		{Rect: rl.NewRectangle(200, 400, 200, 20)}, // Платформа 1
		{Rect: rl.NewRectangle(500, 300, 150, 20)}, // Платформа 2
	}

	// Создание стен
	walls := []Wall{
		{Rect: rl.NewRectangle(400, 300, 20, 250)}, // Вертикальная стена 1
		{Rect: rl.NewRectangle(100, 300, 20, 250)}, // Вертикальная стена 2
	}

	// Создание взламываемых объектов
	hackables := []HackableObject{
		{Rect: rl.NewRectangle(250, 380, 20, 20), Hacked: false},
		{Rect: rl.NewRectangle(550, 280, 20, 20), Hacked: false},
	}

	// Константы физики
	const gravity = 800.0
	const moveSpeed = 200.0
	const jumpForce = -400.0
	const wallSlideSpeed = 100.0
	const wallJumpForceX = 300.0
	const wallJumpForceY = -350.0
	const wallJumpLockoutDuration = 0.05 // Задержка в секундах после прыжка от стены

	// Перемотка времени
	var stateHistory []GameState
	var isRewinding bool
	var currentIndex int = -1
	var wallJumpLockoutTimer float32 = 0 // Таймер блокировки движения после прыжка от стены

	// Инициализация начального состояния
	initialState := GameState{
		PlayerPos:      player.Position,
		PlayerVel:      player.Velocity,
		PlayerOnGround: player.OnGround,
		PlayerOnWall:   player.OnWall,
		WallNormal:     player.WallNormal,
		Hackables:      make([]bool, len(hackables)),
	}
	for i := range hackables {
		initialState.Hackables[i] = hackables[i].Hacked
	}
	stateHistory = append(stateHistory, initialState)

	// Основной игровой цикл
	for !rl.WindowShouldClose() {
		dt := rl.GetFrameTime()

		// Обновление таймера блокировки
		if wallJumpLockoutTimer > 0 {
			wallJumpLockoutTimer -= dt
			if wallJumpLockoutTimer < 0 {
				wallJumpLockoutTimer = 0
			}
		}

		if isRewinding {
			if currentIndex > 0 {
				currentIndex--
				setGameState(&player, hackables, stateHistory[currentIndex])
				wallJumpLockoutTimer = 0 // Сброс таймера при перемотке
			} else {
				isRewinding = false
			}
		} else {
			// Управление
			if rl.IsKeyPressed(rl.KeyR) && len(stateHistory) > 0 {
				isRewinding = true
				currentIndex = len(stateHistory) - 1
				continue
			}

			// Прыжок
			if rl.IsKeyPressed(rl.KeySpace) {
				if player.OnGround {
					player.Velocity.Y = jumpForce
					player.OnGround = false
				} else if player.OnWall {
					player.Velocity.Y = wallJumpForceY
					player.Velocity.X = -player.WallNormal * wallJumpForceX // Устанавливаем скорость прыжка от стены
					player.OnWall = false
					player.WallNormal = 0
					wallJumpLockoutTimer = wallJumpLockoutDuration // Запускаем таймер блокировки
				}
			}

			// Движение (игнорируем A/D, если на стене или активен таймер блокировки)
			if wallJumpLockoutTimer == 0 && !player.OnWall {
				player.Velocity.X = 0
				if rl.IsKeyDown(rl.KeyA) {
					player.Velocity.X = -moveSpeed
				}
				if rl.IsKeyDown(rl.KeyD) {
					player.Velocity.X = moveSpeed
				}
			}

			// Гравитация и скольжение по стене
			if player.OnWall && player.Velocity.Y > 0 {
				player.Velocity.Y = wallSlideSpeed // Медленное скольжение
			} else {
				player.Velocity.Y += gravity * dt
			}

			// Обновление позиции
			player.Position.X += player.Velocity.X * dt
			player.Position.Y += player.Velocity.Y * dt

			// Проверка коллизий с платформами
			player.OnGround = false
			playerRect := rl.NewRectangle(player.Position.X, player.Position.Y, player.Width, player.Height)
			for _, platform := range platforms {
				if rl.CheckCollisionRecs(playerRect, platform.Rect) {
					if player.Velocity.Y > 0 {
						player.Position.Y = platform.Rect.Y - player.Height
						player.Velocity.Y = 0
						player.OnGround = true
						player.OnWall = false
						player.WallNormal = 0
					} else if player.Velocity.Y < 0 {
						player.Position.Y = platform.Rect.Y + platform.Rect.Height
						player.Velocity.Y = 0
					}
				}
			}

			// Проверка коллизий со стенами
			player.OnWall = false
			player.WallNormal = 0
			if !player.OnGround {
				for _, wall := range walls {
					if rl.CheckCollisionRecs(playerRect, wall.Rect) {
						player.OnWall = true
						if player.Position.X < wall.Rect.X {
							player.WallNormal = 1 // Правая сторона стены
							player.Position.X = wall.Rect.X - player.Width
						} else {
							player.WallNormal = -1 // Левая сторона стены
							player.Position.X = wall.Rect.X + wall.Rect.Width
						}
						player.Velocity.X = 0
						break
					}
				}
			}

			// Границы экрана
			if player.Position.X < 0 {
				player.Position.X = 0
			}
			if player.Position.X > float32(rl.GetScreenWidth())-player.Width {
				player.Position.X = float32(rl.GetScreenWidth()) - player.Width
			}

			// Взлом объектов
			if rl.IsKeyPressed(rl.KeyE) {
				for i, obj := range hackables {
					if rl.CheckCollisionRecs(playerRect, obj.Rect) {
						hackables[i].Hacked = !hackables[i].Hacked
					}
				}
			}

			// Сохранение состояния
			currentState := GameState{
				PlayerPos:      player.Position,
				PlayerVel:      player.Velocity,
				PlayerOnGround: player.OnGround,
				PlayerOnWall:   player.OnWall,
				WallNormal:     player.WallNormal,
				Hackables:      make([]bool, len(hackables)),
			}
			for i := range hackables {
				currentState.Hackables[i] = hackables[i].Hacked
			}
			stateHistory = append(stateHistory, currentState)
			if len(stateHistory) > 600 {
				stateHistory = stateHistory[1:] // Ограничение истории до 600 кадров (~14.4 КБ)
			}
		}

		// Отрисовка
		rl.BeginDrawing()
		rl.ClearBackground(rl.Black)

		// Платформы
		for _, platform := range platforms {
			rl.DrawRectangleRec(platform.Rect, rl.Gray)
		}

		// Стены
		for _, wall := range walls {
			rl.DrawRectangleRec(wall.Rect, rl.DarkGray)
		}

		// Взламываемые объекты
		for _, obj := range hackables {
			color := rl.Red
			if obj.Hacked {
				color = rl.Green
			}
			rl.DrawRectangleRec(obj.Rect, color)
		}

		// Игрок
		rl.DrawRectangleRec(rl.NewRectangle(player.Position.X, player.Position.Y, player.Width, player.Height), rl.Blue)

		// Инструкции
		rl.DrawText("A/D - Move, Space - Jump/Wall Jump, E - Hack, R - Rewind", 10, 10, 20, rl.White)

		rl.EndDrawing()

		// Остановка перемотки
		if rl.IsKeyReleased(rl.KeyR) && isRewinding {
			isRewinding = false
			if currentIndex < len(stateHistory)-1 {
				stateHistory = stateHistory[:currentIndex+1]
			}
			wallJumpLockoutTimer = 0 // Сброс таймера при выходе из перемотки
		}
	}

	rl.CloseWindow()
}

// setGameState восстанавливает состояние игры
func setGameState(player *Player, hackables []HackableObject, state GameState) {
	player.Position = state.PlayerPos
	player.Velocity = state.PlayerVel
	player.OnGround = state.PlayerOnGround
	player.OnWall = state.PlayerOnWall
	player.WallNormal = state.WallNormal
	for i := range hackables {
		hackables[i].Hacked = state.Hackables[i]
	}
}
