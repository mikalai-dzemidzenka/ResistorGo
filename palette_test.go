package resistor

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"slices"
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// The Endesga 64 table has to match the published palette exactly: a truncated
// paste or an edited entry would silently change every menu colour.
func TestEndesgaPalette(t *testing.T) {
	if len(EndesgaPalette) != 64 {
		t.Fatalf("EndesgaPalette has %d colours, want 64", len(EndesgaPalette))
	}
	seen := map[string]int{}
	for i, c := range EndesgaPalette {
		if c.A != 255 {
			t.Errorf("colour %d has alpha %d, want 255", i, c.A)
		}
		h := fmt.Sprintf("%02x%02x%02x", c.R, c.G, c.B)
		if j, dup := seen[h]; dup {
			t.Errorf("colour %d (%s) repeats colour %d", i, h, j)
		}
		seen[h] = i
	}
	// the published list's first and last entries
	for _, c := range []struct {
		i    int
		want string
	}{{0, "ff0040"}, {63, "571c27"}} {
		got := fmt.Sprintf("%02x%02x%02x", EndesgaPalette[c.i].R, EndesgaPalette[c.i].G, EndesgaPalette[c.i].B)
		if got != c.want {
			t.Errorf("colour %d is %s, want %s", c.i, got, c.want)
		}
	}
}

// A texture's opaque pixels bring in their nearest Endesga colours; transparent
// ones bring in nothing, and the fixed HUD colours are always there.
func TestLevelPaletteFromTextures(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 2, 1))
	img.Set(0, 0, color.NRGBA{60, 60, 60, 255}) // nearest is 3d3d3d
	img.Set(1, 0, color.NRGBA{0, 255, 0, 0})    // transparent
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(t.TempDir(), "block.png")
	if err := os.WriteFile(p, buf.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
	base := levelPalette(nil)
	got := levelPalette([]string{p})
	if len(got) != len(base)+1 || got[len(got)-1] != (rl.Color{61, 61, 61, 255}) {
		t.Errorf("palette %v, want %v plus 3d3d3d", got, base)
	}
	if !slices.Contains(base, energyBarFill) {
		t.Error("the fixed colours are missing")
	}
}

// Close entries gain the colour between them; far ones do not, and the
// original entries keep their order at the front.
func TestWithMidpoints(t *testing.T) {
	a, b, far := rl.Color{60, 60, 60, 255}, rl.Color{66, 76, 110, 255}, rl.Color{255, 255, 255, 255}
	got := withMidpoints([]rl.Color{a, b, far}, 64)
	want := []rl.Color{a, b, far, {63, 68, 85, 255}}
	if !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}
