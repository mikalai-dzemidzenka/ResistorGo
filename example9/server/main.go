package main

import (
	"fmt"
	"log"
	"math"
	"time"

	"github.com/codecat/go-enet"

	"resistor/example9/shared"
)

const (
	ServerPort = 8095

	WorldWidth  float32 = 1280
	WorldHeight float32 = 720
)

type Rect struct {
	X float32
	Y float32
	W float32
	H float32
}

type Player struct {
	ID uint16

	PeerID uint16
	Peer   enet.Peer

	X float32
	Y float32

	VX float32
	VY float32

	Grounded bool

	Input shared.Input
}

var players = map[uint16]*Player{}

var nextPlayerID uint16 = 1

var tick uint32

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

func main() {
	enet.Initialize()

	defer enet.Deinitialize()

	host, err := enet.NewHost(
		enet.NewListenAddress(ServerPort),
		32,
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

	fmt.Println("================================")
	fmt.Println("      ENET PLATFORMER SERVER")
	fmt.Println("================================")
	fmt.Println()
	fmt.Printf(
		"Listening on 0.0.0.0:%d\n",
		ServerPort,
	)
	fmt.Println()

	ticker := time.NewTicker(
		time.Second / shared.TickRate,
	)

	defer ticker.Stop()

	for {
		// ------------------------------------------------
		// Process all pending network events.
		// ------------------------------------------------

		for {
			event := host.Service(0)

			if event.GetType() == enet.EventNone {
				break
			}

			handleEvent(event)
		}

		// ------------------------------------------------
		// Fixed update.
		// ------------------------------------------------

		select {
		case <-ticker.C:
			update()

		default:
		}
	}
}

// ------------------------------------------------------------
// Network
// ------------------------------------------------------------

func handleEvent(event enet.Event) {
	switch event.GetType() {

	case enet.EventConnect:
		handleConnect(event)

	case enet.EventDisconnect:
		handleDisconnect(event)

	case enet.EventReceive:
		handleReceive(event)
	}
}

func handleConnect(event enet.Event) {
	peer := event.GetPeer()

	peerID := peer.GetIncomingPeerId()

	playerID := nextPlayerID
	nextPlayerID++

	player := &Player{
		ID:     playerID,
		PeerID: peerID,
		Peer:   peer,

		X: 100 + float32(playerID-1)*80,
		Y: 300,

		VX: 0,
		VY: 0,

		Grounded: false,
	}

	players[peerID] = player

	fmt.Printf(
		"CONNECT peer=%d player=%d address=%s\n",
		peerID,
		playerID,
		peer.GetAddress(),
	)

	data := shared.EncodeWelcome(
		shared.Welcome{
			PlayerID: playerID,
		},
	)

	err := peer.SendBytes(
		data,
		0,
		enet.PacketFlagReliable,
	)

	if err != nil {
		log.Printf(
			"send welcome: %v",
			err,
		)
	}
}

func handleDisconnect(event enet.Event) {
	peer := event.GetPeer()

	peerID := peer.GetIncomingPeerId()

	player, ok := players[peerID]

	if !ok {
		return
	}

	fmt.Printf(
		"DISCONNECT peer=%d player=%d\n",
		peerID,
		player.ID,
	)

	delete(
		players,
		peerID,
	)
}

func handleReceive(event enet.Event) {
	packet := event.GetPacket()

	if packet == nil {
		return
	}

	// IMPORTANT:
	// packet must be destroyed after GetData().
	data := packet.GetData()

	// Copy the bytes before destroying the packet.
	// DecodeInput doesn't retain the slice, so this is
	// technically enough for this immediate decode.
	inputData := append(
		[]byte(nil),
		data...,
	)

	packet.Destroy()

	if len(inputData) == 0 {
		return
	}

	if inputData[0] != shared.MsgInput {
		return
	}

	input, err := shared.DecodeInput(
		inputData,
	)

	if err != nil {
		return
	}

	peer := event.GetPeer()

	peerID := peer.GetIncomingPeerId()

	player, ok := players[peerID]

	if !ok {
		return
	}

	// Ignore an old input packet.
	if input.Sequence < player.Input.Sequence {
		return
	}

	player.Input = input
}

// ------------------------------------------------------------
// Update
// ------------------------------------------------------------

func update() {
	tick++

	const dt float32 = 1.0 / shared.TickRate

	for _, player := range players {
		updatePlayer(
			player,
			dt,
		)
	}

	// 60 / 20 = 3
	if tick%shared.SnapshotEvery == 0 {
		broadcastSnapshot()
	}
}

// ------------------------------------------------------------
// Physics
// ------------------------------------------------------------

func updatePlayer(
	player *Player,
	dt float32,
) {
	input := player.Input

	// ------------------------------------------------
	// Horizontal
	// ------------------------------------------------

	if input.Left && !input.Right {

		player.VX = -shared.MoveSpeed

	} else if input.Right && !input.Left {

		player.VX = shared.MoveSpeed

	} else {

		player.VX *= shared.GroundFriction

		if math.Abs(
			float64(player.VX),
		) < 1 {
			player.VX = 0
		}
	}

	// ------------------------------------------------
	// Jump
	// ------------------------------------------------

	if input.Jump && player.Grounded {

		player.VY = shared.JumpVelocity

		player.Grounded = false
	}

	// ------------------------------------------------
	// Gravity
	// ------------------------------------------------

	player.VY += shared.Gravity * dt

	if player.VY > shared.MaxFallSpeed {
		player.VY = shared.MaxFallSpeed
	}

	// ------------------------------------------------
	// Horizontal movement
	// ------------------------------------------------

	player.X += player.VX * dt

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

		playerRect.X = player.X
	}

	// ------------------------------------------------
	// Vertical movement
	// ------------------------------------------------

	player.Y += player.VY * dt

	player.Grounded = false

	playerRect.X = player.X
	playerRect.Y = player.Y

	for _, platform := range platforms {

		if !intersects(
			playerRect,
			platform,
		) {
			continue
		}

		if player.VY > 0 {

			// Falling.
			player.Y =
				platform.Y -
					shared.PlayerHeight

			player.VY = 0

			player.Grounded = true

		} else if player.VY < 0 {

			// Ceiling.
			player.Y =
				platform.Y +
					platform.H

			player.VY = 0
		}

		playerRect.Y = player.Y
	}

	// ------------------------------------------------
	// World bounds
	// ------------------------------------------------

	if player.X < 0 {
		player.X = 0
		player.VX = 0
	}

	if player.X+shared.PlayerWidth > WorldWidth {

		player.X =
			WorldWidth -
				shared.PlayerWidth

		player.VX = 0
	}

	if player.Y > WorldHeight+100 {

		player.X = 100
		player.Y = 200

		player.VX = 0
		player.VY = 0
	}
}

// ------------------------------------------------------------
// Snapshot
// ------------------------------------------------------------

func broadcastSnapshot() {
	snapshot := shared.Snapshot{
		Tick: tick,

		Players: make(
			[]shared.PlayerState,
			0,
			len(players),
		),
	}

	for _, player := range players {

		snapshot.Players = append(
			snapshot.Players,
			shared.PlayerState{
				ID: player.ID,

				X: player.X,
				Y: player.Y,

				VX: player.VX,
				VY: player.VY,

				Grounded: player.Grounded,

				LastProcessedInput: player.Input.Sequence,
			},
		)
	}

	data := shared.EncodeSnapshot(
		snapshot,
	)

	for _, player := range players {

		err := player.Peer.SendBytes(
			data,
			0,
			0,
		)

		if err != nil {
			log.Printf(
				"snapshot player=%d: %v",
				player.ID,
				err,
			)
		}
	}
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
