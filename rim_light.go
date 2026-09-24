package resistor

import (
	"math"
	"slices"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// The rim light lights the edge of a fighter's silhouette that faces a level
// light, in that light's colour, the way Katana Zero separates its characters
// from the dark. It is lit light added on top of the composited scene, so the
// ambient dark cannot dim it, and it paints only the fighter's own outermost
// pixels.

// rimMaxLights is how many lights may rim one fighter at once; the shader's
// uniform arrays are this long.
const rimMaxLights = 4

// rimMinStrength is how weak a light may be at a fighter and still rim it.
// Below it the rim would vanish into the palette anyway.
const rimMinStrength = 0.02

// rimStrength scales every rim: how bright the edge gets compared with the
// light's own brightness at the fighter.
const rimStrength = 1.0

// rimCap is the most a rim adds, as a fraction of white, however strong the
// lights reaching the fighter are together. Far from a light a rim is well under
// it and unchanged; close to one, or inside several, the whole rim is scaled
// down to it, keeping its colour. Without it a fighter beside a lamp was traced
// in solid white.
const rimCap = 0.35

// rimLight is one light that rims a fighter: where it is in world pixels, its
// colour, and how strong it is at the fighter.
type rimLight struct {
	pos      rl.Vector2
	color    rl.Color
	strength float32
}

// rimStrengthAt is how strong l is at p: its intensity with the falloff the
// light shader uses, (1 - d/R)².
func rimStrengthAt(l LevelLight, p rl.Vector2) float32 {
	if l.Radius <= 0 || l.Intensity <= 0 {
		return 0
	}
	d := rl.Vector2Distance(l.center(), p)
	f := max(1-d/float32(l.Radius), 0)
	return l.Intensity * f * f
}

// segmentHitsRect reports whether the segment from a to b touches r, by
// clipping the segment against the rectangle's two slabs. A segment starting
// inside r hits it.
func segmentHitsRect(a, b rl.Vector2, r rl.Rectangle) bool {
	t0, t1 := float32(0), float32(1)
	clip := func(p, d, lo, hi float32) bool {
		if d == 0 {
			return p >= lo && p <= hi // parallel: inside the slab or never
		}
		ta, tb := (lo-p)/d, (hi-p)/d
		if ta > tb {
			ta, tb = tb, ta
		}
		t0, t1 = max(t0, ta), min(t1, tb)
		return t0 <= t1
	}
	return clip(a.X, b.X-a.X, r.X, r.X+r.Width) &&
		clip(a.Y, b.Y-a.Y, r.Y, r.Y+r.Height)
}

// rimLights picks the lights that rim a fighter whose body is body: measured at
// the body's centre, strong enough to show, not shadowed by any occluder, and
// at most rimMaxLights of them, strongest first.
func rimLights(lights []LevelLight, occluders []rl.Rectangle, body rl.Rectangle) []rimLight {
	centre := rl.Vector2{X: body.X + body.Width/2, Y: body.Y + body.Height/2}
	var result []rimLight
	for _, l := range lights {
		s := rimStrengthAt(l, centre)
		if s < rimMinStrength {
			continue
		}
		if slices.ContainsFunc(occluders, func(r rl.Rectangle) bool {
			return segmentHitsRect(l.center(), centre, r)
		}) {
			continue
		}
		result = append(result, rimLight{pos: l.center(), color: l.color(), strength: s})
	}
	slices.SortStableFunc(result, func(a, b rimLight) int {
		switch {
		case a.strength > b.strength:
			return -1
		case a.strength < b.strength:
			return 1
		}
		return 0
	})
	if len(result) > rimMaxLights {
		result = result[:rimMaxLights]
	}
	return result
}

// rimLocal puts a light into the sprite frame's own pixels: relative to where
// the sprite is drawn, and mirrored when the fighter faces left, because a
// negative source width reverses the frame across the quad - the texel drawn at
// the left of the screen is the frame's right-hand one.
func rimLocal(light, spritePos rl.Vector2, frameWidth float32, facingLeft bool) rl.Vector2 {
	p := rl.Vector2{X: light.X - spritePos.X, Y: light.Y - spritePos.Y}
	if facingLeft {
		p.X = frameWidth - p.X
	}
	return p
}

// rimLoc holds RimShader's uniform locations, looked up once when it loads.
var rimLoc struct {
	texelSize, frameBounds              int32
	topTexture, topTexelSize, topBounds int32
	hasTop                              int32
	lightCount, lightPos, lightColor    int32
	cap                                 int32
	bodies, bodiesSize, bodyId          int32
}

// BodyMask holds every fighter's silhouette on screen, each as its own id,
// drawn with the same camera and size as the scene so a rim pixel finds the
// body under it at its own coordinates.
var (
	BodyMask       rl.RenderTexture2D
	BodyMaskShader rl.Shader
	bodyMaskID     int32
)

// maxBodies is how many fighters the mask can tell apart: an id is one byte.
const maxBodies = 255

func bodyID(i int) float32 { return float32(i+1) / 255 }

func loadRimShader() {
	RimShader = rl.LoadShaderFromMemory("", fsRim)
	loc := func(name string) int32 {
		l := rl.GetShaderLocation(RimShader, name)
		if l < 0 {
			rl.TraceLog(rl.LogWarning, "rim: shader has no uniform %s", name)
		}
		return l
	}
	rimLoc.texelSize = loc("texelSize")
	rimLoc.frameBounds = loc("frameBounds")
	rimLoc.topTexture = loc("topTexture")
	rimLoc.topTexelSize = loc("topTexelSize")
	rimLoc.topBounds = loc("topBounds")
	rimLoc.hasTop = loc("hasTop")
	rimLoc.lightCount = loc("lightCount")
	rimLoc.lightPos = loc("lightPos")
	rimLoc.lightColor = loc("lightColor")
	rimLoc.cap = loc("rimCap")
	rimLoc.bodies = loc("bodies")
	rimLoc.bodiesSize = loc("bodiesSize")
	rimLoc.bodyId = loc("bodyId")
	BodyMaskShader = rl.LoadShaderFromMemory("", fsBodyMask)
	bodyMaskID = rl.GetShaderLocation(BodyMaskShader, "bodyId")
}

// drawBodyMask draws each fighter's silhouette into BodyMask as its index in
// actors, the order drawRim is given them in. It runs before the scene pass:
// render textures do not nest.
func drawBodyMask(actors []*Actor, cam rl.Camera2D) {
	rl.BeginTextureMode(BodyMask)
	rl.ClearBackground(rl.Blank)
	rl.BeginMode2D(cam)
	rl.BeginShaderMode(BodyMaskShader)
	for i, a := range actors[:min(len(actors), maxBodies)] {
		if a.deathHidden() {
			continue
		}
		// a batch per fighter, so each is drawn with its own id
		rl.SetShaderValue(BodyMaskShader, bodyMaskID, []float32{bodyID(i)}, rl.ShaderUniformFloat)
		tex, src, pos := a.spriteDraw()
		rl.DrawTextureRec(tex, src, pos, rl.White)
		if tex, src, pos, ok := a.topDraw(); ok {
			rl.DrawTextureRec(tex, src, pos, rl.White)
		}
		rl.DrawRenderBatchActive()
	}
	rl.EndShaderMode()
	rl.EndMode2D()
	rl.EndTextureMode()
}

// rimQuad grows a sprite draw by one pixel on every side, so the ring of empty
// pixels around the fighter is inside the quad and the shader can paint it. The
// negative source width that mirrors a left-facing fighter is kept: the region
// grows the same either way, and it is the sign that flips it.
func rimQuad(src rl.Rectangle, pos rl.Vector2) (rl.Rectangle, rl.Rectangle) {
	w := float32(math.Abs(float64(src.Width)))
	grown := rl.Rectangle{X: src.X - 1, Y: src.Y - 1, Width: w + 2, Height: src.Height + 2}
	if src.Width < 0 {
		grown.Width = -grown.Width
	}
	dst := rl.Rectangle{X: pos.X - 1, Y: pos.Y - 1, Width: w + 2, Height: src.Height + 2}
	return grown, dst
}

// drawRim draws the ring of light around a's silhouette, lit by the level
// lights that reach it: the sprite's quad again, a pixel larger all round, through
// RimShader. i is a's place in the list drawBodyMask was given, so the ring stops
// at every other fighter's body. The caller has the blend mode set to additive
// and the camera applied.
func drawRim(a *Actor, i int, lights []LevelLight, occluders []rl.Rectangle) {
	if a.deathHidden() || i >= maxBodies {
		return
	}
	tex, src, pos := a.spriteDraw()
	if tex.ID == 0 || tex.Width == 0 || tex.Height == 0 || src.Height == 0 {
		return
	}
	rims := rimLights(lights, occluders, a.outlineRect())
	if len(rims) == 0 {
		return
	}

	texW, texH := float32(tex.Width), float32(tex.Height)
	frameW := float32(math.Abs(float64(src.Width)))
	left := src.X / texW
	right := (src.X + frameW) / texW
	bottom := (src.Y + src.Height) / texH

	var lp [2 * rimMaxLights]float32
	var lc [3 * rimMaxLights]float32
	for i, r := range rims {
		p := rimLocal(r.pos, pos, frameW, src.Width < 0)
		lp[2*i], lp[2*i+1] = p.X, p.Y
		k := r.strength * rimStrength / 255
		lc[3*i] = float32(r.color.R) * k
		lc[3*i+1] = float32(r.color.G) * k
		lc[3*i+2] = float32(r.color.B) * k
	}

	grown, dst := rimQuad(src, pos)
	rl.BeginShaderMode(RimShader)
	rl.SetShaderValue(RimShader, rimLoc.texelSize, []float32{1 / texW, 1 / texH}, rl.ShaderUniformVec2)
	rl.SetShaderValue(RimShader, rimLoc.frameBounds, []float32{left, src.Y / texH, right, bottom}, rl.ShaderUniformVec4)
	setRimTop(a)
	rl.SetShaderValue(RimShader, rimLoc.lightCount, []float32{float32(len(rims))}, rl.ShaderUniformFloat)
	rl.SetShaderValueV(RimShader, rimLoc.lightPos, lp[:], rl.ShaderUniformVec2, rimMaxLights)
	rl.SetShaderValueV(RimShader, rimLoc.lightColor, lc[:], rl.ShaderUniformVec3, rimMaxLights)
	rl.SetShaderValue(RimShader, rimLoc.cap, []float32{rimCap}, rl.ShaderUniformFloat)
	rl.SetShaderValue(RimShader, rimLoc.bodyId, []float32{bodyID(i)}, rl.ShaderUniformFloat)
	rl.SetShaderValue(RimShader, rimLoc.bodiesSize, []float32{float32(BodyMask.Texture.Width), float32(BodyMask.Texture.Height)}, rl.ShaderUniformVec2)
	rl.SetShaderValueTexture(RimShader, rimLoc.bodies, BodyMask.Texture)
	rl.DrawTexturePro(tex, grown, dst, rl.Vector2{}, 0, rl.White)
	rl.EndShaderMode()
}

// setRimTop hands the shader the fighter's top layer, so the rim runs around
// the body and the layer as one silhouette. With no layer the sampler is left
// switched off and the silhouette is the body alone.
func setRimTop(a *Actor) {
	tex, src, _, ok := a.topDraw()
	if !ok || tex.ID == 0 || tex.Width == 0 || tex.Height == 0 {
		rl.SetShaderValue(RimShader, rimLoc.hasTop, []float32{0}, rl.ShaderUniformFloat)
		return
	}
	texW, texH := float32(tex.Width), float32(tex.Height)
	frameW := float32(math.Abs(float64(src.Width)))
	rl.SetShaderValue(RimShader, rimLoc.hasTop, []float32{1}, rl.ShaderUniformFloat)
	rl.SetShaderValue(RimShader, rimLoc.topTexelSize, []float32{1 / texW, 1 / texH}, rl.ShaderUniformVec2)
	rl.SetShaderValue(RimShader, rimLoc.topBounds, []float32{
		src.X / texW, src.Y / texH, (src.X + frameW) / texW, (src.Y + src.Height) / texH,
	}, rl.ShaderUniformVec4)
	rl.SetShaderValueTexture(RimShader, rimLoc.topTexture, tex)
}
