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

// The punches and the block are in the binary.
func TestSoundsAreEmbedded(t *testing.T) {
	for _, p := range append(append([]string{}, punchPaths...), blockPaths...) {
		if len(asset(p)) == 0 {
			t.Errorf("%s is empty", p)
		}
	}
}

// A punch plays on its voices in turn, so the one before is left to finish.
func TestSoundVoicesTakeTurns(t *testing.T) {
	p := &voicedSound{voices: make([]rl.Sound, SoundVoices)}
	for want := range 2 * SoundVoices {
		if got := p.nextVoice(); got != want%SoundVoices {
			t.Fatalf("punch %d went to voice %d, want %d", want, got, want%SoundVoices)
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
// blow, block-all, the start of a clash and a shove are silent.
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
		"a clash":         {func(w *World, def, att *Actor) { startClash(att, def, clashKinds[0]) }, nil},
		"a shove":         {func(w *World, def, att *Actor) { def.enterShovedState(att) }, nil},
	} {
		if n := listen(c.blow, c.prep...); n != 0 {
			t.Errorf("%s was heard %d times, want silence", name, n)
		}
	}
}

// A blow stopped by a guard or by block-all is heard as a block, once, and is
// no punch; a landed hit, a clash and a shove are no block.
func TestABlockIsHeardWhenABlowIsStopped(t *testing.T) {
	blocked, punched := 0, 0
	onBlock = func() { blocked++ }
	onPunch = func() { punched++ }
	t.Cleanup(func() { onBlock, onPunch = nil, nil })
	guardAll := func(def *Actor) { def.Updown = 3 }
	hurt := func(def *Actor) { def.Structure = MaxStructure - 1 }
	for name, c := range map[string]struct {
		blow  func(w *World, def, att *Actor)
		prep  []func(*Actor)
		block int
	}{
		"a guarded block": {func(w *World, def, att *Actor) { def.Updown = att.Updown; def.confirmBlock(att) }, []func(*Actor){hurt}, 1},
		"a full guard":    {func(w *World, def, att *Actor) { def.Updown = att.Updown; def.confirmBlock(att) }, nil, 1},
		"block-all":       {func(w *World, def, att *Actor) { def.confirmBlock(att) }, []func(*Actor){guardAll}, 1},
		"a hit":           {func(w *World, def, att *Actor) { def.takeHit(att) }, nil, 0},
		"a clash":         {func(w *World, def, att *Actor) { startClash(att, def, clashKinds[0]) }, nil, 0},
		"a shove":         {func(w *World, def, att *Actor) { def.enterShovedState(att) }, nil, 0},
	} {
		blocked, punched = 0, 0
		bleeding(t, c.blow, c.prep...)
		if blocked != c.block {
			t.Errorf("%s was heard as %d blocks, want %d", name, blocked, c.block)
		}
		if c.block > 0 && punched != 0 {
			t.Errorf("%s was heard as a punch", name)
		}
	}
}

// A block the blood has not looked at yet - a fighter just seen, or a restore
// by time manipulation - is not heard.
func TestABlockIsNotHeardAfterAReseed(t *testing.T) {
	blocked := 0
	onBlock = func() { blocked++ }
	t.Cleanup(func() { onBlock = nil })
	floor()
	clearBlood()
	t.Cleanup(clearBlood)
	w := NewWorld(true)
	def := w.Spawn(rl.Vector2{X: 0}, false)
	blood.watch(w)
	def.Blocks++
	blood.reseed(w)
	blood.watch(w)
	if blocked != 0 {
		t.Errorf("a reseed was heard as %d blocks", blocked)
	}
}

// A clash is heard as a punch at each of its impacts, once each, as it plays
// out - not twice for the two fighters in it, and not at its start.
func TestAClashIsHeardAtItsImpacts(t *testing.T) {
	for _, k := range clashKinds {
		t.Run(k.left, func(t *testing.T) { clashHeard(t, k) })
	}
}

func clashHeard(t *testing.T, k clashKind) {
	heard := 0
	onPunch = func() { heard++ }
	t.Cleanup(func() { onPunch = nil })
	floor()
	clearBlood()
	t.Cleanup(clearBlood)
	w := NewWorld(true)
	a := w.Spawn(rl.Vector2{X: 0}, false)
	b := w.Spawn(rl.Vector2{X: 30}, false)
	blood.watch(w)
	startClash(a, b, k)
	blood.watch(w)
	if heard != 0 {
		t.Fatalf("the clash was heard %d times as it started, want silence", heard)
	}
	for range 120 {
		w.Tick++
		for _, f := range w.Actors {
			f.Update(float32(FixedTimeStep))
		}
		blood.watch(w)
		if a.State != StateClash && b.State != StateClash {
			break
		}
	}
	if a.State == StateClash || b.State == StateClash {
		t.Fatalf("the clash never ended")
	}
	if heard != len(k.impacts) {
		t.Errorf("the clash was heard %d times, want %d", heard, len(k.impacts))
	}
}
