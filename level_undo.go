package resistor

import (
	"maps"
	"slices"
)

// undoDepth is how many edits Ctrl+Z reaches back. The oldest is forgotten
// past it.
const undoDepth = 200

// editState is everything an edit can change: the placed objects, the
// palette's rectangles and the defaults file they are saved to, the lights,
// the lit areas and the ambient. id names it, so the history can tell whether
// undoing has come back to what is on disk.
type editState struct {
	id       uint64
	objects  map[int]LevelLayer
	palette  map[string]ObjectRects // each palette entry's rectangles, by texture
	defaults map[string]ObjectRects
	lights   []LevelLight
	areas    []LitArea
	ambient  Ambient
}

// editHistory is the editor's undo and redo. It does not watch each change:
// every edit counts itself in gen, and at the start of a frame with no gesture
// under way - no drag, no mouse button held, no number being typed - a gen that
// moved since the last look is one finished edit. The state as it was before
// it goes on the undo stack. So a drag, however many frames it runs, is one
// step, and so is a number typed a digit at a time.
type editHistory struct {
	undo, redo []editState
	stable     editState // the state as of the last finished edit
	has        bool      // stable has been taken
	gen        uint64    // edits made; bumped by edited
	stableGen  uint64    // gen when stable was taken
	nextID     uint64
	savedID    uint64 // the state on disk; 0 for none the history has seen
}

// edited is every edit's mark: the level has unsaved changes, and the history
// has a step to take.
func (le *LevelEditor) edited() {
	le.dirty = true
	le.hist.gen++
}

// cloneObjects copies the placed objects deep enough that editing one copy
// never changes the other. Layers are edited in place; an object's rectangle
// lists are replaced whole rather than written into, so they can be shared.
func cloneObjects(objects map[int]LevelLayer) map[int]LevelLayer {
	out := make(map[int]LevelLayer, len(objects))
	for n, layer := range objects {
		out[n] = maps.Clone(layer)
	}
	return out
}

func (le *LevelEditor) snapshot() editState {
	le.hist.nextID++
	palette := make(map[string]ObjectRects, len(le.levelObjects))
	for _, e := range le.levelObjects {
		palette[e.Resource] = e.rects()
	}
	return editState{
		id:       le.hist.nextID,
		objects:  cloneObjects(le.objects),
		palette:  palette,
		defaults: maps.Clone(le.defaults),
		lights:   slices.Clone(le.lights),
		areas:    slices.Clone(le.areas),
		ambient:  le.ambient,
	}
}

// commitEdit takes a step onto the undo stack if an edit has finished since the
// last one. gesture is whether one is still under way; a step is not taken in
// the middle of it.
func (le *LevelEditor) commitEdit(gesture bool) {
	h := &le.hist
	if !h.has {
		h.stable, h.has, h.stableGen = le.snapshot(), true, h.gen
		if !le.dirty {
			h.savedID = h.stable.id
		}
		return
	}
	if gesture || h.gen == h.stableGen {
		return
	}
	h.undo = append(h.undo, h.stable)
	if len(h.undo) > undoDepth {
		h.undo = slices.Delete(h.undo, 0, len(h.undo)-undoDepth)
	}
	h.redo = nil
	h.stable, h.stableGen = le.snapshot(), h.gen
}

// midGesture is whether an edit is still being made: something is being
// dragged, the mouse is held - a slider in the panel - or a number is being
// typed.
func (le *LevelEditor) midGesture(mouseDown bool) bool {
	return mouseDown || le.drag.active || le.lightDrag.mode != lightDragNone || le.areaDrag.active || le.editField >= 0
}

// undo puts the level back as it was before the last edit. It reports whether
// there was one to undo.
func (le *LevelEditor) undo() bool {
	le.commitEdit(false)
	h := &le.hist
	if len(h.undo) == 0 {
		return false
	}
	h.redo = append(h.redo, h.stable)
	h.stable = h.undo[len(h.undo)-1]
	h.undo = h.undo[:len(h.undo)-1]
	le.restore(h.stable)
	return true
}

// redo takes back the last undo. It reports whether there was one.
func (le *LevelEditor) redo() bool {
	le.commitEdit(false)
	h := &le.hist
	if len(h.redo) == 0 {
		return false
	}
	h.undo = append(h.undo, h.stable)
	h.stable = h.redo[len(h.redo)-1]
	h.redo = h.redo[:len(h.redo)-1]
	le.restore(h.stable)
	return true
}

// restore makes the editor hold s. The level is unsaved unless s is what was
// last saved. A selection that no longer points at anything lets go.
func (le *LevelEditor) restore(s editState) {
	le.objects = cloneObjects(s.objects)
	for i, e := range le.levelObjects {
		// an entry imported since has no rectangles to go back to
		if r, ok := s.palette[e.Resource]; ok {
			le.levelObjects[i].setRects(r)
		}
	}
	if le.objIdx >= 0 && le.objIdx < len(le.levelObjects) {
		le.obj.setRects(le.levelObjects[le.objIdx].rects())
	}
	le.defaults = maps.Clone(s.defaults)
	le.lights = slices.Clone(s.lights)
	le.areas = slices.Clone(s.areas)
	le.ambient = s.ambient
	le.hist.stableGen = le.hist.gen
	le.dirty = s.id != le.hist.savedID

	if !le.sel.proto {
		if _, ok := le.objects[le.sel.layer][le.sel.pos]; !ok {
			le.sel = selection{proto: true}
		}
	}
	if le.lightSel >= len(le.lights) {
		le.lightSel = -1
	}
	if le.areaSel >= len(le.areas) {
		le.areaSel = -1
	}
	le.drag, le.lightDrag, le.areaDrag = rectDrag{}, lightDrag{}, rectDrag{}
	le.editField = -1
	le.buildIndex()
	le.clampActiveIdx()
}

// saved marks the state as of now as the one on disk.
func (le *LevelEditor) saved() {
	le.commitEdit(false)
	le.hist.savedID = le.hist.stable.id
}
