package gamepkg

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// AnimationID определяет тип анимации
type AnimationID int

const (
	AnimIdle AnimationID = iota
	AnimRun
	AnimAttackHigh
	AnimAttackMid
	AnimAttackLow
	AnimAttackHigh2
	AnimAttackMid2
	AnimAttackLow2
	AnimBackAttackHigh
	AnimBackAttackMid
	AnimBackAttackLow
	AnimHitHigh
	AnimHitMid
	AnimHitLow
	AnimUpDownLowToMid
	AnimUpDownMidToHigh
	AnimUpDownHighToMid
	AnimUpDownMidToLow
	AnimCount // Количество анимаций (для создания слайса)
)

// AnimationState представляет состояние одной анимации
type AnimationState struct {
	CurrentFrame int
	Timer        float32
}

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

	// Оптимизированное хранение анимаций через слайс (более эффективно чем map)
	Animations []AnimationState
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

// getAnimationState возвращает состояние анимации из слайса
func getAnimationState(animations []AnimationState, id AnimationID) AnimationState {
	if int(id) < len(animations) {
		return animations[id]
	}
	return AnimationState{CurrentFrame: 0, Timer: 0}
}

// setAnimationState устанавливает состояние анимации в слайс
func setAnimationState(animations []AnimationState, id AnimationID, currentFrame int, timer float32) {
	if int(id) < len(animations) {
		animations[id] = AnimationState{
			CurrentFrame: currentFrame,
			Timer:        timer,
		}
	}
}

// SaveState сохраняет текущее состояние игрока
func (trs *TimeRewindSystem) SaveState(actor *Actor) {
	// Создаем слайс для анимаций (фиксированный размер)
	animations := make([]AnimationState, AnimCount)

	// Сохраняем состояние всех анимаций
	setAnimationState(animations, AnimIdle, actor.IdleAnim.CurrentFrame, actor.IdleAnim.Timer)
	setAnimationState(animations, AnimRun, actor.RunAnim.CurrentFrame, actor.RunAnim.Timer)
	setAnimationState(animations, AnimAttackHigh, actor.AttackHighAnim.CurrentFrame, actor.AttackHighAnim.Timer)
	setAnimationState(animations, AnimAttackMid, actor.AttackMidAnim.CurrentFrame, actor.AttackMidAnim.Timer)
	setAnimationState(animations, AnimAttackLow, actor.AttackLowAnim.CurrentFrame, actor.AttackLowAnim.Timer)
	setAnimationState(animations, AnimAttackHigh2, actor.AttackHighAnim2.CurrentFrame, actor.AttackHighAnim2.Timer)
	setAnimationState(animations, AnimAttackMid2, actor.AttackMidAnim2.CurrentFrame, actor.AttackMidAnim2.Timer)
	setAnimationState(animations, AnimAttackLow2, actor.AttackLowAnim2.CurrentFrame, actor.AttackLowAnim2.Timer)
	setAnimationState(animations, AnimBackAttackHigh, actor.BackAttackHighAnim.CurrentFrame, actor.BackAttackHighAnim.Timer)
	setAnimationState(animations, AnimBackAttackMid, actor.BackAttackMidAnim.CurrentFrame, actor.BackAttackMidAnim.Timer)
	setAnimationState(animations, AnimBackAttackLow, actor.BackAttackLowAnim.CurrentFrame, actor.BackAttackLowAnim.Timer)
	setAnimationState(animations, AnimHitHigh, actor.HitHighAnim.CurrentFrame, actor.HitHighAnim.Timer)
	setAnimationState(animations, AnimHitMid, actor.HitMidAnim.CurrentFrame, actor.HitMidAnim.Timer)
	setAnimationState(animations, AnimHitLow, actor.HitLowAnim.CurrentFrame, actor.HitLowAnim.Timer)
	setAnimationState(animations, AnimUpDownLowToMid, actor.UpDownLowToMidAnim.CurrentFrame, actor.UpDownLowToMidAnim.Timer)
	setAnimationState(animations, AnimUpDownMidToHigh, actor.UpDownMidToHighAnim.CurrentFrame, actor.UpDownMidToHighAnim.Timer)
	setAnimationState(animations, AnimUpDownHighToMid, actor.UpDownHighToMidAnim.CurrentFrame, actor.UpDownHighToMidAnim.Timer)
	setAnimationState(animations, AnimUpDownMidToLow, actor.UpDownMidToLowAnim.CurrentFrame, actor.UpDownMidToLowAnim.Timer)

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
		Animations:        animations,
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

	// Отладочная информация каждые 60 кадров (1 секунда)
	if trs.CurrentIndex%60 == 0 {
		fmt.Printf("SAVE STATE: Index %d/%d, States %d\n", trs.CurrentIndex, len(trs.States)-1, len(trs.States))
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

	// Загружаем состояния анимаций из слайса
	animState := getAnimationState(state.Animations, AnimIdle)
	actor.IdleAnim.CurrentFrame = animState.CurrentFrame
	actor.IdleAnim.Timer = animState.Timer

	animState = getAnimationState(state.Animations, AnimRun)
	actor.RunAnim.CurrentFrame = animState.CurrentFrame
	actor.RunAnim.Timer = animState.Timer

	animState = getAnimationState(state.Animations, AnimAttackHigh)
	actor.AttackHighAnim.CurrentFrame = animState.CurrentFrame
	actor.AttackHighAnim.Timer = animState.Timer

	animState = getAnimationState(state.Animations, AnimAttackMid)
	actor.AttackMidAnim.CurrentFrame = animState.CurrentFrame
	actor.AttackMidAnim.Timer = animState.Timer

	animState = getAnimationState(state.Animations, AnimAttackLow)
	actor.AttackLowAnim.CurrentFrame = animState.CurrentFrame
	actor.AttackLowAnim.Timer = animState.Timer

	animState = getAnimationState(state.Animations, AnimAttackHigh2)
	actor.AttackHighAnim2.CurrentFrame = animState.CurrentFrame
	actor.AttackHighAnim2.Timer = animState.Timer

	animState = getAnimationState(state.Animations, AnimAttackMid2)
	actor.AttackMidAnim2.CurrentFrame = animState.CurrentFrame
	actor.AttackMidAnim2.Timer = animState.Timer

	animState = getAnimationState(state.Animations, AnimAttackLow2)
	actor.AttackLowAnim2.CurrentFrame = animState.CurrentFrame
	actor.AttackLowAnim2.Timer = animState.Timer

	animState = getAnimationState(state.Animations, AnimBackAttackHigh)
	actor.BackAttackHighAnim.CurrentFrame = animState.CurrentFrame
	actor.BackAttackHighAnim.Timer = animState.Timer

	animState = getAnimationState(state.Animations, AnimBackAttackMid)
	actor.BackAttackMidAnim.CurrentFrame = animState.CurrentFrame
	actor.BackAttackMidAnim.Timer = animState.Timer

	animState = getAnimationState(state.Animations, AnimBackAttackLow)
	actor.BackAttackLowAnim.CurrentFrame = animState.CurrentFrame
	actor.BackAttackLowAnim.Timer = animState.Timer

	animState = getAnimationState(state.Animations, AnimHitHigh)
	actor.HitHighAnim.CurrentFrame = animState.CurrentFrame
	actor.HitHighAnim.Timer = animState.Timer

	animState = getAnimationState(state.Animations, AnimHitMid)
	actor.HitMidAnim.CurrentFrame = animState.CurrentFrame
	actor.HitMidAnim.Timer = animState.Timer

	animState = getAnimationState(state.Animations, AnimHitLow)
	actor.HitLowAnim.CurrentFrame = animState.CurrentFrame
	actor.HitLowAnim.Timer = animState.Timer

	animState = getAnimationState(state.Animations, AnimUpDownLowToMid)
	actor.UpDownLowToMidAnim.CurrentFrame = animState.CurrentFrame
	actor.UpDownLowToMidAnim.Timer = animState.Timer

	animState = getAnimationState(state.Animations, AnimUpDownMidToHigh)
	actor.UpDownMidToHighAnim.CurrentFrame = animState.CurrentFrame
	actor.UpDownMidToHighAnim.Timer = animState.Timer

	animState = getAnimationState(state.Animations, AnimUpDownHighToMid)
	actor.UpDownHighToMidAnim.CurrentFrame = animState.CurrentFrame
	actor.UpDownHighToMidAnim.Timer = animState.Timer

	animState = getAnimationState(state.Animations, AnimUpDownMidToLow)
	actor.UpDownMidToLowAnim.CurrentFrame = animState.CurrentFrame
	actor.UpDownMidToLowAnim.Timer = animState.Timer
}

// StartRewind начинает перемотку времени назад
func (trs *TimeRewindSystem) StartRewind() {
	trs.IsRewinding = true
	trs.RewindTimer = 0
	trs.RewindSpeed = 0 // Начинаем с остановленной скорости
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

	if trs.RewindSpeed == 0 {
		return
	}

	trs.RewindTimer += rl.GetFrameTime() * trs.RewindSpeed

	// Используем цикл, чтобы перематывать сразу несколько кадров при большой скорости
	for trs.RewindTimer >= 0.016 || trs.RewindTimer <= -0.016 {
		if trs.RewindSpeed > 0 {
			// Движение вперед
			if trs.CurrentIndex < len(trs.States)-1 {
				trs.CurrentIndex++
				trs.LoadState(actor, trs.CurrentIndex)
				trs.RewindTimer -= 0.016
			} else {
				trs.RewindSpeed = 0
				trs.RewindTimer = 0
				break
			}
		} else if trs.RewindSpeed < 0 {
			// Движение назад
			if trs.CurrentIndex > 0 {
				trs.CurrentIndex--
				trs.LoadState(actor, trs.CurrentIndex)
				trs.RewindTimer += 0.016
			} else {
				trs.RewindSpeed = 0
				trs.RewindTimer = 0
				break
			}
		} else {
			break
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
