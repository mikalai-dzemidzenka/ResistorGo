package resistor

import (
	"encoding/json"
	"os"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// objectDefaultsPath holds the rectangles the editor's palette entries are born
// with, keyed by texture path. It sits beside the level rather than inside it:
// the level's fingerprint is a network contract, and the editor's defaults have
// no business changing it. It is not embedded either - only the editor reads it,
// and a release build has no editor.
const objectDefaultsPath = "level_objects.json"

// ObjectRects is an object's two rectangles: what fighters collide with, and
// what casts a shadow. A nil one means the object has neither - a texture that
// should not be stood on, or should not cast a shadow, says so this way.
type ObjectRects struct {
	Collision *rl.Rectangle `json:"collisionRect,omitempty"`
	Light     *rl.Rectangle `json:"lightRect,omitempty"`
}

// clone copies the rectangles themselves, not the pointers to them, so a tile
// placed from a palette entry does not share them with the entry.
func (o ObjectRects) clone() ObjectRects {
	return ObjectRects{Collision: cloneRect(o.Collision), Light: cloneRect(o.Light)}
}

// sameRects reports whether two objects have the same rectangles: a missing
// rectangle matches only a missing one, and two present ones match by value.
func sameRects(a, b ObjectRects) bool {
	return sameRect(a.Collision, b.Collision) && sameRect(a.Light, b.Light)
}

func sameRect(a, b *rl.Rectangle) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}

func cloneRect(r *rl.Rectangle) *rl.Rectangle {
	if r == nil {
		return nil
	}
	c := *r
	return &c
}

func (o LevelObject) rects() ObjectRects {
	return ObjectRects{Collision: o.CollisionRect, Light: o.LightRect}
}

func (o *LevelObject) setRects(r ObjectRects) {
	o.CollisionRect = r.Collision
	o.LightRect = r.Light
}

// loadObjectDefaults reads the palette's rectangles. Anything missing or
// malformed falls back to the rectangles built into the game: bad editor data
// must never stop the editor from opening.
func loadObjectDefaults(path string) map[string]ObjectRects {
	if !DevTools {
		return nil // a release build has no editor to have a palette for
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	m := make(map[string]ObjectRects)
	if err := json.Unmarshal(b, &m); err != nil {
		rl.TraceLog(rl.LogError, "failed to load object defaults: json unmarshal: %s", err.Error())
		return nil
	}
	return m
}

// saveObjectDefaults writes the palette's rectangles. Like the level, it is
// written only by a development build, and only when asked. Nothing edited
// writes nothing, so saving a level does not litter the repo with an empty file.
// It reports whether the defaults are safely on disk.
func saveObjectDefaults(path string, defaults map[string]ObjectRects) bool {
	if !DevTools {
		return false
	}
	if len(defaults) == 0 {
		return true
	}
	b, err := json.MarshalIndent(defaults, "", "  ")
	if err != nil {
		rl.TraceLog(rl.LogError, "failed to save object defaults: marshal: %s", err.Error())
		return false
	}
	if err := os.WriteFile(path, b, 0o644); err != nil {
		rl.TraceLog(rl.LogError, "failed to save object defaults: write file: %s", err.Error())
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
