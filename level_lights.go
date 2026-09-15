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

// Ambient is the light a level has everywhere its lights do not reach: the
// colour of its dark at full brightness, and how much of that colour there is.
// The two are kept apart so a level is authored as "what colour is the night"
// and "how dark is it" rather than as three dim numbers.
type Ambient struct {
	Color    [3]uint8 `json:"color"`
	Strength float32  `json:"strength"`
}

// defaultAmbient lights a level that carries no ambient of its own: a cool
// night, deep enough for a light to tell against and light enough to read the
// level by outside one.
var defaultAmbient = Ambient{Color: [3]uint8{120, 140, 255}, Strength: 0.4}

// ambientClear is the colour the light map starts from, before any light is
// added into it: the ambient's colour weighted by its strength. Opaque, because
// the composite multiplies alpha along with colour and nothing the scene is
// drawn into may end up transparent.
func ambientClear(a Ambient) rl.Color {
	ch := func(v uint8) uint8 {
		return uint8(clamp(roundf(float32(v)*a.Strength), 0, 255))
	}
	return rl.Color{R: ch(a.Color[0]), G: ch(a.Color[1]), B: ch(a.Color[2]), A: 255}
}

// The level file is an object keyed by layer number, plus one key for the
// lights and one for the ambient. It kept its shape when lights were added to
// it, and again when the ambient was: a level with neither is exactly the file
// it always was.
const (
	lightsKey  = "lights"
	ambientKey = "ambient"
)

// parseLevel reads a level file: every integer key is a layer of tiles, the
// lights key is the lights, the ambient key is the ambient, and anything else
// is skipped rather than failing the whole level. A file with no ambient key
// gets the default one - a file that carries an ambient gets it as written, a
// strength of zero included, which is a level that is black outside its lights.
func parseLevel(b []byte) (map[int]LevelLayer, []LevelLight, Ambient, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(b, &raw); err != nil {
		return nil, nil, defaultAmbient, err
	}
	objects := make(map[int]LevelLayer)
	var lights []LevelLight
	ambient := defaultAmbient
	for k, v := range raw {
		switch k {
		case lightsKey:
			if err := json.Unmarshal(v, &lights); err != nil {
				return nil, nil, defaultAmbient, err
			}
			continue
		case ambientKey:
			var a Ambient
			if err := json.Unmarshal(v, &a); err != nil {
				return nil, nil, defaultAmbient, err
			}
			ambient = a
			continue
		}
		n, err := strconv.Atoi(k)
		if err != nil {
			rl.TraceLog(rl.LogWarning, "level: skipping unknown key %q", k)
			continue
		}
		var layer LevelLayer
		if err := json.Unmarshal(v, &layer); err != nil {
			return nil, nil, defaultAmbient, err
		}
		objects[n] = layer
	}
	return objects, lights, ambient, nil
}

// marshalFile is the level file as saved: the tiles, as marshal gives them to
// the fingerprint, with the lights and the ambient beside them.
func (le *LevelEditor) marshalFile() ([]byte, error) {
	m := make(map[string]any)
	for n, layer := range le.objects {
		m[strconv.Itoa(n)] = layer
	}
	if len(le.lights) > 0 {
		m[lightsKey] = le.lights
	}
	// Always written, even when it matches the default: a level that carries
	// its own ambient goes on looking as it was authored when the default
	// changes under it, and a file without the key is only ever one written
	// before the level had an ambient at all.
	m[ambientKey] = le.ambient
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

// computeLights builds the frame's light map in LightAccum: the level's ambient
// light, with every level light in view added into it, for compositeLights to
// light the scene with. The lights take turns with one set of shadow textures,
// so each is added up the moment it is computed, before the next one overwrites
// them. Additive blending saturates the same whether the lights are added one
// by one or added up first, so this is what drawing each onto the map would
// give.
func computeLights(lights []LevelLight, ambient Ambient, occluders []rl.Rectangle, cam rl.Camera2D) {
	rl.BeginTextureMode(LightAccum)
	// where no light reaches, the map is the ambient alone - and opaque, as
	// compositeLights multiplies alpha along with colour
	rl.ClearBackground(ambientClear(ambient))
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

// lightGlow is how much of the light map is added back over the modulated
// scene. Multiplying alone leaves a light invisible over anything black - the
// sky, an unlit tile - because nothing is there to be lit; this much of it
// added on top keeps a light readable in the air, without washing the dark the
// way the flat black overlay it replaced did.
const lightGlow = 0.2

// compositeLights lights the scene with the light map, in screen space -
// computeLights already put every light where the camera shows it. The scene is
// modulated by the map, so an unlit pixel is what the scene drew there times
// the ambient and a shadow keeps the colour of whatever stands in it, and then
// the map is added back at lightGlow for the glow. Whatever is drawn after this
// - the outlines, the bars, the editor's overlay - is untouched by either.
func compositeLights() {
	w, h := float32(LightAccum.Texture.Width), float32(LightAccum.Texture.Height)
	src := rl.NewRectangle(0, 0, w, -h) // флип render-текстуры
	dst := rl.NewRectangle(0, 0, w, h)
	for _, pass := range []struct {
		blend rl.BlendMode
		tint  rl.Color
	}{
		{rl.BlendMultiplied, rl.White},                   // scene * map
		{rl.BlendAdditive, rl.Fade(rl.White, lightGlow)}, // + map * lightGlow
	} {
		rl.BeginBlendMode(pass.blend)
		rl.DrawTexturePro(LightAccum.Texture, src, dst, rl.Vector2{}, 0, pass.tint)
		rl.EndBlendMode()
	}
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

// setAmbient replaces the level's ambient light, if that changes anything.
func (le *LevelEditor) setAmbient(a Ambient) {
	if le.ambient == a {
		return
	}
	le.ambient = a
	le.dirty = true
}

// deselectLight lets go of the selected light without touching it. It is what a
// right click on empty space does: with no light selected the panel edits the
// level's ambient, and a left click on empty space places a light rather than
// letting go of one, so the selection needs a way out of its own. Letting go is
// not an edit, so it leaves the level as saved as it was.
func (le *LevelEditor) deselectLight() {
	le.lightSel = -1
	le.lightDrag = lightDrag{}
	le.editField = -1
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
// places one, and a right click removes the light under it, or lets go of the
// selected one when there is none.
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
		} else {
			le.deselectLight()
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
		// with no light selected the panel edits the level's ambient: what
		// lights everything the lights do not reach
		rlg.Label(row(), fmt.Sprintf("%d lights - LMB to place one", len(le.lights)))
		rlg.Label(row(), "Ambient - the level's dark")
		a := le.ambient
		pct := int32(roundf(a.Strength * 100))
		v := le.drawIntFields(row, []intField{
			{"R", int32(a.Color[0]), 0, 255},
			{"G", int32(a.Color[1]), 0, 255},
			{"B", int32(a.Color[2]), 0, 255},
			{"Amb%", pct, 0, 1000},
		})
		na := a
		na.Color = [3]uint8{uint8(v[0]), uint8(v[1]), uint8(v[2])}
		if v[3] != pct { // only a typed change, so rounding never rewrites the strength
			na.Strength = float32(v[3]) / 100
		}
		le.setAmbient(na)
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
	rlg.Label(row(), "Drag center/edge; Del, RMB remove; RMB space deselects")
}
