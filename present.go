package resistor

import rl "github.com/gen2brain/raylib-go/raylib"

// VW x VH is the virtual screen: everything is drawn into it at this fixed size,
// then scaled to fit the window. The view never depends on the window size.
const (
	VW = SW * WindowScale
	VH = SH * WindowScale
)

var VirtualScreen rl.RenderTexture2D

// dstRect is where the virtual screen lands in the window, letterboxed.
var dstRect = rl.NewRectangle(0, 0, VW, VH)

func initPresent() {
	VirtualScreen = rl.LoadRenderTexture(VW, VH)
	letterbox()
}

// letterbox recomputes the destination rectangle for the current window and
// points the mouse at virtual coordinates, so raygui and the cursor agree with
// what is drawn.
func letterbox() {
	dstRect = letterboxFor(float32(rl.GetScreenWidth()), float32(rl.GetScreenHeight()))
	scale := dstRect.Width / VW
	rl.SetMouseOffset(int(-dstRect.X), int(-dstRect.Y))
	rl.SetMouseScale(1/scale, 1/scale)
	rl.SetTextureFilter(VirtualScreen.Texture, presentFilter(scale))
}

// presentFilter picks how the finished virtual screen is sampled into the
// window. At a whole number scale every virtual pixel covers the same square of
// window pixels, so point sampling stays exact and crisp. At 1.25x or 1.5x -
// what the 1600x900 and 1920x1080 window sizes come to - point sampling would
// blow up some rows of pixels and not others; bilinear spreads that unevenness
// into a slight softness instead of a visible stripe pattern.
func presentFilter(scale float32) rl.TextureFilterMode {
	if scale == roundf(scale) {
		return rl.FilterPoint
	}
	return rl.FilterBilinear
}

// letterboxFor centres a VW x VH image in a winW x winH window at the largest
// scale that fits, keeping its aspect ratio.
func letterboxFor(winW, winH float32) rl.Rectangle {
	scale := min(winW/VW, winH/VH)
	w, h := VW*scale, VH*scale
	return rl.NewRectangle((winW-w)/2, (winH-h)/2, w, h)
}

// present draws the finished virtual screen to the window, with black bars.
func present() {
	rl.BeginDrawing()
	rl.ClearBackground(rl.Black)
	src := rl.NewRectangle(0, 0, VW, -VH) // флип render-текстуры
	rl.DrawTexturePro(VirtualScreen.Texture, src, dstRect, rl.Vector2{}, 0, rl.White)
	rl.EndDrawing()
}
