package resistor

import "testing"

// A level saved before layer 0 was the fighters' loads its tiles behind them,
// and every load of it fingerprints the same, so client and server agree.
func TestLayerZeroLoadsBehindTheFighters(t *testing.T) {
	old := []byte(`{"blocks":[{"resource":"a"}],"0":[[0,39,0]]}`)
	objects, _, _, _, err := parseLevel(old)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := objects[0]; ok {
		t.Errorf("layer 0 still holds tiles after the load")
	}
	if obj, ok := objects[-1][Vector2i{X: 39}]; !ok || obj.Layer != -1 {
		t.Errorf("the tile loaded as %+v (found %t), want it on layer -1", obj, ok)
	}
	again, _, _, _, _ := parseLevel(old)
	if (&LevelEditor{objects: objects}).Hash() != (&LevelEditor{objects: again}).Hash() {
		t.Errorf("two loads of the same file fingerprint differently")
	}
}

func TestStepLayerSkipsTheFighters(t *testing.T) {
	if got := stepLayer(-1, 1); got != 1 {
		t.Errorf("up from -1 = %d, want 1", got)
	}
	if got := stepLayer(1, -1); got != -1 {
		t.Errorf("down from 1 = %d, want -1", got)
	}
	l := -1
	for _, dir := range []int{1, 1, -1, -1, -1, 1, 1, 1} {
		if l = stepLayer(l, dir); l == 0 {
			t.Fatalf("a step landed on layer 0")
		}
	}
}

// A fresh editor places on layer -1; tiles on either side of the fighters share
// a cell, and erasing takes only the current layer's.
func TestLayersShareACell(t *testing.T) {
	le := newLevelEditor(39, 83, []LevelObject{{Resource: "x"}}, nil)
	cell := Vector2i{X: 39}
	le.place(cell)
	if obj, ok := le.objects[-1][cell]; !ok || obj.Layer != -1 {
		t.Fatalf("a fresh editor placed %+v (found %t), want it on layer -1", obj, ok)
	}
	le.layer = stepLayer(le.layer, 1)
	le.place(cell)
	le.erase(cell)
	if _, ok := le.objects[-1][cell]; !ok {
		t.Errorf("erasing on layer 1 took the layer -1 tile")
	}
	if _, ok := le.objects[1][cell]; ok {
		t.Errorf("erasing on layer 1 left its tile")
	}
}

func TestSplitAtFighters(t *testing.T) {
	le := newLevelEditor(39, 83, nil, nil)
	for _, l := range []int{2, -1, 1, 0, -3} {
		le.objects[l] = LevelLayer{{}: {Resource: "x", Layer: l}}
	}
	le.buildIndex()
	back, front := splitAtFighters(le.objectsIndexed)
	layers := func(objs []LevelObject) (ls []int) {
		for _, o := range objs {
			ls = append(ls, o.Layer)
		}
		return ls
	}
	if got := layers(back); len(got) != 3 || got[0] != -3 || got[1] != -1 || got[2] != 0 {
		t.Errorf("drawn behind the fighters: layers %v, want [-3 -1 0]", got)
	}
	if got := layers(front); len(got) != 2 || got[0] != 1 || got[1] != 2 {
		t.Errorf("drawn in front of the fighters: layers %v, want [1 2]", got)
	}
}
