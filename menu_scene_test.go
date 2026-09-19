package resistor

import (
	"slices"
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// hueDegrees is a colour's hue in degrees, 0 being pure red.
func hueDegrees(c rl.Color) float64 {
	r, g, b := float64(c.R)/255, float64(c.G)/255, float64(c.B)/255
	mx := max(r, max(g, b))
	mn := min(r, min(g, b))
	d := mx - mn
	if d == 0 {
		return 0
	}
	var h float64
	switch mx {
	case r:
		h = (g - b) / d
		if h < 0 {
			h += 6
		}
	case g:
		h = (b-r)/d + 2
	default:
		h = (r-g)/d + 4
	}
	return h * 60
}

// fromRed is how far a colour's hue is from pure red, in degrees.
func fromRed(c rl.Color) float64 {
	h := hueDegrees(c)
	return min(h, 360-h)
}

// Every colour the scene is painted with has to be a palette red: the scene is
// drawn only from this one list, so nothing can drift off-red unnoticed.
func TestSceneColorsArePaletteReds(t *testing.T) {
	reds := map[rl.Color]bool{}
	for _, i := range menuRedIndexes {
		c := EndesgaPalette[i]
		if d := fromRed(c); d > 20 {
			t.Errorf("index %d (%02x%02x%02x) is %.0f degrees from red, past the 20 the scene allows", i, c.R, c.G, c.B, d)
		}
		reds[c] = true
	}

	used := map[string][]rl.Color{
		"the sky":       {menuSky, menuSkyTop},
		"the sun":       nil,
		"the ground":    {menuGround},
		"the rain":      {menuRain, menuRainTip},
		"the models":    {menuModelHead, menuModelJab, menuModelKick},
		"the grid ramp": nil,
		"the chrome": {menuNeon, menuAccent, menuBright, menuMuted, menuDim,
			menuPanelCol, menuRowFill, menuRowFillOn},
	}
	for _, i := range menuSunColors {
		used["the sun"] = append(used["the sun"], EndesgaPalette[i])
	}
	for _, i := range neonRamps[neonGrid] {
		used["the grid ramp"] = append(used["the grid ramp"], EndesgaPalette[i])
	}
	for what, cs := range used {
		if len(cs) == 0 {
			t.Errorf("%s uses no colours at all", what)
		}
		for _, c := range cs {
			if !reds[c] {
				t.Errorf("%s uses %02x%02x%02x, which is not one of the scene's reds", what, c.R, c.G, c.B)
			}
		}
	}
}

// Cycling steps between palette reds, so a captured frame can never hold a
// colour from outside the palette, and every role does change.
func TestNeonColorsStayInPalette(t *testing.T) {
	for role := neonRole(0); role < neonRoleCount; role++ {
		first := neonColor(role, 0)
		changed := false
		for step := 0; step < 2000; step++ {
			tt := float64(step) * 0.05 // 100 seconds, finer than one ramp step
			c := neonColor(role, tt)
			if !slices.Contains(EndesgaPalette, c) {
				t.Errorf("role %d at t=%.2f gives %+v, which is not an Endesga 64 colour", role, tt, c)
				break
			}
			if d := fromRed(c); d > 20 {
				t.Errorf("role %d at t=%.2f gives a colour %.0f degrees from red", role, tt, d)
				break
			}
			if tt <= 10 && c != first {
				changed = true
			}
		}
		if !changed {
			t.Errorf("role %d never changes colour within ten seconds", role)
		}
		if neonColor(role, -1) != neonColor(role, -1) {
			t.Errorf("role %d is not stable for a negative time", role)
		}
	}
}
