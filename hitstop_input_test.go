package resistor

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// frozen puts a grounded fighter into a hit-stop freeze with a press already
// made: the press lands on the first frozen tick, which is the tick the old
// code threw it away on. It returns the fighter and the step, with the first
// frozen tick already taken.
func frozen(t *testing.T, press func(in *Input)) (*Actor, *World, float32) {
	t.Helper()
	floor()
	w := NewWorld(true)
	a := w.Spawn(rl.Vector2{}, true)
	step := float32(FixedTimeStep)

	a.FreezeTicks = HitStopTicks
	press(&a.Input)
	a.Update(step) // the frozen tick the press was made on
	a.Input = Input{}
	return a, w, step
}

// thaw runs out the rest of the freeze, checking the press is still held and
// has not acted, and returns having taken the first live tick.
func thaw(t *testing.T, a *Actor, step float32, held func(*Actor) bool) {
	t.Helper()
	for a.FreezeTicks > 0 {
		if !held(a) {
			t.Fatalf("the press was dropped with %d frozen ticks left", a.FreezeTicks)
		}
		if a.State != StateIdle {
			t.Fatalf("a frozen fighter acted: state %d", a.State)
		}
		a.Update(step)
	}
	a.Update(step) // the first live tick: the fold-back happens here
}

// The bug: a click made during hit-stop was latched, consumed by sampleInput
// and then thrown away by Update's early return, so the fastest counter got
// nothing. It is now held for the freeze and acts on the first live tick.
func TestAttackHeldThroughHitStop(t *testing.T) {
	a, _, step := frozen(t, func(in *Input) { in.Attack = true })
	thaw(t, a, step, func(a *Actor) bool { return a.BufAttack })

	if a.State != StateAttack {
		t.Errorf("the counter did not come out: state %d, want the attack %d", a.State, StateAttack)
	}
	if a.BufAttack {
		t.Errorf("the held press outlived the freeze it was held for")
	}
}

// Jump had the same hole and no AttackQueued to fall back on, so it is held by
// the same mechanism rather than a second one.
func TestJumpHeldThroughHitStop(t *testing.T) {
	a, _, step := frozen(t, func(in *Input) { in.Jump = true })
	thaw(t, a, step, func(a *Actor) bool { return a.BufJump })

	if a.State != StateJumpStart {
		t.Errorf("the held jump did not leave the ground: state %d, want %d", a.State, StateJumpStart)
	}
	if a.BufJump {
		t.Errorf("the held press outlived the freeze it was held for")
	}
}

// A landed swing freezes the attacker too, so mashing through the hit-stop of
// your own hit used to lose the chain that mashing through a whiff kept.
func TestChainPressHeldThroughHitStop(t *testing.T) {
	w, att, _ := swing(t, 0)
	step := float32(FixedTimeStep)

	att.AttackQueued = false
	att.FreezeTicks = HitStopTicks
	att.Input.Attack = true
	att.Update(step)
	att.Input = Input{}

	for att.FreezeTicks > 0 {
		if !att.BufAttack {
			t.Fatalf("the chain press was dropped with %d frozen ticks left", att.FreezeTicks)
		}
		att.Update(step)
	}
	att.Update(step) // the first live tick

	if !att.AttackQueued {
		t.Errorf("the chain press did not reach AttackQueued")
	}
	if att.BufAttack {
		t.Errorf("the held press outlived the freeze")
	}
	_ = w
}

// Holding the press for the whole freeze must not turn one click into six.
func TestOnePressIsOneActionAcrossAFreeze(t *testing.T) {
	a, _, step := frozen(t, func(in *Input) { in.Attack = true })

	swings := 0
	prev := a.State
	for range HitStopTicks + 30 {
		a.Update(step)
		if a.State == StateAttack && prev != StateAttack {
			swings++
		}
		prev = a.State
	}
	if swings != 1 {
		t.Errorf("one press produced %d swings, want exactly 1", swings)
	}
}

// A flinch does not lock a fighter out, but a stun does. Nothing in the buffer
// says so: the first live tick of a stunned fighter is in StateAttacked, whose
// arm reads no input, so the press is folded in and dropped by that tick. If
// StateAttacked ever starts reading Input this test is what catches it.
func TestStunSwallowsTheHeldPress(t *testing.T) {
	w, att, def := swing(t, 0)
	step := float32(FixedTimeStep)

	def.enterAttackedState(att) // sets the shared freeze through hitFeedback
	def.Input.Attack = true
	def.Update(step)
	def.Input = Input{}
	if !def.BufAttack {
		t.Fatalf("the press was not held during the stun's freeze")
	}

	for range HitStopTicks + 40 {
		def.Update(step)
		if def.State == StateAttack {
			t.Fatalf("the stunned fighter attacked out of a lockout")
		}
	}
	if def.BufAttack {
		t.Errorf("the press survived the stun instead of being dropped by it")
	}
	_ = w
}

// A press held on the killing blow must not cross into the fighter that
// replaces it.
func TestDeathDiscardsTheHeldPress(t *testing.T) {
	_, att, def := swing(t, 0)
	step := float32(FixedTimeStep)

	def.FreezeTicks = HitStopTicks
	def.Input.Attack = true
	def.Input.Jump = true
	def.Update(step)
	def.Input = Input{}
	if !def.BufAttack || !def.BufJump {
		t.Fatalf("the press was not held during the freeze")
	}

	def.State = StateIdle
	def.Structure = 0.5 // the next hit breaks it
	def.takeHit(att)
	if def.State != StateDeath {
		t.Fatalf("the fighter did not die: state %d", def.State)
	}
	if def.BufAttack || def.BufJump {
		t.Errorf("death kept a held press: attack=%v jump=%v", def.BufAttack, def.BufJump)
	}
}

// Energy cannot move while a fighter is frozen - regen sits below the freeze
// return and nothing spends - so a press made against a pool that could afford
// it is still affordable when it acts.
func TestHeldPressIsStillAffordable(t *testing.T) {
	a, _, step := frozen(t, func(in *Input) { in.Attack = true })
	a.Energy = EnergySwing // exactly one swing left

	for a.FreezeTicks > 0 {
		if a.Energy != EnergySwing {
			t.Fatalf("energy moved during the freeze: %d", a.Energy)
		}
		a.Update(step)
	}
	a.Update(step)

	if a.State != StateAttack {
		t.Errorf("the held press was refused: state %d, want the attack %d", a.State, StateAttack)
	}
	if a.Energy >= EnergySwing {
		t.Errorf("the swing was not paid for: %d energy left", a.Energy)
	}
}
