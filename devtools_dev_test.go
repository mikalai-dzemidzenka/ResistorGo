//go:build !release

package resistor

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDevToolsInDevBuild(t *testing.T) {
	if !DevTools {
		t.Errorf("an untagged build has DevTools false, want the developer tools compiled in")
	}
}

// F2 toggles edit mode, but only in single player and never while time
// manipulation is on.
func TestEditModeToggle(t *testing.T) {
	defer func() { Editing = false }()
	Editing = false

	toggleEditing(true, false)
	if !Editing {
		t.Errorf("F2 in single player did not enter edit mode")
	}
	toggleEditing(true, false)
	if Editing {
		t.Errorf("a second F2 did not leave edit mode")
	}

	toggleEditing(true, true)
	if Editing {
		t.Errorf("edit mode was entered while time manipulation was on")
	}
	toggleEditing(false, false)
	if Editing {
		t.Errorf("edit mode was entered in a networked match")
	}
}

// R does nothing while editing: the match is already held by edit mode.
func TestNoTimeManipulationWhileEditing(t *testing.T) {
	defer func() { Editing = false; ResetTime() }()
	ResetTime()
	Editing = true
	UpdateTime(true, 0, 1.0/60)
	if TimeMode != TimeOff {
		t.Errorf("R while editing started time manipulation (mode %d)", TimeMode)
	}
}

// Clicks made while editing must not fire as an attack when edit mode is left.
func TestLeavingEditModeDropsLatchedInput(t *testing.T) {
	defer func() { Editing = false; pressedAttack, pressedJump = false, false }()
	Editing = true
	pressedAttack, pressedJump = true, true
	toggleEditing(true, false)
	if pressedAttack || pressedJump {
		t.Errorf("leaving edit mode kept latched input: attack %t jump %t", pressedAttack, pressedJump)
	}
}

// Ctrl+S goes through SaveLevel: it writes the level and clears the unsaved mark.
func TestSaveLevelWritesAndClearsDirty(t *testing.T) {
	le := NewLevelEditor(39, 83)
	le.LoadLevel("level_1.json")
	le.dirty = true
	path := filepath.Join(t.TempDir(), "level.json")
	le.SaveLevel(path)
	b, err := os.ReadFile(path)
	if err != nil || len(b) == 0 {
		t.Fatalf("saving wrote nothing to %s: %v", path, err)
	}
	if le.dirty {
		t.Errorf("the level is still marked unsaved after saving")
	}
}
