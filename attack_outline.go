package resistor

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// The outline marks the fighter the local player's next swing would strike,
// across the band of its body that swing would land on. It says where the
// swing goes and that it reaches - never whether it would be blocked, which
// stays the player's read.
const spriteFrame = 48 // one animation frame, in world pixels

// An exact palette entry, so the palette pass leaves it unchanged where the
// mark is at full strength. The fade takes it through the palette's darker reds
// as it thins out, which is what gives the gradient its steps.
var attackOutlineColor = rl.Red

// outlineTarget reports whether aimer's swing marks target this frame. It is
// the drawing rule in one place: there is a fighter aiming, it is not the one
// being marked, it is no ally, neither is dead, the aim is an attack height rather than
// block-all, time is not showing a recorded frame, and the swing reaches.
//
// It reads only replicated state, so a client marks what the server would.
func outlineTarget(aimer, target *Actor) bool {
	if aimer == nil || target == nil || aimer == target {
		return false
	}
	if aimer.allyOf(target) {
		return false // a swing passes through an ally, so it is no target
	}
	if aimer.State == StateDeath || target.State == StateDeath {
		return false
	}
	if TimeShowsHistory() {
		return false
	}
	box, ok := aimer.attackBoxAt(aimer.Updown, aimer.drawPos())
	if !ok {
		return false // block-all has no attack to aim with
	}
	return aabb(box, target.outlineRect())
}

// outlineRect is the target's body as the outline marks it: the collision box,
// placed from the position the sprite is drawn at so the outline tracks the
// interpolated fighter rather than the simulated one.
func (a *Actor) outlineRect() rl.Rectangle {
	pos := a.drawPos()
	return rl.Rectangle{
		X:      pos.X + a.CollisionBox.X,
		Y:      pos.Y + a.CollisionBox.Y,
		Width:  a.CollisionBox.Width,
		Height: a.CollisionBox.Height,
	}
}

// attackOutlineBand is the slice of the marked fighter the swing would take, in
// whole world pixels: the swing's own vertical extent at the aimer's height,
// across the fighter's sprite. The outline traces the body inside this band;
// the band decides which rows of it are marked, not what shape the mark is.
func attackOutlineBand(aimer, target *Actor) rl.Rectangle {
	box, _ := aimer.attackBoxAt(aimer.Updown, aimer.drawPos())
	pos := target.drawPos()
	return rl.Rectangle{
		X:      roundf(pos.X),
		Y:      roundf(box.Y),
		Width:  spriteFrame,
		Height: roundf(box.Height),
	}
}

// drawAttackOutline marks target if the local player's swing would strike it,
// by drawing its sprite a second time through the outline shader: only the
// transparent pixels touching the body, and only across the band, survive.
//
// It runs in the overlay pass, above the lighting, so a marked fighter in
// shadow carries the same outline as one under a lamp.
func drawAttackOutline(aimer, target *Actor) {
	if !outlineTarget(aimer, target) {
		return
	}
	tex, src, pos := target.spriteDraw()
	if tex.ID == 0 || tex.Width == 0 || tex.Height == 0 || src.Height == 0 {
		return
	}
	band := attackOutlineBand(aimer, target)

	// The band, as a range of the sprite's own height. The shader is given the
	// range unclipped, so the fade stays centred on the row the swing lands on
	// even when the band runs past the top or bottom of the sprite; clipping it
	// is only how we tell that it misses the sprite altogether.
	top := (band.Y - pos.Y) / src.Height
	bottom := (band.Y + band.Height - pos.Y) / src.Height
	frameBottom := src.Height / float32(tex.Height)
	if clamp(bottom, 0, frameBottom)-clamp(top, 0, frameBottom) <= 0 {
		return // the swing passes over or under the sprite entirely
	}

	// this frame's own rectangle in the sheet, so the outline cannot sample the
	// pose next to it. The source width is negative when the fighter faces
	// left, but the frame occupies the same columns either way.
	texW, texH := float32(tex.Width), float32(tex.Height)
	left := src.X / texW
	right := (src.X + float32(math.Abs(float64(src.Width)))) / texW

	set := func(name string, v []float32, kind rl.ShaderUniformDataType) {
		rl.SetShaderValue(OutlineShader, rl.GetShaderLocation(OutlineShader, name), v, kind)
	}
	rl.BeginShaderMode(OutlineShader)
	set("texelSize", []float32{1 / texW, 1 / texH}, rl.ShaderUniformVec2)
	set("frameBounds", []float32{left, 0, right, frameBottom}, rl.ShaderUniformVec4)
	set("band", []float32{top, bottom}, rl.ShaderUniformVec2)
	set("outlineColor", []float32{
		float32(attackOutlineColor.R) / 255,
		float32(attackOutlineColor.G) / 255,
		float32(attackOutlineColor.B) / 255,
		float32(attackOutlineColor.A) / 255,
	}, rl.ShaderUniformVec4)
	rl.DrawTextureRec(tex, src, pos, rl.White)
	rl.EndShaderMode()
}
