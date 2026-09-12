package resistor

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"slices"
)

// World is one simulation of actors: the server's, the offline game's or a client's view.
type World struct {
	Actors    []*Actor
	Tick      uint32
	Authority bool // resolves hits and blocks; false for a client's predicted world
	Local     *Actor
	History   *LagHistory // server only
	nextID    uint16
	pending   []pendingSpawn
	dying     []*Actor
}

var SV, CL *World

// OnFighterRemoved and OnFighterRespawned let a server keep its own registries
// in step with deaths and respawns. Both are nil in an offline game, which
// keeps no registries of its own.
var (
	OnFighterRemoved   func(a *Actor)
	OnFighterRespawned func(owner uint16, a *Actor)
)

// pendingSpawn is a fighter due back after a death.
type pendingSpawn struct {
	due      uint32
	isPlayer bool
	owner    uint16
}

func NewWorld(authority bool) *World {
	return &World{Authority: authority}
}

func (w *World) Spawn(pos rl.Vector2, isPlayer bool) *Actor {
	w.nextID++
	a := NewActor(pos, isPlayer)
	a.ID = w.nextID
	w.Add(a)
	return a
}

func (w *World) Add(a *Actor) {
	a.W = w
	w.Actors = append(w.Actors, a)
}

func (w *World) Remove(a *Actor) {
	w.Actors = slices.DeleteFunc(w.Actors, func(b *Actor) bool { return b == a })
	for _, b := range w.Actors {
		if b.Attacker == a {
			b.Attacker = nil
		}
	}
	if w.Local == a {
		w.Local = nil
	}
	if w.History != nil {
		w.History.Remove(a)
	}
}

// Kill books a dead fighter's removal and its return. The removal waits for the
// end of the tick: a fighter dies inside its own Update, and taking it out of
// the world there would cut the actor loop off at the knees.
// Only the authority decides this; a client is told by the next snapshot.
func (w *World) Kill(a *Actor) {
	if !w.Authority || a.Dying {
		return
	}
	a.Dying = true
	w.dying = append(w.dying, a)
	w.pending = append(w.pending, pendingSpawn{due: w.Tick + RespawnTicks, isPlayer: a.IsPlayer, owner: a.Owner})
}

// reap takes the fighters that died during the last tick out of the world.
func (w *World) reap() {
	for _, a := range w.dying {
		if OnFighterRemoved != nil {
			OnFighterRemoved(a)
		}
		w.Remove(a)
	}
	w.dying = w.dying[:0]
}

// updateRespawns brings back the fighters whose time is up, at a free spawn
// position and with everything the old fighter had reset.
func (w *World) updateRespawns() {
	if !w.Authority {
		return
	}
	w.reap()
	if len(w.pending) == 0 {
		return
	}
	kept := w.pending[:0]
	for _, s := range w.pending {
		if w.Tick < s.due {
			kept = append(kept, s)
			continue
		}
		a := w.Spawn(w.spawnPos(), s.isPlayer)
		a.Owner = s.owner
		if OnFighterRespawned != nil {
			OnFighterRespawned(s.owner, a)
		}
	}
	w.pending = kept
}

var spawnXs = []float32{0, 50, -50, 100, -100, 150, -150, 200, -200, -250, -300, -350, -400, -450, -500}

// spawnPos picks the first spawn column not overlapped by another actor's collision box.
func (w *World) spawnPos() rl.Vector2 {
	for _, x := range spawnXs {
		lo, hi := x+14, x+28
		free := true
		for _, a := range w.Actors {
			r := a.collisionRect()
			if r.X < hi && r.X+r.Width > lo {
				free = false
				break
			}
		}
		if free {
			return rl.Vector2{X: x}
		}
	}
	return rl.Vector2{}
}

func (w *World) ByID(id uint16) *Actor {
	for _, a := range w.Actors {
		if a.ID == id {
			return a
		}
	}
	return nil
}
