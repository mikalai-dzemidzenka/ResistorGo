package resistor

import (
	"reflect"
	"testing"
)

// A unit quad with texcoords and normals, exercising fan triangulation.
const quadOBJ = `# comment
mtllib ignored.mtl
o Quad
v 0 0 0
v 1 0 0
v 1 1 0
v 0 1 0
vt 0 0
vt 1 0
vt 1 1
vt 0 1
vn 0 0 1
f 1/1/1 2/2/1 3/3/1 4/4/1
`

func TestParseOBJFanTriangulatesQuads(t *testing.T) {
	verts, texs, norms, err := parseOBJ([]byte(quadOBJ))
	if err != nil {
		t.Fatalf("parseOBJ: %v", err)
	}
	want := []float32{
		0, 0, 0, 1, 0, 0, 1, 1, 0, // 1 2 3
		0, 0, 0, 1, 1, 0, 0, 1, 0, // 1 3 4
	}
	if !reflect.DeepEqual(verts, want) {
		t.Errorf("verts = %v, want %v", verts, want)
	}
	if len(texs) != 12 {
		t.Errorf("len(texs) = %d, want 12", len(texs))
	}
	if len(norms) != len(verts) {
		t.Errorf("len(norms) = %d, want %d", len(norms), len(verts))
	}
	for i := 0; i < len(norms); i += 3 {
		if norms[i] != 0 || norms[i+1] != 0 || norms[i+2] != 1 {
			t.Fatalf("norms[%d:] = %v, want 0 0 1", i, norms[i:i+3])
		}
	}
}

// Blender's export omits vt when the mesh is unwrapped, and indices may count
// back from the end of the list read so far.
func TestParseOBJMissingTexcoordsAndNegativeIndices(t *testing.T) {
	src := `v 0 0 0
v 1 0 0
v 0 1 0
vn 0 0 1
f -3//-1 -2//-1 -1//-1
`
	verts, texs, norms, err := parseOBJ([]byte(src))
	if err != nil {
		t.Fatalf("parseOBJ: %v", err)
	}
	if texs != nil {
		t.Errorf("texs = %v, want nil for a file with no vt", texs)
	}
	want := []float32{0, 0, 0, 1, 0, 0, 0, 1, 0}
	if !reflect.DeepEqual(verts, want) {
		t.Errorf("verts = %v, want %v", verts, want)
	}
	if len(norms) != 9 {
		t.Errorf("len(norms) = %d, want 9", len(norms))
	}
}

func TestParseOBJRejectsEmpty(t *testing.T) {
	if _, _, _, err := parseOBJ([]byte("# nothing here\n")); err == nil {
		t.Error("parseOBJ of a file with no faces: want error, got nil")
	}
}

// Every embedded model must parse into whole triangles, or the menu draws nothing.
func TestEmbeddedModelsParse(t *testing.T) {
	for _, name := range []string{
		"resources/main_menu/head.obj",
		"resources/main_menu/jab1.obj",
		"resources/main_menu/jab2.obj",
		"resources/main_menu/kick.obj",
		"resources/main_menu/martelo.obj",
	} {
		verts, _, norms, err := parseOBJ(asset(name))
		if err != nil {
			t.Errorf("%s: %v", name, err)
			continue
		}
		if len(verts) == 0 || len(verts)%9 != 0 {
			t.Errorf("%s: %d floats, not a whole number of triangles", name, len(verts))
		}
		if norms == nil {
			t.Errorf("%s: no normals", name)
		}
	}
}
