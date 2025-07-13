package gamepkg

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

func ResetAnimation(anims ...*Animation) {
	for _, anim := range anims {
		anim.Timer = 0
		anim.CurrentFrame = 0
	}
}

// ResetTransitionAnimations сбрасывает анимации переходов для актора
func ResetTransitionAnimations(actor *Actor) {
	actor.UpDownLowToMidAnim.CurrentFrame = 0
	actor.UpDownLowToMidAnim.Timer = 0
	actor.UpDownMidToHighAnim.CurrentFrame = 0
	actor.UpDownMidToHighAnim.Timer = 0
	actor.UpDownHighToMidAnim.CurrentFrame = 0
	actor.UpDownHighToMidAnim.Timer = 0
	actor.UpDownMidToLowAnim.CurrentFrame = 0
	actor.UpDownMidToLowAnim.Timer = 0
}

func SelectAnimation(player *Player) *Animation {
	if player.IsHit {
		switch player.AttackPart {
		case Head:
			return &player.HitHighAnim
		case Torso:
			return &player.HitMidAnim
		case Legs:
			return &player.HitLowAnim
		}
	} else if player.IsAttacking {
		switch player.AttackPart {
		case Head:
			return &player.AttackHighAnim
		case Torso:
			return &player.AttackMidAnim
		case Legs:
			return &player.AttackLowAnim
		}
	} else if player.IsBackAttacking {
		switch player.AttackPart {
		case Head:
			return &player.BackAttackHighAnim
		case Torso:
			return &player.BackAttackMidAnim
		case Legs:
			return &player.BackAttackLowAnim
		}
	} else if player.IsMoving {
		return &player.RunAnim
	}
	return &player.IdleAnim
}

func UpdateAnimation(player *Player) {
	anim := SelectAnimation(player)
	anim.Timer += rl.GetFrameTime()
	if anim.Timer >= anim.FrameTime {
		anim.Timer = 0
		anim.CurrentFrame = (anim.CurrentFrame + 1) % anim.FrameCount
	}
}
