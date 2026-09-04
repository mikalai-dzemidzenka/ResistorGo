package resistor

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"math"
)

func runMainMenu(dt float32) {
	rl.BeginTextureMode(MainMenuTexture)
	rl.ClearBackground(rl.Color{0, 0, 0, 0})
	rl.BeginMode3D(rl.NewCamera3D(rl.Vector3{0, -1.5, 10}, rl.Vector3{0, 0, 0}, rl.Vector3{0, 1, 0}, 120, rl.CameraPerspective))
	var linesColor = rl.Red
	linesColor.A = 200
	for i := float32(-30); i <= 30; i++ {
		rl.DrawLine3D(rl.Vector3{X: i}, rl.Vector3{X: i * 1.5, Z: 10}, linesColor)
	}
	for i := float32(-30); i <= 30; i++ {
		rl.DrawLine3D(rl.Vector3{Y: float32(math.Sin(GameTimeSec)+1.2) * i}, rl.Vector3{X: -15 * WindowScale, Y: i * 3 * WindowScale}, linesColor)
		rl.DrawLine3D(rl.Vector3{Y: float32(math.Sin(GameTimeSec)+1.2) * i}, rl.Vector3{X: 15 * WindowScale, Y: i * 3 * WindowScale}, linesColor)
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
	rl.BeginDrawing()
	rl.ClearBackground(rl.Black)
	{
		rl.DrawTextureEx(MainMenuTexture.Texture, rl.Vector2{
			X: 0,
			Y: 0,
		}, 0, WindowScale, rl.Red)
		tex := resources["resources/main_menu/resistor.png"]
		col := rl.White
		if GameTimeSec > 3 {
			col.A = 255
		} else if GameTimeSec > 2 {
			col.A = uint8(math.Min((math.Sin(GameTimeSec-2))*255, 255))
		} else {
			col.A = 0
		}
		rl.DrawTextureEx(tex, rl.Vector2{((SW / 2) - float32(tex.Width)/2) * WindowScale, ((SH / 2) - float32(tex.Height)) * WindowScale}, 0, WindowScale, col)
	}
	rl.EndDrawing()
}

func head() {
	const headScale = 0.75 * WindowScale
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
	const jabScale = 0.25 * WindowScale
	colorJab := rl.Red
	if GameTimeSec > 1.5 {
		colorJab.A = uint8((math.Sin(GameTimeSec-1.5) + 1) * 127)
	} else {
		colorJab.A = 0
	}
	rl.DrawModelWiresEx(
		models["resources/main_menu/jab1.obj"],
		rl.Vector3{X: -1.5 * WindowScale, Y: WindowScale},
		rl.Vector3{0, 1, 0},
		float32(-GameTimeSec*100),
		rl.Vector3{jabScale, -jabScale, jabScale},
		colorJab,
	)
}

func jab2() {
	const jabScale = 0.25 * WindowScale
	colorJab := rl.Red
	if GameTimeSec > 1.8 {
		colorJab.A = uint8((math.Sin(GameTimeSec-1.8) + 1) * 127)
	} else {
		colorJab.A = 0
	}
	rl.DrawModelWiresEx(
		models["resources/main_menu/jab2.obj"],
		rl.Vector3{X: WindowScale * 1.5, Y: WindowScale},
		rl.Vector3{0, 1, 0},
		float32(-GameTimeSec*100),
		rl.Vector3{jabScale, -jabScale, jabScale},
		colorJab,
	)
}

func kick() {
	const jabScale = 0.25 * WindowScale
	colorJab := rl.Red
	if GameTimeSec > 4 {
		colorJab.A = uint8((math.Sin(GameTimeSec-4) + 1) * 127)
	} else {
		colorJab.A = 0
	}
	rl.DrawModelWiresEx(
		models["resources/main_menu/kick.obj"],
		rl.Vector3{X: WindowScale, Y: WindowScale},
		rl.Vector3{0, 1, 0},
		float32(GameTimeSec*35),
		rl.Vector3{jabScale, -jabScale, jabScale},
		colorJab,
	)
}

func martelo() {
	const jabScale = 0.25 * WindowScale
	colorJab := rl.Red
	if GameTimeSec > 4.5 {
		colorJab.A = uint8((math.Sin(GameTimeSec-4.5) + 1) * 127)
	} else {
		colorJab.A = 0
	}
	rl.DrawModelWiresEx(
		models["resources/main_menu/martelo.obj"],
		rl.Vector3{X: -WindowScale, Y: WindowScale},
		rl.Vector3{0, 1, 0},
		float32(GameTimeSec*50),
		rl.Vector3{jabScale, -jabScale, jabScale},
		colorJab,
	)
}
