package resistor

import rl "github.com/gen2brain/raylib-go/raylib"

// A fighter a gamemode put on RedTeam or BlueTeam stands on a line of its
// team's colour, the width of its body, on the ground under its feet. It is
// drawn over the lighting with the bars, so a dark corner cannot hide whose
// side a fighter is on, and it stays clear of the red structure bar over the
// head. Every other team - no team, the bots' - has no line, so a match
// without a gamemode looks as it always has.
const TeamMarkerHeight = 1

// Exact palette entries, so the palette pass leaves them as they are.
var (
	redTeamColor  = rl.Red
	blueTeamColor = rl.SkyBlue
)

// teamColor is the colour a team is drawn in, and false for a team that is
// not drawn.
func teamColor(team uint8) (rl.Color, bool) {
	switch team {
	case RedTeam:
		return redTeamColor, true
	case BlueTeam:
		return blueTeamColor, true
	}
	return rl.Color{}, false
}

// showTeamMarker reports whether a fighter's team line is drawn: it is part of
// the HUD, and a dying fighter or a recorded frame has none.
func showTeamMarker(a *Actor) bool {
	_, ok := teamColor(a.Team)
	return ok && showHUD() && a.State != StateDeath && !TimeShowsHistory()
}

// teamMarkerRect is the line under a fighter's feet, from the position its
// sprite is drawn at, in whole world pixels.
func teamMarkerRect(a *Actor) rl.Rectangle {
	pos := a.drawPos()
	return rl.Rectangle{
		X:      pos.X + roundf(a.CollisionBox.X),
		Y:      pos.Y + roundf(a.CollisionBox.Y+a.CollisionBox.Height),
		Width:  roundf(a.CollisionBox.Width),
		Height: TeamMarkerHeight,
	}
}

func (a *Actor) drawTeamMarker() {
	if !showTeamMarker(a) {
		return
	}
	col, _ := teamColor(a.Team)
	rl.DrawRectangleRec(teamMarkerRect(a), col)
}
