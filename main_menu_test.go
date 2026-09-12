package resistor

import "testing"

func TestMenuLayout(t *testing.T) {
	const winW, winH = SW * WindowScale, SH * WindowScale
	if winW%menuPixel != 0 || winH%menuPixel != 0 {
		t.Errorf("menuPixel %d does not divide the %dx%d window evenly", menuPixel, winW, winH)
	}
	for i := range 6 { // the host screen uses six rows
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
