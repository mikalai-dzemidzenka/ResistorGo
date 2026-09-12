package resistor

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"math"
)

const (
	LagHistoryLen   = 64 // ticks, > 1 s at 60 Hz
	DefaultMaxUnlag = 15 // ticks, 250 ms
)

type lagRec struct {
	Tick   uint32
	Pos    rl.Vector2
	Box    rl.Rectangle
	State  ActorState
	Updown uint8
	Frame  uint8
	ok     bool
}

// LagHistory keeps recent server-side actor state for rewinding hit tests.
type LagHistory struct {
	m map[*Actor]*[LagHistoryLen]lagRec
}

func NewLagHistory() *LagHistory {
	return &LagHistory{m: make(map[*Actor]*[LagHistoryLen]lagRec)}
}

// Record stores every actor's state for the world's current tick.
func (h *LagHistory) Record(w *World) {
	for _, a := range w.Actors {
		ring := h.m[a]
		if ring == nil {
			ring = new([LagHistoryLen]lagRec)
			h.m[a] = ring
		}
		ring[w.Tick%LagHistoryLen] = lagRec{
			Tick:   w.Tick,
			Pos:    a.Pos,
			Box:    a.CollisionBox,
			State:  a.State,
			Updown: a.Updown,
			Frame:  a.AnimCurrentFrame,
			ok:     true,
		}
	}
}

func (h *LagHistory) Remove(a *Actor) {
	delete(h.m, a)
}

// at returns a's recorded state at fractional tick t. Positions are interpolated
// between ticks; the tick after the newest record is a's live position.
func (h *LagHistory) at(a *Actor, t float64, now uint32) (lagRec, bool) {
	ring := h.m[a]
	if ring == nil || t < 0 {
		return lagRec{}, false
	}
	t0 := uint32(math.Floor(t))
	frac := float32(t - float64(t0))
	r0 := ring[t0%LagHistoryLen]
	if !r0.ok || r0.Tick != t0 {
		return lagRec{}, false
	}
	if frac == 0 {
		return r0, true
	}
	p1 := a.Pos
	if t0+1 < now {
		r1 := ring[(t0+1)%LagHistoryLen]
		if !r1.ok || r1.Tick != t0+1 {
			return r0, true
		}
		p1 = r1.Pos
	}
	r0.Pos = rl.Vector2{
		X: r0.Pos.X + (p1.X-r0.Pos.X)*frac,
		Y: r0.Pos.Y + (p1.Y-r0.Pos.Y)*frac,
	}
	return r0, true
}

type lagSave struct {
	a      *Actor
	pos    rl.Vector2
	updown uint8
}

// rewindOthers moves every actor except attacker to its state at tick t.
// Only position and guard height are rewound; State stays live so eligibility
// checks use the current state. Undo with restore.
func (h *LagHistory) rewindOthers(w *World, attacker *Actor, t float64) []lagSave {
	if t >= float64(w.Tick) {
		return nil
	}
	var saves []lagSave
	for _, a := range w.Actors {
		if a == attacker {
			continue
		}
		r, ok := h.at(a, t, w.Tick)
		if !ok {
			continue
		}
		saves = append(saves, lagSave{a: a, pos: a.Pos, updown: a.Updown})
		a.Pos = r.Pos
		a.Updown = r.Updown
	}
	return saves
}

func restore(saves []lagSave) {
	for _, s := range saves {
		s.a.Pos = s.pos
		s.a.Updown = s.updown
	}
}

// clampViewTime limits how far back the server rewinds: never older than
// maxUnlag ticks and never newer than now.
func clampViewTime(view float64, now uint32, maxUnlag uint32) float64 {
	lo := float64(now) - float64(maxUnlag)
	if lo < 0 {
		lo = 0
	}
	return clamp(view, lo, float64(now))
}
