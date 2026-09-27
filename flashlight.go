package resistor

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// FlashlightKey turns the pistol flashlight on or off from inside a match. It
// flips the same setting the options screen shows.
const FlashlightKey = rl.KeyF3

const (
	flashlightRange     = 256 // px the beam reaches
	flashlightLength    = 0.7 // the beam's half length, as a share of the range
	flashlightWidth     = 0.1 // its half width, as a share of the range
	flashlightIntensity = 1   // over 1 the near beam saturates and the far end lifts
)

// flashlightColor is a warm white, so the beam lands on the palette's yellows.
var flashlightColor = rl.Color{R: 255, G: 235, B: 170, A: 255}

// flashlightMask is where a beam is cut out of a round light, one beam at a
// time, at the resolution the lights are computed at.
var flashlightMask rl.RenderTexture2D

func toggleFlashlight() {
	Cfg.Flashlight = !Cfg.Flashlight
	SaveSettings()
}

type beam struct {
	at, dir rl.Vector2
}

// flashlightBeams is a beam from the muzzle along the aim of every living
// fighter holding the pistol, while the setting is on. Presentation only: each
// machine lights what it shows, and nothing of it is replicated.
func flashlightBeams(actors []*Actor) []beam {
	if !Cfg.Flashlight {
		return nil
	}
	var bs []beam
	for _, a := range actors {
		if a.Weapon != WeaponPistol || a.State == StateDeath {
			continue
		}
		// the middle of the muzzle pixel, not its corner
		at := framePoint(a.drawPos(), a.Orientation, muzzleAt[aimFrame(a.Aim)])
		bs = append(bs, beam{rl.Vector2{X: at.X + 0.5, Y: at.Y + 0.5}, a.shotDir()})
	}
	return bs
}

// addFlashlights adds every beam to the light map computeLights built. Each is
// a round shadowed light at the muzzle, multiplied by an ellipse along the aim
// with its near end at the muzzle, so the beam is an oval pool rather than a
// wedge, then added like any other light.
func addFlashlights(actors []*Actor, occluders []rl.Rectangle, cam rl.Camera2D) {
	view := viewRect(cam, float32(LightAccum.Texture.Width), float32(LightAccum.Texture.Height))
	for _, b := range flashlightBeams(actors) {
		if !lightVisible(LevelLight{X: int32(b.at.X), Y: int32(b.at.Y), Radius: flashlightRange}, view) {
			continue
		}
		sharedLight.Radius = flashlightRange
		sharedLight.Color = flashlightColor
		sharedLight.Intensity = flashlightIntensity
		tex := sharedLight.Compute(Sh, b.at, occluders)
		size := float32(tex.Width)

		rl.BeginTextureMode(flashlightMask)
		rl.ClearBackground(rl.Black)
		drawBeamShape(size, b.dir)
		rl.SetBlendFactors(rl.DstColor, rl.Zero, rl.FuncAdd) // mask * light
		rl.BeginBlendMode(rl.BlendCustom)
		rl.DrawTexturePro(tex, rl.NewRectangle(0, 0, size, -size), rl.NewRectangle(0, 0, size, size), rl.Vector2{}, 0, rl.White)
		rl.EndBlendMode()
		rl.EndTextureMode()

		rl.BeginTextureMode(LightAccum)
		rl.BeginMode2D(cam)
		rl.BeginBlendMode(rl.BlendAdditive)
		addLight(flashlightMask.Texture, b.at, flashlightRange)
		rl.EndBlendMode()
		rl.EndMode2D()
		rl.EndTextureMode()
	}
}

// drawBeamShape draws the beam's ellipse into a size x size mask whose centre is
// the muzzle: nested ellipses that all start at the muzzle, the smallest white
// and each larger one a step greyer, so the beam is brightest at the gun and
// its far edge is soft.
func drawBeamShape(size float32, dir rl.Vector2) {
	a := size / 2 * flashlightLength
	b := size / 2 * flashlightWidth
	rl.PushMatrix()
	rl.Translatef(size/2, size/2, 0)
	rl.Rotatef(float32(math.Atan2(float64(dir.Y), float64(dir.X))*180/math.Pi), 0, 0, 1)
	for _, k := range []float32{1, 0.8, 0.6, 0.4} { // outermost and dimmest first
		v := uint8(255 * (1.4 - k))
		// centred its own half length ahead, so its near tip is on the muzzle
		rl.PushMatrix()
		rl.Translatef(a*k, 0, 0)
		rl.DrawEllipse(0, 0, a*k, b*k, rl.Color{R: v, G: v, B: v, A: 255})
		rl.PopMatrix()
	}
	rl.PopMatrix()
}
