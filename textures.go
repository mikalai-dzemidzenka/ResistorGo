package main

import rl "github.com/gen2brain/raylib-go/raylib"

const (
	TEXTURE_PLAYER_IDLE = iota
	TEXTURE_PLAYER_IDLE_1
	TEXTURE_PLAYER_IDLE_2
	TEXTURE_PLAYER_IDLE_3
	TEXTURE_PLAYER_IDLE_4
	TEXTURE_PLAYER_LEN
)

const (
	TEXTURE_BLOCK_1 = 400 + iota
	TEXTURE_BLOCK_2
)

var Textures [TEXTURE_PLAYER_LEN]rl.Texture2D
