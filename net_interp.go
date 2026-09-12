package resistor

import (
	"cmp"
	rl "github.com/gen2brain/raylib-go/raylib"
	"math"
	"slices"
)

const (
	InterpBufferLen = 32
	InterpDelaySec  = 0.1
	MaxExtrapTicks  = 15 // 250 ms
	interpMaxRate   = 0.05
	interpResync    = 30 // ticks behind before the clock jumps forward
	teleportSlack   = 16
)

func interpDelayTicks() float64 {
	return max(InterpDelaySec, 2*SnapshotEvery*FixedTimeStep) / FixedTimeStep
}

// Interp buffers snapshots and keeps a smoothed estimate of the server clock.
type Interp struct {
	Snaps   []Snapshot // ascending by tick
	Est     float64    // estimated current server tick
	Delay   float64    // interpolation delay in ticks
	latest  uint32
	since   float64 // seconds since the newest snapshot arrived
	started bool
}

type InterpState struct {
	Rec   ActorRec // discrete fields, from the older snapshot
	Pos   rl.Vector2
	Ticks float64 // interpolated actor tick counter (animation progress)
}

func (ip *Interp) Add(s Snapshot) {
	i, found := slices.BinarySearchFunc(ip.Snaps, s.Tick, func(e Snapshot, t uint32) int { return cmp.Compare(e.Tick, t) })
	if found {
		return
	}
	ip.Snaps = slices.Insert(ip.Snaps, i, s)
	if n := len(ip.Snaps); n > InterpBufferLen {
		ip.Snaps = slices.Delete(ip.Snaps, 0, n-InterpBufferLen)
	}
	if !ip.started {
		ip.started = true
		ip.Est = float64(s.Tick)
		ip.latest = s.Tick
		ip.since = 0
		return
	}
	if s.Tick > ip.latest {
		ip.latest = s.Tick
		ip.since = 0
	}
}

// Advance moves the clock estimate forward by dt seconds. It runs at most 5 %
// faster or slower than real time to converge on the arrival-based target, and
// never moves backwards.
func (ip *Interp) Advance(dt float64) {
	if !ip.started {
		return
	}
	ip.since += dt
	target := float64(ip.latest) + ip.since/FixedTimeStep
	diff := target - ip.Est
	if diff > interpResync {
		ip.Est = target
		return
	}
	ip.Est += dt / FixedTimeStep * (1 + clamp(diff*0.01, -interpMaxRate, interpMaxRate))
}

func (ip *Interp) RenderTick() float64 {
	return ip.Est - ip.Delay
}

// IDs lists every actor id present in the buffer.
func (ip *Interp) IDs() []uint16 {
	var ids []uint16
	for _, s := range ip.Snaps {
		for _, r := range s.Actors {
			if !slices.Contains(ids, r.ID) {
				ids = append(ids, r.ID)
			}
		}
	}
	return ids
}

func findRec(s *Snapshot, id uint16) (ActorRec, bool) {
	for _, r := range s.Actors {
		if r.ID == id {
			return r, true
		}
	}
	return ActorRec{}, false
}

func roundPos(x, y float64) rl.Vector2 {
	return rl.Vector2{X: float32(math.Round(x)), Y: float32(math.Round(y))}
}

// Sample returns actor id's state at render tick rt, or false if it is not visible then.
func (ip *Interp) Sample(id uint16, rt float64) (InterpState, bool) {
	j := 0
	for j < len(ip.Snaps) && float64(ip.Snaps[j].Tick) <= rt {
		j++
	}
	if j == 0 {
		return InterpState{}, false // rt is older than anything buffered
	}
	a := &ip.Snaps[j-1]
	ra, ok := findRec(a, id)
	if !ok {
		return InterpState{}, false // not spawned yet (or already gone) at rt
	}
	if j == len(ip.Snaps) {
		// no newer data: extrapolate from velocity, then hold
		gap := min(rt-float64(a.Tick), MaxExtrapTicks)
		t := gap * FixedTimeStep
		return InterpState{
			Rec:   ra,
			Pos:   roundPos(float64(ra.X)+float64(ra.SX)*t, float64(ra.Y)+float64(ra.SY)*t),
			Ticks: float64(ra.Ticks) + gap,
		}, true
	}
	b := &ip.Snaps[j]
	rb, ok := findRec(b, id)
	if !ok {
		if rt > float64(a.Tick) {
			return InterpState{}, false // removed after a
		}
		return InterpState{Rec: ra, Pos: roundPos(float64(ra.X), float64(ra.Y)), Ticks: float64(ra.Ticks)}, true
	}
	span := float64(b.Tick - a.Tick)
	t := (rt - float64(a.Tick)) / span
	x := float64(ra.X) + float64(rb.X-ra.X)*t
	y := float64(ra.Y) + float64(rb.Y-ra.Y)*t
	if isTeleport(ra, rb, span) {
		x, y = float64(ra.X), float64(ra.Y)
	}
	return InterpState{
		Rec:   ra,
		Pos:   roundPos(x, y),
		Ticks: float64(ra.Ticks) + (float64(rb.Ticks)-float64(ra.Ticks))*t,
	}, true
}

// isTeleport reports whether an actor moved further between two snapshots than it could travel.
func isTeleport(a, b ActorRec, ticks float64) bool {
	dx, dy := float64(b.X-a.X), float64(b.Y-a.Y)
	limit := MaxFallSpeed*ticks*FixedTimeStep + teleportSlack
	return dx*dx+dy*dy > limit*limit
}
