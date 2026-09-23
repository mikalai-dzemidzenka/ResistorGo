package resistor

import (
	"fmt"
	"slices"
	"testing"
	"time"
)

// The feed shows the last six lines, each for ten seconds, and draws what its
// font cannot as a question mark.
func TestMessageFeed(t *testing.T) {
	var f messageFeed
	now := time.Unix(1000, 0)
	for i := range 8 {
		f.add(fmt.Sprintf("line %d", i), now.Add(time.Duration(i)*time.Millisecond))
	}
	want := []string{"line 2", "line 3", "line 4", "line 5", "line 6", "line 7"}
	if got := f.shown(now.Add(time.Second)); !slices.Equal(got, want) {
		t.Errorf("shown %q, want %q", got, want)
	}
	if got := f.shown(now.Add(FeedDuration + 4*time.Millisecond)); !slices.Equal(got, want[3:]) {
		t.Errorf("just past ten seconds of the first lines: %q, want %q", got, want[3:])
	}
	if got := f.shown(now.Add(FeedDuration + time.Second)); len(got) != 0 {
		t.Errorf("after ten seconds: %q", got)
	}
	f.add("drei Straßen\tweiter", now)
	if got := f.shown(now); got[0] != "drei Stra?en?weiter" {
		t.Errorf("drawn as %q", got[0])
	}
}
