package resistor

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"

	lua "github.com/Shopify/go-lua"
)

// Scripting, in the manner of San Andreas Multiplayer: the server runs one
// gamemode and any number of filterscripts, each a Lua file in a state of its
// own, and tells them what happens in the match through global callbacks.
// Clients and single player run none; the scripts reach players through what
// they change in the world and the lines of text they send.

// Gamemode and Filterscripts are the scripts a server started from this
// process runs, set from the command line; ScriptDir is where they are read.
var (
	Gamemode      string
	Filterscripts []string
	ScriptDir     = "scripts"
)

// ScriptBudget is how many Lua instructions one call into a script - a
// callback, a timer, a script's main chunk - may run before it is stopped.
// Roughly a twentieth of a second in go-lua: long enough for any sane
// callback, short enough that a loop that never ends costs the server a few
// ticks rather than the match.
const (
	ScriptBudget     = 5_000_000
	scriptHookEvery  = 10_000
	scriptTimersKey  = "resistor.timers"
	scriptMaxTimerMs = 24 * 60 * 60 * 1000
)

// scriptTickRate is server ticks per second, for timers given in milliseconds.
var scriptTickRate = math.Round(1 / FixedTimeStep)

type scriptKind uint8

const (
	filterScript scriptKind = iota
	gamemodeScript
)

type script struct {
	name  string
	kind  scriptKind
	l     *lua.State
	host  *scriptHost
	steps int // instructions run in the current call, counted by the hook
}

type scriptTimer struct {
	id        int
	s         *script
	every     uint32 // ticks
	due       uint32
	repeating bool
	dead      bool
}

// scriptHost is every script a server runs, filterscripts first in the order
// they were named and then the gamemode, which is the order every callback is
// delivered in.
type scriptHost struct {
	scripts   []*script
	timers    []*scriptTimer
	nextTimer int
	events    []scriptEvent
	w         *World
	srv       *Server // nil when a test runs scripts without one
	logf      func(format string, args ...any)
}

func newScriptHost(w *World, logf func(string, ...any)) *scriptHost {
	if logf == nil {
		logf = func(format string, args ...any) { fmt.Printf(format+"\n", args...) }
	}
	return &scriptHost{w: w, logf: logf}
}

// loadScripts reads and runs the named filterscripts and then the gamemode
// from dir, calling each one's init callback once it has run. A script that is
// missing, does not compile or fails while its main chunk runs is an error: a
// server with half its rules loaded is not a server to start. An error in an
// init callback is only logged, as it would be in any other callback.
func (h *scriptHost) loadScripts(dir, gamemode string, filterscripts []string) error {
	for _, name := range filterscripts {
		if err := h.load(dir, "filterscripts/"+name+".lua", name, filterScript); err != nil {
			return err
		}
	}
	if gamemode != "" {
		if err := h.load(dir, "gamemodes/"+gamemode+".lua", gamemode, gamemodeScript); err != nil {
			return err
		}
	}
	return nil
}

// load runs the script at rel under dir. The chunk is named by rel, which is
// what error messages show: Lua cuts a long chunk name down to its end.
func (h *scriptHost) load(dir, rel, name string, kind scriptKind) error {
	src, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(rel)))
	if err != nil {
		return fmt.Errorf("script %v", err)
	}
	s := &script{name: name, kind: kind, host: h}
	s.newState()
	if err := lua.LoadBuffer(s.l, string(src), "@"+rel, "t"); err != nil {
		return fmt.Errorf("script %s", s.errorText(err))
	}
	s.steps = 0
	if err := s.l.ProtectedCall(0, 0, 0); err != nil {
		return fmt.Errorf("script %s", s.errorText(err))
	}
	h.scripts = append(h.scripts, s)
	if kind == gamemodeScript {
		s.call("OnGameModeInit")
	} else {
		s.call("OnFilterScriptInit")
	}
	return nil
}

// errorText pops the error a failed load or call left on the stack.
func (s *script) errorText(err error) string {
	msg, ok := s.l.ToString(-1)
	s.l.Pop(1)
	if !ok || msg == "" {
		msg = err.Error()
	}
	return msg
}

// newState is a Lua state with the sandbox's libraries and nothing else: no
// io, os, debug or package, and nothing in the base library that reads or
// runs code from a file or a string.
func (s *script) newState() {
	l := lua.NewState()
	s.l = l
	for _, lib := range []lua.RegistryFunction{
		{Name: "_G", Function: lua.BaseOpen},
		{Name: "string", Function: lua.StringOpen},
		{Name: "table", Function: lua.TableOpen},
		{Name: "math", Function: lua.MathOpen},
		{Name: "bit32", Function: lua.Bit32Open},
	} {
		lua.Require(l, lib.Name, lib.Function, true)
		l.Pop(1)
	}
	for _, name := range []string{"dofile", "loadfile", "load", "loadstring"} {
		l.PushNil()
		l.SetGlobal(name)
	}
	l.Register("print", func(l *lua.State) int {
		parts := make([]string, 0, l.Top())
		for i := 1; i <= l.Top(); i++ {
			str, _ := lua.ToStringMeta(l, i)
			l.Pop(1)
			parts = append(parts, str)
		}
		s.host.logf("[script %s] %s", s.name, strings.Join(parts, "\t"))
		return 0
	})
	l.NewTable()
	l.SetField(lua.RegistryIndex, scriptTimersKey)
	registerScriptTimers(s)
	registerScriptAPI(s)
	lua.SetDebugHook(l, func(l *lua.State, _ lua.Debug) {
		s.steps += scriptHookEvery
		if s.steps > ScriptBudget {
			lua.Errorf(l, "stopped: ran past %d instructions", ScriptBudget)
		}
	}, lua.MaskCount, scriptHookEvery)
}

// call calls the global function name in the script, if it defines one,
// protected and on a fresh budget. A failure is logged with the script's name
// and goes no further.
func (s *script) call(name string, args ...any) {
	l := s.l
	l.Global(name)
	if !l.IsFunction(-1) {
		l.Pop(1)
		return
	}
	for _, a := range args {
		pushScriptValue(l, a)
	}
	s.run(len(args), name)
}

// run calls the function under its nargs arguments on the stack.
func (s *script) run(nargs int, what string) bool {
	s.steps = 0
	if err := s.l.ProtectedCall(nargs, 0, 0); err != nil {
		s.host.logf("[script %s] %s: %s", s.name, what, s.errorText(err))
		return false
	}
	return true
}

func pushScriptValue(l *lua.State, v any) {
	switch v := v.(type) {
	case nil:
		l.PushNil()
	case bool:
		l.PushBoolean(v)
	case int:
		l.PushInteger(v)
	case uint16:
		l.PushInteger(int(v))
	case uint32:
		l.PushInteger(int(v))
	case float32:
		l.PushNumber(float64(v))
	case float64:
		l.PushNumber(v)
	case string:
		l.PushString(v)
	default:
		panic(fmt.Sprintf("pushScriptValue: %T", v))
	}
}

// call delivers a callback to every script, filterscripts first.
func (h *scriptHost) call(name string, args ...any) {
	for _, s := range h.scripts {
		s.call(name, args...)
	}
}

// close tells every script the server is going away and drops their timers.
func (h *scriptHost) close() {
	for _, s := range h.scripts {
		if s.kind == gamemodeScript {
			s.call("OnGameModeExit")
		} else {
			s.call("OnFilterScriptExit")
		}
	}
	h.timers = nil
	h.scripts = nil
}

func (h *scriptHost) tick() uint32 {
	if h.w == nil {
		return 0
	}
	return h.w.Tick
}

// registerScriptTimers gives a script SetTimer and KillTimer. A timer's
// function is kept in the state's registry under the timer's id, since
// go-lua has no references to hold it by.
func registerScriptTimers(s *script) {
	h := s.host
	s.l.Register("SetTimer", func(l *lua.State) int {
		lua.CheckType(l, 1, lua.TypeFunction)
		ms := min(max(lua.CheckNumber(l, 2), 0), scriptMaxTimerMs)
		repeating := l.ToBoolean(3)
		every := max(uint32(math.Ceil(ms*scriptTickRate/1000-1e-9)), 1)
		h.nextTimer++
		t := &scriptTimer{id: h.nextTimer, s: s, every: every, due: h.tick() + every, repeating: repeating}
		l.Field(lua.RegistryIndex, scriptTimersKey)
		l.PushValue(1)
		l.RawSetInt(-2, t.id)
		l.Pop(1)
		h.timers = append(h.timers, t)
		l.PushInteger(t.id)
		return 1
	})
	s.l.Register("KillTimer", func(l *lua.State) int {
		id := lua.CheckInteger(l, 1)
		for _, t := range h.timers {
			if t.id == id && t.s == s {
				t.kill()
			}
		}
		return 0
	})
}

func (t *scriptTimer) kill() {
	if t.dead {
		return
	}
	t.dead = true
	l := t.s.l
	l.Field(lua.RegistryIndex, scriptTimersKey)
	l.PushNil()
	l.RawSetInt(-2, t.id)
	l.Pop(1)
}

// fireTimers calls every timer due by now, in the order they were set. A timer
// set by one of them is not due before the next tick, and one killed by one of
// them does not fire.
func (h *scriptHost) fireTimers() {
	now := h.tick()
	for i := 0; i < len(h.timers); i++ {
		t := h.timers[i]
		if t.dead || now < t.due {
			continue
		}
		if t.repeating {
			t.due += t.every
		}
		l := t.s.l
		l.Field(lua.RegistryIndex, scriptTimersKey)
		l.RawGetInt(-1, t.id)
		l.Remove(-2)
		t.s.run(0, fmt.Sprintf("timer %d", t.id))
		if !t.repeating {
			t.kill()
		}
	}
	live := h.timers[:0]
	for _, t := range h.timers {
		if !t.dead {
			live = append(live, t)
		}
	}
	h.timers = live
}

// scriptEvent is something that happened in the match during a tick, kept
// until the tick is over and then delivered: a callback run in the middle of
// one could be undone by the lag compensation's restore, or change the list of
// fighters the server is walking. Fighters are kept by id and looked up when
// the event is delivered, so one gone by then is simply not found.
type scriptEvent struct {
	kind    scriptEventKind
	fighter uint16
	other   uint16  // the attacker or killer; 0 for nobody
	amount  float32 // the structure a hit cost
}

type scriptEventKind uint8

const (
	evScriptSpawn scriptEventKind = iota
	evScriptHit
	evScriptBlock
	evScriptDeath
)

// scriptEvent queues e for the scripts, if the world runs any.
func (w *World) scriptEvent(kind scriptEventKind, fighter, other *Actor, amount float32) {
	if w == nil || w.Script == nil {
		return
	}
	e := scriptEvent{kind: kind, fighter: fighter.ID, amount: amount}
	if other != nil {
		e.other = other.ID
	}
	w.Script.events = append(w.Script.events, e)
}

// deliver hands the tick's events to the scripts, in the order they happened.
// Events the scripts cause while they are told - a bot added, a fighter
// killed - wait for the next tick, so a script cannot keep a tick going for
// ever by answering its own events.
func (h *scriptHost) deliver() {
	events := h.events
	h.events = nil
	for _, e := range events {
		switch e.kind {
		case evScriptSpawn:
			a := h.w.ByID(e.fighter)
			if a == nil {
				continue
			}
			h.call("OnFighterSpawn", e.fighter, a.Owner)
		case evScriptHit:
			h.call("OnFighterHit", e.fighter, e.other, e.amount)
		case evScriptBlock:
			h.call("OnFighterBlock", e.fighter, e.other)
		case evScriptDeath:
			h.call("OnFighterDeath", e.fighter, e.other)
		}
	}
}

// endTick is the scripts' turn, once a server tick has run: what happened in
// it, and then the timers due.
func (h *scriptHost) endTick() {
	h.deliver()
	h.fireTimers()
}
