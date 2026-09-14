package resistor

import (
	"encoding/json"
	"fmt"
	"strconv"

	rlg "github.com/gen2brain/raylib-go/raygui"
	rl "github.com/gen2brain/raylib-go/raylib"
)

// LevelLight is a light placed in the level: a point that lights the circle of
// Radius around it in Color at Intensity, shadowed by the tiles' light
// rectangles. Positions and radius are whole level pixels, as tiles are.
type LevelLight struct {
	X         int32    `json:"x"`
	Y         int32    `json:"y"`
	Radius    int32    `json:"radius"`
	Color     [3]uint8 `json:"color"`
	Intensity float32  `json:"intensity"`
}

// defaultLight is what a light placed with no light selected looks like.
var defaultLight = LevelLight{Radius: 128, Color: [3]uint8{255, 255, 255}, Intensity: 1}

func (l LevelLight) center() rl.Vector2 {
	return rl.Vector2{X: float32(l.X), Y: float32(l.Y)}
}

// handle is the point on the light's circle that its radius is dragged by.
func (l LevelLight) handle() rl.Vector2 {
	return rl.Vector2{X: float32(l.X + l.Radius), Y: float32(l.Y)}
}

func (l LevelLight) color() rl.Color {
	return rl.Color{R: l.Color[0], G: l.Color[1], B: l.Color[2], A: 255}
}

// The level file is an object keyed by layer number, plus one key for the
// lights. It kept its shape when lights were added to it: a level with no
// lights is exactly the file it always was.
const lightsKey = "lights"

// parseLevel reads a level file: every integer key is a layer of tiles, the
// lights key is the lights, and anything else is skipped rather than failing
// the whole level.
func parseLevel(b []byte) (map[int]LevelLayer, []LevelLight, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(b, &raw); err != nil {
		return nil, nil, err
	}
	objects := make(map[int]LevelLayer)
	var lights []LevelLight
	for k, v := range raw {
		if k == lightsKey {
			if err := json.Unmarshal(v, &lights); err != nil {
				return nil, nil, err
			}
			continue
		}
		n, err := strconv.Atoi(k)
		if err != nil {
			rl.TraceLog(rl.LogWarning, "level: skipping unknown key %q", k)
			continue
		}
		var layer LevelLayer
		if err := json.Unmarshal(v, &layer); err != nil {
			return nil, nil, err
		}
		objects[n] = layer
	}
	return objects, lights, nil
}

// marshalFile is the level file as saved: the tiles, as marshal gives them to
// the fingerprint, and the lights beside them.
func (le *LevelEditor) marshalFile() ([]byte, error) {
	m := make(map[string]any)
	for n, layer := range le.objects {
		m[strconv.Itoa(n)] = layer
	}
	if len(le.lights) > 0 {
		m[lightsKey] = le.lights
	}
	return json.Marshal(m)
}

// lightVisible reports whether any of a light's area is in view. A light that
// is only partly in view is drawn whole, so nothing pops at the screen's edge.
func lightVisible(l LevelLight, view rl.Rectangle) bool {
	r := float32(l.Radius)
	return aabb(rl.Rectangle{X: float32(l.X) - r, Y: float32(l.Y) - r, Width: 2 * r, Height: 2 * r}, view)
}

// viewRect is the part of the level a camera shows on a w x h target.
func viewRect(cam rl.Camera2D, w, h float32) rl.Rectangle {
	a := rl.GetScreenToWorld2D(rl.Vector2{}, cam)
	b := rl.GetScreenToWorld2D(rl.Vector2{X: w, Y: h}, cam)
	return rl.Rectangle{X: a.X, Y: a.Y, Width: b.X - a.X, Height: b.Y - a.Y}
}

// computeLights adds up the light of every level light in view into
// LightAccum, to be added to the scene in one draw by drawLightAccum. The
// lights take turns with one set of shadow textures, so each is added up the
// moment it is computed, before the next one overwrites them. Additive blending
// saturates the same whether the lights are added to the scene one by one or
// added up first, so this is what drawing each onto the scene would give.
func computeLights(lights []LevelLight, occluders []rl.Rectangle, cam rl.Camera2D) {
	rl.BeginTextureMode(LightAccum)
	rl.ClearBackground(rl.Black) // opaque: drawLightAccum's blend weighs it by its alpha
	rl.EndTextureMode()
	view := viewRect(cam, float32(LightAccum.Texture.Width), float32(LightAccum.Texture.Height))
	for _, l := range lights {
		if !lightVisible(l, view) {
			continue
		}
		sharedLight.Radius = float32(l.Radius)
		sharedLight.Color = l.color()
		sharedLight.Intensity = l.Intensity
		tex := sharedLight.Compute(Sh, l.center(), occluders)
		rl.BeginTextureMode(LightAccum)
		rl.BeginMode2D(cam)
		rl.BeginBlendMode(rl.BlendAdditive)
		addLight(tex, l.center(), sharedLight.Radius)
		rl.EndBlendMode()
		rl.EndMode2D()
		rl.EndTextureMode()
	}
}

// drawLightAccum adds the level's light to the scene being drawn, in screen
// space - computeLights already put each light where the camera shows it.
func drawLightAccum() {
	w, h := float32(LightAccum.Texture.Width), float32(LightAccum.Texture.Height)
	rl.BeginBlendMode(rl.BlendAdditive)
	rl.DrawTexturePro(LightAccum.Texture,
		rl.NewRectangle(0, 0, w, -h), // флип render-текстуры
		rl.NewRectangle(0, 0, w, h),
		rl.Vector2{}, 0, rl.White)
	rl.EndBlendMode()
}

// What a light drag has hold of.
const (
	lightDragNone = iota
	lightDragMove
	lightDragRadius
)

// lightDrag is a light being dragged: by its center or its radius handle, and
// how far the cursor was from that point when it took hold.
type lightDrag struct {
	mode int
	off  rl.Vector2
}

// lightAt is the light a click at p takes hold of, and by what. Radius
// handles are tried before centers - a handle can lie on another light's
// center - and later lights before earlier ones, as they are drawn on top.
func (le *LevelEditor) lightAt(p rl.Vector2, tol float32) (int, int) {
	grab := max(tol*2, 2)
	for i := len(le.lights) - 1; i >= 0; i-- {
		if distance(p, le.lights[i].handle()) <= grab {
			return i, lightDragRadius
		}
	}
	for i := len(le.lights) - 1; i >= 0; i-- {
		if distance(p, le.lights[i].center()) <= grab {
			return i, lightDragMove
		}
	}
	return -1, lightDragNone
}

// placeLight puts a light at p and selects it. It takes the radius, color and
// intensity of the light selected before it, so a row of matching lights is a
// row of clicks.
func (le *LevelEditor) placeLight(p rl.Vector2) {
	l := defaultLight
	if le.lightSel >= 0 && le.lightSel < len(le.lights) {
		l = le.lights[le.lightSel]
	}
	l.X, l.Y = int32(roundf(p.X)), int32(roundf(p.Y))
	le.lights = append(le.lights, l)
	le.lightSel = len(le.lights) - 1
	le.lightDrag = lightDrag{}
	le.editField = -1
	le.dirty = true
}

// beginLightDrag takes hold of the light under p, selecting it.
func (le *LevelEditor) beginLightDrag(p rl.Vector2, tol float32) bool {
	i, mode := le.lightAt(p, tol)
	if i < 0 {
		return false
	}
	le.lightSel = i
	le.editField = -1
	anchor := le.lights[i].center()
	if mode == lightDragRadius {
		anchor = le.lights[i].handle()
	}
	le.lightDrag = lightDrag{mode: mode, off: sub(anchor, p)}
	return true
}

// dragLight moves what the drag has hold of to the cursor at p, in whole
// pixels: the center moves the light, the handle sets its radius, never below
// a pixel.
func (le *LevelEditor) dragLight(p rl.Vector2) {
	if le.lightSel < 0 || le.lightSel >= len(le.lights) || le.lightDrag.mode == lightDragNone {
		le.lightDrag = lightDrag{}
		return
	}
	to := roundVec2(rl.Vector2{X: p.X + le.lightDrag.off.X, Y: p.Y + le.lightDrag.off.Y})
	l := le.lights[le.lightSel]
	if le.lightDrag.mode == lightDragMove {
		l.X, l.Y = int32(to.X), int32(to.Y)
	} else {
		l.Radius = max(1, int32(to.X)-l.X)
	}
	le.setLight(le.lightSel, l)
}

// setLight replaces light i, if that changes anything.
func (le *LevelEditor) setLight(i int, l LevelLight) {
	if i < 0 || i >= len(le.lights) || le.lights[i] == l {
		return
	}
	le.lights[i] = l
	le.dirty = true
}

// deleteLight removes light i, keeping the selection on the light it was on.
func (le *LevelEditor) deleteLight(i int) {
	if i < 0 || i >= len(le.lights) {
		return
	}
	le.lights = append(le.lights[:i], le.lights[i+1:]...)
	switch {
	case le.lightSel == i:
		le.lightSel = -1
	case le.lightSel > i:
		le.lightSel--
	}
	le.lightDrag = lightDrag{}
	le.editField = -1
	le.dirty = true
}

// updateLightTool is the light tool's handling of the mouse, the cursor being
// at world in the level: a left click on a light takes hold of it, on nothing
// places one, and a right click on a light removes it.
func (le *LevelEditor) updateLightTool(world rl.Vector2) {
	if le.lightDrag.mode != lightDragNone {
		if rl.IsMouseButtonDown(rl.MouseButtonLeft) {
			le.dragLight(world)
		} else {
			le.lightDrag = lightDrag{}
		}
		return
	}
	if le.overGui(MousePos) {
		return
	}
	tol := grabTolerance(Cam.PixelSize())
	switch {
	case rl.IsMouseButtonPressed(rl.MouseButtonRight):
		if i, _ := le.lightAt(world, tol); i >= 0 {
			le.deleteLight(i)
		}
	case rl.IsMouseButtonPressed(rl.MouseButtonLeft):
		if !le.beginLightDrag(world, tol) {
			le.placeLight(world)
		}
	}
}

// markerBacking is the dark edge under the light markers' center cross and
// radius handle.
var markerBacking = rl.Fade(rl.Black, 0.8)

// DrawOverlay draws what the editor puts over the lit scene: the light
// markers, while the light tool is out. They go over the lighting rather than
// under it with the tiles, because a light is brightest at the very center its
// marker is on, and its own glow would wash the marker out.
func (le *LevelEditor) DrawOverlay() {
	if DevTools && Editing && le.tool == toolLight {
		le.drawLights()
	}
}

// drawLights marks every light while the light tool is out: the circle its
// light reaches, a cross on its center and the handle its radius is dragged
// by. The selected light is white, the others their own color, faded. All of it
// is filled shapes a level pixel wide, not GL lines: a line is one screen pixel
// wide at any zoom, and the palette pass, sampling the frame in blocks of the
// zoom, would keep only scattered dots of it.
func (le *LevelEditor) drawLights() {
	for i, l := range le.lights {
		col := rl.Fade(l.color(), 0.5)
		if i == le.lightSel {
			col = rl.White
		}
		c := l.center()
		r := float32(l.Radius)
		rl.DrawRing(c, max(0, r-0.5), r+0.5, 0, 360, clamp(l.Radius/2, 24, 256), col)
		// The center is where a light is brightest - saturated white, for a white
		// light - so the cross and the handle each sit on a dark backing that
		// shows on the glow and on the dark alike.
		const arm = 4
		for _, s := range []struct {
			thick float32
			col   rl.Color
		}{{3, markerBacking}, {1, col}} {
			rl.DrawLineEx(rl.Vector2{X: c.X - arm, Y: c.Y}, rl.Vector2{X: c.X + arm, Y: c.Y}, s.thick, s.col)
			rl.DrawLineEx(rl.Vector2{X: c.X, Y: c.Y - arm}, rl.Vector2{X: c.X, Y: c.Y + arm}, s.thick, s.col)
		}
		h := l.handle()
		rl.DrawRectangleRec(rl.Rectangle{X: h.X - 2, Y: h.Y - 2, Width: 4, Height: 4}, markerBacking)
		rl.DrawRectangleRec(rl.Rectangle{X: h.X - 1, Y: h.Y - 1, Width: 2, Height: 2}, col)
	}
}

// drawLightPanel is the light tool's part of the panel: the selected light's
// numbers, which set it exactly when typed into.
func (le *LevelEditor) drawLightPanel(row func() rl.Rectangle) {
	if le.lightSel < 0 || le.lightSel >= len(le.lights) {
		rlg.Label(row(), fmt.Sprintf("%d lights - LMB to place one", len(le.lights)))
		return
	}
	l := le.lights[le.lightSel]
	rlg.Label(row(), fmt.Sprintf("Light %d of %d", le.lightSel+1, len(le.lights)))
	pct := int32(roundf(l.Intensity * 100))
	v := le.drawIntFields(row, []intField{
		{"X", l.X, -fieldLim, fieldLim},
		{"Y", l.Y, -fieldLim, fieldLim},
		{"Rad", l.Radius, 1, fieldLim},
		{"In%", pct, 0, 1000},
		{"R", int32(l.Color[0]), 0, 255},
		{"G", int32(l.Color[1]), 0, 255},
		{"B", int32(l.Color[2]), 0, 255},
	})
	nl := l
	nl.X, nl.Y, nl.Radius = v[0], v[1], v[2]
	if v[3] != pct { // only a typed change, so rounding never rewrites the intensity
		nl.Intensity = float32(v[3]) / 100
	}
	nl.Color = [3]uint8{uint8(v[4]), uint8(v[5]), uint8(v[6])}
	le.setLight(le.lightSel, nl)
	rlg.Label(row(), "Drag center/edge; Del, RMB remove")
}
