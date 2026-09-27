package resistor

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	rlg "github.com/gen2brain/raylib-go/raygui"
	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	menuMain = iota
	menuHost
	menuJoin
	menuOptions
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
	winSizeEdit     bool
)

var mainMenuInit bool

func runMainMenu(dt float32) {
	if Srv != nil || Cli != nil {
		netFrame(dt) // connecting; switches to the level once the first snapshot arrives
	}
	if !mainMenuInit {
		mainMenuInit = true
		MainMenuTexture = rl.LoadRenderTexture(menuSceneW, menuSceneH)
		rl.SetTextureFilter(MainMenuTexture.Texture, rl.FilterPoint)
		MenuBackground = rl.LoadRenderTexture(menuSceneW, menuSceneH)
		rl.SetTextureFilter(MenuBackground.Texture, rl.FilterPoint)
	}
	t := GameTimeSec
	gridCol := neonColor(neonGrid, t)

	menuInput()

	// the scene: the sky with its sun, the neon grid, the rain and the CRT
	drawMenuBackground(t, dt, gridCol)

	rl.BeginTextureMode(MainMenuTexture)
	rl.ClearBackground(menuSky)
	drawSceneBackground()

	// 72 screen px per unit; the texture is drawn unflipped, so -Y is up on screen.
	// Models sit right of the menu panel, never behind the logo or the widgets.
	// They do not rock with the background: their placement is what the menu
	// layout guarantees rest on.
	rl.BeginMode3D(rl.NewCamera3D(rl.Vector3{0, 5, 10}, rl.Vector3{0, 0, 0}, rl.Vector3{0, 1, 0}, 10, rl.CameraOrthographic))
	head()
	jab1()
	jab2()
	kick()
	martelo()
	rl.EndMode3D()
	rl.EndTextureMode()

	rl.BeginTextureMode(MenuFrame)
	rl.ClearBackground(rl.Black)
	rl.DrawTextureEx(MainMenuTexture.Texture, rl.Vector2{}, 0, menuPixel, rl.White)
	drawMenuPanel()
	drawMenuRules()
	drawMenuLogo()
	drawMenuUI()
	rl.EndTextureMode()

	// everything the menu shows goes through the palette pass, so every pixel of
	// it - fades, blends and widget text included - is an Endesga 64 colour
	rl.BeginTextureMode(VirtualScreen)
	rl.ClearBackground(rl.Black)
	rl.BeginShaderMode(PaletteShader)
	rl.SetShaderValueTexture(PaletteShader, rl.GetShaderLocation(PaletteShader, "texture1"), MenuPaletteTexture)
	rl.SetShaderValue(PaletteShader, rl.GetShaderLocation(PaletteShader, "resolution"), []float32{VW, VH}, rl.ShaderUniformVec2)
	rl.SetShaderValue(PaletteShader, rl.GetShaderLocation(PaletteShader, "pixelSize"), []float32{1}, rl.ShaderUniformFloat)
	rl.SetShaderValue(PaletteShader, rl.GetShaderLocation(PaletteShader, "paletteSize"), []float32{float32(len(EndesgaPalette))}, rl.ShaderUniformFloat)
	rl.SetShaderValue(PaletteShader, rl.GetShaderLocation(PaletteShader, "ditherStrength"), []float32{0}, rl.ShaderUniformFloat)
	rl.DrawTexturePro(MenuFrame.Texture,
		rl.Rectangle{Width: VW, Height: -VH}, // флип render-текстуры
		rl.Rectangle{Width: VW, Height: VH}, rl.Vector2{}, 0, rl.White)
	rl.EndShaderMode()
	rl.EndTextureMode()
}

// menuPixel is the on-screen size of one scene pixel. The scene renders into a
// window/menuPixel texture and is blown back up by menuPixel, so a smaller value
// means more detail at the same size. Must divide the window evenly.
const menuPixel = 2

// The logo and the widgets use the left menuPanelW pixels of the window; the models use the rest.
const (
	menuPanelW     = 560
	menuRowX       = 190
	menuRowY       = 360
	menuRowW       = 300
	menuRowH       = 40
	menuRowGap     = 8
	menuLabelW     = 170           // label plus gap, left of a row
	menuLogoBottom = menuRowY - 56 // clear of the header above it by menuLogoGap
	menuLogoGap    = 10
)

// The terminal the menu is framed as, in window pixels: a rule across the top
// and the bottom of the window with a readout on each, and the panel the menu
// column stands on. The panel is wide enough for the logo at the scale
// logoScale picks and tall enough for the six rows of the host screen.
const (
	menuRuleTop     = 48
	menuRuleBottom  = VH - 48
	menuRuleInset   = 28 // how far short of the window edges a rule stops
	menuPanelLeft   = 8
	menuPanelTop    = 160
	menuPanelBottom = menuRuleBottom - 8
	menuTextSpacing = 2 // the gap between glyphs of every text the menu draws
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

// menuText draws one line of the menu's text. Everything the chrome writes goes
// through here, so it is all the default font at one spacing.
func menuText(s string, x, y, size float32, col rl.Color) {
	rl.DrawTextEx(rl.GetFontDefault(), s, rl.Vector2{X: x, Y: y}, size, menuTextSpacing, col)
}

// menuTextRight draws one line of text ending at x.
func menuTextRight(s string, x, y, size float32, col rl.Color) {
	w := rl.MeasureTextEx(rl.GetFontDefault(), s, size, menuTextSpacing).X
	menuText(s, x-w, y, size, col)
}

// menuLine draws a one pixel high horizontal rule.
func menuLine(x1, x2, y, width float32, col rl.Color) {
	rl.DrawLineEx(rl.Vector2{X: x1, Y: y}, rl.Vector2{X: x2, Y: y}, width, col)
}

// The highlight of the main screen. A row is chosen with the mouse or with the
// keyboard, so the first screen can be worked without a mouse at all; the other
// screens have text fields and spinners on their rows and are mouse-driven, as
// they were.
var (
	menuSel  int  // the highlighted row
	menuTake bool // the highlighted row was chosen this frame
)

// menuMainLabels are the rows of the main screen, in order.
var menuMainLabels = []string{"Single Player", "Level Editor", "Host Game", "Join Game", "Options", "Quit"}

// menuInput reads the keys the chrome advertises: the highlight on the main
// screen, and Escape backing out of any other screen the way its Back button
// does.
func menuInput() {
	if menuScreen == menuMain {
		menuNavigate(len(menuMainLabels))
	} else {
		menuTake = false
		if rl.IsKeyPressed(rl.KeyEscape) {
			menuBack()
		}
	}
}

// menuNavigate moves the highlight over the n rows of the main screen and notes
// whether the highlighted row was chosen. Moving the mouse over a row takes the
// highlight, so the two ways of picking never disagree about what is lit.
func menuNavigate(n int) {
	if rl.IsKeyPressed(rl.KeyDown) || rl.IsKeyPressed(rl.KeyS) {
		menuSel = (menuSel + 1) % n
	}
	if rl.IsKeyPressed(rl.KeyUp) || rl.IsKeyPressed(rl.KeyW) {
		menuSel = (menuSel + n - 1) % n
	}
	menuSel = clamp(menuSel, 0, n-1)

	mouse := rl.GetMousePosition()
	moved := rl.GetMouseDelta()
	click := rl.IsMouseButtonPressed(rl.MouseLeftButton)
	hovered := -1
	for i := range n {
		if rl.CheckCollisionPointRec(mouse, menuRow(i)) {
			hovered = i
			if moved.X != 0 || moved.Y != 0 || click {
				menuSel = i
			}
		}
	}
	menuTake = rl.IsKeyPressed(rl.KeyEnter) || rl.IsKeyPressed(rl.KeySpace) || (click && hovered >= 0)
}

// menuBack leaves the screen that is up, as its last row does.
func menuBack() {
	hostPortEdit, hostMaxEdit, hostBotsEdit, hostMaxSpecEdit = false, false, false, false
	joinAddrEdit, winSizeEdit = false, false
	switch menuScreen {
	case menuHost, menuJoin:
		menuError = ""
		menuScreen = menuMain
	case menuOptions:
		SaveSettings()
		menuScreen = menuMain
	case menuConnecting:
		leaveMatch("", connectingFrom)
	}
}

// menuScreenName is the readout under the logo: which screen is up.
func menuScreenName() string {
	switch menuScreen {
	case menuHost:
		return "HOST // OPEN A SESSION"
	case menuJoin:
		return "JOIN // REACH A SESSION"
	case menuOptions:
		return "OPTIONS // THIS MACHINE"
	case menuConnecting:
		return "LINK // HANDSHAKE"
	}
	return "MAIN // SELECT A ROUTE"
}

// drawMenuPanel draws the panel the menu column stands on, and the header above
// the logo.
func drawMenuPanel() {
	p := rl.Rectangle{
		X: menuPanelLeft, Y: menuPanelTop,
		Width:  menuPanelW - menuPanelLeft,
		Height: menuPanelBottom - menuPanelTop,
	}
	rl.DrawRectangleRec(p, rl.Fade(menuPanelCol, 0.90))
	rl.DrawRectangleLinesEx(p, 1, rl.Fade(menuDim, 0.85))

	x := float32(menuRowX - menuLabelW) // the left edge of the label column
	menuLine(x, x+200, menuPanelTop+24, 2, menuAccent)
	menuText("NEO-MINSK", x, menuPanelTop+36, 22, menuBright)
	menuText("SECURE TERMINAL // ACCESS", x, menuPanelTop+66, 12, menuMuted)
	menuText(menuScreenName(), x, menuLogoBottom+8, 12, menuMuted)
}

// drawMenuRules draws the two rules the window is framed by and their readouts:
// what this terminal is, what it is talking to, and how the screen that is up is
// worked.
func drawMenuRules() {
	menuLine(menuRuleInset, VW-menuRuleInset, menuRuleTop, 1, rl.Fade(menuNeon, 0.45))
	menuText("RESISTOR.SYS", 30, 21, 18, menuNeon)
	menuText("SECURE CONNECTION // NEO-MINSK", 200, 23, 14, menuMuted)

	status := "SYS.ONLINE"
	switch {
	case Cli != nil:
		status = "LINK ACTIVE"
	case Srv != nil:
		status = "HOST ACTIVE"
	}
	menuTextRight(status, VW-30, 23, 14, menuNeon)

	menuLine(menuRuleInset, VW-menuRuleInset, menuRuleBottom, 1, rl.Fade(menuNeon, 0.30))
	hint := "W/S  NAVIGATE     ENTER  SELECT"
	if menuScreen != menuMain {
		hint = "MOUSE  SELECT     ESC  BACK"
	}
	menuText(hint, 30, VH-31, 14, menuMuted)
	menuTextRight(fmt.Sprintf("PROTO %d // BUILD 0.1", ProtocolVersion), VW-30, VH-31, 14, menuAccent)
}

// drawMenuLogo draws the game's logo over the widget column, with an offset
// ghost of itself behind it. Its fade-in is unchanged.
func drawMenuLogo() {
	tex := resources["resources/main_menu/resistor.png"]
	a := 0.0
	switch {
	case GameTimeSec > 3:
		a = 1
	case GameTimeSec > 2:
		a = math.Min(math.Sin(GameTimeSec-2), 1)
	}
	col, ghost := rl.White, menuAccent
	col.A = uint8(a * 255)
	ghost.A = uint8(a * 110)

	s := logoScale(tex.Width)
	x := float32(math.Floor(menuRowX + menuRowW/2 - float64(tex.Width)*float64(s)/2))
	y := menuLogoBottom - float32(tex.Height)*s
	rl.DrawTextureEx(tex, rl.Vector2{X: x + 3, Y: y + 3}, 0, s, ghost)
	rl.DrawTextureEx(tex, rl.Vector2{X: x, Y: y}, 0, s, col)
}

// drawMenuRow draws row i as a terminal row: its number, its label, a marker
// down its left edge, and a pulsing glow around it while it is highlighted.
func drawMenuRow(r rl.Rectangle, label string, i int, active bool) {
	edge, fill, mark := menuDim, menuRowFill, menuMuted
	if active {
		edge, fill, mark = menuNeon, menuRowFillOn, menuNeon
		pulse := float32(0.22 + 0.10*math.Sin(GameTimeSec*5))
		glow := rl.Rectangle{X: r.X - 4, Y: r.Y - 4, Width: r.Width + 8, Height: r.Height + 8}
		rl.DrawRectangleRec(glow, rl.Fade(menuNeon, pulse))
	}
	rl.DrawRectangleRec(r, fill)
	rl.DrawRectangleLinesEx(r, 1, rl.Fade(edge, 0.80))
	rl.DrawRectangleRec(rl.Rectangle{X: r.X, Y: r.Y, Width: 4, Height: r.Height}, mark)

	menuText(fmt.Sprintf("%02d", i+1), r.X+14, r.Y+14, 14, mark)
	menuText(label, r.X+48, r.Y+11, 20, menuBright)
	if active {
		menuText(">", r.X+r.Width-22, r.Y+11, 20, menuAccent)
	}
}

// menuButton draws row i as a menu row and reports whether it was chosen this
// frame. On the main screen the keyboard moves the highlight; on the others the
// row under the mouse is the lit one.
func menuButton(i int, label string) bool {
	r := menuRow(i)
	hover := rl.CheckCollisionPointRec(rl.GetMousePosition(), r)
	if menuScreen == menuMain {
		drawMenuRow(r, label, i, i == menuSel)
		return menuTake && i == menuSel
	}
	drawMenuRow(r, label, i, hover)
	return hover && rl.IsMouseButtonPressed(rl.MouseLeftButton)
}

// menuChoice draws the rows of the main screen and returns the one chosen this
// frame, or -1.
func menuChoice(labels []string) int {
	chosen := -1
	for i, s := range labels {
		if menuButton(i, s) {
			chosen = i
		}
	}
	return chosen
}

// menuWidgetStyle paints the raygui widgets the menu still uses - the text
// fields, the spinners, the slider, the toggles and the dropdown - in the
// scene's reds, so they match the rows drawn by hand. DEFAULT reaches every
// control; loadStyle puts the dark editor style back when a level starts.
func menuWidgetStyle() {
	set := func(p rlg.PropertyID, col rl.Color) {
		rlg.SetStyle(rlg.DEFAULT, p, rlg.PropertyValue(rl.ColorToInt(col)))
	}
	set(rlg.BORDER_COLOR_NORMAL, menuDim)
	set(rlg.BASE_COLOR_NORMAL, menuRowFill)
	set(rlg.TEXT_COLOR_NORMAL, menuBright)
	set(rlg.BORDER_COLOR_FOCUSED, menuNeon)
	set(rlg.BASE_COLOR_FOCUSED, menuRowFillOn)
	set(rlg.TEXT_COLOR_FOCUSED, menuBright)
	set(rlg.BORDER_COLOR_PRESSED, menuAccent)
	set(rlg.BASE_COLOR_PRESSED, menuRowFillOn)
	set(rlg.TEXT_COLOR_PRESSED, menuBright)
	set(rlg.BORDER_COLOR_DISABLED, menuDim)
	set(rlg.BASE_COLOR_DISABLED, menuRowFill)
	set(rlg.TEXT_COLOR_DISABLED, menuMuted)
	set(rlg.LINE_COLOR, menuDim)
	set(rlg.BACKGROUND_COLOR, menuRowFill)
}

func drawMenuUI() {
	menuWidgetStyle()
	row := menuRow
	label := func(i int, text string) {
		r := row(i)
		menuText(text, r.X-menuLabelW, r.Y+12, 16, menuBright)
	}

	switch menuScreen {
	case menuMain:
		switch menuChoice(menuMainLabels) {
		case 0:
			menuError = ""
			startOffline()
		case 1:
			menuError = ""
			startOffline()
			Editing = true
		case 2:
			menuError = ""
			menuScreen = menuHost
		case 3:
			menuError = ""
			menuScreen = menuJoin
		case 4:
			menuError = ""
			menuScreen = menuOptions
		case 5:
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
		if menuButton(4, "Start") {
			hostGame()
		}
		if menuButton(5, "Back") {
			menuBack()
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
		if menuButton(1, "Connect") {
			join(false)
		}
		if menuButton(2, "Join as spectator") {
			join(true)
		}
		if menuButton(3, "Back") {
			menuBack()
		}
	case menuOptions:
		label(0, "Fullscreen")
		fullscreen := Cfg.Fullscreen
		fsText := "Off"
		if fullscreen {
			fsText = "On"
		}
		if rlg.Toggle(row(0), fsText, &fullscreen); fullscreen != Cfg.Fullscreen {
			Cfg.Fullscreen = fullscreen
			ApplySettings()
		}
		label(1, "Window size")
		label(2, "Music volume")
		vol := float32(Cfg.MusicVolume)
		if rlg.Slider(row(2), "", strconv.Itoa(Cfg.MusicVolume), &vol, 0, 100); int(vol) != Cfg.MusicVolume {
			Cfg.MusicVolume = int(vol)
			applyMusicVolume()
		}
		label(3, "Sound volume")
		sfx := float32(Cfg.SoundVolume)
		if rlg.Slider(row(3), "", strconv.Itoa(Cfg.SoundVolume), &sfx, 0, 100); int(sfx) != Cfg.SoundVolume {
			Cfg.SoundVolume = int(sfx)
		}
		// two toggles share the row: the options screen has no room for a seventh
		label(4, "HUD / Light")
		onOff := func(on bool) string {
			if on {
				return "On"
			}
			return "Off"
		}
		left, right := row(4), row(4)
		left.Width = (left.Width - menuRowGap) / 2
		right.Width, right.X = left.Width, left.X+left.Width+menuRowGap
		hud := !Cfg.HideHUD
		rlg.Toggle(left, "F1 HUD: "+onOff(hud), &hud)
		Cfg.HideHUD = !hud
		rlg.Toggle(right, "F3 Light: "+onOff(Cfg.Flashlight), &Cfg.Flashlight)
		if menuButton(5, "Back") {
			menuBack()
		}
		// drawn last so its open list covers the rows below it
		size := int32(Cfg.WindowSize)
		if rlg.DropdownBox(row(1), windowSizesLabel(), &size, winSizeEdit) {
			winSizeEdit = !winSizeEdit
		}
		if int(size) != Cfg.WindowSize {
			Cfg.WindowSize = int(size)
			ApplySettings()
		}
	case menuConnecting:
		text := "Connecting..."
		if Cli != nil {
			text = "Connecting to " + Cli.Addr + "..."
		}
		menuText(text, menuRowX, menuRow(0).Y+12, 20, menuBright)
		if menuButton(1, "Cancel") {
			menuBack()
		}
	}
	if menuError != "" {
		// above the rows (the host screen uses all six), from the label column so long messages stay in the panel
		r := menuRow(0)
		menuText(menuError, r.X-menuLabelW, r.Y-34, 20, menuAccent)
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
	color := menuModelHead
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
	colorJab := menuModelJab
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
	colorJab := menuModelJab
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
	colorJab := menuModelKick
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
	colorJab := menuModelKick
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
