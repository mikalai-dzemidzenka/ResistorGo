package resistor

import (
	"fmt"
	rl "github.com/gen2brain/raylib-go/raylib"
	"testing"
	"time"
)

// respawnClient is a welcomed client holding fighter sa, as after its first snapshot.
func respawnClient(t *testing.T, sa *Actor) *Client {
	t.Helper()
	floor()
	Stage = StageLevel // reconcile must not enter the level (needs a window)
	t.Cleanup(func() { Stage = StageMainMenu })
	c := &Client{W: NewWorld(false), Welcomed: true, ActorID: sa.ID}
	c.onSnapshot(Snapshot{Tick: 3, Yours: sa.ID, Actors: []ActorRec{recOf(sa)}})
	c.applyInterp()
	if c.W.Local == nil || c.W.Local.ID != sa.ID {
		t.Fatalf("client does not hold fighter %d: %v", sa.ID, c.W.Local)
	}
	return c
}

// Each player is named its own fighter in the snapshots it is sent, none while
// it has none, and the new one after a respawn.
func TestSnapshotNamesOwnFighter(t *testing.T) {
	initNet()
	defer shutdownNet()
	floor()

	cfg := DefaultServerConfig()
	cfg.Port = 27197
	cfg.MaxPlayers = 2
	cfg.Bots = 0
	s, err := NewServer(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	h := Hello{Proto: ProtocolVersion, LevelHash: LE.Hash(), Role: RoleFighter}
	f1 := hello(t, s, nil, cfg.Port, h)
	f2 := hello(t, s, nil, cfg.Port, h)
	w1, err1 := decodeWelcome(f1.msg(MsgWelcome))
	w2, err2 := decodeWelcome(f2.msg(MsgWelcome))
	if err1 != nil || err2 != nil {
		t.Fatal("both fighters not welcomed")
	}
	yours := func(r *rawPeer) uint16 {
		t.Helper()
		snap, err := decodeSnapshot(lastMsg(r, MsgSnapshot))
		if err != nil {
			t.Fatal(err)
		}
		return snap.Yours
	}
	pumpFor(200*time.Millisecond, s, nil, f1, f2)
	if y1, y2 := yours(f1), yours(f2); y1 != w1.ActorID || y2 != w2.ActorID {
		t.Fatalf("named %d and %d, want %d and %d", y1, y2, w1.ActorID, w2.ActorID)
	}

	s.W.Kill(fighterOf(s, w1.ActorID))
	pumpFor(300*time.Millisecond, s, nil, f1, f2)
	if y1, y2 := yours(f1), yours(f2); y1 != 0 || y2 != w2.ActorID {
		t.Fatalf("while the first is dead: named %d and %d, want 0 and %d", y1, y2, w2.ActorID)
	}

	pumpFor(2200*time.Millisecond, s, nil, f1, f2)
	var p1 *netPlayer
	for _, p := range s.Peers {
		if p.ID == w1.PlayerID {
			p1 = p
		}
	}
	if p1 == nil || p1.Actor == nil || p1.Actor.ID == w1.ActorID {
		t.Fatal("the first player was not given a new fighter")
	}
	if y1, y2 := yours(f1), yours(f2); y1 != p1.Actor.ID || y2 != w2.ActorID {
		t.Errorf("after the respawn: named %d and %d, want %d and %d", y1, y2, p1.Actor.ID, w2.ActorID)
	}
	f1.h.Destroy()
	f2.h.Destroy()
}

// Named another fighter, a client drops the corpse it held and builds the new
// one fresh, with none of the old one's death.
func TestClientAdoptsNamedFighter(t *testing.T) {
	sv := NewWorld(true)
	old := sv.Spawn(rl.Vector2{}, true)
	c := respawnClient(t, old)
	held := c.W.Local
	held.State, held.DeathTicks, held.Dying = StateDeath, 20, true
	held.Attacker = held

	fresh := sv.Spawn(rl.Vector2{X: 40}, true)
	c.onSnapshot(Snapshot{Tick: 6, Yours: fresh.ID, Actors: []ActorRec{recOf(fresh)}})
	c.applyInterp()
	local := c.W.Local
	if c.ActorID != fresh.ID || local == nil || local == held || local.ID != fresh.ID {
		t.Fatalf("client reconciles against %d (local %v), want a new fighter %d", c.ActorID, local, fresh.ID)
	}
	if local.DeathTicks != 0 || local.Dying || local.Attacker != nil {
		t.Errorf("the new fighter carries death state: ticks %d dying %t attacker %v", local.DeathTicks, local.Dying, local.Attacker)
	}
	if c.W.ByID(old.ID) != nil {
		t.Errorf("the old fighter is still in the client's world")
	}
}

// Named no fighter, a client holds none until it is named one, and nothing it
// runs each frame trips over that.
func TestClientNamedNone(t *testing.T) {
	sv := NewWorld(true)
	sa := sv.Spawn(rl.Vector2{}, true)
	c := respawnClient(t, sa)

	for tick := uint32(6); tick <= 60; tick += 3 {
		c.onSnapshot(Snapshot{Tick: tick})
		c.Interp.Advance(3 * FixedTimeStep)
		c.applyInterp()
		presentEffects(c.W)
		if c.W.Local != nil || c.ActorID != 0 {
			t.Fatalf("tick %d: client holds %v (actor %d) while named none", tick, c.W.Local, c.ActorID)
		}
		if !c.ready() {
			t.Fatalf("tick %d: a client waiting for a respawn is not ready, so it would time out", tick)
		}
	}
	if len(c.W.Actors) != 0 {
		t.Errorf("client world has %d actors, want none", len(c.W.Actors))
	}

	next := sv.Spawn(rl.Vector2{X: 40}, true)
	c.onSnapshot(Snapshot{Tick: 63, Yours: next.ID, Actors: []ActorRec{recOf(next)}})
	c.applyInterp()
	if c.W.Local == nil || c.W.Local.ID != next.ID {
		t.Errorf("client holds %v, want the fighter it was named, %d", c.W.Local, next.ID)
	}
}

// Named the fighter it already holds, a client predicts it on untouched: the
// same object, never corrected.
func TestClientKeepsSameFighter(t *testing.T) {
	sv := NewWorld(true)
	sa := sv.Spawn(rl.Vector2{}, true)
	c := respawnClient(t, sa)
	local := c.W.Local
	dt := float32(FixedTimeStep)

	for tick := 1; tick <= 240; tick++ {
		in := Input{Left: tick%90 < 40, Right: tick%90 > 60, Attack: tick%45 == 7, Jump: tick == 120, Updown: uint8(tick/30) % 3}
		c.Seq++
		in.Seq = c.Seq
		c.Ring[in.Seq%CmdRing] = in
		local.Input = in
		local.Update(dt)
		sa.Input = in
		sa.Update(dt)
		if tick%SnapshotEvery == 0 {
			c.onSnapshot(Snapshot{Tick: uint32(3 + tick), Ack: c.Seq, Yours: sa.ID, Actors: []ActorRec{recOf(sa)}})
			c.applyInterp()
			if c.W.Local != local {
				t.Fatalf("tick %d: the fighter was rebuilt", tick)
			}
		}
	}
	if c.Corrections != 0 {
		t.Errorf("%d corrections in a match with no lag", c.Corrections)
	}
}

// A client's own fighter lives exactly as long as the server sends it.
func TestClientOwnFighterLifetime(t *testing.T) {
	sv := NewWorld(true)
	sa := sv.Spawn(rl.Vector2{}, true)
	c := respawnClient(t, sa)
	local := c.W.Local

	for tick := uint32(6); tick <= 60; tick += 3 {
		c.onSnapshot(Snapshot{Tick: tick, Yours: sa.ID, Actors: []ActorRec{recOf(sa)}})
		c.Interp.Advance(3 * FixedTimeStep)
		c.applyInterp()
		if c.W.Local != local {
			t.Fatalf("tick %d: own fighter removed while the server still sends it", tick)
		}
	}
	// the server stops sending it, even before it stops naming it
	c.onSnapshot(Snapshot{Tick: 63, Yours: sa.ID})
	c.applyInterp()
	if c.W.Local != nil || c.W.ByID(sa.ID) != nil {
		t.Errorf("own fighter still in the client's world after the server stopped sending it")
	}
}

// Past its flicker a dead fighter a client holds stops counting and stays hidden.
func TestDeathFlickerBounded(t *testing.T) {
	w := NewWorld(false)
	a := w.Spawn(rl.Vector2{}, true)
	a.State, a.AnimFinished = StateDeath, true
	for i := range 600 {
		a.updateDeath()
		if int(a.DeathTicks) >= DeathFlickerTicks && !a.deathHidden() {
			t.Fatalf("update %d: shown past the end of the flicker", i)
		}
	}
	if a.DeathTicks != DeathFlickerTicks {
		t.Errorf("DeathTicks %d after 600 updates, want it held at %d", a.DeathTicks, DeathFlickerTicks)
	}

	// the authority still kills it on the tick the flicker ends
	sw := NewWorld(true)
	b := sw.Spawn(rl.Vector2{}, true)
	b.State, b.AnimFinished = StateDeath, true
	for range DeathFlickerTicks {
		b.updateDeath()
	}
	if !b.Dying {
		t.Errorf("the authority did not kill the fighter when its flicker ended")
	}
}

// The reported bug end to end: a client's fighter dies, the corpse leaves the
// client's world, it holds nothing for the gap, then predicts and controls the
// fighter it respawns with - never blinking past its flicker on the way.
func TestNetDeathHandsOverControl(t *testing.T) {
	initNet()
	defer shutdownNet()
	Stage = StageLevel // reconcile must not enter the level (needs a window)
	defer func() { Stage = StageMainMenu }()
	floor()

	cfg := DefaultServerConfig()
	cfg.Port = 27198
	cfg.MaxPlayers = 1
	cfg.Bots = 0
	s, err := NewServer(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { Srv = nil; s.Close() }()
	if err := startClient(fmt.Sprintf("127.0.0.1:%d", cfg.Port), false); err != nil {
		t.Fatal(err)
	}
	c := Cli
	pumpFor(1200*time.Millisecond, s, c)
	if c.W.Local == nil {
		t.Fatal("client not in the match")
	}
	var p *netPlayer
	for _, q := range s.Peers {
		p = q
	}
	dead := p.Actor
	killer := s.W.Spawn(rl.Vector2{X: 300}, false)
	dead.Structure = HitStructure
	dead.takeHit(killer)

	var reaped, gone, gap, back time.Time
	end := time.Now().Add(4 * time.Second)
	last := time.Now()
	for time.Now().Before(end) && back.IsZero() {
		now := time.Now()
		dt := float32(now.Sub(last).Seconds())
		last = now
		s.Frame(dt)
		c.Frame(dt)
		if c.closed {
			t.Fatal("the client left the match while dead")
		}
		if reaped.IsZero() && s.W.ByID(dead.ID) == nil {
			reaped = now
		}
		if a := c.W.ByID(dead.ID); a != nil && a.DeathTicks > DeathFlickerTicks {
			t.Fatalf("the corpse's flicker ran to %d", a.DeathTicks)
		}
		if !reaped.IsZero() && gone.IsZero() && c.W.ByID(dead.ID) == nil {
			gone = now
		}
		local := c.W.Local
		switch {
		case !gone.IsZero() && p.Actor == nil && local != nil:
			t.Fatalf("the client holds fighter %d during the gap", local.ID)
		case !gone.IsZero() && p.Actor == nil:
			gap = now
		case p.Actor != nil && p.Actor != dead && local != nil && local.ID == p.Actor.ID:
			back = now
		}
		time.Sleep(5 * time.Millisecond)
	}
	if gone.IsZero() {
		t.Fatal("the corpse never left the client's world")
	}
	if d := gone.Sub(reaped); d > 2*SnapshotEvery*time.Second/60+50*time.Millisecond {
		t.Errorf("the corpse left the client's world %v after the server reaped it", d)
	}
	if gap.IsZero() {
		t.Error("the client never had a gap without a fighter")
	}
	if back.IsZero() {
		t.Fatal("control never returned to the client after the respawn")
	}

	// and it is under the client's control: the client predicts it, and the
	// server keeps executing the client's commands for it
	seq, acked := c.Seq, p.Cmds.last
	pumpFor(400*time.Millisecond, s, c)
	if c.ActorID != p.Actor.ID || c.W.Local == nil || c.W.Local.ID != p.Actor.ID {
		t.Errorf("client acts for %d (local %v), server holds %d", c.ActorID, c.W.Local, p.Actor.ID)
	}
	if c.Seq <= seq || p.Cmds.last <= acked {
		t.Errorf("commands stalled after the respawn: client seq %d -> %d, server executed %d -> %d", seq, c.Seq, acked, p.Cmds.last)
	}
}
