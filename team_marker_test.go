package resistor

import (
	"slices"
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// Only a gamemode's two sides are drawn, each in its own palette colour; a
// human with no team and a bot on the bots' team are not.
func TestTeamColor(t *testing.T) {
	for _, c := range []struct {
		team uint8
		want rl.Color
		ok   bool
	}{
		{NoTeam, rl.Color{}, false},
		{BotTeam, rl.Color{}, false},
		{RedTeam, redTeamColor, true},
		{BlueTeam, blueTeamColor, true},
		{4, rl.Color{}, false},
	} {
		if col, ok := teamColor(c.team); col != c.want || ok != c.ok {
			t.Errorf("team %d: %v %t, want %v %t", c.team, col, ok, c.want, c.ok)
		}
	}
	if redTeamColor == blueTeamColor {
		t.Error("red and blue are drawn alike")
	}
	for name, c := range map[string]rl.Color{"red": redTeamColor, "blue": blueTeamColor} {
		if !slices.Contains(Palette, c) {
			t.Errorf("the %s team colour %v is not a palette entry", name, c)
		}
	}
}

// The line shows for a fighter on a side while it lives, and goes with the rest
// of the HUD and over a recorded frame.
func TestShowTeamMarker(t *testing.T) {
	defer func() { Cfg.HideHUD = false }()
	a := NewActor(ZeroVec2, true)
	if showTeamMarker(a) {
		t.Error("a human with no team shows a team line")
	}
	a.Team = BotTeam
	if showTeamMarker(a) {
		t.Error("a fighter on the bots' team shows a team line")
	}
	a.Team = BlueTeam
	if !showTeamMarker(a) {
		t.Fatal("a fighter on blue shows no team line")
	}

	a.State = StateDeath
	if showTeamMarker(a) {
		t.Error("a dying fighter shows a team line")
	}
	a.State = StateIdle

	Cfg.HideHUD = true
	if showTeamMarker(a) {
		t.Error("the team line is drawn with the HUD turned off")
	}
	Cfg.HideHUD = false

	rewindRecord()
	TimeMode = TimePaused
	defer func() { TimeMode = TimeOff }()
	if showTeamMarker(a) {
		t.Error("a team line is drawn over a recorded frame")
	}
	TimeMode = TimeOff
	if !showTeamMarker(a) {
		t.Error("the team line did not come back when time manipulation was left")
	}
}

// The line is the body's width, on the ground right under the feet, on whole
// pixels even when the fighter is drawn between them.
func TestTeamMarkerRect(t *testing.T) {
	a := NewActor(ZeroVec2, true)
	a.InterpolatedPos = rl.Vector2{X: 100.4, Y: 50.6}
	r := teamMarkerRect(a)
	pos := a.drawPos()
	body := a.CollisionBox
	want := rl.Rectangle{
		X:      pos.X + body.X,
		Y:      pos.Y + body.Y + body.Height,
		Width:  body.Width,
		Height: TeamMarkerHeight,
	}
	if r != want {
		t.Errorf("line %+v, want %+v", r, want)
	}
	for _, v := range []float32{r.X, r.Y, r.Width, r.Height} {
		if v != roundf(v) {
			t.Errorf("line %+v is off the pixel grid", r)
			break
		}
	}
}
