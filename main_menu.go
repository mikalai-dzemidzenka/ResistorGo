package resistor

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"math"
)

var MainMenuTexture rl.RenderTexture2D

func runMainMenu(dt float32) {
	Cam.Camera2D.Zoom = 4
	rl.BeginDrawing()
	{
		rl.BeginTextureMode(MainMenuTexture)
		rl.ClearBackground(rl.Color{0, 0, 0, 0})
		rl.BeginMode3D(rl.NewCamera3D(rl.Vector3{0, -1.5, 10}, rl.Vector3{0, 0, 0}, rl.Vector3{0, 1, 0}, 120, rl.CameraPerspective))
		var linesColor = rl.Red
		linesColor.A = 200
		for i := float32(-30); i <= 30; i++ {
			rl.DrawLine3D(rl.Vector3{X: i}, rl.Vector3{X: i * 1.5, Z: 10}, linesColor)
		}
		for i := float32(-30); i <= 30; i++ {
			rl.DrawLine3D(rl.Vector3{Y: float32(math.Sin(GameTimeSec)+1.2) * i}, rl.Vector3{X: -15, Y: i * 3}, linesColor)
			rl.DrawLine3D(rl.Vector3{Y: float32(math.Sin(GameTimeSec)+1.2) * i}, rl.Vector3{X: 15, Y: i * 3}, linesColor)
		}
		rl.EndMode3D()

		rl.BeginMode3D(rl.NewCamera3D(rl.Vector3{0, 5, 10}, rl.Vector3{0, 0, 0}, rl.Vector3{0, 1, 0}, 10, rl.CameraOrthographic))
		head()
		jab1()
		jab2()
		kick()
		martelo()
		rl.EndMode3D()
		rl.EndTextureMode()
		rl.ClearBackground(rl.Black)
		rl.BeginMode2D(Cam.Camera2D)
		{
			rl.DrawTextureEx(MainMenuTexture.Texture, rl.Vector2{
				X: Pl.Pos.X - SW/2,
				Y: Pl.Pos.Y - SH/2,
			}, 0, 1, rl.Red)
			tex := resources["resources/main_menu/resistor.png"]
			col := rl.White
			if GameTimeSec > 3 {
				col.A = 255
			} else if GameTimeSec > 2 {
				col.A = uint8(math.Min((math.Sin(GameTimeSec-2))*255, 255))
			} else {
				col.A = 0
			}
			rl.DrawTexture(tex, -tex.Width/2, -20, col)
		}
		rl.EndMode2D()
	}
	rl.EndDrawing()
}

func head() {
	const headScale = 0.75
	color := rl.Red
	color.A = uint8((math.Sin(GameTimeSec) + 1) * 127)
	rl.DrawModelWiresEx(
		models["resources/main_menu/head.obj"],
		rl.Vector3{},
		rl.Vector3{0, 1, 0},
		float32(GameTimeSec*100),
		rl.Vector3{headScale, -headScale, headScale},
		color,
	)
}

func jab1() {
	const jabScale = 0.25
	colorJab := rl.Red
	if GameTimeSec > 1.5 {
		colorJab.A = uint8((math.Sin(GameTimeSec-1.5) + 1) * 127)
	} else {
		colorJab.A = 0
	}
	rl.DrawModelWiresEx(
		models["resources/main_menu/jab1.obj"],
		rl.Vector3{X: -1.5, Y: 1},
		rl.Vector3{0, 1, 0},
		float32(-GameTimeSec*100),
		rl.Vector3{jabScale, -jabScale, jabScale},
		colorJab,
	)
}

func jab2() {
	const jabScale = 0.25
	colorJab := rl.Red
	if GameTimeSec > 1.8 {
		colorJab.A = uint8((math.Sin(GameTimeSec-1.8) + 1) * 127)
	} else {
		colorJab.A = 0
	}
	rl.DrawModelWiresEx(
		models["resources/main_menu/jab2.obj"],
		rl.Vector3{X: 1.5, Y: 1},
		rl.Vector3{0, 1, 0},
		float32(-GameTimeSec*100),
		rl.Vector3{jabScale, -jabScale, jabScale},
		colorJab,
	)
}

func kick() {
	const jabScale = 0.25
	colorJab := rl.Red
	if GameTimeSec > 4 {
		colorJab.A = uint8((math.Sin(GameTimeSec-4) + 1) * 127)
	} else {
		colorJab.A = 0
	}
	rl.DrawModelWiresEx(
		models["resources/main_menu/kick.obj"],
		rl.Vector3{X: 1, Y: 1},
		rl.Vector3{0, 1, 0},
		float32(GameTimeSec*35),
		rl.Vector3{jabScale, -jabScale, jabScale},
		colorJab,
	)
}

func martelo() {
	const jabScale = 0.25
	colorJab := rl.Red
	if GameTimeSec > 4.5 {
		colorJab.A = uint8((math.Sin(GameTimeSec-4.5) + 1) * 127)
	} else {
		colorJab.A = 0
	}
	rl.DrawModelWiresEx(
		models["resources/main_menu/martelo.obj"],
		rl.Vector3{X: -1, Y: 1},
		rl.Vector3{0, 1, 0},
		float32(GameTimeSec*50),
		rl.Vector3{jabScale, -jabScale, jabScale},
		colorJab,
	)
}
