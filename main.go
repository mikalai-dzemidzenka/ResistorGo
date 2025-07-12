package main

import (
	"fmt"
	rl "github.com/gen2brain/raylib-go/raylib"
	"math"
)

// Physics constants
const gravity = 800.0
const moveSpeed = 200.0
const jumpForce = -400.0
const wallSlideSpeed = 100.0
const wallJumpForceX = 300.0
const wallJumpForceY = -350.0
const climbSpeed = 100.0
const climbThreshold = 10.0         // Increased for better wall detection
const wallJumpLockoutDuration = 0.2 // Lockout duration for wall jump

// Player represents the player character
type Player struct {
	Position   rl.Vector2
	Velocity   rl.Vector2
	Width      float32
	Height     float32
	OnGround   bool
	OnWall     bool
	OnCeiling  bool
	WallNormal float32 // -1 for left wall, 1 for right wall, 0 if not on wall
}

// Platform represents a static platform
type Platform struct {
	Rect rl.Rectangle
}

// Wall represents a vertical wall
type Wall struct {
	Rect rl.Rectangle
}

// HackableObject represents a hackable object
type HackableObject struct {
	Rect   rl.Rectangle
	Hacked bool
}

// GameState stores game state for time rewinding
type GameState struct {
	PlayerPos       rl.Vector2
	PlayerVel       rl.Vector2
	PlayerOnGround  bool
	PlayerOnWall    bool
	PlayerOnCeiling bool
	WallNormal      float32
	Hackables       []bool
}

func main() {
	// Initialize window
	rl.InitWindow(800, 600, "Gunpoint Clone")
	rl.SetTargetFPS(60)

	// Create player
	player := Player{
		Position:   rl.NewVector2(100, 500),
		Velocity:   rl.NewVector2(0, 0),
		Width:      20,
		Height:     40,
		OnGround:   false,
		OnWall:     false,
		OnCeiling:  false,
		WallNormal: 0,
	}

	// Create platforms (including ceilings)
	platforms := []Platform{
		{Rect: rl.NewRectangle(0, 550, 800, 50)},   // Floor
		{Rect: rl.NewRectangle(200, 400, 200, 20)}, // Platform 1
		{Rect: rl.NewRectangle(500, 300, 150, 20)}, // Platform 2
		{Rect: rl.NewRectangle(300, 200, 200, 20)}, // Ceiling platform
	}

	// Create walls
	walls := []Wall{
		{Rect: rl.NewRectangle(400, 200, 20, 350)}, // Vertical wall 1 (extended to ceiling)
		{Rect: rl.NewRectangle(100, 200, 20, 350)}, // Vertical wall 2 (extended to ceiling)
	}

	// Create hackable objects
	hackables := []HackableObject{
		{Rect: rl.NewRectangle(250, 380, 20, 20), Hacked: false},
		{Rect: rl.NewRectangle(550, 280, 20, 20), Hacked: false},
	}

	// Time rewind variables
	var stateHistory []GameState
	var isRewinding bool
	var currentIndex int = -1
	var wallJumpLockoutTimer float32 = 0 // Timer for locking movement after wall jump

	// Initialize initial state
	initialState := GameState{
		PlayerPos:       player.Position,
		PlayerVel:       player.Velocity,
		PlayerOnGround:  player.OnGround,
		PlayerOnWall:    player.OnWall,
		PlayerOnCeiling: player.OnCeiling,
		WallNormal:      player.WallNormal,
		Hackables:       make([]bool, len(hackables)),
	}
	for i := range hackables {
		initialState.Hackables[i] = hackables[i].Hacked
	}
	stateHistory = append(stateHistory, initialState)

	// Main game loop
	for !rl.WindowShouldClose() {
		dt := rl.GetFrameTime()

		// Update wall jump lockout timer
		if wallJumpLockoutTimer > 0 {
			wallJumpLockoutTimer -= dt
			if wallJumpLockoutTimer < 0 {
				wallJumpLockoutTimer = 0
			}
		}

		if isRewinding {
			if currentIndex > 0 {
				currentIndex--
				setGameState(&player, hackables, stateHistory[currentIndex])
				wallJumpLockoutTimer = 0 // Reset timer during rewind
			} else {
				isRewinding = false
			}
		} else {
			// Handle input
			if rl.IsKeyPressed(rl.KeyR) && len(stateHistory) > 0 {
				isRewinding = true
				currentIndex = len(stateHistory) - 1
				continue
			}

			// Handle attaching to surfaces
			if rl.IsKeyDown(rl.KeyUp) && !player.OnGround && !player.OnWall && !player.OnCeiling {
				// Check for wall first
				for _, wall := range walls {
					near, normal := isNearWall(player, wall)
					if near {
						player.OnWall = true
						player.WallNormal = normal
						if normal == 1 {
							player.Position.X = wall.Rect.X - player.Width
						} else {
							player.Position.X = wall.Rect.X + wall.Rect.Width
						}
						player.Velocity.X = 0
						player.Velocity.Y = 0
						break
					}
				}
				// Check for ceiling if not on wall
				if !player.OnWall {
					for _, platform := range platforms {
						if isNearCeiling(player, platform) {
							player.OnCeiling = true
							player.Position.Y = platform.Rect.Y + platform.Rect.Height
							player.Velocity.Y = 0
							player.Velocity.X = 0
							break
						}
					}
				}
			}

			// Handle jumping
			if rl.IsKeyPressed(rl.KeySpace) {
				if player.OnGround {
					player.Velocity.Y = jumpForce
					player.OnGround = false
				} else if player.OnWall {
					player.Velocity.Y = wallJumpForceY
					player.Velocity.X = -player.WallNormal * wallJumpForceX
					player.OnWall = false
					player.WallNormal = 0
					wallJumpLockoutTimer = wallJumpLockoutDuration
				} else if player.OnCeiling {
					player.OnCeiling = false
					player.Velocity.Y = jumpForce
				}
			}

			// Handle movement
			if player.OnWall {
				if rl.IsKeyDown(rl.KeyUp) {
					player.Velocity.Y = -climbSpeed
				} else if rl.IsKeyDown(rl.KeyDown) {
					player.Velocity.Y = climbSpeed
				} else {
					player.Velocity.Y = wallSlideSpeed // Slide down
				}
				// Update position
				newY := player.Position.Y + player.Velocity.Y*dt
				// Check if still on wall
				stillOnWall := false
				testRect := rl.NewRectangle(player.Position.X, newY, player.Width, player.Height)
				for _, wall := range walls {
					if rl.CheckCollisionRecs(testRect, wall.Rect) && (player.WallNormal == getWallNormal(player.Position.X, wall)) {
						stillOnWall = true
						break
					}
				}
				if stillOnWall {
					player.Position.Y = newY
				} else {
					player.OnWall = false
					player.WallNormal = 0
				}
				// Check for transition to ceiling
				if player.OnWall && rl.IsKeyDown(rl.KeyUp) {
					for _, platform := range platforms {
						if isNearCeiling(player, platform) {
							player.OnWall = false
							player.WallNormal = 0
							player.OnCeiling = true
							player.Position.Y = platform.Rect.Y + platform.Rect.Height
							player.Velocity.Y = 0
							player.Velocity.X = 0
							break
						}
					}
				}
			} else if player.OnCeiling {
				if !rl.IsKeyDown(rl.KeyUp) {
					player.OnCeiling = false
				} else {
					// Allow left/right movement
					if rl.IsKeyDown(rl.KeyA) {
						player.Velocity.X = -moveSpeed
					} else if rl.IsKeyDown(rl.KeyD) {
						player.Velocity.X = moveSpeed
					} else {
						player.Velocity.X = 0
					}
					// Update position
					newX := player.Position.X + player.Velocity.X*dt
					// Check if still on ceiling
					stillOnCeiling := false
					testRect := rl.NewRectangle(newX, player.Position.Y, player.Width, player.Height)
					for _, platform := range platforms {
						if rl.CheckCollisionRecs(testRect, platform.Rect) {
							stillOnCeiling = true
							break
						}
					}
					if stillOnCeiling {
						player.Position.X = newX
					} else {
						player.OnCeiling = false
					}
					// Check for transition to wall
					if player.OnCeiling {
						for _, wall := range walls {
							near, normal := isNearWall(player, wall)
							if near {
								player.OnCeiling = false
								player.OnWall = true
								player.WallNormal = normal
								if normal == 1 {
									player.Position.X = wall.Rect.X - player.Width
								} else {
									player.Position.X = wall.Rect.X + wall.Rect.Width
								}
								player.Velocity.X = 0
								player.Velocity.Y = 0
								break
							}
						}
					}
				}
			} else {
				// Normal movement
				if wallJumpLockoutTimer == 0 {
					player.Velocity.X = 0
					if rl.IsKeyDown(rl.KeyA) {
						player.Velocity.X = -moveSpeed
					}
					if rl.IsKeyDown(rl.KeyD) {
						player.Velocity.X = moveSpeed
					}
				}
				player.Velocity.Y += gravity * dt
				player.Position.X += player.Velocity.X * dt
				player.Position.Y += player.Velocity.Y * dt

				// Handle collisions with platforms for OnGround
				player.OnGround = false
				playerRect := rl.NewRectangle(player.Position.X, player.Position.Y, player.Width, player.Height)
				for _, platform := range platforms {
					if rl.CheckCollisionRecs(playerRect, platform.Rect) {
						if player.Velocity.Y > 0 {
							player.Position.Y = platform.Rect.Y - player.Height
							player.Velocity.Y = 0
							player.OnGround = true
							player.OnWall = false
							player.OnCeiling = false
							player.WallNormal = 0
						} else if player.Velocity.Y < 0 {
							player.Position.Y = platform.Rect.Y + platform.Rect.Height
							player.Velocity.Y = 0
						}
					}
				}

				// Handle collisions with walls for normal movement
				for _, wall := range walls {
					if rl.CheckCollisionRecs(playerRect, wall.Rect) {
						if player.Velocity.X > 0 {
							player.Position.X = wall.Rect.X - player.Width
						} else if player.Velocity.X < 0 {
							player.Position.X = wall.Rect.X + wall.Rect.Width
						}
						player.Velocity.X = 0
					}
				}
			}

			// Handle hacking
			if rl.IsKeyPressed(rl.KeyE) {
				playerRect := rl.NewRectangle(player.Position.X, player.Position.Y, player.Width, player.Height)
				for i, obj := range hackables {
					if rl.CheckCollisionRecs(playerRect, obj.Rect) {
						hackables[i].Hacked = !hackables[i].Hacked
					}
				}
			}

			// Store current state for rewinding
			currentState := GameState{
				PlayerPos:       player.Position,
				PlayerVel:       player.Velocity,
				PlayerOnGround:  player.OnGround,
				PlayerOnWall:    player.OnWall,
				PlayerOnCeiling: player.OnCeiling,
				WallNormal:      player.WallNormal,
				Hackables:       make([]bool, len(hackables)),
			}
			for i := range hackables {
				currentState.Hackables[i] = hackables[i].Hacked
			}
			stateHistory = append(stateHistory, currentState)
			if len(stateHistory) > 600 {
				stateHistory = stateHistory[1:] // Limit history to 600 frames (~15 KB)
			}
		}

		// Render
		rl.BeginDrawing()
		rl.ClearBackground(rl.Black)

		// Draw platforms
		for _, platform := range platforms {
			rl.DrawRectangleRec(platform.Rect, rl.Gray)
			rl.DrawRectangleLines(int32(platform.Rect.X), int32(platform.Rect.Y), int32(platform.Rect.Width), int32(platform.Rect.Height), rl.Green)
		}

		// Draw walls and climb threshold zones
		for _, wall := range walls {
			rl.DrawRectangleRec(wall.Rect, rl.DarkGray)
			rl.DrawRectangleLines(int32(wall.Rect.X), int32(wall.Rect.Y), int32(wall.Rect.Width), int32(wall.Rect.Height), rl.Red)
			// Draw climb threshold zones
			rl.DrawRectangleLines(int32(wall.Rect.X-climbThreshold), int32(wall.Rect.Y), int32(wall.Rect.Width+2*climbThreshold), int32(wall.Rect.Height), rl.Yellow)
		}

		// Draw hackable objects
		for _, obj := range hackables {
			color := rl.Red
			if obj.Hacked {
				color = rl.Green
			}
			rl.DrawRectangleRec(obj.Rect, color)
		}

		// Draw player
		playerRect := rl.NewRectangle(player.Position.X, player.Position.Y, player.Width, player.Height)
		rl.DrawRectangleRec(playerRect, rl.Blue)

		// Draw debug information
		nearWall := false
		var wallNormal float32
		var distanceToWall float32
		for _, wall := range walls {
			if near, normal := isNearWall(player, wall); near {
				nearWall = true
				wallNormal = normal
				if normal == 1 {
					distanceToWall = float32(math.Abs(float64((player.Position.X + player.Width) - wall.Rect.X)))
				} else {
					distanceToWall = float32(math.Abs(float64(player.Position.X - (wall.Rect.X + wall.Rect.Width))))
				}
				break
			}
		}
		rl.DrawText(fmt.Sprintf("Player Pos: (%.1f, %.1f)", player.Position.X, player.Position.Y), 10, 30, 20, rl.White)
		rl.DrawText(fmt.Sprintf("OnGround: %v", player.OnGround), 10, 50, 20, rl.White)
		rl.DrawText(fmt.Sprintf("OnWall: %v, WallNormal: %.1f", player.OnWall, player.WallNormal), 10, 70, 20, rl.White)
		rl.DrawText(fmt.Sprintf("OnCeiling: %v", player.OnCeiling), 10, 90, 20, rl.White)
		rl.DrawText(fmt.Sprintf("NearWall: %v, Normal: %.1f, Distance: %.1f", nearWall, wallNormal, distanceToWall), 10, 110, 20, rl.White)
		rl.DrawText(fmt.Sprintf("Up key: %v", rl.IsKeyDown(rl.KeyUp)), 10, 130, 20, rl.White)
		rl.DrawText(fmt.Sprintf("WallJumpLockout: %.2f", wallJumpLockoutTimer), 10, 150, 20, rl.White)
		rl.DrawText("A/D - Move, Space - Jump/Wall Jump, E - Hack, R - Rewind, Up - Climb", 10, 10, 20, rl.White)

		rl.EndDrawing()

		// Stop rewinding
		if rl.IsKeyReleased(rl.KeyR) && isRewinding {
			isRewinding = false
			if currentIndex < len(stateHistory)-1 {
				stateHistory = stateHistory[:currentIndex+1]
			}
			wallJumpLockoutTimer = 0 // Reset timer after rewind
		}
	}

	rl.CloseWindow()
}

// Helper functions
func isNearCeiling(player Player, platform Platform) bool {
	return math.Abs(float64(player.Position.Y-(platform.Rect.Y+platform.Rect.Height))) < float64(climbThreshold) &&
		player.Position.X < platform.Rect.X+platform.Rect.Width &&
		player.Position.X+player.Width > platform.Rect.X
}

func isNearWall(player Player, wall Wall) (bool, float32) {
	if math.Abs(float64((player.Position.X+player.Width)-wall.Rect.X)) < float64(climbThreshold) &&
		player.Position.Y < wall.Rect.Y+wall.Rect.Height &&
		player.Position.Y+player.Height > wall.Rect.Y {
		return true, 1 // Wall on right
	}
	if math.Abs(float64(player.Position.X-(wall.Rect.X+wall.Rect.Width))) < float64(climbThreshold) &&
		player.Position.Y < wall.Rect.Y+wall.Rect.Height &&
		player.Position.Y+player.Height > wall.Rect.Y {
		return true, -1 // Wall on left
	}
	return false, 0
}

func getWallNormal(playerX float32, wall Wall) float32 {
	if playerX < wall.Rect.X {
		return 1 // Right of wall
	}
	return -1 // Left of wall
}

func setGameState(player *Player, hackables []HackableObject, state GameState) {
	player.Position = state.PlayerPos
	player.Velocity = state.PlayerVel
	player.OnGround = state.PlayerOnGround
	player.OnWall = state.PlayerOnWall
	player.OnCeiling = state.PlayerOnCeiling
	player.WallNormal = state.WallNormal
	for i := range hackables {
		hackables[i].Hacked = state.Hackables[i]
	}
}
