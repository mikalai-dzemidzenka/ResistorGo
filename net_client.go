package resistor

import (
	"errors"
	"fmt"
	"github.com/codecat/go-enet"
	rl "github.com/gen2brain/raylib-go/raylib"
	"math"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"time"
)

const (
	ConnectTimeout = 5 * time.Second
	LostTimeout    = 5 * time.Second
	CmdRing        = 128 // > 2 s of commands
	CmdRedundancy  = 3   // newest + 2 previous in every packet
	SmoothMaxPx    = 8
	SmoothTau      = 0.035 // s; ~95 % of a correction is gone after 100 ms
	KeepaliveEvery = time.Second
)

type Client struct {
	Host        enet.Host
	Peer        enet.Peer
	Addr        string
	Start       time.Time
	LastRecv    time.Time
	Welcomed    bool
	PlayerID    uint16
	ActorID     uint16
	W           *World
	Seq         uint32
	Ring        [CmdRing]Input
	LastSnap    uint32
	LastAck     uint32
	HasSnap     bool
	Interp      Interp
	Corrections int
	CorrOffset  rl.Vector2
	Sim         NetSim
	snapTimes   []time.Time
	closed      bool
	Spectator   bool // joined without a fighter
	lastKeep    time.Time
	Feed        messageFeed // lines the server sent, for the HUD
}

var Cli *Client

// menu screen to return to if the connection fails before the match starts
var connectingFrom = menuJoin

func parseAddr(s string) (string, uint16, error) {
	s = strings.TrimSpace(s)
	host, portStr := s, ""
	if i := strings.LastIndexByte(s, ':'); i >= 0 {
		host, portStr = s[:i], s[i+1:]
	}
	if host == "" {
		return "", 0, errors.New("invalid address")
	}
	port := DefaultPort
	if portStr != "" {
		p, err := strconv.Atoi(portStr)
		if err != nil || p < 1 || p > 65535 {
			return "", 0, errors.New("invalid port")
		}
		port = p
	}
	ips, err := net.LookupIP(host)
	if err != nil {
		return "", 0, fmt.Errorf("cannot resolve %s", host)
	}
	for _, ip := range ips {
		if v4 := ip.To4(); v4 != nil {
			return v4.String(), uint16(port), nil
		}
	}
	return "", 0, fmt.Errorf("no IPv4 address for %s", host)
}

func startClient(addr string, spectator bool) error {
	ip, port, err := parseAddr(addr)
	if err != nil {
		return err
	}
	initNet()
	h, err := enet.NewHost(nil, 1, ChannelCount, 0, 0)
	if err != nil {
		return err
	}
	peer, err := h.Connect(enet.NewAddress(ip, port), ChannelCount, 0)
	if err != nil {
		h.Destroy()
		return err
	}
	now := time.Now()
	Cli = &Client{
		Host:      h,
		Peer:      peer,
		Addr:      fmt.Sprintf("%s:%d", ip, port),
		Start:     now,
		LastRecv:  now,
		W:         NewWorld(false),
		Spectator: spectator,
	}
	Cli.Interp.Delay = interpDelayTicks()
	CL = Cli.W
	NetMode = NetClient
	return nil
}

// startListen hosts a server in this process and joins it over loopback.
func startListen(cfg ServerConfig) error {
	s, err := NewServer(cfg)
	if err != nil {
		return err
	}
	if err := startClient(fmt.Sprintf("127.0.0.1:%d", cfg.Port), false); err != nil {
		s.Close()
		return err
	}
	Srv = s
	SV = s.W
	NetMode = NetListen
	return nil
}

func (c *Client) send(ch uint8, data []byte) {
	c.Sim.send(time.Now(), c.Peer, ch, data)
}

// ready reports whether the client is in the match: a fighter once its own
// fighter arrived, a spectator once the first snapshot arrived.
func (c *Client) ready() bool {
	return c.W.Local != nil || (c.Spectator && c.HasSnap)
}

func (c *Client) fail(reason string) {
	screen := menuMain
	if !c.ready() {
		screen = connectingFrom
	}
	leaveMatch(reason, screen)
}

// Frame polls the network, runs client ticks (once in the level) and updates remote actors.
func (c *Client) Frame(dt float32) {
	now := time.Now()
	c.poll(now)
	for _, p := range c.Sim.due(&c.Sim.in, now) {
		if c.closed {
			return
		}
		c.recv(p.data)
	}
	if c.closed {
		return
	}
	if !c.ready() {
		if now.Sub(c.Start) > ConnectTimeout {
			c.fail("connection timed out")
			return
		}
	} else if now.Sub(c.LastRecv) > LostTimeout {
		c.fail("connection lost")
		return
	}
	if c.Spectator && c.Welcomed && now.Sub(c.lastKeep) >= KeepaliveEvery {
		c.lastKeep = now
		c.send(ChanUnreliable, encodeKeepalive())
	}

	c.Interp.Advance(float64(dt))
	if Stage == StageLevel && c.W.Local != nil {
		Accumulator += float64(dt)
		for Accumulator > FixedTimeStep {
			c.tick()
			Accumulator -= FixedTimeStep
		}
	}
	c.applyInterp()
	for _, p := range c.Sim.due(&c.Sim.out, now) {
		sendBytes(c.Peer, p.ch, p.data)
	}
}

func (c *Client) poll(now time.Time) {
	for !c.closed {
		ev := c.Host.Service(0)
		switch ev.GetType() {
		case enet.EventNone:
			return
		case enet.EventConnect:
			role := RoleFighter
			if c.Spectator {
				role = RoleSpectator
			}
			c.send(ChanReliable, encodeHello(Hello{Proto: ProtocolVersion, LevelHash: LE.Hash(), Role: role}))
		case enet.EventDisconnect:
			if c.Welcomed {
				c.fail("connection lost")
			} else {
				c.fail("connection timed out")
			}
		case enet.EventReceive:
			pkt := ev.GetPacket()
			data := pkt.GetData()
			ch := ev.GetChannelID()
			pkt.Destroy()
			c.Sim.receive(now, ch, data, c.recv)
		}
	}
}

func (c *Client) recv(data []byte) {
	if len(data) == 0 || c.closed {
		return
	}
	c.LastRecv = time.Now()
	switch data[0] {
	case MsgWelcome:
		w, err := decodeWelcome(data)
		if err == nil && !c.Welcomed {
			c.Welcomed = true
			c.PlayerID, c.ActorID = w.PlayerID, w.ActorID
		}
	case MsgReject:
		if r, err := decodeReject(data); err == nil {
			c.fail(rejectText(r))
		}
	case MsgBye:
		if b, err := decodeBye(data); err == nil {
			c.fail(byeText(b))
		}
	case MsgSnapshot:
		if s, err := decodeSnapshot(data); err == nil && c.Welcomed {
			c.onSnapshot(s)
		}
	case MsgText:
		if text, err := decodeText(data); err == nil {
			c.Feed.add(text, time.Now())
		}
	}
}

func (c *Client) onSnapshot(s Snapshot) {
	if c.HasSnap && s.Tick <= c.LastSnap {
		return // old or duplicate
	}
	c.HasSnap = true
	c.LastSnap = s.Tick
	now := time.Now()
	c.snapTimes = append(c.snapTimes, now)
	c.LastAck = s.Ack
	c.Interp.Add(s)
	if c.Spectator && Stage != StageLevel {
		enterLevel()
	}
	for _, r := range s.Actors {
		if r.ID == c.ActorID {
			c.reconcile(r, s.Ack)
		}
	}
}

// reconcile resets the local fighter to the server state and replays unacknowledged commands.
func (c *Client) reconcile(r ActorRec, ack uint32) {
	local := c.W.Local
	if local == nil {
		local = NewActor(rl.Vector2{}, true)
		c.W.Add(local)
		c.W.Local = local
		r.apply(local, c.W)
		local.PrevPos = local.Pos
		if Stage != StageLevel {
			enterLevel()
		}
		return
	}
	before := recOf(local)
	beforePos := local.Pos
	events := local.Events

	r.apply(local, c.W)
	for seq := ack + 1; seq <= c.Seq; seq++ {
		in := c.Ring[seq%CmdRing]
		if in.Seq != seq {
			continue // overwritten: older than the ring
		}
		local.Input = in
		local.Update(float32(FixedTimeStep))
	}
	local.Events = events // effects of replayed ticks already played

	if recOf(local) != before {
		c.Corrections++
		d := rl.Vector2{X: beforePos.X - local.Pos.X, Y: beforePos.Y - local.Pos.Y}
		if abs(d.X) <= SmoothMaxPx && abs(d.Y) <= SmoothMaxPx {
			c.CorrOffset.X += d.X
			c.CorrOffset.Y += d.Y
		} else {
			c.CorrOffset = rl.Vector2{}
		}
	}
}

// tick samples input, sends it and predicts the local fighter one tick ahead.
func (c *Client) tick() {
	local := c.W.Local
	in := sampleInput(local)
	c.Seq++
	in.Seq = c.Seq
	rt := max(c.Interp.RenderTick(), 0)
	in.ViewTick = uint32(rt)
	in.ViewFrac = uint8((rt - math.Floor(rt)) * 256)
	c.Ring[c.Seq%CmdRing] = in

	cmds := make([]Input, 0, CmdRedundancy)
	for i := uint32(0); i < CmdRedundancy && i < c.Seq; i++ {
		if prev := c.Ring[(c.Seq-i)%CmdRing]; prev.Seq == c.Seq-i {
			cmds = append(cmds, prev)
		}
	}
	c.send(ChanUnreliable, encodeCmds(cmds))

	local.Input = in
	local.Update(float32(FixedTimeStep))
	c.W.Tick++
}

// applyInterp positions every remote actor at the interpolation time.
func (c *Client) applyInterp() {
	rt := c.Interp.RenderTick()
	visible := make(map[uint16]bool)
	for _, id := range c.Interp.IDs() {
		if id == c.ActorID {
			continue
		}
		st, ok := c.Interp.Sample(id, rt)
		if !ok {
			continue
		}
		a := c.W.ByID(id)
		if a == nil {
			a = NewActor(st.Pos, false)
			a.Remote = true
			c.W.Add(a)
		}
		st.Rec.apply(a, c.W)
		a.Pos = st.Pos
		a.PrevPos = st.Pos
		a.InterpolatedPos = st.Pos
		elapsed := max(st.Ticks-float64(a.AnimStartTick), 0) * FixedTimeStep
		a.AnimCurrentFrame, a.AnimFinished = animFrame(a.AnimOnce, a.AnimReverse, a.AnimFrame, a.AnimTotalFrames, a.AnimDuration, elapsed)
		visible[id] = true
	}
	for i := len(c.W.Actors) - 1; i >= 0; i-- {
		if a := c.W.Actors[i]; a.Remote && !visible[a.ID] {
			c.W.Remove(a)
		}
	}
}

// smoothCorrection adds the decaying reconciliation offset to the drawn position only.
func (c *Client) smoothCorrection(dt float32, local *Actor) {
	local.InterpolatedPos.X += c.CorrOffset.X
	local.InterpolatedPos.Y += c.CorrOffset.Y
	k := float32(math.Exp(-float64(dt) / SmoothTau))
	c.CorrOffset.X *= k
	c.CorrOffset.Y *= k
	if abs(c.CorrOffset.X) < 0.25 && abs(c.CorrOffset.Y) < 0.25 {
		c.CorrOffset = rl.Vector2{}
	}
}

func (c *Client) snapRate() int {
	cut := time.Now().Add(-time.Second)
	i := 0
	for i < len(c.snapTimes) && c.snapTimes[i].Before(cut) {
		i++
	}
	c.snapTimes = c.snapTimes[i:]
	return len(c.snapTimes)
}

// Close says goodbye to the server and destroys the host.
func (c *Client) Close() {
	if c.closed {
		return
	}
	c.closed = true
	for _, p := range c.Sim.out {
		sendBytes(c.Peer, p.ch, p.data)
	}
	sendBytes(c.Peer, ChanReliable, encodeBye(ByeQuit))
	c.Peer.DisconnectLater(0)
	serviceFor(c.Host, 100*time.Millisecond, 1)
	c.Host.Destroy()
}

// NetSim adds artificial latency and loss to the client's traffic (debug).
type NetSim struct {
	LatencyMs int // round trip; each direction is delayed by half
	LossPct   int // applied to the unreliable channel only
	in, out   []simPacket
}

type simPacket struct {
	at   time.Time
	ch   uint8
	data []byte
}

func (s *NetSim) lost(ch uint8) bool {
	return ch == ChanUnreliable && s.LossPct > 0 && rand.Intn(100) < s.LossPct
}

func (s *NetSim) delay() time.Duration {
	return time.Duration(s.LatencyMs) * time.Millisecond / 2
}

func (s *NetSim) send(now time.Time, peer enet.Peer, ch uint8, data []byte) {
	if s.lost(ch) {
		return
	}
	if s.LatencyMs == 0 && len(s.out) == 0 {
		sendBytes(peer, ch, data)
		return
	}
	s.out = append(s.out, simPacket{at: now.Add(s.delay()), ch: ch, data: data})
}

func (s *NetSim) receive(now time.Time, ch uint8, data []byte, handle func([]byte)) {
	if s.lost(ch) {
		return
	}
	if s.LatencyMs == 0 && len(s.in) == 0 {
		handle(data)
		return
	}
	s.in = append(s.in, simPacket{at: now.Add(s.delay()), ch: ch, data: data})
}

// due pops the packets from q whose delay has elapsed.
func (s *NetSim) due(q *[]simPacket, now time.Time) []simPacket {
	n := 0
	for n < len(*q) && !(*q)[n].at.After(now) {
		n++
	}
	out := (*q)[:n:n]
	*q = (*q)[n:]
	return out
}

func (s *NetSim) controls() {
	if rl.IsKeyPressed(rl.KeyF5) {
		s.LatencyMs = max(s.LatencyMs-50, 0)
	}
	if rl.IsKeyPressed(rl.KeyF6) {
		s.LatencyMs = min(s.LatencyMs+50, 500)
	}
	if rl.IsKeyPressed(rl.KeyF7) {
		s.LossPct = max(s.LossPct-5, 0)
	}
	if rl.IsKeyPressed(rl.KeyF8) {
		s.LossPct = min(s.LossPct+5, 50)
	}
}
