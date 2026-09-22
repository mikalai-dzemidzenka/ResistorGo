package resistor

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// Both tracks the game plays are in the binary.
func TestMusicIsEmbedded(t *testing.T) {
	for _, p := range []string{menuMusicPath, levelMusicPath} {
		if len(asset(p)) == 0 {
			t.Errorf("%s is empty", p)
		}
	}
}

// The menu's track on the menu, the level's in a match, ducked while the match
// is held, and neither anywhere else.
func TestMusicTargets(t *testing.T) {
	for _, c := range []struct {
		stage       int
		held        bool
		menu, level float32
	}{
		{StageMainMenu, false, 1, 0},
		{StageMainMenu, true, 1, 0},
		{StageLevel, false, 0, 1},
		{StageLevel, true, 0, LevelMusicDuck},
		{-1, false, 0, 0},
	} {
		menu, level := musicTargets(c.stage, c.held)
		if menu != c.menu || level != c.level {
			t.Errorf("stage %d held %v: menu %v level %v, want %v %v", c.stage, c.held, menu, level, c.menu, c.level)
		}
	}
	if LevelMusicDuck <= 0 || LevelMusicDuck >= 1 {
		t.Errorf("the duck is %v: held, the level track should play on, quieter", LevelMusicDuck)
	}
}

// A rewinding track stops at silence and starts from the top when wanted again;
// the menu's pauses and resumes where it was. Ducked, both keep playing.
func TestNextPlayback(t *testing.T) {
	for _, c := range []struct {
		name                     string
		rewind, stopped, playing bool
		fade                     float32
		want                     playback
	}{
		{"level fades out", true, false, true, 0, stopPlayback},
		{"level silent and stopped", true, true, false, 0, keepSilent},
		{"level wanted again", true, true, false, 0.1, startPlayback},
		{"level ducked", true, false, true, LevelMusicDuck, keepPlaying},
		{"menu fades out", false, false, true, 0, pausePlayback},
		{"menu silent and paused", false, false, false, 0, keepSilent},
		{"menu wanted again", false, false, false, 0.1, resumePlayback},
		{"menu at first", false, true, false, 1, startPlayback},
	} {
		if got := nextPlayback(c.rewind, c.stopped, c.playing, c.fade); got != c.want {
			t.Errorf("%s: %d, want %d", c.name, got, c.want)
		}
	}
}

// Both punches are in the binary.
func TestPunchesAreEmbedded(t *testing.T) {
	for _, p := range punchPaths {
		if len(asset(p)) == 0 {
			t.Errorf("%s is empty", p)
		}
	}
}

// A punch plays on its voices in turn, so the one before is left to finish.
func TestPunchVoicesTakeTurns(t *testing.T) {
	p := &punchSound{voices: make([]rl.Sound, PunchVoices)}
	for want := range 2 * PunchVoices {
		if got := p.nextVoice(); got != want%PunchVoices {
			t.Fatalf("punch %d went to voice %d, want %d", want, got, want%PunchVoices)
		}
	}
}

// The audio draws only from its own random source, as the blood does: the bots
// draw from the global one.
func TestAudioHasItsOwnRandom(t *testing.T) {
	f, err := parser.ParseFile(token.NewFileSet(), "audio.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	ast.Inspect(f, func(n ast.Node) bool {
		sel, ok := n.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		if pkg, ok := sel.X.(*ast.Ident); ok && pkg.Name == "rand" && sel.Sel.Name != "New" && sel.Sel.Name != "NewSource" && sel.Sel.Name != "Rand" {
			t.Errorf("audio.go calls the global rand.%s", sel.Sel.Name)
		}
		return true
	})
}

// A landed hit is heard once, whether it flinches, stuns or kills; a blocked
// blow, block-all, a clash and a shove are silent.
func TestAPunchIsHeardWhenAHitLands(t *testing.T) {
	heard := 0
	onPunch = func() { heard++ }
	t.Cleanup(func() { onPunch = nil })
	listen := func(blow func(w *World, def, att *Actor), prep ...func(*Actor)) int {
		heard = 0
		bleeding(t, blow, prep...)
		return heard
	}
	for name, blow := range map[string]func(w *World, def, att *Actor){
		"a flinch": func(w *World, def, att *Actor) { def.takeHit(att) },
		"a stun": func(w *World, def, att *Actor) {
			def.HitCount = StunEveryHits - 1
			def.takeHit(att)
		},
		"a kill": func(w *World, def, att *Actor) {
			def.Structure = 1
			def.takeHit(att)
		},
	} {
		if n := listen(blow); n != 1 {
			t.Errorf("%s was heard %d times, want once", name, n)
		}
	}
	guardAll := func(def *Actor) { def.Updown = 3 }
	hurt := func(def *Actor) { def.Structure = MaxStructure - 1 }
	for name, c := range map[string]struct {
		blow func(w *World, def, att *Actor)
		prep []func(*Actor)
	}{
		"a guarded block": {func(w *World, def, att *Actor) { def.Updown = att.Updown; def.confirmBlock(att) }, []func(*Actor){hurt}},
		"block-all":       {func(w *World, def, att *Actor) { def.confirmBlock(att) }, []func(*Actor){guardAll}},
		"a clash":         {func(w *World, def, att *Actor) { startClash(att, def) }, nil},
		"a shove":         {func(w *World, def, att *Actor) { def.enterShovedState(att) }, nil},
	} {
		if n := listen(c.blow, c.prep...); n != 0 {
			t.Errorf("%s was heard %d times, want silence", name, n)
		}
	}
}
