package resistor

import (
	"math"
	"math/rand"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// What a bot can see and how long it keeps looking. Feel numbers, to be turned
// by eye in play.
const (
	// How far ahead a bot notices anything at all: five tiles, about half the
	// level the camera shows at the zoom it starts at.
	SightRange = 240
	// How far up and down it sees per pixel ahead - 1 is 45 degrees either side
	// of the way it faces, so a fighter is in view while it is no further above
	// or below than it is ahead. The tangent of the half-angle, so no
	// trigonometry is needed to test it.
	SightSlope = 1
	// How long a bot keeps trying to reach the spot it last saw someone at
	// before giving the search up: 3 seconds.
	SightSearchTicks = 180
	// How long it stands looking around once it gets there: 0.75 seconds.
	SightLingerTicks = 45
)

// What a bot believes its own jumps can do. Every number is derived from the
// movement constants rather than measured off a level and typed in, so retuning
// gravity, a take-off or an animation carries the belief along with it - a bot
// left believing in a jump it no longer has walks into holes.
//
// The physics is the last word: these only decide. A jump the bot commits to is
// the same jump any fighter takes.
const (
	// The bot decides with a little less than the arc actually gives it, so a
	// decision made a pixel early is still a jump it makes. It also covers the
	// difference between the smooth arc these formulas describe and the stepped
	// one the fixed timestep actually flies.
	JumpMargin = 0.85

	// What the two jumps put into Speed.Y on the tick they leave the ground,
	// in pixels a second.
	jumpTakeoff    = -MaxJumpSpeed / 60
	runJumpTakeoff = -MaxRunJumpSpeed / 60

	// How far the feet rise, v^2/2g at the top of the arc.
	JumpRise    = jumpTakeoff * jumpTakeoff / (2 * Gravity) * JumpMargin
	RunJumpRise = runJumpTakeoff * runJumpTakeoff / (2 * Gravity) * JumpMargin

	// How long the fighter is off the ground, 2v/g, in ticks.
	JumpAirTicks    = 2 * jumpTakeoff * 60 / Gravity
	RunJumpAirTicks = 2 * runJumpTakeoff * 60 / Gravity

	// The wind-up each jump spends still on the ground: the standing jump's
	// whole jump-start animation, and the run-jump's two anticipation frames.
	// Ground the fighter covers on its feet, so a bot that means to leave an
	// edge has to press its jump this long before reaching it.
	JumpWindupTicks    = 60 * JumpStartSeconds
	RunJumpWindupTicks = 60 * RunJumpSeconds * RunJumpLaunchFrame / RunJumpFrames

	// How finely the ground ahead is read. Four pixels is a step's height and
	// narrower than anything a fighter could stand on anyway.
	probeStep = 4
)

// reach is how far the bot would travel horizontally while off the ground at
// the speed it is going now. Its own speed rather than the top speed: a bot
// that has only just started moving cannot cross what a bot at a run can, and
// believing otherwise is how it walks into a gap.
func (a *Actor) reach(airTicks float32) float32 {
	return a.reachAt(abs(a.Speed.X), airTicks)
}

// reachAt is the same distance at a speed the bot does not have yet.
func (a *Actor) reachAt(speed, airTicks float32) float32 {
	return speed / 60 * airTicks * JumpMargin
}

// takeoff is how fast the bot will be going after run pixels of run-up: what it
// has now plus what the acceleration adds, and never more than it can go.
//
// A gap has to be judged at the speed the bot will leave it at, not the speed
// it is going while it decides. Judging by the speed it has now, a bot standing
// a stride from the edge finds it cannot cross, stops - and so never builds the
// speed that would have carried it over. It would stand there for ever.
func (a *Actor) takeoff(run float32) float32 {
	v := abs(a.Speed.X)
	return min(sqrtf(v*v+2*a.Acc*max(run, 0)), MaxSpeed)
}

// front is the leading edge of the bot's collision box going dir - the face
// that meets whatever is in its way.
//
// The probes take the direction rather than reading Orientation, because the
// bot decides which way to go before it has turned that way: a probe reading
// the facing would answer for the way it came from.
func (a *Actor) front(dir int8) float32 {
	l := a.Pos.X + a.CollisionBox.X
	if dir < 0 {
		return l
	}
	return l + a.CollisionBox.Width
}

// support is the column a fighter stands or falls by: the middle of its
// collision box, which is the one checkCollision looks for ground under.
func (a *Actor) support() float32 {
	return a.Pos.X + a.CollisionBox.X + a.CollisionBox.Width/2
}

// feet and head are the bottom and top of the collision box.
func (a *Actor) feet() float32 { return a.Pos.Y + a.CollisionBox.Y + a.CollisionBox.Height }
func (a *Actor) head() float32 { return a.Pos.Y + a.CollisionBox.Y }

// wallAhead returns how high the bot must get and how far ahead the wall is:
// the nearest solid standing in its way within reach, and the top of the column
// that solid belongs to.
//
// What counts as a wall rather than a step is Actor.isWall, the same question
// blockSides asks - a bot's idea of what stops it and the collision code's
// cannot be allowed to drift apart. But isWall only knows about what is level
// with the body, and the height to get to is the top of the whole wall, so from
// the rectangle that blocks the bot the column is walked upward through
// whatever is stacked on it.
func (a *Actor) wallAhead(dir int8, reach float32) (top, dist float32, ok bool) {
	step := float32(dir)
	// How far ahead the bot would meet c, or -1 for one that is not in its way.
	face := func(c rl.Rectangle) float32 {
		if dir < 0 {
			return c.X + c.Width
		}
		return c.X
	}
	ahead := func(c rl.Rectangle) float32 { return (face(c) - a.front(dir)) * step }
	dist = reach + 1
	for _, c := range LE.solids {
		if !a.isWall(c) {
			continue
		}
		if d := ahead(c); d >= 0 && d < dist {
			dist = d
		}
	}
	if dist > reach {
		return 0, 0, false
	}
	// the lowest rectangle of the column that is in the way, then everything
	// resting on it
	top = a.feet()
	for _, c := range LE.solids {
		if !a.isWall(c) || abs(ahead(c)-dist) > 0.5 {
			continue
		}
		top = min(top, c.Y)
	}
	for again := true; again; {
		again = false
		for _, c := range LE.solids {
			if abs(ahead(c)-dist) > 0.5 {
				continue
			}
			if c.Y < top && c.Y+c.Height >= top-1 {
				top, again = c.Y, true
			}
		}
	}
	return top, dist, true
}

// groundAt returns the top of the highest ground under the column at x that the
// bot could stand on: no higher above its feet than a step it can walk up, and
// no further below than drop. Nothing within that is a hole.
//
// The step's worth of headroom matters: without it a rise of a few pixels reads
// as a hole, and the bot stops dead at something it would have walked over.
//
// The column test takes the left edge and not the right, the way the ground
// test in checkCollision does, so a column landing exactly on the seam between
// two tiles belongs to one of them rather than falling between the two.
func groundAt(x, feetY, drop float32) (float32, bool) {
	top := float32(0)
	found := false
	for _, c := range LE.solids {
		if x < c.X || x >= c.X+c.Width {
			continue
		}
		if c.Y < feetY-StepHeight || c.Y > feetY+drop {
			continue
		}
		if !found || c.Y < top {
			top, found = c.Y, true
		}
	}
	return top, found
}

// solidAt reports whether the column at x has anything at all in it that stops
// the bot dropping out of the world: ground it can stand on, or a rise it
// cannot, which is a wall and somebody else's problem. Only an empty column is
// a hole.
//
// This is the question edgeAhead asks, and it is not the one groundAt answers.
// A step up is not ground the bot can stand on from here, but a bot that read
// it as a hole would brake at the foot of a staircase - and, in the air with no
// wall check in front of it, brake backwards off the ledge it just left.
func solidAt(x, feetY, drop float32) bool {
	for _, c := range LE.solids {
		if x < c.X || x >= c.X+c.Width {
			continue
		}
		if c.Y <= feetY+drop {
			return true
		}
	}
	return false
}

// edgeAhead returns the last column ahead of the bot that still has ground
// under it before the ground runs out, and whether it runs out at all within
// lookahead. A bot standing over nothing reports no edge: it is already falling
// and the question does not arise.
//
// Measured from where the fighter is actually held up, not from the leading
// face of its collision box. A fighter stands or falls by the middle of that
// box - it is the one column checkCollision looks under - and the face is
// half a body ahead of it. Probing from the face, a bot walking at a ledge
// lost the hole the moment the face crossed it, which is several pixels before
// it would fall, and walked those last pixels with nothing telling it to stop.
func (a *Actor) edgeAhead(dir int8, lookahead, drop float32) (float32, bool) {
	from := a.support()
	last := from
	if !solidAt(from, a.feet(), drop) {
		return 0, false
	}
	for d := float32(probeStep); d <= lookahead; d += probeStep {
		x := from + d*float32(dir)
		if !solidAt(x, a.feet(), drop) {
			return last, true
		}
		last = x
	}
	return 0, false
}

// landingBeyond returns the first column past a gap that has ground the bot
// could land on, searching out to reach from the edge it would leave, and
// whether there is one.
func (a *Actor) landingBeyond(dir int8, edgeX, reach, drop float32) (float32, bool) {
	for d := float32(probeStep); d <= reach; d += probeStep {
		x := edgeX + d*float32(dir)
		if _, ok := groundAt(x, a.feet(), drop); ok {
			return x, true
		}
	}
	return 0, false
}

// botMind is everything a bot carries from one tick to the next: where it last
// saw the fighter it was after, and how long it will go on looking there.
//
// Two states and no enum for them, because there are only two: the bot has
// somewhere to be, or it has not. searching says which.
type botMind struct {
	lastSeen  rl.Vector2
	hasLast   bool // something was seen at lastSeen; the zero vector is a real position
	searching bool
	patience  uint16 // ticks left to reach the spot, then to stand and look
}

// forget puts the bot back to having nothing in mind.
func (m *botMind) forget() { *m = botMind{} }

// mark notes where a fighter is being seen right now. Nothing is being searched
// for while something is in view.
func (m *botMind) mark(pos rl.Vector2) {
	m.lastSeen, m.hasLast = pos, true
	m.searching, m.patience = false, 0
}

// remember starts the search for the spot last marked, replacing whatever was
// being searched for before: a bot looks for one thing at a time.
func (m *botMind) remember() {
	m.searching, m.patience = true, SightSearchTicks
}

// eye is the point a fighter sees from and is seen at: the centre of its
// collision box. Pos is the sprite's top-left, so a line drawn from there
// between two fighters on the same floor runs through their foreheads, and a
// waist-high ledge standing between them would not break it.
func (a *Actor) eye() rl.Vector2 {
	return rl.Vector2{
		X: a.Pos.X + a.CollisionBox.X + a.CollisionBox.Width/2,
		Y: a.Pos.Y + a.CollisionBox.Y + a.CollisionBox.Height/2,
	}
}

// clipSlab narrows the segment's parameter range to the stretch of it inside
// one slab - the band between lo and hi on a single axis - and reports whether
// any of the segment is left. A segment that does not move on that axis is
// either inside the band for its whole length or outside it for all of it,
// which is what stands in for the division that cannot be made.
func clipSlab(from, d, lo, hi float32, t0, t1 *float32) bool {
	if d == 0 {
		return from >= lo && from <= hi
	}
	a, b := (lo-from)/d, (hi-from)/d
	if a > b {
		a, b = b, a
	}
	*t0 = max(*t0, a)
	*t1 = min(*t1, b)
	return *t0 <= *t1
}

// lineBlocked reports whether any solid part of the level lies on the segment
// from p to q: the slab method, clipping the segment against each rectangle's
// X band and then its Y band, anything surviving both meaning the segment is
// inside that rectangle for some stretch of its length.
//
// It walks LE.solids the way collision does. A thin rectangle is no harder for
// it than a thick one, which is why it clips rather than marching the line in
// steps - the level is full of collision rectangles a step would jump over.
func lineBlocked(p, q rl.Vector2) bool {
	dx, dy := q.X-p.X, q.Y-p.Y
	for _, c := range LE.solids {
		t0, t1 := float32(0), float32(1)
		if !clipSlab(p.X, dx, c.X, c.X+c.Width, &t0, &t1) {
			continue
		}
		if !clipSlab(p.Y, dy, c.Y, c.Y+c.Height, &t0, &t1) {
			continue
		}
		return true
	}
	return false
}

// canSee reports whether a has b in view: b is on the side a faces, within
// range, inside the cone either side of that facing, and with nothing solid on
// the line between them. Every test is cheaper than the one after it, so the
// sweep of the level is only reached by a fighter that is worth sweeping for.
//
// Nothing within a body's width is behind the bot, whichever way it faces. Two
// fighters at that distance are touching - one is standing on the other - and a
// bot that called that "behind me" would be blind to the fighter it is trading
// blows with every time the two of them crossed over.
func (a *Actor) canSee(b *Actor) bool {
	if b == nil || b == a {
		return false
	}
	from, to := a.eye(), b.eye()
	dx, dy := to.X-from.X, to.Y-from.Y
	if dx*float32(a.Orientation) < 0 && abs(dx) > a.CollisionBox.Width {
		return false
	}
	if dx*dx+dy*dy > SightRange*SightRange {
		return false
	}
	if abs(dy) > abs(dx)*SightSlope {
		return false
	}
	return !lineBlocked(from, to)
}

// seen returns the nearest fighter a can see, or nil when it can see none.
func (a *Actor) seen() *Actor {
	var best *Actor
	bestD := float32(math.MaxFloat32)
	from := a.eye()
	for _, b := range a.W.Actors {
		if !a.canSee(b) {
			continue
		}
		to := b.eye()
		dx, dy := to.X-from.X, to.Y-from.Y
		if d := dx*dx + dy*dy; d < bestD {
			best, bestD = b, d
		}
	}
	return best
}

// How the bot reads the ground it is walking on. Feel numbers, like the sight
// ones, and stated where the probes that use them can be read beside them.
const (
	// Near enough to the goal in X to count as arrived, and to stop shuffling.
	botArrived = 8
	// Close enough to a wall to jump from a standstill and still catch it.
	botAdjacent = 12
	// How far short of a hole it has no way across the bot comes to rest. It is
	// a fixed distance and not a braking distance: a threshold that moved with
	// the speed has an equilibrium in it, and the bot found it - pressing and
	// releasing on alternate ticks, going nowhere, because at that speed the
	// step it takes is smaller than the pixel the position is truncated to.
	botStandoff = 32
	// A drop the bot is willing to take: ground this far below the ledge it is
	// leaving is somewhere to go, not a hole. Two tiles.
	botDrop = 96
	// How far ahead it reads the ground for holes. Far enough to commit a
	// run-jump before the edge arrives, whatever the speed.
	botLook = 96
	// How much room a bot wants behind a gap it means to jump: the ground it
	// takes to get up to a run from a standstill, plus the ground the run-jump
	// spends winding up. Derived, so retuning the acceleration or the animation
	// carries it along.
	//
	// A gap is crossed by the running jump or not at all. The standing jump is
	// no use for one: it winds up for the best part of a second, on its feet,
	// and a fighter walking at an edge for that long is over it before the jump
	// ever leaves the ground.
	botRunUp = RunReach*RunReach/(2*DefaultAcc) + RunReach*RunJumpWindupTicks/60
)

// clears reports whether a jump that rises rise pixels gets the bot onto a wall
// whose top stands height above its feet - either over it outright, or high
// enough for the climb to catch the ledge, which needs the head at the top of
// the arc to reach the wall's top.
//
// The climb is what does most of the work: the head reaching a ledge is worth
// the fighter's whole height over the bare rise. This is the conservative half
// of the engine's rule - a ledge the head stops just short of is often caught
// anyway, because the grab only needs the head somewhere inside the rectangle -
// so a bot refuses a wall now and then that it could in fact have climbed.
func (a *Actor) clears(rise, height float32) bool {
	return height <= rise+a.CollisionBox.Height
}

// running reports whether a jump asked for now would be the running one.
func (a *Actor) running() bool { return a.State == StateRun }

// brake is how far the bot still travels after it stops pressing forward, with
// friction alone bringing it to rest: v^2/2f, the same shape as a jump's rise.
//
// A bot at a run carries a long way - from top speed, more than twice the
// distance it reads ahead of itself - so a hole has to be seen this far off or
// the bot slides into it having decided not to.
func (a *Actor) brake() float32 {
	return a.Speed.X * a.Speed.X / (2 * a.Friction)
}

// overWall is where a fighter would stand on top of the wall the bot is up
// against, and whether there is one it could get onto. It is what the bot tries
// when a search runs out of ground: the answer to "it was here a moment ago and
// now it is not" is usually that it went up.
func (a *Actor) overWall() (rl.Vector2, bool) {
	dir := a.Orientation
	top, dist, ok := a.wallAhead(dir, botLook)
	if !ok || dist > botAdjacent || !a.clears(RunJumpRise, a.feet()-top) {
		return rl.Vector2{}, false
	}
	// past the face of the wall, standing on it
	x := a.front(dir) + dist*float32(dir)
	if dir > 0 {
		x += 4
	} else {
		x -= 32
	}
	return rl.Vector2{X: x, Y: top - 48}, true
}

// holdShortOf is how a bot stops at a hole it has no way across, gap pixels
// ahead of it.
//
// Not pressing forward is usually enough: friction takes back very nearly what
// the acceleration gave, so the bot coasts to rest a stride short of the edge.
// Where there is less room than it takes to coast, coasting ends up over the
// edge, and the only thing left is to turn and push the other way - which
// brakes far harder, a fighter that turns losing two thirds of its speed on the
// spot. That costs the bot sight of whatever it was after, which is the cheaper
// of the two prices.
//
// The standoff is a fixed distance and not a braking distance: a threshold that
// moves with the speed has an equilibrium in it, and the bot found it, pressing
// and releasing on alternate ticks while going nowhere, because at that speed
// the step it takes is smaller than the pixel the position is truncated to.
func (a *Actor) holdShortOf(dir int8, gap float32) step {
	switch {
	case gap < a.brake():
		return step{dir: -dir}
	case gap <= max(a.brake()+botAdjacent, botStandoff):
		return step{stop: true}
	}
	return step{dir: dir}
}

// brakeForHole reports whether the ground ahead runs out with nothing beyond it
// the bot could land on, and what to do about it. It is the danger half of the
// hole reading with none of the jumping: what a bot still in the air needs,
// where the jump is already spent and only the landing is still in question.
func (a *Actor) brakeForHole(dir int8) (step, bool) {
	edge, ok := a.edgeAhead(dir, botLook+a.brake(), botDrop)
	if !ok {
		return step{}, false
	}
	// A gap it could cross is no danger: mid-jump over one, the far side is
	// where it means to be.
	if _, across := a.landingBeyond(dir, edge, a.reachAt(MaxSpeed, RunJumpAirTicks), botDrop); across {
		return step{}, false
	}
	s := a.holdShortOf(dir, (edge-a.support())*float32(dir))
	return s, s.dir != dir
}

// downLedge is where a fighter would stand on the ground below the ledge the
// bot is on, and whether there is one close enough ahead to drop to. It is the
// other half of overWall: a fighter that goes down is lost the same way one
// that goes up is, at the moment the ledge comes between them, and leaves the
// same useless spot behind - the ground under the bot's own feet.
func (a *Actor) downLedge() (rl.Vector2, bool) {
	dir := a.Orientation
	from := a.support()
	for d := float32(probeStep); d <= botLook; d += probeStep {
		x := from + d*float32(dir)
		top, ok := groundAt(x, a.feet(), botDrop)
		if !ok {
			return rl.Vector2{}, false // a hole it cannot take, not a step down
		}
		if top > a.feet()+StepHeight {
			return rl.Vector2{X: x - a.CollisionBox.X - a.CollisionBox.Width/2, Y: top - 48}, true
		}
	}
	return rl.Vector2{}, false
}

// roomBehind reports whether the ground behind the bot runs on far enough to
// take a run-up on. Backing off into a second hole would trade one gap it
// cannot cross for one it falls into.
func (a *Actor) roomBehind(dir int8) bool {
	_, hole := a.edgeAhead(-dir, botRunUp, botDrop)
	return !hole
}

// step is what a bot means to do with its feet this tick: which way to walk,
// whether to leave the ground, and whether there is simply nothing to be done -
// no way from here to where it was going, so come to rest rather than press
// into the level.
type step struct {
	dir  int8
	jump bool
	stop bool
}

// botStep answers the one question the bot asks about the level: where do I put
// my feet. It reads only what is immediately in front of and under the bot, so
// it cannot plan a route - but it is the whole of the decision, and a planner
// would replace it by handing back a waypoint instead of the goal itself,
// without botInput knowing the difference.
func botStep(a *Actor, goal rl.Vector2) step {
	dir := int8(0)
	switch {
	case goal.X > a.Pos.X+botArrived:
		dir = 1
	case goal.X < a.Pos.X-botArrived:
		dir = -1
	default:
		return step{stop: true} // already there
	}
	if !a.OnGround {
		// The arc is set, but the ground it comes down on is not read yet. A
		// fighter that lands at a run with a hole in front of it has no room
		// left to stop, so the ground ahead is looked at from up here too -
		// only to brake, never to jump, the jump being already spent.
		if s, danger := a.brakeForHole(dir); danger {
			return s
		}
		return step{dir: dir} // otherwise steer, and let the arc finish
	}

	// A wall is dealt with before a hole: it is the nearer of the two whenever
	// both are ahead, and the hole beyond it is not this bot's problem yet.
	if top, dist, ok := a.wallAhead(dir, botLook); ok {
		height := a.feet() - top
		switch {
		case a.running() && a.canRunJump() && a.clears(RunJumpRise, height):
			// run at it, and press early enough that the push-off frame comes
			// before the wall does
			if dist <= a.reach(RunJumpWindupTicks)+probeStep {
				return step{dir: dir, jump: true}
			}
			return step{dir: dir}
		case a.clears(JumpRise, height):
			// the standing jump clears it, and it is the standing jump that
			// has to be used: a run-jump asked for with nothing in the pool is
			// refused outright, so the bot stops first and jumps from its feet
			if a.running() {
				return step{}
			}
			if dist <= botAdjacent {
				return step{dir: dir, jump: true}
			}
			return step{dir: dir}
		default:
			return step{stop: true} // too tall to get onto
		}
	}

	// A hole in the ground: ground below within a drop it will take is not a
	// hole at all, which is how following a target down a ledge comes for free.
	if edge, ok := a.edgeAhead(dir, botLook+a.brake(), botDrop); ok {
		gap := (edge - a.support()) * float32(dir)
		// Only the running jump crosses a gap, so what matters is the speed the
		// bot will be doing at the edge, not the speed it is doing while it
		// decides - and whether that is a run at all.
		speed := a.takeoff(gap)
		reaches := func(v float32) bool {
			_, ok := a.landingBeyond(dir, edge, a.reachAt(v, RunJumpAirTicks), botDrop)
			return ok
		}
		switch {
		case a.canRunJump() && speed >= RunReach && reaches(speed):
			// Press early enough that the push-off frame comes before the edge.
			if gap <= a.reach(RunJumpWindupTicks)+probeStep {
				return step{dir: dir, jump: true}
			}
			return step{dir: dir}

		case a.canRunJump() && reaches(MaxSpeed) && gap < botRunUp && a.roomBehind(dir):
			// It could clear this at a run and cannot from here: there is not
			// enough ground left to get up to one. So back away from the edge
			// far enough to take a run at it. Backing off turns the bot and
			// costs it sight of whatever it was after, but the spot it last saw
			// that thing at is still across the gap, so it turns round and comes
			// back the moment it has the room.
			return step{dir: -dir}

		default:
			// Nothing to land on, or no way to get up the speed.
			return a.holdShortOf(dir, gap)
		}
	}

	return step{dir: dir}
}

// Colours of the sight overlay, dim enough to read the level through.
var (
	sightConeColor   = rl.Color{R: 120, G: 200, B: 255, A: 40}  // what it could see
	sightBlindColor  = rl.Color{R: 120, G: 200, B: 255, A: 90}  // the edge of the cone
	sightTargetColor = rl.Color{R: 120, G: 255, B: 140, A: 220} // the line to what it sees
	sightMemoryColor = rl.Color{R: 255, G: 200, B: 90, A: 200}  // the spot it is searching
)

// drawSight draws what a bot can see, for debug mode: the cone it looks down,
// a line to the fighter it has picked out of it, and the spot it is walking to
// when it has lost one.
//
// Only where the sight is real. A bot on a client never runs any of this - the
// server does, and the client is only told where the bot went - so drawing a
// cone there would be drawing a guess.
func (a *Actor) drawSight() {
	if a.IsPlayer || a.W == nil || !a.W.Authority || a.State == StateDeath {
		return
	}
	eye := a.eye()
	half := float32(math.Atan(SightSlope) * 180 / math.Pi)
	mid := float32(0) // raylib measures from +X, clockwise, with Y down
	if a.Orientation < 0 {
		mid = 180
	}
	rl.DrawCircleSector(eye, SightRange, mid-half, mid+half, 24, sightConeColor)
	rl.DrawCircleSectorLines(eye, SightRange, mid-half, mid+half, 24, sightBlindColor)

	if t := a.seen(); t != nil {
		rl.DrawLineV(eye, t.eye(), sightTargetColor)
		rl.DrawCircleLinesV(t.eye(), 4, sightTargetColor)
		return
	}
	if a.Bot.searching {
		spot := rl.Vector2{X: a.Bot.lastSeen.X + 21, Y: a.Bot.lastSeen.Y + 27} // its eye, had it stood there
		rl.DrawLineV(eye, spot, sightMemoryColor)
		rl.DrawCircleLinesV(spot, 3, sightMemoryColor)
		rl.DrawLineV(rl.Vector2{X: spot.X - 4, Y: spot.Y - 4}, rl.Vector2{X: spot.X + 4, Y: spot.Y + 4}, sightMemoryColor)
		rl.DrawLineV(rl.Vector2{X: spot.X - 4, Y: spot.Y + 4}, rl.Vector2{X: spot.X + 4, Y: spot.Y - 4}, sightMemoryColor)
	}
}

// alerted tells a bot where an attack came from. A blow is the one thing that
// reaches a fighter from outside its sight, and a bot that took no notice of one
// could be hit from behind all day - so the attacker's position becomes the
// place the bot last saw someone, which turns it round and sends it to look. If
// the attacker is still standing there when it arrives, it sees them and the
// fight goes on as any other.
//
// It learns a position and nothing more. The attacker is not made visible by
// this: an attacker that strikes and steps aside leaves the bot searching the
// place the blow came from, rather than tracking someone it cannot see. What
// the bot can see outranks it either way, because sight marks the same memory
// every tick it has anything in view.
//
// A human is never steered by it.
func (a *Actor) alerted(attacker *Actor) {
	if a.IsPlayer || attacker == nil || attacker == a {
		return
	}
	a.Bot.mark(attacker.Pos)
}

// botInput produces the bot's command for this tick (authoritative worlds only).
//
// See, then decide, then step. A bot acts on the nearest fighter it can
// actually see; failing that, on the spot it last saw one at; and failing that,
// on nothing at all - it stands where it is, still drifting its guard, and it
// does not turn. Turning toward something it cannot see would put that thing in
// view on the next tick and the cone would mean nothing.
//
// A broken fighter does not think at all: it holds no key, looks at nothing and
// remembers nothing while the fall plays out. Actor.Update drops a dead
// fighter's command and face refuses a dead turn, so this is not what keeps the
// body still - it is that a corpse has no business seeing.
func botInput(a *Actor) Input {
	if a.State == StateDeath {
		return Input{}
	}
	if a.State == StateAttacked {
		return a.Input
	}
	in := Input{Updown: a.Updown}
	if rand.Int()%61 == 0 {
		in.Updown = uint8(rand.Int() % 3)
	}

	if t := a.seen(); t != nil {
		// Seeing anything at all ends a search before its counter is ever
		// read: what is there now beats where something used to be.
		a.Bot.mark(t.Pos)
		// Close enough to swing at: stop walking into it and start throwing.
		// Close is both ways about: a fighter standing on a ledge overhead is
		// within arm's reach across the floor and nowhere near the end of a
		// swing, and a bot that only measured the one distance would stand
		// under it throwing at the air instead of climbing up to it.
		const near = 50
		if abs(t.Pos.X-a.Pos.X) <= near && abs(t.Pos.Y-a.Pos.Y) <= a.CollisionBox.Height {
			a.face(facing(t.Pos.X - a.Pos.X))
			if rand.Int()%61 == 3 {
				in.Attack = true
			}
			return in
		}
		return a.walk(in, t.Pos)
	}

	// Nothing in view. If something was a moment ago, go and look where it was.
	if !a.Bot.searching {
		if !a.Bot.hasLast {
			return in // nothing was ever seen: stand
		}
		a.Bot.remember()
	}
	a.Bot.patience--
	if a.Bot.patience == 0 {
		a.Bot.forget()
		return in
	}
	s := botStep(a, a.Bot.lastSeen)
	if s.stop {
		// Look that way even with nowhere to walk: a blow landed from right
		// behind leaves the spot too close to step toward, and a bot that
		// would not turn for it could be hit from behind all day.
		if dx := a.Bot.lastSeen.X - a.Pos.X; abs(dx) > 1 {
			a.face(facing(dx))
		}
		// Nothing here, and a wall right in front: try the top of it before
		// giving up. A fighter that hops a ledge is lost at the moment the wall
		// comes between the two of them, which leaves the spot it was last seen
		// at on the ground the bot is already standing on - so the search ends
		// the tick it starts, with the quarry standing directly overhead and a
		// wall the bot is pressed against hiding it for good.
		if spot, ok := a.overWall(); ok {
			a.Bot.lastSeen = spot
			return a.walkStep(in, botStep(a, spot))
		}
		// And the same the other way: over the edge it went down.
		if spot, ok := a.downLedge(); ok {
			a.Bot.lastSeen = spot
			return a.walkStep(in, botStep(a, spot))
		}
		// Arrived, or no way to get there. Either way the walking is over:
		// what is left is a moment standing and looking around, and the
		// search is not the sort of thing a bot swings during.
		if a.Bot.patience > SightLingerTicks {
			a.Bot.patience = SightLingerTicks
		}
		return in
	}
	return a.walkStep(in, s)
}

// walk turns a step into the command that takes it, and turns the bot the way
// it is going. Left and Right are the keys' names, not the level's directions:
// Right is the key that moves a fighter toward smaller X. See sampleInput.
func (a *Actor) walk(in Input, goal rl.Vector2) Input {
	return a.walkStep(in, botStep(a, goal))
}

// walkStep is walk once the step has already been decided.
func (a *Actor) walkStep(in Input, s step) Input {
	if s.dir != 0 {
		a.face(s.dir)
		in.Left = s.dir > 0
		in.Right = s.dir < 0
	}
	in.Jump = s.jump
	return in
}

// facing is the way to turn to look at something dx away.
func facing(dx float32) int8 {
	if dx < 0 {
		return -1
	}
	return 1
}
