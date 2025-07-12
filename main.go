package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	screenWidth  = 800
	screenHeight = 450

	gravity   = float32(0.5)
	jumpSpeed = float32(-12.0)
)

type Player struct {
	Position    rl.Vector2   // Позиция игрока (x, y)
	Size        rl.Vector2   // Размер (ширина, высота)
	Speed       float32      // Скорость движения
	Health      int          // Здоровье
	IsJumping   bool         // Флаг прыжка
	IsAttacking bool         // Флаг атаки
	IsBlocking  bool         // Флаг блока
	AttackBox   rl.Rectangle // Хитбокс атаки
}

func main() {
	// Инициализация окна
	rl.InitWindow(screenWidth, screenHeight, "Fist Fight Prototype")
	rl.SetTargetFPS(60)

	// Создание игроков
	player1 := Player{
		Position:    rl.NewVector2(100, screenHeight-50),
		Size:        rl.NewVector2(40, 60),
		Speed:       5.0,
		Health:      100,
		IsJumping:   false,
		IsAttacking: false,
		IsBlocking:  false,
		AttackBox:   rl.NewRectangle(0, 0, 50, 30),
	}
	player2 := Player{
		Position:    rl.NewVector2(screenWidth-140, screenHeight-50),
		Size:        rl.NewVector2(40, 60),
		Speed:       5.0,
		Health:      100,
		IsJumping:   false,
		IsAttacking: false,
		IsBlocking:  false,
		AttackBox:   rl.NewRectangle(0, 0, 50, 30),
	}

	for !rl.WindowShouldClose() {
		// Обновление
		updatePlayer(&player1, rl.KeyA, rl.KeyD, rl.KeyW, rl.KeyS, true)
		updatePlayer(&player2, rl.KeyLeft, rl.KeyRight, rl.KeyUp, rl.KeyDown, false)

		// Проверка столкновений и атаки
		checkAttack(&player1, &player2)
		checkAttack(&player2, &player1)

		// Отрисовка
		rl.BeginDrawing()
		rl.ClearBackground(rl.RayWhite)

		// Рисуем игроков
		rl.DrawRectangleV(player1.Position, player1.Size, rl.Red)
		rl.DrawRectangleV(player2.Position, player2.Size, rl.Blue)

		// Рисуем хитбоксы атаки, если активны
		if player1.IsAttacking {
			rl.DrawRectangleRec(player1.AttackBox, rl.Fade(rl.Red, 0.5))
		}
		if player2.IsAttacking {
			rl.DrawRectangleRec(player2.AttackBox, rl.Fade(rl.Blue, 0.5))
		}

		// Отображаем здоровье
		rl.DrawText("Player 1 HP: "+string(rune(player1.Health)), 10, 10, 20, rl.Black)
		rl.DrawText("Player 2 HP: "+string(rune(player2.Health)), screenWidth-150, 10, 20, rl.Black)

		rl.EndDrawing()
	}

	rl.CloseWindow()
}

func updatePlayer(player *Player, leftKey, rightKey, jumpKey, attackKey int32, isPlayer1 bool) {
	// Движение влево/вправо
	if rl.IsKeyDown(leftKey) {
		player.Position.X -= player.Speed
	}
	if rl.IsKeyDown(rightKey) {
		player.Position.X += player.Speed
	}

	// Прыжок
	if rl.IsKeyPressed(jumpKey) && !player.IsJumping {
		player.IsJumping = true
		player.Position.Y += jumpSpeed
	}

	// Гравитация
	if player.IsJumping {
		player.Position.Y += gravity
		if player.Position.Y >= screenHeight-player.Size.Y {
			player.Position.Y = screenHeight - player.Size.Y
			player.IsJumping = false
		}
	}

	// Атака и блок
	if rl.IsKeyPressed(attackKey) {
		player.IsAttacking = true
		// Устанавливаем хитбокс атаки
		if isPlayer1 {
			player.AttackBox = rl.NewRectangle(player.Position.X+player.Size.X, player.Position.Y, 50, 30)
		} else {
			player.AttackBox = rl.NewRectangle(player.Position.X-50, player.Position.Y, 50, 30)
		}
	} else if rl.IsKeyReleased(attackKey) {
		player.IsAttacking = false
	}

	// Блок (используем Shift для игрока 1 и Ctrl для игрока 2)
	if isPlayer1 && rl.IsKeyDown(rl.KeyLeftShift) || !isPlayer1 && rl.IsKeyDown(rl.KeyLeftControl) {
		player.IsBlocking = true
	} else {
		player.IsBlocking = false
	}

	// Ограничение по границам экрана
	if player.Position.X < 0 {
		player.Position.X = 0
	}
	if player.Position.X > screenWidth-player.Size.X {
		player.Position.X = screenWidth - player.Size.X
	}
}

func checkAttack(attacker, defender *Player) {
	if attacker.IsAttacking && !defender.IsBlocking {
		playerRect := rl.NewRectangle(defender.Position.X, defender.Position.Y, defender.Size.X, defender.Size.Y)
		if rl.CheckCollisionRecs(attacker.AttackBox, playerRect) {
			defender.Health -= 10
			attacker.IsAttacking = false // Сбрасываем атаку после попадания
			if defender.Health < 0 {
				defender.Health = 0
			}
		}
	}
}
