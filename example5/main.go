package main

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	screenWidth  = 960
	screenHeight = 640

	tileSize = 4

	mapWidth  = 240
	mapHeight = 160

	gravity   = 800.0
	moveSpeed = 90.0
	jumpSpeed = 250.0

	ambient = 0.025
)

type Player struct {
	Position rl.Vector2
	Velocity rl.Vector2

	Width  float32
	Height float32

	OnGround bool
}

type Light struct {
	X float64
	Y float64

	// Направление в радианах.
	Angle float64

	// Половина угла конуса.
	// Например 30 градусов => полный конус 60°.
	ConeAngle float64

	Radius    float64
	Intensity float64
}

var world [mapHeight][mapWidth]uint8

// ============================================================
// WORLD
// ============================================================

func buildWorld() {
	for y := 0; y < mapHeight; y++ {
		for x := 0; x < mapWidth; x++ {
			if x == 0 ||
				y == 0 ||
				x == mapWidth-1 ||
				y == mapHeight-1 {
				world[y][x] = 1
			}
		}
	}

	// Ground
	for x := 1; x < mapWidth-1; x++ {
		world[135][x] = 1
		world[136][x] = 1
		world[137][x] = 1
		world[138][x] = 1
	}

	// Platforms
	for x := 15; x < 65; x++ {
		world[108][x] = 1
	}

	for x := 82; x < 135; x++ {
		world[90][x] = 1
	}

	for x := 155; x < 205; x++ {
		world[72][x] = 1
	}

	// Walls
	for y := 108; y < 135; y++ {
		world[y][68] = 1
	}

	for y := 72; y < 90; y++ {
		world[y][145] = 1
	}

	// Big box
	for x := 90; x <= 115; x++ {
		world[115][x] = 1
	}

	for y := 115; y < 135; y++ {
		world[y][90] = 1
		world[y][115] = 1
	}

	// Small blocks
	for x := 170; x < 185; x++ {
		world[120][x] = 1
	}

	for x := 190; x < 205; x++ {
		world[105][x] = 1
	}
}

func solidTile(x, y int) bool {
	if x < 0 ||
		y < 0 ||
		x >= mapWidth ||
		y >= mapHeight {
		return true
	}

	return world[y][x] != 0
}

// ============================================================
// PLAYER
// ============================================================

func updatePlayer(p *Player, dt float32) {
	p.Velocity.X = 0

	if rl.IsKeyDown(rl.KeyA) ||
		rl.IsKeyDown(rl.KeyLeft) {
		p.Velocity.X = -moveSpeed
	}

	if rl.IsKeyDown(rl.KeyD) ||
		rl.IsKeyDown(rl.KeyRight) {
		p.Velocity.X = moveSpeed
	}

	if rl.IsKeyPressed(rl.KeySpace) &&
		p.OnGround {

		p.Velocity.Y = -jumpSpeed
		p.OnGround = false
	}

	p.Velocity.Y += gravity * dt

	if p.Velocity.Y > 500 {
		p.Velocity.Y = 500
	}

	movePlayerX(p, p.Velocity.X*dt)
	movePlayerY(p, p.Velocity.Y*dt)
}

func movePlayerX(p *Player, dx float32) {
	p.Position.X += dx

	left := int(math.Floor(float64(p.Position.X)))
	right := int(math.Floor(float64(
		p.Position.X + p.Width - 0.001,
	)))

	top := int(math.Floor(float64(p.Position.Y)))
	bottom := int(math.Floor(float64(
		p.Position.Y + p.Height - 0.001,
	)))

	if dx > 0 {
		tx := right / tileSize

		for ty := top / tileSize; ty <= bottom/tileSize; ty++ {
			if solidTile(tx, ty) {
				p.Position.X =
					float32(tx*tileSize) - p.Width

				p.Velocity.X = 0
				return
			}
		}
	}

	if dx < 0 {
		tx := left / tileSize

		for ty := top / tileSize; ty <= bottom/tileSize; ty++ {
			if solidTile(tx, ty) {
				p.Position.X =
					float32((tx + 1) * tileSize)

				p.Velocity.X = 0
				return
			}
		}
	}
}

func movePlayerY(p *Player, dy float32) {
	p.Position.Y += dy
	p.OnGround = false

	left := int(math.Floor(float64(p.Position.X)))
	right := int(math.Floor(float64(
		p.Position.X + p.Width - 0.001,
	)))

	top := int(math.Floor(float64(p.Position.Y)))
	bottom := int(math.Floor(float64(
		p.Position.Y + p.Height - 0.001,
	)))

	if dy > 0 {
		ty := bottom / tileSize

		for tx := left / tileSize; tx <= right/tileSize; tx++ {
			if solidTile(tx, ty) {
				p.Position.Y =
					float32(ty*tileSize) - p.Height

				p.Velocity.Y = 0
				p.OnGround = true
				return
			}
		}
	}

	if dy < 0 {
		ty := top / tileSize

		for tx := left / tileSize; tx <= right/tileSize; tx++ {
			if solidTile(tx, ty) {
				p.Position.Y =
					float32((ty + 1) * tileSize)

				p.Velocity.Y = 0
				return
			}
		}
	}
}

// ============================================================
// RECTANGLE / RAY
// ============================================================

func segmentIntersectsRect(
	x0, y0 float64,
	x1, y1 float64,
	r rl.Rectangle,
) bool {

	minX := float64(r.X)
	maxX := float64(r.X + r.Width)

	minY := float64(r.Y)
	maxY := float64(r.Y + r.Height)

	dx := x1 - x0
	dy := y1 - y0

	tmin := 0.0
	tmax := 1.0

	if math.Abs(dx) < 0.000001 {
		if x0 < minX || x0 > maxX {
			return false
		}
	} else {
		invD := 1.0 / dx

		t1 := (minX - x0) * invD
		t2 := (maxX - x0) * invD

		if t1 > t2 {
			t1, t2 = t2, t1
		}

		if t1 > tmin {
			tmin = t1
		}

		if t2 < tmax {
			tmax = t2
		}

		if tmin > tmax {
			return false
		}
	}

	if math.Abs(dy) < 0.000001 {
		if y0 < minY || y0 > maxY {
			return false
		}
	} else {
		invD := 1.0 / dy

		t1 := (minY - y0) * invD
		t2 := (maxY - y0) * invD

		if t1 > t2 {
			t1, t2 = t2, t1
		}

		if t1 > tmin {
			tmin = t1
		}

		if t2 < tmax {
			tmax = t2
		}

		if tmin > tmax {
			return false
		}
	}

	return true
}

// ============================================================
// DDA RAYCAST
// ============================================================

func raycastWorld(
	x0, y0 float64,
	x1, y1 float64,
) bool {

	dx := x1 - x0
	dy := y1 - y0

	mapX := int(math.Floor(x0))
	mapY := int(math.Floor(y0))

	targetX := int(math.Floor(x1))
	targetY := int(math.Floor(y1))

	if mapX < 0 ||
		mapY < 0 ||
		mapX >= mapWidth ||
		mapY >= mapHeight {
		return false
	}

	if targetX < 0 ||
		targetY < 0 ||
		targetX >= mapWidth ||
		targetY >= mapHeight {
		return false
	}

	if mapX == targetX &&
		mapY == targetY {
		return true
	}

	stepX := 0
	stepY := 0

	if dx > 0 {
		stepX = 1
	} else if dx < 0 {
		stepX = -1
	}

	if dy > 0 {
		stepY = 1
	} else if dy < 0 {
		stepY = -1
	}

	var tDeltaX float64
	var tDeltaY float64

	if dx != 0 {
		tDeltaX = math.Abs(1 / dx)
	} else {
		tDeltaX = math.Inf(1)
	}

	if dy != 0 {
		tDeltaY = math.Abs(1 / dy)
	} else {
		tDeltaY = math.Inf(1)
	}

	var tMaxX float64
	var tMaxY float64

	if dx > 0 {
		tMaxX =
			(float64(mapX+1) - x0) /
				math.Abs(dx)
	} else if dx < 0 {
		tMaxX =
			(x0 - float64(mapX)) /
				math.Abs(dx)
	} else {
		tMaxX = math.Inf(1)
	}

	if dy > 0 {
		tMaxY =
			(float64(mapY+1) - y0) /
				math.Abs(dy)
	} else if dy < 0 {
		tMaxY =
			(y0 - float64(mapY)) /
				math.Abs(dy)
	} else {
		tMaxY = math.Inf(1)
	}

	for {
		if tMaxX < tMaxY {
			tMaxX += tDeltaX
			mapX += stepX
		} else {
			tMaxY += tDeltaY
			mapY += stepY
		}

		if mapX == targetX &&
			mapY == targetY {
			return true
		}

		if mapX < 0 ||
			mapY < 0 ||
			mapX >= mapWidth ||
			mapY >= mapHeight {
			return false
		}

		if world[mapY][mapX] != 0 {
			return false
		}
	}
}

// ============================================================
// DIRECTED LIGHT
// ============================================================

// Возвращает true, если точка находится внутри конуса света.
func pointInsideCone(
	x, y float64,
	light Light,
) (bool, float64) {

	dx := x - light.X
	dy := y - light.Y

	distance := math.Hypot(dx, dy)

	if distance >= light.Radius {
		return false, 0
	}

	if distance < 0.00001 {
		return true, 1
	}

	// Направление на точку.
	pointAngle :=
		math.Atan2(dy, dx)

	// Разница углов.
	angleDifference :=
		normalizeAngle(
			pointAngle - light.Angle,
		)

	absAngle :=
		math.Abs(angleDifference)

	if absAngle > light.ConeAngle {
		return false, 0
	}

	// 1 в центре.
	// 0 на краю конуса.
	edge :=
		1.0 -
			absAngle/light.ConeAngle

	return true, edge
}

func normalizeAngle(a float64) float64 {
	for a > math.Pi {
		a -= math.Pi * 2
	}

	for a < -math.Pi {
		a += math.Pi * 2
	}

	return a
}

// ============================================================
// LIGHT RAYCAST
// ============================================================

func raycastLight(
	x, y float64,
	light Light,
	player *Player,
) bool {

	// Сначала проверяем конус.
	inside, _ :=
		pointInsideCone(
			x,
			y,
			light,
		)

	if !inside {
		return false
	}

	// Стены.
	if !raycastWorld(
		x,
		y,
		light.X,
		light.Y,
	) {
		return false
	}

	// Игрок.
	playerRect := rl.Rectangle{
		X: player.Position.X / tileSize,
		Y: player.Position.Y / tileSize,

		Width: player.Width / tileSize,

		Height: player.Height / tileSize,
	}

	// Точка внутри игрока.
	if x >= float64(playerRect.X) &&
		x <= float64(playerRect.X+playerRect.Width) &&
		y >= float64(playerRect.Y) &&
		y <= float64(playerRect.Y+playerRect.Height) {

		return true
	}

	// Игрок блокирует свет.
	if segmentIntersectsRect(
		x,
		y,
		light.X,
		light.Y,
		playerRect,
	) {
		return false
	}

	return true
}

// ============================================================
// LIGHTMAP
// ============================================================

func buildLightmap(
	lights []Light,
	player *Player,
) []rl.Color {

	pixels :=
		make([]rl.Color, mapWidth*mapHeight)

	for y := 0; y < mapHeight; y++ {

		for x := 0; x < mapWidth; x++ {

			px :=
				float64(x) + 0.5

			py :=
				float64(y) + 0.5

			value := ambient

			for _, light := range lights {

				inside, edge :=
					pointInsideCone(
						px,
						py,
						light,
					)

				if !inside {
					continue
				}

				dx :=
					light.X - px

				dy :=
					light.Y - py

				distance :=
					math.Hypot(dx, dy)

				if !raycastLight(
					px,
					py,
					light,
					player,
				) {
					continue
				}

				// Радиальное затухание.
				radial :=
					1.0 -
						distance/light.Radius

				radial *= radial

				// Затухание к краю конуса.
				cone := edge * edge

				value +=
					radial *
						cone *
						light.Intensity
			}

			if value > 1 {
				value = 1
			}

			if value < 0 {
				value = 0
			}

			v :=
				uint8(value * 255)

			pixels[y*mapWidth+x] = rl.Color{
				R: v,
				G: v,
				B: v,
				A: 255,
			}
		}
	}

	return pixels
}

// ============================================================
// WORLD TEXTURE
// ============================================================

func createWorldTexture() rl.Texture2D {

	image :=
		rl.GenImageColor(
			mapWidth,
			mapHeight,
			rl.Black,
		)

	for y := 0; y < mapHeight; y++ {

		for x := 0; x < mapWidth; x++ {

			if world[y][x] != 0 {

				rl.ImageDrawPixel(
					image,
					int32(x),
					int32(y),

					rl.Color{
						R: 255,
						G: 0,
						B: 0,
						A: 255,
					},
				)
			}
		}
	}

	texture :=
		rl.LoadTextureFromImage(
			image,
		)

	rl.UnloadImage(image)

	rl.SetTextureFilter(
		texture,
		rl.TextureFilterNearest,
	)

	return texture
}

// ============================================================
// SHADER
// ============================================================

const vertexShader = `
#version 330

in vec3 vertexPosition;
in vec2 vertexTexCoord;
in vec4 vertexColor;

out vec2 fragTexCoord;
out vec4 fragColor;

uniform mat4 mvp;

void main()
{
    fragTexCoord = vertexTexCoord;
    fragColor = vertexColor;

    gl_Position =
        mvp *
        vec4(vertexPosition, 1.0);
}
`

const fragmentShader = `
#version 330

in vec2 fragTexCoord;
in vec4 fragColor;

out vec4 finalColor;

uniform sampler2D texture0;
uniform sampler2D lightMap;

uniform int debugLight;

const vec3 palette[12] = vec3[](

    // Empty
    vec3(0.006, 0.009, 0.015),
    vec3(0.015, 0.022, 0.038),
    vec3(0.032, 0.050, 0.075),
    vec3(0.065, 0.095, 0.135),
    vec3(0.120, 0.165, 0.220),
    vec3(0.210, 0.280, 0.360),

    // Block
    vec3(0.018, 0.016, 0.014),
    vec3(0.045, 0.040, 0.032),
    vec3(0.100, 0.085, 0.065),
    vec3(0.190, 0.155, 0.110),
    vec3(0.360, 0.280, 0.180),
    vec3(0.680, 0.500, 0.280)
);

void main()
{
    float light =
        texture(
            lightMap,
            fragTexCoord
        ).r;

    light =
        clamp(
            light,
            0.0,
            1.0
        );

    if (debugLight == 1)
    {
        finalColor =
            vec4(
                light,
                light,
                light,
                1.0
            );

        return;
    }

    float material =
        texture(
            texture0,
            fragTexCoord
        ).r;

    int materialID = 0;

    if (material > 0.5)
        materialID = 1;

    int level =
        int(
            floor(
                light * 5.0 + 0.5
            )
        );

    level =
        clamp(
            level,
            0,
            5
        );

    int paletteIndex =
        materialID * 6 +
        level;

    finalColor =
        vec4(
            palette[paletteIndex],
            1.0
        );
}
`

// ============================================================
// MAIN
// ============================================================

func main() {

	rl.InitWindow(
		screenWidth,
		screenHeight,
		"Directional 2D Light",
	)

	defer rl.CloseWindow()

	rl.SetTargetFPS(60)

	buildWorld()

	// ========================================================
	// TEXTURES
	// ========================================================

	worldTexture :=
		createWorldTexture()

	defer rl.UnloadTexture(worldTexture)

	lightImage :=
		rl.GenImageColor(
			mapWidth,
			mapHeight,
			rl.Black,
		)

	lightTexture :=
		rl.LoadTextureFromImage(
			lightImage,
		)

	rl.UnloadImage(lightImage)

	defer rl.UnloadTexture(lightTexture)

	rl.SetTextureFilter(
		worldTexture,
		rl.TextureFilterNearest,
	)

	rl.SetTextureFilter(
		lightTexture,
		rl.TextureFilterLinear,
	)

	// ========================================================
	// SHADER
	// ========================================================

	shader :=
		rl.LoadShaderFromMemory(
			vertexShader,
			fragmentShader,
		)

	defer rl.UnloadShader(shader)

	lightMapLocation :=
		rl.GetShaderLocation(
			shader,
			"lightMap",
		)

	// ========================================================
	// PLAYER
	// ========================================================

	player := Player{
		Position: rl.Vector2{
			X: 200,
			Y: 400,
		},

		Width:  7,
		Height: 13,
	}

	// ========================================================
	// LIGHT
	// ========================================================

	lights := []Light{

		// Свет игрока.
		{
			Radius:    22,
			Intensity: 1.3,

			// 90 градусов = вниз.
			Angle: math.Pi / 2,

			// Полный конус = 70 градусов.
			ConeAngle: math.Pi * 70 / 180,
		},

		// Фонарик мыши.
		{
			Radius:    28,
			Intensity: 2.0,

			ConeAngle: math.Pi * 25 / 180,
		},
	}

	// ========================================================
	// CAMERA
	// ========================================================

	camera := rl.Camera2D{
		Offset: rl.Vector2{
			X: screenWidth / 2,
			Y: screenHeight / 2,
		},

		Target: rl.Vector2{
			X: player.Position.X,
			Y: player.Position.Y,
		},

		Zoom: 2,
	}

	// ========================================================
	// LOOP
	// ========================================================

	for !rl.WindowShouldClose() {

		dt :=
			rl.GetFrameTime()

		if dt > 0.033 {
			dt = 0.033
		}

		// ----------------------------------------------------
		// PLAYER
		// ----------------------------------------------------

		updatePlayer(
			&player,
			dt,
		)

		// ----------------------------------------------------
		// LIGHT FROM PLAYER
		// ----------------------------------------------------

		lights[0].X =
			float64(
				player.Position.X+
					player.Width*0.5,
			) /
				float64(tileSize)

		lights[0].Y =
			float64(
				player.Position.Y+
					player.Height*0.5,
			) /
				float64(tileSize)

		// ----------------------------------------------------
		// MOUSE LIGHT POSITION
		// ----------------------------------------------------

		mouse :=
			rl.GetMousePosition()

		mouseWorld :=
			rl.GetScreenToWorld2D(
				mouse,
				camera,
			)

		lights[1].X =
			float64(mouseWorld.X) /
				float64(tileSize)

		lights[1].Y =
			float64(mouseWorld.Y) /
				float64(tileSize)

		// ----------------------------------------------------
		// MOUSE LIGHT DIRECTION
		// ----------------------------------------------------

		playerCenter :=
			rl.Vector2{
				X: player.Position.X +
					player.Width*0.5,

				Y: player.Position.Y +
					player.Height*0.5,
			}

		dx :=
			mouseWorld.X -
				playerCenter.X

		dy :=
			mouseWorld.Y -
				playerCenter.Y

		lights[0].Angle =
			math.Atan2(
				float64(dy),
				float64(dx),
			)

		// ----------------------------------------------------
		// LIGHTMAP
		// ----------------------------------------------------

		lightPixels :=
			buildLightmap(
				lights,
				&player,
			)

		rl.UpdateTexture(
			lightTexture,
			lightPixels,
		)

		// ----------------------------------------------------
		// CAMERA
		// ----------------------------------------------------

		camera.Target =
			rl.Vector2{
				X: player.Position.X +
					player.Width*0.5,

				Y: player.Position.Y +
					player.Height*0.5,
			}

		// ----------------------------------------------------
		// ZOOM
		// ----------------------------------------------------

		wheel :=
			rl.GetMouseWheelMove()

		if wheel > 0 {
			camera.Zoom++
		}

		if wheel < 0 {
			camera.Zoom--
		}

		if camera.Zoom < 1 {
			camera.Zoom = 1
		}

		if camera.Zoom > 8 {
			camera.Zoom = 8
		}

		// ====================================================
		// DRAW
		// ====================================================

		rl.BeginDrawing()

		rl.ClearBackground(
			rl.Color{
				R: 3,
				G: 4,
				B: 7,
				A: 255,
			},
		)

		rl.BeginMode2D(camera)

		// ----------------------------------------------------
		// WORLD + LIGHT
		// ----------------------------------------------------

		rl.BeginShaderMode(shader)

		rl.SetShaderValueTexture(
			shader,
			lightMapLocation,
			lightTexture,
		)

		rl.DrawTexturePro(
			worldTexture,

			rl.Rectangle{
				X:      0,
				Y:      0,
				Width:  mapWidth,
				Height: mapHeight,
			},

			rl.Rectangle{
				X:      0,
				Y:      0,
				Width:  mapWidth * tileSize,
				Height: mapHeight * tileSize,
			},

			rl.Vector2{},

			0,
			rl.White,
		)

		rl.EndShaderMode()

		// ----------------------------------------------------
		// PLAYER
		// ----------------------------------------------------

		rl.DrawRectangleRec(
			rl.Rectangle{
				X:      player.Position.X,
				Y:      player.Position.Y,
				Width:  player.Width,
				Height: player.Height,
			},

			rl.Color{
				R: 220,
				G: 220,
				B: 220,
				A: 255,
			},
		)

		// ----------------------------------------------------
		// LIGHT DIRECTION
		// ----------------------------------------------------

		for _, light := range lights {

			x :=
				light.X *
					float64(tileSize)

			y :=
				light.Y *
					float64(tileSize)

			dirX :=
				math.Cos(light.Angle)

			dirY :=
				math.Sin(light.Angle)

			endX :=
				x +
					dirX*
						light.Radius*
						float64(tileSize)

			endY :=
				y +
					dirY*
						light.Radius*
						float64(tileSize)

			rl.DrawLine(
				int32(x),
				int32(y),
				int32(endX),
				int32(endY),
				rl.Color{
					R: 255,
					G: 210,
					B: 60,
					A: 100,
				},
			)

			rl.DrawCircle(
				int32(x),
				int32(y),
				2,
				rl.Color{
					R: 255,
					G: 210,
					B: 60,
					A: 255,
				},
			)
		}

		rl.EndMode2D()

		// ====================================================
		// UI
		// ====================================================

		rl.DrawText(
			"A/D — движение",
			10,
			10,
			18,
			rl.White,
		)

		rl.DrawText(
			"SPACE — прыжок",
			10,
			32,
			18,
			rl.White,
		)

		rl.DrawText(
			"Mouse — направление фонарика",
			10,
			54,
			18,
			rl.White,
		)

		rl.DrawText(
			"Wheel — zoom",
			10,
			76,
			18,
			rl.White,
		)

		rl.DrawText(
			"F1 — lightmap",
			10,
			98,
			18,
			rl.White,
		)

		rl.DrawText(
			"Pixel: 4x4",
			10,
			120,
			18,
			rl.White,
		)

		rl.DrawText(
			"Zoom: "+zoomText(camera.Zoom),
			10,
			142,
			18,
			rl.White,
		)

		rl.EndDrawing()
	}
}

func zoomText(z float32) string {

	switch int(z) {

	case 1:
		return "1x"

	case 2:
		return "2x"

	case 3:
		return "3x"

	case 4:
		return "4x"

	case 5:
		return "5x"

	case 6:
		return "6x"

	case 7:
		return "7x"

	case 8:
		return "8x"
	}

	return "?"
}
