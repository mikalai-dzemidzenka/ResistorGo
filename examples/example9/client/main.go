package main

import (
	"fmt"
	"log"
	"math"
	"resistor/examples/example9/shared"

	"github.com/codecat/go-enet"
	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	ServerIP   = "127.0.0.1"
	ServerPort = 8095

	ScreenWidth  = 1280
	ScreenHeight = 720
)

type Rect struct {
	X float32
	Y float32
	W float32
	H float32
}

type Player struct {
	ID uint16

	X float32
	Y float32

	VX float32
	VY float32

	Grounded bool

	// Remote interpolation.
	RenderX float32
	RenderY float32

	TargetX float32
	TargetY float32
}

type PendingInput struct {
	Input shared.Input
}

var players = map[uint16]*Player{}

var localPlayerID uint16

var connected bool

var sequence uint32

var pendingInputs []PendingInput

var lastSnapshotTick uint32

var platforms = []Rect{
	{
		X: 0,
		Y: 560,
		W: 1280,
		H: 40,
	},

	{
		X: 150,
		Y: 450,
		W: 220,
		H: 30,
	},

	{
		X: 500,
		Y: 380,
		W: 250,
		H: 30,
	},

	{
		X: 850,
		Y: 460,
		W: 250,
		H: 30,
	},

	{
		X: 1030,
		Y: 300,
		W: 150,
		H: 30,
	},
}

// ------------------------------------------------------------
// Main
// ------------------------------------------------------------

func main() {
	enet.Initialize()

	defer enet.Deinitialize()

	host, err := enet.NewHost(
		nil,
		1,
		1,
		0,
		0,
	)

	if err != nil {
		log.Fatal(
			"NewHost:",
			err,
		)
	}

	defer host.Destroy()

	peer, err := host.Connect(
		enet.NewAddress(
			ServerIP,
			ServerPort,
		),
		1,
		0,
	)

	if err != nil {
		log.Fatal(
			"Connect:",
			err,
		)
	}

	rl.InitWindow(
		ScreenWidth,
		ScreenHeight,
		"ENet 2D Platformer",
	)

	defer rl.CloseWindow()

	rl.SetTargetFPS(60)

	fmt.Println(
		"Connecting to",
		ServerIP,
		ServerPort,
	)

	for !rl.WindowShouldClose() {

		// ------------------------------------------------
		// Network
		// ------------------------------------------------

		processNetwork(host)

		// ------------------------------------------------
		// Input
		// ------------------------------------------------

		input := readInput()

		if connected {

			sendInput(peer, input)

			// ------------------------------------------------
			// Client prediction.
			// ------------------------------------------------

			player, ok := players[localPlayerID]

			if ok {
				simulatePlayer(player, input, 1.0/shared.TickRate)
			}
		}

		// ------------------------------------------------
		// Remote interpolation.
		// ------------------------------------------------

		updateRemotePlayers(rl.GetFrameTime())

		// ------------------------------------------------
		// Draw.
		// ------------------------------------------------

		rl.BeginDrawing()

		rl.ClearBackground(rl.NewColor(25, 30, 40, 255))

		drawWorld()
		drawPlayers()
		drawUI()

		rl.EndDrawing()
	}
}

// ------------------------------------------------------------
// Network
// ------------------------------------------------------------

func processNetwork(
	host enet.Host,
) {
	for {

		event := host.Service(0)

		if event.GetType() == enet.EventNone {
			break
		}

		switch event.GetType() {

		case enet.EventConnect:

			fmt.Println(
				"Connected to server",
			)

		case enet.EventDisconnect:

			fmt.Println(
				"Disconnected from server",
			)

			connected = false

		case enet.EventReceive:
			handlePacket(event)
		}
	}
}

func handlePacket(
	event enet.Event,
) {
	packet := event.GetPacket()

	if packet == nil {
		return
	}

	data := packet.GetData()

	// Copy before Destroy.
	packetData := append(
		[]byte(nil),
		data...,
	)

	packet.Destroy()

	if len(packetData) == 0 {
		return
	}

	switch packetData[0] {

	case shared.MsgWelcome:

		handleWelcome(
			packetData,
		)

	case shared.MsgSnapshot:

		handleSnapshot(
			packetData,
		)
	}
}

// ------------------------------------------------------------
// Welcome
// ------------------------------------------------------------

func handleWelcome(
	data []byte,
) {
	welcome, err :=
		shared.DecodeWelcome(data)

	if err != nil {
		log.Println(err)
		return
	}

	localPlayerID =
		welcome.PlayerID

	connected = true

	fmt.Printf(
		"My player ID: %d\n",
		localPlayerID,
	)

	players[localPlayerID] =
		&Player{
			ID: localPlayerID,

			X: 100,
			Y: 200,

			RenderX: 100,
			RenderY: 200,

			TargetX: 100,
			TargetY: 200,
		}
}

// ------------------------------------------------------------
// Snapshot
// ------------------------------------------------------------

func handleSnapshot(
	data []byte,
) {
	snapshot, err :=
		shared.DecodeSnapshot(data)

	if err != nil {
		log.Println(err)
		return
	}

	if snapshot.Tick <
		lastSnapshotTick {
		return
	}

	lastSnapshotTick =
		snapshot.Tick

	seen := make(map[uint16]bool, len(snapshot.Players))
	for _, state := range snapshot.Players {
		seen[state.ID] = true

		player, ok := players[state.ID]

		if !ok {
			player = &Player{
				ID: state.ID,

				X: state.X,
				Y: state.Y,

				RenderX: state.X,
				RenderY: state.Y,

				TargetX: state.X,
				TargetY: state.Y,
			}

			players[state.ID] = player
		}

		player.VX = state.VX
		player.VY = state.VY
		player.Grounded = state.Grounded

		// ------------------------------------------------
		// Local player.
		// ------------------------------------------------

		if state.ID == localPlayerID {

			// Authoritative server state.
			player.X = state.X

			player.Y = state.Y

			// Remove acknowledged inputs.
			remaining := pendingInputs[:0]

			for _, pending := range pendingInputs {

				if pending.Input.Sequence > state.LastProcessedInput {

					remaining = append(remaining, pending)
				}
			}

			pendingInputs = remaining

			// Re-apply inputs that server
			// has not processed yet.
			for _, pending := range pendingInputs {
				simulatePlayer(player, pending.Input, 1.0/shared.TickRate)
			}

			player.RenderX = player.X

			player.RenderY = player.Y

		} else {
			// Remote player.
			player.TargetX = state.X

			player.TargetY = state.Y
		}
	}

	// Remove disconnected players.
	for id := range players {
		if id == localPlayerID {
			continue
		}

		if !seen[id] {
			delete(players, id)
		}
	}
}

// ------------------------------------------------------------
// Input
// ------------------------------------------------------------

func readInput() shared.Input {
	sequence++

	return shared.Input{
		Sequence: sequence,

		Left: rl.IsKeyDown(rl.KeyA) ||
			rl.IsKeyDown(rl.KeyLeft),

		Right: rl.IsKeyDown(rl.KeyD) ||
			rl.IsKeyDown(rl.KeyRight),

		Jump: rl.IsKeyPressed(rl.KeySpace) ||
			rl.IsKeyPressed(rl.KeyW) ||
			rl.IsKeyPressed(rl.KeyUp),
	}
}

func sendInput(
	peer enet.Peer,
	input shared.Input,
) {
	data := shared.EncodeInput(input)

	err := peer.SendBytes(data, 0, 0)

	if err != nil {
		return
	}

	pendingInputs = append(pendingInputs,
		PendingInput{
			Input: input,
		})

	// Safety limit.
	if len(pendingInputs) > 120 {

		pendingInputs = pendingInputs[len(pendingInputs)-120:]
	}
}

// ------------------------------------------------------------
// Physics
// ------------------------------------------------------------

func simulatePlayer(
	player *Player,
	input shared.Input,
	dt float32,
) {
	// ------------------------------------------------
	// Horizontal
	// ------------------------------------------------

	if input.Left &&
		!input.Right {

		player.VX =
			-shared.MoveSpeed

	} else if input.Right &&
		!input.Left {

		player.VX =
			shared.MoveSpeed

	} else {

		player.VX *=
			shared.GroundFriction

		if math.Abs(
			float64(player.VX),
		) < 1 {

			player.VX = 0
		}
	}

	// ------------------------------------------------
	// Jump
	// ------------------------------------------------

	if input.Jump &&
		player.Grounded {

		player.VY =
			shared.JumpVelocity

		player.Grounded =
			false
	}

	// ------------------------------------------------
	// Gravity
	// ------------------------------------------------

	player.VY +=
		shared.Gravity * dt

	if player.VY >
		shared.MaxFallSpeed {

		player.VY =
			shared.MaxFallSpeed
	}

	// ------------------------------------------------
	// Horizontal
	// ------------------------------------------------

	player.X +=
		player.VX * dt

	playerRect := Rect{
		X: player.X,
		Y: player.Y,
		W: shared.PlayerWidth,
		H: shared.PlayerHeight,
	}

	for _, platform := range platforms {

		if !intersects(
			playerRect,
			platform,
		) {
			continue
		}

		if player.VX > 0 {

			player.X =
				platform.X -
					shared.PlayerWidth

		} else if player.VX < 0 {

			player.X =
				platform.X +
					platform.W
		}

		player.VX = 0

		playerRect.X =
			player.X
	}

	// ------------------------------------------------
	// Vertical
	// ------------------------------------------------

	player.Y +=
		player.VY * dt

	player.Grounded = false

	playerRect.X =
		player.X

	playerRect.Y =
		player.Y

	for _, platform := range platforms {

		if !intersects(
			playerRect,
			platform,
		) {
			continue
		}

		if player.VY > 0 {

			player.Y =
				platform.Y -
					shared.PlayerHeight

			player.VY = 0

			player.Grounded = true

		} else if player.VY < 0 {

			player.Y =
				platform.Y +
					platform.H

			player.VY = 0
		}

		playerRect.Y =
			player.Y
	}

	// ------------------------------------------------
	// Bounds
	// ------------------------------------------------

	if player.X < 0 {

		player.X = 0
		player.VX = 0
	}

	if player.X+
		shared.PlayerWidth >
		ScreenWidth {

		player.X =
			ScreenWidth -
				shared.PlayerWidth

		player.VX = 0
	}

	if player.Y >
		ScreenHeight+100 {

		player.X = 100
		player.Y = 200

		player.VX = 0
		player.VY = 0
	}
}

// ------------------------------------------------------------
// Remote interpolation
// ------------------------------------------------------------

func updateRemotePlayers(
	dt float32,
) {
	speed := float32(12)

	t := speed * dt

	if t > 1 {
		t = 1
	}

	for id, player := range players {

		if id == localPlayerID {
			continue
		}

		player.RenderX +=
			(player.TargetX -
				player.RenderX) * t

		player.RenderY +=
			(player.TargetY -
				player.RenderY) * t
	}
}

// ------------------------------------------------------------
// Rendering
// ------------------------------------------------------------

func drawWorld() {
	// Grid.

	for x := 0; x <= ScreenWidth; x += 64 {

		rl.DrawLine(
			int32(x),
			0,
			int32(x),
			ScreenHeight,
			rl.NewColor(
				40,
				45,
				55,
				255,
			),
		)
	}

	for y := 0; y <= ScreenHeight; y += 64 {

		rl.DrawLine(
			0,
			int32(y),
			ScreenWidth,
			int32(y),
			rl.NewColor(
				40,
				45,
				55,
				255,
			),
		)
	}

	// Platforms.

	for _, platform := range platforms {

		rl.DrawRectangle(
			int32(platform.X),
			int32(platform.Y),
			int32(platform.W),
			int32(platform.H),
			rl.NewColor(
				80,
				90,
				105,
				255,
			),
		)

		rl.DrawRectangleLines(
			int32(platform.X),
			int32(platform.Y),
			int32(platform.W),
			int32(platform.H),
			rl.NewColor(
				130,
				140,
				160,
				255,
			),
		)
	}
}

func drawPlayers() {
	for id, player := range players {

		var x float32
		var y float32

		if id == localPlayerID {

			x = player.X
			y = player.Y

		} else {

			x = player.RenderX
			y = player.RenderY
		}

		var color rl.Color

		if id == localPlayerID {

			color = rl.NewColor(
				80,
				220,
				120,
				255,
			)

		} else {

			color = rl.NewColor(
				220,
				100,
				100,
				255,
			)
		}

		rl.DrawRectangle(
			int32(x),
			int32(y),
			int32(shared.PlayerWidth),
			int32(shared.PlayerHeight),
			color,
		)

		rl.DrawText(
			fmt.Sprintf(
				"%d",
				id,
			),
			int32(x),
			int32(y)-22,
			16,
			rl.RayWhite,
		)
	}
}

func drawUI() {
	status := "DISCONNECTED"

	if connected {
		status = "CONNECTED"
	}

	rl.DrawText(
		"ENet 2D Platformer",
		20,
		20,
		24,
		rl.RayWhite,
	)

	rl.DrawText(
		"Status: "+status,
		20,
		52,
		18,
		rl.RayWhite,
	)

	rl.DrawText(
		fmt.Sprintf(
			"Player ID: %d",
			localPlayerID,
		),
		20,
		78,
		18,
		rl.RayWhite,
	)

	rl.DrawText(
		fmt.Sprintf(
			"Players: %d",
			len(players),
		),
		20,
		104,
		18,
		rl.RayWhite,
	)

	rl.DrawText(
		"WASD / Arrow Keys - Move",
		20,
		130,
		18,
		rl.RayWhite,
	)

	rl.DrawText(
		"SPACE / W / UP - Jump",
		20,
		156,
		18,
		rl.RayWhite,
	)

	rl.DrawText(
		fmt.Sprintf(
			"Server tick: %d",
			lastSnapshotTick,
		),
		20,
		182,
		18,
		rl.RayWhite,
	)
}

// ------------------------------------------------------------
// Collision
// ------------------------------------------------------------

func intersects(
	a Rect,
	b Rect,
) bool {
	return a.X < b.X+b.W &&
		a.X+a.W > b.X &&
		a.Y < b.Y+b.H &&
		a.Y+a.H > b.Y
}
