package resistor

import (
	"slices"
	"testing"
)

func TestCmdQueueOrderAndDuplicates(t *testing.T) {
	var cq cmdQueue
	var ran []uint32
	run := func(in Input) { ran = append(ran, in.Seq) }

	// redundant, reordered and duplicated packets
	for _, s := range []uint32{3, 1, 2, 2, 1, 3} {
		cq.push(Input{Seq: s})
	}
	for range 3 {
		cq.drain(run)
	}
	if !slices.Equal(ran, []uint32{1, 2, 3}) {
		t.Fatalf("executed %v, want [1 2 3]", ran)
	}
	cq.push(Input{Seq: 2}) // already executed
	cq.drain(run)
	if len(ran) != 3 {
		t.Errorf("old command executed again: %v", ran)
	}
}

func TestCmdQueueSteady(t *testing.T) {
	var cq cmdQueue
	for tick := uint32(1); tick <= 100; tick++ {
		cq.push(Input{Seq: tick})
		n := 0
		cq.drain(func(Input) { n++ })
		if n != 1 {
			t.Fatalf("tick %d executed %d commands, want 1", tick, n)
		}
	}
}

func TestCmdQueueSpeedHack(t *testing.T) {
	var cq cmdQueue
	var seq uint32
	executed := 0
	const ticks = 600
	for range ticks {
		for range 2 {
			seq++
			cq.push(Input{Seq: seq})
		}
		cq.drain(func(Input) { executed++ })
	}
	if executed > ticks+CmdBurst {
		t.Errorf("executed %d commands in %d ticks", executed, ticks)
	}
	if len(cq.q) > MaxQueuedCmds {
		t.Errorf("queue grew to %d", len(cq.q))
	}
}

func TestSpawnPosAvoidsActors(t *testing.T) {
	s := &Server{W: NewWorld(true)}
	var spawned []*Actor
	for range 5 {
		spawned = append(spawned, s.W.Spawn(s.W.spawnPos(), true))
	}
	for i, a := range spawned {
		for _, b := range spawned[i+1:] {
			ra, rb := a.collisionRect(), b.collisionRect()
			if ra.X < rb.X+rb.Width && ra.X+ra.Width > rb.X {
				t.Errorf("actors %d and %d spawned overlapping at x=%v and x=%v", a.ID, b.ID, a.Pos.X, b.Pos.X)
			}
		}
	}
}

func TestCmdQueueCatchUpAfterStall(t *testing.T) {
	var cq cmdQueue
	for range 20 {
		cq.drain(func(Input) {}) // stalled: tokens accumulate up to the burst
	}
	for s := uint32(1); s <= 12; s++ {
		cq.push(Input{Seq: s})
	}
	n := 0
	cq.drain(func(Input) { n++ })
	if n != CmdBurst {
		t.Errorf("after a stall executed %d, want burst %d", n, CmdBurst)
	}
}
