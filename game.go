package resistor

import (
	"encoding/json"
	"fmt"
	"github.com/codecat/go-enet"
	rlg "github.com/gen2brain/raylib-go/raygui"
	rl "github.com/gen2brain/raylib-go/raylib"
	"slices"
	"strconv"
	"strings"
)

const (
	SW          = 160
	SH          = 90
	WindowScale = 8
)

const (
	StageMainMenu = iota
	StageLevel
)

const (
	NetOffline = iota
	NetListen
	NetClient
	NetDedicated
)

// LevelPath is the level the game plays. It is the one level there has ever
// been, unless a development build was given another with -level: a test course
// of walls, ledges and holes is worth playing a bot against, and is not worth
// putting into the level people are actually building. Loading and saving both
// go through it, so a level opened this way cannot be saved over level_1.json.
var LevelPath = "level_1.json"

var (
	Debug                = DevTools
	GameTimeSec          float64
	Accumulator          float64
	PhysicsTime          float64
	FixedTimeStep        float64 = 1.0 / 60.0
	MouseWorldPos        Vector2i
	MousePos             rl.Vector2
	CamOffset                        = rl.Vector2{X: SW * WindowScale * 0.9 / 2, Y: SH * WindowScale / 2}
	Cam                  Camera      = NewCamera(3, CamOffset)
	LE                   LevelEditor = NewLevelEditor(39, 83)
	Stage                int         = StageMainMenu
	MainMenuTexture      rl.RenderTexture2D
	MenuBackground       rl.RenderTexture2D
	MenuFrame            rl.RenderTexture2D
	MenuPaletteTexture   rl.Texture2D
	ScreenFullTexture    rl.RenderTexture2D
	ScreenPaletteTexture rl.Texture2D
	PaletteShader        rl.Shader
	Sh                   *Shaders
	sharedLight          *ZimaLight         // one set of shadow textures, taken in turn by every light
	LightAccum           rl.RenderTexture2D // every light in view, added up
	Cur                  *Cursor            = &Cursor{
		AtkPos: rl.Vector2{SW * WindowScale / 2, SH * WindowScale / 2},
		energy: 1,
		GUIZ:   2,
	}
	cursorOwner   *Actor // the local fighter whose energy Cur was last given
	WhiteShader   rl.Shader
	OutlineShader rl.Shader
	RimShader     rl.Shader
	NetMode       = NetOffline
	quit          bool
	drawList      []*Actor
)

type Vector2i struct {
	X int32
	Y int32
}

func (v Vector2i) String() string {
	return fmt.Sprintf("%d,%d", v.X, v.Y)
}

func (v Vector2i) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.String())
}

func (v *Vector2i) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	tokens := strings.Split(s, ",")
	if len(tokens) != 2 {
		return fmt.Errorf("malformed input: %s", b)
	}
	x, errX := strconv.Atoi(tokens[0])
	y, errY := strconv.Atoi(tokens[1])
	if errX != nil || errY != nil {
		return fmt.Errorf("malformed input: %s", b)
	}
	v.X = int32(x)
	v.Y = int32(y)
	return nil
}

func v2i(v rl.Vector2) Vector2i {
	return Vector2i{
		X: int32(v.X),
		Y: int32(v.Y),
	}
}

func v2f(v Vector2i) rl.Vector2 {
	return rl.Vector2{
		X: float32(v.X),
		Y: float32(v.Y),
	}
}

func Game() {
	LoadSettings()
	rl.InitWindow(VW, VH, "Resistor")
	defer rl.CloseWindow()
	rl.SetExitKey(0)

	loadStyle()

	rl.SetTargetFPS(60)

	initPresent()
	initAudio()
	defer closeAudio()
	ApplySettings()

	LoadTextures()

	defer shutdownNet()

	LE = NewLevelEditor(39, 83)
	LE.LoadLevel(LevelPath)

	for !rl.WindowShouldClose() && !quit {
		if DevTools && rl.IsKeyPressed(rl.KeyGrave) {
			Debug = !Debug
		}
		dt := rl.GetFrameTime()
		if dt > 0.25 {
			dt = 0.25
		}

		GameTimeSec += float64(dt)
		MousePos = rl.GetMousePosition()
		MouseWorldPos = v2i(rl.GetScreenToWorld2D(MousePos, Cam.View()))
		// A dropped file always comes off the queue, so one dropped during a
		// match is not waiting to be imported when edit mode is entered. Only
		// edit mode acts on it.
		if rl.IsFileDropped() {
			dropped := rl.LoadDroppedFiles()
			if DevTools && Editing && Stage == StageLevel {
				LE.importDropped(dropped)
			}
		}
		updateAudio(dt)
		switch Stage {
		case StageMainMenu:
			runMainMenu(dt)
		case StageLevel:
			runLevel(dt)
		}
		present()
	}
	leaveMatch("", menuMain)
}

// netReady is whether ENet has been started. Single player never touches the
// network, so it is started on demand rather than at launch: the first host or
// join pays for it, and a session that only ever plays offline does not.
var netReady bool

// initNet starts ENet if it has not been started already.
func initNet() {
	if netReady {
		return
	}
	enet.Initialize()
	netReady = true
}

// shutdownNet stops ENet, if anything ever started it.
func shutdownNet() {
	if !netReady {
		return
	}
	enet.Deinitialize()
	netReady = false
}

func loadStyle() {
	rlg.LoadStyleFromMemory(asset("resources/style_dark.rgs"))
	rlg.SetStyle(rlg.DEFAULT, rlg.TEXT_SIZE, 20)
}

// OfflineBots is how many bots a single-player match puts in the level with the
// local fighter. Each fights whoever is nearest, bot or not, and comes back
// after RespawnTicks like anyone else, so the count holds for the whole match.
const OfflineBots = 3

// startOffline starts single player: one authoritative world, the local fighter
// and OfflineBots bots on the free spawn columns beside it.
func startOffline() {
	NetMode = NetOffline
	CL = NewWorld(true)
	CL.Local = CL.Spawn(rl.Vector2{}, true)
	for range OfflineBots {
		CL.Spawn(CL.spawnPos(), false)
	}
	enterLevel()
}

func enterLevel() {
	Stage = StageLevel
	Accumulator = 0
	Spec = SpecCam{Follow: true, Pos: Cam.Camera2D.Target}
	// the menu paints the raygui widgets in its own reds (menuWidgetStyle);
	// the editor and the level HUD get the dark style back
	loadStyle()
	rl.DisableCursor()
}

// leaveMatch tears down any network session and returns to the menu.
func leaveMatch(reason string, screen int) {
	if Cli != nil {
		Cli.Close()
		Cli = nil
	}
	if Srv != nil {
		Srv.Close()
		Srv = nil
	}
	CL, SV = nil, nil
	NetMode = NetOffline
	ResetTime()
	Editing = false
	Stage = StageMainMenu
	menuScreen = screen
	menuError = reason
	if rl.IsWindowReady() {
		rl.EnableCursor()
		loadStyle()
	}
}

func netFrame(dt float32) {
	if Srv != nil {
		Srv.Frame(dt)
	}
	if Cli != nil {
		Cli.Frame(dt)
	}
}

func stepOffline(dt float32) {
	Accumulator += float64(dt)
	for Accumulator > FixedTimeStep {
		PhysicsTime += FixedTimeStep
		CL.Tick++
		CL.updateRespawns()
		rewindRecord()
		for _, actor := range CL.Actors {
			actor.recordRewind()
			if actor.IsPlayer {
				// A death nils CL.Local through World.Remove, and the respawn
				// spawns a fighter nothing points at. Adopting it here is the
				// offline answer to the question a snapshot answers over the
				// wire: whose fighter is this. Without it the player keeps
				// control - input routes on IsPlayer, not on Local - while the
				// camera and every effect keyed on the local fighter stay with
				// the corpse.
				if CL.Local == nil {
					CL.Local = actor
				}
				actor.Input = sampleInput(actor)
			} else {
				actor.Input = botInput(actor)
			}
			actor.Update(float32(FixedTimeStep))
		}
		Accumulator -= FixedTimeStep
	}
}

// presentEffects turns simulation events and states into local-only camera effects.
func presentEffects(w *World) {
	local := w.Local
	shake := false
	for _, a := range w.Actors {
		if a == local && a.Events&EvAttackStart != 0 {
			Cam.ZoomSpeed += AttackZoomSpeed
		}
		if a.Events&EvNoEnergy != 0 {
			a.DenyUntil = GameTimeSec + EnergyDenySec
		}
		a.Events = 0
		if a.State == StateAttacked && local != nil && (a == local || a.Attacker == local) {
			shake = true
		}
	}
	Cam.Offset = CamOffset
	if shake {
		const offset = 2
		Cam.Offset.Y += []float32{-offset, offset, -offset, offset, -offset, offset, -offset, offset}[int(GameTimeSec*1000)%8]
		Cam.Offset.X += []float32{offset, -offset, offset, -offset, offset, -offset, offset, -offset}[int(GameTimeSec*1000)%8]
	}
}

var levelInit bool

func runLevel(dt float32) {
	if !levelInit {
		levelInit = true
		WhiteShader = rl.LoadShaderFromMemory("", fsWhite)
		OutlineShader = rl.LoadShaderFromMemory("", fsOutline)
		loadRimShader()
		LE.LoadLevel(LevelPath)

		Sh = loadShaders()
		// 512 is the resolution every light has always been drawn at; radius,
		// color and intensity are set per light as it takes its turn
		sharedLight = newZimaLight(512, 128, rl.White, 1)

		ScreenFullTexture = rl.LoadRenderTexture(SW*WindowScale, SH*WindowScale)
		LightAccum = rl.LoadRenderTexture(SW*WindowScale, SH*WindowScale)
		cornerFade = newCornerFade()
	}
	if rl.IsKeyPressed(rl.KeyEscape) {
		leaveMatch("", menuMain)
		return
	}
	if rl.IsKeyPressed(HUDKey) {
		toggleHUD()
	}

	spectating := Cli != nil && Cli.Spectator
	if !spectating && !matchHeld() {
		latchInput()
	}
	if NetMode == NetOffline {
		if DevTools && rl.IsKeyPressed(rl.KeyF2) {
			toggleEditing(true, TimeMode != TimeOff)
		}
		// Time manipulation is single player only, and holds the simulation
		// while it is on: no step, and no dt banked up for one either.
		if UpdateTime(rl.IsKeyPressed(rl.KeyR), rl.GetMouseDelta().X, dt) {
			for _, actor := range CL.Actors {
				actor.restoreRewind()
			}
		}
		if !matchHeld() {
			stepOffline(dt)
		}
	} else {
		netFrame(dt)
		if Stage != StageLevel || CL == nil {
			return
		}
	}
	w := CL
	local := w.Local

	alpha := float32(Accumulator / FixedTimeStep)

	energy := float32(1) // with no fighter of our own, the cross is as if rested
	if local != nil {
		energy = float32(local.Energy) / MaxEnergy
	}
	// a new fighter or a rewound frame is not energy won back
	live := local != nil && local == cursorOwner && !TimeShowsHistory()
	cursorOwner = local
	Cur.Update(dt, energy, live)
	for _, actor := range w.Actors {
		if !actor.Remote {
			actor.Interp(alpha)
		}
	}
	if Cli != nil && local != nil {
		Cli.smoothCorrection(dt, local)
	}
	presentEffects(w)

	if Editing {
		LE.Update(dt)
	}
	if DevTools && Debug && Cli != nil {
		Cli.Sim.controls()
	}
	if local != nil {
		Cam.SetTarget(rl.Vector2{X: local.Pos.X + 24, Y: local.Pos.Y + 24})
	} else if spectating {
		updateSpectatorCam(dt, w)
	}
	Cam.Update(dt)

	occluders := LE.LightOccluders()
	computeLights(LE.lights, LE.areas, LE.ambient, occluders, Cam.View())
	rl.BeginTextureMode(ScreenFullTexture)
	{
		rl.ClearBackground(rl.Black)
		rl.BeginMode2D(Cam.View())

		LE.Draw() // TODO sort objects with actors

		// sort a copy: the simulation order of w.Actors must stay stable
		drawList = append(drawList[:0], w.Actors...)
		slices.SortFunc(drawList, func(a, b *Actor) int {
			return int(b.Pos.X - a.Pos.X)
		})
		locTime := rl.GetShaderLocation(WhiteShader, "time")
		for _, a := range drawList {
			if a.Flickering() {
				rl.BeginShaderMode(WhiteShader)
				rl.SetShaderValue(WhiteShader, locTime, []float32{float32(rl.GetTime())}, rl.ShaderUniformFloat)
				a.Draw()
				rl.EndShaderMode()
			} else {
				a.Draw()
			}
		}
		if DevTools && Debug && Srv != nil {
			for _, r := range Srv.LastRewound {
				rl.DrawRectangleLinesEx(r, 1, rl.Orange)
			}
		}

		rl.EndMode2D()
		compositeLights()
		rl.BeginMode2D(Cam.View())

		// the rim is light added on top of the lit scene, so the ambient dark
		// cannot dim it; the overlays below go over it
		rl.BeginBlendMode(rl.BlendAdditive)
		for _, a := range drawList {
			drawRim(a, LE.lights, occluders)
		}
		rl.EndBlendMode()

		// A bot's sight is drawn here rather than with the other debug boxes,
		// which are inside the lit scene: the cone is a faint wash, and
		// multiplied by the light map it would disappear into any dark corner -
		// exactly where knowing what the bot can see matters most.
		if DevTools && Debug {
			for _, a := range drawList {
				a.drawSight()
			}
		}

		// the attack outline and the structure bars go on top of the lighting,
		// so it cannot tint them; the outline first, under the bars
		for _, a := range drawList {
			drawAttackOutline(w.Local, a)
		}
		for _, a := range drawList {
			a.drawStructureBar()
		}
		LE.DrawOverlay()

		rl.EndMode2D()
	}
	rl.EndTextureMode()

	rl.BeginTextureMode(VirtualScreen)
	rl.ClearBackground(rl.Black)

	locTexture1 := rl.GetShaderLocation(PaletteShader, "texture1")
	locPaletteSize := rl.GetShaderLocation(PaletteShader, "paletteSize")
	locDither := rl.GetShaderLocation(PaletteShader, "ditherStrength")
	locResolution := rl.GetShaderLocation(PaletteShader, "resolution")
	locPixelSize := rl.GetShaderLocation(PaletteShader, "pixelSize")
	rl.BeginShaderMode(PaletteShader)
	rl.SetShaderValueTexture(PaletteShader, locTexture1, ScreenPaletteTexture)
	rl.SetShaderValue(PaletteShader, locResolution, []float32{SW * WindowScale, SH * WindowScale}, rl.ShaderUniformVec2)
	rl.SetShaderValue(PaletteShader, locPixelSize, []float32{Cam.PixelSize()}, rl.ShaderUniformFloat)
	rl.SetShaderValue(PaletteShader, locPaletteSize, []float32{float32(len(Palette))}, rl.ShaderUniformFloat)
	rl.SetShaderValue(PaletteShader, locDither, []float32{0}, rl.ShaderUniformFloat)

	src := rl.NewRectangle(0, 0, SW*WindowScale, -SH*WindowScale) // флип render-текстуры
	dst := rl.NewRectangle(0, 0, SW*WindowScale, SH*WindowScale)
	rl.DrawTexturePro(ScreenFullTexture.Texture, src, dst, rl.Vector2{}, 0, rl.White)
	rl.EndShaderMode()
	if Editing {
		LE.DrawGui()
		drawEditMarker()
	}
	if spectating {
		drawSpectatorHUD()
	} else {
		Cur.DrawGUI()
	}

	if DevTools && Debug {
		fps := strconv.Itoa(int(rl.GetFPS()))
		rl.DrawText("fps:"+fps, 20, 20, 2, rl.White)
		rl.DrawText(fmt.Sprintf("gameTime:%f", GameTimeSec), 20, 30, 2, rl.White)
		rl.DrawText(fmt.Sprintf("physTime:%f", PhysicsTime), 20, 40, 2, rl.White)
		rl.DrawText(fmt.Sprintf("camera zoom speed %f", Cam.ZoomSpeed), 20, 50, 2, rl.White)
		drawNetDebug()
	}
	rl.EndTextureMode()
}

func drawNetDebug() {
	y := int32(60)
	line := func(s string) {
		rl.DrawText(s, 20, y, 2, rl.White)
		y += 10
	}
	if c := Cli; c != nil {
		rtt := 0
		if c.Welcomed {
			rtt = int(c.Peer.GetRoundTripTime())
		}
		line(fmt.Sprintf("rtt: %d ms (enet %d + sim %d)  loss sim: %d%%  [F5/F6 lag, F7/F8 loss]",
			rtt+c.Sim.LatencyMs, rtt, c.Sim.LatencyMs, c.Sim.LossPct))
		line(fmt.Sprintf("interp: %.0f ms  snapshots/s: %d  corrections: %d",
			c.Interp.Delay*FixedTimeStep*1000, c.snapRate(), c.Corrections))
		line(fmt.Sprintf("render tick: %.1f  newest snapshot: %d  cmd seq: %d  unacked cmds: %d", c.Interp.RenderTick(), c.LastSnap, c.Seq, c.Seq-c.LastAck))
	}
	if s := Srv; s != nil {
		line(fmt.Sprintf("server tick: %d  humans: %d  bots: %d  lagcomp: %t", s.W.Tick, s.humans(), len(s.Bots), s.Cfg.LagComp))
	}
}
