package resistor

import (
	"encoding/binary"
	"errors"
	"hash/fnv"
	"math"
)

const (
	ProtocolVersion  uint16 = 5
	DefaultPort             = 27015
	ChanUnreliable   uint8  = 0 // commands, snapshots
	ChanReliable     uint8  = 1 // hello, welcome, reject, bye
	ChannelCount            = 2
	SnapshotEvery           = 3 // server ticks per snapshot (60/3 = 20 Hz)
	MaxCmdsPerPacket        = 8
)

const (
	MsgHello byte = iota + 1
	MsgWelcome
	MsgReject
	MsgBye
	MsgCmds
	MsgSnapshot
	MsgKeepalive // spectator -> server, keeps an idle connection from timing out
)

const (
	RejectVersion uint8 = iota + 1
	RejectLevel
	RejectFull
	RejectSpectatorsFull
)

// joining role carried in Hello
const (
	RoleFighter uint8 = iota
	RoleSpectator
)

func roleName(r uint8) string {
	if r == RoleSpectator {
		return "spectator"
	}
	return "player"
}

const (
	ByeServerClosed uint8 = iota + 1
	ByeQuit
)

var errPacket = errors.New("malformed packet")

func rejectText(r uint8) string {
	switch r {
	case RejectVersion:
		return "version mismatch"
	case RejectLevel:
		return "level mismatch"
	case RejectFull:
		return "server full"
	case RejectSpectatorsFull:
		return "spectators full"
	}
	return "rejected by server"
}

func byeText(b uint8) string {
	if b == ByeServerClosed {
		return "server closed"
	}
	return "disconnected"
}

func levelHash(b []byte) uint64 {
	h := fnv.New64a()
	h.Write(b)
	return h.Sum64()
}

type Hello struct {
	Proto     uint16
	LevelHash uint64
	Role      uint8
}

type Welcome struct {
	PlayerID uint16
	ActorID  uint16
	Tick     uint32
}

type Snapshot struct {
	Tick   uint32
	Ack    uint32 // last command sequence executed for the receiving player
	Actors []ActorRec
}

// ActorRec is the full replicated state of one actor.
type ActorRec struct {
	ID          uint16
	X, Y        float32
	SX, SY      float32
	State       uint8
	Flags       uint16
	Updown      uint8
	PrevUpdown  uint8
	Anim        uint8
	AnimFrame   uint8
	AnimTotal   uint8
	AnimYOffset int8
	AnimDurMs   uint16
	Ticks       uint32
	AnimStart   uint32
	AttackCount uint8
	FreezeTicks uint8
	Structure   uint8 // tenths of a hit-unit
	Energy      uint8 // whole energy points, 0..MaxEnergy
	EnergyCalm  uint8 // ticks since the last swing, 0..EnergyCalmTicks
	AttackerID  uint16
}

const actorRecSize = 44

const (
	fOrientLeft uint16 = 1 << iota
	fOnGround
	fAnimOnce
	fAnimReverse
	fAnimWhite
	fAttackQueued
	fClimbRight
	fClimbLeft
)

// The landed-hit count toward the next stun (0..2) rides in the two flag bits
// after the ones above, so the record does not grow to carry it.
const (
	fHitCountShift        = 8
	fHitCountMask  uint16 = 3 << fHitCountShift
)

const (
	bLeft uint8 = 1 << iota
	bRight
	bAttack
	bJump
	bFaceLeft
)

func recOf(a *Actor) ActorRec {
	r := ActorRec{
		ID:          a.ID,
		X:           a.Pos.X,
		Y:           a.Pos.Y,
		SX:          a.Speed.X,
		SY:          a.Speed.Y,
		State:       uint8(a.State),
		Updown:      a.Updown,
		PrevUpdown:  a.PrevUpdown,
		Anim:        animID(a.Anim),
		AnimFrame:   a.AnimFrame,
		AnimTotal:   a.AnimTotalFrames,
		AnimYOffset: a.AnimYOffset,
		AnimDurMs:   uint16(math.Round(a.AnimDuration * 1000)),
		Ticks:       a.Ticks,
		AnimStart:   a.AnimStartTick,
		AttackCount: a.AttackCount,
		FreezeTicks: a.FreezeTicks,
		Structure:   uint8(math.Round(float64(clamp(a.Structure, 0, MaxStructure)) * 10)),
		Energy:      uint8(clamp(a.Energy, 0, MaxEnergy)),
		EnergyCalm:  uint8(min(a.EnergyCalm, EnergyCalmTicks)),
	}
	flag := func(f uint16, on bool) {
		if on {
			r.Flags |= f
		}
	}
	flag(fOrientLeft, a.Orientation < 0)
	flag(fOnGround, a.OnGround)
	flag(fAnimOnce, a.AnimOnce)
	flag(fAnimReverse, a.AnimReverse)
	flag(fAnimWhite, a.AnimWhite)
	flag(fAttackQueued, a.AttackQueued)
	flag(fClimbRight, a.CanClimbRight)
	flag(fClimbLeft, a.CanClimbLeft)
	r.Flags |= uint16(min(a.HitCount, StunEveryHits-1)) << fHitCountShift
	if a.Attacker != nil {
		r.AttackerID = a.Attacker.ID
	}
	return r
}

// apply copies the record into a; the attacker is resolved in w (may be nil).
func (r ActorRec) apply(a *Actor, w *World) {
	a.ID = r.ID
	a.Pos.X, a.Pos.Y = r.X, r.Y
	a.Speed.X, a.Speed.Y = r.SX, r.SY
	a.State = ActorState(r.State)
	a.Orientation = 1
	if r.Flags&fOrientLeft != 0 {
		a.Orientation = -1
	}
	a.OnGround = r.Flags&fOnGround != 0
	a.AnimOnce = r.Flags&fAnimOnce != 0
	a.AnimReverse = r.Flags&fAnimReverse != 0
	a.AnimWhite = r.Flags&fAnimWhite != 0
	a.AttackQueued = r.Flags&fAttackQueued != 0
	a.CanClimbRight = r.Flags&fClimbRight != 0
	a.CanClimbLeft = r.Flags&fClimbLeft != 0
	a.HitCount = min(uint8((r.Flags&fHitCountMask)>>fHitCountShift), StunEveryHits-1)
	a.Updown = r.Updown
	a.PrevUpdown = r.PrevUpdown
	a.Anim = animPath(r.Anim)
	a.AnimFrame = r.AnimFrame
	a.AnimTotalFrames = r.AnimTotal
	a.AnimYOffset = r.AnimYOffset
	a.AnimDuration = float64(r.AnimDurMs) / 1000
	a.Ticks = r.Ticks
	a.AnimStartTick = r.AnimStart
	a.AttackCount = r.AttackCount
	a.FreezeTicks = r.FreezeTicks
	a.Structure = float32(r.Structure) / 10
	a.Energy = int32(min(r.Energy, MaxEnergy))
	a.EnergyCalm = uint16(min(r.EnergyCalm, EnergyCalmTicks))
	a.Attacker = nil
	if r.AttackerID != 0 && w != nil {
		a.Attacker = w.ByID(r.AttackerID)
	}
	a.advanceAnim()
}

// --- encoding ---

func putU16(b []byte, v uint16) []byte  { return binary.LittleEndian.AppendUint16(b, v) }
func putU32(b []byte, v uint32) []byte  { return binary.LittleEndian.AppendUint32(b, v) }
func putU64(b []byte, v uint64) []byte  { return binary.LittleEndian.AppendUint64(b, v) }
func putF32(b []byte, v float32) []byte { return putU32(b, math.Float32bits(v)) }

type reader struct {
	b   []byte
	bad bool
}

func newReader(b []byte, typ byte) *reader {
	r := &reader{b: b}
	if r.u8() != typ {
		r.bad = true
	}
	return r
}

func (r *reader) take(n int) []byte {
	if r.bad || len(r.b) < n {
		r.bad = true
		return nil
	}
	v := r.b[:n]
	r.b = r.b[n:]
	return v
}

func (r *reader) u8() uint8 {
	if v := r.take(1); v != nil {
		return v[0]
	}
	return 0
}

func (r *reader) u16() uint16 {
	if v := r.take(2); v != nil {
		return binary.LittleEndian.Uint16(v)
	}
	return 0
}

func (r *reader) u32() uint32 {
	if v := r.take(4); v != nil {
		return binary.LittleEndian.Uint32(v)
	}
	return 0
}

func (r *reader) u64() uint64 {
	if v := r.take(8); v != nil {
		return binary.LittleEndian.Uint64(v)
	}
	return 0
}

func (r *reader) f32() float32 { return math.Float32frombits(r.u32()) }

func (r *reader) err() error {
	if r.bad {
		return errPacket
	}
	return nil
}

func encodeHello(h Hello) []byte {
	b := []byte{MsgHello}
	b = putU16(b, h.Proto)
	b = putU64(b, h.LevelHash)
	return append(b, h.Role)
}

// decodeHello accepts the version-1 layout without a role byte, so an old
// client can still be told "version mismatch" instead of being ignored.
func decodeHello(b []byte) (Hello, error) {
	r := newReader(b, MsgHello)
	h := Hello{Proto: r.u16(), LevelHash: r.u64()}
	if !r.bad && len(r.b) > 0 {
		h.Role = r.u8()
		if h.Role > RoleSpectator {
			r.bad = true
		}
	}
	return h, r.err()
}

func encodeKeepalive() []byte { return []byte{MsgKeepalive} }

func encodeWelcome(w Welcome) []byte {
	b := []byte{MsgWelcome}
	b = putU16(b, w.PlayerID)
	b = putU16(b, w.ActorID)
	return putU32(b, w.Tick)
}

func decodeWelcome(b []byte) (Welcome, error) {
	r := newReader(b, MsgWelcome)
	w := Welcome{PlayerID: r.u16(), ActorID: r.u16(), Tick: r.u32()}
	return w, r.err()
}

func encodeReject(reason uint8) []byte { return []byte{MsgReject, reason} }

func decodeReject(b []byte) (uint8, error) {
	r := newReader(b, MsgReject)
	v := r.u8()
	return v, r.err()
}

func encodeBye(reason uint8) []byte { return []byte{MsgBye, reason} }

func decodeBye(b []byte) (uint8, error) {
	r := newReader(b, MsgBye)
	v := r.u8()
	return v, r.err()
}

func encodeCmds(cmds []Input) []byte {
	b := []byte{MsgCmds, uint8(len(cmds))}
	for _, c := range cmds {
		var buttons uint8
		if c.Left {
			buttons |= bLeft
		}
		if c.Right {
			buttons |= bRight
		}
		if c.Attack {
			buttons |= bAttack
		}
		if c.Jump {
			buttons |= bJump
		}
		if c.FaceLeft {
			buttons |= bFaceLeft
		}
		b = putU32(b, c.Seq)
		b = append(b, buttons, c.Updown)
		b = putU32(b, c.ViewTick)
		b = append(b, c.ViewFrac)
	}
	return b
}

func decodeCmds(b []byte) ([]Input, error) {
	r := newReader(b, MsgCmds)
	n := int(r.u8())
	if n == 0 || n > MaxCmdsPerPacket {
		return nil, errPacket
	}
	cmds := make([]Input, 0, n)
	for range n {
		var c Input
		c.Seq = r.u32()
		buttons := r.u8()
		c.Updown = r.u8()
		c.ViewTick = r.u32()
		c.ViewFrac = r.u8()
		c.Left = buttons&bLeft != 0
		c.Right = buttons&bRight != 0
		c.Attack = buttons&bAttack != 0
		c.Jump = buttons&bJump != 0
		c.FaceLeft = buttons&bFaceLeft != 0
		if c.Updown > 3 {
			r.bad = true
		}
		cmds = append(cmds, c)
	}
	if err := r.err(); err != nil {
		return nil, err
	}
	return cmds, nil
}

func encodeSnapshot(s Snapshot) []byte {
	b := make([]byte, 0, 10+len(s.Actors)*actorRecSize)
	b = append(b, MsgSnapshot)
	b = putU32(b, s.Tick)
	b = putU32(b, s.Ack)
	b = append(b, uint8(len(s.Actors)))
	for _, a := range s.Actors {
		b = putU16(b, a.ID)
		b = putF32(b, a.X)
		b = putF32(b, a.Y)
		b = putF32(b, a.SX)
		b = putF32(b, a.SY)
		b = append(b, a.State)
		b = putU16(b, a.Flags)
		b = append(b, a.Updown, a.PrevUpdown, a.Anim, a.AnimFrame, a.AnimTotal, uint8(a.AnimYOffset))
		b = putU16(b, a.AnimDurMs)
		b = putU32(b, a.Ticks)
		b = putU32(b, a.AnimStart)
		b = append(b, a.AttackCount, a.FreezeTicks, a.Structure, a.Energy, a.EnergyCalm)
		b = putU16(b, a.AttackerID)
	}
	return b
}

func decodeSnapshot(b []byte) (Snapshot, error) {
	r := newReader(b, MsgSnapshot)
	s := Snapshot{Tick: r.u32(), Ack: r.u32()}
	n := int(r.u8())
	if r.bad || len(r.b) < n*actorRecSize {
		return Snapshot{}, errPacket
	}
	s.Actors = make([]ActorRec, 0, n)
	for range n {
		var a ActorRec
		a.ID = r.u16()
		a.X = r.f32()
		a.Y = r.f32()
		a.SX = r.f32()
		a.SY = r.f32()
		a.State = r.u8()
		a.Flags = r.u16()
		a.Updown = r.u8()
		a.PrevUpdown = r.u8()
		a.Anim = r.u8()
		a.AnimFrame = r.u8()
		a.AnimTotal = r.u8()
		a.AnimYOffset = int8(r.u8())
		a.AnimDurMs = r.u16()
		a.Ticks = r.u32()
		a.AnimStart = r.u32()
		a.AttackCount = r.u8()
		a.FreezeTicks = r.u8()
		a.Structure = r.u8()
		a.Energy = r.u8()
		a.EnergyCalm = r.u8()
		a.AttackerID = r.u16()
		if a.State > uint8(StateDeath) || a.Updown > 3 {
			r.bad = true
		}
		s.Actors = append(s.Actors, a)
	}
	if err := r.err(); err != nil {
		return Snapshot{}, err
	}
	return s, nil
}
