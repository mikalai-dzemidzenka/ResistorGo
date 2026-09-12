package resistor

import (
	"math"
	"math/rand"
	"testing"
)

func irec(id uint16, x, y float32) ActorRec {
	return ActorRec{ID: id, X: x, Y: y}
}

func isnap(tick uint32, recs ...ActorRec) Snapshot {
	return Snapshot{Tick: tick, Actors: recs}
}

func TestInterpBracketAndRound(t *testing.T) {
	var ip Interp
	a := irec(1, 0, 0)
	a.State, a.Ticks = uint8(StateRun), 100
	b := irec(1, 10, 0)
	b.State, b.Ticks = uint8(StateAttack), 103
	ip.Add(isnap(10, a))
	ip.Add(isnap(13, b))

	st, ok := ip.Sample(1, 11.5)
	if !ok || st.Pos.X != 5 || st.Rec.State != uint8(StateRun) || st.Ticks != 101.5 {
		t.Errorf("mid: %+v %v", st, ok)
	}
	st, _ = ip.Sample(1, 11)
	if st.Pos.X != 3 { // 3.33 rounded
		t.Errorf("rounding: got %v", st.Pos.X)
	}
	if _, ok := ip.Sample(1, 9); ok {
		t.Error("visible before the oldest snapshot")
	}
}

func TestInterpExtrapolation(t *testing.T) {
	var ip Interp
	r := irec(1, 0, 0)
	r.SX = 60 // px/s = 1 px per tick
	ip.Add(isnap(10, r))

	if st, _ := ip.Sample(1, 19); st.Pos.X != 9 { // 150 ms gap keeps moving
		t.Errorf("150 ms: x=%v", st.Pos.X)
	}
	for _, rt := range []float64{10 + MaxExtrapTicks, 40, 130} { // 2 s gap holds after 250 ms
		if st, _ := ip.Sample(1, rt); st.Pos.X != MaxExtrapTicks {
			t.Errorf("rt %v: x=%v, want %v", rt, st.Pos.X, MaxExtrapTicks)
		}
	}
}

func TestInterpTeleport(t *testing.T) {
	var ip Interp
	ip.Add(isnap(10, irec(1, 0, 0)))
	ip.Add(isnap(13, irec(1, 500, 0)))
	if st, _ := ip.Sample(1, 11.5); st.Pos.X != 0 {
		t.Errorf("teleport interpolated: x=%v", st.Pos.X)
	}
	if st, _ := ip.Sample(1, 13); st.Pos.X != 500 {
		t.Errorf("after teleport: x=%v", st.Pos.X)
	}
	// a fast fall is not a teleport
	ip = Interp{}
	ip.Add(isnap(10, irec(1, 0, 0)))
	ip.Add(isnap(13, irec(1, 0, 45)))
	if st, _ := ip.Sample(1, 11.5); st.Pos.Y == 0 {
		t.Error("fall treated as teleport")
	}
}

func TestInterpSpawnAndRemove(t *testing.T) {
	var ip Interp
	ip.Add(isnap(10, irec(1, 0, 0)))
	ip.Add(isnap(13, irec(1, 0, 0), irec(2, 50, 0)))
	ip.Add(isnap(16, irec(2, 50, 0)))

	if _, ok := ip.Sample(2, 11); ok {
		t.Error("actor 2 visible before its first snapshot")
	}
	if st, ok := ip.Sample(2, 13); !ok || st.Pos.X != 50 {
		t.Errorf("actor 2 at spawn: %+v %v", st, ok)
	}
	if _, ok := ip.Sample(1, 12); !ok {
		t.Error("actor 1 hidden before its last snapshot")
	}
	if _, ok := ip.Sample(1, 14); ok {
		t.Error("actor 1 visible after it was removed")
	}
}

func TestInterpClockSmoothUnderJitter(t *testing.T) {
	ip := Interp{Delay: interpDelayTicks()}
	rnd := rand.New(rand.NewSource(1))
	type arrival struct {
		tick uint32
		at   float64
	}
	var pending []arrival
	for tick := uint32(3); tick < 600; tick += SnapshotEvery {
		pending = append(pending, arrival{tick, float64(tick)*FixedTimeStep + 0.05 + (rnd.Float64()-0.5)*0.04})
	}
	const dt = 1.0 / 60
	prev := math.Inf(-1)
	for frame := 0; frame < 560; frame++ {
		now := float64(frame) * dt
		for len(pending) > 0 && pending[0].at <= now {
			ip.Add(isnap(pending[0].tick))
			pending = pending[1:]
		}
		ip.Advance(dt)
		if !ip.started {
			continue
		}
		if prev != math.Inf(-1) {
			step := ip.Est - prev
			if step < 0 {
				t.Fatalf("frame %d: clock went backwards by %v", frame, -step)
			}
			if frame > 30 && (step < 0.94 || step > 1.06) {
				t.Fatalf("frame %d: clock step %v ticks, want within 5%%", frame, step)
			}
		}
		prev = ip.Est
	}
	// the estimate tracks the server clock minus the one-way latency (~3 ticks)
	if want := 560.0 - 3; math.Abs(ip.Est-want) > 3 {
		t.Errorf("estimate %v, want about %v", ip.Est, want)
	}
}
