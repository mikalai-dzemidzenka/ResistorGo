package resistor

import (
	rlg "github.com/gen2brain/raylib-go/raygui"
	rl "github.com/gen2brain/raylib-go/raylib"
	"math"
	"strconv"
	"strings"
)

const (
	menuMain = iota
	menuHost
	menuJoin
	menuConnecting
)

var (
	menuScreen      int
	menuError       string
	hostPort        = strconv.Itoa(DefaultPort)
	hostPortEdit    bool
	hostMaxPlayers  int32 = 8
	hostMaxEdit     bool
	hostBots        int32 = 1
	hostBotsEdit    bool
	hostMaxSpec     int32 = 4
	hostMaxSpecEdit bool
	joinAddr        = "127.0.0.1"
	joinAddrEdit    bool
)

var mainMenuInit bool

func runMainMenu(dt float32) {
	if Srv != nil || Cli != nil {
		netFrame(dt) // connecting; switches to the level once the first snapshot arrives
	}
	if !mainMenuInit {
		mainMenuInit = true
		MainMenuTexture = rl.LoadRenderTexture(SW, SH)
		rl.SetTextureFilter(MainMenuTexture.Texture, rl.FilterPoint)
	}
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
	drawMenuUI()
	rl.EndDrawing()
}

func drawMenuUI() {
	const w, h, gap = 300, 44, 10
	x := float32(SW*WindowScale/2 - w/2)
	y := float32(SH*WindowScale/2 + 60)
	row := func(i int) rl.Rectangle { return rl.Rectangle{X: x, Y: y + float32(i)*(h+gap), Width: w, Height: h} }
	label := func(i int, text string) {
		r := row(i)
		r.X -= 170
		r.Width = 160
		rlg.Label(r, text)
	}

	switch menuScreen {
	case menuMain:
		if rlg.Button(row(0), "Single Player") {
			menuError = ""
			startOffline()
		}
		if rlg.Button(row(1), "Host Game") {
			menuError = ""
			menuScreen = menuHost
		}
		if rlg.Button(row(2), "Join Game") {
			menuError = ""
			menuScreen = menuJoin
		}
		if rlg.Button(row(3), "Quit") {
			quit = true
		}
	case menuHost:
		label(0, "Port")
		if rlg.TextBox(row(0), &hostPort, 6, hostPortEdit) {
			hostPortEdit = !hostPortEdit
		}
		label(1, "Max players")
		if rlg.Spinner(row(1), "", &hostMaxPlayers, 2, 16, hostMaxEdit) {
			hostMaxEdit = !hostMaxEdit
		}
		label(2, "Bots")
		if rlg.Spinner(row(2), "", &hostBots, 0, 15, hostBotsEdit) {
			hostBotsEdit = !hostBotsEdit
		}
		label(3, "Max spectators")
		if rlg.Spinner(row(3), "", &hostMaxSpec, 0, 8, hostMaxSpecEdit) {
			hostMaxSpecEdit = !hostMaxSpecEdit
		}
		if rlg.Button(row(4), "Start") {
			hostGame()
		}
		if rlg.Button(row(5), "Back") {
			menuError = ""
			menuScreen = menuMain
		}
	case menuJoin:
		label(0, "Address")
		if rlg.TextBox(row(0), &joinAddr, 64, joinAddrEdit) {
			joinAddrEdit = !joinAddrEdit
		}
		join := func(spectator bool) {
			menuError = ""
			if err := startClient(joinAddr, spectator); err != nil {
				menuError = err.Error()
			} else {
				connectingFrom = menuJoin
				menuScreen = menuConnecting
			}
		}
		if rlg.Button(row(1), "Connect") {
			join(false)
		}
		if rlg.Button(row(2), "Join as spectator") {
			join(true)
		}
		if rlg.Button(row(3), "Back") {
			menuError = ""
			menuScreen = menuMain
		}
	case menuConnecting:
		text := "Connecting..."
		if Cli != nil {
			text = "Connecting to " + Cli.Addr + "..."
		}
		rlg.Label(row(0), text)
		if rlg.Button(row(1), "Cancel") {
			leaveMatch("", connectingFrom)
		}
	}
	if menuError != "" {
		// above the rows: the host screen uses all six
		rl.DrawText(menuError, int32(x), int32(y-34), 20, rl.Red)
	}
}

func hostGame() {
	menuError = ""
	port, err := strconv.Atoi(strings.TrimSpace(hostPort))
	if err != nil || port < 1 || port > 65535 {
		menuError = "invalid port"
		return
	}
	cfg := DefaultServerConfig()
	cfg.Port = uint16(port)
	cfg.MaxPlayers = int(hostMaxPlayers)
	cfg.Bots = int(min(hostBots, hostMaxPlayers-1))
	cfg.MaxSpectators = int(hostMaxSpec)
	if err := startListen(cfg); err != nil {
		menuError = err.Error()
		return
	}
	connectingFrom = menuHost
	menuScreen = menuConnecting
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
