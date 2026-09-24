package resistor

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// dropLevel is a solid floor with its top at y 300, from x 0 to 400, and a
// drop-through platform over it with its top at y 150, from x 0 to 200, 16 px
// thick. It returns a function that puts the level before it back.
func dropLevel() func() {
	saved := LE
	le := newLevelEditor(39, 83, nil, nil)
	le.objects[0] = LevelLayer{
		{}:             {CollisionRects: []rl.Rectangle{rect(0, 300, 400, 40)}},
		{X: 0, Y: 150}: {X: 0, Y: 150, DropRects: []rl.Rectangle{rect(0, 0, 200, 16)}},
	}
	le.buildIndex()
	LE = le
	return func() { LE = saved }
}

// standOn puts a fighter of w centred on x, standing with its feet on top, and
// settles it there.
func standOn(w *World, x, top float32) *Actor {
	a := w.Spawn(rl.Vector2{}, true)
	a.Pos.X = x - (a.CollisionBox.X + a.CollisionBox.Width/2)
	a.Pos.Y = top - 48
	a.Orientation = 1
	a.OnGround = true
	a.Input.FaceLeft = false
	return a
}

func TestDropThroughIsFloorFromAbove(t *testing.T) {
	defer dropLevel()()
	step := float32(FixedTimeStep)

	t.Run("walked on", func(t *testing.T) {
		w := NewWorld(true)
		a := standOn(w, 30, 150)
		a.Input.Left = true // D: to the right
		for range 60 {
			if a.Pos.X+a.CollisionBox.X+a.CollisionBox.Width/2 > 180 {
				break // short of the far edge
			}
			a.Update(step)
		}
		if a.Pos.X < 60 {
			t.Fatalf("the fighter did not walk: x %v", a.Pos.X)
		}
		a.checkCollision()
		if feet(a) != 150 || !a.OnGround {
			t.Errorf("walking across the platform left the feet at %v (on ground %t), want 150", feet(a), a.OnGround)
		}
	})

	t.Run("landed on", func(t *testing.T) {
		w := NewWorld(true)
		a := standOn(w, 100, 100)
		a.OnGround = false
		a.State = StateFall
		for range 90 {
			a.Update(step)
		}
		if feet(a) != 150 || !a.OnGround {
			t.Errorf("falling onto the platform left the feet at %v (on ground %t), want 150", feet(a), a.OnGround)
		}
	})

	t.Run("fast landing", func(t *testing.T) {
		w := NewWorld(true)
		a := standOn(w, 100, 140)
		a.OnGround = false
		a.State = StateFall
		a.Speed.Y = MaxFallSpeed
		for range 10 {
			a.Update(step)
		}
		if feet(a) != 150 {
			t.Errorf("falling at full speed onto the platform left the feet at %v, want 150", feet(a))
		}
	})
}

func TestDropThroughIsOpenFromBelow(t *testing.T) {
	defer dropLevel()()

	// head just inside the platform's underside, rising: it goes on rising
	a := jumper(100, 160)
	a.checkCollision()
	if a.Speed.Y != -5 || a.Pos.Y != 160-a.CollisionBox.Y || a.OnGround {
		t.Errorf("rising into the platform: speed %v, y %v, on ground %t; want it left rising where it was", a.Speed.Y, a.Pos.Y, a.OnGround)
	}

	// feet inside it on the way up: not stood on until it comes down
	b := jumper(100, 150-42+4)
	b.checkCollision()
	if b.OnGround || b.Speed.Y != -5 {
		t.Errorf("rising with the feet in the platform: on ground %t, speed %v; want still rising", b.OnGround, b.Speed.Y)
	}

	// head level with the platform just short of its right edge, facing it
	c := jumper(215, 155)
	c.Orientation = -1
	c.checkCollision()
	if !c.CanClimbLeft {
		t.Errorf("rising at the platform's edge did not catch its ledge")
	}
}

// A run-jump from the floor under a drop-through platform takes the head up
// through it and comes down on top: the platform stands higher than the head,
// so a platform that stopped heads would have sent it back to the floor.
func TestJumpUpOntoADropThrough(t *testing.T) {
	saved := LE
	defer func() { LE = saved }()
	LE = NewLevelEditor(39, 83)
	LE.objects[0] = LevelLayer{
		{}:       {CollisionRects: []rl.Rectangle{rect(-200, 300, 2400, 40)}},
		{Y: 245}: {Y: 245, DropRects: []rl.Rectangle{rect(-200, 0, 2400, 6)}},
	}
	LE.buildIndex()
	w := NewWorld(true)
	a := standOn(w, 100, 300)
	if head := a.Pos.Y + a.CollisionBox.Y; head <= 251 {
		t.Fatalf("the head at %v is not under the platform's underside at 251", head)
	}
	a.State = StateRun
	a.Speed.X = RunReach + 20
	a.Input.Left = true // running right
	a.Input.Jump = true
	a.Update(float32(FixedTimeStep))
	a.Input.Jump = false
	for range 200 {
		if a.OnGround && a.State != StateRunJump {
			break
		}
		a.Update(float32(FixedTimeStep))
	}
	if feet(a) != 245 || !a.OnGround {
		t.Errorf("the jump ended with the feet at %v (on ground %t), want standing on the platform at 245", feet(a), a.OnGround)
	}
}

func TestDropThroughPassesAFighterBelowIt(t *testing.T) {
	defer dropLevel()()

	// below the top, standing still or falling: not held, not stopped, no ledge
	for _, sy := range []float32{0, 100} {
		a := jumper(100, 160) // head inside the platform
		a.Speed.Y = sy
		a.checkCollision()
		if a.OnGround || a.Pos.Y != 160-a.CollisionBox.Y || a.Speed.Y != sy {
			t.Errorf("speed %v: a fighter below the top was held (%t) or moved to y %v", sy, a.OnGround, a.Pos.Y)
		}
		b := jumper(215, 155)
		b.Orientation = -1
		b.Speed.Y = sy
		b.checkCollision()
		if b.CanClimbLeft {
			t.Errorf("speed %v: a fighter falling past the platform caught its ledge", sy)
		}
	}

	// feet a few px into it, having started the tick below the top
	a := NewActor(rl.Vector2{}, true)
	a.Pos = rl.Vector2{X: 80, Y: 150 - 48 + 5}
	a.checkCollision()
	if a.OnGround {
		t.Errorf("a fighter placed below the top with no speed was stood on the platform")
	}
}

func TestOnDropThrough(t *testing.T) {
	defer dropLevel()()
	w := NewWorld(true)
	if a := standOn(w, 100, 150); !a.onDropThrough() {
		t.Errorf("standing on the platform is not standing on a drop-through")
	}
	if a := standOn(w, 300, 300); a.onDropThrough() {
		t.Errorf("standing on the solid floor counts as drop-through")
	}
	if a := standOn(w, 100, 120); a.onDropThrough() {
		t.Errorf("a fighter in the air counts as on a drop-through")
	}

	// a solid sharing the platform's top under the same feet
	LE.objects[0][Vector2i{X: 50, Y: 150}] = LevelObject{X: 50, Y: 150, CollisionRects: []rl.Rectangle{rect(0, 0, 100, 16)}}
	LE.buildIndex()
	if a := standOn(w, 100, 150); a.onDropThrough() {
		t.Errorf("a fighter held up by a solid too counts as on a drop-through")
	}
}
