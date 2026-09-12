package resistor

import (
	"fmt"
	"github.com/codecat/go-enet"
	"testing"
	"time"
)

// rawPeer is a bare ENet client used to exercise the server's handshake.
type rawPeer struct {
	h    enet.Host
	p    enet.Peer
	got  [][]byte
	conn bool
}

func dialRaw(t *testing.T, port uint16) *rawPeer {
	t.Helper()
	h, err := enet.NewHost(nil, 1, ChannelCount, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	p, err := h.Connect(enet.NewAddress("127.0.0.1", port), ChannelCount, 0)
	if err != nil {
		t.Fatal(err)
	}
	return &rawPeer{h: h, p: p}
}

func (r *rawPeer) pump() {
	for {
		ev := r.h.Service(0)
		switch ev.GetType() {
		case enet.EventNone:
			return
		case enet.EventConnect:
			r.conn = true
		case enet.EventReceive:
			pkt := ev.GetPacket()
			r.got = append(r.got, pkt.GetData())
			pkt.Destroy()
		}
	}
}

func (r *rawPeer) msg(typ byte) []byte {
	for _, b := range r.got {
		if len(b) > 0 && b[0] == typ {
			return b
		}
	}
	return nil
}

func pumpFor(d time.Duration, s *Server, c *Client, raws ...*rawPeer) {
	end := time.Now().Add(d)
	last := time.Now()
	for time.Now().Before(end) {
		now := time.Now()
		dt := float32(now.Sub(last).Seconds())
		last = now
		if s != nil {
			s.Frame(dt)
		}
		if c != nil && !c.closed {
			c.Frame(dt)
		}
		for _, r := range raws {
			r.pump()
		}
		time.Sleep(5 * time.Millisecond)
	}
}

// hello connects a raw peer and sends a handshake, then waits for the answer.
func hello(t *testing.T, s *Server, c *Client, port uint16, h Hello) *rawPeer {
	t.Helper()
	r := dialRaw(t, port)
	for i := 0; i < 100 && !r.conn; i++ {
		pumpFor(10*time.Millisecond, s, c, r)
	}
	if !r.conn {
		t.Fatal("raw peer did not connect")
	}
	sendBytes(r.p, ChanReliable, encodeHello(h))
	pumpFor(300*time.Millisecond, s, c, r)
	return r
}

func expectReject(t *testing.T, r *rawPeer, want uint8) {
	t.Helper()
	b := r.msg(MsgReject)
	if b == nil {
		t.Errorf("no reject, want %q", rejectText(want))
		return
	}
	if got, _ := decodeReject(b); got != want {
		t.Errorf("reject %q, want %q", rejectText(got), rejectText(want))
	}
}

func TestNetIntegration(t *testing.T) {
	enet.Initialize()
	defer enet.Deinitialize()
	Stage = StageLevel // reconcile must not enter the level (needs a window)
	defer func() { Stage = StageMainMenu }()
	LE = NewLevelEditor(39, 83)

	cfg := DefaultServerConfig()
	cfg.Port = 27191
	cfg.MaxPlayers = 2
	cfg.Bots = 1
	s, err := NewServer(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if len(s.Bots) != 1 {
		t.Fatalf("server has %d bots, want 1", len(s.Bots))
	}
	if _, err := NewServer(cfg); err == nil {
		t.Error("second server bound the same port")
	}

	if err := startClient(fmt.Sprintf("127.0.0.1:%d", cfg.Port)); err != nil {
		t.Fatal(err)
	}
	c := Cli
	pumpFor(1500*time.Millisecond, s, c)
	if !c.Welcomed || c.W.Local == nil {
		t.Fatalf("client not in the match: welcomed=%t local=%v", c.Welcomed, c.W.Local)
	}
	if r := c.snapRate(); r < 18 || r > 22 {
		t.Errorf("snapshot rate %d/s, want 20±2", r)
	}
	if n := len(c.W.Actors); n != 2 {
		t.Errorf("client sees %d actors, want own fighter + 1 bot", n)
	}
	for _, p := range s.Peers {
		if p.Actor == nil {
			continue
		}
		// the client reports the tick it renders others at: server tick minus interp delay and latency
		if behind := int64(s.W.Tick) - int64(p.Actor.Input.ViewTick); behind < 3 || behind > 14 {
			t.Errorf("view tick %d is %d ticks behind server tick %d, want about 6", p.Actor.Input.ViewTick, behind, s.W.Tick)
		}
	}
	t.Logf("client seq %d, corrections %d, render tick %.1f, server tick %d", c.Seq, c.Corrections, c.Interp.RenderTick(), s.W.Tick)

	hash := LE.Hash()
	v := hello(t, s, c, cfg.Port, Hello{Proto: ProtocolVersion + 1, LevelHash: hash})
	expectReject(t, v, RejectVersion)
	l := hello(t, s, c, cfg.Port, Hello{Proto: ProtocolVersion, LevelHash: hash + 1})
	expectReject(t, l, RejectLevel)

	h2 := hello(t, s, c, cfg.Port, Hello{Proto: ProtocolVersion, LevelHash: hash})
	if h2.msg(MsgWelcome) == nil {
		t.Fatal("second human not welcomed")
	}
	if s.humans() != 2 || len(s.Bots) != 0 {
		t.Errorf("after second human: humans %d bots %d, want 2 0", s.humans(), len(s.Bots))
	}
	full := hello(t, s, c, cfg.Port, Hello{Proto: ProtocolVersion, LevelHash: hash})
	expectReject(t, full, RejectFull)

	h2.p.Disconnect(0)
	pumpFor(700*time.Millisecond, s, c, h2)
	if s.humans() != 1 || len(s.Bots) != 1 {
		t.Errorf("after second human left: humans %d bots %d, want 1 1", s.humans(), len(s.Bots))
	}
	if n := len(c.W.Actors); n != 2 {
		t.Errorf("client sees %d actors after leave/refill, want 2", n)
	}

	for _, r := range []*rawPeer{v, l, h2, full} {
		r.h.Destroy()
	}

	Srv = nil // leaveMatch must not close s a second time
	s.Close()
	pumpFor(300*time.Millisecond, nil, c)
	if menuError != "server closed" || Cli != nil {
		t.Errorf("after server close: menuError %q, client %v", menuError, Cli)
	}
}

func TestNetSimulatedLag(t *testing.T) {
	enet.Initialize()
	defer enet.Deinitialize()
	Stage = StageLevel
	defer func() { Stage = StageMainMenu }()
	LE = NewLevelEditor(39, 83)

	cfg := DefaultServerConfig()
	cfg.Port = 27193
	s, err := NewServer(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if err := startClient(fmt.Sprintf("127.0.0.1:%d", cfg.Port)); err != nil {
		t.Fatal(err)
	}
	c := Cli
	c.Sim.LatencyMs = 200
	c.Sim.LossPct = 5
	pumpFor(2500*time.Millisecond, s, c)

	if !c.Welcomed || c.W.Local == nil {
		t.Fatal("client did not join under simulated lag")
	}
	// predicted every tick regardless of the server: ~12+ commands are in flight
	if unacked := c.Seq - c.LastAck; unacked < 10 {
		t.Errorf("only %d unacknowledged commands; simulated latency not applied", unacked)
	}
	if r := c.snapRate(); r < 14 {
		t.Errorf("snapshot rate %d/s under 5%% loss", r)
	}
	t.Logf("seq %d, unacked %d, corrections %d, snapshots/s %d", c.Seq, c.Seq-c.LastAck, c.Corrections, c.snapRate())

	// server process dies without saying goodbye
	s.Host.Destroy()
	start := time.Now()
	for !c.closed && time.Since(start) < 12*time.Second {
		pumpFor(50*time.Millisecond, nil, c)
	}
	if menuError != "connection lost" || time.Since(start) > 10*time.Second {
		t.Errorf("after server death: %q after %v, want \"connection lost\" within 10 s", menuError, time.Since(start))
	}
	t.Logf("connection lost detected after %v", time.Since(start).Round(100*time.Millisecond))
}

func TestClientConnectTimeout(t *testing.T) {
	if testing.Short() {
		t.Skip("waits for the 5 s connect timeout")
	}
	enet.Initialize()
	defer enet.Deinitialize()
	LE = NewLevelEditor(39, 83)
	if err := startClient("127.0.0.1:27192"); err != nil { // nothing listens there
		t.Fatal(err)
	}
	connectingFrom = menuJoin
	start := time.Now()
	c := Cli
	for !c.closed && time.Since(start) < ConnectTimeout+2*time.Second {
		pumpFor(50*time.Millisecond, nil, c)
	}
	if menuError != "connection timed out" || menuScreen != menuJoin {
		t.Errorf("menuError %q screen %d, want timeout on the join screen", menuError, menuScreen)
	}
	t.Logf("gave up after %v", time.Since(start).Round(100*time.Millisecond))
}

func TestParseAddr(t *testing.T) {
	if ip, port, err := parseAddr("127.0.0.1"); err != nil || ip != "127.0.0.1" || port != DefaultPort {
		t.Errorf("default port: %s %d %v", ip, port, err)
	}
	if _, port, err := parseAddr(" localhost:4000 "); err != nil || port != 4000 {
		t.Errorf("host:port: %d %v", port, err)
	}
	for _, bad := range []string{"", ":27015", "127.0.0.1:0", "127.0.0.1:99999", "127.0.0.1:abc", "no-such-host.invalid"} {
		if _, _, err := parseAddr(bad); err == nil {
			t.Errorf("%q accepted", bad)
		}
	}
}
