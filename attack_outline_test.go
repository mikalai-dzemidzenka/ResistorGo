package resistor

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// The one attack box both the swing and the outline are built from: the three
// heights at the offsets the fighter carries, mirrored when it faces left, and
// nothing at all for block-all.
func TestAttackBoxAt(t *testing.T) {
	a := NewActor(rl.Vector2{X: 100, Y: 50}, true)
	a.Orientation = 1
	pos := a.Pos

	// the offsets are spelled out rather than read back from the fighter, so
	// moving one fails here instead of quietly moving the outline with it
	for _, c := range []struct {
		updown           uint8
		offY, wide, high float32
	}{
		{0, 24, 24, 6}, // low
		{1, 14, 24, 6}, // mid
		{2, 5, 24, 6},  // high
	} {
		r, ok := a.attackBoxAt(c.updown, pos)
		if !ok {
			t.Errorf("height %d reports no attack box", c.updown)
			continue
		}
		want := rl.Rectangle{X: pos.X + 24, Y: pos.Y + c.offY, Width: c.wide, Height: c.high}
		if r != want {
			t.Errorf("height %d: box %+v, want %+v", c.updown, r, want)
		}
	}

	// the three are distinct and ordered up the body: low below mid below high
	low, _ := a.attackBoxAt(0, pos)
	mid, _ := a.attackBoxAt(1, pos)
	high, _ := a.attackBoxAt(2, pos)
	if !(high.Y < mid.Y && mid.Y < low.Y) {
		t.Errorf("the heights are not ordered up the body: high %v mid %v low %v", high.Y, mid.Y, low.Y)
	}

	if _, ok := a.attackBoxAt(3, pos); ok {
		t.Errorf("block-all reports an attack box")
	}

	// facing left drops the 24-pixel offset, putting the box over the fighter
	a.Orientation = -1
	left, _ := a.attackBoxAt(1, pos)
	if want := pos.X + a.AttackBoxMid.X; left.X != want {
		t.Errorf("facing left the box starts at %v, want %v", left.X, want)
	}

	// and it is built from the position given, not the fighter's own
	moved, _ := a.attackBoxAt(1, rl.Vector2{X: pos.X + 10, Y: pos.Y + 3})
	if moved.X != left.X+10 || moved.Y != left.Y+3 {
		t.Errorf("the box ignored the position it was given: %+v against %+v", moved, left)
	}
}

// aim puts a fighter at pos, facing right, aiming at height updown, with its
// interpolated position settled on its simulated one.
func aim(w *World, pos rl.Vector2, updown uint8) *Actor {
	a := w.Spawn(pos, true)
	a.Orientation = 1
	a.Updown = updown
	a.InterpolatedPos = pos
	return a
}

// Only a fighter the swing would reach is marked: the outline appearing means
// the swing connects from here.
func TestOutlineOnlyInRange(t *testing.T) {
	defer ResetTime()
	ResetTime()
	w := NewWorld(true)
	att := aim(w, ZeroVec2, 1)

	// the mid box reaches from x+24 for 24 pixels; the body starts 14 in
	near := aim(w, rl.Vector2{X: 30}, 1)
	if !outlineTarget(att, near) {
		t.Errorf("a fighter inside the swing's reach is not marked")
	}

	far := aim(w, rl.Vector2{X: 80}, 1)
	if outlineTarget(att, far) {
		t.Errorf("a fighter beyond the swing's reach is marked")
	}
}

// The height decides what is in reach: a fighter the high box misses is marked
// at the height that does reach it.
func TestOutlineFollowsHeight(t *testing.T) {
	defer ResetTime()
	ResetTime()
	w := NewWorld(true)
	att := aim(w, ZeroVec2, 1)
	// standing low enough that the high box passes over its body, while the
	// low box still catches it
	target := aim(w, rl.Vector2{X: 30, Y: 12}, 1)

	att.Updown = 0
	low := outlineTarget(att, target)
	att.Updown = 2
	high := outlineTarget(att, target)
	if !low || high {
		t.Errorf("standing low: marked at low %t, at high %t - want marked at low only", low, high)
	}
}

// The swing goes the way the fighter faces, so what is behind it is not marked.
func TestOutlineNotBehind(t *testing.T) {
	defer ResetTime()
	ResetTime()
	w := NewWorld(true)
	att := aim(w, rl.Vector2{X: 60}, 1)
	behind := aim(w, rl.Vector2{X: 40}, 1)
	if outlineTarget(att, behind) {
		t.Errorf("a fighter behind the aimer is marked")
	}
	att.Orientation = -1 // turn around and it is in front
	if !outlineTarget(att, behind) {
		t.Errorf("turning to face a fighter in reach did not mark it")
	}
}

// Every reason not to draw one, in one place.
func TestOutlineExclusions(t *testing.T) {
	defer ResetTime()
	ResetTime()
	w := NewWorld(true)
	att := aim(w, ZeroVec2, 1)
	target := aim(w, rl.Vector2{X: 30}, 1)
	if !outlineTarget(att, target) {
		t.Fatalf("the case under test is not marked to begin with")
	}

	att.Updown = 3 // block-all has no attack to aim with
	if outlineTarget(att, target) {
		t.Errorf("block-all marks a target")
	}
	att.Updown = 1

	target.State = StateDeath
	if outlineTarget(att, target) {
		t.Errorf("a dying fighter is marked")
	}
	target.State = StateIdle

	att.State = StateDeath
	if outlineTarget(att, target) {
		t.Errorf("a dying fighter marks a target")
	}
	att.State = StateIdle

	if outlineTarget(att, att) {
		t.Errorf("the aimer marks itself")
	}
	if outlineTarget(nil, target) {
		t.Errorf("a nil aimer - a spectator, or a player between respawns - marks a target")
	}

	rewindRecord()
	TimeMode = TimePaused
	if outlineTarget(att, target) {
		t.Errorf("a recorded frame is marked")
	}
	TimeMode = TimeOff
	if !outlineTarget(att, target) {
		t.Errorf("the outline did not come back when time manipulation was left")
	}
}

// A swing reaching two fighters at once marks both.
func TestOutlineMarksEveryFighterInRange(t *testing.T) {
	defer ResetTime()
	ResetTime()
	w := NewWorld(true)
	att := aim(w, ZeroVec2, 1)
	first := aim(w, rl.Vector2{X: 26}, 1)
	second := aim(w, rl.Vector2{X: 32}, 1)
	if !outlineTarget(att, first) || !outlineTarget(att, second) {
		t.Errorf("two fighters in reach: first marked %t, second marked %t - want both", outlineTarget(att, first), outlineTarget(att, second))
	}
}

// The band is the slice of the marked fighter the swing would take: at the
// height being aimed, as wide as that fighter's body and no wider.
func TestOutlineBandLayout(t *testing.T) {
	defer ResetTime()
	ResetTime()
	w := NewWorld(true)
	att := aim(w, ZeroVec2, 1)
	target := aim(w, rl.Vector2{X: 30}, 1)

	var bands [3]rl.Rectangle
	for h := uint8(0); h < 3; h++ {
		att.Updown = h
		bands[h] = attackOutlineBand(att, target)
	}
	if !(bands[2].Y < bands[1].Y && bands[1].Y < bands[0].Y) {
		t.Errorf("the band is not ordered up the body: high %v mid %v low %v", bands[2].Y, bands[1].Y, bands[0].Y)
	}

	att.Updown = 1
	band := attackOutlineBand(att, target)
	// the band covers the marked fighter's sprite frame: the shader clips it to
	// the body's own pixels from there, so the band decides the rows and the
	// sprite decides the shape
	if band.X != target.drawPos().X || band.Width != spriteFrame {
		t.Errorf("the band spans %v..%v, want the sprite frame %v..%v", band.X, band.X+band.Width, target.drawPos().X, target.drawPos().X+spriteFrame)
	}
	if body := target.outlineRect(); band.X > body.X || band.X+band.Width < body.X+body.Width {
		t.Errorf("the band %v..%v does not cover the body %v..%v", band.X, band.X+band.Width, body.X, body.X+body.Width)
	}
	if box, _ := att.attackBoxAt(1, att.drawPos()); band.Height != box.Height {
		t.Errorf("the band is %v tall, want the swing's %v", band.Height, box.Height)
	}
	for _, v := range []float32{band.X, band.Y, band.Width, band.Height} {
		if v != roundf(v) {
			t.Errorf("the band has an edge off the pixel grid: %+v", band)
			break
		}
	}
}

// The band is laid out from the position the sprite is drawn at, so it cannot
// lag the fighter it marks.
func TestOutlineBandFollowsTheSprite(t *testing.T) {
	defer ResetTime()
	ResetTime()
	w := NewWorld(true)
	att := aim(w, ZeroVec2, 1)
	target := aim(w, rl.Vector2{X: 30}, 1)

	settled := attackOutlineBand(att, target)
	target.InterpolatedPos.X = target.Pos.X + 6 // mid-interpolation, ahead of the simulation
	moved := attackOutlineBand(att, target)
	if moved.X != settled.X+6 {
		t.Errorf("the band sat at %v while the sprite was drawn 6 pixels on at %v", moved.X, settled.X+6)
	}
}

// The outline says nothing about the defence: the marked fighter's guard,
// structure, energy and stun progress leave it untouched.
func TestOutlineIgnoresTheDefence(t *testing.T) {
	defer ResetTime()
	ResetTime()
	w := NewWorld(true)
	att := aim(w, ZeroVec2, 1)
	target := aim(w, rl.Vector2{X: 30}, 1)
	want := attackOutlineBand(att, target)

	for _, guard := range []uint8{0, 1, 2, 3} { // including the guard that would block it
		target.Updown = guard
		if !outlineTarget(att, target) {
			t.Errorf("guard %d stopped the fighter being marked", guard)
		}
		if got := attackOutlineBand(att, target); got != want {
			t.Errorf("guard %d moved the band to %+v, want %+v", guard, got, want)
		}
	}

	target.Structure = 1
	target.Energy = 0
	target.HitCount = StunEveryHits - 1
	if !outlineTarget(att, target) {
		t.Errorf("a hurt fighter is not marked")
	}
	if got := attackOutlineBand(att, target); got != want {
		t.Errorf("a hurt fighter's band is %+v, want %+v", got, want)
	}
}
