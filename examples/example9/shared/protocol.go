package shared

import (
	"bytes"
	"encoding/binary"
	"errors"
)

const (
	MsgInput    byte = 1
	MsgWelcome  byte = 2
	MsgSnapshot byte = 3
)

const (
	TickRate     = 60
	SnapshotRate = 20

	TickDuration = 1.0 / TickRate

	SnapshotEvery = TickRate / SnapshotRate
)

const (
	PlayerWidth  float32 = 32
	PlayerHeight float32 = 48

	MoveSpeed    float32 = 260
	JumpVelocity float32 = -520

	Gravity      float32 = 1500
	MaxFallSpeed float32 = 900

	GroundFriction float32 = 0.80
)

// ------------------------------------------------------------
// Input
// ------------------------------------------------------------

type Input struct {
	Sequence uint32

	Left  bool
	Right bool
	Jump  bool
}

func EncodeInput(input Input) []byte {
	var b bytes.Buffer

	b.WriteByte(MsgInput)

	_ = binary.Write(
		&b,
		binary.LittleEndian,
		input.Sequence,
	)

	var buttons byte

	if input.Left {
		buttons |= 1 << 0
	}

	if input.Right {
		buttons |= 1 << 1
	}

	if input.Jump {
		buttons |= 1 << 2
	}

	b.WriteByte(buttons)

	return b.Bytes()
}

func DecodeInput(data []byte) (Input, error) {
	if len(data) < 6 {
		return Input{}, errors.New("invalid input packet")
	}

	if data[0] != MsgInput {
		return Input{}, errors.New("not an input packet")
	}

	var input Input

	input.Sequence = binary.LittleEndian.Uint32(
		data[1:5],
	)

	buttons := data[5]

	input.Left = buttons&(1<<0) != 0
	input.Right = buttons&(1<<1) != 0
	input.Jump = buttons&(1<<2) != 0

	return input, nil
}

// ------------------------------------------------------------
// Welcome
// ------------------------------------------------------------

type Welcome struct {
	PlayerID uint16
}

func EncodeWelcome(w Welcome) []byte {
	var b bytes.Buffer

	b.WriteByte(MsgWelcome)

	_ = binary.Write(
		&b,
		binary.LittleEndian,
		w.PlayerID,
	)

	return b.Bytes()
}

func DecodeWelcome(data []byte) (Welcome, error) {
	if len(data) < 3 {
		return Welcome{}, errors.New("invalid welcome packet")
	}

	if data[0] != MsgWelcome {
		return Welcome{}, errors.New("not a welcome packet")
	}

	return Welcome{
		PlayerID: binary.LittleEndian.Uint16(data[1:3]),
	}, nil
}

// ------------------------------------------------------------
// Player state
// ------------------------------------------------------------

type PlayerState struct {
	ID uint16

	X float32
	Y float32

	VX float32
	VY float32

	Grounded bool

	LastProcessedInput uint32
}

// ------------------------------------------------------------
// Snapshot
// ------------------------------------------------------------

type Snapshot struct {
	Tick uint32

	Players []PlayerState
}

func EncodeSnapshot(snapshot Snapshot) []byte {
	var b bytes.Buffer

	b.WriteByte(MsgSnapshot)

	_ = binary.Write(
		&b,
		binary.LittleEndian,
		snapshot.Tick,
	)

	count := uint16(len(snapshot.Players))

	_ = binary.Write(
		&b,
		binary.LittleEndian,
		count,
	)

	for _, player := range snapshot.Players {

		_ = binary.Write(
			&b,
			binary.LittleEndian,
			player.ID,
		)

		_ = binary.Write(
			&b,
			binary.LittleEndian,
			player.X,
		)

		_ = binary.Write(
			&b,
			binary.LittleEndian,
			player.Y,
		)

		_ = binary.Write(
			&b,
			binary.LittleEndian,
			player.VX,
		)

		_ = binary.Write(
			&b,
			binary.LittleEndian,
			player.VY,
		)

		var grounded byte

		if player.Grounded {
			grounded = 1
		}

		b.WriteByte(grounded)

		_ = binary.Write(
			&b,
			binary.LittleEndian,
			player.LastProcessedInput,
		)
	}

	return b.Bytes()
}

func DecodeSnapshot(data []byte) (Snapshot, error) {
	if len(data) < 7 {
		return Snapshot{}, errors.New("invalid snapshot")
	}

	if data[0] != MsgSnapshot {
		return Snapshot{}, errors.New("not a snapshot")
	}

	reader := bytes.NewReader(data[1:])

	var tick uint32
	var count uint16

	if err := binary.Read(
		reader,
		binary.LittleEndian,
		&tick,
	); err != nil {
		return Snapshot{}, err
	}

	if err := binary.Read(
		reader,
		binary.LittleEndian,
		&count,
	); err != nil {
		return Snapshot{}, err
	}

	players := make(
		[]PlayerState,
		0,
		count,
	)

	for i := 0; i < int(count); i++ {

		var player PlayerState

		if err := binary.Read(
			reader,
			binary.LittleEndian,
			&player.ID,
		); err != nil {
			return Snapshot{}, err
		}

		if err := binary.Read(
			reader,
			binary.LittleEndian,
			&player.X,
		); err != nil {
			return Snapshot{}, err
		}

		if err := binary.Read(
			reader,
			binary.LittleEndian,
			&player.Y,
		); err != nil {
			return Snapshot{}, err
		}

		if err := binary.Read(
			reader,
			binary.LittleEndian,
			&player.VX,
		); err != nil {
			return Snapshot{}, err
		}

		if err := binary.Read(
			reader,
			binary.LittleEndian,
			&player.VY,
		); err != nil {
			return Snapshot{}, err
		}

		var grounded byte

		if err := binary.Read(
			reader,
			binary.LittleEndian,
			&grounded,
		); err != nil {
			return Snapshot{}, err
		}

		player.Grounded = grounded != 0

		if err := binary.Read(
			reader,
			binary.LittleEndian,
			&player.LastProcessedInput,
		); err != nil {
			return Snapshot{}, err
		}

		players = append(
			players,
			player,
		)
	}

	return Snapshot{
		Tick:    tick,
		Players: players,
	}, nil
}
