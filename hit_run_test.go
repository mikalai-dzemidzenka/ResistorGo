package resistor

import "testing"

// The cost climbs half a point per hit already in the run, and stops at the
// ceiling: 1, 1.5, 2, 2.5, 3, then 3 for as long as the run goes on.
func TestHitCostClimbsAndCaps(t *testing.T) {
	_, att, def := swing(t, 0)
	for i, want := range []float32{1, 1.5, 2, 2.5, 3, 3, 3} {
		if got := def.hitCost(); got != want {
			t.Errorf("hit %d of the run costs %v, want %v", i+1, got, want)
		}
		def.HitRun = min(def.HitRun+1, HitRunCapped)
	}
	_ = att
}

// A whole run against a fresh fighter: five hits take it to 2 structure and the
// sixth ends it.
func TestARunCostsMoreAsItGoes(t *testing.T) {
	_, att, def := swing(t, 0)
	for i, want := range []float32{11, 9.5, 7.5, 5, 2} {
		def.State = StateIdle // the stun's own exit is not what is under test
		def.takeHit(att)
		if def.Structure != want {
			t.Errorf("after hit %d: %v structure, want %v", i+1, def.Structure, want)
		}
	}
	def.State = StateIdle
	def.takeHit(att)
	if def.State != StateDeath {
		t.Errorf("the sixth hit of a run left state %d, want death %d", def.State, StateDeath)
	}
}

// The stun does not relieve the pressure: the fourth hit of a run costs what a
// fourth hit costs, although the third stunned.
func TestTheRunSurvivesTheStun(t *testing.T) {
	_, att, def := swing(t, 0)
	for range 3 {
		def.State = StateIdle
		def.takeHit(att)
	}
	if def.HitCount != 0 {
		t.Fatalf("three hits left a stun count of %d, want the stun to have cleared it", def.HitCount)
	}
	before := def.Structure
	def.State = StateIdle
	def.takeHit(att)
	if cost := before - def.Structure; cost != 2.5 {
		t.Errorf("the fourth hit cost %v, want the 2.5 a fourth hit of a run costs", cost)
	}
}

// Any stopped attack breaks the run: a guard at the attacker's height, and
// block-all too.
func TestABlockBreaksTheRun(t *testing.T) {
	for _, c := range []struct {
		name  string
		guard uint8
	}{
		{"a matching guard", 1},
		{"block-all", 3},
	} {
		_, att, def := swing(t, 0)
		for range 3 {
			def.State = StateIdle
			def.takeHit(att)
		}
		if def.HitRun == 0 {
			t.Fatalf("%s: the run did not build up", c.name)
		}

		def.Updown = c.guard
		def.confirmBlock(att)
		if def.HitRun != 0 {
			t.Errorf("%s did not break the run: %d still counted", c.name, def.HitRun)
		}

		def.Updown, def.State = 0, StateIdle
		before := def.Structure
		def.takeHit(att)
		if cost := before - def.Structure; cost != HitStructure {
			t.Errorf("the hit after %s cost %v, want the base %v", c.name, cost, float32(HitStructure))
		}
	}
}

// Being left alone ends a run, on the same recovery that clears the stun count.
func TestRecoveryEndsTheRun(t *testing.T) {
	w, att, def := swing(t, 0)
	att.State = StateIdle
	for range 3 {
		def.State = StateIdle
		def.takeHit(att)
	}
	run(w, StructureCalmTicks+1)
	if def.HitRun != 0 {
		t.Errorf("recovery left a run of %d", def.HitRun)
	}

	before := def.Structure
	def.State = StateIdle
	def.takeHit(att)
	if cost := before - def.Structure; cost != HitStructure {
		t.Errorf("the hit after recovering cost %v, want the base %v", cost, float32(HitStructure))
	}
}

// The run survives the wire, is clamped to what the cost can use, and leaves
// the count it shares a word with alone.
func TestHitRunReplicates(t *testing.T) {
	w := NewWorld(true)
	a := w.Spawn(ZeroVec2, true)
	a.HitCount = StunEveryHits - 1
	a.Orientation, a.CanClimbLeft = -1, true
	for _, c := range []struct{ sent, want uint8 }{{0, 0}, {1, 1}, {HitRunCapped, HitRunCapped}, {9, HitRunCapped}} {
		a.HitRun = c.sent
		b := NewActor(ZeroVec2, true)
		recOf(a).apply(b, w)
		if b.HitRun != c.want {
			t.Errorf("a run of %d arrived as %d, want %d", c.sent, b.HitRun, c.want)
		}
		if b.HitCount != StunEveryHits-1 {
			t.Errorf("a run of %d disturbed the stun count: %d", c.sent, b.HitCount)
		}
		if b.Orientation != -1 || !b.CanClimbLeft {
			t.Errorf("a run of %d disturbed the flags it shares a word with", c.sent)
		}
	}
}

// A broken run replicates as well as a growing one.
func TestABrokenRunReplicates(t *testing.T) {
	w, att, def := swing(t, 0)
	sent := func() uint8 {
		b := NewActor(ZeroVec2, true)
		recOf(def).apply(b, w)
		return b.HitRun
	}
	for range 3 {
		def.State = StateIdle
		def.takeHit(att)
	}
	if n := sent(); n != 3 {
		t.Errorf("three hits in a row replicate as a run of %d, want 3", n)
	}
	def.Updown = att.Updown
	def.confirmBlock(att)
	if n := sent(); n != 0 {
		t.Errorf("after a block the run replicates as %d, want 0", n)
	}
}
