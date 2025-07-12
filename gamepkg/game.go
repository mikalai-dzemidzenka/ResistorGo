package gamepkg

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

func CheckAttack(attacker, defender *Player, isHuman bool) {
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

		anim := SelectAnimation(attacker)
		if attacker.AttackPart == defender.TargetPart {
			attacker.IsAttacking = false
			attacker.IsBackAttacking = true
			attacker.BackAttackTimer = 0
			ResetAnimation(&attacker.BackAttackHighAnim, &attacker.BackAttackMidAnim, &attacker.BackAttackLowAnim)
			if isHuman && attacker.HasPendingAttack {
				attacker.IsAttacking = true
				attacker.IsBackAttacking = false
				attacker.AttackPart = attacker.PendingAttackPart
				attacker.HasPendingAttack = false
				attacker.PendingAttackPart = None
				var xOffset float32 = AttackBoxWidth / 2
				if !attacker.FacingRight {
					xOffset = -AttackBoxWidth / 2
				}
				SetAttackBox(attacker, xOffset)
				ResetAnimation(&attacker.AttackHighAnim, &attacker.AttackMidAnim, &attacker.AttackLowAnim)
				attacker.AttackTimer = 0
			}
		} else if anim.CurrentFrame == 5 && rl.CheckCollisionRecs(attacker.AttackBox, defenderBox) {
			defender.Health -= 10
			if defender.Health < 0 {
				defender.Health = 0
			}
			defender.IsHit = true
			defender.HitTimer = 0
			defender.AttackPart = attacker.AttackPart
			ResetAnimation(&defender.HitHighAnim, &defender.HitMidAnim, &defender.HitLowAnim)
		}
	}
}

func DrawPlayer(player *Player, ai *Player, color rl.Color, isHuman bool, camera *rl.Camera2D) {
	anim := SelectAnimation(player)
	var flip float32 = 1.0
	if !player.FacingRight {
		flip = -1.0
	}

	sourceRect := rl.NewRectangle(float32(anim.CurrentFrame)*anim.FrameWidth, 0, anim.FrameWidth*flip, float32(anim.Texture.Height))
	destRect := rl.NewRectangle(player.Position.X+player.Size.X/2, player.Position.Y+player.Size.Y/2, anim.FrameWidth*SpriteScale, float32(anim.Texture.Height)*SpriteScale)
	origin := rl.NewVector2(anim.FrameWidth*SpriteScale/2, float32(anim.Texture.Height)*SpriteScale/2)
	rl.DrawTexturePro(anim.Texture, sourceRect, destRect, origin, 0, rl.White)

	rl.DrawRectangleLinesEx(player.HeadBox, 1/camera.Zoom, rl.Fade(color, 0.5)) // Толщина линий корректируется
	rl.DrawRectangleLinesEx(player.TorsoBox, 1/camera.Zoom, rl.Fade(color, 0.5))
	rl.DrawRectangleLinesEx(player.LegsBox, 1/camera.Zoom, rl.Fade(color, 0.5))

	// Отображаем AttackBox всегда с учётом направления удара
	attackBoxColor := rl.Blue
	if player.IsAttacking && anim.CurrentFrame >= 5 {
		attackBoxColor = rl.Red
	}
	rl.DrawRectangleLinesEx(player.AttackBox, 1/camera.Zoom, rl.Fade(attackBoxColor, 0.5))

	if isHuman && player.TargetPart != None && ai != nil {
		var targetBox rl.Rectangle
		switch player.TargetPart {
		case Head:
			targetBox = ai.HeadBox
		case Torso:
			targetBox = ai.TorsoBox
		case Legs:
			targetBox = ai.LegsBox
		}
		rl.DrawRectangleRec(targetBox, rl.Fade(color, 0.3))
	}
}

func UnloadTextures(players []*Player) {
	for _, player := range players {
		rl.UnloadTexture(player.IdleAnim.Texture)
		rl.UnloadTexture(player.RunAnim.Texture)
		rl.UnloadTexture(player.AttackHighAnim.Texture)
		rl.UnloadTexture(player.AttackMidAnim.Texture)
		rl.UnloadTexture(player.AttackLowAnim.Texture)
		rl.UnloadTexture(player.BackAttackHighAnim.Texture)
		rl.UnloadTexture(player.BackAttackMidAnim.Texture)
		rl.UnloadTexture(player.BackAttackLowAnim.Texture)
		rl.UnloadTexture(player.HitHighAnim.Texture)
		rl.UnloadTexture(player.HitMidAnim.Texture)
		rl.UnloadTexture(player.HitLowAnim.Texture)
	}
}
