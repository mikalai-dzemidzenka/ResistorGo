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
// Теперь хранит только активную анимацию (ID, кадр, таймер)
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

	ActiveAnimID AnimationID
	AnimFrame    int
	AnimTimer    float32
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
	// Определяем активную анимацию
	anim := SelectActorAnimation(actor)
	var animID AnimationID
	switch anim {
	case &actor.IdleAnim:
		animID = AnimIdle
	case &actor.RunAnim:
		animID = AnimRun
	case &actor.AttackHighAnim:
		animID = AnimAttackHigh
	case &actor.AttackMidAnim:
		animID = AnimAttackMid
	case &actor.AttackLowAnim:
		animID = AnimAttackLow
	case &actor.AttackHighAnim2:
		animID = AnimAttackHigh2
	case &actor.AttackMidAnim2:
		animID = AnimAttackMid2
	case &actor.AttackLowAnim2:
		animID = AnimAttackLow2
	case &actor.BackAttackHighAnim:
		animID = AnimBackAttackHigh
	case &actor.BackAttackMidAnim:
		animID = AnimBackAttackMid
	case &actor.BackAttackLowAnim:
		animID = AnimBackAttackLow
	case &actor.HitHighAnim:
		animID = AnimHitHigh
	case &actor.HitMidAnim:
		animID = AnimHitMid
	case &actor.HitLowAnim:
		animID = AnimHitLow
	case &actor.UpDownLowToMidAnim:
		animID = AnimUpDownLowToMid
	case &actor.UpDownMidToHighAnim:
		animID = AnimUpDownMidToHigh
	case &actor.UpDownHighToMidAnim:
		animID = AnimUpDownHighToMid
	case &actor.UpDownMidToLowAnim:
		animID = AnimUpDownMidToLow
	default:
		animID = AnimIdle
	}

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
		ActiveAnimID:      animID,
		AnimFrame:         anim.CurrentFrame,
		AnimTimer:         anim.Timer,
	}

	trs.States = append(trs.States, state)
	trs.CurrentIndex++

	if len(trs.States) > trs.MaxStates {
		trs.States = trs.States[1:]
		trs.CurrentIndex--
	}

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

	// Восстанавливаем только активную анимацию, остальные не трогаем
	switch state.ActiveAnimID {
	case AnimIdle:
		actor.IdleAnim.CurrentFrame = state.AnimFrame
		actor.IdleAnim.Timer = state.AnimTimer
	case AnimRun:
		actor.RunAnim.CurrentFrame = state.AnimFrame
		actor.RunAnim.Timer = state.AnimTimer
	case AnimAttackHigh:
		actor.AttackHighAnim.CurrentFrame = state.AnimFrame
		actor.AttackHighAnim.Timer = state.AnimTimer
	case AnimAttackMid:
		actor.AttackMidAnim.CurrentFrame = state.AnimFrame
		actor.AttackMidAnim.Timer = state.AnimTimer
	case AnimAttackLow:
		actor.AttackLowAnim.CurrentFrame = state.AnimFrame
		actor.AttackLowAnim.Timer = state.AnimTimer
	case AnimAttackHigh2:
		actor.AttackHighAnim2.CurrentFrame = state.AnimFrame
		actor.AttackHighAnim2.Timer = state.AnimTimer
	case AnimAttackMid2:
		actor.AttackMidAnim2.CurrentFrame = state.AnimFrame
		actor.AttackMidAnim2.Timer = state.AnimTimer
	case AnimAttackLow2:
		actor.AttackLowAnim2.CurrentFrame = state.AnimFrame
		actor.AttackLowAnim2.Timer = state.AnimTimer
	case AnimBackAttackHigh:
		actor.BackAttackHighAnim.CurrentFrame = state.AnimFrame
		actor.BackAttackHighAnim.Timer = state.AnimTimer
	case AnimBackAttackMid:
		actor.BackAttackMidAnim.CurrentFrame = state.AnimFrame
		actor.BackAttackMidAnim.Timer = state.AnimTimer
	case AnimBackAttackLow:
		actor.BackAttackLowAnim.CurrentFrame = state.AnimFrame
		actor.BackAttackLowAnim.Timer = state.AnimTimer
	case AnimHitHigh:
		actor.HitHighAnim.CurrentFrame = state.AnimFrame
		actor.HitHighAnim.Timer = state.AnimTimer
	case AnimHitMid:
		actor.HitMidAnim.CurrentFrame = state.AnimFrame
		actor.HitMidAnim.Timer = state.AnimTimer
	case AnimHitLow:
		actor.HitLowAnim.CurrentFrame = state.AnimFrame
		actor.HitLowAnim.Timer = state.AnimTimer
	case AnimUpDownLowToMid:
		actor.UpDownLowToMidAnim.CurrentFrame = state.AnimFrame
		actor.UpDownLowToMidAnim.Timer = state.AnimTimer
	case AnimUpDownMidToHigh:
		actor.UpDownMidToHighAnim.CurrentFrame = state.AnimFrame
		actor.UpDownMidToHighAnim.Timer = state.AnimTimer
	case AnimUpDownHighToMid:
		actor.UpDownHighToMidAnim.CurrentFrame = state.AnimFrame
		actor.UpDownHighToMidAnim.Timer = state.AnimTimer
	case AnimUpDownMidToLow:
		actor.UpDownMidToLowAnim.CurrentFrame = state.AnimFrame
		actor.UpDownMidToLowAnim.Timer = state.AnimTimer
	}
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
