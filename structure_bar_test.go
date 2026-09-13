package resistor

import (
	"slices"
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// The hit count survives the wire, is clamped to what a stun leaves room for,
// and does not disturb the flags it shares a word with.
func TestHitCountReplicates(t *testing.T) {
	w := NewWorld(true)
	a := w.Spawn(ZeroVec2, true)
	a.Orientation, a.CanClimbLeft, a.AnimWhite = -1, true, true
	for _, c := range []struct{ sent, want uint8 }{{0, 0}, {1, 1}, {2, 2}, {5, StunEveryHits - 1}} {
		a.HitCount = c.sent
		b := NewActor(ZeroVec2, true)
		recOf(a).apply(b, w)
		if b.HitCount != c.want {
			t.Errorf("hit count %d arrived as %d, want %d", c.sent, b.HitCount, c.want)
		}
		if b.Orientation != -1 || !b.CanClimbLeft || !b.AnimWhite {
			t.Errorf("hit count %d disturbed the other flags: orientation %d climb-left %t white %t", c.sent, b.Orientation, b.CanClimbLeft, b.AnimWhite)
		}
	}
}

// "Stun progress is replicated": two hits in, the record says two; the stun
// clears it.
func TestStunProgressReplicates(t *testing.T) {
	w, att, def := swing(t, 0)
	sent := func() uint8 {
		b := NewActor(ZeroVec2, true)
		recOf(def).apply(b, w)
		return b.HitCount
	}
	def.takeHit(att)
	def.takeHit(att)
	if n := sent(); n != 2 {
		t.Errorf("after two hits the record carries %d, want 2", n)
	}
	def.takeHit(att) // the stun
	if n := sent(); n != 0 {
		t.Errorf("after the stun the record carries %d, want 0", n)
	}
}

func TestShowStructureBar(t *testing.T) {
	defer ResetTime()
	ResetTime()
	w, att, def := swing(t, 0)
	att.State = StateIdle

	if showStructureBar(def) {
		t.Errorf("a fighter at full structure has a bar")
	}
	def.takeHit(att)
	if !showStructureBar(def) {
		t.Errorf("a fighter that has lost structure has no bar")
	}

	// time manipulation showing a recorded frame hides every bar
	rewindRecord()
	TimeMode = TimePaused
	if showStructureBar(def) {
		t.Errorf("a bar is shown over a recorded frame")
	}
	TimeMode = TimeOff
	if !showStructureBar(def) {
		t.Errorf("the bar did not come back when time manipulation was left")
	}

	run(w, StructureCalmTicks+180) // comfortably longer than one hit takes to refill
	if def.Structure != MaxStructure {
		t.Fatalf("regeneration left %v structure, want full", def.Structure)
	}
	if showStructureBar(def) {
		t.Errorf("the bar is still shown after regeneration refilled the pool")
	}

	def.Structure = HitStructure
	def.takeHit(att)
	if def.State != StateDeath {
		t.Fatalf("the killing hit left state %d, want death", def.State)
	}
	if showStructureBar(def) {
		t.Errorf("a dying fighter has a bar")
	}
}

// fillWidth is the total filled width of a layout.
func fillWidth(r structureBarRects) float32 { return r.Fill[0].Width + r.Fill[1].Width }

func TestStructureBarLayout(t *testing.T) {
	a := NewActor(ZeroVec2, true)
	a.InterpolatedPos = rl.Vector2{X: 10.4, Y: -3.6} // interpolation leaves fractions

	a.Structure = 9
	r := structureBarLayout(a)
	if w := fillWidth(r); w != StructureBarWidth/4 {
		t.Errorf("at 9 of 12 structure the fill is %v px, want a quarter of %d", w, StructureBarWidth)
	}
	mid := r.Track.X + r.Track.Width/2
	if r.Fill[0].X+r.Fill[0].Width != mid || r.Fill[1].X != mid || r.Fill[0].Width != r.Fill[1].Width {
		t.Errorf("the fill %+v is not symmetric about the middle of the bar at %v", r.Fill, mid)
	}

	a.Structure = 1
	if w := fillWidth(structureBarLayout(a)); w < StructureBarWidth-4 || w >= StructureBarWidth {
		t.Errorf("one hit from breaking the fill is %v of %d px, want nearly full", w, StructureBarWidth)
	}

	a.Structure = 8
	before := fillWidth(structureBarLayout(a))
	a.Structure = 8 + BlockStructure
	if after := fillWidth(structureBarLayout(a)); after >= before {
		t.Errorf("a block from 8 to 8.5 left the fill at %v px, want less than %v", after, before)
	}

	a.Structure = 7.3
	a.HitCount = 5 // more than a stun leaves room for
	r = structureBarLayout(a)
	if len(r.Pips) != StunEveryHits-1 {
		t.Errorf("%d pips drawn for a hit count of %d, want it clamped to %d", len(r.Pips), a.HitCount, StunEveryHits-1)
	}

	// pips sit in a row centred above the bar and never stick out past its
	// ends, so at arm's length they cannot land on the other fighter's bar
	for _, hits := range []uint8{1, 2} {
		a.HitCount = hits
		p := structureBarLayout(a)
		for _, q := range p.Pips {
			if q.Y+q.Height > p.Track.Y {
				t.Errorf("%d pips: pip %+v is not above the bar at y=%v", hits, q, p.Track.Y)
			}
			if q.X < p.Track.X || q.X+q.Width > p.Track.X+p.Track.Width {
				t.Errorf("%d pips: pip %+v sticks out past the bar %+v", hits, q, p.Track)
			}
		}
		first, last := p.Pips[0], p.Pips[len(p.Pips)-1]
		if left, right := first.X-p.Track.X, p.Track.X+p.Track.Width-(last.X+last.Width); left != right {
			t.Errorf("%d pips are not centred over the bar: %v px of bar to their left, %v to their right", hits, left, right)
		}
	}
	a.HitCount = 5
	rects := append([]rl.Rectangle{r.Track, r.Fill[0], r.Fill[1]}, r.Pips...)
	for _, q := range rects {
		for _, v := range []float32{q.X, q.Y, q.Width, q.Height} {
			if v != roundf(v) {
				t.Errorf("rectangle %+v is off the pixel grid", q)
				break
			}
		}
	}
	if r.Track.Y+r.Track.Height > a.drawPos().Y+a.CollisionBox.Y {
		t.Errorf("the bar at y=%v overlaps the fighter, whose head is at y=%v", r.Track.Y, a.drawPos().Y+a.CollisionBox.Y)
	}
}

// The bar's colours must survive the palette pass unchanged.
func TestStructureBarColoursInPalette(t *testing.T) {
	for name, c := range map[string]rl.Color{"track": structureBarTrack, "fill": structureBarFill, "pip": stunPipColor, "energy fill": energyBarFill} {
		if !slices.Contains(Palette, c) {
			t.Errorf("the %s colour %v is not a palette entry", name, c)
		}
	}
}

// The energy bar: what is left, under the structure bar, same width, on the
// pixel grid, and never past either end.
func TestEnergyBarLayout(t *testing.T) {
	a := NewActor(ZeroVec2, true)
	a.InterpolatedPos = rl.Vector2{X: 10.4, Y: -3.6} // interpolation leaves fractions

	a.Energy = MaxEnergy
	r := structureBarLayout(a)
	if r.EnergyFill.Width != StructureBarWidth {
		t.Errorf("at full energy the fill is %v px, want the whole %d", r.EnergyFill.Width, StructureBarWidth)
	}
	if r.Energy.Width != r.Track.Width || r.Energy.X != r.Track.X {
		t.Errorf("the energy bar %+v does not line up with the structure bar %+v", r.Energy, r.Track)
	}
	if r.Energy.Y < r.Track.Y+r.Track.Height {
		t.Errorf("the energy bar at y=%v is not below the structure bar ending at y=%v", r.Energy.Y, r.Track.Y+r.Track.Height)
	}

	a.Energy = MaxEnergy / 2
	if w := structureBarLayout(a).EnergyFill.Width; w != StructureBarWidth/2 {
		t.Errorf("at half energy the fill is %v px, want %d", w, StructureBarWidth/2)
	}

	a.Energy = 0
	if w := structureBarLayout(a).EnergyFill.Width; w != 0 {
		t.Errorf("at no energy the fill is %v px, want nothing", w)
	}

	// spending drains it and blocking grows it back
	a.Energy = 40
	before := structureBarLayout(a).EnergyFill.Width
	a.Energy = 40 + EnergyBlockRestore
	if after := structureBarLayout(a).EnergyFill.Width; after <= before {
		t.Errorf("a block left the fill at %v px, want more than %v", after, before)
	}

	// every edge on the pixel grid, and never past either end of the bar
	for e := int32(0); e <= MaxEnergy; e++ {
		a.Energy = e
		p := structureBarLayout(a)
		if p.EnergyFill.X < p.Energy.X || p.EnergyFill.X+p.EnergyFill.Width > p.Energy.X+p.Energy.Width {
			t.Fatalf("at %d energy the fill %+v sticks out past the track %+v", e, p.EnergyFill, p.Energy)
		}
		for _, q := range []rl.Rectangle{p.Energy, p.EnergyFill} {
			for _, v := range []float32{q.X, q.Y, q.Width, q.Height} {
				if v != roundf(v) {
					t.Fatalf("at %d energy the rectangle %+v is off the pixel grid", e, q)
				}
			}
		}
	}
}

// The energy bar follows its own visibility rule, not the structure bar's.
func TestShowEnergyBar(t *testing.T) {
	a := NewActor(ZeroVec2, true)

	if showEnergyBar(a) {
		t.Error("a fighter at full energy shows an energy bar")
	}
	a.Energy = MaxEnergy - EnergySwing
	if !showEnergyBar(a) {
		t.Error("a fighter that has swung shows no energy bar")
	}
	if showStructureBar(a) {
		t.Error("the structure bar appeared for a fighter that only spent energy")
	}

	a.Energy, a.Structure = MaxEnergy, MaxStructure-1
	if showEnergyBar(a) {
		t.Error("a hurt fighter at full energy shows an energy bar")
	}
	if !showStructureBar(a) {
		t.Error("a hurt fighter shows no structure bar")
	}

	a.Energy = 0
	a.State = StateDeath
	if showEnergyBar(a) {
		t.Error("a dying fighter shows an energy bar")
	}
	a.State = StateIdle

	rewindRecord() // a recorded frame to show
	TimeMode = TimePaused
	defer func() { TimeMode = TimeOff }()
	if showEnergyBar(a) {
		t.Error("an energy bar is drawn over a recorded frame")
	}
	TimeMode = TimeOff
	if !showEnergyBar(a) {
		t.Error("the energy bar did not come back when time manipulation was left")
	}
}
