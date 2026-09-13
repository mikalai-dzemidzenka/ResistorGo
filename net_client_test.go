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

// energyRun drives a server fighter and a predicted client fighter with the
// same commands, the server `lag` ticks behind, and reports the swings the
// server threw that the client had not predicted. The client runs ahead of the
// server, so their pools are not comparable tick for tick; what a correct
// prediction means here is that replaying from a snapshot reproduces what the
// client already had, which is what c.Corrections counts.
func energyRun(t *testing.T, lag int, ticks int, in func(tick int) Input) (c *Client, mispredicted int) {
	t.Helper()
	Stage = StageLevel // so reconcile does not try to enter the level (needs a window)
	defer func() { Stage = StageMainMenu }()

	sv := NewWorld(true)
	sa := sv.Spawn(rl.Vector2{}, true)
	c = &Client{W: NewWorld(false), ActorID: sa.ID}
	c.reconcile(recOf(sa), 0)
	local := c.W.Local
	dt := float32(FixedTimeStep)

	var executed uint32
	for tick := 1; tick <= ticks; tick++ {
		cmd := in(tick)
		c.Seq++
		cmd.Seq = c.Seq
		c.Ring[cmd.Seq%CmdRing] = cmd
		local.Input = cmd
		local.Update(dt)

		if tick > lag {
			executed++
			sa.Input = c.Ring[executed%CmdRing]
			swinging := sa.State == StateAttack
			sa.Update(dt)
			if !swinging && sa.State == StateAttack && local.State != StateAttack {
				mispredicted++
			}
		}
		if tick%SnapshotEvery == 0 {
			c.reconcile(recOf(sa), executed)
		}
	}
	return c, mispredicted
}

// A client holding attack through an empty pool must throw the refilled swing
// on the same tick the server does. Predicting the spend but not the refill
// leaves the client poorer than the server, so it refuses swings the server
// allows and they arrive later as attacks the player never saw start.
func TestPredictionRefillsWithTheServer(t *testing.T) {
	c, mispredicted := energyRun(t, 9, 900, func(tick int) Input {
		return Input{Attack: true, Updown: 1} // mash: the pool empties, then refills under the finger
	})
	if mispredicted != 0 {
		t.Errorf("%d swings came from the server that the client did not predict", mispredicted)
	}
	if c.Corrections != 0 {
		t.Errorf("%d corrections over a pool emptying and refilling, want a prediction that needs none", c.Corrections)
	}
	if c.W.Local.Energy < 0 || c.W.Local.Energy > MaxEnergy {
		t.Errorf("the predicted pool left its range at %v", c.W.Local.Energy)
	}
}

// Replay costs a swing once: reconciliation rewinds to the snapshot's pool and
// spends forward again, rather than charging on top of what the server sent.
func TestReconcileSpendsEachSwingOnce(t *testing.T) {
	Stage = StageLevel
	defer func() { Stage = StageMainMenu }()

	sv := NewWorld(true)
	sa := sv.Spawn(rl.Vector2{}, true)
	c := &Client{W: NewWorld(false), ActorID: sa.ID}
	c.reconcile(recOf(sa), 0)
	local := c.W.Local
	dt := float32(FixedTimeStep)

	// ten commands the server has not executed yet, one of them a swing
	for seq := 1; seq <= 10; seq++ {
		cmd := Input{Seq: uint32(seq), Attack: seq == 3, Updown: 1}
		c.Seq = uint32(seq)
		c.Ring[cmd.Seq%CmdRing] = cmd
		local.Input = cmd
		local.Update(dt)
	}
	if local.Energy != MaxEnergy-EnergySwing {
		t.Fatalf("the predicted swing cost %v, want %d", MaxEnergy-local.Energy, EnergySwing)
	}

	// the server has executed none of them: its pool is still full
	c.reconcile(recOf(sa), 0)
	if local.Energy != MaxEnergy-EnergySwing {
		t.Errorf("after replay: %v energy, want the one swing charged once against the snapshot's %d", local.Energy, int32(MaxEnergy))
	}
}
