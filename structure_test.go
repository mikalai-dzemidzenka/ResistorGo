package resistor

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// swing builds an authoritative world with an attacker mid-swing, its hit frame
// already passed, and a defender standing inside the attack box.
func swing(t *testing.T, defUpdown uint8) (*World, *Actor, *Actor) {
	t.Helper()
	floor()
	w := NewWorld(true)
	att := w.Spawn(rl.Vector2{}, true)
	def := w.Spawn(rl.Vector2{X: 20}, true)
	att.Orientation = 1
	att.Updown = 1
	def.Updown = defUpdown
	att.State = StateAttack
	att.Ticks = 100
	att.SetAnim("resources/threefour/attack-mid.png", AnimOptions{animDuration: 0.3, totalFrames: 8, once: true})
	att.Ticks = 114 // past the hit frame
	att.advanceAnim()
	return w, att, def
}

// run advances every fighter in the world for n ticks, ignoring hit-stop.
func run(w *World, n int) {
	for range n {
		w.Tick++
		for _, a := range append([]*Actor{}, w.Actors...) {
			a.FreezeTicks = 0
			a.Update(float32(FixedTimeStep))
		}
		w.updateRespawns()
	}
}

func TestStructureStartsFull(t *testing.T) {
	w := NewWorld(true)
	if a := w.Spawn(rl.Vector2{}, true); a.Structure != MaxStructure {
		t.Errorf("a new fighter has %v structure, want %v", a.Structure, float32(MaxStructure))
	}
}

// The two places structure moves: a hit spends it, a block feeds it, block-all
// spends less than a hit would have.
func TestStructureCosts(t *testing.T) {
	_, att, def := swing(t, 0)

	def.takeHit(att)
	if want := float32(MaxStructure - HitStructure); def.Structure != want {
		t.Errorf("after a hit: %v structure, want %v", def.Structure, want)
	}

	def.Structure = 8
	def.Updown = att.Updown // guarding at the attacker's height
	def.confirmBlock(att)
	if def.Structure != 8+BlockStructure {
		t.Errorf("after a block: %v structure, want %v", def.Structure, float32(8+BlockStructure))
	}

	def.Structure = 8
	def.Updown = 3 // block-all
	def.confirmBlock(att)
	if def.Structure != 8-BlockAllStructure {
		t.Errorf("after a block-all: %v structure, want %v", def.Structure, float32(8-BlockAllStructure))
	}
	if def.Structure <= 8-HitStructure {
		t.Errorf("block-all cost %v, which is not less than a hit's %v", 8-def.Structure, float32(HitStructure))
	}

	def.Structure = MaxStructure
	def.Updown = att.Updown
	def.confirmBlock(att)
	if def.Structure != MaxStructure {
		t.Errorf("blocking at full structure overfilled it to %v", def.Structure)
	}
}

func TestStructureRegen(t *testing.T) {
	w, att, def := swing(t, 0)
	att.State = StateIdle // stop the attacker resolving anything else
	def.Structure = MaxStructure - 4
	def.CalmTicks = 0

	// the pool is a float32 fed a fraction per tick, so a total reached part way
	// through a refill carries the accumulated rounding; compare to a tolerance
	// finer than anything the bar can show
	nearly := func(got, want float32) bool { return got-want < 0.01 && want-got < 0.01 }

	run(w, 90+240) // 1.5s of waiting, then 4s of refilling
	if !nearly(def.Structure, MaxStructure) {
		t.Errorf("after 5.5s of calm: %v structure, want it back at %v", def.Structure, float32(MaxStructure))
	}

	// half way there: 3.5s of calm once refilled all 4, and now returns 2. This
	// is what tells the rate apart - waiting long enough passes at either rate.
	def.Structure = MaxStructure - 4
	def.CalmTicks = 0
	run(w, 90+120) // 1.5s of waiting, then 2s of refilling
	if want := float32(MaxStructure - 2); !nearly(def.Structure, want) {
		t.Errorf("after 3.5s of calm: %v structure, want %v - half of what was lost", def.Structure, want)
	}

	// pressure keeps it down: a hit every second
	def.Structure = MaxStructure - 4
	for range 4 {
		def.takeHit(att)
		run(w, 60)
	}
	if def.Structure >= MaxStructure-4 {
		t.Errorf("structure regenerated to %v under a hit every second", def.Structure)
	}
}

// Two hits flinch, the third stuns, and the count starts again after it. The
// structure a run costs is escalating and is covered on its own; here the pool
// is topped up between hits so six of them can be counted without the defender
// dying partway through.
func TestEveryThirdHitStuns(t *testing.T) {
	_, att, def := swing(t, 0)
	for i, want := range []ActorState{StateIdle, StateIdle, StateAttacked, StateIdle, StateIdle, StateAttacked} {
		def.State = StateIdle // the stun's own exit is not what is under test
		def.Structure = MaxStructure
		def.takeHit(att)
		if def.State != want {
			t.Errorf("hit %d left state %d, want %d", i+1, def.State, want)
		}
		if def.Structure >= MaxStructure {
			t.Errorf("hit %d cost no structure", i+1)
		}
	}
}

// A stun throws the defender further than a flinch does, and a flinch leaves it
// free to act.
func TestFlinchKeepsControl(t *testing.T) {
	_, att, def := swing(t, 0)
	def.takeHit(att)
	flinch := def.Speed.X
	if def.State != StateIdle {
		t.Errorf("a flinch put the defender in state %d, want it left alone", def.State)
	}
	if def.FlinchTicks == 0 {
		t.Errorf("a flinch did not start the flicker")
	}
	def.takeHit(att)
	def.takeHit(att) // the third stuns
	if def.State != StateAttacked {
		t.Fatalf("the third hit left state %d, want the stun %d", def.State, StateAttacked)
	}
	if abs(def.Speed.X) <= abs(flinch) {
		t.Errorf("the stun threw the defender %v, no further than the flinch's %v", def.Speed.X, flinch)
	}
}

// Only the stunning hit plays the hit animation; the two before it leave the
// fighter's own animation alone and just flicker.
func TestOnlyTheStunPlaysTheHitAnimation(t *testing.T) {
	_, att, def := swing(t, 0)
	const hitAnim = "resources/threefour/hit-top.png"
	for i := range 2 {
		def.SetAnimFrame("resources/threefour/updown.png", def.Updown)
		def.takeHit(att)
		if def.Anim == hitAnim {
			t.Errorf("hit %d played the hit animation, want the fighter's own left alone", i+1)
		}
		if !def.Flickering() {
			t.Errorf("hit %d did not flicker", i+1)
		}
	}
	def.takeHit(att) // the third stuns
	if def.Anim != hitAnim {
		t.Errorf("the stun played %q, want the hit animation", def.Anim)
	}
	if !def.Flickering() {
		t.Errorf("the stun did not flicker over its opening frames")
	}
}

// A flinch leaves the fighter's own animation running: it is hit mid-stride and
// keeps striding, on the same frame it would have reached un-hit.
func TestFlinchDoesNotInterruptTheAnimation(t *testing.T) {
	w, att, def := swing(t, 0)
	def.State = StateRun
	def.SetAnim("resources/threefour/run.png", AnimOptions{animDuration: 0.4, totalFrames: 10})
	anim, startTick := def.Anim, def.AnimStartTick

	def.takeHit(att)
	if def.State != StateRun {
		t.Errorf("a flinch put the running defender in state %d, want it left running", def.State)
	}
	if def.Anim != anim || def.AnimStartTick != startTick {
		t.Errorf("a flinch restarted the animation as %q at tick %d, want %q at %d", def.Anim, def.AnimStartTick, anim, startTick)
	}
	if !def.Flickering() {
		t.Errorf("the flinched fighter is not flickering")
	}

	// and it goes on running: the animation advances while the flicker holds
	frame := def.AnimCurrentFrame
	def.FreezeTicks = 0
	run(w, 3)
	if def.Anim != anim {
		t.Errorf("the run animation became %q three ticks into a flinch", def.Anim)
	}
	if def.AnimCurrentFrame == frame {
		t.Errorf("the run animation is frozen on frame %d during a flinch", frame)
	}
}

// The flicker runs out with the flinch hold, and swinging out of a flinch takes
// it along: nothing may leave a fighter white for good.
func TestFlickerEndsWithTheFlinch(t *testing.T) {
	_, att, def := swing(t, 0)
	def.takeHit(att)
	def.FreezeTicks = 0
	for range FlinchHoldTicks {
		if !def.Flickering() {
			t.Fatalf("the flicker stopped with %d flinch ticks left", def.FlinchTicks)
		}
		def.Update(float32(FixedTimeStep))
	}
	if def.Flickering() {
		t.Errorf("the flicker outlasted the flinch hold")
	}

	def.takeHit(att)
	def.FreezeTicks = 0
	if !def.enterAttackState() {
		t.Fatal("the flinching fighter could not swing")
	}
	if def.Flickering() {
		t.Errorf("swinging out of a flinch left the fighter flickering")
	}
}

// A swing lands once however long its box overlaps, and a chained swing lands again.
func TestOneLandingPerSwing(t *testing.T) {
	_, att, def := swing(t, 0)
	for range 10 {
		att.updateAttack()
	}
	if want := float32(MaxStructure - HitStructure); def.Structure != want {
		t.Errorf("a swing held over the defender left %v structure, want one hit's worth (%v)", def.Structure, want)
	}

	att.SwingHit = false // as starting the next swing does
	att.updateAttack()
	// the second hit of the run costs a step more than the first
	if want := float32(MaxStructure - HitStructure - (HitStructure + HitStructureStep)); def.Structure != want {
		t.Errorf("after a second swing: %v structure, want %v", def.Structure, want)
	}
}

// A block is only worth structure if the guard is still up at the hit frame.
func TestBlockResolvesAtTheHitFrame(t *testing.T) {
	_, att, def := swing(t, 1) // guard matches the attacker's height
	def.Structure = 8

	att.AnimCurrentFrame = 0 // before the hit frame
	att.updateAttack()
	if def.State != StateBlock {
		t.Fatalf("the defender is in state %d, want it guarding (%d)", def.State, StateBlock)
	}
	if def.Structure != 8 {
		t.Errorf("raising a guard moved structure to %v before the swing resolved", def.Structure)
	}

	def.Updown = 2 // guard dropped to the wrong height
	att.advanceAnim()
	att.updateAttack()
	if want := float32(8 - HitStructure); def.Structure != want {
		t.Errorf("a guard dropped before the hit frame left %v structure, want a hit's %v", def.Structure, want)
	}
}

// Structure at zero kills, from a hit or from block-all, and a dying fighter is
// out of the fight.
func TestStructureBreakKills(t *testing.T) {
	_, att, def := swing(t, 0)
	// a run kills when the pool runs out, not after a fixed number of hits: the
	// escalating cost is what decides how many that is
	hits := 0
	for range 12 {
		if def.State == StateDeath {
			break
		}
		if def.Structure <= 0 {
			t.Fatalf("after hit %d the fighter is at %v structure and still alive", hits, def.Structure)
		}
		def.State = StateIdle
		def.takeHit(att)
		hits++
	}
	if def.State != StateDeath {
		t.Errorf("twelve hits left state %d, want death (%d)", def.State, StateDeath)
	}
	if hits != 6 {
		t.Errorf("a clean run killed in %d hits, want the 6 the escalating cost takes", hits)
	}
	if def.Structure != 0 {
		t.Errorf("a broken fighter has %v structure, want 0", def.Structure)
	}

	// nothing more can be done to it
	before := def.Structure
	att.SwingHit = false
	att.updateAttack()
	if def.Structure != before || def.State != StateDeath {
		t.Errorf("a swing at a dying fighter changed it: %v structure, state %d", def.Structure, def.State)
	}

	_, att2, def2 := swing(t, 3) // block-all
	def2.Structure = BlockAllStructure
	def2.confirmBlock(att2)
	if def2.State != StateDeath {
		t.Errorf("block-all that emptied structure left state %d, want death (%d)", def2.State, StateDeath)
	}
}

// The body falls, flickers, leaves, and a new fighter takes its place.
func TestDeathRemovesAndRespawns(t *testing.T) {
	w, att, def := swing(t, 0)
	att.State = StateIdle
	def.Structure = HitStructure
	def.takeHit(att)
	if def.State != StateDeath {
		t.Fatalf("the killing hit left state %d, want death", def.State)
	}

	gone := -1
	for tick := range 400 {
		run(w, 1)
		if gone < 0 && w.ByID(def.ID) == nil {
			gone = tick
		}
	}
	if gone < 0 {
		t.Fatalf("the body is still in the world after 400 ticks")
	}
	if gone < DeathFlickerTicks {
		t.Errorf("the body left after %d ticks, before the fall and flicker could play", gone)
	}

	var spawned *Actor
	for _, a := range w.Actors {
		if a != att {
			spawned = a
		}
	}
	if spawned == nil {
		t.Fatalf("nothing respawned within %d ticks of the body leaving", 400-gone)
	}
	if spawned.Structure != MaxStructure || spawned.State != StateIdle || spawned.HitCount != 0 {
		t.Errorf("respawned with %v structure, state %d, %d hits: want a clean fighter", spawned.Structure, spawned.State, spawned.HitCount)
	}
}

// The flicker hides the body for part of the time and shows it for the rest.
func TestDeathFlickerAlternates(t *testing.T) {
	_, att, def := swing(t, 0)
	def.Structure = HitStructure
	def.takeHit(att)
	def.AnimFinished = true

	hidden, shown := 0, 0
	for range DeathFlickerTicks {
		if def.deathHidden() {
			hidden++
		} else {
			shown++
		}
		def.DeathTicks++
	}
	if hidden == 0 || shown == 0 {
		t.Errorf("over the flicker the body was hidden %d ticks and shown %d: want both", hidden, shown)
	}
}

// A predicted world animates swings but decides nothing: no structure is spent,
// nobody dies and nobody is removed.
func TestPredictedWorldSpendsNoStructure(t *testing.T) {
	for _, auth := range []bool{true, false} {
		w := NewWorld(auth)
		att := w.Spawn(rl.Vector2{}, true)
		def := w.Spawn(rl.Vector2{X: 20}, true)
		att.Orientation, att.Updown, def.Updown = 1, 1, 0
		att.State = StateAttack
		att.Ticks = 100
		att.SetAnim("resources/threefour/attack-mid.png", AnimOptions{animDuration: 0.3, totalFrames: 8, once: true})
		att.Ticks = 114
		att.advanceAnim()
		def.Structure = HitStructure // one hit from death, if anything lands

		att.updateAttack()

		if spent := def.Structure < HitStructure; spent != auth {
			t.Errorf("authority %t: structure spent = %t", auth, spent)
		}
		if died := def.State == StateDeath; died != auth {
			t.Errorf("authority %t: fighter died = %t", auth, died)
		}
		if w.ByID(def.ID) == nil {
			t.Errorf("authority %t: the fighter was removed during the swing", auth)
		}
	}
}
