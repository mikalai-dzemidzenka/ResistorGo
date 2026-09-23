package resistor

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// scriptDir writes scripts into a fresh directory laid out as the server
// expects: files named "gamemodes/x" or "filterscripts/x" get ".lua" added.
func scriptDir(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, src := range files {
		path := filepath.Join(dir, filepath.FromSlash(name)+".lua")
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

// scriptLog collects what the host logs.
type scriptLog struct{ lines []string }

func (s *scriptLog) logf(format string, args ...any) {
	s.lines = append(s.lines, fmt.Sprintf(format, args...))
}

func (s *scriptLog) has(sub string) bool {
	for _, l := range s.lines {
		if strings.Contains(l, sub) {
			return true
		}
	}
	return false
}

// loaded is a host with the given scripts loaded, failing the test if they
// do not load.
func loaded(t *testing.T, w *World, gamemode string, filters []string, files map[string]string) (*scriptHost, *scriptLog) {
	t.Helper()
	log := &scriptLog{}
	h := newScriptHost(w, log.logf)
	if err := h.loadScripts(scriptDir(t, files), gamemode, filters); err != nil {
		t.Fatal(err)
	}
	return h, log
}

// Filterscripts load in the order named and before the gamemode, every
// callback goes to them in that order, and none sees another's globals.
func TestScriptsLoadInOrderAndApart(t *testing.T) {
	h, log := loaded(t, nil, "gm", []string{"b", "a"}, map[string]string{
		"filterscripts/a": `secret = "a"; function OnFilterScriptInit() print("init a") end
			function Ping() print("ping a", secret) end`,
		"filterscripts/b": `function OnFilterScriptInit() print("init b") end
			function Ping() print("ping b", secret) end`,
		"gamemodes/gm": `function OnGameModeInit() print("init gm") end
			function Ping() print("ping gm", secret) end`,
	})
	h.call("Ping")
	want := []string{
		"[script b] init b", "[script a] init a", "[script gm] init gm",
		"[script b] ping b\tnil", "[script a] ping a\ta", "[script gm] ping gm\tnil",
	}
	if strings.Join(log.lines, "|") != strings.Join(want, "|") {
		t.Errorf("log:\n%s\nwant:\n%s", strings.Join(log.lines, "\n"), strings.Join(want, "\n"))
	}
}

// A script can reach neither the disk nor the process, nor load code.
func TestScriptSandbox(t *testing.T) {
	for _, call := range []string{
		`io.open("x", "w")`, `os.execute("echo")`, `debug.getinfo(1)`,
		`require("x")`, `dofile("x")`, `loadfile("x")`, `load("return 1")`, `loadstring("return 1")`,
	} {
		h, log := loaded(t, nil, "gm", nil, map[string]string{
			"gamemodes/gm": "function Try() " + call + " end",
		})
		h.call("Try")
		if !log.has("[script gm] Try:") {
			t.Errorf("%s did not fail: %q", call, log.lines)
		}
	}
	// and what it does have works
	h, log := loaded(t, nil, "gm", nil, map[string]string{
		"gamemodes/gm": `function Try() print(string.upper("a"), table.concat({1, 2}, ","), math.floor(2.5), bit32.band(6, 3)) end`,
	})
	h.call("Try")
	if !log.has("[script gm] A\t1,2\t2\t2") {
		t.Errorf("the allowed libraries failed: %q", log.lines)
	}
}

// An error stops only the call it happened in: it is logged with the script's
// name, and the next script still gets the callback.
func TestScriptErrorIsContained(t *testing.T) {
	h, log := loaded(t, nil, "gm", []string{"bad"}, map[string]string{
		"filterscripts/bad": `function OnFighterDeath() local x = nil; return x.y end`,
		"gamemodes/gm":      `deaths = 0; function OnFighterDeath() deaths = deaths + 1; print("deaths", deaths) end`,
	})
	h.call("OnFighterDeath", 1, 0)
	h.call("OnFighterDeath", 2, 0)
	if !log.has("[script bad] OnFighterDeath:") || !log.has("[script gm] deaths\t2") {
		t.Errorf("log: %q", log.lines)
	}
}

// A loop that never ends is stopped by the budget, promptly, and the scripts
// after it still run.
func TestScriptBudgetStopsARunawayCall(t *testing.T) {
	h, log := loaded(t, nil, "gm", []string{"loop"}, map[string]string{
		"filterscripts/loop": `function OnPlayerConnect() while true do end end`,
		"gamemodes/gm":       `function OnPlayerConnect(id) print("hello", id) end`,
	})
	start := time.Now()
	h.call("OnPlayerConnect", 1)
	if d := time.Since(start); d > 2*time.Second {
		t.Errorf("the runaway call took %v", d)
	}
	if !log.has("[script loop] OnPlayerConnect:") || !log.has("stopped") || !log.has("[script gm] hello\t1") {
		t.Errorf("log: %q", log.lines)
	}
	// and the budget is per call: the next call runs again
	log.lines = nil
	h.call("OnPlayerConnect", 2)
	if !log.has("[script gm] hello\t2") {
		t.Errorf("the second call did not run: %q", log.lines)
	}
}

// A script that is missing, does not compile or fails as it runs keeps the
// server from starting, and the error names the file.
func TestScriptLoadFailures(t *testing.T) {
	for name, c := range map[string]struct {
		files map[string]string
		want  string
	}{
		"missing":       {map[string]string{}, "gm.lua"},
		"syntax":        {map[string]string{"gamemodes/gm": "function (\n"}, "gm.lua:1"},
		"runtime error": {map[string]string{"gamemodes/gm": "\n\nerror('broken')"}, "gm.lua:3"},
	} {
		h := newScriptHost(nil, (&scriptLog{}).logf)
		err := h.loadScripts(scriptDir(t, c.files), "gm", nil)
		if err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: error %v, want one naming %q", name, err, c.want)
		}
	}
	// an error in the init callback is only logged
	h, log := loaded(t, nil, "gm", nil, map[string]string{"gamemodes/gm": "function OnGameModeInit() error('oops') end"})
	if len(h.scripts) != 1 || !log.has("OnGameModeInit") {
		t.Errorf("an init error was fatal or not logged: %q", log.lines)
	}
}

// Timers fire in ticks: a one-shot once, a repeating one every period until
// killed, and one killed from its own function never again.
func TestScriptTimers(t *testing.T) {
	w := NewWorld(true)
	h, log := loaded(t, w, "gm", nil, map[string]string{"gamemodes/gm": `
		once = SetTimer(function() print("once") end, 100, false)
		every = SetTimer(function() print("every") end, 1000, true)
		n = 0
		self = SetTimer(function() n = n + 1; print("self", n); KillTimer(self) end, 10, true)
		function StopEvery() KillTimer(every) end
	`})
	count := func(line string) int {
		c := 0
		for _, l := range log.lines {
			if l == "[script gm] "+line {
				c++
			}
		}
		return c
	}
	for range 180 {
		w.Tick++
		h.fireTimers()
	}
	if count("once") != 1 || count("every") != 3 || count("self\t1") != 1 || count("self\t2") != 0 {
		t.Errorf("after 3 s: %q", log.lines)
	}
	h.call("StopEvery")
	for range 180 {
		w.Tick++
		h.fireTimers()
	}
	if count("every") != 3 {
		t.Errorf("a killed timer went on firing: %d", count("every"))
	}
	if len(h.timers) != 0 {
		t.Errorf("%d timers left alive", len(h.timers))
	}
}

// A repeating 1000 ms timer is every 60 ticks, and anything shorter than a
// tick is one tick.
func TestScriptTimerPeriod(t *testing.T) {
	w := NewWorld(true)
	h, _ := loaded(t, w, "gm", nil, map[string]string{"gamemodes/gm": `
		SetTimer(function() end, 1000, true)
		SetTimer(function() end, 1, true)
		SetTimer(function() end, 0, false)
	`})
	for i, want := range []uint32{60, 1, 1} {
		if got := h.timers[i].every; got != want {
			t.Errorf("timer %d is every %d ticks, want %d", i+1, got, want)
		}
	}
}
