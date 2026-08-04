package resistor

import (
	"cmp"
	"encoding/json"
	rl "github.com/gen2brain/raylib-go/raylib"
	"os"
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

type LevelEditor struct {
	obj            LevelObject // current object
	layer          int         // current layer
	gridW          int32
	gridH          int32
	objects        map[int]LevelLayer // layer - pos - object
	objectsIndexed []LevelObject
}
type LevelObject struct {
	Resource      string        `json:"resource,omitempty"`
	Layer         int           `json:"layer,omitempty"`
	X             int32         `json:"x,omitempty"`
	Y             int32         `json:"y,omitempty"`
	CollisionRect *rl.Rectangle `json:"collisionRect,omitempty"`
}

func NewLevelEditor(gridW, gridH int32) LevelEditor {
	return LevelEditor{
		obj: LevelObject{
			Resource: "resources/blocks/e1m1_wall1.png",
			Layer:    0,
			X:        0,
			Y:        0,
			CollisionRect: &rl.Rectangle{
				X:      0,
				Y:      86,
				Width:  40,
				Height: 20,
			},
		},
		gridW:   gridW,
		gridH:   gridH,
		objects: make(map[int]LevelLayer),
	}
}

func (le *LevelEditor) updateCursor() {
	le.obj.X = MouseScreenPos.X
	le.obj.Y = MouseScreenPos.Y
	if MouseScreenPos.X < 0 {
		le.obj.X -= le.gridW
	}
	if MouseScreenPos.Y < 0 {
		le.obj.Y -= le.gridH
	}
	if MouseScreenPos.X != 0 {
		le.obj.X /= le.gridW
		le.obj.X *= le.gridW
	}

	if MouseScreenPos.Y != 0 {
		le.obj.Y /= le.gridH
		le.obj.Y *= le.gridH
	}
}

func (le *LevelEditor) Update(dt float32) {
	le.updateCursor()

	pos := Vector2i{
		X: le.obj.X,
		Y: le.obj.Y,
	}
	if rl.IsMouseButtonPressed(rl.MouseButtonRight) {
		if le.objects[le.layer] != nil {
			delete(le.objects[le.layer], pos)
		}

		le.buildIndex()
	} else if rl.IsMouseButtonPressed(rl.MouseButtonLeft) {
		if le.objects[le.layer] == nil {
			le.objects[le.layer] = make(map[Vector2i]LevelObject)
		}

		le.objects[le.layer][pos] = LevelObject{
			Resource: le.obj.Resource,
			X:        pos.X,
			Y:        pos.Y,
			Layer:    le.layer,
			CollisionRect: &rl.Rectangle{
				X:      le.obj.CollisionRect.X,
				Y:      le.obj.CollisionRect.Y,
				Width:  le.obj.CollisionRect.Width,
				Height: le.obj.CollisionRect.Height,
			},
		}
		le.buildIndex()
	}
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
}

func (le *LevelEditor) Draw() {
	for _, obj := range le.objectsIndexed {
		rl.DrawTexture(resources[obj.Resource], obj.X, obj.Y, rl.White)
	}
	for _, obj := range le.objectsIndexed {
		if obj.CollisionRect != nil {
			rl.DrawRectangleLinesEx(rl.Rectangle{
				X:      float32(obj.X) + obj.CollisionRect.X,
				Y:      float32(obj.Y) + obj.CollisionRect.Y,
				Width:  obj.CollisionRect.Width,
				Height: obj.CollisionRect.Height,
			}, 1, rl.SkyBlue)
		}
	}

	le.DrawCursor()

	le.DrawGrid()
}

func (le *LevelEditor) DrawCursor() {
	rl.DrawTexture(resources[le.obj.Resource], le.obj.X, le.obj.Y, rl.NewColor(255, 255, 255, 127))
	if le.obj.CollisionRect != nil {
		rl.DrawRectangleLinesEx(rl.Rectangle{
			X:      float32(le.obj.X) + le.obj.CollisionRect.X,
			Y:      float32(le.obj.Y) + le.obj.CollisionRect.Y,
			Width:  le.obj.CollisionRect.Width,
			Height: le.obj.CollisionRect.Height,
		}, 1, rl.SkyBlue)
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

func (le *LevelEditor) LoadLevel(path string) {
	b, err := os.ReadFile(path)
	if err != nil {
		rl.TraceLog(rl.LogError, "failed to load level: read file: %s", err.Error())
		return
	}
	err = json.Unmarshal(b, &le.objects)
	if err != nil {
		rl.TraceLog(rl.LogError, "failed to load level: json unmarshal: %s", err.Error())
		return
	}

	le.buildIndex()
}

func (le *LevelEditor) SaveLevel(path string) {
	m := make(map[int]any)
	for lNum, layer := range le.objects {
		m[lNum] = layer
	}
	b, err := json.Marshal(m)
	if err != nil {
		rl.TraceLog(rl.LogError, "failed to save level: marshal level: %s", err.Error())
		return
	}
	f, err := os.Create(path)
	if err != nil {
		rl.TraceLog(rl.LogError, "failed to save level: create file: %s", err.Error())
		return
	}
	defer f.Close()

	_, err = f.Write(b)
	if err != nil {
		rl.TraceLog(rl.LogError, "failed to save level: save file: %s", err.Error())
		return
	}
}
