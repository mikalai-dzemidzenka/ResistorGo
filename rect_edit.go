package resistor

import rl "github.com/gen2brain/raylib-go/raylib"

// handle is the part of a rectangle the cursor has hold of.
type handle int

const (
	handleNone handle = iota
	handleBody
	handleLeft
	handleRight
	handleTop
	handleBottom
	handleTopLeft
	handleTopRight
	handleBottomLeft
	handleBottomRight
)

func (h handle) hasLeft() bool {
	return h == handleLeft || h == handleTopLeft || h == handleBottomLeft
}

func (h handle) hasRight() bool {
	return h == handleRight || h == handleTopRight || h == handleBottomRight
}

func (h handle) hasTop() bool {
	return h == handleTop || h == handleTopLeft || h == handleTopRight
}

func (h handle) hasBottom() bool {
	return h == handleBottom || h == handleBottomLeft || h == handleBottomRight
}

// rectHandleAt reports what the point p has hold of on r: a corner first, then
// an edge, then the body, then nothing. tol is the grab tolerance in world
// pixels - the caller derives it from the zoom, so a handle stays the same size
// on screen however far in the camera is.
func rectHandleAt(r rl.Rectangle, p rl.Vector2, tol float32) handle {
	near := func(a, b float32) bool { return abs(a-b) <= tol }
	left, right := near(p.X, r.X), near(p.X, r.X+r.Width)
	top, bottom := near(p.Y, r.Y), near(p.Y, r.Y+r.Height)
	if p.X < r.X-tol || p.X > r.X+r.Width+tol || p.Y < r.Y-tol || p.Y > r.Y+r.Height+tol {
		return handleNone
	}
	switch {
	case left && top:
		return handleTopLeft
	case right && top:
		return handleTopRight
	case left && bottom:
		return handleBottomLeft
	case right && bottom:
		return handleBottomRight
	case left:
		return handleLeft
	case right:
		return handleRight
	case top:
		return handleTop
	case bottom:
		return handleBottom
	case p.X >= r.X && p.X <= r.X+r.Width && p.Y >= r.Y && p.Y <= r.Y+r.Height:
		return handleBody
	}
	return handleNone
}

// handleAnchor is the point on r that h holds: the corner or edge being
// dragged, or the rectangle's own corner when the whole body is being moved.
// On an axis the handle does not constrain the cursor's own position is used,
// so the grab offset on that axis is zero.
func handleAnchor(r rl.Rectangle, h handle, p rl.Vector2) rl.Vector2 {
	if h == handleBody {
		return rl.Vector2{X: r.X, Y: r.Y}
	}
	a := p
	if h.hasLeft() {
		a.X = r.X
	}
	if h.hasRight() {
		a.X = r.X + r.Width
	}
	if h.hasTop() {
		a.Y = r.Y
	}
	if h.hasBottom() {
		a.Y = r.Y + r.Height
	}
	return a
}

// dragRect moves what h holds to the point to. A body drag translates; an edge
// or a corner moves only that side, clamped against the opposite one so a
// rectangle can never be dragged thinner than a pixel, nor inside out.
func dragRect(r rl.Rectangle, h handle, to rl.Vector2) rl.Rectangle {
	if h == handleBody {
		r.X, r.Y = to.X, to.Y
		return r
	}
	if h.hasLeft() {
		x := min(to.X, r.X+r.Width-1)
		r.Width += r.X - x
		r.X = x
	}
	if h.hasRight() {
		r.Width = max(1, to.X-r.X)
	}
	if h.hasTop() {
		y := min(to.Y, r.Y+r.Height-1)
		r.Height += r.Y - y
		r.Y = y
	}
	if h.hasBottom() {
		r.Height = max(1, to.Y-r.Y)
	}
	return r
}

// A rectangle is stored as an offset from its object's origin, and dragged
// where it is drawn - in the level. These two convert between the two, in whole
// pixels: rounding, not truncating, because half the level is on the negative
// side of the origin and truncation there loses a pixel every drag.
func rectToWorld(r rl.Rectangle, origin rl.Vector2) rl.Rectangle {
	r.X += origin.X
	r.Y += origin.Y
	return r
}

func rectToLocal(r rl.Rectangle, origin rl.Vector2) rl.Rectangle {
	r.X = roundf(r.X - origin.X)
	r.Y = roundf(r.Y - origin.Y)
	r.Width = roundf(r.Width)
	r.Height = roundf(r.Height)
	return r
}

func roundVec2(v rl.Vector2) rl.Vector2 {
	return rl.Vector2{X: roundf(v.X), Y: roundf(v.Y)}
}
