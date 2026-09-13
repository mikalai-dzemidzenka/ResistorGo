package resistor

import rl "github.com/gen2brain/raylib-go/raylib"

// The structure bar is posture-style: empty at full structure, filling from the
// middle outward as structure is lost. Sizes are in world pixels, so the bar is
// part of the scene and zooms with it.
const (
	StructureBarWidth  = 16
	StructureBarHeight = 2
	StructureBarLift   = 3 // gap between the bar and the top of the collision box
	StunPipSize        = 2
	StunPipGap         = 2 // between pips; even, so a row of them centres on a whole pixel
	StunPipLift        = 1 // gap between the pips and the bar below them

	// The energy bar sits under the structure bar, the same width. It shows
	// what the fighter has left where the bar above shows what it has lost, so
	// what a swing costs can be retuned without touching the HUD.
	EnergyBarHeight = 1
	EnergyBarDrop   = 1   // gap between the structure bar and the energy bar below it
	EnergyDenySec   = 0.2 // how long a refused swing flashes the empty bar
)

// Exact palette entries: the palette pass maps every pixel to its nearest
// entry, so these come through it unchanged.
var (
	structureBarTrack = rl.NewColor(39, 35, 36, 255)
	structureBarFill  = rl.NewColor(234, 50, 60, 255)
	stunPipColor      = rl.NewColor(242, 221, 161, 255)
	energyBarFill     = rl.NewColor(0, 228, 48, 255) // green, so it is never read as part of the red bar above
)

// showStructureBar reports whether a fighter's bar is drawn: only while it has
// lost structure, never while it is dying, and never over a recorded frame,
// which does not hold structure.
func showStructureBar(a *Actor) bool {
	return a.Structure < MaxStructure && a.State != StateDeath && !TimeShowsHistory()
}

// showEnergyBar reports whether a fighter's energy bar is drawn. Energy has its
// own rule rather than the structure bar's: a fighter that has swung shows one
// at full structure, and one at full energy shows none however hurt it is.
func showEnergyBar(a *Actor) bool {
	return a.Energy < MaxEnergy && a.State != StateDeath && !TimeShowsHistory()
}

// structureBarRects is the bar laid out in whole world pixels.
type structureBarRects struct {
	Track  rl.Rectangle
	Fill   [2]rl.Rectangle // the halves either side of the middle
	Pips   []rl.Rectangle
	Energy rl.Rectangle // the energy track, under the bar
	// EnergyFill is the energy left, drained from the right as it is spent.
	EnergyFill rl.Rectangle
}

// structureBarLayout places the bar over a fighter's head, from the position
// its sprite is drawn at. Every edge lands on a whole world pixel, so the
// camera's whole-number zoom keeps it on the pixel grid.
func structureBarLayout(a *Actor) structureBarRects {
	pos := a.drawPos()
	cx := pos.X + roundf(a.CollisionBox.X+a.CollisionBox.Width/2)
	top := pos.Y + roundf(a.CollisionBox.Y) - StructureBarLift - StructureBarHeight
	left := cx - StructureBarWidth/2

	r := structureBarRects{Track: rl.Rectangle{X: left, Y: top, Width: StructureBarWidth, Height: StructureBarHeight}}
	lost := clamp(MaxStructure-a.Structure, 0, MaxStructure)
	half := roundf(StructureBarWidth / 2 * lost / MaxStructure)
	r.Fill[0] = rl.Rectangle{X: cx - half, Y: top, Width: half, Height: StructureBarHeight}
	r.Fill[1] = rl.Rectangle{X: cx, Y: top, Width: half, Height: StructureBarHeight}

	// the pips: a row centred over the bar and never wider than it, so at arm's
	// length they cannot land on the other fighter's bar
	n := int(min(a.HitCount, StunEveryHits-1))
	row := float32(n*StunPipSize + max(n-1, 0)*StunPipGap)
	x := cx - row/2
	y := top - StunPipLift - StunPipSize
	for range n {
		r.Pips = append(r.Pips, rl.Rectangle{X: x, Y: y, Width: StunPipSize, Height: StunPipSize})
		x += StunPipSize + StunPipGap
	}

	// the energy bar, under the structure bar: what is left, not what is spent
	ey := top + StructureBarHeight + EnergyBarDrop
	r.Energy = rl.Rectangle{X: left, Y: ey, Width: StructureBarWidth, Height: EnergyBarHeight}
	held := clamp(a.Energy, 0, MaxEnergy)
	r.EnergyFill = r.Energy
	r.EnergyFill.Width = roundf(StructureBarWidth * float32(held) / MaxEnergy)
	return r
}

// drawStructureBar draws the structure bar, its stun pips and the energy bar,
// each when there is anything to show. The two bars appear independently.
func (a *Actor) drawStructureBar() {
	bar, energy := showStructureBar(a), showEnergyBar(a)
	denied := a.DenyUntil > GameTimeSec && a.State != StateDeath && !TimeShowsHistory()
	if !bar && !energy && !denied {
		return
	}
	r := structureBarLayout(a)
	if bar {
		rl.DrawRectangleRec(r.Track, structureBarTrack)
		for _, f := range r.Fill {
			if f.Width > 0 {
				rl.DrawRectangleRec(f, structureBarFill)
			}
		}
		for _, p := range r.Pips {
			rl.DrawRectangleRec(p, stunPipColor)
		}
	}
	if energy || denied { // a refusal shows the empty bar it could not pay from
		rl.DrawRectangleRec(r.Energy, structureBarTrack)
		if r.EnergyFill.Width > 0 {
			rl.DrawRectangleRec(r.EnergyFill, energyBarFill)
		}
	}
}
