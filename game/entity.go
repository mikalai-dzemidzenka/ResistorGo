package game

import (
	"cmp"
	"encoding/json"
	"os"
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
	GetZ() int8
	Draw()
}

type Updatable interface {
	Update(float32)
}

type Serializable interface {
	json.Marshaler
}

type GUIDrawable interface {
	GetGUIZ() int8
	DrawGUI()
}

type NoSerialize interface {
	NoSerialize()
}

var Entities map[int]any

func SaveEntities(path string) {
	entities := make([]any, 0, len(Entities))
	for _, e := range Entities {
		if _, ok := e.(NoSerialize); ok {
			continue
		}
		entities = append(entities, e)
	}
	b, err := json.MarshalIndent(entities, "", "  ")
	if err != nil {
		panic(err)
	}
	f, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	_, err = f.Write(b)
	if err != nil {
		panic(err)
	}
}

var Drawables []Drawable
var Updaters []Updatable
var DrawGUIs []GUIDrawable

func Update(dt float32) {
	for i := range Updaters {
		Updaters[i].Update(dt)
	}
}

func Draw() {
	for i := range Drawables {
		Drawables[i].Draw()
	}
}

func DrawGUI() {
	for i := range DrawGUIs {
		DrawGUIs[i].DrawGUI()
	}
}

func EvalEntities() {
	for i := range Entities {
		{
			u, ok := Entities[i].(Updatable)
			if ok {
				Updaters = append(Updaters, u)
			}
		}
		{
			d, ok := Entities[i].(Drawable)
			if ok {
				Drawables = append(Drawables, d)
			}
		}
		{
			dg, ok := Entities[i].(GUIDrawable)
			if ok {
				DrawGUIs = append(DrawGUIs, dg)
			}
		}
	}

	slices.SortFunc(Drawables, func(a, b Drawable) int {
		return cmp.Compare(a.GetZ(), b.GetZ())
	})

	slices.SortFunc(DrawGUIs, func(a, b GUIDrawable) int {
		return cmp.Compare(a.GetGUIZ(), b.GetGUIZ())
	})
}
