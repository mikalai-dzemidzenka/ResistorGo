# Recovered: the withdrawn first implementation

Recovered from the transcript of session `ba331b21-fb37-4693-afb3-199d8ada7f04`
(2026-09-18, 19:05-23:46 UTC), in the state it had when it was withdrawn on
2026-09-19. Reference only, never compiled from here - it is Markdown so that
`go build ./...` does not pick it up as a package.

It implements the **withdrawn premise**: it shoves from `StateIdle` *or*
`StateRun`, shoves the way the fighter already faces without turning, and moves
the crowder 20px. design.md - Decisions says what changes; the places that
change are marked `// CHANGES:` below.

Already committed since, and not repeated here: the `stateCount` sentinel and
the decoder bound (`2781f87`), `TestEveryStateSurvivesTheWire`, and the bot
course test holding the bot's swing back.

## resources.go

Appended to the texture list in `LoadTextures`, after `fall.png`:

```go
	"resources/threefour/push.png",
```

## actor.go

### The state

Appended to the `ActorState` enum after `StateFall`, before `stateCount`:

```go
	StatePush
```

### Constants, after `FlinchHoldTicks`

```go
	// A shove is what two fighters standing in one another do about it. It
	// costs the fighter shoved nothing but its ground - no structure, no stun -
	// and it is the only answer to the spot where neither can reach the other:
	// the arm covers the half of the sprite in front, so a fighter on another's
	// feet has nothing to hit, and cannot step back either, facing the way it
	// walks.
	// CHANGES: the premise is now a fighter on your back, see proposal.md
	PushSeconds = 0.35
	PushFrames  = 7
	PushFrame   = 3 // the frame both arms are out: where the shove lands
	// How far the shove moves the fighter in front. Far enough to clear a
	// collision box, so the two are no longer standing in one another, and near
	// enough to stay inside the reach of a swing, so a shove opens the fight
	// back up rather than ending it.
	// CHANGES: 16, bounded to [11, 21] by a test
	PushDistance = 20
	PushSpeed    = PunchSpeed // what it leaves in the shoved fighter, as a hit does
```

### Field, after `SwingHit`

```go
	PushDone         bool    // this fighter's current shove has already moved someone
```

### `canMove`

```go
	canMove := a.State != StateAttack && a.State != StateBlock && a.State != StateAttacked && a.State != StatePush && a.Updown != 3
```

### The trigger, in `Actor.Update` after `a.updateFalling()`, before `switch a.State`

```go
	// Two fighters cannot stand in one another. Whoever is on their feet and
	// free to act shoves the other out of the way rather than swinging at a
	// body its arm passes over. A fighter guarding everything is left to stand
	// its ground: block-all is what it does instead of moving.
	// CHANGES: StateIdle only, and pass the crowder to enterPushState
	if a.OnGround && a.Updown != 3 && (a.State == StateIdle || a.State == StateRun) && a.crowding() != nil {
		a.enterPushState()
	}
```

### The case, after `case StateAttack`

```go
	case StatePush:
		a.updatePush()
```

### The functions, before `enterDeathState`

```go
// crowding returns a fighter standing in this one that this one cannot swing
// at: the collision boxes overlap, and the arm - which covers the half of the
// sprite in front - passes over the body rather than into it.
//
// This is a spot two fighters can walk into and neither get out of. A bot
// creeps in four pixels with every swing it throws and a player can simply walk
// in; there, nobody can hit anybody, and nobody can step back either.
func (a *Actor) crowding() *Actor {
	if a.W == nil {
		return nil
	}
	for _, other := range a.W.Actors {
		if other == a || other.State == StateDeath {
			continue
		}
		if !aabb(a.collisionRect(), other.collisionRect()) {
			continue
		}
		if a.swingReachFacing(other, a.Orientation) {
			continue // standing close, but there is still something to hit
		}
		return other
	}
	return nil
}

// swingReachFacing reports whether a swing thrown facing o would meet t: the
// box this fighter's guard height throws, carried forward by the dash the swing
// takes, against the box a hit is decided against. The same rectangles the hit
// test asks, so nothing that decides on reach can drift away from what reaches.
func (a *Actor) swingReachFacing(t *Actor, o int8) bool {
	box, ok := a.attackBoxFacing(a.Updown, a.Pos, o)
	if !ok {
		// guarding with block-all: no swing to measure, so measure the one it
		// would throw the moment it stopped guarding
		box, _ = a.attackBoxFacing(1, a.Pos, o)
	}
	box.X += float32(o) * AttackDash
	return aabb(box, t.collisionRect())
}

// enterPushState starts the shove: both arms out, and whoever is standing in
// the fighter is taken out of the way. It cannot be steered while it plays.
//
// The shove goes the way the fighter already faces, and the fighter does not
// turn to throw it. Two fighters standing in one another are within a body's
// width, where which side the other is on is a pixel of noise - turning to it
// pointed a bot away from the fight as often as toward it, and a bot that
// turns away loses sight of what it was following on the tick it turns.
// CHANGES: take the crowder and face it - the window no longer contains dx = 0,
// so the side is never noise. The version that turned, from 23:27 UTC, was:
//
//	func (a *Actor) pushDir(other *Actor) int8 {
//		if d := other.Pos.X - a.Pos.X; d != 0 {
//			return facing(d)
//		}
//		return a.Orientation
//	}
//
//	func (a *Actor) enterPushState(other *Actor) {
//		a.face(a.pushDir(other))
//		...
func (a *Actor) enterPushState() {
	a.State = StatePush
	a.PushDone = false
	a.Speed.X = 0
	a.SetAnim("resources/threefour/push.png", AnimOptions{animDuration: PushSeconds, totalFrames: PushFrames, once: true})
}

// updatePush lands the shove on the frame the arms come out, once, and ends on
// the animation. The shove is the authority's alone, as a hit is: a client is
// shown the animation, and the ground it changed hands, by the snapshot.
func (a *Actor) updatePush() {
	if a.W != nil && a.W.Authority && !a.PushDone && a.AnimCurrentFrame >= PushFrame {
		a.PushDone = true
		for _, other := range a.W.Actors {
			if other == a || other.State == StateDeath {
				continue
			}
			if aabb(a.collisionRect(), other.collisionRect()) {
				a.shove(other)
			}
		}
	}
	if a.AnimFinished {
		a.enterIdleState()
	}
}

// shove moves a fighter out of the way, the way the arms point.
//
// The ground it gives up is a displacement and not only speed: a fighter
// mid-swing holds its own speed down to the lunge, and would otherwise stand
// its ground through a shove. Walls are asked the same way every other move
// asks them, so nobody is shoved into one.
// CHANGES: dir is away from the shover; set Speed.X before blockSides
func (a *Actor) shove(other *Actor) {
	dir := float32(a.Orientation)
	fromX := other.Pos.X
	other.Pos.X = float32(int(other.Pos.X + dir*PushDistance))
	other.blockSides(fromX)
	other.Speed.X = dir * PushSpeed
}
```

## bot.go

`inSwingRange` became a caller, its body replaced up to the `tooClose` comment.
Since then `584ad7b` added a comment block there about `tooClose`; keep it.

```go
func (a *Actor) inSwingRange(t *Actor) bool {
	return a.swingReachFacing(t, facing(t.Pos.X-a.Pos.X))
}
```

## jump_test.go

`TestAnimTableIsAppendOnly` pins the previous last two by number and holds the
newest to the end:

```go
		"resources/threefour/turnover.png":   45,
		"resources/threefour/run-jump.png":   46,
		"resources/threefour/fall.png":       47,
	} {
		if got := animID(path); got != want {
			t.Errorf("%s has id %d, want %d - something was inserted rather than appended", path, got, want)
		}
	}

	// the newest one is on the end, and survives the round trip
	const newest = "resources/threefour/push.png"
	id := animID(newest)
	if want := uint8(len(animTable) - 1); id != want {
		t.Errorf("%s has id %d, want %d - it is not at the end of the table", newest, id, want)
	}
	if back := animPath(id); back != newest {
		t.Errorf("id %d came back as %q, want %q", id, back, newest)
	}
}
```

The two fixtures that stood both fighters on one column move the runner clear:

```go
	// clear of the other: two fighters standing in one another shove each other
	// apart, and this test is about the jump
	runner := ground(w, 120)
	runner.State = StateRun
```

```go
	kept := ground(w, 120) // clear of the other, so neither shoves the other
	kept.State = StateRun
```

## push_test.go

Every `crowd(t, 0)` here stands the two at `dx = 0`, which the corrected reach no
longer treats as crowding: those setups move into the window, `[-13, -3]` facing
right. `TestCloseButReachableDoesNotShove`'s sweep and
`TestAShoveLeavesTheFightInReach` need both facings and the whole window.

```go
package resistor

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// shoveTicks is how long the shove takes to land: the frame both arms come out,
// at the rate the animation runs, and a tick either side of it.
const shoveTicks = int(60*PushSeconds/PushFrames*PushFrame) + 2

// crowd puts two fighters dx apart on a floor, settled on their feet. dx of 0
// stands them in one another, where neither has anything to swing at.
func crowd(t *testing.T, dx float32) (*World, *Actor, *Actor) {
	t.Helper()
	floor()
	w := NewWorld(true)
	a := w.Spawn(rl.Vector2{}, true)
	b := w.Spawn(rl.Vector2{X: dx}, true)
	a.Orientation, b.Orientation = 1, 1
	for _, f := range []*Actor{a, b} {
		f.checkCollision()
		if !f.OnGround {
			t.Fatalf("the fighter at %v is not standing on the floor", f.Pos.X)
		}
	}
	return w, a, b
}

// Two fighters cannot stand in one another: one shoves the other out of the way
// and they end up a body's width apart, with nothing else changed hands.
func TestFightersStandingInOneAnotherShoveApart(t *testing.T) {
	w, a, b := crowd(t, 0)
	if !aabb(a.collisionRect(), b.collisionRect()) {
		t.Fatal("the two are meant to be standing in one another")
	}
	structure, energy := b.Structure, b.Energy

	run(w, 1)
	if a.State != StatePush {
		t.Fatalf("the crowded fighter is in state %d, want the shove %d", a.State, StatePush)
	}
	if a.Anim != "resources/threefour/push.png" {
		t.Errorf("the shove plays %q, want push.png", a.Anim)
	}

	run(w, int(60*PushSeconds)+2) // the shove plays out
	if aabb(a.collisionRect(), b.collisionRect()) {
		t.Errorf("after the shove the two are still standing in one another: %v and %v", a.Pos.X, b.Pos.X)
	}
	if a.State == StatePush {
		t.Error("the shove never ended")
	}
	if b.Structure != structure {
		t.Errorf("the shove cost the fighter shoved %v structure, want nothing", structure-b.Structure)
	}
	if b.Energy != energy {
		t.Errorf("the shove moved the shoved fighter's energy by %v, want nothing", b.Energy-energy)
	}
	if b.HitCount != 0 || b.State == StateAttacked {
		t.Error("the shove was counted as a hit")
	}
}

// The shove leaves the fighter it moved inside the reach of a swing: it opens
// the fight back up rather than ending it.
func TestAShoveLeavesTheFightInReach(t *testing.T) {
	w, a, b := crowd(t, 0)
	run(w, int(60*PushSeconds)+2)
	if !a.swingReachFacing(b, a.Orientation) {
		t.Errorf("the shove put the other fighter %v away, out of reach of a swing", b.Pos.X-a.Pos.X)
	}
}

// Standing close is not standing in: where there is still something to hit, a
// fighter swings rather than shoving. Otherwise the shove would take the place
// of close-quarters fighting altogether.
func TestCloseButReachableDoesNotShove(t *testing.T) {
	for dx := float32(1); dx < 14; dx++ { // inside a collision box's width
		w, a, b := crowd(t, dx)
		if !aabb(a.collisionRect(), b.collisionRect()) {
			t.Fatalf("dx=%v: the two are meant to be overlapping", dx)
		}
		if !a.swingReachFacing(b, a.Orientation) {
			t.Fatalf("dx=%v: a swing is meant to reach from here", dx)
		}
		run(w, 1)
		if a.State == StatePush {
			t.Errorf("dx=%v: the fighter shoved a fighter it could have hit", dx)
		}
	}
}

// A fighter guarding everything stands its ground: block-all is what it does
// instead of moving, so it neither shoves nor is moved out of the way by its
// own doing.
func TestBlockAllStandsItsGround(t *testing.T) {
	w, a, _ := crowd(t, 0)
	a.Updown, a.PrevUpdown = 3, 3
	run(w, 1)
	if a.State == StatePush {
		t.Error("a fighter guarding everything shoved instead of standing its ground")
	}
}

// The shove comes off the feet. A fighter in the air has no shove to throw, any
// more than it has a swing.
func TestAFighterInTheAirDoesNotShove(t *testing.T) {
	w, a, _ := crowd(t, 0)
	a.OnGround, a.Speed.Y = false, -10
	a.State = StateJump
	run(w, 1)
	if a.State == StatePush {
		t.Error("a fighter in the air threw a shove")
	}
}

// A fighter mid-swing is moved too. Its own state holds its speed down to the
// lunge, so a shove that only pushed speed around would leave it standing
// exactly where it was.
func TestAShoveMovesAFighterMidSwing(t *testing.T) {
	w, _, b := crowd(t, 0)
	b.Updown = 1
	b.State = StateAttack
	b.SetAnim("resources/threefour/attack-mid.png", AnimOptions{animDuration: 0.3, totalFrames: 8, once: true})
	from := b.Pos.X

	run(w, shoveTicks)
	if b.Pos.X == from {
		t.Errorf("the fighter mid-swing was not moved: still at %v", b.Pos.X)
	}
	if b.State != StateAttack {
		t.Errorf("the shove put the swinging fighter into state %d, want its swing %d left alone", b.State, StateAttack)
	}
}

// The dead are out of the fight, and are not shoved out of the way either.
func TestTheDeadAreNotShoved(t *testing.T) {
	w, a, b := crowd(t, 0)
	b.State = StateDeath
	from := b.Pos.X
	run(w, shoveTicks)
	if a.State == StatePush {
		t.Error("a fighter shoved a body")
	}
	if b.Pos.X != from {
		t.Errorf("the body was moved to %v from %v", b.Pos.X, from)
	}
}

// The shove is the authority's: a client is shown the animation and the ground
// it changed hands by the snapshot, and invents neither.
func TestAClientDoesNotShove(t *testing.T) {
	floor()
	w := NewWorld(false) // a predicting client's world
	a := w.Spawn(rl.Vector2{}, true)
	b := w.Spawn(rl.Vector2{}, true)
	a.Orientation, b.Orientation = 1, 1
	for _, f := range []*Actor{a, b} {
		f.checkCollision()
	}
	from := b.Pos.X
	run(w, shoveTicks)
	if b.Pos.X != from {
		t.Errorf("a client moved a fighter with a shove of its own: %v to %v", from, b.Pos.X)
	}
}
```
