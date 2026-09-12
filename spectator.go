package resistor

import (
	"fmt"
	rl "github.com/gen2brain/raylib-go/raylib"
	"slices"
)

const SpecPanSpeed = 400 // screen px/s, the same at every zoom

// SpecCam is the spectator's camera: free, or following one fighter.
type SpecCam struct {
	Follow   bool
	FollowID uint16     // 0 while waiting for the first fighter to become visible
	Pos      rl.Vector2 // camera target in world space
}

// SpecInput is one frame of spectator controls.
type SpecInput struct {
	Pan              rl.Vector2 // -1..1 per axis from the keys
	Drag             rl.Vector2 // mouse movement while the right button is held, screen px
	Next, Prev, Free bool
}

var Spec SpecCam

// nextFighter steps dir (+1/-1) through the sorted ids, wrapping. If cur is not
// in the list it returns the first id after (or before) it.
func nextFighter(ids []uint16, cur uint16, dir int) (uint16, bool) {
	if len(ids) == 0 {
		return 0, false
	}
	if i := slices.Index(ids, cur); i >= 0 {
		return ids[(i+dir+len(ids))%len(ids)], true
	}
	if dir >= 0 {
		for _, id := range ids {
			if id > cur {
				return id, true
			}
		}
		return ids[0], true
	}
	for j := len(ids) - 1; j >= 0; j-- {
		if ids[j] < cur {
			return ids[j], true
		}
	}
	return ids[len(ids)-1], true
}

func panStep(pos, dir rl.Vector2, zoom, dt float32) rl.Vector2 {
	step := SpecPanSpeed / zoom * dt
	return rl.Vector2{X: pos.X + dir.X*step, Y: pos.Y + dir.Y*step}
}

// dragStep keeps the world point under the cursor under the cursor.
func dragStep(pos, delta rl.Vector2, zoom float32) rl.Vector2 {
	return rl.Vector2{X: pos.X - delta.X/zoom, Y: pos.Y - delta.Y/zoom}
}

func followPoint(a *Actor) rl.Vector2 {
	return rl.Vector2{X: a.Pos.X + 24, Y: a.Pos.Y + 24}
}

// Update applies one frame of input. fighters are the visible fighters; view is
// where the camera currently looks, so leaving follow mode does not move the view.
func (c *SpecCam) Update(in SpecInput, fighters []*Actor, zoom, dt float32, view rl.Vector2) {
	ids := make([]uint16, 0, len(fighters))
	for _, a := range fighters {
		ids = append(ids, a.ID)
	}
	slices.Sort(ids)

	moving := in.Pan != (rl.Vector2{}) || in.Drag != (rl.Vector2{})
	if c.Follow && (in.Free || moving) {
		c.Follow = false
		c.Pos = view
	}
	if in.Next || in.Prev {
		dir := 1
		if in.Prev {
			dir = -1
		}
		if id, ok := nextFighter(ids, c.FollowID, dir); ok {
			c.Follow = true
			c.FollowID = id
		}
	}

	if !c.Follow {
		c.Pos = panStep(c.Pos, in.Pan, zoom, dt)
		c.Pos = dragStep(c.Pos, in.Drag, zoom)
		return
	}
	switch {
	case c.FollowID == 0:
		if len(ids) > 0 {
			c.FollowID = ids[0]
		}
	case !slices.Contains(ids, c.FollowID):
		if id, ok := nextFighter(ids, c.FollowID, 1); ok {
			c.FollowID = id
		} else {
			c.Follow = false
			c.FollowID = 0
			return
		}
	}
	for _, a := range fighters {
		if a.ID == c.FollowID {
			c.Pos = followPoint(a)
		}
	}
}

func readSpecInput() SpecInput {
	var in SpecInput
	if rl.IsKeyDown(rl.KeyA) || rl.IsKeyDown(rl.KeyLeft) {
		in.Pan.X--
	}
	if rl.IsKeyDown(rl.KeyD) || rl.IsKeyDown(rl.KeyRight) {
		in.Pan.X++
	}
	if rl.IsKeyDown(rl.KeyW) || rl.IsKeyDown(rl.KeyUp) {
		in.Pan.Y--
	}
	if rl.IsKeyDown(rl.KeyS) || rl.IsKeyDown(rl.KeyDown) {
		in.Pan.Y++
	}
	if rl.IsMouseButtonDown(rl.MouseButtonRight) {
		d := rl.GetMouseDelta()
		// the delta can jump after the window regains focus; never more than a screen per frame
		in.Drag = rl.Vector2{X: clamp(d.X, -SW*WindowScale, SW*WindowScale), Y: clamp(d.Y, -SH*WindowScale, SH*WindowScale)}
	}
	in.Next = rl.IsKeyPressed(rl.KeyE)
	in.Prev = rl.IsKeyPressed(rl.KeyQ)
	in.Free = rl.IsKeyPressed(rl.KeySpace)
	return in
}

// updateSpectatorCam drives Cam from the spectator camera: follow mode is smoothed
// like the player camera, free mode moves the view directly.
func updateSpectatorCam(dt float32, w *World) {
	Spec.Update(readSpecInput(), w.Actors, Cam.Zoom, dt, Cam.Camera2D.Target)
	if Spec.Follow {
		Cam.SetTarget(Spec.Pos)
	} else {
		Cam.Target = Spec.Pos
		Cam.Camera2D.Target = Spec.Pos
	}
}

func drawSpectatorHUD() {
	mode := "free camera"
	if Spec.Follow {
		mode = "waiting for a fighter"
		if Spec.FollowID != 0 {
			mode = fmt.Sprintf("following fighter #%d", Spec.FollowID)
		}
	}
	const x = 20
	y := int32(SH*WindowScale - 64)
	rl.DrawText("SPECTATING", x, y, 20, rl.Red)
	rl.DrawText(mode, x+150, y, 20, rl.White)
	rl.DrawText("Q/E prev/next fighter   Space free camera   WASD/arrows pan   right-drag move   wheel zoom   Esc leave", x, y+28, 20, rl.LightGray)
}
