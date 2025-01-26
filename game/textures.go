package game

import rl "github.com/gen2brain/raylib-go/raylib"

const (
	TEXTURE_PLAYER_IDLE = iota
	TEXTURE_PLAYER_IDLE_1
	TEXTURE_PLAYER_IDLE_2
	TEXTURE_PLAYER_IDLE_3
	TEXTURE_PLAYER_IDLE_4

	TEXTURE_BLOCK_1
	TEXTURE_BG_1
	TEXTURE_PLATFORM_1

	TEXTURE_LEN
)

var Textures [TEXTURE_LEN]rl.Texture2D

func LoadTextures() {
	Textures[TEXTURE_PLAYER_IDLE] = rl.LoadTexture("resources/idle.png")
	Textures[TEXTURE_PLAYER_IDLE_1] = rl.LoadTexture("resources/idle-1.png")
	Textures[TEXTURE_PLAYER_IDLE_2] = rl.LoadTexture("resources/idle-2.png")
	Textures[TEXTURE_PLAYER_IDLE_3] = rl.LoadTexture("resources/idle-3.png")
	Textures[TEXTURE_PLAYER_IDLE_4] = rl.LoadTexture("resources/idle-4.png")

	Textures[TEXTURE_BLOCK_1] = rl.LoadTexture("resources/e1m1_bg_wall1.png")
	Textures[TEXTURE_BG_1] = rl.LoadTexture("resources/e1m1_bg1.png")
	Textures[TEXTURE_PLATFORM_1] = rl.LoadTexture("resources/e1m1_platform1.png")
}
