package resistor

import (
	"encoding/json"
	"os"
	"slices"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// objectDefaultsPath holds the rectangles the editor's palette entries are born
// with, keyed by texture path. It sits beside the level rather than inside it:
// the level's fingerprint is a network contract, and the editor's defaults have
// no business changing it. It is not embedded either - only the editor reads it,
// and a release build has no editor.
const objectDefaultsPath = "level_objects.json"

// ObjectRects is an object's rectangles: the ones fighters collide with, the
// ones that cast a shadow, and the drop-through ones - floor from above that a
// fighter can climb down through - any number of each, in order. No collision
// rectangles, or no light rectangles, is how a texture says it should not be
// stood on, or should not cast a shadow.
type ObjectRects struct {
	Collision []rl.Rectangle
	Light     []rl.Rectangle
	Drop      []rl.Rectangle
}

// clone copies the lists themselves, so a tile placed from a palette entry does
// not share them with the entry.
func (o ObjectRects) clone() ObjectRects {
	return ObjectRects{Collision: slices.Clone(o.Collision), Light: slices.Clone(o.Light), Drop: slices.Clone(o.Drop)}
}

// sameRects reports whether two objects have the same rectangles: the same
// collision, light and drop-through rectangles, each in the same order.
func sameRects(a, b ObjectRects) bool {
	return slices.Equal(a.Collision, b.Collision) && slices.Equal(a.Light, b.Light) && slices.Equal(a.Drop, b.Drop)
}

func (o LevelObject) rects() ObjectRects {
	return ObjectRects{Collision: o.CollisionRects, Light: o.LightRects, Drop: o.DropRects}
}

func (o *LevelObject) setRects(r ObjectRects) {
	o.CollisionRects = r.Collision
	o.LightRects = r.Light
	o.DropRects = r.Drop
}

// loadObjectDefaults reads the palette's rectangles. Anything missing or
// malformed falls back to the rectangles built into the game: bad editor data
// must never stop the editor from opening.
func loadObjectDefaults(path string) map[string]ObjectRects {
	m := make(map[string]ObjectRects)
	if !readEditorJSON(path, "object defaults", &m) {
		return nil
	}
	return m
}

// saveObjectDefaults writes the palette's rectangles. Like the level, it is
// written only by a development build, and only when asked. Nothing edited
// writes nothing, so saving a level does not litter the repo with an empty file.
// It reports whether the defaults are safely on disk.
func saveObjectDefaults(path string, defaults map[string]ObjectRects) bool {
	if len(defaults) == 0 {
		return DevTools
	}
	return writeEditorJSON(path, "object defaults", defaults)
}

// presetsPath holds the editor's rectangle presets: slots of rectangles that
// any palette entry or placed tile can be given with one key. Like the
// defaults, only the editor reads it, and only Ctrl+S writes it.
const presetsPath = "level_presets.json"

const presetSlots = 9 // keys 1 to 9

// Presets are the slots, nil where nothing has been saved.
type Presets [presetSlots]*ObjectRects

// loadPresets reads the presets; anything missing or malformed is no presets.
func loadPresets(path string) Presets {
	var p Presets
	if !readEditorJSON(path, "presets", &p) {
		return Presets{}
	}
	return p
}

// savePresets writes the presets. No preset writes nothing.
func savePresets(path string, p Presets) bool {
	if p == (Presets{}) {
		return DevTools
	}
	return writeEditorJSON(path, "presets", p)
}

// readEditorJSON reads one of the editor's own files into v, reporting whether
// it did. Only a development build has an editor to read them.
func readEditorJSON(path, what string, v any) bool {
	if !DevTools {
		return false
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	if err := json.Unmarshal(b, v); err != nil {
		rl.TraceLog(rl.LogError, "failed to load %s: json unmarshal: %s", what, err.Error())
		return false
	}
	return true
}

// writeEditorJSON writes one of the editor's own files, reporting whether it is
// safely on disk.
func writeEditorJSON(path, what string, v any) bool {
	if !DevTools {
		return false
	}
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		rl.TraceLog(rl.LogError, "failed to save %s: marshal: %s", what, err.Error())
		return false
	}
	if err := os.WriteFile(path, b, 0o644); err != nil {
		rl.TraceLog(rl.LogError, "failed to save %s: write file: %s", what, err.Error())
		return false
	}
	return true
}

// builtinRects are the rectangles compiled into the game, by texture path. They
// are what a texture starts with when the defaults file says nothing about it.
func builtinRects() map[string]ObjectRects {
	m := make(map[string]ObjectRects)
	for _, obj := range LoadLevelObjects() {
		m[obj.Resource] = obj.rects()
	}
	return m
}

// buildPalette is every placeable block texture, each with the rectangles it is
// born with: the defaults file first, then what is built into the game, then
// none at all. It loads no textures, so it is safe before there is a window.
func buildPalette(defaults map[string]ObjectRects) []LevelObject {
	builtin := builtinRects()
	paths := blockTextures()
	palette := make([]LevelObject, 0, len(paths))
	for _, p := range paths {
		rects, ok := defaults[p]
		if !ok {
			rects = builtin[p]
		}
		obj := LevelObject{Resource: p}
		obj.setRects(rects.clone())
		palette = append(palette, obj)
	}
	return palette
}

// On disk one rectangle of a kind is "collisionRect" or "lightRect", exactly as
// it was before objects could have several, so a level or defaults file with no
// more than one of each per object keeps its bytes and its fingerprint. Two or
// more are "collisionRects" or "lightRects", and none is neither.
func rectsToWire(rs []rl.Rectangle) (*rl.Rectangle, []rl.Rectangle) {
	switch len(rs) {
	case 0:
		return nil, nil
	case 1:
		r := rs[0]
		return &r, nil
	}
	return nil, rs
}

// rectsFromWire is the list a kind's two keys give; key is the kind's
// one-rectangle key, for the warning.
func rectsFromWire(one *rl.Rectangle, many []rl.Rectangle, key string) []rl.Rectangle {
	if many != nil {
		if one != nil {
			rl.TraceLog(rl.LogWarning, "level: both %[1]s and %[1]ss; keeping %[1]ss", key)
		}
		return many
	}
	if one != nil {
		return []rl.Rectangle{*one}
	}
	return nil
}

// levelObjectWire is a LevelObject as the level file holds it, in the order its
// fields have always been written.
type levelObjectWire struct {
	Resource       string         `json:"resource,omitempty"`
	Layer          int            `json:"layer,omitempty"`
	X              int32          `json:"x,omitempty"`
	Y              int32          `json:"y,omitempty"`
	CollisionRect  *rl.Rectangle  `json:"collisionRect,omitempty"`
	CollisionRects []rl.Rectangle `json:"collisionRects,omitempty"`
	LightRect      *rl.Rectangle  `json:"lightRect,omitempty"`
	LightRects     []rl.Rectangle `json:"lightRects,omitempty"`
	DropRect       *rl.Rectangle  `json:"dropRect,omitempty"`
	DropRects      []rl.Rectangle `json:"dropRects,omitempty"`
}

func (o LevelObject) MarshalJSON() ([]byte, error) {
	w := levelObjectWire{Resource: o.Resource, Layer: o.Layer, X: o.X, Y: o.Y}
	w.CollisionRect, w.CollisionRects = rectsToWire(o.CollisionRects)
	w.LightRect, w.LightRects = rectsToWire(o.LightRects)
	w.DropRect, w.DropRects = rectsToWire(o.DropRects)
	return json.Marshal(w)
}

func (o *LevelObject) UnmarshalJSON(b []byte) error {
	var w levelObjectWire
	if err := json.Unmarshal(b, &w); err != nil {
		return err
	}
	*o = LevelObject{
		Resource:       w.Resource,
		Layer:          w.Layer,
		X:              w.X,
		Y:              w.Y,
		CollisionRects: rectsFromWire(w.CollisionRect, w.CollisionRects, "collisionRect"),
		LightRects:     rectsFromWire(w.LightRect, w.LightRects, "lightRect"),
		DropRects:      rectsFromWire(w.DropRect, w.DropRects, "dropRect"),
	}
	return nil
}

// objectRectsWire is an ObjectRects as the texture defaults file holds it.
type objectRectsWire struct {
	CollisionRect  *rl.Rectangle  `json:"collisionRect,omitempty"`
	CollisionRects []rl.Rectangle `json:"collisionRects,omitempty"`
	LightRect      *rl.Rectangle  `json:"lightRect,omitempty"`
	LightRects     []rl.Rectangle `json:"lightRects,omitempty"`
	DropRect       *rl.Rectangle  `json:"dropRect,omitempty"`
	DropRects      []rl.Rectangle `json:"dropRects,omitempty"`
}

func (o ObjectRects) MarshalJSON() ([]byte, error) {
	var w objectRectsWire
	w.CollisionRect, w.CollisionRects = rectsToWire(o.Collision)
	w.LightRect, w.LightRects = rectsToWire(o.Light)
	w.DropRect, w.DropRects = rectsToWire(o.Drop)
	return json.Marshal(w)
}

func (o *ObjectRects) UnmarshalJSON(b []byte) error {
	var w objectRectsWire
	if err := json.Unmarshal(b, &w); err != nil {
		return err
	}
	*o = ObjectRects{
		Collision: rectsFromWire(w.CollisionRect, w.CollisionRects, "collisionRect"),
		Light:     rectsFromWire(w.LightRect, w.LightRects, "lightRect"),
		Drop:      rectsFromWire(w.DropRect, w.DropRects, "dropRect"),
	}
	return nil
}
