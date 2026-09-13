package resistor

import (
	"strings"
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func TestEnergyStartsFull(t *testing.T) {
	w := NewWorld(true)
	if a := w.Spawn(rl.Vector2{}, true); a.Energy != MaxEnergy {
		t.Errorf("a new fighter has %v energy, want %v", a.Energy, int32(MaxEnergy))
	}
}

// The pool is clamped at both ends: a swing cannot take it below zero and a
// restore cannot push it past the cap.
func TestEnergyClamps(t *testing.T) {
	w := NewWorld(true)
	a := w.Spawn(rl.Vector2{}, true)

	a.Energy = 10
	a.spendEnergy()
	if a.Energy != 0 {
		t.Errorf("spending %d from 10 left %v, want 0", EnergySwing, a.Energy)
	}

	a.Energy = MaxEnergy
	a.restoreEnergy(EnergyBlockRestore)
	if a.Energy != MaxEnergy {
		t.Errorf("restoring at full overfilled the pool to %v", a.Energy)
	}

	a.Energy = MaxEnergy - 1
	a.restoreEnergy(EnergyBlockRestore)
	if a.Energy != MaxEnergy {
		t.Errorf("restoring one below full gave %v, want %v", a.Energy, int32(MaxEnergy))
	}
}

// Spending restarts the calm wait; restoring leaves it alone, so a block pays
// on top of refilling that is already under way.
func TestEnergySpendResetsCalm(t *testing.T) {
	w := NewWorld(true)
	a := w.Spawn(rl.Vector2{}, true)

	a.EnergyCalm = EnergyCalmTicks
	a.restoreEnergy(EnergyBlockRestore)
	if a.EnergyCalm != EnergyCalmTicks {
		t.Errorf("restoring reset the calm wait to %d", a.EnergyCalm)
	}

	a.spendEnergy()
	if a.EnergyCalm != 0 {
		t.Errorf("after a swing the calm wait is %d, want 0", a.EnergyCalm)
	}
}

// A swing costs the same wherever it starts and however it ends, and a fighter
// that cannot pay for one does not swing at all.
func TestSwingCostsEnergy(t *testing.T) {
	w := NewWorld(true)
	a := w.Spawn(rl.Vector2{}, true)
	a.Updown = 1

	for i := 1; i <= MaxEnergy/EnergySwing; i++ {
		a.State = StateIdle
		if !a.enterAttackState() {
			t.Fatalf("swing %d was refused at %v energy", i, a.Energy)
		}
		if want := int32(MaxEnergy - i*EnergySwing); a.Energy != want {
			t.Errorf("after swing %d: %v energy, want %v", i, a.Energy, want)
		}
	}

	if a.canSwing() {
		t.Errorf("another swing is affordable at %v energy, want the pool short of the %d it costs", a.Energy, EnergySwing)
	}

	// nothing left: the swing does not happen and nothing is taken
	a.Energy = 0
	a.State, a.Events = StateIdle, 0
	anim := a.Anim
	if a.enterAttackState() {
		t.Error("a swing came out at 0 energy")
	}
	if a.Energy != 0 || a.State != StateIdle || a.Anim != anim {
		t.Errorf("a refused swing changed the fighter: %v energy, state %d", a.Energy, a.State)
	}
	if a.Events&EvNoEnergy == 0 {
		t.Error("a refused swing raised no event")
	}

	// part of a swing is not enough either
	a.Energy, a.Events = EnergySwing-4, 0
	if a.enterAttackState() || a.Energy != EnergySwing-4 {
		t.Errorf("a swing came out at %v energy, below the %d it costs", a.Energy, EnergySwing)
	}
}

// A held attack chains until the pool runs out, and the swings it does throw
// cost exactly one swing each.
func TestChainedSwingsCostEnergy(t *testing.T) {
	w := NewWorld(true)
	a := w.Spawn(rl.Vector2{}, true)
	a.Updown = 1
	a.Input.Attack = true

	run(w, 1) // the first swing
	if a.State != StateAttack {
		t.Fatalf("the first swing did not start: state %d", a.State)
	}
	if a.Energy != MaxEnergy-EnergySwing {
		t.Fatalf("the first swing cost %v, want %d", MaxEnergy-a.Energy, EnergySwing)
	}

	// hold attack: the chain takes the pool down one swing at a time. Watch only
	// until it runs dry - the calm wait then refills it and the held button
	// swings again, which is right but is the regen rule, not this one.
	swings := 1
	for tick := 0; tick < 800 && a.canSwing(); tick++ {
		before := a.Energy
		run(w, 1)
		switch d := before - a.Energy; {
		case d == 0:
		case d == EnergySwing:
			swings++
		default:
			t.Fatalf("the pool moved by %v in one tick, want either nothing or the %d a swing costs", d, EnergySwing)
		}
	}
	if want := int32(MaxEnergy % EnergySwing); a.Energy != want {
		t.Errorf("the chain left %v energy, want %v - a whole pool spent in swings", a.Energy, want)
	}
	if want := MaxEnergy / EnergySwing; swings != want {
		t.Errorf("the chain threw %d swings from a full pool, want %d", swings, want)
	}
	// the last swing is still in flight when the pool runs dry; let it finish and
	// the chain should end there, well inside the calm wait before any refill
	run(w, 40)
	if !strings.Contains(a.Anim, "back") && a.State == StateAttack {
		t.Errorf("the chain started another swing on an empty pool: %s", a.Anim)
	}
}

// A swing pays once, whatever it meets: a defender it overlaps for the whole
// swing, a block, or nothing at all.
func TestSwingPaysOncePerSwing(t *testing.T) {
	_, att, _ := swing(t, 1) // the defender guards at the attacker's height
	att.Energy = 60

	for range 10 { // the box sits over the defender the whole time
		att.updateAttack()
	}
	if att.Energy != 60 {
		t.Errorf("a swing in flight cost %v more energy, want nothing after it started", 60-att.Energy)
	}

	// the same swing against empty air
	w2 := NewWorld(true)
	solo := w2.Spawn(rl.Vector2{}, true)
	solo.Updown = 1
	solo.enterAttackState()
	before := solo.Energy
	for range 10 {
		solo.updateAttack()
	}
	if solo.Energy != before {
		t.Errorf("swinging at air cost %v after the swing started", before-solo.Energy)
	}
	if MaxEnergy-before != EnergySwing {
		t.Errorf("a swing at air cost %v, want the same %d as one that lands", MaxEnergy-before, EnergySwing)
	}
}

// A block at the attacker's height pays; block-all stops the attack but funds
// nothing; a hit pays nothing at all.
func TestBlockRestoresEnergy(t *testing.T) {
	_, att, def := swing(t, 1)

	def.Energy = 40
	def.Updown = att.Updown // guarding at the attacker's height
	def.confirmBlock(att)
	if def.Energy != 40+EnergyBlockRestore {
		t.Errorf("after a block: %v energy, want %v", def.Energy, int32(40+EnergyBlockRestore))
	}

	def.Energy = 40
	def.Updown = 3 // block-all
	def.confirmBlock(att)
	if def.Energy != 40 {
		t.Errorf("block-all paid %v energy, want nothing", def.Energy-40)
	}

	def.Energy = 40
	def.Updown = att.Updown
	def.takeHit(att)
	if def.Energy != 40 {
		t.Errorf("a hit changed the defender's energy by %v, want nothing", def.Energy-40)
	}

	// a block is the way back into the fight from too low to swing
	def.Energy = 10
	def.Updown = att.Updown
	def.confirmBlock(att)
	if def.Energy != 25 || !def.canSwing() {
		t.Errorf("blocking at 10 energy gave %v, want 25 and a swing available", def.Energy)
	}
}

// Energy refills once the fighter stops swinging, and swinging keeps it down.
func TestEnergyRegen(t *testing.T) {
	w := NewWorld(true)
	a := w.Spawn(rl.Vector2{}, true)
	a.Updown = 1
	a.Energy = MaxEnergy - 60
	a.EnergyCalm = 0

	run(w, EnergyCalmTicks+240) // 1.5s of waiting, then 4s of refilling
	if a.Energy != MaxEnergy {
		t.Errorf("after 5.5s of calm: %v energy, want it back at %v", a.Energy, int32(MaxEnergy))
	}

	// a swing a second never gets the 1.5s wait finished, so a full pool's worth
	// of swings empties it with nothing refilled in between
	a.Energy = MaxEnergy
	for i := range MaxEnergy / EnergySwing {
		a.State = StateIdle
		if !a.enterAttackState() {
			t.Fatalf("swing %d a second in was refused at %v energy", i+1, a.Energy)
		}
		run(w, 60)
	}
	if want := int32(MaxEnergy % EnergySwing); a.Energy != want {
		t.Errorf("a swing a second left %v energy, want %v - nothing regenerated in between", a.Energy, want)
	}
	if a.canSwing() {
		t.Errorf("the pool still affords a swing at %v energy", a.Energy)
	}
}

// Only spending holds the refill back: taking a hit and blocking leave the
// wait running.
func TestEnergyCalmOnlyResetsOnSpending(t *testing.T) {
	w := NewWorld(true)
	att := w.Spawn(rl.Vector2{}, true)
	def := w.Spawn(rl.Vector2{X: 20}, true)
	att.Orientation, att.Updown = 1, 1
	def.Energy = 40
	def.EnergyCalm = EnergyCalmTicks // already refilling

	def.takeHit(att)
	if def.EnergyCalm != EnergyCalmTicks {
		t.Errorf("a hit reset the energy wait to %d", def.EnergyCalm)
	}

	def.Updown = 3
	def.confirmBlock(att) // block-all: no restore, and no reset either
	if def.EnergyCalm != EnergyCalmTicks {
		t.Errorf("a block-all reset the energy wait to %d", def.EnergyCalm)
	}

	def.State, def.Updown = StateIdle, 1
	def.enterAttackState()
	if def.EnergyCalm != 0 {
		t.Errorf("a swing left the energy wait at %d, want it restarted", def.EnergyCalm)
	}
}
