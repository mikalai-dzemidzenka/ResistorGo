package resistor

import (
	"slices"
	"testing"
)

// The panel has to hold the whole menu column - the header, the logo at the
// scale logoScale picks, every row with its label and the error line - and the
// panel itself has to stay between the two rules the window is framed by.
func TestMenuPanelHoldsTheColumn(t *testing.T) {
	if menuPanelTop < menuRuleTop || menuPanelBottom > menuRuleBottom {
		t.Errorf("panel spans y %d..%d, outside the rules at %d and %d", menuPanelTop, menuPanelBottom, menuRuleTop, menuRuleBottom)
	}
	for i := range 6 { // the host screen uses six rows, the most of any screen
		r := menuRow(i)
		if r.X-menuLabelW < menuPanelLeft || r.X+r.Width > menuPanelW {
			t.Errorf("row %d with its label spans x %v..%v, want inside the panel at %d..%d", i, r.X-menuLabelW, r.X+r.Width, menuPanelLeft, menuPanelW)
		}
		if r.Y < menuPanelTop || r.Y+r.Height > menuPanelBottom {
			t.Errorf("row %d spans y %v..%v, want inside the panel at %d..%d", i, r.Y, r.Y+r.Height, menuPanelTop, menuPanelBottom)
		}
	}
	// the logo, at the size it is actually drawn: 209x28 pixels of art,
	// centred over the widget column
	const logoW, logoH = 209, 28
	s := logoScale(logoW)
	x := menuRowX + menuRowW/2 - logoW*s/2
	if x < menuPanelLeft || x+logoW*s > menuPanelW {
		t.Errorf("the logo spans x %v..%v, want inside the panel at %d..%d", x, x+logoW*s, menuPanelLeft, menuPanelW)
	}
	if top := menuLogoBottom - logoH*s; top < menuPanelTop {
		t.Errorf("the logo starts at y %v, above the panel at %d", top, menuPanelTop)
	}
	// the header above the logo, the last of it being the screen name
	if menuLogoBottom+8+12 > menuRow(0).Y-34 {
		t.Error("the screen name under the logo runs into the error line")
	}
}

// The rain is laid out in normalized coordinates from a fixed seed, so it is
// the same shower every run and it covers the scene whatever the scene is
// drawn at.
func TestMenuRainCoversTheScene(t *testing.T) {
	drops := makeMenuDrops(7)
	if len(drops) != menuDropCount {
		t.Fatalf("got %d streaks, want %d", len(drops), menuDropCount)
	}
	for i, d := range drops {
		if d.x < 0 || d.x > 1 || d.y < 0 || d.y > 1 {
			t.Errorf("streak %d sits at %v,%v, outside the scene", i, d.x, d.y)
		}
		if d.speed <= 0 {
			t.Errorf("streak %d falls at %v, want faster than nothing", i, d.speed)
		}
		if d.length <= 0 || d.length > menuSceneH/4 {
			t.Errorf("streak %d is %v pixels long, want between 0 and a quarter of the scene", i, d.length)
		}
	}
	if !slices.Equal(drops, makeMenuDrops(7)) {
		t.Error("the same seed gives a different shower")
	}
}

func TestMenuLayout(t *testing.T) {
	const winW, winH = SW * WindowScale, SH * WindowScale
	if winW%menuPixel != 0 || winH%menuPixel != 0 {
		t.Errorf("menuPixel %d does not divide the %dx%d window evenly", menuPixel, winW, winH)
	}
	for i := range 6 { // the host screen uses six rows, the most of any screen
		r := menuRow(i)
		if r.Y+r.Height > winH-16 {
			t.Errorf("row %d ends at %v, want <= %v", i, r.Y+r.Height, winH-16)
		}
		if r.X-menuLabelW < 16 || r.X+r.Width > menuPanelW {
			t.Errorf("row %d with its label spans x %v..%v, want inside 16..%d", i, r.X-menuLabelW, r.X+r.Width, menuPanelW)
		}
	}
	// the error line sits 34px above the first row and must stay below the logo
	if r := menuRow(0); r.Y-34 < menuLogoBottom {
		t.Errorf("error line at %v overlaps the logo ending at %v", r.Y-34, menuLogoBottom)
	}
	if s := logoScale(209); s != 2 {
		t.Errorf("logoScale(209) = %v, want 2", s)
	}
	// centred over the widget column, the logo has this much room inside the panel
	const room = 2 * (menuPanelW - menuRowX - menuRowW/2)
	for w := int32(1); w <= 2000; w++ {
		s := logoScale(w)
		if s < 1 || s > WindowScale || s != float32(int(s)) {
			t.Fatalf("logoScale(%d) = %v, want an integer in [1, %d]", w, s, WindowScale)
		}
		if s > 1 && s*float32(w) > room {
			t.Fatalf("logoScale(%d) = %v: logo %v px wide, more than the %d px in the panel", w, s, s*float32(w), room)
		}
	}
}
