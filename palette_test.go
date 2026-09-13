package resistor

import "fmt"

import "testing"

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
