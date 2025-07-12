package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"math/rand"
	"strconv"
	"time"
)

const (
	screenWidth  = 800
	screenHeight = 450

	gravity   = float32(0.5)
	jumpSpeed = float32(-12.0)
)

type BodyPart int

const (
	Head BodyPart = iota
	Torso
	Legs
	None
)

type Player struct {
	Position       rl.Vector2   // Позиция
	Size           rl.Vector2   // Размер
	Speed          float32      // Скорость
	Health         int          // Здоровье
	IsJumping      bool         // Флаг прыжка
	TargetPart     BodyPart     // Целевая часть тела (для атаки/блока)
	HeadBox        rl.Rectangle // Хитбокс головы
	TorsoBox       rl.Rectangle // Хитбокс торса
	LegsBox        rl.Rectangle // Хитбокс ног
	IsAttacking    bool         // Флаг атаки
	AttackPart     BodyPart     // Часть тела для атаки
	AttackBox      rl.Rectangle // Хитбокс атаки
	LastAttackTime float32      // Время последней атаки (для ИИ)
}

func main() {
	// Инициализация
	rl.InitWindow(screenWidth, screenHeight, "Fist Fight with AI")
	rl.SetTargetFPS(60)
	rand.Seed(time.Now().UnixNano())

	// Игрок
	player := Player{
		Position:    rl.NewVector2(100, screenHeight-90),
		Size:        rl.NewVector2(40, 90),
		Speed:       5.0,
		Health:      100,
		IsJumping:   false,
		TargetPart:  None,
		HeadBox:     rl.NewRectangle(100, screenHeight-90, 40, 30),
		TorsoBox:    rl.NewRectangle(100, screenHeight-60, 40, 30),
		LegsBox:     rl.NewRectangle(100, screenHeight-30, 40, 30),
		IsAttacking: false,
		AttackPart:  None,
		AttackBox:   rl.NewRectangle(0, 0, 50, 30),
	}

	// Противник (ИИ)
	ai := Player{
		Position:       rl.NewVector2(screenWidth-140, screenHeight-90),
		Size:           rl.NewVector2(40, 90),
		Speed:          4.0,
		Health:         100,
		IsJumping:      false,
		TargetPart:     None,
		HeadBox:        rl.NewRectangle(screenWidth-140, screenHeight-90, 40, 30),
		TorsoBox:       rl.NewRectangle(screenWidth-140, screenHeight-60, 40, 30),
		LegsBox:        rl.NewRectangle(screenWidth-140, screenHeight-30, 40, 30),
		IsAttacking:    false,
		AttackPart:     None,
		AttackBox:      rl.NewRectangle(0, 0, 50, 30),
		LastAttackTime: 0,
	}

	aiAttackInterval := float32(1.5) // Интервал атаки ИИ (в секундах)

	for !rl.WindowShouldClose() {
		// Обновление
		updatePlayer(&ai, &player, true)
		updateAI(&ai, &player, aiAttackInterval)

		// Проверка атак
		checkAttack(&player, &ai, true)
		checkAttack(&ai, &player, false)

		// Отрисовка
		rl.BeginDrawing()
		rl.ClearBackground(rl.RayWhite)

		// Рисуем игроков
		drawPlayer(&player, rl.Red, true)
		drawPlayer(&ai, rl.Blue, false)

		// Отображаем здоровье
		rl.DrawText("Player HP: "+strconv.Itoa(player.Health), 10, 10, 20, rl.Black)
		rl.DrawText("AI HP: "+strconv.Itoa(ai.Health), screenWidth-150, 10, 20, rl.Black)

		rl.EndDrawing()
	}

	rl.CloseWindow()
}

func updatePlayer(ai *Player, player *Player, isHuman bool) {
	// Движение
	if rl.IsKeyDown(rl.KeyA) {
		player.Position.X -= player.Speed
	}
	if rl.IsKeyDown(rl.KeyD) {
		player.Position.X += player.Speed
	}

	// Прыжок
	if rl.IsKeyPressed(rl.KeyW) && !player.IsJumping {
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

	// Обновление хитбоксов
	player.HeadBox.X = player.Position.X
	player.HeadBox.Y = player.Position.Y
	player.TorsoBox.X = player.Position.X
	player.TorsoBox.Y = player.Position.Y + 30
	player.LegsBox.X = player.Position.X
	player.LegsBox.Y = player.Position.Y + 60

	// Управление мышью (только для игрока)
	if isHuman {
		mousePos := rl.GetMousePosition()
		player.TargetPart = None
		if rl.CheckCollisionPointRec(mousePos, ai.HeadBox) {
			player.TargetPart = Head
		} else if rl.CheckCollisionPointRec(mousePos, ai.TorsoBox) {
			player.TargetPart = Torso
		} else if rl.CheckCollisionPointRec(mousePos, ai.LegsBox) {
			player.TargetPart = Legs
		}

		// Атака
		if rl.IsMouseButtonPressed(rl.MouseLeftButton) && player.TargetPart != None {
			player.IsAttacking = true
			player.AttackPart = player.TargetPart
			player.AttackBox = rl.NewRectangle(player.Position.X+player.Size.X, player.Position.Y+float32(player.AttackPart)*30, 50, 30)
		} else if rl.IsMouseButtonReleased(rl.MouseLeftButton) {
			player.IsAttacking = false
			player.AttackPart = None
		}
	}

	// Ограничение по границам
	if player.Position.X < 0 {
		player.Position.X = 0
	}
	if player.Position.X > screenWidth-player.Size.X {
		player.Position.X = screenWidth - player.Size.X
	}
}

func updateAI(ai *Player, player *Player, attackInterval float32) {
	// Движение ИИ: следовать за игроком
	if player.Position.X > ai.Position.X+50 {
		ai.Position.X += ai.Speed
	} else if player.Position.X < ai.Position.X-50 {
		ai.Position.X -= ai.Speed
	}

	// Случайный прыжок
	if !ai.IsJumping && rand.Float32() < 0.01 {
		ai.IsJumping = true
		ai.Position.Y += jumpSpeed
	}

	// Гравитация
	if ai.IsJumping {
		ai.Position.Y += gravity
		if ai.Position.Y >= screenHeight-ai.Size.Y {
			ai.Position.Y = screenHeight - ai.Size.Y
			ai.IsJumping = false
		}
	}

	// Обновление хитбоксов
	ai.HeadBox.X = ai.Position.X
	ai.HeadBox.Y = ai.Position.Y
	ai.TorsoBox.X = ai.Position.X
	ai.TorsoBox.Y = ai.Position.Y + 30
	ai.LegsBox.X = ai.Position.X
	ai.LegsBox.Y = ai.Position.Y + 60

	// ИИ: выбор части тела для атаки и блока
	ai.LastAttackTime += rl.GetFrameTime()
	if ai.LastAttackTime >= attackInterval {
		ai.LastAttackTime = 0
		ai.IsAttacking = true
		ai.AttackPart = BodyPart(rand.Intn(3)) // Случайная часть тела
		ai.AttackBox = rl.NewRectangle(ai.Position.X-50, ai.Position.Y+float32(ai.AttackPart)*30, 50, 30)
		ai.TargetPart = BodyPart(rand.Intn(3)) // Случайный блок
	} else if ai.IsAttacking && ai.LastAttackTime > 0.3 {
		ai.IsAttacking = false
		ai.AttackPart = None
	}

	// Ограничение по границам
	if ai.Position.X < 0 {
		ai.Position.X = 0
	}
	if ai.Position.X > screenWidth-ai.Size.X {
		ai.Position.X = screenWidth - ai.Size.X
	}
}

func checkAttack(attacker, defender *Player, isHuman bool) {
	if attacker.IsAttacking {
		var defenderBox rl.Rectangle
		switch attacker.AttackPart {
		case Head:
			defenderBox = defender.HeadBox
		case Torso:
			defenderBox = defender.TorsoBox
		case Legs:
			defenderBox = defender.LegsBox
		}

		// Проверка блока
		if attacker.AttackPart == defender.TargetPart {
			// Блокировка успешна
			attacker.IsAttacking = false
			if isHuman {
				attacker.AttackPart = None
			}
		} else if rl.CheckCollisionRecs(attacker.AttackBox, defenderBox) {
			// Удар попал
			defender.Health -= 10
			attacker.IsAttacking = false
			if isHuman {
				attacker.AttackPart = None
			}
			if defender.Health < 0 {
				defender.Health = 0
			}
		}
	}
}

func drawPlayer(player *Player, color rl.Color, isHuman bool) {
	// Рисуем части тела
	if player.TargetPart == Head {
		rl.DrawRectangleRec(player.HeadBox, rl.Fade(color, 0.7))
	} else {
		rl.DrawRectangleRec(player.HeadBox, color)
	}
	if player.TargetPart == Torso {
		rl.DrawRectangleRec(player.TorsoBox, rl.Fade(color, 0.7))
	} else {
		rl.DrawRectangleRec(player.TorsoBox, color)
	}
	if player.TargetPart == Legs {
		rl.DrawRectangleRec(player.LegsBox, rl.Fade(color, 0.7))
	} else {
		rl.DrawRectangleRec(player.LegsBox, color)
	}

	// Рисуем хитбокс атаки
	if player.IsAttacking {
		rl.DrawRectangleRec(player.AttackBox, rl.Fade(color, 0.5))
	}
}
