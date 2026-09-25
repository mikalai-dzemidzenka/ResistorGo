//go:build release

package resistor

import "testing"

func TestReleaseHasNoDevTools(t *testing.T) {
	if DevTools {
		t.Errorf("a release build has DevTools true")
	}
	if Debug {
		t.Errorf("a release build starts with the debug overlay on")
	}
}
