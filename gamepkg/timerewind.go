package gamepkg

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

// TimeState представляет состояние игрока в определенный момент времени
type TimeState struct {
	Position          rl.Vector2
	TargetPosition    rl.Vector2
	Velocity          float32
	Health            int
	IsAttacking       bool
	AttackPart        BodyPart
	AttackTimer       float32
	IsBackAttacking   bool
	BackAttackTimer   float32
	IsHit             bool
	HitTimer          float32
	IsMoving          bool
	FacingRight       bool
	HasPendingAttack  bool
	PendingAttackPart BodyPart
	IsPendingAttack2  bool
	IsCurrentAttack2  bool
	AttackCooldown    float32
	LastAttackTime    float32

	// Анимации (текущий кадр и таймер для каждой)
	IdleAnimFrame           int
	IdleAnimTimer           float32
	RunAnimFrame            int
	RunAnimTimer            float32
	AttackHighAnimFrame     int
	AttackHighAnimTimer     float32
	AttackMidAnimFrame      int
	AttackMidAnimTimer      float32
	AttackLowAnimFrame      int
	AttackLowAnimTimer      float32
	AttackHighAnim2Frame    int
	AttackHighAnim2Timer    float32
	AttackMidAnim2Frame     int
	AttackMidAnim2Timer     float32
	AttackLowAnim2Frame     int
	AttackLowAnim2Timer     float32
	BackAttackHighAnimFrame int
	BackAttackHighAnimTimer float32
	BackAttackMidAnimFrame  int
	BackAttackMidAnimTimer  float32
	BackAttackLowAnimFrame  int
	BackAttackLowAnimTimer  float32
	HitHighAnimFrame        int
	HitHighAnimTimer        float32
	HitMidAnimFrame         int
	HitMidAnimTimer         float32
	HitLowAnimFrame         int
	HitLowAnimTimer         float32
}

// TimeRewindSystem управляет системой перемотки времени
type TimeRewindSystem struct {
	States       []TimeState
	MaxStates    int
	CurrentIndex int
	IsRewinding  bool
	RewindSpeed  float32
	RewindTimer  float32
}

// NewTimeRewindSystem создает новую систему перемотки времени
func NewTimeRewindSystem(maxStates int) *TimeRewindSystem {
	return &TimeRewindSystem{
		States:       make([]TimeState, 0, maxStates),
		MaxStates:    maxStates,
		CurrentIndex: -1,
		IsRewinding:  false,
		RewindSpeed:  1.0, // Скорость перемотки (1.0 = нормальная скорость)
		RewindTimer:  0,
	}
}

// SaveState сохраняет текущее состояние игрока
func (trs *TimeRewindSystem) SaveState(actor *Actor) {
	state := TimeState{
		Position:          actor.Position,
		TargetPosition:    actor.TargetPosition,
		Velocity:          actor.Velocity,
		Health:            actor.Health,
		IsAttacking:       actor.IsAttacking,
		AttackPart:        actor.AttackPart,
		AttackTimer:       actor.AttackTimer,
		IsBackAttacking:   actor.IsBackAttacking,
		BackAttackTimer:   actor.BackAttackTimer,
		IsHit:             actor.IsHit,
		HitTimer:          actor.HitTimer,
		IsMoving:          actor.IsMoving,
		FacingRight:       actor.FacingRight,
		HasPendingAttack:  actor.HasPendingAttack,
		PendingAttackPart: actor.PendingAttackPart,
		IsPendingAttack2:  actor.IsPendingAttack2,
		IsCurrentAttack2:  actor.IsCurrentAttack2,
		AttackCooldown:    actor.AttackCooldown,
		LastAttackTime:    actor.LastAttackTime,

		IdleAnimFrame:           actor.IdleAnim.CurrentFrame,
		IdleAnimTimer:           actor.IdleAnim.Timer,
		RunAnimFrame:            actor.RunAnim.CurrentFrame,
		RunAnimTimer:            actor.RunAnim.Timer,
		AttackHighAnimFrame:     actor.AttackHighAnim.CurrentFrame,
		AttackHighAnimTimer:     actor.AttackHighAnim.Timer,
		AttackMidAnimFrame:      actor.AttackMidAnim.CurrentFrame,
		AttackMidAnimTimer:      actor.AttackMidAnim.Timer,
		AttackLowAnimFrame:      actor.AttackLowAnim.CurrentFrame,
		AttackLowAnimTimer:      actor.AttackLowAnim.Timer,
		AttackHighAnim2Frame:    actor.AttackHighAnim2.CurrentFrame,
		AttackHighAnim2Timer:    actor.AttackHighAnim2.Timer,
		AttackMidAnim2Frame:     actor.AttackMidAnim2.CurrentFrame,
		AttackMidAnim2Timer:     actor.AttackMidAnim2.Timer,
		AttackLowAnim2Frame:     actor.AttackLowAnim2.CurrentFrame,
		AttackLowAnim2Timer:     actor.AttackLowAnim2.Timer,
		BackAttackHighAnimFrame: actor.BackAttackHighAnim.CurrentFrame,
		BackAttackHighAnimTimer: actor.BackAttackHighAnim.Timer,
		BackAttackMidAnimFrame:  actor.BackAttackMidAnim.CurrentFrame,
		BackAttackMidAnimTimer:  actor.BackAttackMidAnim.Timer,
		BackAttackLowAnimFrame:  actor.BackAttackLowAnim.CurrentFrame,
		BackAttackLowAnimTimer:  actor.BackAttackLowAnim.Timer,
		HitHighAnimFrame:        actor.HitHighAnim.CurrentFrame,
		HitHighAnimTimer:        actor.HitHighAnim.Timer,
		HitMidAnimFrame:         actor.HitMidAnim.CurrentFrame,
		HitMidAnimTimer:         actor.HitMidAnim.Timer,
		HitLowAnimFrame:         actor.HitLowAnim.CurrentFrame,
		HitLowAnimTimer:         actor.HitLowAnim.Timer,
	}

	// Обрезание истории теперь не происходит автоматически!
	// Добавляем новое состояние
	trs.States = append(trs.States, state)
	trs.CurrentIndex++

	// Ограничиваем количество состояний
	if len(trs.States) > trs.MaxStates {
		trs.States = trs.States[1:]
		trs.CurrentIndex--
	}
}

// LoadState загружает состояние игрока из определенного момента времени
func (trs *TimeRewindSystem) LoadState(actor *Actor, index int) {
	if index < 0 || index >= len(trs.States) {
		return
	}

	state := trs.States[index]
	actor.Position = state.Position
	actor.TargetPosition = state.TargetPosition
	actor.Velocity = state.Velocity
	actor.Health = state.Health
	actor.IsAttacking = state.IsAttacking
	actor.AttackPart = state.AttackPart
	actor.AttackTimer = state.AttackTimer
	actor.IsBackAttacking = state.IsBackAttacking
	actor.BackAttackTimer = state.BackAttackTimer
	actor.IsHit = state.IsHit
	actor.HitTimer = state.HitTimer
	actor.IsMoving = state.IsMoving
	actor.FacingRight = state.FacingRight
	actor.HasPendingAttack = state.HasPendingAttack
	actor.PendingAttackPart = state.PendingAttackPart
	actor.IsPendingAttack2 = state.IsPendingAttack2
	actor.IsCurrentAttack2 = state.IsCurrentAttack2
	actor.AttackCooldown = state.AttackCooldown
	actor.LastAttackTime = state.LastAttackTime

	actor.IdleAnim.CurrentFrame = state.IdleAnimFrame
	actor.IdleAnim.Timer = state.IdleAnimTimer
	actor.RunAnim.CurrentFrame = state.RunAnimFrame
	actor.RunAnim.Timer = state.RunAnimTimer
	actor.AttackHighAnim.CurrentFrame = state.AttackHighAnimFrame
	actor.AttackHighAnim.Timer = state.AttackHighAnimTimer
	actor.AttackMidAnim.CurrentFrame = state.AttackMidAnimFrame
	actor.AttackMidAnim.Timer = state.AttackMidAnimTimer
	actor.AttackLowAnim.CurrentFrame = state.AttackLowAnimFrame
	actor.AttackLowAnim.Timer = state.AttackLowAnimTimer
	actor.AttackHighAnim2.CurrentFrame = state.AttackHighAnim2Frame
	actor.AttackHighAnim2.Timer = state.AttackHighAnim2Timer
	actor.AttackMidAnim2.CurrentFrame = state.AttackMidAnim2Frame
	actor.AttackMidAnim2.Timer = state.AttackMidAnim2Timer
	actor.AttackLowAnim2.CurrentFrame = state.AttackLowAnim2Frame
	actor.AttackLowAnim2.Timer = state.AttackLowAnim2Timer
	actor.BackAttackHighAnim.CurrentFrame = state.BackAttackHighAnimFrame
	actor.BackAttackHighAnim.Timer = state.BackAttackHighAnimTimer
	actor.BackAttackMidAnim.CurrentFrame = state.BackAttackMidAnimFrame
	actor.BackAttackMidAnim.Timer = state.BackAttackMidAnimTimer
	actor.BackAttackLowAnim.CurrentFrame = state.BackAttackLowAnimFrame
	actor.BackAttackLowAnim.Timer = state.BackAttackLowAnimTimer
	actor.HitHighAnim.CurrentFrame = state.HitHighAnimFrame
	actor.HitHighAnim.Timer = state.HitHighAnimTimer
	actor.HitMidAnim.CurrentFrame = state.HitMidAnimFrame
	actor.HitMidAnim.Timer = state.HitMidAnimTimer
	actor.HitLowAnim.CurrentFrame = state.HitLowAnimFrame
	actor.HitLowAnim.Timer = state.HitLowAnimTimer
}

// StartRewind начинает перемотку времени назад
func (trs *TimeRewindSystem) StartRewind() {
	if trs.CurrentIndex > 0 {
		trs.IsRewinding = true
		trs.RewindTimer = 0
	}
}

// StopRewind останавливает перемотку времени
func (trs *TimeRewindSystem) StopRewind() {
	trs.IsRewinding = false
}

// UpdateRewind обновляет перемотку времени
func (trs *TimeRewindSystem) UpdateRewind(actor *Actor) {
	if !trs.IsRewinding {
		return
	}

	trs.RewindTimer += rl.GetFrameTime() * trs.RewindSpeed

	// Каждые 0.016 секунды (60 FPS) перематываем на один кадр назад
	if trs.RewindTimer >= 0.016 {
		trs.RewindTimer = 0

		if trs.CurrentIndex > 0 {
			trs.CurrentIndex--
			trs.LoadState(actor, trs.CurrentIndex)
		} else {
			// Достигли начала истории
			trs.IsRewinding = false
		}
	}
}

// GetRewindProgress возвращает прогресс перемотки (0.0 - 1.0)
func (trs *TimeRewindSystem) GetRewindProgress() float32 {
	if len(trs.States) == 0 {
		return 0.0
	}
	return float32(trs.CurrentIndex) / float32(len(trs.States)-1)
}

// CanRewind проверяет, можно ли перематывать время назад
func (trs *TimeRewindSystem) CanRewind() bool {
	return trs.CurrentIndex > 0
}

// ClearHistory очищает историю состояний
func (trs *TimeRewindSystem) ClearHistory() {
	trs.States = make([]TimeState, 0, trs.MaxStates)
	trs.CurrentIndex = -1
	trs.IsRewinding = false
}

// CutHistoryToCurrent обрезает историю до текущего состояния (использовать после перемотки и нового действия)
func (trs *TimeRewindSystem) CutHistoryToCurrent() {
	if trs.CurrentIndex < len(trs.States)-1 {
		trs.States = trs.States[:trs.CurrentIndex+1]
	}
}
