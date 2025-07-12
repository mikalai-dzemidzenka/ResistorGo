package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"math/rand"
	"strconv"
	"time"
)

const (
	screenWidth         = 800
	screenHeight        = 450
	spriteSize          = 48
	spriteScale         = 1.5
	gravity             = 0.0
	jumpSpeed           = 0.0
	attackFps           = 24
	attackFrameTime     = 0.45 / 8               // 0.45 секунды на 8 кадров (атака)
	backAttackFrameTime = float32(4) / attackFps // 4 кадра для возврата (~0.1667 сек)
	hitFrameTime        = 0.45 / 4               // 0.45 секунды на 4 кадра (урон, 12 FPS)
	runFps              = 18
	runFrameTime        = 1.0 / runFps
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
	Size               rl.Vector2
	Speed              float32
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

func main() {
	rl.InitWindow(screenWidth, screenHeight, "Fist Fight with Animations")
	rl.SetTargetFPS(60)
	rand.Seed(time.Now().UnixNano())

	player := Player{
		Position:           rl.NewVector2(100, screenHeight-90*spriteScale),
		Size:               rl.NewVector2(48*spriteScale, 48*spriteScale),
		Speed:              5.0,
		Health:             100,
		TargetPart:         None,
		HeadBox:            rl.NewRectangle(0, 0, 9*spriteScale, 10*spriteScale),
		TorsoBox:           rl.NewRectangle(0, 0, 19*spriteScale, 12*spriteScale),
		LegsBox:            rl.NewRectangle(0, 0, 16*spriteScale, 22*spriteScale),
		IsAttacking:        false,
		AttackPart:         None,
		PendingAttackPart:  None,
		HasPendingAttack:   false,
		AttackBox:          rl.NewRectangle(0, 0, 50*spriteScale, 30*spriteScale),
		IsMoving:           false,
		FacingRight:        true,
		IsHit:              false,
		HitTimer:           0,
		HitDuration:        0.45,
		IdleAnim:           Animation{Texture: rl.LoadTexture("resources/threefour/idle.png"), FrameCount: 1, FrameWidth: 48, FrameTime: 0.1},
		RunAnim:            Animation{Texture: rl.LoadTexture("resources/threefour/run.png"), FrameCount: 10, FrameWidth: 48, FrameTime: runFrameTime},
		AttackHighAnim:     Animation{Texture: rl.LoadTexture("resources/threefour/attack-high.png"), FrameCount: 8, FrameWidth: 48, FrameTime: attackFrameTime},
		AttackMidAnim:      Animation{Texture: rl.LoadTexture("resources/threefour/attack-mid.png"), FrameCount: 8, FrameWidth: 48, FrameTime: attackFrameTime},
		AttackLowAnim:      Animation{Texture: rl.LoadTexture("resources/threefour/attack-low.png"), FrameCount: 8, FrameWidth: 48, FrameTime: attackFrameTime},
		BackAttackHighAnim: Animation{Texture: rl.LoadTexture("resources/threefour/back-attack-top-1.png"), FrameCount: 4, FrameWidth: 48, FrameTime: backAttackFrameTime},
		BackAttackMidAnim:  Animation{Texture: rl.LoadTexture("resources/threefour/back-attack-mid-1.png"), FrameCount: 4, FrameWidth: 48, FrameTime: backAttackFrameTime},
		BackAttackLowAnim:  Animation{Texture: rl.LoadTexture("resources/threefour/back-attack-low-1.png"), FrameCount: 4, FrameWidth: 48, FrameTime: backAttackFrameTime},
		HitHighAnim:        Animation{Texture: rl.LoadTexture("resources/threefour/hit-top.png"), FrameCount: 4, FrameWidth: 48, FrameTime: hitFrameTime},
		HitMidAnim:         Animation{Texture: rl.LoadTexture("resources/threefour/hit-mid.png"), FrameCount: 4, FrameWidth: 48, FrameTime: hitFrameTime},
		HitLowAnim:         Animation{Texture: rl.LoadTexture("resources/threefour/hit-low.png"), FrameCount: 4, FrameWidth: 48, FrameTime: hitFrameTime},
		AttackDuration:     0.45,
		AttackTimer:        0,
		BackAttackDuration: float32(4) / attackFps,
		BackAttackTimer:    0,
	}

	ai := Player{
		Position:           rl.NewVector2(screenWidth-140, screenHeight-90*spriteScale),
		Size:               rl.NewVector2(48*spriteScale, 48*spriteScale),
		Speed:              4.0,
		Health:             100,
		TargetPart:         None,
		HeadBox:            rl.NewRectangle(0, 0, 9*spriteScale, 10*spriteScale),
		TorsoBox:           rl.NewRectangle(0, 0, 19*spriteScale, 12*spriteScale),
		LegsBox:            rl.NewRectangle(0, 0, 16*spriteScale, 22*spriteScale),
		IsAttacking:        false,
		AttackPart:         None,
		PendingAttackPart:  None,
		HasPendingAttack:   false,
		AttackBox:          rl.NewRectangle(0, 0, 50*spriteScale, 30*spriteScale),
		IsMoving:           false,
		FacingRight:        false,
		IsHit:              false,
		HitTimer:           0,
		HitDuration:        0.45,
		IdleAnim:           Animation{Texture: rl.LoadTexture("resources/threefour/idle.png"), FrameCount: 1, FrameWidth: 48, FrameTime: 0.1},
		RunAnim:            Animation{Texture: rl.LoadTexture("resources/threefour/run.png"), FrameCount: 10, FrameWidth: 48, FrameTime: runFrameTime},
		AttackHighAnim:     Animation{Texture: rl.LoadTexture("resources/threefour/attack-high.png"), FrameCount: 8, FrameWidth: 48, FrameTime: attackFrameTime},
		AttackMidAnim:      Animation{Texture: rl.LoadTexture("resources/threefour/attack-mid.png"), FrameCount: 8, FrameWidth: 48, FrameTime: attackFrameTime},
		AttackLowAnim:      Animation{Texture: rl.LoadTexture("resources/threefour/attack-low.png"), FrameCount: 8, FrameWidth: 48, FrameTime: attackFrameTime},
		BackAttackHighAnim: Animation{Texture: rl.LoadTexture("resources/threefour/back-attack-top-1.png"), FrameCount: 4, FrameWidth: 48, FrameTime: backAttackFrameTime},
		BackAttackMidAnim:  Animation{Texture: rl.LoadTexture("resources/threefour/back-attack-mid-1.png"), FrameCount: 4, FrameWidth: 48, FrameTime: backAttackFrameTime},
		BackAttackLowAnim:  Animation{Texture: rl.LoadTexture("resources/threefour/back-attack-low-1.png"), FrameCount: 4, FrameWidth: 48, FrameTime: backAttackFrameTime},
		HitHighAnim:        Animation{Texture: rl.LoadTexture("resources/threefour/hit-top.png"), FrameCount: 4, FrameWidth: 48, FrameTime: hitFrameTime},
		HitMidAnim:         Animation{Texture: rl.LoadTexture("resources/threefour/hit-mid.png"), FrameCount: 4, FrameWidth: 48, FrameTime: hitFrameTime},
		HitLowAnim:         Animation{Texture: rl.LoadTexture("resources/threefour/hit-low.png"), FrameCount: 4, FrameWidth: 48, FrameTime: hitFrameTime},
		AttackDuration:     0.45,
		AttackTimer:        0,
		BackAttackDuration: float32(4) / attackFps,
		BackAttackTimer:    0,
	}

	aiAttackInterval := float32(1.5)

	for !rl.WindowShouldClose() {
		updatePlayer(&player, &ai, true)
		updateAI(&ai, &player, aiAttackInterval)

		updateAnimation(&player)
		updateAnimation(&ai)

		checkAttack(&player, &ai, true)
		checkAttack(&ai, &player, false)

		rl.BeginDrawing()
		rl.ClearBackground(rl.RayWhite)

		drawPlayer(&player, &ai, rl.Red, true)
		drawPlayer(&ai, nil, rl.Blue, false)

		rl.DrawText("Player HP: "+strconv.Itoa(player.Health), 10, 10, 20, rl.Black)
		rl.DrawText("AI HP: "+strconv.Itoa(ai.Health), screenWidth-150, 10, 20, rl.Black)

		rl.EndDrawing()
	}

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
	rl.UnloadTexture(ai.IdleAnim.Texture)
	rl.UnloadTexture(ai.RunAnim.Texture)
	rl.UnloadTexture(ai.AttackHighAnim.Texture)
	rl.UnloadTexture(ai.AttackMidAnim.Texture)
	rl.UnloadTexture(ai.AttackLowAnim.Texture)
	rl.UnloadTexture(ai.BackAttackHighAnim.Texture)
	rl.UnloadTexture(ai.BackAttackMidAnim.Texture)
	rl.UnloadTexture(ai.BackAttackLowAnim.Texture)
	rl.UnloadTexture(ai.HitHighAnim.Texture)
	rl.UnloadTexture(ai.HitMidAnim.Texture)
	rl.UnloadTexture(ai.HitLowAnim.Texture)
	rl.CloseWindow()
}

func updatePlayer(player *Player, ai *Player, isHuman bool) {
	if player.IsHit {
		player.HitTimer += rl.GetFrameTime()
		if player.HitTimer >= player.HitDuration {
			player.IsHit = false
			player.HitTimer = 0
		}
		return
	}

	if player.IsAttacking {
		player.AttackTimer += rl.GetFrameTime()
		if player.AttackTimer >= player.AttackDuration {
			player.IsAttacking = false
			player.IsBackAttacking = true
			player.BackAttackTimer = 0
			player.BackAttackHighAnim.Timer = 0
			player.BackAttackMidAnim.Timer = 0
			player.BackAttackLowAnim.Timer = 0
			player.BackAttackHighAnim.CurrentFrame = 0
			player.BackAttackMidAnim.CurrentFrame = 0
			player.BackAttackLowAnim.CurrentFrame = 0
			if player.HasPendingAttack {
				player.IsAttacking = true
				player.IsBackAttacking = false
				player.AttackPart = player.PendingAttackPart
				player.HasPendingAttack = false
				player.PendingAttackPart = None
				player.AttackBox = rl.NewRectangle(player.Position.X+player.Size.X, player.Position.Y+float32(player.AttackPart)*30*spriteScale, 50*spriteScale, 30*spriteScale)
				player.AttackHighAnim.CurrentFrame = 0
				player.AttackMidAnim.CurrentFrame = 0
				player.AttackLowAnim.CurrentFrame = 0
				player.AttackHighAnim.Timer = 0
				player.AttackMidAnim.Timer = 0
				player.AttackLowAnim.Timer = 0
				player.AttackTimer = 0
			}
		}
		return
	}

	if player.IsBackAttacking {
		player.BackAttackTimer += rl.GetFrameTime()
		if player.BackAttackTimer >= player.BackAttackDuration {
			player.IsBackAttacking = false
			player.AttackPart = None
			player.BackAttackTimer = 0
			if player.HasPendingAttack {
				player.IsAttacking = true
				player.AttackPart = player.PendingAttackPart
				player.HasPendingAttack = false
				player.PendingAttackPart = None
				player.AttackBox = rl.NewRectangle(player.Position.X+player.Size.X, player.Position.Y+float32(player.AttackPart)*30*spriteScale, 50*spriteScale, 30*spriteScale)
				player.AttackHighAnim.CurrentFrame = 0
				player.AttackMidAnim.CurrentFrame = 0
				player.AttackLowAnim.CurrentFrame = 0
				player.AttackHighAnim.Timer = 0
				player.AttackMidAnim.Timer = 0
				player.AttackLowAnim.Timer = 0
				player.AttackTimer = 0
			}
		}
		return
	}

	player.IsMoving = false

	if rl.IsKeyDown(rl.KeyA) {
		player.Position.X -= player.Speed
		player.IsMoving = true
		player.FacingRight = false
	}
	if rl.IsKeyDown(rl.KeyD) {
		player.Position.X += player.Speed
		player.IsMoving = true
		player.FacingRight = true
	}

	centerX := player.Position.X + player.Size.X/2
	centerY := player.Position.Y + player.Size.Y/2
	player.HeadBox.X = centerX - (9*spriteScale)/2 - 1*spriteScale
	player.HeadBox.Y = centerY - (10*spriteScale)/2 - 15*spriteScale
	player.TorsoBox.X = centerX - (19*spriteScale)/2 - 1*spriteScale
	player.TorsoBox.Y = centerY - (12*spriteScale)/2 - 4*spriteScale
	player.LegsBox.X = centerX - (16*spriteScale)/2 - 1*spriteScale
	player.LegsBox.Y = centerY - (22*spriteScale)/2 + 13*spriteScale

	if isHuman && ai != nil {
		mousePos := rl.GetMousePosition()
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
				player.AttackBox = rl.NewRectangle(player.Position.X+player.Size.X, player.Position.Y+float32(player.AttackPart)*30*spriteScale, 50*spriteScale, 30*spriteScale)
				player.AttackHighAnim.CurrentFrame = 0
				player.AttackMidAnim.CurrentFrame = 0
				player.AttackLowAnim.CurrentFrame = 0
				player.AttackHighAnim.Timer = 0
				player.AttackMidAnim.Timer = 0
				player.AttackLowAnim.Timer = 0
				player.AttackTimer = 0
			}
		}
	}

	if player.Position.X < 0 {
		player.Position.X = 0
	}
	if player.Position.X > screenWidth-player.Size.X {
		player.Position.X = screenWidth - player.Size.X
	}
}

func updateAI(ai *Player, player *Player, attackInterval float32) {
	if ai.IsHit {
		ai.HitTimer += rl.GetFrameTime()
		if ai.HitTimer >= ai.HitDuration {
			ai.IsHit = false
			ai.HitTimer = 0
		}
		return
	}

	if ai.IsAttacking {
		ai.AttackTimer += rl.GetFrameTime()
		if ai.AttackTimer >= ai.AttackDuration {
			ai.IsAttacking = false
			ai.IsBackAttacking = true
			ai.BackAttackTimer = 0
			ai.BackAttackHighAnim.Timer = 0
			ai.BackAttackMidAnim.Timer = 0
			ai.BackAttackLowAnim.Timer = 0
			ai.BackAttackHighAnim.CurrentFrame = 0
			ai.BackAttackMidAnim.CurrentFrame = 0
			ai.BackAttackLowAnim.CurrentFrame = 0
		}
		return
	}

	if ai.IsBackAttacking {
		ai.BackAttackTimer += rl.GetFrameTime()
		if ai.BackAttackTimer >= ai.BackAttackDuration {
			ai.IsBackAttacking = false
			ai.AttackPart = None
			ai.BackAttackTimer = 0
		}
		return
	}

	ai.IsMoving = false

	if player.Position.X > ai.Position.X+50 {
		ai.Position.X += ai.Speed
		ai.IsMoving = true
		ai.FacingRight = true
	} else if player.Position.X < ai.Position.X-50 {
		ai.Position.X -= ai.Speed
		ai.IsMoving = true
		ai.FacingRight = false
	}

	centerX := ai.Position.X + ai.Size.X/2
	centerY := ai.Position.Y + ai.Size.Y/2
	ai.HeadBox.X = centerX - (9*spriteScale)/2 - 1*spriteScale
	ai.HeadBox.Y = centerY - (10*spriteScale)/2 - 15*spriteScale
	ai.TorsoBox.X = centerX - (19*spriteScale)/2 - 1*spriteScale
	ai.TorsoBox.Y = centerY - (12*spriteScale)/2 - 4*spriteScale
	ai.LegsBox.X = centerX - (16*spriteScale)/2 - 1*spriteScale
	ai.LegsBox.Y = centerY - (22*spriteScale)/2 + 13*spriteScale

	ai.LastAttackTime += rl.GetFrameTime()
	if ai.LastAttackTime >= attackInterval {
		ai.LastAttackTime = 0
		ai.IsAttacking = true
		ai.AttackPart = BodyPart(rand.Intn(3))
		ai.AttackBox = rl.NewRectangle(ai.Position.X-50*spriteScale, ai.Position.Y+float32(ai.AttackPart)*30*spriteScale, 50*spriteScale, 30*spriteScale)
		ai.TargetPart = BodyPart(rand.Intn(3))
		ai.AttackHighAnim.CurrentFrame = 0
		ai.AttackMidAnim.CurrentFrame = 0
		ai.AttackLowAnim.CurrentFrame = 0
		ai.AttackHighAnim.Timer = 0
		ai.AttackMidAnim.Timer = 0
		ai.AttackLowAnim.Timer = 0
		ai.AttackTimer = 0
	}

	if ai.Position.X < 0 {
		ai.Position.X = 0
	}
	if ai.Position.X > screenWidth-ai.Size.X {
		ai.Position.X = screenWidth - ai.Size.X
	}
}

func updateAnimation(player *Player) {
	var anim *Animation
	if player.IsHit {
		switch player.AttackPart {
		case Head:
			anim = &player.HitHighAnim
		case Torso:
			anim = &player.HitMidAnim
		case Legs:
			anim = &player.HitLowAnim
		default:
			anim = &player.IdleAnim
		}
	} else if player.IsAttacking {
		switch player.AttackPart {
		case Head:
			anim = &player.AttackHighAnim
		case Torso:
			anim = &player.AttackMidAnim
		case Legs:
			anim = &player.AttackLowAnim
		default:
			anim = &player.IdleAnim
		}
	} else if player.IsBackAttacking {
		switch player.AttackPart {
		case Head:
			anim = &player.BackAttackHighAnim
		case Torso:
			anim = &player.BackAttackMidAnim
		case Legs:
			anim = &player.BackAttackLowAnim
		default:
			anim = &player.IdleAnim
		}
	} else if player.IsMoving {
		anim = &player.RunAnim
	} else {
		anim = &player.IdleAnim
	}

	anim.Timer += rl.GetFrameTime()
	if anim.Timer >= anim.FrameTime {
		anim.Timer = 0
		anim.CurrentFrame = (anim.CurrentFrame + 1) % anim.FrameCount
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

		if attacker.AttackPart == defender.TargetPart {
			attacker.IsAttacking = false
			attacker.IsBackAttacking = true
			attacker.BackAttackTimer = 0
			attacker.BackAttackHighAnim.Timer = 0
			attacker.BackAttackMidAnim.Timer = 0
			attacker.BackAttackLowAnim.Timer = 0
			attacker.BackAttackHighAnim.CurrentFrame = 0
			attacker.BackAttackMidAnim.CurrentFrame = 0
			attacker.BackAttackLowAnim.CurrentFrame = 0
			if isHuman && attacker.HasPendingAttack {
				attacker.IsAttacking = true
				attacker.IsBackAttacking = false
				attacker.AttackPart = attacker.PendingAttackPart
				attacker.HasPendingAttack = false
				attacker.PendingAttackPart = None
				attacker.AttackBox = rl.NewRectangle(attacker.Position.X+attacker.Size.X, attacker.Position.Y+float32(attacker.AttackPart)*30*spriteScale, 50*spriteScale, 30*spriteScale)
				attacker.AttackHighAnim.CurrentFrame = 0
				attacker.AttackMidAnim.CurrentFrame = 0
				attacker.AttackLowAnim.CurrentFrame = 0
				attacker.AttackHighAnim.Timer = 0
				attacker.AttackMidAnim.Timer = 0
				attacker.AttackLowAnim.Timer = 0
				attacker.AttackTimer = 0
			}
		} else if rl.CheckCollisionRecs(attacker.AttackBox, defenderBox) {
			defender.Health -= 10
			if defender.Health < 0 {
				defender.Health = 0
			}
			defender.IsHit = true
			defender.HitTimer = 0
			defender.AttackPart = attacker.AttackPart
			defender.HitHighAnim.CurrentFrame = 0
			defender.HitMidAnim.CurrentFrame = 0
			defender.HitLowAnim.CurrentFrame = 0
			defender.HitHighAnim.Timer = 0
			defender.HitMidAnim.Timer = 0
			defender.HitLowAnim.Timer = 0
		}
	}
}

func drawPlayer(player *Player, ai *Player, color rl.Color, isHuman bool) {
	var anim *Animation
	var flip float32 = 1.0
	if !player.FacingRight {
		flip = -1.0
	}
	if player.IsHit {
		switch player.AttackPart {
		case Head:
			anim = &player.HitHighAnim
		case Torso:
			anim = &player.HitMidAnim
		case Legs:
			anim = &player.HitLowAnim
		default:
			anim = &player.IdleAnim
		}
	} else if player.IsAttacking {
		switch player.AttackPart {
		case Head:
			anim = &player.AttackHighAnim
		case Torso:
			anim = &player.AttackMidAnim
		case Legs:
			anim = &player.AttackLowAnim
		default:
			anim = &player.IdleAnim
		}
	} else if player.IsBackAttacking {
		switch player.AttackPart {
		case Head:
			anim = &player.BackAttackHighAnim
		case Torso:
			anim = &player.BackAttackMidAnim
		case Legs:
			anim = &player.BackAttackLowAnim
		default:
			anim = &player.IdleAnim
		}
	} else if player.IsMoving {
		anim = &player.RunAnim
	} else {
		anim = &player.IdleAnim
	}

	sourceRect := rl.NewRectangle(float32(anim.CurrentFrame)*anim.FrameWidth, 0, anim.FrameWidth*flip, float32(anim.Texture.Height))
	destRect := rl.NewRectangle(player.Position.X+player.Size.X/2, player.Position.Y+player.Size.Y/2, anim.FrameWidth*spriteScale, float32(anim.Texture.Height)*spriteScale)
	origin := rl.NewVector2(anim.FrameWidth*spriteScale/2, float32(anim.Texture.Height)*spriteScale/2)
	rl.DrawTexturePro(anim.Texture, sourceRect, destRect, origin, 0, rl.White)

	rl.DrawRectangleLinesEx(player.HeadBox, 1, rl.Fade(color, 0.5))
	rl.DrawRectangleLinesEx(player.TorsoBox, 1, rl.Fade(color, 0.5))
	rl.DrawRectangleLinesEx(player.LegsBox, 1, rl.Fade(color, 0.5))

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
