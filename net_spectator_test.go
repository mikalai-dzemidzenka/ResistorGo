package resistor

import (
	"fmt"
	"github.com/codecat/go-enet"
	"testing"
	"time"
)

func snapshotsIn(r *rawPeer) int {
	n := 0
	for _, b := range r.got {
		if len(b) > 0 && b[0] == MsgSnapshot {
			n++
		}
	}
	return n
}

func fighterOf(s *Server, id uint16) *Actor {
	return s.W.ByID(id)
}

// Server-side spectator rules, driven by raw ENet peers.
func TestSpectatorServer(t *testing.T) {
	enet.Initialize()
	defer enet.Deinitialize()
	LE = NewLevelEditor(39, 83)

	cfg := DefaultServerConfig()
	cfg.Port = 27194
	cfg.MaxPlayers = 2
	cfg.MaxSpectators = 1
	cfg.Bots = 1
	s, err := NewServer(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	hash := LE.Hash()
	fighter := Hello{Proto: ProtocolVersion, LevelHash: hash, Role: RoleFighter}
	spectator := Hello{Proto: ProtocolVersion, LevelHash: hash, Role: RoleSpectator}

	f1 := hello(t, s, nil, cfg.Port, fighter)
	w1, err := decodeWelcome(f1.msg(MsgWelcome))
	if err != nil {
		t.Fatal("first fighter not welcomed")
	}
	if s.humans() != 1 || len(s.Bots) != 1 {
		t.Fatalf("humans %d bots %d, want 1 1 (all slots taken)", s.humans(), len(s.Bots))
	}

	// all player slots taken by a human and a bot: a spectator still gets in and no bot is removed
	sp := hello(t, s, nil, cfg.Port, spectator)
	ws, err := decodeWelcome(sp.msg(MsgWelcome))
	if err != nil || ws.ActorID != 0 {
		t.Fatalf("spectator welcome %+v %v, want actor 0", ws, err)
	}
	if len(s.Bots) != 1 || s.humans() != 1 || s.spectators() != 1 || len(s.W.Actors) != 2 {
		t.Errorf("after spectator: bots %d humans %d spectators %d actors %d, want 1 1 1 2", len(s.Bots), s.humans(), s.spectators(), len(s.W.Actors))
	}
	pumpFor(300*time.Millisecond, s, nil, sp, f1)
	if n := snapshotsIn(sp); n < 3 {
		t.Errorf("spectator received %d snapshots in ~0.6 s", n)
	}
	if snap, err := decodeSnapshot(lastMsg(sp, MsgSnapshot)); err != nil || len(snap.Actors) != 2 || snap.Ack != 0 {
		t.Errorf("spectator snapshot: %d actors ack %d err %v, want 2 actors ack 0", len(snap.Actors), snap.Ack, err)
	}

	sp2 := hello(t, s, nil, cfg.Port, spectator)
	expectReject(t, sp2, RejectSpectatorsFull)

	// forged commands from the spectator never move a fighter
	human := fighterOf(s, w1.ActorID)
	before := recOf(human)
	var cmds []Input
	for seq := uint32(1); seq <= 3; seq++ {
		cmds = append(cmds, Input{Seq: seq, Left: true, Attack: true})
	}
	for range 10 {
		sendBytes(sp.p, ChanUnreliable, encodeCmds(cmds))
		pumpFor(20*time.Millisecond, s, nil, sp, f1)
	}
	if recOf(human) != before {
		t.Error("spectator commands changed the human fighter")
	}
	for _, p := range s.Peers {
		if p.Role == RoleSpectator && (len(p.Cmds.q) != 0 || p.Cmds.last != 0) {
			t.Errorf("spectator commands were queued: %d queued, last %d", len(p.Cmds.q), p.Cmds.last)
		}
	}

	// the spectator leaving changes nothing for fighters
	sp.p.Disconnect(0)
	pumpFor(500*time.Millisecond, s, nil, sp, f1)
	if s.spectators() != 0 || s.humans() != 1 || len(s.Bots) != 1 {
		t.Errorf("after spectator left: spectators %d humans %d bots %d, want 0 1 1", s.spectators(), s.humans(), len(s.Bots))
	}

	// fill the human slots: a fighter is then rejected, but the match can still be watched
	f2 := hello(t, s, nil, cfg.Port, fighter)
	if f2.msg(MsgWelcome) == nil || s.humans() != 2 {
		t.Fatalf("second fighter not accepted (humans %d)", s.humans())
	}
	f3 := hello(t, s, nil, cfg.Port, fighter)
	expectReject(t, f3, RejectFull)
	sp3 := hello(t, s, nil, cfg.Port, spectator)
	if sp3.msg(MsgWelcome) == nil {
		t.Error("spectator rejected from a server full of fighters")
	}

	// a version-1 client (no role byte) is told "version mismatch"
	old := dialRaw(t, cfg.Port)
	for i := 0; i < 100 && !old.conn; i++ {
		pumpFor(10*time.Millisecond, s, nil, old)
	}
	sendBytes(old.p, ChanReliable, encodeHello(Hello{Proto: 1, LevelHash: hash})[:11])
	pumpFor(300*time.Millisecond, s, nil, old)
	expectReject(t, old, RejectVersion)

	for _, r := range []*rawPeer{f1, sp, sp2, f2, f3, sp3, old} {
		r.h.Destroy()
	}
}

func lastMsg(r *rawPeer, typ byte) []byte {
	for i := len(r.got) - 1; i >= 0; i-- {
		if len(r.got[i]) > 0 && r.got[i][0] == typ {
			return r.got[i]
		}
	}
	return nil
}

// A real spectator Client: enters without a fighter, sees everyone, sends no commands, stays connected.
func TestSpectatorClient(t *testing.T) {
	enet.Initialize()
	defer enet.Deinitialize()
	Stage = StageLevel // entering the level needs a window
	defer func() { Stage = StageMainMenu }()
	LE = NewLevelEditor(39, 83)

	cfg := DefaultServerConfig()
	cfg.Port = 27195
	cfg.MaxPlayers = 3
	cfg.Bots = 2
	s, err := NewServer(cfg)
	if err != nil {
		t.Fatal(err)
	}
	f := hello(t, s, nil, cfg.Port, Hello{Proto: ProtocolVersion, LevelHash: LE.Hash()})
	if f.msg(MsgWelcome) == nil || s.humans() != 1 || len(s.Bots) != 2 {
		t.Fatalf("setup: humans %d bots %d", s.humans(), len(s.Bots))
	}

	if err := startClient(fmt.Sprintf("127.0.0.1:%d", cfg.Port), true); err != nil {
		t.Fatal(err)
	}
	c := Cli
	pumpFor(1500*time.Millisecond, s, c, f)
	if !c.ready() || c.W.Local != nil || c.ActorID != 0 {
		t.Fatalf("spectator: ready %t local %v actor %d", c.ready(), c.W.Local, c.ActorID)
	}
	if n := len(c.W.Actors); n != 3 {
		t.Errorf("spectator sees %d fighters, want 3 (1 human + 2 bots)", n)
	}
	if c.Seq != 0 {
		t.Errorf("spectator sent %d commands", c.Seq)
	}
	if s.spectators() != 1 || len(s.W.Actors) != 3 {
		t.Errorf("server: spectators %d actors %d, want 1 3", s.spectators(), len(s.W.Actors))
	}

	// idle for 3x the (shortened) server timeout: keepalives keep it connected
	old := PeerTimeout
	PeerTimeout = 1500 * time.Millisecond
	defer func() { PeerTimeout = old }()
	pumpFor(3*PeerTimeout, s, c)
	if c.closed || s.spectators() != 1 {
		t.Errorf("idle spectator dropped: client closed %t, server spectators %d", c.closed, s.spectators())
	}
	if c.snapRate() < 18 {
		t.Errorf("idle spectator receives %d snapshots/s", c.snapRate())
	}

	f.h.Destroy()
	c.Close()
	Cli = nil
	s.Close()
}
