package resistor

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// solidsLevel installs a level of the given collision rectangles, in world
// space, and returns a function that restores the level before it.
func solidsLevel(rects ...rl.Rectangle) func() {
	saved := LE
	LE = newLevelEditor(39, 83, nil, nil)
	LE.objects[0] = LevelLayer{{}: {CollisionRects: rects}}
	LE.buildIndex()
	return func() { LE = saved }
}

// The floor every wall below stands on: its top is at y 100.
var wallFloor = rect(-200, 100, 600, 20)

// stander is a fighter standing on wallFloor, facing right, with the right
// side of its collision box at right.
func stander(right float32) *Actor {
	a := NewActor(rl.Vector2{}, false)
	a.Pos.X = right - (a.CollisionBox.X + a.CollisionBox.Width)
	a.Pos.Y = 100 - (a.CollisionBox.Y + a.CollisionBox.Height)
	a.OnGround = true
	a.Orientation = 1
	return a
}

func boxRight(a *Actor) float32 { return a.Pos.X + a.CollisionBox.X + a.CollisionBox.Width }
func boxLeft(a *Actor) float32  { return a.Pos.X + a.CollisionBox.X }
func feet(a *Actor) float32     { return a.Pos.Y + a.CollisionBox.Y + a.CollisionBox.Height }

func TestBlockSides(t *testing.T) {
	wall := rect(100, 40, 20, 60) // on the floor, 60 tall

	t.Run("moving right stops at the face", func(t *testing.T) {
		defer solidsLevel(wallFloor, wall)()
		a := stander(95)
		a.Speed.X = 100
		from := a.Pos.X
		a.Pos.X += 8
		a.blockSides(from)
		if boxRight(a) != 100 || a.Speed.X != 0 {
			t.Errorf("right side at %v, speed %v; want 100 and 0", boxRight(a), a.Speed.X)
		}
	})

	t.Run("moving left stops at the face", func(t *testing.T) {
		defer solidsLevel(wallFloor, wall)()
		a := stander(139) // left side at 125
		a.Speed.X = -100
		from := a.Pos.X
		a.Pos.X -= 8
		a.blockSides(from)
		if boxLeft(a) != 120 || a.Speed.X != 0 {
			t.Errorf("left side at %v, speed %v; want 120 and 0", boxLeft(a), a.Speed.X)
		}
	})

	t.Run("a one-pixel wall at full speed", func(t *testing.T) {
		defer solidsLevel(wallFloor, rect(100, 40, 1, 60))()
		a := stander(98)
		from := a.Pos.X
		a.Pos.X = float32(int(a.Pos.X + MaxSpeed*float32(FixedTimeStep)))
		if boxLeft(a) > 101 || boxRight(a) <= 101 {
			t.Fatalf("the move does not carry the box across the wall: %v..%v", boxLeft(a), boxRight(a))
		}
		a.blockSides(from)
		if boxRight(a) != 100 {
			t.Errorf("a one-pixel wall let the fighter through to %v", boxRight(a))
		}
	})

	t.Run("overlapping: never pushed, never deeper, free to leave", func(t *testing.T) {
		defer solidsLevel(wallFloor, wall)()
		a := stander(104)
		a.blockSides(a.Pos.X)
		if boxRight(a) != 104 {
			t.Errorf("a fighter overlapping a wall was moved to %v", boxRight(a))
		}
		from := a.Pos.X
		a.Pos.X += 3
		a.blockSides(from)
		if boxRight(a) != 104 {
			t.Errorf("an overlapping fighter went deeper, to %v", boxRight(a))
		}
		from = a.Pos.X
		a.Pos.X -= 3
		a.blockSides(from)
		if boxRight(a) != 101 {
			t.Errorf("an overlapping fighter could not leave: at %v, want 101", boxRight(a))
		}
	})

	t.Run("moving away is untouched", func(t *testing.T) {
		defer solidsLevel(wallFloor, wall)()
		a := stander(100)
		a.Speed.X = -50
		from := a.Pos.X
		a.Pos.X -= 4
		a.blockSides(from)
		if boxRight(a) != 96 || a.Speed.X != -50 {
			t.Errorf("walking away from a wall: right side %v, speed %v", boxRight(a), a.Speed.X)
		}
	})

	for _, c := range []struct {
		rise float32
		wall bool
	}{{4, false}, {5, true}} {
		defer solidsLevel(wallFloor, rect(100, 100-c.rise, 40, c.rise))()
		a := stander(98)
		from := a.Pos.X
		a.Pos.X += 5
		a.blockSides(from)
		if stopped := boxRight(a) == 100; stopped != c.wall {
			t.Errorf("a %v px rise: stopped %t, want %t", c.rise, stopped, c.wall)
		}
	}
}

// hold runs n ticks of Update with a direction held: +1 for right, -1 for
// left. Input.Left is the key that moves a fighter toward +x.
func hold(a *Actor, dir, n int, each func()) {
	for range n {
		a.Input = Input{Left: dir > 0, Right: dir < 0}
		a.Update(float32(FixedTimeStep))
		if each != nil {
			each()
		}
	}
}

func TestWallsStopFighters(t *testing.T) {
	wall := rect(100, 40, 20, 60)

	t.Run("walking", func(t *testing.T) {
		defer solidsLevel(wallFloor, wall)()
		a := stander(90)
		hold(a, 1, 120, func() {
			if boxRight(a) > 100 {
				t.Fatalf("walked into the wall, right side at %v", boxRight(a))
			}
		})
		if boxRight(a) != 100 || feet(a) != 100 {
			t.Errorf("after walking at the wall: right side %v, feet %v; want against it, on the floor", boxRight(a), feet(a))
		}
	})

	t.Run("running", func(t *testing.T) {
		defer solidsLevel(wallFloor, wall)()
		a := stander(40)
		a.State, a.Speed.X = StateRun, MaxSpeed
		// the tick the wall cuts the run short leaves the fighter against it with
		// no speed; a run that happens to end exactly on the face is cut the
		// tick after
		stopped := false
		for range 60 {
			hold(a, 1, 1, nil)
			if boxRight(a) > 100 {
				t.Fatalf("ran into the wall, right side at %v", boxRight(a))
			}
			if boxRight(a) == 100 && a.Speed.X == 0 {
				stopped = true
				break
			}
		}
		if !stopped {
			t.Fatalf("the wall never stopped the run: right side at %v, speed %v", boxRight(a), a.Speed.X)
		}
		hold(a, 1, 2, nil)
		if a.State == StateRun {
			t.Errorf("a fighter stopped by a wall is still running (speed %v)", a.Speed.X)
		}
	})

	t.Run("knocked back", func(t *testing.T) {
		defer solidsLevel(wallFloor, wall)()
		a := stander(96)
		a.State, a.Speed.X = StateAttacked, StunKnockback
		a.SetAnim("resources/threefour/hit-top.png", AnimOptions{animDuration: 2, totalFrames: 4, once: true})
		for range 30 {
			a.Update(float32(FixedTimeStep))
			if boxRight(a) > 100 {
				t.Fatalf("knocked into the wall, right side at %v", boxRight(a))
			}
		}
		if boxRight(a) != 100 {
			t.Errorf("a knockback toward the wall ended at %v, want against it", boxRight(a))
		}
	})

	t.Run("steering in the air", func(t *testing.T) {
		defer solidsLevel(wallFloor, rect(100, 50, 40, 50))()
		a := stander(98)
		a.Pos.Y = 24 // head at 30, above the wall's top at 50; feet at 72, below it
		a.OnGround, a.State, a.Speed.X, a.Speed.Y = false, StateFall, 200, 40
		a.SetAnim("resources/threefour/fall.png", AnimOptions{animDuration: 0.3, totalFrames: 3})
		y := a.Pos.Y
		hold(a, 1, 1, nil)
		if boxRight(a) != 100 || a.Speed.X != 0 {
			t.Errorf("steering into a wall in the air: right side %v, speed %v; want against it, 0", boxRight(a), a.Speed.X)
		}
		if a.State != StateFall || a.Pos.Y <= y {
			t.Errorf("stopped by a wall in the air, the fighter is in state %d at y %v, want still falling below %v", a.State, a.Pos.Y, y)
		}
	})
}

func TestSwingDashAgainstAWall(t *testing.T) {
	defer solidsLevel(wallFloor, rect(100, 40, 20, 60))()
	a := stander(100)
	x := a.Pos.X
	a.Input = Input{Attack: true}
	a.Update(float32(FixedTimeStep))
	if a.State != StateAttack {
		t.Fatalf("the swing did not come out: state %d", a.State)
	}
	if a.Pos.X != x {
		t.Errorf("a swing against a wall dashed into it: x %v, was %v", a.Pos.X, x)
	}

	b := stander(100)
	b.Orientation = -1
	x = b.Pos.X
	b.Input = Input{Attack: true}
	b.Update(float32(FixedTimeStep))
	if b.Pos.X != x-AttackDash {
		t.Errorf("a swing away from a wall dashed to %v, want %v", b.Pos.X, x-AttackDash)
	}
}

func TestStepsAndFloors(t *testing.T) {
	t.Run("a four-pixel step", func(t *testing.T) {
		defer solidsLevel(wallFloor, rect(100, 96, 300, 24))()
		a := stander(90)
		hold(a, 1, 40, nil) // about 100 px: onto the step and well short of its far end
		if !a.OnGround || feet(a) != 96 || a.State == StateJumpStart || a.State == StateJump {
			t.Errorf("walking onto a 4 px step: feet at %v, on ground %t, state %d; want standing on it at 96", feet(a), a.OnGround, a.State)
		}
		if boxLeft(a) < 100 {
			t.Errorf("the fighter stopped at the step, left side at %v", boxLeft(a))
		}
	})

	t.Run("a five-pixel rise", func(t *testing.T) {
		defer solidsLevel(wallFloor, rect(100, 95, 60, 25))()
		a := stander(90)
		hold(a, 1, 90, nil)
		if boxRight(a) != 100 || feet(a) != 100 {
			t.Errorf("a 5 px rise: right side %v, feet %v; want stopped against it on the floor", boxRight(a), feet(a))
		}
	})

	t.Run("a floor of level tiles", func(t *testing.T) {
		defer solidsLevel(rect(0, 100, 40, 20), rect(39, 100, 40, 20), rect(78, 100, 40, 20), rect(117, 100, 40, 20), rect(156, 100, 200, 20))()
		a := stander(20)
		a.State, a.Speed.X = StateRun, MaxSpeed
		prev := a.Pos.X
		hold(a, 1, 30, func() {
			if a.Pos.X <= prev || a.Speed.X < MaxSpeed-1 {
				t.Fatalf("running across level tiles lost speed at x %v (speed %v)", a.Pos.X, a.Speed.X)
			}
			prev = a.Pos.X
		})
	})
}

// The current level has no rectangle that is a wall to a fighter on its main
// floor: a fighter runs the length of its unbroken stretch and back.
func TestLevel1FloorHasNoWalls(t *testing.T) {
	saved := LE
	defer func() { LE = saved }()
	LE = NewLevelEditor(39, 83)
	LE.LoadLevel("level_1.json")

	a := NewActor(rl.Vector2{}, false)
	a.Pos.X = -470 - (a.CollisionBox.X + a.CollisionBox.Width/2)
	a.Pos.Y = 86 - (a.CollisionBox.Y + a.CollisionBox.Height) // the row of tiles at y 0, top at 86
	a.OnGround = true
	a.Update(float32(FixedTimeStep))
	if !a.OnGround {
		t.Fatalf("the fighter is not on the level's floor at x %v", a.Pos.X)
	}
	for _, leg := range []struct {
		dir    int
		target float32
	}{{1, 110}, {-1, -470}} {
		center := func() float32 { return a.Pos.X + a.CollisionBox.X + a.CollisionBox.Width/2 }
		for i := 0; ; i++ {
			if i > 240 {
				t.Fatalf("stopped short of %v at %v", leg.target, center())
			}
			if (leg.dir > 0 && center() >= leg.target) || (leg.dir < 0 && center() <= leg.target) {
				break
			}
			hold(a, leg.dir, 1, nil)
			if !a.OnGround {
				t.Fatalf("left the floor at x %v", center())
			}
		}
	}
}

func TestLedgeAgainstAWall(t *testing.T) {
	wall := rect(100, 50, 40, 50) // 50 tall: its top is above the head of a fighter on the floor

	t.Run("jumping at it", func(t *testing.T) {
		defer solidsLevel(wallFloor, wall)()
		a := stander(100)
		a.Input = Input{Left: true, Jump: true}
		a.Update(float32(FixedTimeStep))
		for range 120 {
			if a.State == StateClimb {
				break
			}
			hold(a, 1, 1, nil)
			if boxRight(a) > 100 && a.State != StateClimb {
				t.Fatalf("jumped into the wall, right side at %v", boxRight(a))
			}
		}
		if a.State != StateClimb {
			t.Errorf("jumping at a wall reached state %d, want the climb", a.State)
		}
	})

	t.Run("coming down past it from a jump", func(t *testing.T) {
		defer solidsLevel(wallFloor, wall)()
		a := stander(100)
		a.Pos.Y = 14 // head at 20, above the top at 50; feet at 62, below it, so it is beside the wall
		a.OnGround, a.State, a.Speed.Y, a.Jumped = false, StateFall, 40, true
		a.SetAnim("resources/threefour/fall.png", AnimOptions{animDuration: 0.3, totalFrames: 3})
		for range 60 {
			hold(a, 1, 1, nil)
			if a.State == StateClimb {
				break
			}
			if boxRight(a) > 100 {
				t.Fatalf("fell into the wall, right side at %v", boxRight(a))
			}
		}
		if a.State != StateClimb {
			t.Errorf("falling past a wall reached state %d, want the climb", a.State)
		}
	})
}

func TestOverlappingAWall(t *testing.T) {
	defer solidsLevel(wallFloor, rect(100, 40, 20, 60))()
	a := stander(104)
	a.Update(float32(FixedTimeStep))
	if boxRight(a) != 104 {
		t.Errorf("a fighter overlapping a wall was moved to %v with no input", boxRight(a))
	}
	hold(a, 1, 30, func() {
		if boxRight(a) > 104 {
			t.Fatalf("walked deeper into the wall, to %v", boxRight(a))
		}
	})
	hold(a, -1, 30, nil)
	if boxRight(a) >= 100 {
		t.Errorf("could not walk out of the wall: right side at %v", boxRight(a))
	}
}
