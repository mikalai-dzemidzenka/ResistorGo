package game

import (
	"cmp"
	"slices"
)

type EntityType int16

const (
	ENTITY_NO EntityType = iota
	ENTITY_PLAYER
	ENTITY_SOLID
	ENTITY_BG
)

type Drawable interface {
	Z() int8
	Draw()
}

type Updatable interface {
	Update(float32)
}

type GUIDrawable interface {
	GuiZ() int8
	DrawGui()
}

var Entities []any

type idx struct {
	z    int8
	draw func()
}

var Drawables []idx
var Updaters []func(float32)
var DrawGUIs []idx

func SortEntities() {
	for i := range Entities {
		{
			u, ok := Entities[i].(Updatable)
			if ok {
				Updaters = append(Updaters, u.Update)
			}
		}
		{
			d, ok := Entities[i].(Drawable)
			if ok {
				Drawables = append(Drawables, idx{d.Z(), d.Draw})
			}
		}
		{
			dg, ok := Entities[i].(GUIDrawable)
			if ok {
				DrawGUIs = append(DrawGUIs, idx{dg.GuiZ(), dg.DrawGui})
			}
		}
	}
	{
		slices.SortFunc(Drawables, func(a, b idx) int {
			return cmp.Compare(a.z, b.z)
		})
	}
	{
		slices.SortFunc(DrawGUIs, func(a, b idx) int {
			return cmp.Compare(a.z, b.z)
		})
	}
}
