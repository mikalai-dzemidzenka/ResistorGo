package gamepkg

import (
	"fmt"
	"math/rand"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// ActorType определяет тип актора (игрок или ИИ)
type ActorType int

const (
	PlayerActor ActorType = iota
	AIActor
)

// Actor объединяет функциональность Player и AI в единую сущность
type Actor struct {
	// Основные свойства
	Type           ActorType
	Position       rl.Vector2
	TargetPosition rl.Vector2 // Целевая позиция для интерполяции
	Velocity       float32    // Текущая скорость
	Size           rl.Vector2
	Speed          float32 // Максимальная скорость (MaxSpeed)
	Health         int
	TargetPart     BodyPart

	// Хитбоксы
	HeadBox  rl.Rectangle
	TorsoBox rl.Rectangle
	LegsBox  rl.Rectangle

	// Система атак
	IsAttacking       bool
	AttackPart        BodyPart
	AttackBox         rl.Rectangle
	PendingAttackPart BodyPart
	HasPendingAttack  bool
	IsPendingAttack2  bool // Флаг для атаки-2
	IsCurrentAttack2  bool // Флаг для текущей атаки-2
	LastAttackTime    float32
	AttackCooldown    float32 // Защита от спама атак

	// Состояние
	IsMoving    bool
	FacingRight bool
	IsHit       bool
	HitTimer    float32
	HitDuration float32

	// Анимации
	IdleAnim           Animation
	WalkAnim           Animation // Анимация ходьбы (нижняя часть тела - ноги)
	WalkHighAnim       Animation // Анимация ходьбы с высоким положением рук
	WalkMidAnim        Animation // Анимация ходьбы со средним положением рук
	WalkLowAnim        Animation // Анимация ходьбы с низким положением рук
	RunAnim            Animation
	AttackHighAnim     Animation
	AttackMidAnim      Animation
	AttackLowAnim      Animation
	AttackHighAnim2    Animation // Новые анимации для атак-2
	AttackMidAnim2     Animation
	AttackLowAnim2     Animation
	BackAttackHighAnim Animation
	BackAttackMidAnim  Animation
	BackAttackLowAnim  Animation
	HitHighAnim        Animation
	HitMidAnim         Animation
	HitLowAnim         Animation
	UpDownAnim         Animation // Анимация для отображения текущей позы (low/mid/high)

	// Анимации переходов между позами
	UpDownLowToMidAnim  Animation // Переход от low к mid (updown-0-1.png)
	UpDownMidToHighAnim Animation // Переход от mid к high (updown-1-2.png)
	UpDownHighToMidAnim Animation // Переход от high к mid (updown-2-1.png)
	UpDownMidToLowAnim  Animation // Переход от mid к low (updown-1-0.png)

	// Состояние перехода между позами
	IsTransitioning    bool
	TransitionFrom     BodyPart
	TransitionTo       BodyPart
	TransitionTimer    float32
	TransitionDuration float32
	CurrentPose        BodyPart // Последняя завершённая поза
	NextTargetPart     BodyPart // Буфер следующей цели

	// Таймеры атак
	AttackDuration     float32
	AttackTimer        float32
	IsBackAttacking    bool
	BackAttackDuration float32
	BackAttackTimer    float32

	// Параметры ИИ
	AttackInterval float32
}

// NewActor создает нового актора указанного типа
func NewActor(position rl.Vector2, actorType ActorType) Actor {
	isHuman := actorType == PlayerActor

	return Actor{
		Type:               actorType,
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
		IsPendingAttack2:   false,
		IsCurrentAttack2:   false,
		AttackBox:          rl.NewRectangle(0, 0, AttackBoxWidth, AttackBoxHeight), // 24x6
		IsMoving:           false,
		FacingRight:        isHuman,
		IsHit:              false,
		HitTimer:           0,
		HitDuration:        0.3333, // Длительность анимации урона (4 кадра, 12 FPS)
		IdleAnim:           Animation{Texture: rl.LoadTexture("resources/threefour/idle.png"), FrameCount: 1, FrameWidth: 48, FrameTime: 0.1},
		WalkAnim:           Animation{Texture: rl.LoadTexture("resources/threefour/walk.png"), FrameCount: 8, FrameWidth: 48, FrameTime: WalkFrameTime},
		WalkHighAnim:       Animation{Texture: rl.LoadTexture("resources/threefour/walk-high.png"), FrameCount: 8, FrameWidth: 48, FrameTime: WalkFrameTime},
		WalkMidAnim:        Animation{Texture: rl.LoadTexture("resources/threefour/walk-mid.png"), FrameCount: 8, FrameWidth: 48, FrameTime: WalkFrameTime},
		WalkLowAnim:        Animation{Texture: rl.LoadTexture("resources/threefour/walk-low.png"), FrameCount: 8, FrameWidth: 48, FrameTime: WalkFrameTime},
		RunAnim:            Animation{Texture: rl.LoadTexture("resources/threefour/run.png"), FrameCount: 10, FrameWidth: 48, FrameTime: RunFrameTime},
		AttackHighAnim:     Animation{Texture: rl.LoadTexture("resources/threefour/attack-high.png"), FrameCount: 8, FrameWidth: 48, FrameTime: AttackFrameTime},
		AttackMidAnim:      Animation{Texture: rl.LoadTexture("resources/threefour/attack-mid.png"), FrameCount: 8, FrameWidth: 48, FrameTime: AttackFrameTime},
		AttackLowAnim:      Animation{Texture: rl.LoadTexture("resources/threefour/attack-low.png"), FrameCount: 8, FrameWidth: 48, FrameTime: AttackFrameTime},
		AttackHighAnim2:    Animation{Texture: rl.LoadTexture("resources/threefour/attack-high-2.png"), FrameCount: 8, FrameWidth: 48, FrameTime: AttackFrameTime},
		AttackMidAnim2:     Animation{Texture: rl.LoadTexture("resources/threefour/attack-mid-2.png"), FrameCount: 8, FrameWidth: 48, FrameTime: AttackFrameTime},
		AttackLowAnim2:     Animation{Texture: rl.LoadTexture("resources/threefour/attack-low-2.png"), FrameCount: 8, FrameWidth: 48, FrameTime: AttackFrameTime},
		BackAttackHighAnim: Animation{Texture: rl.LoadTexture("resources/threefour/back-attack-top-1.png"), FrameCount: 4, FrameWidth: 48, FrameTime: BackAttackFrameTime},
		BackAttackMidAnim:  Animation{Texture: rl.LoadTexture("resources/threefour/back-attack-mid-1.png"), FrameCount: 4, FrameWidth: 48, FrameTime: BackAttackFrameTime},
		BackAttackLowAnim:  Animation{Texture: rl.LoadTexture("resources/threefour/back-attack-low-1.png"), FrameCount: 4, FrameWidth: 48, FrameTime: BackAttackFrameTime},
		HitHighAnim:        Animation{Texture: rl.LoadTexture("resources/threefour/hit-top.png"), FrameCount: 4, FrameWidth: 48, FrameTime: HitFrameTime},
		HitMidAnim:         Animation{Texture: rl.LoadTexture("resources/threefour/hit-mid.png"), FrameCount: 4, FrameWidth: 48, FrameTime: HitFrameTime},
		HitLowAnim:         Animation{Texture: rl.LoadTexture("resources/threefour/hit-low.png"), FrameCount: 4, FrameWidth: 48, FrameTime: HitFrameTime},
		UpDownAnim:         Animation{Texture: rl.LoadTexture("resources/threefour/updown.png"), FrameCount: 3, FrameWidth: 48, FrameTime: 0.1}, // 3 кадра: low/mid/high

		// Анимации переходов между позами (4 кадра каждая)
		UpDownLowToMidAnim:  Animation{Texture: rl.LoadTexture("resources/threefour/updown-0-1.png"), FrameCount: 4, FrameWidth: 48, FrameTime: UpDownTransitionFrameTime},
		UpDownMidToHighAnim: Animation{Texture: rl.LoadTexture("resources/threefour/updown-1-2.png"), FrameCount: 4, FrameWidth: 48, FrameTime: UpDownTransitionFrameTime},
		UpDownHighToMidAnim: Animation{Texture: rl.LoadTexture("resources/threefour/updown-2-1.png"), FrameCount: 4, FrameWidth: 48, FrameTime: UpDownTransitionFrameTime},
		UpDownMidToLowAnim:  Animation{Texture: rl.LoadTexture("resources/threefour/updown-1-0.png"), FrameCount: 4, FrameWidth: 48, FrameTime: UpDownTransitionFrameTime},

		// Состояние перехода
		IsTransitioning:    false,
		TransitionFrom:     None,
		TransitionTo:       None,
		TransitionTimer:    0,
		TransitionDuration: UpDownTransitionDuration, // 0.1 сек
		CurrentPose:        Torso,
		NextTargetPart:     None,

		AttackDuration:     0.45,
		AttackTimer:        0,
		BackAttackDuration: float32(4) / AttackFps,
		BackAttackTimer:    0,
		AttackInterval:     1.5, // Интервал атак для ИИ
		AttackCooldown:     0.0, // Кулдаун между атаками
	}
}

// UpdateActor объединяет логику обновления игрока и ИИ
func UpdateActor(actor *Actor, opponent *Actor, camera *rl.Camera2D, mousePos rl.Vector2) {
	// Обработка атак в зависимости от типа актора (ДО обработки состояний)
	if actor.Type == PlayerActor {
		// Логика атак игрока
		if opponent != nil {
			actor.TargetPart = None
			if rl.CheckCollisionPointRec(mousePos, opponent.HeadBox) {
				actor.TargetPart = Head
			} else if rl.CheckCollisionPointRec(mousePos, opponent.TorsoBox) {
				actor.TargetPart = Torso
			} else if rl.CheckCollisionPointRec(mousePos, opponent.LegsBox) {
				actor.TargetPart = Legs
			}

			// Обновляем кулдаун атаки
			if actor.AttackCooldown > 0 {
				actor.AttackCooldown -= rl.GetFrameTime()
			}

			if rl.IsMouseButtonPressed(rl.MouseLeftButton) && actor.TargetPart != None && actor.AttackCooldown <= 0 {
				if actor.IsAttacking || actor.IsBackAttacking {
					// Если уже атакуем, ставим атаку-2 в очередь
					actor.HasPendingAttack = true
					actor.PendingAttackPart = actor.TargetPart
					actor.IsPendingAttack2 = true // Флаг для атаки-2
				} else {
					// Начинаем обычную атаку
					actor.IsAttacking = true
					actor.AttackPart = actor.TargetPart
					actor.IsCurrentAttack2 = false // Обычная атака
					var xOffset float32 = 6.0
					if !actor.FacingRight {
						xOffset = -6.0
					}
					SetActorAttackBox(actor, xOffset)
					ResetAnimation(&actor.AttackHighAnim, &actor.AttackMidAnim, &actor.AttackLowAnim)
					actor.AttackTimer = 0
					actor.AttackCooldown = 0.1 // Небольшой кулдаун после начала атаки
				}
			}
		}
	} else {
		// Логика движения ИИ - теперь используем новую систему позиционирования
		// Эта функция будет вызываться из main.go с передачей всех врагов
		// Здесь оставляем только базовую логику атак
		actor.LastAttackTime += rl.GetFrameTime()
		if actor.LastAttackTime >= actor.AttackInterval {
			actor.LastAttackTime = 0
			actor.IsAttacking = true
			actor.AttackPart = BodyPart(rand.Intn(3))
			actor.IsCurrentAttack2 = false // ИИ всегда использует обычные атаки
			var xOffset float32 = -6.0
			if actor.FacingRight {
				xOffset = 6.0
			}
			SetActorAttackBox(actor, xOffset)
			actor.TargetPart = BodyPart(rand.Intn(3))
			ResetAnimation(&actor.AttackHighAnim, &actor.AttackMidAnim, &actor.AttackLowAnim)
			actor.AttackTimer = 0
		}
	}

	// Обработка состояния получения урона
	if actor.IsHit {
		actor.HitTimer += rl.GetFrameTime()
		HandleActorAnimationEnd(actor, actor.HitTimer, actor.HitDuration, func() {
			actor.IsHit = false
			actor.HitTimer = 0
		})
		return
	}

	// Обработка состояния атаки
	if actor.IsAttacking {
		actor.AttackTimer += rl.GetFrameTime()
		HandleActorAnimationEnd(actor, actor.AttackTimer, actor.AttackDuration, func() {
			actor.IsAttacking = false
			actor.IsCurrentAttack2 = false // Сбрасываем флаг текущей атаки-2
			actor.IsBackAttacking = true
			actor.BackAttackTimer = 0
			ResetAnimation(&actor.BackAttackHighAnim, &actor.BackAttackMidAnim, &actor.BackAttackLowAnim)
		})
		return
	}

	// Обработка состояния возврата из атаки
	if actor.IsBackAttacking {
		actor.BackAttackTimer += rl.GetFrameTime()
		HandleActorAnimationEnd(actor, actor.BackAttackTimer, actor.BackAttackDuration, func() {
			actor.IsBackAttacking = false
			actor.AttackPart = Torso
			actor.BackAttackTimer = 0
		})
		return
	}

	// --- Плавное движение ---
	actor.IsMoving = false
	targetVelocity := float32(0.0)
	acceleration := float32(80.0)  // Уменьшено с 120 для снижения скольжения
	deceleration := float32(120.0) // Уменьшено с 200 для снижения скольжения
	dt := rl.GetFrameTime()

	if actor.Type == PlayerActor {
		if rl.IsKeyDown(rl.KeyA) {
			targetVelocity = -actor.Speed // Влево
			actor.IsMoving = true
			actor.FacingRight = false
		}
		if rl.IsKeyDown(rl.KeyD) {
			targetVelocity = actor.Speed // Вправо
			actor.IsMoving = true
			actor.FacingRight = true
		}
	} else {
		if opponent.Position.X > actor.Position.X+50 {
			targetVelocity = actor.Speed // Вправо
			actor.IsMoving = true
			actor.FacingRight = true
		} else if opponent.Position.X < actor.Position.X-50 {
			targetVelocity = -actor.Speed // Влево
			actor.IsMoving = true
			actor.FacingRight = false
		}
	}

	if targetVelocity != 0 {
		if (targetVelocity > 0 && actor.Velocity < targetVelocity) || (targetVelocity < 0 && actor.Velocity > targetVelocity) {
			actor.Velocity += acceleration * dt * sign(targetVelocity)
			if (targetVelocity > 0 && actor.Velocity > targetVelocity) || (targetVelocity < 0 && actor.Velocity < targetVelocity) {
				actor.Velocity = targetVelocity
			}
		} else if (targetVelocity > 0 && actor.Velocity > targetVelocity) || (targetVelocity < 0 && actor.Velocity < targetVelocity) {
			actor.Velocity -= deceleration * dt * sign(actor.Velocity)
			if (targetVelocity > 0 && actor.Velocity < targetVelocity) || (targetVelocity < 0 && actor.Velocity > targetVelocity) {
				actor.Velocity = targetVelocity
			}
		}
	} else {
		// Быстрое торможение до нуля
		if actor.Velocity > 0 {
			actor.Velocity -= BrakeSpeed * dt
			if actor.Velocity < 0 {
				actor.Velocity = 0
			}
		} else if actor.Velocity < 0 {
			actor.Velocity += BrakeSpeed * dt
			if actor.Velocity > 0 {
				actor.Velocity = 0
			}
		}
	}

	// Обновляем целевую позицию
	actor.TargetPosition.X += actor.Velocity * dt

	// Интерполяция позиции
	actor.Position.X = Lerp(actor.Position.X, actor.TargetPosition.X, 0.95)

	// Обновляем хитбоксы после изменения позиции
	UpdateActorHitboxes(actor)
}

// UpdateActorHitboxes обновляет хитбоксы актора после изменения позиции
func UpdateActorHitboxes(actor *Actor) {
	centerX := actor.Position.X + actor.Size.X/2
	centerY := actor.Position.Y + actor.Size.Y/2

	// Хитбоксы основаны на центре актора
	actor.HeadBox.X = centerX - (9*SpriteScale)/2 - 1*SpriteScale // 9x10
	actor.HeadBox.Y = centerY - (10*SpriteScale)/2 - 15*SpriteScale
	actor.HeadBox.Width = 9 * SpriteScale
	actor.HeadBox.Height = 10 * SpriteScale

	actor.TorsoBox.X = centerX - (19*SpriteScale)/2 - 1*SpriteScale // 19x12
	actor.TorsoBox.Y = centerY - (12*SpriteScale)/2 - 4*SpriteScale
	actor.TorsoBox.Width = 19 * SpriteScale
	actor.TorsoBox.Height = 12 * SpriteScale

	actor.LegsBox.X = centerX - (16*SpriteScale)/2 - 1*SpriteScale // 16x22
	actor.LegsBox.Y = centerY - (22*SpriteScale)/2 + 13*SpriteScale
	actor.LegsBox.Width = 16 * SpriteScale
	actor.LegsBox.Height = 22 * SpriteScale

	// AttackBox всегда присутствует и следует за актором
	// При взгляде вправо - в правой половине спрайта, при взгляде влево - в левой половине
	var attackBoxX float32
	if actor.FacingRight {
		// Взгляд вправо - AttackBox в правой половине спрайта (24-48px)
		attackBoxX = actor.Position.X + 24
	} else {
		// Взгляд влево - AttackBox в левой половине спрайта (0-24px)
		attackBoxX = actor.Position.X
	}

	// Определяем часть тела для атаки
	attackPart := actor.AttackPart
	if actor.Type == PlayerActor && !actor.IsAttacking {
		attackPart = actor.TargetPart
		if attackPart == None {
			attackPart = Torso
		}
	} else if attackPart == None {
		attackPart = Torso
	}

	// Устанавливаем позицию AttackBox в зависимости от части тела
	var attackBoxY float32
	switch attackPart {
	case Head:
		attackBoxY = actor.HeadBox.Y + actor.HeadBox.Height/2 - AttackBoxHeight/2
	case Torso:
		attackBoxY = actor.TorsoBox.Y + actor.TorsoBox.Height/2 - AttackBoxHeight/2
	case Legs:
		attackBoxY = actor.LegsBox.Y + actor.LegsBox.Height/2 - AttackBoxHeight/2
	default:
		attackBoxY = actor.TorsoBox.Y + actor.TorsoBox.Height/2 - AttackBoxHeight/2
	}

	actor.AttackBox = rl.NewRectangle(attackBoxX, attackBoxY, AttackBoxWidth, AttackBoxHeight)
}

// SetActorAttackBox устанавливает позицию атаки для актора
func SetActorAttackBox(actor *Actor, xOffset float32) {
	centerX := actor.Position.X + actor.Size.X/2
	var attackBoxY float32
	switch actor.AttackPart {
	case Head:
		attackBoxY = actor.HeadBox.Y + actor.HeadBox.Height/2 - AttackBoxHeight/2
	case Torso:
		attackBoxY = actor.TorsoBox.Y + actor.TorsoBox.Height/2 - AttackBoxHeight/2
	case Legs:
		attackBoxY = actor.LegsBox.Y + actor.LegsBox.Height/2 - AttackBoxHeight/2
	default:
		attackBoxY = actor.TorsoBox.Y + actor.TorsoBox.Height/2 - AttackBoxHeight/2
	}

	attackBoxX := centerX + xOffset
	if attackBoxX < actor.Position.X {
		attackBoxX = actor.Position.X
	} else if attackBoxX+AttackBoxWidth > actor.Position.X+actor.Size.X {
		attackBoxX = actor.Position.X + actor.Size.X - AttackBoxWidth
	}

	actor.AttackBox = rl.NewRectangle(attackBoxX, attackBoxY, AttackBoxWidth, AttackBoxHeight)
}

// HandleActorAnimationEnd обрабатывает завершение анимации для актора
func HandleActorAnimationEnd(actor *Actor, timer, duration float32, nextState func()) {
	if timer >= duration {
		nextState()
		// Обрабатываем ожидающую атаку только если актор не в состоянии получения урона
		if actor.HasPendingAttack && !actor.IsHit {
			actor.IsAttacking = true
			actor.IsBackAttacking = false
			actor.AttackPart = actor.PendingAttackPart
			var xOffset float32 = 6.0
			if actor.Type == AIActor || !actor.FacingRight {
				xOffset = -6.0
			}
			SetActorAttackBox(actor, xOffset)

			// Устанавливаем флаг текущей атаки-2
			actor.IsCurrentAttack2 = actor.IsPendingAttack2

			// Выбираем анимацию в зависимости от типа атаки
			if actor.IsPendingAttack2 {
				// Сбрасываем анимации атаки-2
				ResetAnimation(&actor.AttackHighAnim2, &actor.AttackMidAnim2, &actor.AttackLowAnim2)
			} else {
				// Сбрасываем обычные анимации атаки
				ResetAnimation(&actor.AttackHighAnim, &actor.AttackMidAnim, &actor.AttackLowAnim)
			}

			actor.AttackTimer = 0
			actor.AttackCooldown = 0.1 // Кулдаун после начала новой атаки
			actor.HasPendingAttack = false
			actor.PendingAttackPart = None
			actor.IsPendingAttack2 = false // Сбрасываем флаг атаки-2
		}
	}
}

// getTransitionAnimation возвращает анимацию перехода между двумя позами
func getTransitionAnimation(actor *Actor, from, to BodyPart) *Animation {
	switch {
	case from == Legs && to == Torso:
		return &actor.UpDownLowToMidAnim
	case from == Torso && to == Head:
		return &actor.UpDownMidToHighAnim
	case from == Head && to == Torso:
		return &actor.UpDownHighToMidAnim
	case from == Torso && to == Legs:
		return &actor.UpDownMidToLowAnim
	default:
		return &actor.UpDownAnim // Возвращаем обычную анимацию если переход не найден
	}
}

// SelectActorAnimation выбирает подходящую анимацию для актора
func SelectActorAnimation(actor *Actor) *Animation {
	if actor.IsHit {
		switch actor.AttackPart {
		case Head:
			return &actor.HitHighAnim
		case Torso:
			return &actor.HitMidAnim
		case Legs:
			return &actor.HitLowAnim
		}
	} else if actor.IsAttacking {
		// Выбираем анимацию в зависимости от того, была ли это атака-2
		if actor.IsCurrentAttack2 {
			switch actor.AttackPart {
			case Head:
				return &actor.AttackHighAnim2
			case Torso:
				return &actor.AttackMidAnim2
			case Legs:
				return &actor.AttackLowAnim2
			}
		} else {
			switch actor.AttackPart {
			case Head:
				return &actor.AttackHighAnim
			case Torso:
				return &actor.AttackMidAnim
			case Legs:
				return &actor.AttackLowAnim
			}
		}
	} else if actor.IsBackAttacking {
		switch actor.AttackPart {
		case Head:
			return &actor.BackAttackHighAnim
		case Torso:
			return &actor.BackAttackMidAnim
		case Legs:
			return &actor.BackAttackLowAnim
		}
	} else if actor.IsMoving {
		// Проверяем скорость для выбора между ходьбой и бегом
		if abs(actor.Velocity) >= WalkToRunThreshold {
			// Бег - используем обычную анимацию бега
			return &actor.RunAnim
		} else {
			// Ходьба - выбираем анимацию в зависимости от положения рук
			switch actor.CurrentPose {
			case Head:
				return &actor.WalkHighAnim
			case Torso:
				return &actor.WalkMidAnim
			case Legs:
				return &actor.WalkLowAnim
			default:
				return &actor.WalkMidAnim
			}
		}
	}

	// --- Логика переходов между позами ---
	if actor.IsTransitioning {
		// Если пользователь хочет новую позу — запоминаем её, но не прерываем текущий переход
		if actor.TargetPart != actor.TransitionTo && actor.TargetPart != None {
			actor.NextTargetPart = actor.TargetPart
		}
		transitionAnim := getTransitionAnimation(actor, actor.TransitionFrom, actor.TransitionTo)
		return transitionAnim
	}

	// После завершения перехода — если есть отложенный переход, запускаем его
	if actor.NextTargetPart != None && actor.NextTargetPart != actor.CurrentPose {
		actor.IsTransitioning = true
		actor.TransitionFrom = actor.CurrentPose
		actor.TransitionTo = actor.NextTargetPart
		actor.TransitionTimer = 0
		transitionAnim := getTransitionAnimation(actor, actor.TransitionFrom, actor.TransitionTo)
		transitionAnim.CurrentFrame = 0
		transitionAnim.Timer = 0
		actor.NextTargetPart = None
		return transitionAnim
	}

	// Если поза изменилась (и нет перехода) — запускаем переход
	if actor.TargetPart != None && actor.TargetPart != actor.CurrentPose {
		actor.IsTransitioning = true
		actor.TransitionFrom = actor.CurrentPose
		actor.TransitionTo = actor.TargetPart
		actor.TransitionTimer = 0
		transitionAnim := getTransitionAnimation(actor, actor.TransitionFrom, actor.TransitionTo)
		transitionAnim.CurrentFrame = 0
		transitionAnim.Timer = 0
		return transitionAnim
	}

	// Обычная поза (нет перехода)
	switch actor.CurrentPose {
	case Head:
		actor.UpDownAnim.CurrentFrame = 2 // high
	case Torso:
		actor.UpDownAnim.CurrentFrame = 1 // mid
	case Legs:
		actor.UpDownAnim.CurrentFrame = 0 // low
	default:
		actor.UpDownAnim.CurrentFrame = 1 // mid по умолчанию
	}
	return &actor.UpDownAnim
}

// abs возвращает абсолютное значение float32
func abs(x float32) float32 {
	if x < 0 {
		return -x
	}
	return x
}

// sign возвращает знак числа (-1, 0, 1)
func sign(x float32) float32 {
	if x > 0 {
		return 1
	} else if x < 0 {
		return -1
	}
	return 0
}

// UpdateActorAnimation обновляет анимацию актора
func UpdateActorAnimation(actor *Actor) {
	anim := SelectActorAnimation(actor)
	anim.Timer += rl.GetFrameTime()
	if anim.Timer >= anim.FrameTime {
		anim.Timer = 0
		anim.CurrentFrame = (anim.CurrentFrame + 1) % anim.FrameCount

		// Если это анимация ходьбы, синхронизируем нижнюю часть
		if anim == &actor.WalkHighAnim || anim == &actor.WalkMidAnim || anim == &actor.WalkLowAnim {
			actor.WalkAnim.CurrentFrame = anim.CurrentFrame
			actor.WalkAnim.Timer = anim.Timer
		}

		// Если это анимация перехода и она завершилась
		if actor.IsTransitioning && anim.CurrentFrame == 0 {
			actor.IsTransitioning = false
			actor.CurrentPose = actor.TransitionTo
			actor.TransitionTimer = 0
		}
	}

	// Обновляем таймер перехода
	if actor.IsTransitioning {
		actor.TransitionTimer += rl.GetFrameTime()
		if actor.TransitionTimer >= actor.TransitionDuration {
			actor.IsTransitioning = false
			actor.CurrentPose = actor.TransitionTo
			actor.TransitionTimer = 0
		}
	}
}

// CheckActorAttack проверяет атаку актора
func CheckActorAttack(attacker, defender *Actor) {
	if attacker.IsAttacking {
		anim := SelectActorAnimation(attacker)

		// Урон наносится только на 6-м кадре анимации атаки (CurrentFrame == 5)
		if anim.CurrentFrame == 5 {
			// Проверяем коллизию AttackBox с хитбоксами защитника
			if rl.CheckCollisionRecs(attacker.AttackBox, defender.HeadBox) ||
				rl.CheckCollisionRecs(attacker.AttackBox, defender.TorsoBox) ||
				rl.CheckCollisionRecs(attacker.AttackBox, defender.LegsBox) {

				// Наносим урон
				defender.Health -= 10
				if defender.Health < 0 {
					defender.Health = 0
				}

				// Активируем анимацию получения урона
				defender.IsHit = true
				defender.HitTimer = 0
				defender.AttackPart = attacker.AttackPart

				// Сбрасываем ожидающую атаку при получении урона
				defender.HasPendingAttack = false
				defender.PendingAttackPart = None

				// Выбираем правильную анимацию получения урона
				switch attacker.AttackPart {
				case Head:
					ResetAnimation(&defender.HitHighAnim)
				case Torso:
					ResetAnimation(&defender.HitMidAnim)
				case Legs:
					ResetAnimation(&defender.HitLowAnim)
				}
			}
		}
	}
}

// DrawActor отрисовывает актора
func DrawActor(actor *Actor, opponent *Actor, color rl.Color, camera *rl.Camera2D) {
	anim := SelectActorAnimation(actor)
	var flip float32 = 1.0
	if !actor.FacingRight {
		flip = -1.0
	}

	// Проверяем, является ли это анимацией ходьбы
	if anim == &actor.WalkHighAnim || anim == &actor.WalkMidAnim || anim == &actor.WalkLowAnim {
		// Отрисовываем анимацию ходьбы с разделением на верхнюю и нижнюю части
		// Сначала рисуем нижнюю часть (ноги) - WalkAnim
		walkSourceRect := rl.NewRectangle(float32(actor.WalkAnim.CurrentFrame)*actor.WalkAnim.FrameWidth, 0, actor.WalkAnim.FrameWidth*flip, float32(actor.WalkAnim.Texture.Height))
		walkDestRect := rl.NewRectangle(actor.Position.X+actor.Size.X/2, actor.Position.Y+actor.Size.Y/2, actor.WalkAnim.FrameWidth*SpriteScale, float32(actor.WalkAnim.Texture.Height)*SpriteScale)
		walkOrigin := rl.NewVector2(actor.WalkAnim.FrameWidth*SpriteScale/2, float32(actor.WalkAnim.Texture.Height)*SpriteScale/2)
		rl.DrawTexturePro(actor.WalkAnim.Texture, walkSourceRect, walkDestRect, walkOrigin, 0, rl.White)

		// Затем рисуем верхнюю часть (руки) - выбранная анимация ходьбы
		upperSourceRect := rl.NewRectangle(float32(anim.CurrentFrame)*anim.FrameWidth, 0, anim.FrameWidth*flip, float32(anim.Texture.Height))
		upperDestRect := rl.NewRectangle(actor.Position.X+actor.Size.X/2, actor.Position.Y+actor.Size.Y/2, anim.FrameWidth*SpriteScale, float32(anim.Texture.Height)*SpriteScale)
		upperOrigin := rl.NewVector2(anim.FrameWidth*SpriteScale/2, float32(anim.Texture.Height)*SpriteScale/2)
		rl.DrawTexturePro(anim.Texture, upperSourceRect, upperDestRect, upperOrigin, 0, rl.White)
	} else {
		// Обычная отрисовка для всех остальных анимаций
		sourceRect := rl.NewRectangle(float32(anim.CurrentFrame)*anim.FrameWidth, 0, anim.FrameWidth*flip, float32(anim.Texture.Height))
		destRect := rl.NewRectangle(actor.Position.X+actor.Size.X/2, actor.Position.Y+actor.Size.Y/2, anim.FrameWidth*SpriteScale, float32(anim.Texture.Height)*SpriteScale)
		origin := rl.NewVector2(anim.FrameWidth*SpriteScale/2, float32(anim.Texture.Height)*SpriteScale/2)
		rl.DrawTexturePro(anim.Texture, sourceRect, destRect, origin, 0, rl.White)
	}

	// Выводим скорость актора рядом с ним
	speedText := fmt.Sprintf("%.2f", actor.Velocity)
	textWidth := rl.MeasureText(speedText, 16)
	textX := int32(actor.Position.X + actor.Size.X/2 - float32(textWidth)/2)
	textY := int32(actor.Position.Y - 24) // чуть выше головы
	rl.DrawText(speedText, textX, textY, 16, rl.Black)

	// Хитбоксы
	rl.DrawRectangleLinesEx(actor.HeadBox, 1/camera.Zoom, rl.Fade(color, 0.5)) // Толщина линий корректируется
	rl.DrawRectangleLinesEx(actor.TorsoBox, 1/camera.Zoom, rl.Fade(color, 0.5))
	rl.DrawRectangleLinesEx(actor.LegsBox, 1/camera.Zoom, rl.Fade(color, 0.5))

	// Отображаем AttackBox всегда с учётом направления удара
	attackBoxColor := rl.Blue
	if actor.IsAttacking && anim.CurrentFrame >= 5 {
		attackBoxColor = rl.Red
	}
	rl.DrawRectangleLinesEx(actor.AttackBox, 1/camera.Zoom, rl.Fade(attackBoxColor, 0.5))

	if actor.Type == PlayerActor && actor.TargetPart != None && opponent != nil {
		var targetBox rl.Rectangle
		switch actor.TargetPart {
		case Head:
			targetBox = opponent.HeadBox
		case Torso:
			targetBox = opponent.TorsoBox
		case Legs:
			targetBox = opponent.LegsBox
		}
		rl.DrawRectangleRec(targetBox, rl.Fade(color, 0.3))
	}
}

// UnloadActorTextures освобождает текстуры актора
func UnloadActorTextures(actor *Actor) {
	rl.UnloadTexture(actor.IdleAnim.Texture)
	rl.UnloadTexture(actor.WalkAnim.Texture)
	rl.UnloadTexture(actor.WalkHighAnim.Texture)
	rl.UnloadTexture(actor.WalkMidAnim.Texture)
	rl.UnloadTexture(actor.WalkLowAnim.Texture)
	rl.UnloadTexture(actor.RunAnim.Texture)
	rl.UnloadTexture(actor.AttackHighAnim.Texture)
	rl.UnloadTexture(actor.AttackMidAnim.Texture)
	rl.UnloadTexture(actor.AttackLowAnim.Texture)
	rl.UnloadTexture(actor.AttackHighAnim2.Texture)
	rl.UnloadTexture(actor.AttackMidAnim2.Texture)
	rl.UnloadTexture(actor.AttackLowAnim2.Texture)
	rl.UnloadTexture(actor.BackAttackHighAnim.Texture)
	rl.UnloadTexture(actor.BackAttackMidAnim.Texture)
	rl.UnloadTexture(actor.BackAttackLowAnim.Texture)
	rl.UnloadTexture(actor.HitHighAnim.Texture)
	rl.UnloadTexture(actor.HitMidAnim.Texture)
	rl.UnloadTexture(actor.HitLowAnim.Texture)
	rl.UnloadTexture(actor.UpDownAnim.Texture)

	// Освобождаем текстуры анимаций переходов
	rl.UnloadTexture(actor.UpDownLowToMidAnim.Texture)
	rl.UnloadTexture(actor.UpDownMidToHighAnim.Texture)
	rl.UnloadTexture(actor.UpDownHighToMidAnim.Texture)
	rl.UnloadTexture(actor.UpDownMidToLowAnim.Texture)
}

// CalculateEnemyPositionAroundPlayer рассчитывает желаемую позицию врага для beat'em up
// Реализует систему очереди атаки: первые 2 врага атакуют, остальные ждут
func CalculateEnemyPositionAroundPlayer(player *Actor, enemies []Actor, enemyIndex int) rl.Vector2 {
	// Первые 2 врага (индексы 0 и 1) - атакующие позиции слева и справа
	if enemyIndex < 2 {
		side := -1 // -1 = слева, 1 = справа
		if enemyIndex == 1 {
			side = 1
		}
		// Атакующие позиции ближе к игроку
		distanceX := float32(35) * float32(side)
		return rl.NewVector2(player.Position.X+distanceX, player.Position.Y)
	}

	// Остальные враги (индексы 2+) - ожидающие позиции сверху и снизу
	queueIndex := enemyIndex - 2
	side := -1 // -1 = сверху, 1 = снизу
	if queueIndex%2 == 1 {
		side = 1
	}

	// Ожидающие позиции с фиксированными Y координатами для создания перспективы
	var fixedY float32
	if side == -1 {
		// Сверху - фиксированная позиция выше игрока
		fixedY = player.Position.Y - 25
	} else {
		// Снизу - фиксированная позиция ниже игрока
		fixedY = player.Position.Y + 25
	}

	// Небольшое смещение по X для разнесения
	var spreadX float32
	if queueIndex%4 < 2 {
		spreadX = float32((queueIndex/2)*30) * -1
	} else {
		spreadX = float32((queueIndex / 2) * 30)
	}

	return rl.NewVector2(player.Position.X+spreadX, fixedY)
}

// UpdateEnemyAI обновляет ИИ врага для движения к желаемой позиции
func UpdateEnemyAI(enemy *Actor, player *Actor, enemies []Actor, enemyIndex int) {
	// Рассчитываем желаемую позицию
	targetPos := CalculateEnemyPositionAroundPlayer(player, enemies, enemyIndex)

	// Определяем направление движения по X
	dx := targetPos.X - enemy.Position.X

	// Устанавливаем целевую скорость
	targetVelocity := float32(0.0)
	acceleration := float32(80.0)  // Уменьшено с 120 для снижения скольжения
	deceleration := float32(120.0) // Уменьшено с 200 для снижения скольжения
	dt := rl.GetFrameTime()

	// Горизонтальное движение
	if abs(dx) > 10 {
		if dx > 0 {
			targetVelocity = enemy.Speed
			enemy.FacingRight = true
		} else {
			targetVelocity = -enemy.Speed
			enemy.FacingRight = false
		}
		enemy.IsMoving = true
	} else {
		targetVelocity = 0
		enemy.IsMoving = false
	}

	// Применяем плавное движение
	if targetVelocity != 0 {
		if (targetVelocity > 0 && enemy.Velocity < targetVelocity) || (targetVelocity < 0 && enemy.Velocity > targetVelocity) {
			enemy.Velocity += acceleration * dt * sign(targetVelocity)
			if (targetVelocity > 0 && enemy.Velocity > targetVelocity) || (targetVelocity < 0 && enemy.Velocity < targetVelocity) {
				enemy.Velocity = targetVelocity
			}
		} else if (targetVelocity > 0 && enemy.Velocity > targetVelocity) || (targetVelocity < 0 && enemy.Velocity < targetVelocity) {
			enemy.Velocity -= deceleration * dt * sign(enemy.Velocity)
			if (targetVelocity > 0 && enemy.Velocity < targetVelocity) || (targetVelocity < 0 && enemy.Velocity > targetVelocity) {
				enemy.Velocity = targetVelocity
			}
		}
	} else {
		// Быстрое торможение
		if enemy.Velocity > 0 {
			enemy.Velocity -= BrakeSpeed * dt
			if enemy.Velocity < 0 {
				enemy.Velocity = 0
			}
		} else if enemy.Velocity < 0 {
			enemy.Velocity += BrakeSpeed * dt
			if enemy.Velocity > 0 {
				enemy.Velocity = 0
			}
		}
	}

	// Все враги двигаются к своим целевым позициям (включая Y)
	// Атакующие враги получают Y игрока, враги в очереди - фиксированные Y позиции

	// Обновляем целевую позицию
	enemy.TargetPosition.X += enemy.Velocity * dt
	enemy.TargetPosition.Y = targetPos.Y // Используем Y из целевой позиции

	// Интерполяция позиции (включая Y для врагов в очереди)
	enemy.Position.X = Lerp(enemy.Position.X, enemy.TargetPosition.X, 0.95)
	enemy.Position.Y = Lerp(enemy.Position.Y, enemy.TargetPosition.Y, 0.95)

	// --- Атака, если враг достаточно близко к игроку ---
	// Только первые 2 врага могут атаковать (система очереди)
	if enemyIndex < 2 {
		// Проверяем расстояние до игрока для начала атаки
		distX := abs(enemy.Position.X - player.Position.X)
		distY := abs(enemy.Position.Y - player.Position.Y)
		if distX < 30 && distY < 20 && !enemy.IsAttacking && !enemy.IsBackAttacking && !enemy.IsHit {
			enemy.IsAttacking = true
			enemy.AttackPart = BodyPart(rand.Intn(3))
			ResetAnimation(&enemy.AttackHighAnim, &enemy.AttackMidAnim, &enemy.AttackLowAnim)
			enemy.AttackTimer = 0
		}
	}
}

// ReorderEnemyQueue перераспределяет очередь врагов, когда один из атакующих падает
func ReorderEnemyQueue(enemies []Actor) {
	// Находим живых врагов
	var aliveEnemies []int
	for i, enemy := range enemies {
		if enemy.Health > 0 {
			aliveEnemies = append(aliveEnemies, i)
		}
	}

	// Если врагов меньше 3, перераспределение не нужно
	if len(aliveEnemies) < 3 {
		return
	}

	// Проверяем, есть ли атакующие позиции (первые 2) свободны
	attackingPositions := 0
	for _, index := range aliveEnemies {
		if index < 2 {
			attackingPositions++
		}
	}

	// Если атакующих позиций меньше 2, перемещаем врага из очереди
	if attackingPositions < 2 {
		// Находим первого врага в очереди (индекс >= 2)
		for _, index := range aliveEnemies {
			if index >= 2 {
				// Перемещаем его в первую свободную атакующую позицию
				if attackingPositions == 0 {
					// Заменяем индекс 0
					enemies[0] = enemies[index]
					enemies[index] = Actor{} // Очищаем старую позицию
				} else if attackingPositions == 1 {
					// Заменяем индекс 1
					enemies[1] = enemies[index]
					enemies[index] = Actor{} // Очищаем старую позицию
				}
				break
			}
		}
	}
}
