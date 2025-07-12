package gamepkg

import (
	"math/rand"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type BodyPart int

const (
	Head BodyPart = iota
	Torso
	Legs
	None
)

type Animation struct {
	Texture      rl.Texture2D
	FrameCount   int
	FrameWidth   float32
	CurrentFrame int
	FrameTime    float32
	Timer        float32
}

type Player struct {
	Position           rl.Vector2
	TargetPosition     rl.Vector2 // Целевая позиция для интерполяции
	Velocity           float32    // Текущая скорость
	Size               rl.Vector2
	Speed              float32 // Максимальная скорость (MaxSpeed)
	Health             int
	TargetPart         BodyPart
	HeadBox            rl.Rectangle
	TorsoBox           rl.Rectangle
	LegsBox            rl.Rectangle
	IsAttacking        bool
	AttackPart         BodyPart
	AttackBox          rl.Rectangle
	PendingAttackPart  BodyPart
	HasPendingAttack   bool
	LastAttackTime     float32
	IsMoving           bool
	FacingRight        bool
	IsHit              bool
	HitTimer           float32
	HitDuration        float32
	IdleAnim           Animation
	RunAnim            Animation
	AttackHighAnim     Animation
	AttackMidAnim      Animation
	AttackLowAnim      Animation
	BackAttackHighAnim Animation
	BackAttackMidAnim  Animation
	BackAttackLowAnim  Animation
	HitHighAnim        Animation
	HitMidAnim         Animation
	HitLowAnim         Animation
	AttackDuration     float32
	AttackTimer        float32
	IsBackAttacking    bool
	BackAttackDuration float32
	BackAttackTimer    float32
}

func NewPlayer(position rl.Vector2, isHuman bool) Player {
	return Player{
		Position:           position,
		TargetPosition:     position,
		Velocity:           0.0,
		Size:               rl.NewVector2(SpriteSize*SpriteScale, SpriteSize*SpriteScale), // 48x48
		Speed:              MaxSpeed,                                                      // 25.0
		Health:             100,
		TargetPart:         None,
		HeadBox:            rl.NewRectangle(0, 0, 9*SpriteScale, 10*SpriteScale),  // 9x10
		TorsoBox:           rl.NewRectangle(0, 0, 19*SpriteScale, 12*SpriteScale), // 19x12
		LegsBox:            rl.NewRectangle(0, 0, 16*SpriteScale, 22*SpriteScale), // 16x22
		IsAttacking:        false,
		AttackPart:         Torso,
		PendingAttackPart:  None,
		HasPendingAttack:   false,
		AttackBox:          rl.NewRectangle(0, 0, AttackBoxWidth, AttackBoxHeight), // 24x6
		IsMoving:           false,
		FacingRight:        isHuman,
		IsHit:              false,
		HitTimer:           0,
		HitDuration:        0.3333, // Длительность анимации урона (4 кадра, 12 FPS)
		IdleAnim:           Animation{Texture: rl.LoadTexture("resources/threefour/idle.png"), FrameCount: 1, FrameWidth: 48, FrameTime: 0.1},
		RunAnim:            Animation{Texture: rl.LoadTexture("resources/threefour/run.png"), FrameCount: 10, FrameWidth: 48, FrameTime: RunFrameTime},
		AttackHighAnim:     Animation{Texture: rl.LoadTexture("resources/threefour/attack-high.png"), FrameCount: 8, FrameWidth: 48, FrameTime: AttackFrameTime},
		AttackMidAnim:      Animation{Texture: rl.LoadTexture("resources/threefour/attack-mid.png"), FrameCount: 8, FrameWidth: 48, FrameTime: AttackFrameTime},
		AttackLowAnim:      Animation{Texture: rl.LoadTexture("resources/threefour/attack-low.png"), FrameCount: 8, FrameWidth: 48, FrameTime: AttackFrameTime},
		BackAttackHighAnim: Animation{Texture: rl.LoadTexture("resources/threefour/back-attack-top-1.png"), FrameCount: 4, FrameWidth: 48, FrameTime: BackAttackFrameTime},
		BackAttackMidAnim:  Animation{Texture: rl.LoadTexture("resources/threefour/back-attack-mid-1.png"), FrameCount: 4, FrameWidth: 48, FrameTime: BackAttackFrameTime},
		BackAttackLowAnim:  Animation{Texture: rl.LoadTexture("resources/threefour/back-attack-low-1.png"), FrameCount: 4, FrameWidth: 48, FrameTime: BackAttackFrameTime},
		HitHighAnim:        Animation{Texture: rl.LoadTexture("resources/threefour/hit-top.png"), FrameCount: 4, FrameWidth: 48, FrameTime: HitFrameTime},
		HitMidAnim:         Animation{Texture: rl.LoadTexture("resources/threefour/hit-mid.png"), FrameCount: 4, FrameWidth: 48, FrameTime: HitFrameTime},
		HitLowAnim:         Animation{Texture: rl.LoadTexture("resources/threefour/hit-low.png"), FrameCount: 4, FrameWidth: 48, FrameTime: HitFrameTime},
		AttackDuration:     0.45,
		AttackTimer:        0,
		BackAttackDuration: float32(4) / AttackFps,
		BackAttackTimer:    0,
	}
}

func UpdateHitboxes(player *Player, isHuman bool) {
	centerX := player.Position.X + player.Size.X/2
	centerY := player.Position.Y + player.Size.Y/2
	player.HeadBox.X = centerX - (9*SpriteScale)/2 - 1*SpriteScale // 9x10
	player.HeadBox.Y = centerY - (10*SpriteScale)/2 - 15*SpriteScale
	player.TorsoBox.X = centerX - (19*SpriteScale)/2 - 1*SpriteScale // 19x12
	player.TorsoBox.Y = centerY - (12*SpriteScale)/2 - 4*SpriteScale
	player.LegsBox.X = centerX - (16*SpriteScale)/2 - 1*SpriteScale // 16x22
	player.LegsBox.Y = centerY - (22*SpriteScale)/2 + 13*SpriteScale

	var xOffset float32 = 6.0
	if !player.FacingRight {
		xOffset = -6.0
	}
	attackPart := player.AttackPart
	if isHuman && !player.IsAttacking {
		attackPart = player.TargetPart
		if attackPart == None {
			attackPart = Torso
		}
	} else if attackPart == None {
		attackPart = Torso
	}

	var attackBoxY float32
	switch attackPart {
	case Head:
		attackBoxY = player.HeadBox.Y + player.HeadBox.Height/2 - AttackBoxHeight/2
	case Torso:
		attackBoxY = player.TorsoBox.Y + player.TorsoBox.Height/2 - AttackBoxHeight/2
	case Legs:
		attackBoxY = player.LegsBox.Y + player.LegsBox.Height/2 - AttackBoxHeight/2
	default:
		attackBoxY = player.TorsoBox.Y + player.TorsoBox.Height/2 - AttackBoxHeight/2
	}

	attackBoxX := centerX + xOffset
	if attackBoxX < player.Position.X {
		attackBoxX = player.Position.X
	} else if attackBoxX+AttackBoxWidth > player.Position.X+player.Size.X {
		attackBoxX = player.Position.X + player.Size.X - AttackBoxWidth
	}

	player.AttackBox = rl.NewRectangle(attackBoxX, attackBoxY, AttackBoxWidth, AttackBoxHeight)
}

func SetAttackBox(player *Player, xOffset float32) {
	centerX := player.Position.X + player.Size.X/2
	var attackBoxY float32
	switch player.AttackPart {
	case Head:
		attackBoxY = player.HeadBox.Y + player.HeadBox.Height/2 - AttackBoxHeight/2
	case Torso:
		attackBoxY = player.TorsoBox.Y + player.TorsoBox.Height/2 - AttackBoxHeight/2
	case Legs:
		attackBoxY = player.LegsBox.Y + player.LegsBox.Height/2 - AttackBoxHeight/2
	default:
		attackBoxY = player.TorsoBox.Y + player.TorsoBox.Height/2 - AttackBoxHeight/2
	}

	attackBoxX := centerX + xOffset
	if attackBoxX < player.Position.X {
		attackBoxX = player.Position.X
	} else if attackBoxX+AttackBoxWidth > player.Position.X+player.Size.X {
		attackBoxX = player.Position.X + player.Size.X - AttackBoxWidth
	}

	player.AttackBox = rl.NewRectangle(attackBoxX, attackBoxY, AttackBoxWidth, AttackBoxHeight)
}

func HandleAnimationEnd(player *Player, timer, duration float32, isHuman bool, nextState func()) {
	if timer >= duration {
		nextState()
		if player.HasPendingAttack {
			player.IsAttacking = true
			player.IsBackAttacking = false
			player.IsHit = false
			player.AttackPart = player.PendingAttackPart
			player.HasPendingAttack = false
			player.PendingAttackPart = None
			var xOffset float32 = 6.0
			if !isHuman || !player.FacingRight {
				xOffset = -6.0
			}
			SetAttackBox(player, xOffset)
			ResetAnimation(&player.AttackHighAnim, &player.AttackMidAnim, &player.AttackLowAnim)
			player.AttackTimer = 0
		}
	}
}

func Lerp(start, end, t float32) float32 {
	return start + t*(end-start)
}

func UpdatePlayer(player *Player, ai *Player, isHuman bool, camera *rl.Camera2D, mousePos rl.Vector2) {
	if player.IsHit {
		player.HitTimer += rl.GetFrameTime()
		HandleAnimationEnd(player, player.HitTimer, player.HitDuration, isHuman, func() {
			player.IsHit = false
			player.HitTimer = 0
		})
		return
	}

	if player.IsAttacking {
		player.AttackTimer += rl.GetFrameTime()
		HandleAnimationEnd(player, player.AttackTimer, player.AttackDuration, isHuman, func() {
			player.IsAttacking = false
			player.IsBackAttacking = true
			player.BackAttackTimer = 0
			ResetAnimation(&player.BackAttackHighAnim, &player.BackAttackMidAnim, &player.BackAttackLowAnim)
		})
		return
	}

	if player.IsBackAttacking {
		player.BackAttackTimer += rl.GetFrameTime()
		HandleAnimationEnd(player, player.BackAttackTimer, player.BackAttackDuration, isHuman, func() {
			player.IsBackAttacking = false
			player.AttackPart = Torso
			player.BackAttackTimer = 0
		})
		return
	}

	// Плавное движение игрока
	player.IsMoving = false
	targetVelocity := float32(0.0)
	if rl.IsKeyDown(rl.KeyA) {
		targetVelocity = -player.Speed // Влево
		player.IsMoving = true
		player.FacingRight = false
	}
	if rl.IsKeyDown(rl.KeyD) {
		targetVelocity = player.Speed // Вправо
		player.IsMoving = true
		player.FacingRight = true
	}

	// Интерполяция скорости
	player.Velocity = Lerp(player.Velocity, targetVelocity, MoveAcceleration) // 0.95
	if targetVelocity == 0 {
		player.Velocity = Lerp(player.Velocity, 0, MoveDeceleration) // 0.98
	}

	// Обновляем целевую позицию
	player.TargetPosition.X += player.Velocity * rl.GetFrameTime()

	// Интерполяция позиции
	player.Position.X = Lerp(player.Position.X, player.TargetPosition.X, 0.95) // Увеличено с 0.9 до 0.95

	// Убираем ограничения позиции - теперь камера следует за игроком
	// Игрок может свободно двигаться по всей сцене

	UpdateHitboxes(player, isHuman)

	if isHuman && ai != nil {
		player.TargetPart = None
		if rl.CheckCollisionPointRec(mousePos, ai.HeadBox) {
			player.TargetPart = Head
		} else if rl.CheckCollisionPointRec(mousePos, ai.TorsoBox) {
			player.TargetPart = Torso
		} else if rl.CheckCollisionPointRec(mousePos, ai.LegsBox) {
			player.TargetPart = Legs
		}

		if rl.IsMouseButtonPressed(rl.MouseLeftButton) && player.TargetPart != None {
			if player.IsAttacking || player.IsBackAttacking {
				player.HasPendingAttack = true
				player.PendingAttackPart = player.TargetPart
			} else {
				player.IsAttacking = true
				player.AttackPart = player.TargetPart
				var xOffset float32 = 6.0
				if !player.FacingRight {
					xOffset = -6.0
				}
				SetAttackBox(player, xOffset)
				ResetAnimation(&player.AttackHighAnim, &player.AttackMidAnim, &player.AttackLowAnim)
				player.AttackTimer = 0
			}
		}
	}
}

func UpdateAI(ai *Player, player *Player, attackInterval float32) {
	if ai.IsHit {
		ai.HitTimer += rl.GetFrameTime()
		HandleAnimationEnd(ai, ai.HitTimer, ai.HitDuration, false, func() {
			ai.IsHit = false
			ai.HitTimer = 0
		})
		return
	}

	if ai.IsAttacking {
		ai.AttackTimer += rl.GetFrameTime()
		HandleAnimationEnd(ai, ai.AttackTimer, ai.AttackDuration, false, func() {
			ai.IsAttacking = false
			ai.IsBackAttacking = true
			ai.BackAttackTimer = 0
			ResetAnimation(&ai.BackAttackHighAnim, &ai.BackAttackMidAnim, &ai.BackAttackLowAnim)
		})
		return
	}

	if ai.IsBackAttacking {
		ai.BackAttackTimer += rl.GetFrameTime()
		HandleAnimationEnd(ai, ai.BackAttackTimer, ai.BackAttackDuration, false, func() {
			ai.IsBackAttacking = false
			ai.AttackPart = Torso
			ai.BackAttackTimer = 0
		})
		return
	}

	// Плавное движение ИИ
	ai.IsMoving = false
	targetVelocity := float32(0.0)
	if player.Position.X > ai.Position.X+50 {
		targetVelocity = ai.Speed // Вправо
		ai.IsMoving = true
		ai.FacingRight = true
	} else if player.Position.X < ai.Position.X-50 {
		targetVelocity = -ai.Speed // Влево
		ai.IsMoving = true
		ai.FacingRight = false
	}

	// Интерполяция скорости
	ai.Velocity = Lerp(ai.Velocity, targetVelocity, MoveAcceleration) // 0.95
	if targetVelocity == 0 {
		ai.Velocity = Lerp(ai.Velocity, 0, MoveDeceleration) // 0.98
	}

	// Обновляем целевую позицию
	ai.TargetPosition.X += ai.Velocity * rl.GetFrameTime()

	// Интерполяция позиции
	ai.Position.X = Lerp(ai.Position.X, ai.TargetPosition.X, 0.95) // Увеличено с 0.9 до 0.95

	// Убираем ограничения позиции для ИИ - теперь камера следует за игроком
	// ИИ может свободно двигаться по всей сцене

	UpdateHitboxes(ai, false)

	ai.LastAttackTime += rl.GetFrameTime()
	if ai.LastAttackTime >= attackInterval {
		ai.LastAttackTime = 0
		ai.IsAttacking = true
		ai.AttackPart = BodyPart(rand.Intn(3))
		var xOffset float32 = -6.0
		if ai.FacingRight {
			xOffset = 6.0
		}
		SetAttackBox(ai, xOffset)
		ai.TargetPart = BodyPart(rand.Intn(3))
		ResetAnimation(&ai.AttackHighAnim, &ai.AttackMidAnim, &ai.AttackLowAnim)
		ai.AttackTimer = 0
	}
}
