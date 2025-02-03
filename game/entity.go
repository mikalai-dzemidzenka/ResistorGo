package game

import (
	"cmp"
	"encoding/json"
	"os"
	"reflect"
	"slices"
	"strconv"
)

var Entities = RuntimeEntities{}

type RuntimeEntities map[int]any

type EntitySerde struct {
	Type   EntityType `json:"type"`
	Entity any        `json:"entity"`
}

func (r *RuntimeEntities) UnmarshalJSON(bytes []byte) error {
	var serde []EntitySerde
	err := json.Unmarshal(bytes, &serde)
	if err != nil {
		return err
	}
	for i, s := range serde {
		typ, ok := EntityRegistry[s.Type]
		if !ok {
			panic("missing type" + strconv.Itoa(int(s.Type)))
		}
		v := reflect.New(typ)
		iface := v.Interface()
		serde[i].Entity = iface
	}
	err = json.Unmarshal(bytes, &serde)
	if err != nil {
		return err
	}

	for i, v := range serde {
		(*r)[i] = v.Entity
	}
	return nil
}

var _ json.Unmarshaler = &Entities

var EntityRegistry = map[EntityType]reflect.Type{
	ENTITY_PLAYER:       reflect.TypeOf(Player{}),
	ENTITY_CAMERA:       reflect.TypeOf(Camera{}),
	ENTITY_SOLID:        reflect.TypeOf(Solid{}),
	ENTITY_LEVEL_EDITOR: reflect.TypeOf(LevelEditor{}),
	ENTITY_CURSOR:       reflect.TypeOf(Cursor{}),
}

type EntityType int16

const (
	ENTITY_NO EntityType = iota
	ENTITY_PLAYER
	ENTITY_CAMERA
	ENTITY_SOLID
	ENTITY_BG
	ENTITY_LEVEL_EDITOR
	ENTITY_CURSOR
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

func SaveEntities(path string) {
	type SerEnt struct {
		I      int
		Entity any
	}
	sortEnt := make([]SerEnt, 0, len(Entities))
	for i, e := range Entities {
		sortEnt = append(sortEnt, SerEnt{
			I:      i,
			Entity: e,
		})
	}
	slices.SortFunc(sortEnt, func(a, b SerEnt) int {
		return cmp.Compare(a.I, b.I)
	})
	entities := make([]any, len(sortEnt))
	for i := range sortEnt {
		entities[i] = sortEnt[i].Entity
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

func LoadEntities(path string) {
	b, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	err = Entities.UnmarshalJSON(b)
	if err != nil {
		panic(err)
	}
}

func EvalEntities() {
	for i := range Entities {
		// TODO get rid of this!!!!!!!!!!!
		{
			c, ok := Entities[i].(*Cursor)
			if ok {
				cur = c
			}
		}
		{
			pl, ok := Entities[i].(*Player)
			if ok {
				p = pl
			}
		}
		{
			c, ok := Entities[i].(*Camera)
			if ok {
				cam = c
			}
		}
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
