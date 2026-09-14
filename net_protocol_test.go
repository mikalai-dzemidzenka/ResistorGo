package resistor

import (
	"fmt"
	"os"
	"reflect"
	"regexp"
	"slices"
	"testing"
)

func sampleRec(id uint16) ActorRec {
	return ActorRec{
		ID: id, X: 12, Y: -34, SX: 1.5, SY: -900,
		State: uint8(StateAttack), Flags: fOrientLeft | fAnimOnce | fAttackQueued | fBufAttack | fBufJump,
		Updown: 2, PrevUpdown: 1, Anim: 17, AnimFrame: 3, AnimTotal: 8, AnimYOffset: -48,
		AnimDurMs: 300, Ticks: 123456, AnimStart: 123400, AttackCount: 2, FreezeTicks: 5, Structure: 73, Energy: 45, AttackerID: 7,
	}
}

func TestProtocolRoundTrip(t *testing.T) {
	h := Hello{Proto: ProtocolVersion, LevelHash: 0xdeadbeefcafef00d, Role: RoleSpectator}
	if got, err := decodeHello(encodeHello(h)); err != nil || got != h {
		t.Errorf("hello: %+v %v", got, err)
	}
	w := Welcome{PlayerID: 3, ActorID: 9, Tick: 777}
	if got, err := decodeWelcome(encodeWelcome(w)); err != nil || got != w {
		t.Errorf("welcome: %+v %v", got, err)
	}
	if got, err := decodeReject(encodeReject(RejectLevel)); err != nil || got != RejectLevel {
		t.Errorf("reject: %v %v", got, err)
	}
	if got, err := decodeBye(encodeBye(ByeServerClosed)); err != nil || got != ByeServerClosed {
		t.Errorf("bye: %v %v", got, err)
	}
	cmds := []Input{
		{Seq: 10, Left: true, Attack: true, Updown: 3, ViewTick: 99, ViewFrac: 128},
		{Seq: 9, Right: true, Jump: true, FaceLeft: true, Updown: 1, ViewTick: 98},
	}
	if got, err := decodeCmds(encodeCmds(cmds)); err != nil || !slices.Equal(got, cmds) {
		t.Errorf("cmds: %+v %v", got, err)
	}
	s := Snapshot{Tick: 42, Ack: 500, Actors: []ActorRec{sampleRec(1), sampleRec(2)}}
	if got, err := decodeSnapshot(encodeSnapshot(s)); err != nil || !reflect.DeepEqual(got, s) {
		t.Errorf("snapshot: %+v %v", got, err)
	}
}

func TestProtocolRejectsMalformed(t *testing.T) {
	cases := []struct {
		name   string
		data   []byte
		decode func([]byte) error
	}{
		// the role byte is optional (version-1 layout), so truncate below the 11-byte core
		{"hello", encodeHello(Hello{Proto: 1, LevelHash: 2})[:11], func(b []byte) error { _, err := decodeHello(b); return err }},
		{"welcome", encodeWelcome(Welcome{1, 2, 3}), func(b []byte) error { _, err := decodeWelcome(b); return err }},
		{"reject", encodeReject(RejectFull), func(b []byte) error { _, err := decodeReject(b); return err }},
		{"bye", encodeBye(ByeQuit), func(b []byte) error { _, err := decodeBye(b); return err }},
		{"cmds", encodeCmds([]Input{{Seq: 1}, {Seq: 2}}), func(b []byte) error { _, err := decodeCmds(b); return err }},
		{"snapshot", encodeSnapshot(Snapshot{Tick: 1, Actors: []ActorRec{sampleRec(1)}}), func(b []byte) error { _, err := decodeSnapshot(b); return err }},
	}
	for _, c := range cases {
		for i := range len(c.data) {
			if err := c.decode(c.data[:i]); err == nil {
				t.Errorf("%s: truncated to %d bytes decoded without error", c.name, i)
			}
		}
		bad := slices.Clone(c.data)
		bad[0] = 99
		if err := c.decode(bad); err == nil {
			t.Errorf("%s: unknown type decoded without error", c.name)
		}
	}
	badUpdown := encodeCmds([]Input{{Seq: 1}})
	badUpdown[2+5] = 7
	if _, err := decodeCmds(badUpdown); err == nil {
		t.Error("cmds with guard height 7 accepted")
	}
	if _, err := decodeCmds([]byte{MsgCmds, 0}); err == nil {
		t.Error("empty cmds accepted")
	}
}

func TestHelloVersions(t *testing.T) {
	v1 := encodeHello(Hello{Proto: 1, LevelHash: 5})[:11] // a version-1 client has no role byte
	h, err := decodeHello(v1)
	if err != nil || h.Proto != 1 || h.LevelHash != 5 || h.Role != RoleFighter {
		t.Errorf("version-1 hello: %+v %v", h, err)
	}
	if _, err := decodeHello(encodeHello(Hello{Proto: ProtocolVersion, Role: 7})); err == nil {
		t.Error("unknown role accepted")
	}
	if got := rejectText(RejectSpectatorsFull); got != "spectators full" {
		t.Errorf("reject text %q", got)
	}
	if b := encodeKeepalive(); len(b) != 1 || b[0] != MsgKeepalive {
		t.Errorf("keepalive %v", b)
	}
}

func TestSnapshotFitsPacket(t *testing.T) {
	var recs []ActorRec
	for i := range 16 {
		recs = append(recs, sampleRec(uint16(i+1)))
	}
	if n := len(encodeSnapshot(Snapshot{Tick: 1, Ack: 1, Actors: recs})); n > 1200 {
		t.Errorf("16-actor snapshot is %d bytes, want <= 1200", n)
	}
}

// The record size constant has to match what the encoder actually writes, or
// the decoder's length check silently rejects good packets.
func TestActorRecSize(t *testing.T) {
	const header = 1 + 4 + 4 + 1 // message id, tick, ack, count
	if n := len(encodeSnapshot(Snapshot{Tick: 1, Ack: 1, Actors: []ActorRec{sampleRec(1)}})); n != header+actorRecSize {
		t.Errorf("one-actor snapshot is %d bytes, want %d", n, header+actorRecSize)
	}
}

// Input held across a hit-stop freeze survives the wire. It has to: reconcile
// applies a record and then replays the command ring, so a buffer that did not
// cross would be dropped on every correction that lands inside a freeze.
func TestHeldInputReplicates(t *testing.T) {
	w := NewWorld(true)
	a := w.Spawn(ZeroVec2, true)
	for _, c := range []struct{ attack, jump bool }{{false, false}, {true, false}, {false, true}, {true, true}} {
		a.BufAttack, a.BufJump = c.attack, c.jump
		b := NewActor(ZeroVec2, true)
		recOf(a).apply(b, w)
		if b.BufAttack != c.attack || b.BufJump != c.jump {
			t.Errorf("held attack=%v jump=%v arrived as attack=%v jump=%v", c.attack, c.jump, b.BufAttack, b.BufJump)
		}
	}
}

// Structure survives the wire to a tenth of a hit-unit.
func TestStructureReplicates(t *testing.T) {
	w := NewWorld(true)
	a := w.Spawn(ZeroVec2, true)
	for _, want := range []float32{0, 0.4, 7.3, 11.6, MaxStructure} {
		a.Structure = want
		b := NewActor(ZeroVec2, true)
		recOf(a).apply(b, w)
		if diff := b.Structure - want; diff > 0.05 || diff < -0.05 {
			t.Errorf("structure %v arrived as %v", want, b.Structure)
		}
	}
}

// Energy survives the wire exactly, at every value it can hold.
func TestEnergyReplicates(t *testing.T) {
	w := NewWorld(true)
	a := w.Spawn(ZeroVec2, true)
	for want := int32(0); want <= MaxEnergy; want++ {
		a.Energy = want
		b := NewActor(ZeroVec2, true)
		recOf(a).apply(b, w)
		if b.Energy != want {
			t.Fatalf("energy %v arrived as %v", want, b.Energy)
		}
	}

	// a value past the cap is clamped rather than dropping the snapshot
	r := recOf(a)
	r.Energy = 250
	b := NewActor(ZeroVec2, true)
	r.apply(b, w)
	if b.Energy != MaxEnergy {
		t.Errorf("an over-cap record gave %v energy, want it clamped to %v", b.Energy, int32(MaxEnergy))
	}
}

func TestLevelHash(t *testing.T) {
	a := []byte(`{"0":{"1,2":{"resource":"x"}}}`)
	b := slices.Clone(a)
	if levelHash(a) != levelHash(b) {
		t.Error("same bytes, different hash")
	}
	b[10] = '3'
	if levelHash(a) == levelHash(b) {
		t.Error("one byte changed, same hash")
	}
}

func TestRecApplyRoundTrip(t *testing.T) {
	w := NewWorld(true)
	a := w.Spawn(ZeroVec2, true)
	b := w.Spawn(ZeroVec2, false)
	a.Pos.X, a.Pos.Y, a.Speed.X = 37, -12, -3.25
	a.Structure = 7.3
	a.Orientation = -1
	a.Attacker = b
	a.Ticks = 50
	a.SetAnim("resources/threefour/attack-high-2.png", AnimOptions{animDuration: SecondAttackSpeed, totalFrames: 6, once: true, startFrame: 3})
	a.Ticks = 55
	a.advanceAnim()
	r := recOf(a)
	c := NewActor(ZeroVec2, true)
	r.apply(c, w)
	if got := recOf(c); got != r {
		t.Errorf("apply/recOf mismatch:\n got %+v\nwant %+v", got, r)
	}
	if c.AnimCurrentFrame != a.AnimCurrentFrame || c.AnimDuration != a.AnimDuration {
		t.Errorf("animation differs: frame %d/%d duration %v/%v", c.AnimCurrentFrame, a.AnimCurrentFrame, c.AnimDuration, a.AnimDuration)
	}
}

// Every animation referenced by actor code must be in animTable, or recOf panics.
func TestAnimTableCoversActorAnims(t *testing.T) {
	src, err := os.ReadFile("actor.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range regexp.MustCompile(`SetAnim(?:Frame)?\("(resources/threefour/[^"%]+\.png)"`).FindAllSubmatch(src, -1) {
		if _, ok := animIDs[string(m[1])]; !ok {
			t.Errorf("%s missing from animTable", m[1])
		}
	}
	// animations built with Sprintf from two different guard heights (0..2)
	for i := range 3 {
		for j := range 3 {
			if i == j {
				continue
			}
			for _, f := range []string{"resources/threefour/updown-%d-%d.png", "resources/threefour/updown-atk-%d-%d.png"} {
				if p := fmt.Sprintf(f, i, j); !slices.Contains(animTable, p) {
					t.Errorf("%s missing from animTable", p)
				}
			}
		}
	}
	if len(animTable) > 255 {
		t.Error("animTable does not fit a byte")
	}
}
