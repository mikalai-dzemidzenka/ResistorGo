package resistor

import (
	"cmp"
	"encoding/json"
	"fmt"
	rlg "github.com/gen2brain/raylib-go/raygui"
	rl "github.com/gen2brain/raylib-go/raylib"
	"image/color"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

type LevelLayer map[Vector2i]LevelObject

func (ll LevelLayer) MarshalJSON() ([]byte, error) {
	if ll == nil || len(ll) == 0 {
		return []byte("null"), nil
	}
	m := make(map[string]any)
	for pos, obj := range ll {
		m[pos.String()] = obj
	}
	return json.Marshal(m)
}

func (ll *LevelLayer) UnmarshalJSON(b []byte) error {
	m := make(map[string]LevelObject)
	err := json.Unmarshal(b, &m)
	if err != nil {
		return err
	}

	for k, v := range m {
		toks := strings.Split(k, ",")
		x, _ := strconv.Atoi(toks[0])
		y, _ := strconv.Atoi(toks[1])
		if (*ll) == nil {
			(*ll) = make(map[Vector2i]LevelObject)
		}
		(*ll)[Vector2i{X: int32(x), Y: int32(y)}] = v
	}
	return nil
}

// The editor's four tools. The place tool is the editor as it always was; the
// rect tool only selects and shapes rectangles, the light tool only places and
// shapes lights, and the area tool only places and shapes lit areas, so one
// click never acts as two tools at once. The two lighting tools are neighbours,
// as the two lighting tasks are.
const (
	toolPlace = iota
	toolRect
	toolLight
	toolArea
	toolCount
)

// nextTool is the tool Q switches to: place, rect, light, area, and round again.
func nextTool(t int) int { return (t + 1) % toolCount }

// Which of an object's three kinds of rectangle the rect tool is shaping.
const (
	rectCollision = iota
	rectLight
	rectDrop
)

// selection is what rectangle editing acts on: the palette entry new objects
// are born from, or one object already placed in the level. A placed object is
// held by layer and grid position rather than by pointer, because buildIndex
// rebuilds the slice of objects on every edit.
type selection struct {
	proto bool
	layer int
	pos   Vector2i
}

// rectDrag is a rectangle being dragged: what the cursor has hold of, and how
// far the cursor was from it when it took hold, so the rectangle does not jump
// on the first frame of the drag.
type rectDrag struct {
	active bool
	h      handle
	off    rl.Vector2
}

type LevelEditor struct {
	isObjectsViewerEnabled bool
	levelObjects           []LevelObject
	obj                    LevelObject // current object
	objIdx                 int         // which palette entry obj came from, -1 for none
	layer                  int         // current layer
	gridW                  int32
	gridH                  int32
	objects                map[int]LevelLayer // layer - pos - object
	objectsIndexed         []LevelObject
	solids                 []rl.Rectangle // every collision rectangle in world space, in the order collision visits them
	drop                   []bool         // beside solids: whether that rectangle is drop-through
	dirty                  bool           // edited since the last save; only Ctrl+S in edit mode writes

	defaults   map[string]ObjectRects // palette rectangles, from level_objects.json
	tool       int                    // toolPlace or toolRect
	sel        selection              // what rectangle editing acts on
	activeRect int                    // rectCollision or rectLight
	activeIdx  int                    // which rectangle of the activeRect kind
	drag       rectDrag
	editField  int     // which numeric field has the keyboard, -1 for none
	dialog     int     // the file dialog that is up, dlgNone for none
	dlgName    string  // the file name typed into it
	dlgErr     string  // why its last Save or Load did not happen
	dlgPick    int32   // the level picked from its list, -1 for none
	dlgScroll  int32   // how far its list is scrolled
	fieldVals  []int32 // what the panel's numeric fields show

	clip     ObjectRects // rectangles copied with Ctrl+C, pasted with Ctrl+V
	clipFrom string      // what they were copied from, for the panel
	hasClip  bool
	presets  Presets // rectangles kept in slots 1 to 9, from level_presets.json

	lights    []LevelLight // the level's lights, saved with it
	lightSel  int          // the selected light, -1 for none
	lightDrag lightDrag
	ambient   Ambient // what lights the level where its lights do not, saved with it

	hist editHistory // undo and redo; see level_undo.go

	areas    []LitArea // the level's lit areas, saved with it
	areaSel  int       // the selected area, -1 for none
	areaDrag rectDrag  // the area tool holds an area by the rect tool's handles
}
type LevelObject struct {
	Resource       string         `json:"resource,omitempty"`
	Layer          int            `json:"layer,omitempty"`
	X              int32          `json:"x,omitempty"`
	Y              int32          `json:"y,omitempty"`
	CollisionRects []rl.Rectangle // relative to the object's origin; saved by MarshalJSON
	LightRects     []rl.Rectangle // likewise; each casts a shadow
	DropRects      []rl.Rectangle // likewise; floor a fighter can climb down through

	// own is whether the object keeps rectangles of its own rather than
	// following its texture's. It is unexported, so it is never saved: the
	// level file and its fingerprint are exactly the rectangles, and on load
	// markOwnRects works it out from them.
	own bool
}

func NewLevelEditor(gridW, gridH int32) LevelEditor {
	defaults := loadObjectDefaults(objectDefaultsPath)
	le := newLevelEditor(gridW, gridH, buildPalette(defaults), defaults)
	le.presets = loadPresets(presetsPath)
	return le
}

func newLevelEditor(gridW, gridH int32, palette []LevelObject, defaults map[string]ObjectRects) LevelEditor {
	if defaults == nil {
		defaults = make(map[string]ObjectRects)
	}
	le := LevelEditor{
		levelObjects: palette,
		objIdx:       -1,
		gridW:        gridW,
		gridH:        gridH,
		objects:      make(map[int]LevelLayer),
		defaults:     defaults,
		sel:          selection{proto: true},
		editField:    -1,
		lightSel:     -1,
		ambient:      defaultAmbient,
		areaSel:      -1,
	}
	le.selectEntry(0) // an empty palette leaves nothing selected, and nothing to place
	return le
}

// snapCell is the grid cell the place tool puts an object in for a world point.
func snapCell(p Vector2i, gridW, gridH int32) Vector2i {
	c := p
	if p.X < 0 {
		c.X -= gridW
	}
	if p.Y < 0 {
		c.Y -= gridH
	}
	if p.X != 0 {
		c.X = c.X / gridW * gridW
	}
	if p.Y != 0 {
		c.Y = c.Y / gridH * gridH
	}
	return c
}

func (le *LevelEditor) updateCursor() {
	c := snapCell(MouseWorldPos, le.gridW, le.gridH)
	le.obj.X, le.obj.Y = c.X, c.Y
}

func (le *LevelEditor) Update(dt float32) {
	// last frame's edit, finished, is a step Ctrl+Z can take back
	le.commitEdit(le.midGesture(rl.IsMouseButtonDown(rl.MouseButtonLeft) || rl.IsMouseButtonDown(rl.MouseButtonRight)))
	ctrl := rl.IsKeyDown(rl.KeyLeftControl) || rl.IsKeyDown(rl.KeyRightControl)
	shift := rl.IsKeyDown(rl.KeyLeftShift) || rl.IsKeyDown(rl.KeyRightShift)
	if le.dialog != dlgNone {
		if rl.IsKeyPressed(rl.KeyEscape) {
			le.dialog = dlgNone
		}
		if rl.IsKeyPressed(rl.KeyEnter) || rl.IsKeyPressed(rl.KeyKpEnter) {
			le.dialogOK()
		}
		return
	}
	if ctrl && rl.IsKeyPressed(rl.KeyS) {
		le.Save()
		return
	}
	// A numeric field that has the keyboard gets every key: typing a number
	// must not also retool the editor. A Ctrl chord is only ever the chord:
	// Ctrl+C must not also make the collision rectangle the active one.
	if le.editField < 0 && ctrl {
		if rl.IsKeyPressed(rl.KeyC) {
			le.copySelection()
		}
		if rl.IsKeyPressed(rl.KeyV) {
			le.pasteSelection()
		}
		// nothing to undo in the middle of a drag: it is not an edit yet
		if !le.midGesture(rl.IsMouseButtonDown(rl.MouseButtonLeft)) {
			if rl.IsKeyPressed(rl.KeyZ) && !shift {
				le.undo()
			} else if rl.IsKeyPressed(rl.KeyY) || rl.IsKeyPressed(rl.KeyZ) && shift {
				le.redo()
			}
		}
	}
	if le.editField < 0 && !ctrl {
		if rl.IsKeyPressed(rl.KeyTab) {
			LE.OpenObjects()
		}
		if rl.IsKeyPressed(rl.KeyQ) {
			le.setTool(nextTool(le.tool))
		}
		if rl.IsKeyPressed(rl.KeyC) {
			le.pickKind(rectCollision)
		}
		if rl.IsKeyPressed(rl.KeyL) {
			le.pickKind(rectLight)
		}
		if rl.IsKeyPressed(rl.KeyP) {
			le.pickKind(rectDrop)
		}
		if le.tool == toolRect && rl.IsKeyPressed(rl.KeyN) {
			le.addActiveRect()
		}
		if rl.IsKeyPressed(rl.KeyF) {
			le.followTexture()
		}
		if le.tool == toolPlace || le.tool == toolRect {
			for i := range presetSlots {
				if rl.IsKeyPressed(int32(rl.KeyOne) + int32(i)) {
					le.usePreset(i, shift)
				}
			}
		}
	}
	if le.editField < 0 && le.tool == toolRect && rl.IsKeyPressed(rl.KeyDelete) {
		le.clearActiveRect()
	}
	if le.editField < 0 && le.tool == toolLight && rl.IsKeyPressed(rl.KeyDelete) {
		le.deleteLight(le.lightSel)
	}
	if le.editField < 0 && le.tool == toolArea && rl.IsKeyPressed(rl.KeyDelete) {
		le.deleteArea(le.areaSel)
	}
	// The rect tool parks the prototype where it was, so its rectangles can be
	// dragged without the object running off after the cursor.
	if le.tool == toolPlace {
		le.updateCursor()
	}
	if le.isObjectsViewerEnabled {
		le.drag = rectDrag{}
		return
	}

	switch le.tool {
	case toolPlace:
		if le.overGui(MousePos) {
			return
		}
		pos := Vector2i{X: le.obj.X, Y: le.obj.Y}
		if rl.IsMouseButtonPressed(rl.MouseButtonRight) {
			le.erase(pos)
		} else if rl.IsMouseButtonPressed(rl.MouseButtonLeft) {
			le.place(pos)
		}
	case toolRect:
		world := rl.GetScreenToWorld2D(MousePos, Cam.View())
		if le.drag.active {
			if rl.IsMouseButtonDown(rl.MouseButtonLeft) {
				le.dragTo(world)
			} else {
				le.drag = rectDrag{}
			}
			return
		}
		if le.overGui(MousePos) || !rl.IsMouseButtonPressed(rl.MouseButtonLeft) {
			return
		}
		// A handle of the active rectangle wins over the object under it, so a
		// click on a rectangle never also selects something else.
		if !le.beginDrag(world, grabTolerance(Cam.PixelSize())) {
			le.selectAt(snapCell(MouseWorldPos, le.gridW, le.gridH))
		}
	case toolLight:
		le.updateLightTool(rl.GetScreenToWorld2D(MousePos, Cam.View()))
	case toolArea:
		le.updateAreaTool(rl.GetScreenToWorld2D(MousePos, Cam.View()))
	}
}

// grabTolerance is how near a rectangle's edge the cursor has to be to take
// hold of it, in level pixels: about four screen pixels at any zoom.
func grabTolerance(zoom float32) float32 {
	return max(1, 4/zoom)
}

// place puts a copy of the selected palette entry in a grid cell. The copy has
// rectangles of its own, or none where the entry has none.
func (le *LevelEditor) place(pos Vector2i) {
	if le.objIdx < 0 {
		return
	}
	if le.objects[le.layer] == nil {
		le.objects[le.layer] = make(map[Vector2i]LevelObject)
	}
	obj := LevelObject{
		Resource: le.obj.Resource,
		X:        pos.X,
		Y:        pos.Y,
		Layer:    le.layer,
	}
	obj.setRects(le.obj.rects().clone())
	le.objects[le.layer][pos] = obj
	le.edited()
	le.buildIndex()
}

func (le *LevelEditor) erase(pos Vector2i) {
	if le.objects[le.layer] != nil {
		delete(le.objects[le.layer], pos)
	}
	le.edited()
	le.buildIndex()
}

func (le *LevelEditor) setTool(tool int) {
	le.tool = tool
	le.drag = rectDrag{}
	le.lightDrag = lightDrag{}
	le.areaDrag = rectDrag{}
	le.editField = -1
}

// setActiveRect makes that kind of rectangle the one being edited, starting
// from its first when the kind changes.
func (le *LevelEditor) setActiveRect(which int) {
	if which != le.activeRect {
		le.activeIdx = 0
	}
	le.activeRect = which
	le.drag = rectDrag{}
	le.editField = -1
	le.clampActiveIdx()
}

// pickKind is C, L and P: that kind's rectangles, or - when they are already what
// is being edited - the next of them, round again after the last.
func (le *LevelEditor) pickKind(which int) {
	if le.activeRect != which {
		le.setActiveRect(which)
		return
	}
	if obj, ok := le.selectedObject(); ok {
		if n := obj.rects().count(which); n > 0 {
			le.activeIdx = (le.activeIdx + 1) % n
		}
	}
	le.drag = rectDrag{}
	le.editField = -1
}

// clampActiveIdx keeps the active index inside the selection's list of the
// active kind, so a selection with fewer rectangles never points past its end.
func (le *LevelEditor) clampActiveIdx() {
	n := 0
	if obj, ok := le.selectedObject(); ok {
		n = obj.rects().count(le.activeRect)
	}
	le.activeIdx = clamp(le.activeIdx, 0, max(0, n-1))
}

// selectEntry makes palette entry i the object to place and the thing whose
// rectangles are edited. The entry keeps the cell the cursor was on.
func (le *LevelEditor) selectEntry(i int) {
	if i < 0 || i >= len(le.levelObjects) {
		return
	}
	x, y := le.obj.X, le.obj.Y
	le.objIdx = i
	le.obj = le.levelObjects[i]
	le.obj.X, le.obj.Y = x, y
	le.sel = selection{proto: true}
	le.drag = rectDrag{}
	le.editField = -1
	le.clampActiveIdx()
}

func (le *LevelEditor) selectEntryByName(name string) bool {
	for i, obj := range le.levelObjects {
		if obj.Resource == name {
			le.selectEntry(i)
			return true
		}
	}
	return false
}

// selectAt selects the topmost object placed in a grid cell. A cell with
// nothing in it leaves the selection alone rather than falling back to editing
// the palette entry without saying so.
func (le *LevelEditor) selectAt(cell Vector2i) bool {
	layers := make([]int, 0, len(le.objects))
	for l := range le.objects {
		layers = append(layers, l)
	}
	slices.Sort(layers)
	for i := len(layers) - 1; i >= 0; i-- {
		if _, ok := le.objects[layers[i]][cell]; ok {
			le.sel = selection{layer: layers[i], pos: cell}
			le.drag = rectDrag{}
			le.editField = -1
			le.clampActiveIdx()
			return true
		}
	}
	return false
}

// selectedObject is what rectangle editing acts on: the palette entry, at the
// cell it is parked on, or the placed object. A placed object that has since
// been deleted is no selection at all.
func (le *LevelEditor) selectedObject() (LevelObject, bool) {
	if le.sel.proto {
		return le.obj, le.objIdx >= 0
	}
	obj, ok := le.objects[le.sel.layer][le.sel.pos]
	return obj, ok
}

func objOrigin(obj LevelObject) rl.Vector2 {
	return rl.Vector2{X: float32(obj.X), Y: float32(obj.Y)}
}

// setSelectedRects replaces the selected thing's rectangles. For the palette
// entry that is the default new objects are born with, and it reaches every
// placed object of that texture that follows it. For a placed object it is that
// object alone, and it detaches it: its texture's changes pass it by from then on.
func (le *LevelEditor) setSelectedRects(r ObjectRects) {
	if le.sel.proto {
		if le.objIdx < 0 {
			return
		}
		le.obj.setRects(r)
		le.levelObjects[le.objIdx].setRects(r)
		le.defaults[le.obj.Resource] = r.clone()
		le.propagate(le.obj.Resource, r)
	} else {
		layer := le.objects[le.sel.layer]
		obj, ok := layer[le.sel.pos]
		if !ok {
			return
		}
		obj.setRects(r)
		obj.own = true
		layer[le.sel.pos] = obj
	}
	le.buildIndex()
	le.edited()
	le.clampActiveIdx()
}

// propagate gives every placed object of a texture that follows it the
// texture's new rectangles. Each gets copies of its own, so an edit of one
// later cannot reach the others through a shared rectangle.
func (le *LevelEditor) propagate(res string, r ObjectRects) {
	for _, layer := range le.objects {
		for pos, obj := range layer {
			if obj.Resource != res || obj.own {
				continue
			}
			obj.setRects(r.clone())
			layer[pos] = obj
		}
	}
}

// paletteRects is the rectangles a texture's palette entry gives, and whether
// the texture has an entry at all.
func (le *LevelEditor) paletteRects(res string) (ObjectRects, bool) {
	for _, e := range le.levelObjects {
		if e.Resource == res {
			return e.rects(), true
		}
	}
	return ObjectRects{}, false
}

// markOwnRects works out, for a level just loaded, which objects follow their
// texture - the level file does not say. An object follows when its rectangles
// are exactly its texture's, and keeps its own when they differ or its texture
// has no palette entry. Only the flag is written: loading never changes a
// rectangle, so the fingerprint is the one the file gives.
func (le *LevelEditor) markOwnRects() {
	for _, layer := range le.objects {
		for pos, obj := range layer {
			rects, ok := le.paletteRects(obj.Resource)
			obj.own = !ok || !sameRects(obj.rects(), rects)
			layer[pos] = obj
		}
	}
}

// followTexture makes the selected placed object follow its texture again,
// with the texture's rectangles as they are now. It writes the object directly:
// going through setSelectedRects would detach it again at once.
func (le *LevelEditor) followTexture() {
	if le.sel.proto {
		return
	}
	layer := le.objects[le.sel.layer]
	obj, ok := layer[le.sel.pos]
	if !ok || !obj.own {
		return
	}
	rects, ok := le.paletteRects(obj.Resource)
	if !ok {
		return
	}
	changed := !sameRects(obj.rects(), rects)
	obj.setRects(rects.clone())
	obj.own = false
	layer[le.sel.pos] = obj
	le.buildIndex()
	le.drag = rectDrag{}
	le.editField = -1
	le.clampActiveIdx()
	if changed {
		le.edited()
	}
}

// copySelection copies the selected thing's rectangles, a rectangle it does
// not have included as absent.
func (le *LevelEditor) copySelection() {
	obj, ok := le.selectedObject()
	if !ok {
		return
	}
	le.clip = obj.rects().clone()
	le.clipFrom = path.Base(obj.Resource)
	if !le.sel.proto {
		le.clipFrom = "tile " + le.clipFrom
	}
	le.hasClip = true
}

// pasteSelection gives the selected thing a copy of what was copied. It is an
// edit like any other: onto a palette entry it reaches the objects following
// that texture, onto a placed object it detaches that object. A paste that
// changes nothing is no edit at all.
func (le *LevelEditor) pasteSelection() {
	if le.hasClip {
		le.giveRects(le.clip)
	}
}

// giveRects gives the selected thing a copy of r, as a paste does.
func (le *LevelEditor) giveRects(r ObjectRects) {
	obj, ok := le.selectedObject()
	if !ok || sameRects(obj.rects(), r) {
		return
	}
	le.setSelectedRects(r.clone())
	le.drag = rectDrag{}
	le.editField = -1
}

// usePreset is a preset slot's key or button: it gives the selection the
// slot's rectangles. An empty slot, or save, keeps the selection's rectangles
// in the slot instead; Ctrl+S writes them.
func (le *LevelEditor) usePreset(i int, save bool) {
	if p := le.presets[i]; p != nil && !save {
		le.giveRects(*p)
		return
	}
	obj, ok := le.selectedObject()
	if !ok {
		return
	}
	r := obj.rects().clone()
	le.presets[i] = &r
	le.dirty = true
}

// list is the kind's list of rectangles: collision, light or drop-through.
func (r *ObjectRects) list(which int) *[]rl.Rectangle {
	switch which {
	case rectLight:
		return &r.Light
	case rectDrop:
		return &r.Drop
	}
	return &r.Collision
}

// count is how many rectangles of that kind there are.
func (r ObjectRects) count(which int) int { return len(*r.list(which)) }

// get is rectangle i of that kind; nil when there is no such rectangle.
func (r ObjectRects) get(which, i int) *rl.Rectangle {
	rs := *r.list(which)
	if i < 0 || i >= len(rs) {
		return nil
	}
	c := rs[i]
	return &c
}

// set replaces rectangle i of that kind with v, or removes it when v is nil. An
// index one past the end appends. The list is copied rather than written in
// place, because the palette entry and the objects placed from it can share it.
func (r *ObjectRects) set(which, i int, v *rl.Rectangle) {
	l := r.list(which)
	rs := slices.Clone(*l)
	switch {
	case v == nil:
		if i >= 0 && i < len(rs) {
			rs = slices.Delete(rs, i, i+1)
		}
	case i >= 0 && i < len(rs):
		rs[i] = *v
	default:
		rs = append(rs, *v)
	}
	*l = rs
}

// activeRect is the selected thing's rectangle being edited, nil when it has
// none, and the origin that rectangle is an offset from.
func (le *LevelEditor) active() (ObjectRects, *rl.Rectangle, rl.Vector2, bool) {
	obj, ok := le.selectedObject()
	if !ok {
		return ObjectRects{}, nil, rl.Vector2{}, false
	}
	rects := obj.rects()
	return rects, rects.get(le.activeRect, le.activeIdx), objOrigin(obj), true
}

// setActive gives the active rectangle a new value, if that changes anything.
func (le *LevelEditor) setActive(nr *rl.Rectangle) {
	rects, cur, _, ok := le.active()
	if !ok || (cur == nil && nr == nil) || (cur != nil && nr != nil && *cur == *nr) {
		return
	}
	rects.set(le.activeRect, le.activeIdx, nr)
	le.setSelectedRects(rects)
}

// addActiveRect is N: another rectangle of the kind being edited. It starts as
// one grid cell at the origin, somewhere visible to drag it from, and becomes
// the one being edited.
func (le *LevelEditor) addActiveRect() {
	rects, _, _, ok := le.active()
	if !ok {
		return
	}
	le.activeIdx = rects.count(le.activeRect) // one past the end: set appends
	le.setActive(&rl.Rectangle{Width: float32(le.gridW), Height: float32(le.gridH)})
}

// clearActiveRect is Delete: the rectangle being edited goes, and every other
// stays. The one of its kind before it becomes the one being edited.
func (le *LevelEditor) clearActiveRect() {
	i := le.activeIdx
	le.setActive(nil)
	le.activeIdx = max(0, i-1)
	le.clampActiveIdx()
	le.drag = rectDrag{}
	le.editField = -1
}

// rectRef names one of an object's rectangles: rectangle i of kind which.
type rectRef struct{ which, i int }

// rectOrder is the order a click tries the selection's rectangles in: the
// active one first, so it wins where rectangles overlap, then every collision
// rectangle in order, then every light rectangle, then every drop-through one.
func (le *LevelEditor) rectOrder(r ObjectRects) []rectRef {
	refs := []rectRef{{le.activeRect, le.activeIdx}}
	for _, which := range []int{rectCollision, rectLight, rectDrop} {
		for i := range r.count(which) {
			if which != le.activeRect || i != le.activeIdx {
				refs = append(refs, rectRef{which, i})
			}
		}
	}
	return refs
}

// beginDrag takes hold of whichever of the selection's rectangles p is on a
// handle of, trying them in rectOrder, and makes it the one being edited.
func (le *LevelEditor) beginDrag(p rl.Vector2, tol float32) bool {
	obj, ok := le.selectedObject()
	if !ok {
		return false
	}
	rects, origin := obj.rects(), objOrigin(obj)
	for _, ref := range le.rectOrder(rects) {
		r := rects.get(ref.which, ref.i)
		if r == nil {
			continue
		}
		wr := rectToWorld(*r, origin)
		h := rectHandleAt(wr, p, tol)
		if h == handleNone {
			continue
		}
		le.activeRect, le.activeIdx = ref.which, ref.i
		le.editField = -1
		le.drag = rectDrag{active: true, h: h, off: sub(handleAnchor(wr, h, p), p)}
		return true
	}
	return false
}

// dragTo moves what the drag has hold of to the cursor at p, in whole pixels.
func (le *LevelEditor) dragTo(p rl.Vector2) {
	_, cur, origin, ok := le.active()
	if !ok || cur == nil || !le.drag.active {
		le.drag = rectDrag{}
		return
	}
	to := roundVec2(rl.Vector2{X: p.X + le.drag.off.X, Y: p.Y + le.drag.off.Y})
	nr := rectToLocal(dragRect(rectToWorld(*cur, origin), le.drag.h, to), origin)
	le.setActive(&nr)
}

func (le *LevelEditor) buildIndex() {
	le.objectsIndexed = make([]LevelObject, 0)
	for _, layer := range le.objects {
		for _, obj := range layer {
			le.objectsIndexed = append(le.objectsIndexed, obj)
		}
	}

	slices.SortFunc(le.objectsIndexed, func(obj1 LevelObject, obj2 LevelObject) int {
		if obj1.Layer != obj2.Layer {
			return cmp.Compare(obj1.Layer, obj2.Layer)
		}
		if obj1.Y != obj2.Y {
			return cmp.Compare(obj2.Y, obj1.Y)
		}
		return cmp.Compare(obj2.X, obj1.X)
	})

	// Collision walks every rectangle of every object, in the objects' order
	// and each object's own, placed in the level the way it always has been.
	// An object's drop-through rectangles follow its collision ones, marked in
	// drop, so everything else that walks solids takes them for floor.
	solids := make([]rl.Rectangle, 0, len(le.objectsIndexed))
	drop := make([]bool, 0, len(le.objectsIndexed))
	place := func(obj LevelObject, c rl.Rectangle, through bool) {
		solids = append(solids, rl.Rectangle{
			X:      float32(obj.X + int32(c.X)),
			Y:      float32(obj.Y + int32(c.Y)),
			Width:  c.Width,
			Height: c.Height,
		})
		drop = append(drop, through)
	}
	for _, obj := range le.objectsIndexed {
		for _, c := range obj.CollisionRects {
			place(obj, c, false)
		}
		for _, c := range obj.DropRects {
			place(obj, c, true)
		}
	}
	le.solids, le.drop = solids, drop
}

// rebuildPalette re-reads which block textures exist, keeping the rectangles
// edited this session and what is selected.
func (le *LevelEditor) rebuildPalette() {
	sel, name := le.sel, le.obj.Resource
	le.levelObjects = buildPalette(le.defaults)
	le.objIdx = -1
	if !le.selectEntryByName(name) {
		le.selectEntry(0)
	}
	if !sel.proto {
		le.sel = sel
	}
}

// objBounds is where an object's texture covers the level, or one grid cell
// when there is no texture to measure.
func (le *LevelEditor) objBounds(obj LevelObject) rl.Rectangle {
	w, h := float32(le.gridW), float32(le.gridH)
	if tex, ok := resources[obj.Resource]; ok {
		w, h = float32(tex.Width), float32(tex.Height)
	}
	return rl.Rectangle{X: float32(obj.X), Y: float32(obj.Y), Width: w, Height: h}
}

func (le *LevelEditor) Draw() {
	for _, obj := range le.objectsIndexed {
		if tex, ok := resources[obj.Resource]; ok {
			rl.DrawTexture(tex, obj.X, obj.Y, rl.White)
		} else {
			le.drawMissing(obj)
		}
	}
	if (DevTools && Debug) || Editing {
		for _, obj := range le.objectsIndexed {
			for _, c := range obj.CollisionRects {
				rl.DrawRectangleLinesEx(rectToWorld(c, objOrigin(obj)), 1, rl.SkyBlue)
			}
			for _, l := range obj.LightRects {
				rl.DrawRectangleLinesEx(rectToWorld(l, objOrigin(obj)), 1, rl.Yellow)
			}
			for _, d := range obj.DropRects {
				rl.DrawRectangleLinesEx(rectToWorld(d, objOrigin(obj)), 1, dropColor)
			}
		}
		le.DrawCursor()

		le.DrawGrid()
	}
	if Editing {
		// Art that is only on disk would be missing from a build made now.
		for _, obj := range le.objectsIndexed {
			if _, ok := resources[obj.Resource]; ok && !inBuild(obj.Resource) {
				rl.DrawRectangleLinesEx(le.objBounds(obj), 1, rl.Orange)
			}
		}
		if le.tool == toolRect {
			le.drawSelection()
		}
	}
}

// drawMissing stands in for an object whose texture the build does not have,
// which would otherwise be an invisible tile that still collides.
func (le *LevelEditor) drawMissing(obj LevelObject) {
	b := le.objBounds(obj)
	rl.DrawRectangleLinesEx(b, 1, rl.Magenta)
	rl.DrawLine(int32(b.X), int32(b.Y), int32(b.X+b.Width), int32(b.Y+b.Height), rl.Magenta)
	rl.DrawLine(int32(b.X+b.Width), int32(b.Y), int32(b.X), int32(b.Y+b.Height), rl.Magenta)
	rl.DrawText(path.Base(obj.Resource), obj.X+2, obj.Y+2, 10, rl.Magenta)
}

// dropColor outlines drop-through rectangles, apart from the sky-blue collision
// ones and the yellow light ones.
var dropColor = rl.Lime

// drawSelection draws the selected thing and all of its rectangles, the one
// being edited filled and with its corners marked.
func (le *LevelEditor) drawSelection() {
	obj, ok := le.selectedObject()
	if !ok {
		return
	}
	rl.DrawRectangleLinesEx(le.objBounds(obj), 1, rl.White)
	rects := obj.rects()
	refs := le.rectOrder(rects)
	// the active one, first in the order, is drawn last, over the others
	for i := len(refs) - 1; i >= 0; i-- {
		r := rects.get(refs[i].which, refs[i].i)
		if r == nil {
			continue
		}
		wr := rectToWorld(*r, objOrigin(obj))
		col := rl.SkyBlue
		switch refs[i].which {
		case rectLight:
			col = rl.Yellow
		case rectDrop:
			col = dropColor
		}
		if i != 0 {
			rl.DrawRectangleLinesEx(wr, 1, rl.Fade(col, 0.5))
			continue
		}
		rl.DrawRectangleRec(wr, rl.Fade(col, 0.35))
		rl.DrawRectangleLinesEx(wr, 1, col)
		for _, c := range []rl.Vector2{
			{X: wr.X, Y: wr.Y}, {X: wr.X + wr.Width, Y: wr.Y},
			{X: wr.X, Y: wr.Y + wr.Height}, {X: wr.X + wr.Width, Y: wr.Y + wr.Height},
		} {
			rl.DrawRectangleRec(rl.Rectangle{X: c.X - 1, Y: c.Y - 1, Width: 2, Height: 2}, rl.White)
		}
	}
}

func (le *LevelEditor) DrawCursor() {
	if le.objIdx < 0 {
		return
	}
	if tex, ok := resources[le.obj.Resource]; ok {
		rl.DrawTexture(tex, le.obj.X, le.obj.Y, rl.NewColor(255, 255, 255, 127))
	}
	for _, c := range le.obj.CollisionRects {
		rl.DrawRectangleLinesEx(rectToWorld(c, objOrigin(le.obj)), 1, rl.SkyBlue)
	}
	for _, d := range le.obj.DropRects {
		rl.DrawRectangleLinesEx(rectToWorld(d, objOrigin(le.obj)), 1, dropColor)
	}
}

func (le *LevelEditor) DrawGrid() {
	rl.PushMatrix()
	{
		a := float32(le.gridH) / float32(le.gridW)
		rl.Translatef(
			float32(le.gridW),
			float32(le.gridH+(5*le.gridH)),
			0)
		rl.Rotatef(90, 1, 0, 0)
		rl.Scalef(1, 0, a)
		rl.DrawGrid(100, float32(le.gridW))
	}
	rl.PopMatrix()
}

// LoadLevel reads the level in path, reporting whether it did.
func (le *LevelEditor) LoadLevel(path string) bool {
	// The disk comes first, so a level saved from the editor takes effect on the
	// next run; the level the game was built with is the fallback. An edited level
	// fingerprints differently, so the handshake keeps it out of other players' matches.
	b, err := os.ReadFile(path)
	if err != nil {
		b, err = assetFS.ReadFile(path)
	}
	if err != nil {
		rl.TraceLog(rl.LogError, "failed to load level: read file: %s", err.Error())
		return false
	}
	objects, lights, areas, ambient, err := parseLevel(b)
	if err != nil {
		rl.TraceLog(rl.LogError, "failed to load level: json unmarshal: %s", err.Error())
		return false
	}
	le.objects, le.lights, le.lightSel, le.ambient = objects, lights, -1, ambient
	le.areas, le.areaSel = areas, -1

	le.markOwnRects()
	le.buildIndex()
	le.hist = editHistory{} // another level's edits are not this one's to undo
	return true
}

// marshal returns the canonical JSON of the level's tiles (map keys are sorted
// by encoding/json). It is the fingerprint's input, so neither the lights nor
// the ambient is in it; marshalFile is what is saved.
func (le *LevelEditor) marshal() ([]byte, error) {
	m := make(map[int]any)
	for lNum, layer := range le.objects {
		m[lNum] = layer
	}
	return json.Marshal(m)
}

// Hash fingerprints the loaded level so client and server can check they match.
func (le *LevelEditor) Hash() uint64 {
	b, _ := le.marshal()
	return levelHash(b)
}

// SaveLevel writes the level to path, reporting whether it is on disk.
func (le *LevelEditor) SaveLevel(path string) bool {
	b, err := le.marshalFile()
	if err != nil {
		rl.TraceLog(rl.LogError, "failed to save level: marshal level: %s", err.Error())
		return false
	}
	if err := os.WriteFile(path, b, 0o644); err != nil {
		rl.TraceLog(rl.LogError, "failed to save level: %s", err.Error())
		return false
	}
	le.dirty = false
	return true
}

func (le *LevelEditor) OpenObjects() {
	le.isObjectsViewerEnabled = !le.isObjectsViewerEnabled
}

// Save writes the level and the palette defaults, reporting whether all of it
// is on disk. It is Ctrl+S and the save dialog in edit mode, and nothing else
// writes either file.
func (le *LevelEditor) Save() bool {
	ok := saveObjectDefaults(objectDefaultsPath, le.defaults)
	ok = savePresets(presetsPath, le.presets) && ok
	saved := le.SaveLevel(LevelPath)
	if !ok {
		le.dirty = true // the level is saved, but not everything is
	}
	if !le.dirty {
		le.saved() // undoing back to here is back to what is on disk
	}
	return ok && saved
}

// The editor's file dialogs. One is up at a time, and while it is it has the
// keyboard and the mouse.
const (
	dlgNone = iota
	dlgNew
	dlgSave
	dlgLoad
)

// openDialog puts a file dialog up, its name field on the open level - or
// empty, for a new one.
func (le *LevelEditor) openDialog(d int) {
	le.dialog, le.dlgName, le.dlgErr, le.dlgPick, le.dlgScroll = d, LevelPath, "", -1, 0
	if d == dlgNew {
		le.dlgName = ""
	}
}

// dialogOK is the dialog's Save or Load, pressed or Enter.
func (le *LevelEditor) dialogOK() {
	name, err := levelFileName(le.dlgName)
	if err != nil {
		le.dlgErr = err.Error()
		return
	}
	switch le.dialog {
	case dlgNew:
		le.newLevel(name)
	case dlgSave:
		le.saveTo(name)
	case dlgLoad:
		le.loadFrom(name)
	}
}

// saveTo writes the level under name, which from then on is the level Ctrl+S
// writes and the game loads. On any failure the dialog stays up with the
// reason, and LevelPath is left alone.
func (le *LevelEditor) saveTo(name string) {
	prev := LevelPath
	LevelPath = name
	if !le.Save() {
		LevelPath = prev
		le.dlgErr = "could not write " + name
		return
	}
	le.dialog = dlgNone
}

// loadFrom opens the level in name in place of the one being edited, unsaved
// edits and all, and makes it the level Ctrl+S writes and the game loads.
func (le *LevelEditor) loadFrom(name string) {
	if !le.LoadLevel(name) {
		le.dlgErr = "could not load " + name
		return
	}
	LevelPath = name
	le.dirty = false
	le.sel = selection{proto: true} // a tile of the old level is not one of this one
	le.setTool(le.tool)
	le.dialog = dlgNone
}

// newLevel starts an empty level in name: written out empty, then opened like
// any other, so it is on disk from the start and Ctrl+S has somewhere to write.
func (le *LevelEditor) newLevel(name string) {
	if err := os.WriteFile(name, []byte("{}\n"), 0o644); err != nil {
		le.dlgErr = "could not write " + name
		return
	}
	le.loadFrom(name)
}

// levelFileName checks a name typed into a file dialog and gives it the .json
// it may lack. Levels live beside the game, so a path is refused, and so are
// the files beside it that are not levels.
func levelFileName(s string) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", fmt.Errorf("type a file name")
	}
	if !strings.EqualFold(filepath.Ext(s), ".json") {
		s += ".json"
	}
	if filepath.Base(s) != s || strings.ContainsAny(s, `/\:`) {
		return "", fmt.Errorf("a file name only, not a path")
	}
	for _, other := range []string{objectDefaultsPath, presetsPath, settingsPath} {
		if strings.EqualFold(s, other) {
			return "", fmt.Errorf("%s is not a level", other)
		}
	}
	return s, nil
}

// levelFiles lists the levels there are to load: those beside the game, and
// the one it was built with.
func levelFiles() []string {
	files := []string{"level_1.json"}
	entries, _ := os.ReadDir(".")
	for _, e := range entries {
		if name, err := levelFileName(e.Name()); err == nil && name == e.Name() && !e.IsDir() {
			files = append(files, name)
		}
	}
	slices.Sort(files)
	return slices.Compact(files)
}

// drawDialog draws the file dialog over the middle of the screen: the file
// name, the levels to pick from when loading, and what the button will throw
// away.
func (le *LevelEditor) drawDialog() {
	title, ok, h := "Save level", "Save (Enter)", float32(176)
	switch le.dialog {
	case dlgNew:
		title, ok = "New level", "Create (Enter)"
	case dlgLoad:
		title, ok, h = "Load level", "Load (Enter)", 380
	}
	const w = 460
	box := rl.Rectangle{X: (VW - w) / 2, Y: (VH - h) / 2, Width: w, Height: h}
	if rlg.WindowBox(box, title) {
		le.dialog = dlgNone
		return
	}
	x, y := box.X+12, box.Y+34
	rlg.Label(rl.Rectangle{X: x, Y: y, Width: w - 24, Height: 24}, "File name (beside the game)")
	rlg.TextBox(rl.Rectangle{X: x, Y: y + 28, Width: w - 24, Height: 30}, &le.dlgName, 64, true)
	y += 66
	if le.dialog == dlgLoad {
		files := levelFiles()
		if rlg.ListView(rl.Rectangle{X: x, Y: y, Width: w - 24, Height: 200}, strings.Join(files, ";"), &le.dlgScroll, &le.dlgPick); le.dlgPick >= 0 && int(le.dlgPick) < len(files) {
			le.dlgName = files[le.dlgPick]
		}
		y += 204
	}
	note := le.dlgErr
	if name, err := levelFileName(le.dlgName); note == "" && err == nil {
		_, statErr := os.Stat(name)
		switch {
		case statErr == nil && (le.dialog == dlgNew || le.dialog == dlgSave && name != LevelPath):
			note = "replaces the existing " + name
		case le.dialog != dlgSave && le.dirty:
			note = "discards the unsaved edits"
		}
	}
	rl.DrawText(note, int32(x), int32(y), 20, rl.Orange)
	b := buttons(rl.Rectangle{X: x, Y: box.Y + h - 40, Width: w - 24, Height: 28}, 2)
	if rlg.Button(b[0], ok) {
		le.dialogOK()
	}
	if rlg.Button(b[1], "Cancel (Esc)") {
		le.dialog = dlgNone
	}
}

// importDropped imports every image dropped on the window, and selects the
// last one.
func (le *LevelEditor) importDropped(files []string) {
	for _, f := range files {
		le.importTexture(f)
	}
}

// droppedTextureName is the name a dropped file is imported under: the file's
// own name in the block textures folder. Only PNGs are taken, and the extension
// is written in lower case, since the build embeds *.png and nothing else. A
// name go:embed would skip - a leading dot or underscore - is refused for the
// same reason: it would place fine and then never ship.
func droppedTextureName(p string) (string, bool) {
	base := filepath.Base(p)
	ext := filepath.Ext(base)
	if !strings.EqualFold(ext, ".png") {
		return "", false
	}
	stem := strings.TrimSuffix(base, ext)
	if stem == "" || strings.HasPrefix(base, ".") || strings.HasPrefix(base, "_") {
		return "", false
	}
	return blocksDir + "/" + stem + ".png", true
}

// importTexture copies a dropped image in beside the other block textures,
// loads it and selects it. The file is proved to decode before anything is
// copied, so a bad drop cannot leave a file that breaks the next run's palette.
// A name that is already taken is never overwritten: that entry is selected.
func (le *LevelEditor) importTexture(src string) bool {
	name, ok := droppedTextureName(src)
	if !ok {
		rl.TraceLog(rl.LogWarning, "import: not a texture that can be placed: %s", src)
		return false
	}
	if _, err := os.Stat(name); err == nil || inBuild(name) {
		rl.TraceLog(rl.LogWarning, "import: %s already exists; selecting it", name)
		if le.selectEntryByName(name) {
			return true
		}
		loadBlockTex(name)
		le.rebuildPalette()
		return le.selectEntryByName(name)
	}
	data, err := os.ReadFile(src)
	if err != nil {
		rl.TraceLog(rl.LogError, "import: read %s: %s", src, err.Error())
		return false
	}
	if !imageDecodable(".png", data) {
		rl.TraceLog(rl.LogError, "import: %s is not an image that can be read", src)
		return false
	}
	if err := os.MkdirAll(blocksDir, 0o755); err != nil {
		rl.TraceLog(rl.LogError, "import: %s", err.Error())
		return false
	}
	f, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		rl.TraceLog(rl.LogError, "import: create %s: %s", name, err.Error())
		return false
	}
	_, err = f.Write(data)
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		os.Remove(name)
		rl.TraceLog(rl.LogError, "import: write %s: %s", name, err.Error())
		return false
	}
	if !loadTexBytes(name, data) {
		return false
	}
	le.rebuildPalette()
	return le.selectEntryByName(name)
}

// The editor's GUI, in virtual-screen pixels. Each rectangle is used both to
// draw and to keep world clicks out, so the two cannot drift apart.
const (
	pickerX, pickerY = 25, 35
	pickerCell       = 100
	pickerCols       = 12
	panelW, panelH   = 380, 428
)

func pickerCellRect(i int) rl.Rectangle {
	return rl.Rectangle{
		X:      float32(pickerX + (i%pickerCols)*pickerCell),
		Y:      float32(pickerY + (i/pickerCols)*pickerCell),
		Width:  pickerCell,
		Height: pickerCell,
	}
}

func editorPanel() rl.Rectangle {
	return rl.Rectangle{X: 20, Y: VH - panelH - 20, Width: panelW, Height: panelH}
}

// overGui reports whether the cursor is on the editor's own GUI, where a click
// belongs to a widget and never to the level.
func (le *LevelEditor) overGui(p rl.Vector2) bool {
	if rl.CheckCollisionPointRec(p, editorPanel()) {
		return true
	}
	if le.isObjectsViewerEnabled {
		for i := range le.levelObjects {
			if rl.CheckCollisionPointRec(p, pickerCellRect(i)) {
				return true
			}
		}
	}
	return false
}

func (le *LevelEditor) DrawGui() {
	// the dialog has the mouse: nothing behind it can be pressed
	if le.dialog != dlgNone {
		rlg.Lock()
	}
	if le.isObjectsViewerEnabled {
		rlg.SetStyle(rlg.BUTTON, rlg.BASE_COLOR_NORMAL, rlg.PropertyValue(rl.ColorToInt(color.RGBA{0, 0, 0, 0})))
		rlg.SetStyle(rlg.BUTTON, rlg.BASE_COLOR_FOCUSED, rlg.PropertyValue(rl.ColorToInt(color.RGBA{0, 0, 0, 25})))
		for i, obj := range le.levelObjects {
			cell := pickerCellRect(i)
			if tex, ok := resources[obj.Resource]; ok {
				rl.DrawTexturePro(tex,
					rl.Rectangle{Width: float32(tex.Width), Height: float32(tex.Height)},
					cell,
					rl.Vector2{},
					0,
					rl.White,
				)
			}
			if !inBuild(obj.Resource) {
				rl.DrawRectangleLinesEx(cell, 3, rl.Orange)
				rl.DrawText("not in build", int32(cell.X)+4, int32(cell.Y+cell.Height)-16, 10, rl.Orange)
			}
			if le.sel.proto && i == le.objIdx {
				rl.DrawRectangleLinesEx(cell, 2, rl.White)
			}
			if rlg.Button(cell, "") {
				le.selectEntry(i)
			}
		}
	}
	le.drawPanel()
	if le.dialog != dlgNone {
		rlg.Unlock()
		le.drawDialog()
	}
}

// drawPanel shows the tool, what is selected and - with the rect tool - the
// rectangle being edited, with a field for each of its numbers.
func (le *LevelEditor) drawPanel() {
	p := editorPanel()
	rl.DrawRectangleRec(p, rl.Fade(rl.Black, 0.7))
	x, y := p.X+10, p.Y+10
	row := func() rl.Rectangle {
		r := rl.Rectangle{X: x, Y: y, Width: p.Width - 20, Height: 24}
		y += 28
		return r
	}

	tool := int32(le.tool)
	r := row()
	r.Width = (r.Width - 6) / toolCount // raygui spaces the toggles 2px apart
	rlg.ToggleGroup(r, "Place;Rect;Light;Area", &tool)
	if int(tool) != le.tool {
		le.setTool(int(tool))
	}
	b := buttons(row(), 5)
	if button(b[0], "Undo", len(le.hist.undo) > 0 || le.hist.gen != le.hist.stableGen) {
		le.undo()
	}
	if button(b[1], "Redo", len(le.hist.redo) > 0) {
		le.redo()
	}
	if button(b[2], "New", true) {
		le.openDialog(dlgNew)
	}
	if button(b[3], "Save", true) {
		le.openDialog(dlgSave)
	}
	if button(b[4], "Load", true) {
		le.openDialog(dlgLoad)
	}

	switch le.tool {
	case toolPlace:
		rlg.Label(row(), "LMB place, RMB delete (Q: next tool)")
		if button(row(), "Pick texture (Tab)", true) {
			le.OpenObjects()
		}
	case toolLight:
		le.drawLightPanel(row)
		return
	case toolArea:
		le.drawAreaPanel(row)
		return
	}

	obj, ok := le.selectedObject()
	switch {
	case !ok:
		rlg.Label(row(), "Nothing selected")
	case le.sel.proto:
		rlg.Label(row(), "Palette: "+path.Base(obj.Resource))
	default:
		rlg.Label(row(), fmt.Sprintf("Tile %d,%d: %s", obj.X, obj.Y, path.Base(obj.Resource)))
		if obj.own {
			b := buttons(row(), 2)
			rlg.Label(b[0], "Own rectangles")
			if button(b[1], "Follow texture (F)", true) {
				le.followTexture()
			}
		} else {
			rlg.Label(row(), "Follows texture")
		}
	}
	if ok && !inBuild(obj.Resource) {
		r := row()
		rl.DrawText("not in build - rebuild to ship it", int32(r.X), int32(r.Y), 20, rl.Orange)
	}
	if le.tool != toolRect || !ok {
		return
	}
	b = buttons(row(), 2)
	if button(b[0], "Copy rects (Ctrl+C)", true) {
		le.copySelection()
	}
	if button(b[1], "Paste (Ctrl+V)", le.hasClip) {
		le.pasteSelection()
	}
	if le.hasClip {
		rlg.Label(row(), "Copied: "+le.clipFrom)
	}
	// preset slots: a number gives its rectangles, + (or Shift) keeps the selection's
	shift := rl.IsKeyDown(rl.KeyLeftShift) || rl.IsKeyDown(rl.KeyRightShift)
	b = buttons(row(), presetSlots)
	for i, p := range le.presets {
		label := "+"
		if p != nil {
			label = strconv.Itoa(i + 1)
		}
		if button(b[i], label, true) {
			le.usePreset(i, shift)
		}
	}

	active := int32(le.activeRect)
	r = row()
	r.Width = (r.Width - 4) / 3 // each of the three toggles, side by side inside the panel
	rlg.ToggleGroup(r, "Collision (C);Light (L);Drop (P)", &active)
	if int(active) != le.activeRect {
		le.setActiveRect(int(active))
	}

	rects, cur, _, _ := le.active()
	n := rects.count(le.activeRect)
	if n > 0 {
		name := [...]string{rectCollision: "Collision", rectLight: "Light", rectDrop: "Drop"}[le.activeRect]
		rlg.Label(row(), fmt.Sprintf("%s %d of %d", name, le.activeIdx+1, n))
	} else {
		rlg.Label(row(), "None of this kind")
	}
	// Each of these changes which rectangle is active. The fields below would
	// write the one they were drawn from onto it, so the panel stops for this
	// frame and draws the new one on the next.
	b = buttons(row(), 3)
	switch {
	case button(b[0], "Add (N)", true):
		le.addActiveRect()
		return
	case button(b[1], "Remove (Del)", cur != nil):
		le.clearActiveRect()
		return
	case button(b[2], "Next", n > 1):
		le.pickKind(le.activeRect)
		return
	}
	if cur == nil {
		return
	}
	v := le.drawIntFields(row, []intField{
		{"X", int32(cur.X), -fieldLim, fieldLim},
		{"Y", int32(cur.Y), -fieldLim, fieldLim},
		{"W", int32(cur.Width), 1, fieldLim},
		{"H", int32(cur.Height), 1, fieldLim},
	})
	le.setActive(&rl.Rectangle{X: float32(v[0]), Y: float32(v[1]), Width: float32(v[2]), Height: float32(v[3])})
	rlg.Label(row(), "Drag to shape. 1-9 preset, Shift+1-9 keep one")
}

// buttons splits a panel row into n buttons side by side.
func buttons(r rl.Rectangle, n int) []rl.Rectangle {
	const gap = 4
	w := (r.Width - gap*float32(n-1)) / float32(n)
	out := make([]rl.Rectangle, n)
	for i := range out {
		out[i] = rl.Rectangle{X: r.X + float32(i)*(w+gap), Y: r.Y, Width: w, Height: r.Height}
	}
	return out
}

// button is a panel button that reports a press. One with nothing to do is
// drawn greyed out and cannot be pressed.
func button(r rl.Rectangle, label string, enabled bool) bool {
	if !enabled {
		rlg.Disable()
		defer rlg.Enable()
	}
	return rlg.Button(r, label) && enabled
}

// intField is one numeric field in the panel: its label, the value it shows
// and the range it accepts.
type intField struct {
	label  string
	val    int32
	lo, hi int32
}

// fieldLim bounds positions and sizes typed into the panel.
const fieldLim = 100000

// drawIntFields shows fields two to a row, taking rows from row, and returns
// what they hold, each clamped to its range: what is being typed in the field
// that has the keyboard, and every other field's own value.
func (le *LevelEditor) drawIntFields(row func() rl.Rectangle, fields []intField) []int32 {
	if len(le.fieldVals) != len(fields) {
		le.fieldVals = make([]int32, len(fields))
		le.editField = -1
	}
	for i, f := range fields {
		if i != le.editField {
			le.fieldVals[i] = f.val // the field being typed in keeps what is typed
		}
	}
	toggled := make([]bool, len(fields))
	var r rl.Rectangle
	for i, f := range fields {
		if i%2 == 0 {
			r = row()
		}
		b := rl.Rectangle{X: r.X + 30 + float32(i%2)*170, Y: r.Y, Width: 120, Height: r.Height}
		toggled[i] = rlg.ValueBox(b, f.label, &le.fieldVals[i], int(f.lo), int(f.hi), le.editField == i)
	}
	le.editField = nextEditField(le.editField, toggled)
	out := make([]int32, len(fields))
	for i, f := range fields {
		out[i] = clamp(le.fieldVals[i], f.lo, f.hi)
	}
	return out
}

// nextEditField is which numeric field has the keyboard after a frame in which
// the fields in toggled were clicked. Clicking the field being typed in, or
// away from it, ends typing; clicking another field moves typing there -
// whichever order raygui reports the two in.
func nextEditField(editing int, toggled []bool) int {
	next := editing
	if editing >= 0 && editing < len(toggled) && toggled[editing] {
		next = -1
	}
	for i, t := range toggled {
		if t && i != editing {
			next = i
		}
	}
	return next
}

func (le *LevelEditor) LightOccluders() []rl.Rectangle {
	var result []rl.Rectangle
	for _, obj := range le.objectsIndexed {
		for _, l := range obj.LightRects {
			result = append(result, rl.Rectangle{
				X:      float32(obj.X) + l.X,
				Y:      float32(obj.Y) + l.Y,
				Width:  l.Width,
				Height: l.Height,
			})
		}
	}
	return result
}

// Editing is edit mode: the match is paused and the level editor takes the
// mouse. Outside it the editor reads no input at all, so an attack is never
// also an edit.
var Editing bool

// toggleEditing handles a press of the edit-mode key. Edit mode is only for a
// single-player match, and never while time manipulation
// holds the match. Leaving it drops any click latched while editing, so it
// does not fire as an attack on the first tick back.
func toggleEditing(offline, timeManipulating bool) {
	if Editing {
		Editing = false
		LE.dialog = dlgNone
		clearLatchedInput()
		return
	}
	if offline && !timeManipulating {
		Editing = true
	}
}

// drawEditMarker says that edit mode is on, and whether the level has edits
// that Ctrl+S has not saved yet. Drawn in screen space, after the palette pass.
func drawEditMarker() {
	text := "EDIT"
	if LE.dirty {
		text = "EDIT (unsaved - Ctrl+S)"
	}
	if n, m := len(LE.hist.undo), len(LE.hist.redo); n+m > 0 {
		text += fmt.Sprintf("   undo %d (Ctrl+Z), redo %d (Ctrl+Y)", n, m)
	}
	rl.DrawText(text, 20, 70, 20, rl.Yellow)
}
