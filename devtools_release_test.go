//go:build release

package resistor

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReleaseHasNoDevTools(t *testing.T) {
	if DevTools {
		t.Errorf("a release build has DevTools true")
	}
	if Debug {
		t.Errorf("a release build starts with the debug overlay on")
	}
}

// A release build never writes the level, however it is asked to.
func TestReleaseNeverWritesLevel(t *testing.T) {
	le := NewLevelEditor(39, 83)
	le.dirty = true
	path := filepath.Join(t.TempDir(), "level_1.json")
	le.SaveLevel(path)
	if _, err := os.Stat(path); err == nil {
		t.Errorf("a release build wrote the level to %s", path)
	}
}
