package resistor

import (
	"cmp"
	"encoding/json"
	"fmt"
	"maps"
	"slices"
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

// LitArea is a rectangle of the level the light map holds at full brightness:
// inside it neither the ambient dark nor a shadow a tile casts darkens what
// stands there. Feather is how many pixels of the edge the brightness fades
// over, so an area never reads as a bright straight line across the level; a
// feather of zero is a hard edge. Whole level pixels, as the tiles and the
// lights are.
type LitArea struct {
	X       int32 `json:"x"`
	Y       int32 `json:"y"`
	Width   int32 `json:"w"`
	Height  int32 `json:"h"`
	Feather int32 `json:"feather"`
}

// defaultArea is what an area placed with no area selected looks like.
var defaultArea = LitArea{Width: 64, Height: 64, Feather: 16}

func (a LitArea) rect() rl.Rectangle {
	return rl.Rectangle{X: float32(a.X), Y: float32(a.Y), Width: float32(a.Width), Height: float32(a.Height)}
}

// feather is how wide an area's fade really is: its own feather, but never
// more than half its shorter side, which would leave it dim in the middle
// rather than bright. Worked out here rather than clamped in the data, so an
// area nudged larger gets the feather it was authored with back.
func (a LitArea) feather() int32 {
	return clamp(a.Feather, 0, min(a.Width, a.Height)/2)
}

// gradRect is one piece of a lit area: a rectangle in world pixels and how much
// light it adds at each of its corners, in the order DrawRectangleGradientEx
// takes them. A corner piece is drawn from cornerFade instead, and its one lit
// corner says which way round it goes.
type gradRect struct {
	rect        rl.Rectangle
	topLeft     float32
	bottomLeft  float32
	bottomRight float32
	topRight    float32
	corner      bool
}

// cornerFade is the fade at a lit area's corner: two texels by two, lit at one
// of them and clear at the other three, sampled between the texels' centers so
// that one draw covers a whole corner. It is filtered, unlike everything else
// the game draws, because the fade is the whole point of it; the clear texels
// are white at zero alpha rather than blank, so that filtering fades the alpha
// without dragging the colour towards black with it.
var cornerFade rl.Texture2D

func newCornerFade() rl.Texture2D {
	img := rl.GenImageColor(2, 2, rl.Color{R: 255, G: 255, B: 255})
	rl.ImageDrawPixel(img, 0, 0, rl.White)
	tex := rl.LoadTextureFromImage(img)
	rl.UnloadImage(img)
	rl.SetTextureFilter(tex, rl.FilterBilinear)
	rl.SetTextureWrap(tex, rl.WrapClamp)
	return tex
}

// cornerFadeSource is the part of cornerFade a corner is drawn from: from the
// lit texel's center to the far one's, so the light runs from full at the
// piece's lit corner to nothing at the opposite one. Mirrored on either axis to
// put the lit texel at the corner the piece carries its light in.
func cornerFadeSource(p gradRect) rl.Rectangle {
	src := rl.Rectangle{X: 0.5, Y: 0.5, Width: 1, Height: 1}
	if p.topRight > 0 || p.bottomRight > 0 {
		src.Width = -1
	}
	if p.bottomLeft > 0 || p.bottomRight > 0 {
		src.Height = -1
	}
	return src
}

// areaPieces cuts a lit area into the rectangles it is drawn as: a core at full
// light, four sides fading out across the feather, and four corners fading out
// across both. They tile the area exactly - no overlap, so nothing is added
// twice, and no gap - and a piece with nothing to it, as the core of an area
// feathered to its own half is, is left out.
//
// The core and the sides are gradient rectangles; the corners are drawn from
// cornerFade instead, and carry the light at their inner corner alone so that
// addAreas can tell which way round each one goes. A gradient quad is
// rasterised as two triangles split from its top left to its bottom right, and
// a corner whose lit vertex is off that diagonal loses one whole triangle to
// it - a hard wedge across the corner. A texture is sampled per pixel instead,
// so every corner fades the same way.
func areaPieces(a LitArea) []gradRect {
	if a.Width <= 0 || a.Height <= 0 {
		return nil
	}
	f := float32(a.feather())
	x0, y0 := float32(a.X), float32(a.Y)
	x1, y1 := x0+float32(a.Width), y0+float32(a.Height)
	inW, inH := float32(a.Width)-2*f, float32(a.Height)-2*f
	pieces := []gradRect{
		{rect: rl.Rectangle{X: x0 + f, Y: y0 + f, Width: inW, Height: inH}, topLeft: 1, bottomLeft: 1, bottomRight: 1, topRight: 1},
		{rect: rl.Rectangle{X: x0, Y: y0 + f, Width: f, Height: inH}, bottomRight: 1, topRight: 1},   // left
		{rect: rl.Rectangle{X: x1 - f, Y: y0 + f, Width: f, Height: inH}, topLeft: 1, bottomLeft: 1}, // right
		{rect: rl.Rectangle{X: x0 + f, Y: y0, Width: inW, Height: f}, bottomLeft: 1, bottomRight: 1}, // top
		{rect: rl.Rectangle{X: x0 + f, Y: y1 - f, Width: inW, Height: f}, topLeft: 1, topRight: 1},   // bottom
		{rect: rl.Rectangle{X: x0, Y: y0, Width: f, Height: f}, bottomRight: 1, corner: true},        // top left
		{rect: rl.Rectangle{X: x1 - f, Y: y0, Width: f, Height: f}, bottomLeft: 1, corner: true},     // top right
		{rect: rl.Rectangle{X: x0, Y: y1 - f, Width: f, Height: f}, topRight: 1, corner: true},       // bottom left
		{rect: rl.Rectangle{X: x1 - f, Y: y1 - f, Width: f, Height: f}, topLeft: 1, corner: true},    // bottom right
	}
	kept := pieces[:0]
	for _, p := range pieces {
		if p.rect.Width > 0 && p.rect.Height > 0 {
			kept = append(kept, p)
		}
	}
	return kept
}

// areaVisible reports whether any of a lit area is in view. The fade is inside
// the area's own rectangle, so the rectangle is the whole test.
func areaVisible(a LitArea, view rl.Rectangle) bool {
	return aabb(a.rect(), view)
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

// ambientDarkness is how dark an ambient of strength s leaves a level, as the
// editor's darkness slider shows it: 0% at full strength or brighter, 100% at
// none.
func ambientDarkness(s float32) int32 {
	return clamp(int32(roundf((1-s)*100)), 0, 100)
}

// darknessStrength is the ambient strength that leaves a level d% dark - the
// strength the darkness slider sets.
func darknessStrength(d int32) float32 {
	return float32(100-clamp(d, 0, 100)) / 100
}

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

// The level file is an object: the blocks key is the table of blocks the tiles
// use, each layer number is that layer's tiles as placements into the table,
// and one key each holds the lights, the lit areas and the ambient. Every list
// is one entry to a line, so a save's diff is the entries it changed. A file
// saved before the block table - each layer an object of whole tiles keyed by
// position - still loads, and is written the new way on the next save.
const (
	lightsKey  = "lights"
	areasKey   = "litAreas"
	ambientKey = "ambient"
)

// parseLevel reads a level file: the blocks key is the block table, every
// integer key is a layer of tiles, the lights key is the lights, the areas key
// is the lit areas, the ambient key is the ambient, and anything else is
// skipped rather than failing the whole level. A file with no ambient key gets the default one - a file that carries
// an ambient gets it as written, a strength of zero included, which is a level
// that is black outside its lights.
func parseLevel(b []byte) (map[int]LevelLayer, []LevelLight, []LitArea, Ambient, error) {
	fail := func(err error) (map[int]LevelLayer, []LevelLight, []LitArea, Ambient, error) {
		return nil, nil, nil, defaultAmbient, err
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(b, &raw); err != nil {
		return fail(err)
	}
	var blocks []LevelObject
	if v, ok := raw[blocksKey]; ok {
		if err := json.Unmarshal(v, &blocks); err != nil {
			return fail(err)
		}
	}
	objects := make(map[int]LevelLayer)
	var lights []LevelLight
	var areas []LitArea
	ambient := defaultAmbient
	for k, v := range raw {
		switch k {
		case lightsKey:
			if err := json.Unmarshal(v, &lights); err != nil {
				return fail(err)
			}
			continue
		case areasKey:
			if err := json.Unmarshal(v, &areas); err != nil {
				return fail(err)
			}
			continue
		case ambientKey:
			var a Ambient
			if err := json.Unmarshal(v, &a); err != nil {
				return fail(err)
			}
			ambient = a
			continue
		case blocksKey:
			continue
		}
		n, err := strconv.Atoi(k)
		if err != nil {
			rl.TraceLog(rl.LogWarning, "level: skipping unknown key %q", k)
			continue
		}
		// a list of placements into the blocks, or - in a file saved before
		// there was a block table - the tiles themselves, keyed by position
		var layer LevelLayer
		var ps []placement
		if err := json.Unmarshal(v, &ps); err == nil {
			if layer, err = decodeLayer(blocks, ps); err != nil {
				return fail(err)
			}
		} else if err := json.Unmarshal(v, &layer); err != nil {
			return fail(err)
		}
		objects[n] = layer
	}
	return objects, lights, areas, ambient, nil
}

// marshalFile is the level file as saved: the tiles, as marshal gives them to
// the fingerprint, with the lights, the lit areas and the ambient beside them.
func (le *LevelEditor) marshalFile() ([]byte, error) {
	blocks, layers, err := encodeLayers(le.objects)
	if err != nil {
		return nil, err
	}
	b := []byte("{")
	field := func(key string) func([]byte, error) {
		return func(v []byte, e error) {
			if err = cmp.Or(err, e); err != nil {
				return
			}
			if len(b) > 1 {
				b = append(b, ',')
			}
			b = append(b, '\n')
			b = strconv.AppendQuote(b, key)
			b = append(append(b, ':'), v...)
		}
	}
	field(blocksKey)(jsonLines(blocks))
	for _, n := range slices.Sorted(maps.Keys(layers)) {
		field(strconv.Itoa(n))(jsonLines(layers[n]))
	}
	// Always written, even when it matches the default: a level that carries
	// its own ambient goes on looking as it was authored when the default
	// changes under it, and a file without the key is only ever one written
	// before the level had an ambient at all.
	field(ambientKey)(json.Marshal(le.ambient))
	if len(le.lights) > 0 {
		field(lightsKey)(jsonLines(le.lights))
	}
	if len(le.areas) > 0 {
		field(areasKey)(jsonLines(le.areas))
	}
	if err != nil {
		return nil, err
	}
	return append(b, "\n}\n"...), nil
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
// light, with every lit area and every level light in view added into it, for
// compositeLights to light the scene with. The lights take turns with one set
// of shadow textures,
// so each is added up the moment it is computed, before the next one overwrites
// them. Additive blending saturates the same whether the lights are added one
// by one or added up first, so this is what drawing each onto the map would
// give.
func computeLights(lights []LevelLight, areas []LitArea, ambient Ambient, occluders []rl.Rectangle, cam rl.Camera2D) {
	rl.BeginTextureMode(LightAccum)
	// where no light reaches, the map is the ambient alone - and opaque, as
	// compositeLights multiplies alpha along with colour
	rl.ClearBackground(ambientClear(ambient))
	rl.EndTextureMode()
	view := viewRect(cam, float32(LightAccum.Texture.Width), float32(LightAccum.Texture.Height))
	addAreas(areas, cam, view)
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

// addAreas lifts the light map to full brightness inside every lit area in
// view. White added at the piece's own alpha, in the blend mode the lights use:
// additive blending scales what it adds by alpha, so a piece that ramps from
// full to nothing is a feathered edge, and adding can only brighten - a light
// already brighter than the area inside it stays exactly as bright. Nothing
// here depends on what was added before it, so the areas and the lights may be
// added in any order.
func addAreas(areas []LitArea, cam rl.Camera2D, view rl.Rectangle) {
	if len(areas) == 0 {
		return
	}
	rl.BeginTextureMode(LightAccum)
	rl.BeginMode2D(cam)
	rl.BeginBlendMode(rl.BlendAdditive)
	for _, a := range areas {
		if !areaVisible(a, view) {
			continue
		}
		for _, p := range areaPieces(a) {
			if p.corner {
				rl.DrawTexturePro(cornerFade, cornerFadeSource(p), p.rect, rl.Vector2{}, 0, rl.White)
				continue
			}
			rl.DrawRectangleGradientEx(p.rect,
				rl.Fade(rl.White, p.topLeft),
				rl.Fade(rl.White, p.bottomLeft),
				rl.Fade(rl.White, p.bottomRight),
				rl.Fade(rl.White, p.topRight))
		}
	}
	rl.EndBlendMode()
	rl.EndMode2D()
	rl.EndTextureMode()
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
	le.edited()
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
	le.edited()
}

// setAmbient replaces the level's ambient light, if that changes anything.
func (le *LevelEditor) setAmbient(a Ambient) {
	if le.ambient == a {
		return
	}
	le.ambient = a
	le.edited()
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
	le.edited()
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

// areaAt is the area a click at p takes hold of, and by which handle. Later
// areas are tried before earlier ones, as they are drawn on top of them.
func (le *LevelEditor) areaAt(p rl.Vector2, tol float32) (int, handle) {
	for i := len(le.areas) - 1; i >= 0; i-- {
		if h := rectHandleAt(le.areas[i].rect(), p, tol); h != handleNone {
			return i, h
		}
	}
	return -1, handleNone
}

// placeArea puts an area centred on p and selects it. It takes the size and
// feather of the area selected before it, so a row of matching areas is a row
// of clicks.
func (le *LevelEditor) placeArea(p rl.Vector2) {
	a := defaultArea
	if le.areaSel >= 0 && le.areaSel < len(le.areas) {
		a = le.areas[le.areaSel]
	}
	a.X = int32(roundf(p.X)) - a.Width/2
	a.Y = int32(roundf(p.Y)) - a.Height/2
	le.areas = append(le.areas, a)
	le.areaSel = len(le.areas) - 1
	le.areaDrag = rectDrag{}
	le.editField = -1
	le.edited()
}

// beginAreaDrag takes hold of the area under p, by whichever handle, selecting
// it. The handles are the rect tool's, so an area is moved by its body and
// resized by an edge or a corner exactly as an object's rectangles are.
func (le *LevelEditor) beginAreaDrag(p rl.Vector2, tol float32) bool {
	i, h := le.areaAt(p, tol)
	if i < 0 {
		return false
	}
	le.areaSel = i
	le.editField = -1
	le.areaDrag = rectDrag{active: true, h: h, off: sub(handleAnchor(le.areas[i].rect(), h, p), p)}
	return true
}

// dragArea moves what the drag has hold of to the cursor at p, in whole pixels.
// dragRect keeps the rectangle a pixel wide at the least and never lets it turn
// inside out.
func (le *LevelEditor) dragArea(p rl.Vector2) {
	if le.areaSel < 0 || le.areaSel >= len(le.areas) || !le.areaDrag.active {
		le.areaDrag = rectDrag{}
		return
	}
	to := roundVec2(rl.Vector2{X: p.X + le.areaDrag.off.X, Y: p.Y + le.areaDrag.off.Y})
	a := le.areas[le.areaSel]
	r := dragRect(a.rect(), le.areaDrag.h, to)
	a.X, a.Y = int32(roundf(r.X)), int32(roundf(r.Y))
	a.Width, a.Height = int32(roundf(r.Width)), int32(roundf(r.Height))
	le.setArea(le.areaSel, a)
}

// setArea replaces area i, if that changes anything.
func (le *LevelEditor) setArea(i int, a LitArea) {
	if i < 0 || i >= len(le.areas) || le.areas[i] == a {
		return
	}
	le.areas[i] = a
	le.edited()
}

// deselectArea lets go of the selected area without touching it, as
// deselectLight does for a light: a right click on empty space. Letting go is
// not an edit, so it leaves the level as saved as it was.
func (le *LevelEditor) deselectArea() {
	le.areaSel = -1
	le.areaDrag = rectDrag{}
	le.editField = -1
}

// deleteArea removes area i, keeping the selection on the area it was on.
func (le *LevelEditor) deleteArea(i int) {
	if i < 0 || i >= len(le.areas) {
		return
	}
	le.areas = append(le.areas[:i], le.areas[i+1:]...)
	switch {
	case le.areaSel == i:
		le.areaSel = -1
	case le.areaSel > i:
		le.areaSel--
	}
	le.areaDrag = rectDrag{}
	le.editField = -1
	le.edited()
}

// updateAreaTool is the area tool's handling of the mouse, the cursor being at
// world in the level: a left click on an area takes hold of it, on nothing
// places one, and a right click removes the area under it, or lets go of the
// selected one when there is none.
func (le *LevelEditor) updateAreaTool(world rl.Vector2) {
	if le.areaDrag.active {
		if rl.IsMouseButtonDown(rl.MouseButtonLeft) {
			le.dragArea(world)
		} else {
			le.areaDrag = rectDrag{}
		}
		return
	}
	if le.overGui(MousePos) {
		return
	}
	tol := grabTolerance(Cam.PixelSize())
	switch {
	case rl.IsMouseButtonPressed(rl.MouseButtonRight):
		if i, _ := le.areaAt(world, tol); i >= 0 {
			le.deleteArea(i)
		} else {
			le.deselectArea()
		}
	case rl.IsMouseButtonPressed(rl.MouseButtonLeft):
		if !le.beginAreaDrag(world, tol) {
			le.placeArea(world)
		}
	}
}

// markerBacking is the dark edge under the light markers' center cross and
// radius handle, and under the lit areas' outlines.
var markerBacking = rl.Fade(rl.Black, 0.8)

// DrawOverlay draws what the editor puts over the lit scene: the light markers
// while the light tool is out, and the lit areas while the area tool is. They
// go over the lighting rather than under it with the tiles, because a light is
// brightest at the very center its marker is on and an area is full brightness
// throughout, and either would wash an outline drawn under it out.
func (le *LevelEditor) DrawOverlay() {
	if !DevTools || !Editing {
		return
	}
	switch le.tool {
	case toolLight:
		le.drawLights()
	case toolArea:
		le.drawAreas()
	}
}

// drawAreas marks every lit area while the area tool is out: the rectangle it
// lights and, inside it, the rectangle its feather fades in from, so how soft
// the edge is can be seen as well as read. The selected area is white, the
// others faded. Filled shapes a level pixel wide, not GL lines: a line is one
// screen pixel at any zoom, and the palette pass, sampling the frame in blocks
// of the zoom, would keep only scattered dots of it. Each outline sits on a
// dark backing, which is what makes it tell against the full brightness inside
// the area.
func (le *LevelEditor) drawAreas() {
	for i, a := range le.areas {
		col := rl.Fade(rl.SkyBlue, 0.7)
		if i == le.areaSel {
			col = rl.White
		}
		r := a.rect()
		rl.DrawRectangleLinesEx(rl.Rectangle{X: r.X - 1, Y: r.Y - 1, Width: r.Width + 2, Height: r.Height + 2}, 1, markerBacking)
		rl.DrawRectangleLinesEx(r, 1, col)
		if f := float32(a.feather()); f > 0 {
			in := rl.Rectangle{X: r.X + f, Y: r.Y + f, Width: r.Width - 2*f, Height: r.Height - 2*f}
			if in.Width > 0 && in.Height > 0 {
				rl.DrawRectangleLinesEx(in, 1, rl.Fade(col, 0.5))
			}
		}
	}
}

// drawAreaPanel is the area tool's part of the panel: the selected area's
// numbers, which set it exactly when typed into.
func (le *LevelEditor) drawAreaPanel(row func() rl.Rectangle) {
	if le.areaSel < 0 || le.areaSel >= len(le.areas) {
		rlg.Label(row(), fmt.Sprintf("%d lit areas - LMB to place one", len(le.areas)))
		rlg.Label(row(), "An area is lit whatever shadows it")
		return
	}
	a := le.areas[le.areaSel]
	rlg.Label(row(), fmt.Sprintf("Area %d of %d", le.areaSel+1, len(le.areas)))
	v := le.drawIntFields(row, []intField{
		{"X", a.X, -fieldLim, fieldLim},
		{"Y", a.Y, -fieldLim, fieldLim},
		{"W", a.Width, 1, fieldLim},
		{"H", a.Height, 1, fieldLim},
		{"Fade", a.Feather, 0, fieldLim},
	})
	na := a
	na.X, na.Y, na.Width, na.Height, na.Feather = v[0], v[1], v[2], v[3], v[4]
	le.setArea(le.areaSel, na)
	b := buttons(row(), 2)
	if button(b[0], "Remove (Del)", true) {
		le.deleteArea(le.areaSel)
	}
	if button(b[1], "Deselect", true) {
		le.deselectArea()
	}
	rlg.Label(row(), "Drag the body or an edge; RMB removes")
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
		dark := ambientDarkness(na.Strength)
		d := float32(dark)
		r := row()
		r.X, r.Width = r.X+70, r.Width-110 // room for the labels either side
		rlg.Slider(r, "Darkness", fmt.Sprintf("%d%%", dark), &d, 0, 100)
		if nd := int32(roundf(d)); nd != dark { // only a drag, as with the fields
			na.Strength = darknessStrength(nd)
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
	b := buttons(row(), 2)
	if button(b[0], "Remove (Del)", true) {
		le.deleteLight(le.lightSel)
	}
	if button(b[1], "Deselect", true) {
		le.deselectLight()
	}
	rlg.Label(row(), "Drag the center or the edge; RMB removes")
}
