package resistor

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	lua "github.com/Shopify/go-lua"
	rl "github.com/gen2brain/raylib-go/raylib"
)

// scripted gives w a host with no scripts, to see what it is told.
func scripted(w *World) *scriptHost {
	h := newScriptHost(w, (&scriptLog{}).logf)
	w.Script = h
	return h
}

// luaRun runs code as a gamemode's Run callback in w, answering to srv (which
// may be nil), and returns the global `out` it leaves as a string.
func luaRun(t *testing.T, w *World, srv *Server, code string) (string, *scriptLog) {
	t.Helper()
	h, log := loaded(t, w, "gm", nil, map[string]string{"gamemodes/gm": "function Run() " + code + " end"})
	h.srv = srv
	w.Script = h
	h.call("Run")
	return scriptGlobal(h, "out"), log
}

// scriptGlobal is the global name in the host's first script, as a string.
func scriptGlobal(h *scriptHost, name string) string {
	l := h.scripts[0].l
	l.Global(name)
	defer l.Pop(1)
	if l.IsNoneOrNil(-1) {
		return "nil"
	}
	s, _ := lua.ToStringMeta(l, -1)
	l.Pop(1)
	return s
}

// Spawns, hits, blocks and deaths are queued once each, with the fighters by
// id and what a hit cost; a world without scripts queues nothing.
func TestScriptEventsAreQueued(t *testing.T) {
	floor()
	w := NewWorld(true)
	h := scripted(w)
	def := w.Spawn(rl.Vector2{}, true)
	att := w.Spawn(rl.Vector2{X: -30}, false)
	want := []scriptEvent{{kind: evScriptSpawn, fighter: def.ID}, {kind: evScriptSpawn, fighter: att.ID}}

	cost := def.hitCost()
	def.takeHit(att)
	want = append(want, scriptEvent{kind: evScriptHit, fighter: def.ID, other: att.ID, amount: cost})
	def.Updown = att.Updown
	def.confirmBlock(att)
	want = append(want, scriptEvent{kind: evScriptBlock, fighter: def.ID, other: att.ID})
	def.Structure = 0.1
	def.takeHit(att)
	want = append(want,
		scriptEvent{kind: evScriptHit, fighter: def.ID, other: att.ID, amount: 0.1},
		scriptEvent{kind: evScriptDeath, fighter: def.ID, other: att.ID})

	if fmt.Sprint(h.events) != fmt.Sprint(want) {
		t.Errorf("events\n got %v\nwant %v", h.events, want)
	}

	plain := NewWorld(true)
	a := plain.Spawn(rl.Vector2{}, true)
	a.takeHit(plain.Spawn(rl.Vector2{X: -30}, false))
	if plain.Script != nil {
		t.Error("a world without scripts grew a script host")
	}
}

// A callback runs once the tick is over, so what it changes holds - even for a
// hit decided while lag compensation had the fighter rewound.
func TestAChangeInOnFighterHitHolds(t *testing.T) {
	floor()
	w := NewWorld(true)
	w.History = NewLagHistory()
	h, _ := loaded(t, w, "gm", nil, map[string]string{"gamemodes/gm": `
		function OnFighterHit(id) SetFighterStructure(id, 12) end`})
	w.Script = h
	att := w.Spawn(rl.Vector2{}, true)
	def := w.Spawn(rl.Vector2{X: 20}, true)
	def.Updown = 0
	for tick := uint32(1); tick <= 20; tick++ {
		w.Tick = tick
		w.History.Record(w)
	}
	def.Pos.X, def.Updown = 200, 1 // out of reach and guarding now; not when it was seen
	att.Updown, att.Orientation = 1, 1
	att.Ticks = 100
	att.SetAnim("resources/threefour/attack-mid.png", AnimOptions{animDuration: 0.3, totalFrames: 8, once: true})
	att.State = StateAttack
	att.Ticks = 114
	att.advanceAnim()

	s := &Server{Cfg: ServerConfig{LagComp: true, MaxUnlag: DefaultMaxUnlag}, W: w, Peers: map[uint16]*netPlayer{}}
	h.srv = s
	p := &netPlayer{ID: 1, Role: RoleFighter, Actor: att} // not accepted: no snapshots to a peer that is not there
	p.Cmds.push(Input{Seq: 1, Updown: 1, ViewTick: 10})
	s.Peers[0] = p
	h.events = nil // the spawns above
	s.Step()

	if def.HitCount == 0 {
		t.Fatal("the rewound swing did not land")
	}
	if r := recOf(def); r.Structure != uint8(MaxStructure*10) {
		t.Errorf("the snapshot has the defender at %v structure, want the %v the script healed it to", float32(r.Structure)/10, MaxStructure)
	}
}

// Players are told of as they connect and leave, with why they left; init and
// exit bracket the server's life.
func TestScriptPlayerCallbacks(t *testing.T) {
	initNet()
	defer shutdownNet()
	floor()
	dir := scriptDir(t, map[string]string{"gamemodes/gm": `
		seen = ""
		function OnGameModeInit() seen = seen .. "init;" end
		function OnGameModeExit() seen = seen .. "exit;" end
		function OnPlayerConnect(id) seen = seen .. "connect " .. id .. ";" end
		function OnPlayerDisconnect(id, why) seen = seen .. "leave " .. id .. " " .. why .. ";" end
		function OnFighterSpawn(f, id) spawned = (spawned or "") .. id .. ";" end
		function Kick(id) KickPlayer(id) end
	`})
	cfg := DefaultServerConfig()
	cfg.Port = 27196
	cfg.Gamemode, cfg.ScriptDir = "gm", dir
	s, err := NewServer(cfg)
	if err != nil {
		t.Fatal(err)
	}
	gm := s.W.Script.scripts[0]
	hash := LE.Hash()
	fighter := Hello{Proto: ProtocolVersion, LevelHash: hash, Role: RoleFighter}
	a := hello(t, s, nil, cfg.Port, fighter)
	b := hello(t, s, nil, cfg.Port, fighter)
	pumpFor(100*time.Millisecond, s, nil, a, b)
	sendBytes(a.p, ChanReliable, encodeBye(0))
	pumpFor(200*time.Millisecond, s, nil, a, b)
	s.W.Script.call("Kick", 2)
	pumpFor(100*time.Millisecond, s, nil, b)
	// still in when the server closes, as a hosted game's own player is
	c := hello(t, s, nil, cfg.Port, fighter)
	pumpFor(100*time.Millisecond, s, nil, c)
	host := s.W.Script
	s.Close()

	// A player's fighter is told of on the tick after it connects, so where it
	// falls among the connects is up to the timing; it is checked on its own.
	want := "init;connect 1;connect 2;leave 1 quit;leave 2 kicked;connect 3;leave 3 quit;exit;"
	if seen := scriptGlobal(host0(host, gm), "seen"); seen != want {
		t.Errorf("callbacks\n got %s\nwant %s", seen, want)
	}
	if spawned := scriptGlobal(host0(host, gm), "spawned"); spawned != "1;2;3;" {
		t.Errorf("fighters spawned for players %q, want 1;2;3;", spawned)
	}
	if reason := scriptLeaveReason("timed out"); reason != "timeout" {
		t.Errorf("a timed out player leaves with %q, want timeout", reason)
	}
}

// host0 is a host holding just s, for reading a script's globals after its
// server let go of it.
func host0(h *scriptHost, s *script) *scriptHost {
	return &scriptHost{scripts: []*script{s}, w: h.w, logf: h.logf}
}

// Players and fighters are listed and found by id, and an id gone stale
// answers nil rather than failing.
func TestScriptFindsPlayersAndFighters(t *testing.T) {
	floor()
	w := NewWorld(true)
	you := w.Spawn(rl.Vector2{}, true)
	you.Owner = 3
	bot := w.Spawn(rl.Vector2{X: 60}, false)
	player := &netPlayer{ID: 3, Accepted: true, Role: RoleFighter, Actor: you}
	s := &Server{W: w, Peers: map[uint16]*netPlayer{
		0: player,
		1: {ID: 4, Accepted: true, Role: RoleSpectator},
	}}
	out, log := luaRun(t, w, s, fmt.Sprintf(`
		out = table.concat(GetPlayers(), ",") .. "|" .. tostring(IsPlayerConnected(3)) .. tostring(IsPlayerConnected(4)) ..
			"|" .. GetPlayerFighter(3) .. "|" .. table.concat(GetFighters(), ",") ..
			"|" .. tostring(IsBot(%d)) .. tostring(IsBot(%d)) .. "|" .. GetFighterPlayer(%d) .. GetFighterPlayer(%d) ..
			"|" .. GetTick()`, you.ID, bot.ID, you.ID, bot.ID))
	want := fmt.Sprintf("3|truefalse|%d|%d,%d|falsetrue|30|0", you.ID, you.ID, bot.ID)
	if out != want {
		t.Errorf("got %q, want %q (log %q)", out, want, log.lines)
	}

	// the bot is gone and the player has come back as a new fighter
	w.Remove(bot)
	w.Remove(you)
	again := w.Spawn(rl.Vector2{}, true)
	player.Actor = again
	out, log = luaRun(t, w, s, fmt.Sprintf(`
		out = tostring(GetFighterStructure(%d)) .. tostring(GetFighterPos(%d)) .. tostring(IsBot(%d)) ..
			tostring(SetFighterTeam(%d, 2)) .. tostring(KillFighter(%d)) .. "|" .. GetPlayerFighter(3) ..
			"|" .. tostring(GetPlayerFighter(99))`, bot.ID, bot.ID, bot.ID, bot.ID, bot.ID))
	want = fmt.Sprintf("nilnilfalsefalsefalse|%d|nil", again.ID)
	if out != want {
		t.Errorf("got %q, want %q (log %q)", out, want, log.lines)
	}
}

// What a script sets is clamped to the fighter's range, and a fighter is not
// put inside the level's solid ground.
func TestScriptSetsFighters(t *testing.T) {
	floor()
	w := NewWorld(true)
	a := w.Spawn(rl.Vector2{}, true)
	for range 10 {
		a.Update(float32(FixedTimeStep)) // settle onto the floor
	}
	stood := a.Pos
	id := a.ID
	out, log := luaRun(t, w, nil, fmt.Sprintf(`
		SetFighterStructure(%[1]d, 99); SetFighterEnergy(%[1]d, -5); SetFighterTeam(%[1]d, 300)
		local s, e, team = GetFighterStructure(%[1]d), GetFighterEnergy(%[1]d), GetFighterTeam(%[1]d)
		local wall = SetFighterPos(%[1]d, 0, %[2]d)
		out = s .. "," .. e .. "," .. team .. "," .. tostring(wall)`, id, floorY))
	if out != fmt.Sprintf("%v,0,255,false", MaxStructure) {
		t.Errorf("got %q (log %q)", out, log.lines)
	}
	if a.Pos != stood {
		t.Errorf("a refused move moved the fighter from %v to %v", stood, a.Pos)
	}
	out, _ = luaRun(t, w, nil, fmt.Sprintf(`
		local ok = SetFighterPos(%d, 40, %d); local x, y = GetFighterPos(%d)
		out = tostring(ok) .. "," .. x .. "," .. y`, id, floorY-48-40, id))
	if out != fmt.Sprintf("true,40,%d", floorY-48-40) {
		t.Errorf("a move into the open: %q", out)
	}
}

// A killed fighter dies with nobody to blame and comes back; a bot a script
// adds is a server bot and is told of as one; a kicked player is dropped.
func TestScriptKillsAddsAndKicks(t *testing.T) {
	floor()
	w := NewWorld(true)
	a := w.Spawn(rl.Vector2{}, true)
	a.Owner = 5
	s := &Server{W: w, Peers: map[uint16]*netPlayer{}}
	h, log := loaded(t, w, "gm", nil, map[string]string{"gamemodes/gm": fmt.Sprintf(`
		deaths, spawns, left = "", "", ""
		function OnFighterDeath(id, killer) deaths = deaths .. id .. ":" .. killer .. ";" end
		function OnFighterSpawn(id, player) spawns = spawns .. player .. ";" end
		function OnPlayerDisconnect(id, why) left = left .. id .. " " .. why end
		function Run() KillFighter(%d); kicked = KickPlayer(6); gone = RemoveBot(AddBot()); kept = AddBot() end`, a.ID)})
	h.srv = s
	w.Script = h
	s.Peers[0] = &netPlayer{ID: 6, Accepted: true, Role: RoleFighter}
	h.call("Run")

	if a.State != StateDeath || a.Structure != 0 {
		t.Errorf("KillFighter left the fighter in state %d with %v structure", a.State, a.Structure)
	}
	if scriptGlobal(h, "kicked") != "true" || len(s.Peers) != 0 {
		t.Errorf("the kick did not drop the player: %s, %d peers (log %q)", scriptGlobal(h, "kicked"), len(s.Peers), log.lines)
	}
	kept := w.ByID(uint16(mustAtoi(t, scriptGlobal(h, "kept"))))
	if scriptGlobal(h, "gone") != "true" || len(s.Bots) != 1 || s.Bots[0] != kept || kept.IsPlayer {
		t.Errorf("AddBot and RemoveBot: gone %s, server bots %v, kept %v", scriptGlobal(h, "gone"), s.Bots, kept)
	}
	h.endTick()
	if got := scriptGlobal(h, "deaths"); got != fmt.Sprintf("%d:0;", a.ID) {
		t.Errorf("deaths %q (log %q)", got, log.lines)
	}
	if got := scriptGlobal(h, "left"); got != "6 kicked" {
		t.Errorf("left %q", got)
	}
	if got := scriptGlobal(h, "spawns"); got != "0;" {
		t.Errorf("the kept bot was told of as %q, want player 0 once", got)
	}

	// the killed fighter falls, is taken away and comes back as a new one
	for range 600 {
		w.Tick++
		w.updateRespawns()
		for _, f := range w.Actors {
			f.Update(float32(FixedTimeStep))
		}
		if w.ByID(a.ID) == nil && len(w.Actors) == 2 {
			break
		}
	}
	back := false
	for _, f := range w.Actors {
		back = back || (f.IsPlayer && f.Owner == 5 && f.ID != a.ID)
	}
	if !back {
		t.Error("the killed fighter never came back")
	}
}

func mustAtoi(t *testing.T, s string) int {
	t.Helper()
	n, err := strconv.Atoi(s)
	if err != nil {
		t.Fatalf("%q is not a number", s)
	}
	return n
}

// A line goes to the one player it is sent to, or to every client, spectators
// included, and lands in the client's feed.
func TestScriptMessagesReachClients(t *testing.T) {
	initNet()
	defer shutdownNet()
	floor()
	cfg := DefaultServerConfig()
	cfg.Port = 27197
	cfg.Gamemode, cfg.ScriptDir = "gm", scriptDir(t, map[string]string{"gamemodes/gm": `
		function ToOne(id) SendPlayerMessage(id, "just you") end
		function ToAll() SendMessageToAll("everyone") end`})
	s, err := NewServer(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	hash := LE.Hash()
	a := hello(t, s, nil, cfg.Port, Hello{Proto: ProtocolVersion, LevelHash: hash, Role: RoleFighter})
	b := hello(t, s, nil, cfg.Port, Hello{Proto: ProtocolVersion, LevelHash: hash, Role: RoleFighter})
	sp := hello(t, s, nil, cfg.Port, Hello{Proto: ProtocolVersion, LevelHash: hash, Role: RoleSpectator})
	s.W.Script.call("ToOne", 1)
	s.W.Script.call("ToAll")
	pumpFor(200*time.Millisecond, s, nil, a, b, sp)

	texts := func(r *rawPeer) []string {
		var out []string
		for _, m := range r.got {
			if len(m) > 0 && m[0] == MsgText {
				text, err := decodeText(m)
				if err != nil {
					t.Fatal(err)
				}
				out = append(out, text)
			}
		}
		return out
	}
	for name, c := range map[string]struct {
		r    *rawPeer
		want string
	}{"player 1": {a, "just you|everyone"}, "player 2": {b, "everyone"}, "the spectator": {sp, "everyone"}} {
		if got := strings.Join(texts(c.r), "|"); got != c.want {
			t.Errorf("%s got %q, want %q", name, got, c.want)
		}
	}

	var c Client
	c.recv(encodeText("into the feed"))
	c.recv([]byte{MsgText, 9, 'x'}) // malformed: ignored
	if got := c.Feed.shown(time.Now()); len(got) != 1 || got[0] != "into the feed" {
		t.Errorf("the client's feed has %q", got)
	}
}

// A gamemode that does not load keeps the server from starting: the error
// names the file, and the port is left free.
func TestABrokenGamemodeStopsTheServer(t *testing.T) {
	initNet()
	defer shutdownNet()
	floor()
	cfg := DefaultServerConfig()
	cfg.Port = 27198
	cfg.Gamemode, cfg.ScriptDir = "broken", scriptDir(t, map[string]string{"gamemodes/broken": "function OnGameModeInit(\n"})
	if s, err := NewServer(cfg); err == nil || !strings.Contains(err.Error(), "gamemodes/broken.lua:") {
		if s != nil {
			s.Close()
		}
		t.Fatalf("a broken gamemode gave %v", err)
	}
	cfg.Gamemode = ""
	s, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("the port was not left free: %v", err)
	}
	s.Close()
	if got := scriptsNamed(ServerConfig{Gamemode: "tdm", Filterscripts: []string{"a", "b"}}); got != ", gamemode tdm, filterscripts a,b" {
		t.Errorf("the startup line says %q", got)
	}
}

// The example scripts load, and the team deathmatch sorts players into two
// even teams that keep across a respawn, leaves bots neutral, and scores only
// a kill of the other side.
func TestExampleScripts(t *testing.T) {
	floor()
	w := NewWorld(true)
	log := &scriptLog{}
	h := newScriptHost(w, log.logf)
	h.srv = &Server{W: w, Peers: map[uint16]*netPlayer{}}
	w.Script = h
	if err := h.loadScripts("scripts", "tdm", []string{"welcome"}); err != nil {
		t.Fatal(err)
	}
	spawn := func(x float32, owner uint16) *Actor {
		a := w.Spawn(rl.Vector2{X: x}, owner > 0)
		a.Owner = owner
		return a
	}
	p1, bot1, p2, bot2 := spawn(0, 1), spawn(60, 0), spawn(120, 2), spawn(180, 0)
	h.endTick()
	if p1.Team == p2.Team || (p1.Team != RedTeam && p1.Team != BlueTeam) || (p2.Team != RedTeam && p2.Team != BlueTeam) {
		t.Fatalf("players on teams %d and %d, want one on Red and one on Blue", p1.Team, p2.Team)
	}
	if bot1.Team != BotTeam || bot2.Team != BotTeam {
		t.Errorf("bots on teams %d and %d, want both left on the bots' team", bot1.Team, bot2.Team)
	}
	red, blue := p1, p2
	if red.Team != RedTeam {
		red, blue = p2, p1
	}

	// breaking a neutral bot scores nothing
	bot1.Structure = 0.1
	bot1.takeHit(red)
	h.endTick()

	blue.Structure = 0.1
	blue.takeHit(red)
	back := spawn(300, blue.Owner)
	back.Team = blue.Team // as a respawn comes back
	h.endTick()
	if back.Team != BlueTeam {
		t.Errorf("a respawned Blue player was moved to team %d", back.Team)
	}
	for _, l := range log.lines {
		// a failed call is logged as "[script name] Callback: error"
		if _, rest, _ := strings.Cut(l, "] "); strings.HasPrefix(rest, "On") || strings.HasPrefix(rest, "timer ") {
			t.Errorf("a script failed: %s", l)
		}
	}
	h.close()
	if !log.has("final score: Red 1 : 0 Blue") {
		t.Errorf("want only the kill of the other side scored: %q", log.lines)
	}
}
