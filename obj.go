package resistor

import (
	"bufio"
	"bytes"
	"errors"
	"strconv"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// liveMeshData keeps the vertex arrays of every mesh built here reachable.
// raylib's Mesh holds raw pointers into them after UploadMesh, so the GC must
// not collect them while the model can still be unloaded or re-uploaded.
var liveMeshData []any

// loadOBJ uploads a model built from Wavefront OBJ bytes, which raylib can only
// do from a file.
func loadOBJ(name string, data []byte) rl.Model {
	verts, texs, norms, err := parseOBJ(data)
	if err != nil {
		rl.TraceLog(rl.LogError, "obj: %s: %s", name, err.Error())
		return rl.Model{}
	}

	mesh := rl.Mesh{
		VertexCount:   int32(len(verts) / 3),
		TriangleCount: int32(len(verts) / 9),
		Vertices:      &verts[0],
	}
	if texs != nil {
		mesh.Texcoords = &texs[0]
	}
	if norms != nil {
		mesh.Normals = &norms[0]
	}
	rl.UploadMesh(&mesh, false)
	liveMeshData = append(liveMeshData, verts, texs, norms)
	return rl.LoadModelFromMesh(mesh)
}

// parseOBJ turns Wavefront OBJ bytes into the flat, non-indexed triangle arrays
// raylib wants. It is deliberately minimal: the menu models are drawn as
// untextured wireframes, so positions, normals and texture coordinates are all
// that matter and materials are ignored. texs and norms are nil when the file
// carries no vt or vn at all.
func parseOBJ(data []byte) (verts, texs, norms []float32, err error) {
	var (
		pos, uv, norm []float32 // v / vt / vn, in file order
	)

	sc := bufio.NewScanner(bytes.NewReader(data))
	for sc.Scan() {
		fields := strings.Fields(sc.Text())
		if len(fields) < 2 {
			continue
		}
		args := fields[1:]
		switch fields[0] {
		case "v":
			pos = append(pos, objFloats(args, 3)...)
		case "vt":
			uv = append(uv, objFloats(args, 2)...)
		case "vn":
			norm = append(norm, objFloats(args, 3)...)
		case "f":
			// A face is any convex polygon - Blender exports quads here - so
			// fan-triangulate it into the flat arrays raylib expects.
			for i := 2; i < len(args); i++ {
				for _, corner := range [3]string{args[0], args[i-1], args[i]} {
					v, t, n := objRef(corner, len(pos)/3, len(uv)/2, len(norm)/3)
					verts = appendElem(verts, pos, v, 3)
					texs = appendElem(texs, uv, t, 2)
					norms = appendElem(norms, norm, n, 3)
				}
			}
		}
	}
	if err := sc.Err(); err != nil {
		return nil, nil, nil, err
	}
	if len(verts) == 0 {
		return nil, nil, nil, errors.New("no faces")
	}
	if len(uv) == 0 {
		texs = nil // every corner was a zero placeholder; drop it
	}
	if len(norm) == 0 {
		norms = nil
	}
	return verts, texs, norms, nil
}

// objFloats parses the first n numbers of an OBJ statement. Anything missing or
// malformed stays zero rather than failing the whole model.
func objFloats(args []string, n int) []float32 {
	out := make([]float32, n)
	for i := 0; i < n && i < len(args); i++ {
		if f, err := strconv.ParseFloat(args[i], 32); err == nil {
			out[i] = float32(f)
		}
	}
	return out
}

// objRef resolves one "v/vt/vn" face corner to zero-based indices. OBJ indices
// are 1-based, may be omitted, and may be negative to count back from the end of
// what has been read so far. -1 means "not referenced".
func objRef(corner string, nv, nt, nn int) (v, t, n int) {
	parts := strings.Split(corner, "/")
	at := func(i, count int) int {
		if i >= len(parts) || parts[i] == "" {
			return -1
		}
		x, err := strconv.Atoi(parts[i])
		if err != nil || x == 0 {
			return -1
		}
		if x < 0 {
			return count + x
		}
		return x - 1
	}
	return at(0, nv), at(1, nt), at(2, nn)
}

// appendElem copies src's idx'th n-component element onto dst, or n zeros when
// the corner referenced nothing, so all three arrays stay in lockstep.
func appendElem(dst, src []float32, idx, n int) []float32 {
	if idx < 0 || (idx+1)*n > len(src) {
		return append(dst, make([]float32, n)...)
	}
	return append(dst, src[idx*n:(idx+1)*n]...)
}
