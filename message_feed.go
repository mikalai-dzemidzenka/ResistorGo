package resistor

import (
	"strings"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// The message feed is the lines of text a server sends - a gamemode's scripts
// telling the players what is going on - shown over the level for a while. It
// is part of the match, not of the HUD over the fighters, so hiding the HUD
// leaves it on screen.
const (
	FeedLines    = 6
	FeedDuration = 10 * time.Second
	feedFont     = 20
	feedLineH    = 24
)

type feedLine struct {
	text  string
	until time.Time
}

type messageFeed struct {
	lines []feedLine // oldest first
}

// add puts a line at the bottom of the feed, pushing the oldest out past
// FeedLines.
func (f *messageFeed) add(text string, now time.Time) {
	f.lines = append(f.lines, feedLine{feedText(text), now.Add(FeedDuration)})
	if n := len(f.lines) - FeedLines; n > 0 {
		f.lines = append(f.lines[:0], f.lines[n:]...)
	}
}

// shown is the lines still on screen at now, oldest first, and it lets the
// ones that are not go.
func (f *messageFeed) shown(now time.Time) []string {
	kept := f.lines[:0]
	for _, l := range f.lines {
		if now.Before(l.until) {
			kept = append(kept, l)
		}
	}
	f.lines = kept
	out := make([]string, len(kept))
	for i, l := range kept {
		out[i] = l.text
	}
	return out
}

// feedText is text as raylib's default font can draw it: printable ASCII,
// anything else a question mark.
func feedText(s string) string {
	return strings.Map(func(r rune) rune {
		if r < ' ' || r > '~' {
			return '?'
		}
		return r
	}, s)
}

// draw draws the feed in window pixels at the lower left, above where the
// spectator HUD's two lines go, newest at the bottom.
func (f *messageFeed) draw(now time.Time) {
	lines := f.shown(now)
	const x = 20
	bottom := int32(SH*WindowScale - 100)
	for i, text := range lines {
		y := bottom - int32(len(lines)-1-i)*feedLineH
		rl.DrawText(text, x+2, y+2, feedFont, rl.Black)
		rl.DrawText(text, x, y, feedFont, rl.White)
	}
}
