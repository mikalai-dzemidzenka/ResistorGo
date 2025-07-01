package game

import (
	"encoding/json"
	"reflect"
	"strconv"
)

type EntityType int16

const (
	ENTITY_NO EntityType = iota
	ENTITY_PLAYER
	ENTITY_SOLID
	ENTITY_BG
	ENTITY_LEN
)

type ComponentType int16

const (
	COMPONENT_TYPE_POSITION ComponentType = iota
	COMPONENT_TYPE_TEXTURE
)

type EntityInfo struct {
	Type           reflect.Type
	ComponentTypes []ComponentType
}

var EntityRegistry = map[EntityType]EntityInfo{
	ENTITY_PLAYER: {
		Type: reflect.TypeOf(Player{}),
		ComponentTypes: []ComponentType{
			COMPONENT_TYPE_POSITION,
			COMPONENT_TYPE_TEXTURE,
		},
	},
	ENTITY_SOLID: {
		Type: reflect.TypeOf(Solid{}),
		ComponentTypes: []ComponentType{
			COMPONENT_TYPE_POSITION,
			COMPONENT_TYPE_TEXTURE,
		},
	},
}

func HasComponent(t EntityType, want ComponentType) bool {
	if inf, ok := EntityRegistry[t]; ok {
		for _, got := range inf.ComponentTypes {
			if got == want {
				return true
			}
		}
	}
	return false
}

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
		info, ok := EntityRegistry[s.Type]
		if !ok {
			panic("missing type" + strconv.Itoa(int(s.Type)))
		}
		v := reflect.New(info.Type)
		iface := v.Interface()
		serde[i].Entity = iface
	}
	err = json.Unmarshal(bytes, &serde)
	if err != nil {
		return err
	}

	*r = make(RuntimeEntities)
	for i, v := range serde {
		(*r)[i] = v.Entity
	}
	return nil
}

var _ json.Unmarshaler = (*RuntimeEntities)(nil)

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
