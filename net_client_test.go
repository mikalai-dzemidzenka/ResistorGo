package resistor

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"testing"
)

// predictionRun drives a server actor and a predicted client actor with the same
// commands; the server runs them `lag` ticks late and snapshots every 3 ticks.
func predictionRun(t *testing.T, lag int, onServer func(tick int, a *Actor)) *Client {
	t.Helper()
	Stage = StageLevel // so reconcile does not try to enter the level (needs a window)
	defer func() { Stage = StageMainMenu }()

	sv := NewWorld(true)
	sa := sv.Spawn(rl.Vector2{}, true)
	c := &Client{W: NewWorld(false), ActorID: sa.ID}
	c.reconcile(recOf(sa), 0)
	local := c.W.Local
	dt := float32(FixedTimeStep)

	var executed uint32
	for tick := 1; tick <= 240; tick++ {
		in := Input{Left: tick%90 < 40, Right: tick%90 > 60, Attack: tick%45 == 7, Jump: tick == 120, Updown: uint8(tick/30) % 3, FaceLeft: tick%90 > 60}
		c.Seq++
		in.Seq = c.Seq
		c.Ring[in.Seq%CmdRing] = in
		local.Input = in
		local.Update(dt)

		if tick > lag {
			executed++
			sa.Input = c.Ring[executed%CmdRing]
			sa.Update(dt)
			if onServer != nil {
				onServer(int(executed), sa)
			}
		}
		if tick%SnapshotEvery == 0 {
			c.reconcile(recOf(sa), executed)
		}
	}
	return c
}

func TestPredictionMatchesServer(t *testing.T) {
	if c := predictionRun(t, 9, nil); c.Corrections != 0 {
		t.Errorf("%d corrections for identical simulation, want 0", c.Corrections)
	}
}

func TestPredictionCorrectsServerHit(t *testing.T) {
	hit := false
	c := predictionRun(t, 9, func(tick int, a *Actor) {
		if tick == 200 && !hit {
			hit = true
			attacker := NewActor(rl.Vector2{}, false)
			attacker.ID = 99
			a.enterAttackedState(attacker)
		}
	})
	if c.Corrections == 0 {
		t.Error("server-only hit produced no correction")
	}
	if c.W.Local.Ticks == 0 {
		t.Error("local actor not simulated")
	}
}
