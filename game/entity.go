package game

var _ GUIDrawable = (*baseGui)(nil)

type baseGui struct {
	guiZ int8
}

func (b baseGui) DrawGui() {}

func (b baseGui) GuiZ() int8 {
	return b.guiZ
}

var _ Drawable = (*baseDraw)(nil)

type baseDraw struct {
	z int8
}

func (b baseDraw) Draw() {}
func (b baseDraw) Z() int8 {
	return b.z
}
