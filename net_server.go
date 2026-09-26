package resistor

import (
	"cmp"
	"fmt"
	"github.com/codecat/go-enet"
	rl "github.com/gen2brain/raylib-go/raylib"
	"os"
	"os/signal"
	"slices"
	"strings"
	"time"
)

const (
	CmdBurst      = 8  // token bucket size: commands a player may run in one tick after a stall
	MaxQueuedCmds = 16 // commands buffered per player; more are dropped
)

// PeerTimeout drops a peer the server has not heard from (a var so tests can shorten it).
var PeerTimeout = 10 * time.Second

type ServerConfig struct {
	Port          uint16
	MaxPlayers    int // human players
	MaxSpectators int
	Bots          int
	LagComp       bool
	MaxUnlag      uint32 // ticks
	// The scripts the server runs, read from ScriptDir; none by default.
	Gamemode      string
	Filterscripts []string
	ScriptDir     string
}

// DefaultServerConfig is the configuration a server starts with, the scripts
// named on the command line included.
func DefaultServerConfig() ServerConfig {
	return ServerConfig{Port: DefaultPort, MaxPlayers: 8, MaxSpectators: 4, LagComp: true, MaxUnlag: DefaultMaxUnlag,
		Gamemode: Gamemode, Filterscripts: Filterscripts, ScriptDir: ScriptDir}
}

// cmdQueue orders a player's commands and rate-limits their execution.
type cmdQueue struct {
	q      []Input
	last   uint32 // last executed sequence
	tokens int
}

func (cq *cmdQueue) push(in Input) {
	if in.Seq <= cq.last {
		return
	}
	i, found := slices.BinarySearchFunc(cq.q, in.Seq, func(q Input, s uint32) int { return cmp.Compare(q.Seq, s) })
	if found || len(cq.q) >= MaxQueuedCmds {
		return
	}
	cq.q = slices.Insert(cq.q, i, in)
}

// drain runs the commands this tick allows: one token per tick, up to CmdBurst saved.
func (cq *cmdQueue) drain(run func(Input)) {
	cq.tokens = min(cq.tokens+1, CmdBurst)
	for cq.tokens > 0 && len(cq.q) > 0 {
		in := cq.q[0]
		cq.q = cq.q[1:]
		cq.tokens--
		cq.last = in.Seq
		run(in)
	}
}

type netPlayer struct {
	ID       uint16
	Peer     enet.Peer
	Role     uint8
	Accepted bool   // handshake accepted
	Actor    *Actor // the fighter; nil for spectators and before acceptance
	Cmds     cmdQueue
	LastRecv time.Time
	lastLog  time.Time
}

type Server struct {
	Cfg          ServerConfig
	Host         enet.Host
	W            *World
	Peers        map[uint16]*netPlayer // by ENet incoming peer id
	Bots         []*Actor
	LastRewound  []rl.Rectangle
	accum        float64
	hash         uint64
	nextPlayerID uint16
}

var Srv *Server

func NewServer(cfg ServerConfig) (*Server, error) {
	initNet()
	// spare peers let the server answer extra clients with a rejection
	host, err := enet.NewHost(enet.NewListenAddress(cfg.Port), uint64(cfg.MaxPlayers+cfg.MaxSpectators+4), ChannelCount, 0, 0)
	if err != nil {
		return nil, fmt.Errorf("cannot bind UDP port %d", cfg.Port)
	}
	w := NewWorld(true)
	w.History = NewLagHistory()
	s := &Server{
		Cfg:   cfg,
		Host:  host,
		W:     w,
		Peers: make(map[uint16]*netPlayer),
		hash:  LE.Hash(),
	}
	// deaths and respawns have to reach the server's own registries
	OnFighterRemoved = func(a *Actor) {
		s.Bots = slices.DeleteFunc(s.Bots, func(b *Actor) bool { return b == a })
		for _, p := range s.Peers {
			if p.Actor == a {
				p.Actor = nil
			}
		}
	}
	OnFighterRespawned = func(owner uint16, a *Actor) {
		if owner == 0 {
			s.Bots = append(s.Bots, a)
			return
		}
		for _, p := range s.Peers {
			if p.ID == owner && p.Role == RoleFighter {
				p.Actor = a
				return
			}
		}
		s.W.Remove(a) // the player left while dead
	}
	if cfg.Gamemode != "" || len(cfg.Filterscripts) > 0 {
		h := newScriptHost(w, s.logf)
		h.srv = s
		w.Script = h
		if err := h.loadScripts(cmp.Or(cfg.ScriptDir, ScriptDir), cfg.Gamemode, cfg.Filterscripts); err != nil {
			host.Destroy()
			return nil, err
		}
	}
	s.fillBots()
	s.logf("listening on UDP port %d (max players %d, max spectators %d, bots %d, lagcomp %t%s)", cfg.Port, cfg.MaxPlayers, cfg.MaxSpectators, cfg.Bots, cfg.LagComp, scriptsNamed(cfg))
	return s, nil
}

// scriptsNamed is what the startup line says of the scripts a server runs.
func scriptsNamed(cfg ServerConfig) string {
	out := ""
	if cfg.Gamemode != "" {
		out += ", gamemode " + cfg.Gamemode
	}
	if len(cfg.Filterscripts) > 0 {
		out += ", filterscripts " + strings.Join(cfg.Filterscripts, ",")
	}
	return out
}

func (s *Server) logf(format string, args ...any) {
	fmt.Printf("[server] "+format+"\n", args...)
}

func (s *Server) count(role uint8) int {
	n := 0
	for _, p := range s.Peers {
		if p.Accepted && p.Role == role {
			n++
		}
	}
	return n
}

func (s *Server) humans() int     { return s.count(RoleFighter) }
func (s *Server) spectators() int { return s.count(RoleSpectator) }

func sendBytes(peer enet.Peer, ch uint8, data []byte) {
	var flags enet.PacketFlags
	if ch == ChanReliable {
		flags = enet.PacketFlagReliable
	}
	_ = peer.SendBytes(data, ch, flags)
}

// Poll drains ENet events. It must run on the main goroutine.
func (s *Server) Poll() {
	now := time.Now()
	for {
		ev := s.Host.Service(0)
		switch ev.GetType() {
		case enet.EventNone:
			for id, p := range s.Peers {
				if now.Sub(p.LastRecv) > PeerTimeout {
					p.Peer.DisconnectNow(0)
					s.drop(id, "timed out")
				}
			}
			return
		case enet.EventConnect:
			peer := ev.GetPeer()
			s.Peers[peer.GetIncomingPeerId()] = &netPlayer{Peer: peer, LastRecv: now}
		case enet.EventDisconnect:
			s.drop(ev.GetPeer().GetIncomingPeerId(), "disconnected")
		case enet.EventReceive:
			pkt := ev.GetPacket()
			data := pkt.GetData() // copies out of the C buffer
			pkt.Destroy()
			if p := s.Peers[ev.GetPeer().GetIncomingPeerId()]; p != nil && len(data) > 0 {
				p.LastRecv = now
				s.handle(p, data)
			}
		}
	}
}

func (s *Server) handle(p *netPlayer, data []byte) {
	switch data[0] {
	case MsgHello:
		h, err := decodeHello(data)
		if err != nil || p.Accepted {
			return
		}
		var reason uint8
		switch {
		case h.Proto != ProtocolVersion:
			reason = RejectVersion
		case h.LevelHash != s.hash:
			reason = RejectLevel
		case h.Role == RoleFighter && s.humans() >= s.Cfg.MaxPlayers:
			reason = RejectFull
		case h.Role == RoleSpectator && s.spectators() >= s.Cfg.MaxSpectators:
			reason = RejectSpectatorsFull
		}
		if reason != 0 {
			s.logf("reject %s %s: %s", roleName(h.Role), p.Peer.GetAddress(), rejectText(reason))
			sendBytes(p.Peer, ChanReliable, encodeReject(reason))
			p.Peer.DisconnectLater(0)
			return
		}
		if h.Role == RoleSpectator {
			s.acceptSpectator(p)
		} else {
			s.accept(p)
		}
	case MsgKeepalive:
		// nothing to do: any packet refreshes LastRecv
	case MsgCmds:
		if p.Actor == nil {
			return // spectators (and unaccepted peers) cannot drive a fighter
		}
		cmds, err := decodeCmds(data)
		if err != nil {
			return
		}
		for _, c := range cmds {
			p.Cmds.push(c)
		}
	case MsgBye:
		p.Peer.DisconnectLater(0)
	}
}

func (s *Server) accept(p *netPlayer) {
	s.makeRoom()
	s.nextPlayerID++
	p.ID = s.nextPlayerID
	p.Role = RoleFighter
	p.Accepted = true
	p.Actor = s.W.Spawn(s.W.spawnPos(), true)
	p.Actor.Owner = p.ID
	s.logf("connect player %d (actor %d) from %s", p.ID, p.Actor.ID, p.Peer.GetAddress())
	sendBytes(p.Peer, ChanReliable, encodeWelcome(Welcome{PlayerID: p.ID, ActorID: p.Actor.ID, Tick: s.W.Tick}))
	if s.W.Script != nil {
		s.W.Script.call("OnPlayerConnect", p.ID)
	}
}

// acceptSpectator admits a watcher: no fighter, no slot, bots untouched. ActorID 0 means no fighter.
func (s *Server) acceptSpectator(p *netPlayer) {
	s.nextPlayerID++
	p.ID = s.nextPlayerID
	p.Role = RoleSpectator
	p.Accepted = true
	s.logf("connect spectator %d from %s", p.ID, p.Peer.GetAddress())
	sendBytes(p.Peer, ChanReliable, encodeWelcome(Welcome{PlayerID: p.ID, Tick: s.W.Tick}))
}

func (s *Server) drop(peerID uint16, why string) {
	p := s.Peers[peerID]
	if p == nil {
		return
	}
	delete(s.Peers, peerID)
	switch {
	case p.Accepted && p.Role == RoleFighter:
		s.logf("disconnect player %d: %s", p.ID, why)
		if p.Actor != nil { // nil while dead and waiting to respawn
			s.W.Remove(p.Actor)
		}
		s.fillBots()
		if s.W.Script != nil {
			s.W.Script.call("OnPlayerDisconnect", p.ID, scriptLeaveReason(why))
		}
	case p.Accepted:
		s.logf("disconnect spectator %d: %s", p.ID, why)
	}
}

// scriptLeaveReason is the reason OnPlayerDisconnect is given for why a
// player was dropped.
func scriptLeaveReason(why string) string {
	switch why {
	case "timed out":
		return "timeout"
	case "kicked":
		return "kicked"
	}
	return "quit"
}

// makeRoom removes a bot when a joining human would exceed the slot count.
func (s *Server) makeRoom() {
	if s.humans()+len(s.Bots) >= s.Cfg.MaxPlayers && len(s.Bots) > 0 {
		b := s.Bots[len(s.Bots)-1]
		s.Bots = s.Bots[:len(s.Bots)-1]
		s.W.Remove(b)
	}
}

func (s *Server) fillBots() {
	for len(s.Bots) < s.Cfg.Bots && s.humans()+len(s.Bots) < s.Cfg.MaxPlayers {
		s.Bots = append(s.Bots, s.W.Spawn(s.W.spawnPos(), false))
	}
}

// Step advances the server by one tick.
func (s *Server) Step() {
	s.W.Tick++
	s.W.updateRespawns()
	players := make([]*netPlayer, 0, len(s.Peers))
	for _, p := range s.Peers {
		if p.Actor != nil {
			players = append(players, p)
		}
	}
	slices.SortFunc(players, func(a, b *netPlayer) int { return cmp.Compare(a.ID, b.ID) })
	for _, p := range players {
		p.Cmds.drain(func(in Input) { s.runCmd(p, in) })
	}
	for _, b := range s.Bots {
		b.Input = botInput(b)
		b.Update(float32(FixedTimeStep))
	}
	if s.W.Script != nil {
		s.W.Script.endTick()
	}
	s.W.History.Record(s.W)
	if s.W.Tick%SnapshotEvery == 0 {
		s.sendSnapshots()
	}
}

// runCmd advances p's fighter by one command, with lag-compensated hit tests while attacking or firing.
func (s *Server) runCmd(p *netPlayer, in Input) {
	a := p.Actor
	a.Input = in
	if Debug && time.Since(p.lastLog) > 5*time.Second {
		p.lastLog = time.Now()
		s.logf("player %d view tick %d, server tick %d (%d behind)", p.ID, in.ViewTick, s.W.Tick, int64(s.W.Tick)-int64(in.ViewTick))
	}
	// a swing, or a shot about to be fired: either is tested against what the
	// player saw
	firing := a.Weapon != WeaponNone && in.Attack
	if !s.Cfg.LagComp || a.State != StateAttack && !firing {
		a.Update(float32(FixedTimeStep))
		return
	}
	t := clampViewTime(float64(in.ViewTick)+float64(in.ViewFrac)/256, s.W.Tick, s.Cfg.MaxUnlag)
	saves := s.W.History.rewindOthers(s.W, a, t)
	if Debug {
		s.LastRewound = s.LastRewound[:0]
		for _, sv := range saves {
			s.LastRewound = append(s.LastRewound, sv.a.collisionRect())
		}
	}
	a.Update(float32(FixedTimeStep))
	restore(saves)
}

// sendText sends a line for the message feed to p, or with p nil to every
// accepted client, spectators included.
func (s *Server) sendText(p *netPlayer, text string) {
	data := encodeText(text)
	for _, q := range s.Peers {
		if q.Accepted && q.Peer != nil && (p == nil || q == p) {
			sendBytes(q.Peer, ChanReliable, data)
		}
	}
}

// sendSnapshots sends the world to every accepted peer; spectators' Ack stays 0.
func (s *Server) sendSnapshots() {
	recs := make([]ActorRec, 0, len(s.W.Actors))
	for _, a := range s.W.Actors {
		recs = append(recs, recOf(a))
	}
	for _, p := range s.Peers {
		if !p.Accepted {
			continue
		}
		var yours uint16
		if p.Actor != nil {
			yours = p.Actor.ID
		}
		data := encodeSnapshot(Snapshot{Tick: s.W.Tick, Ack: p.Cmds.last, Yours: yours, Actors: recs})
		sendBytes(p.Peer, ChanUnreliable, data)
	}
}

// Frame runs the server from a render loop (listen server).
func (s *Server) Frame(dt float32) {
	s.Poll()
	s.accum += float64(dt)
	for s.accum > FixedTimeStep {
		s.Step()
		s.accum -= FixedTimeStep
	}
}

// Close tells every client the server is going away and shuts the host down.
func (s *Server) Close() {
	if s.W.Script != nil {
		// every player still in - a hosted game's own player among them -
		// leaves with the server, and the scripts hear it before they exit
		for _, id := range s.W.Script.playerIDs() {
			s.W.Script.call("OnPlayerDisconnect", id, "quit")
		}
		s.W.Script.close()
	}
	for _, p := range s.Peers {
		sendBytes(p.Peer, ChanReliable, encodeBye(ByeServerClosed))
		p.Peer.DisconnectLater(0)
	}
	serviceFor(s.Host, 200*time.Millisecond, len(s.Peers))
	s.Host.Destroy()
	s.logf("closed")
}

// serviceFor flushes a host until `peers` disconnects are seen or d elapses.
func serviceFor(h enet.Host, d time.Duration, peers int) {
	end := time.Now().Add(d)
	for peers > 0 && time.Now().Before(end) {
		ev := h.Service(10)
		switch ev.GetType() {
		case enet.EventReceive:
			ev.GetPacket().Destroy()
		case enet.EventDisconnect:
			peers--
		}
	}
}

// Dedicated runs a headless server until the process is interrupted.
func Dedicated(cfg ServerConfig) {
	initNet()
	defer shutdownNet()

	NetMode = NetDedicated
	LE = NewLevelEditor(39, 83)
	LE.LoadLevel(LevelPath)

	s, err := NewServer(cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	Srv = s
	SV = s.W

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	tick := time.Duration(FixedTimeStep * float64(time.Second))
	next := time.Now()
	for {
		select {
		case <-sig:
			s.Close()
			return
		default:
		}
		s.Poll()
		for now := time.Now(); !now.Before(next); next = next.Add(tick) {
			s.Step()
		}
		time.Sleep(min(time.Until(next), 2*time.Millisecond))
	}
}
