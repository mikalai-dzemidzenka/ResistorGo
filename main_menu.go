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
		MainMenuTexture = rl.LoadRenderTexture(SW*WindowScale/menuPixel, SH*WindowScale/menuPixel)
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
		rl.DrawLine3D(rl.Vector3{Y: float32(math.Sin(GameTimeSec)+1.2) * i}, rl.Vector3{X: -15, Y: i * 3}, linesColor)
		rl.DrawLine3D(rl.Vector3{Y: float32(math.Sin(GameTimeSec)+1.2) * i}, rl.Vector3{X: 15, Y: i * 3}, linesColor)
	}
	rl.EndMode3D()

	// 72 screen px per unit; the texture is drawn unflipped, so -Y is up on screen.
	// Models sit right of the menu panel, never behind the logo or the widgets.
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
		}, 0, menuPixel, rl.Red)
		tex := resources["resources/main_menu/resistor.png"]
		col := rl.White
		if GameTimeSec > 3 {
			col.A = 255
		} else if GameTimeSec > 2 {
			col.A = uint8(math.Min((math.Sin(GameTimeSec-2))*255, 255))
		} else {
			col.A = 0
		}
		s := logoScale(tex.Width)
		x := math.Floor(menuRowX + menuRowW/2 - float64(tex.Width)*float64(s)/2)
		rl.DrawTextureEx(tex, rl.Vector2{float32(x), menuLogoBottom - float32(tex.Height)*s}, 0, s, col)
	}
	drawMenuUI()
	rl.EndDrawing()
}

// menuPixel is the on-screen size of one scene pixel. The scene renders into a
// window/menuPixel texture and is blown back up by menuPixel, so a smaller value
// means more detail at the same size. Must divide the window evenly.
const menuPixel = 4

// The logo and the widgets use the left menuPanelW pixels of the window; the models use the rest.
const (
	menuPanelW     = 560
	menuRowX       = 190
	menuRowY       = 360
	menuRowW       = 300
	menuRowH       = 40
	menuRowGap     = 8
	menuLabelW     = 170 // label plus gap, left of a row
	menuLogoBottom = menuRowY - 64
)

// menuRow is the i-th widget row of a menu screen, in window pixels.
func menuRow(i int) rl.Rectangle {
	return rl.Rectangle{
		X:      menuRowX,
		Y:      menuRowY + float32(i)*(menuRowH+menuRowGap),
		Width:  menuRowW,
		Height: menuRowH,
	}
}

// logoScale is the largest integer scale at which a logo texW pixels wide,
// centred over the widget column, stays inside the menu panel.
func logoScale(texW int32) float32 {
	s := float32(math.Floor(2 * (menuPanelW - menuRowX - menuRowW/2) / float64(texW)))
	return max(1, min(s, WindowScale))
}

func drawMenuUI() {
	row := menuRow
	label := func(i int, text string) {
		r := row(i)
		r.X -= menuLabelW
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
		// above the rows (the host screen uses all six), from the label column so long messages stay in the panel
		r := menuRow(0)
		rl.DrawText(menuError, int32(r.X-menuLabelW), int32(r.Y-34), 20, rl.Red)
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
	const headScale = 1.3
	color := rl.Red
	color.A = uint8((math.Sin(GameTimeSec) + 1) * 127)
	rl.DrawModelWiresEx(
		models["resources/main_menu/head.obj"],
		rl.Vector3{X: 3.9, Y: -2.7},
		rl.Vector3{0, 1, 0},
		float32(GameTimeSec*100),
		rl.Vector3{headScale, -headScale, headScale},
		color,
	)
}

func jab1() {
	const jabScale = 0.55
	colorJab := rl.Red
	if GameTimeSec > 1.5 {
		colorJab.A = uint8((math.Sin(GameTimeSec-1.5) + 1) * 127)
	} else {
		colorJab.A = 0
	}
	rl.DrawModelWiresEx(
		models["resources/main_menu/jab1.obj"],
		rl.Vector3{X: 0.7, Y: -1.25},
		rl.Vector3{0, 1, 0},
		float32(-GameTimeSec*100),
		rl.Vector3{jabScale, -jabScale, jabScale},
		colorJab,
	)
}

func jab2() {
	const jabScale = 0.55
	colorJab := rl.Red
	if GameTimeSec > 1.8 {
		colorJab.A = uint8((math.Sin(GameTimeSec-1.8) + 1) * 127)
	} else {
		colorJab.A = 0
	}
	rl.DrawModelWiresEx(
		models["resources/main_menu/jab2.obj"],
		rl.Vector3{X: 7.1, Y: -1.25},
		rl.Vector3{0, 1, 0},
		float32(-GameTimeSec*100),
		rl.Vector3{jabScale, -jabScale, jabScale},
		colorJab,
	)
}

func kick() {
	const jabScale = 0.7
	colorJab := rl.Red
	if GameTimeSec > 4 {
		colorJab.A = uint8((math.Sin(GameTimeSec-4) + 1) * 127)
	} else {
		colorJab.A = 0
	}
	rl.DrawModelWiresEx(
		models["resources/main_menu/kick.obj"],
		rl.Vector3{X: 6.4, Y: 4.7},
		rl.Vector3{0, 1, 0},
		float32(GameTimeSec*35),
		rl.Vector3{jabScale, -jabScale, jabScale},
		colorJab,
	)
}

func martelo() {
	const jabScale = 0.7
	colorJab := rl.Red
	if GameTimeSec > 4.5 {
		colorJab.A = uint8((math.Sin(GameTimeSec-4.5) + 1) * 127)
	} else {
		colorJab.A = 0
	}
	rl.DrawModelWiresEx(
		models["resources/main_menu/martelo.obj"],
		rl.Vector3{X: 1.4, Y: 4.7},
		rl.Vector3{0, 1, 0},
		float32(GameTimeSec*50),
		rl.Vector3{jabScale, -jabScale, jabScale},
		colorJab,
	)
}
