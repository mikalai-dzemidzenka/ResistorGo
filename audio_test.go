package resistor

import "testing"

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
