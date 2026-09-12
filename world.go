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
}

var SV, CL *World

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

func (w *World) ByID(id uint16) *Actor {
	for _, a := range w.Actors {
		if a.ID == id {
			return a
		}
	}
	return nil
}
