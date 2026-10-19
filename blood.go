package resistor

import (
	"math"
	"math/rand"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// Blood is what a landed blow leaves behind: droplets thrown away from it that
// fly, land on the level and stay there as small clumps of stain. It is the view's
// alone - every machine makes its own from the structure drops it is shown,
// nothing of it is sent, and nothing in the match reads it.
const (
	BloodFlinch = 10 // droplets a flinching hit sprays
	BloodStun   = 30
	BloodKill   = 60

	BloodSpeedMin = 40.0 // px/s at launch
	BloodSpeedMax = 140.0
	BloodGravity  = 400.0 // px/s^2
	BloodLife     = 2.0   // seconds a droplet may fly before it is dropped

	BloodDroplets = 512  // in flight at once; past it the oldest goes
	BloodStains   = 4096 // stuck to the level; past it the oldest goes

	// A droplet flies between these many seconds before it splashes onto the
	// wall behind the fight, if there is painted wall under it; the spread is
	// what scatters a spray into a splatter rather than a ring.
	BloodReachMin = 0.04
	BloodReachMax = 0.3
	// BloodStreakMax is how many pixels a streak's tail runs back along the
	// droplet's flight behind its head.
	BloodStreakMax = 4

	// A droplet is a clump of 1 to BloodDropMax pixels, and splashes out by 1
	// to BloodSplatMax more when it lands.
	BloodDropMax  = 3
	BloodSplatMax = 4

	// Blood on a wall runs down it up to BloodRunMax pixels, slowing as it goes,
	// for BloodRunTime seconds after it lands.
	BloodRunMax  = 12
	BloodRunTime = 3.0
)

// A blot is the shape of a clump of blood: which pixels of the 5x5 square
// around its center it covers, bit (dy+2)*5 + (dx+2). The center is always in.
type blot uint32

const (
	blotR      = 2
	blotSide   = 2*blotR + 1
	blotCenter = blot(1) << (blotR*blotSide + blotR)
)

func blotBit(dx, dy int) blot {
	return blot(1) << ((dy+blotR)*blotSide + dx + blotR)
}

// has reports whether the blot covers the pixel dx, dy from its center.
func (bl blot) has(dx, dy int) bool {
	return dx >= -blotR && dx <= blotR && dy >= -blotR && dy <= blotR && bl&blotBit(dx, dy) != 0
}

// pixels calls f with every pixel the blot covers, as offsets from its center.
func (bl blot) pixels(f func(dx, dy int)) {
	for dy := -blotR; dy <= blotR; dy++ {
		for dx := -blotR; dx <= blotR; dx++ {
			if bl.has(dx, dy) {
				f(dx, dy)
			}
		}
	}
}

// grow spreads a blot by up to n more pixels, each one next to a pixel it
// already covers, so a clump stays in one piece.
func (b *bloodState) grow(bl blot, n int) blot {
	bl |= blotCenter
	for ; n > 0; n-- {
		// the pixels next to the clump, not yet in it
		var edge [blotSide * blotSide][2]int
		k := 0
		for dy := -blotR; dy <= blotR; dy++ {
			for dx := -blotR; dx <= blotR; dx++ {
				if !bl.has(dx, dy) && (bl.has(dx+1, dy) || bl.has(dx-1, dy) || bl.has(dx, dy+1) || bl.has(dx, dy-1)) {
					edge[k] = [2]int{dx, dy}
					k++
				}
			}
		}
		if k == 0 {
			break // the blot is full
		}
		p := edge[b.rng.Intn(k)]
		bl |= blotBit(p[0], p[1])
	}
	return bl
}

// keep is the part of a blot centered on x, y whose pixels pass ok. The center
// is kept whatever ok says: it is where the droplet came to rest.
func (bl blot) keep(x, y int, ok func(x, y int) bool) blot {
	out := blotCenter
	bl.pixels(func(dx, dy int) {
		if ok(x+dx, y+dy) {
			out |= blotBit(dx, dy)
		}
	})
	return out
}

// bloodReds are the level palette's reds, darkest first.
var bloodReds = [...]rl.Color{
	rl.NewColor(87, 28, 39, 255),
	rl.NewColor(137, 30, 43, 255),
	rl.NewColor(196, 36, 48, 255),
	rl.NewColor(234, 50, 60, 255),
}

type droplet struct {
	pos, vel rl.Vector2
	age      float32
	reach    float32 // seconds before it may splash onto the backdrop
	red      uint8
	shape    blot // the clump it flies as
}

// A stain keeps the world tick it landed on. Time manipulation shows only the
// stains a frame had, and resuming from a frame deletes the ones after it; so
// a stain never moves or changes once stuck, and the tick is all a rewind
// needs of it.
//
// A stain on the backdrop is a streak: tx, ty is its tail, back along the way
// the droplet flew. A stain on a floor or a wall it hit has no tail. Either is
// a clump at its head, its shape the droplet's splashed out where it landed.
// One on a wall runs run pixels down it; how much of that has run by now is
// the ticks since it landed, so the stain itself still never changes.
type stain struct {
	x, y   int16
	tx, ty int8
	red    uint8
	run    uint8
	tick   uint32
	shape  blot
}

// seenFighter is what the blood last saw of a fighter: a drop in structure
// between two looks is a blow that landed.
type seenFighter struct {
	structure float32
	state     ActorState
	updown    uint8
	blocks    uint8
	frame     uint8 // the animation frame, for the clash's impacts
}

type bloodState struct {
	drops []droplet
	// the stains are a ring: the oldest at start, n of them in landing order
	stains  [BloodStains]stain
	start   int
	n       int
	seen    map[*Actor]seenFighter
	history bool // the last frame showed a recorded frame
	// its own random source, so spraying blood can never change what the bots,
	// which draw from the global one, decide
	rng *rand.Rand
}

var blood = newBlood()

func newBlood() *bloodState {
	return &bloodState{seen: map[*Actor]seenFighter{}, rng: rand.New(rand.NewSource(1))}
}

// clearBlood wipes the level clean, for a new match or a restart.
func clearBlood() {
	blood = newBlood()
}

// stainAt returns the i-th stain in landing order.
func (b *bloodState) stainAt(i int) *stain {
	return &b.stains[(b.start+i)%BloodStains]
}

func (b *bloodState) addStain(s stain) {
	if b.n < BloodStains {
		*b.stainAt(b.n) = s
		b.n++
		return
	}
	b.stains[b.start] = s
	b.start = (b.start + 1) % BloodStains
}

// dropStainsAfter deletes the stains that landed after tick. They are in
// landing order, so the newest end is enough.
func (b *bloodState) dropStainsAfter(tick uint32) {
	for b.n > 0 && b.stainAt(b.n-1).tick > tick {
		b.n--
	}
}

// spray throws n droplets from at, mostly toward dir (-1 or +1) and upward,
// some straight up, a few back the other way.
func (b *bloodState) spray(at rl.Vector2, dir float32, n int) {
	for range n {
		angle := (-20 + b.rng.Float64()*90) * math.Pi / 180 // from a little under level to steeply up
		speed := BloodSpeedMin + b.rng.Float64()*(BloodSpeedMax-BloodSpeedMin)
		vx := float32(math.Cos(angle)*speed) * dir
		if b.rng.Intn(7) == 0 {
			vx *= -0.4 // a few fly back past the blow
		}
		d := droplet{
			pos:   at,
			vel:   rl.Vector2{X: vx, Y: -float32(math.Sin(angle) * speed)},
			reach: float32(BloodReachMin + b.rng.Float64()*(BloodReachMax-BloodReachMin)),
			red:   uint8(b.rng.Intn(len(bloodReds))),
		}
		d.shape = b.grow(0, b.rng.Intn(BloodDropMax))
		if len(b.drops) >= BloodDroplets {
			copy(b.drops, b.drops[1:])
			b.drops = b.drops[:len(b.drops)-1]
		}
		b.drops = append(b.drops, d)
	}
}

// pixelSolid reports whether the level pixel at p is inside the level's collision.
func pixelSolid(solids []rl.Rectangle, p rl.Vector2) bool {
	for _, r := range solids {
		if p.X >= r.X && p.X < r.X+r.Width && p.Y >= r.Y && p.Y < r.Y+r.Height {
			return true
		}
	}
	return false
}

// A mask is which pixels of a texture are painted: alpha over half.
type mask struct {
	w, h int
	bits []uint64
}

func (m *mask) at(x, y int) bool {
	if x < 0 || y < 0 || x >= m.w || y >= m.h {
		return false
	}
	i := y*m.w + x
	return m.bits[i/64]&(1<<(i%64)) != 0
}

// texMasks holds a mask for every texture loaded, by resource name, so blood
// can ask whether a pixel of a tile is painted without reading the GPU.
var texMasks = map[string]*mask{}

// maskOf reads which pixels of an image are painted.
func maskOf(img *rl.Image) *mask {
	w, h := int(img.Width), int(img.Height)
	m := &mask{w: w, h: h, bits: make([]uint64, (w*h+63)/64)}
	for i, c := range rl.LoadImageColors(img) {
		if c.A > 127 {
			m.bits[i/64] |= 1 << (i % 64)
		}
	}
	return m
}

// paintedAt reports whether the level pixel at p is painted on any tile: where
// a droplet may splash onto the wall behind the fight.
func paintedAt(tiles []LevelObject, p rl.Vector2) bool {
	x, y := int(math.Floor(float64(p.X))), int(math.Floor(float64(p.Y)))
	for _, t := range tiles {
		if m := texMasks[t.Resource]; m != nil && m.at(x-int(t.X), y-int(t.Y)) {
			return true
		}
	}
	return false
}

// streak is the stain a droplet leaves splashing onto the backdrop at p: its
// head there, its tail up to BloodStreakMax pixels back along its flight. The
// head's clump covers only painted pixels, so no blood hangs off a tile's edge.
func (b *bloodState) streak(d droplet, p rl.Vector2, solids []rl.Rectangle, tiles []LevelObject, tick uint32) stain {
	x, y := int(math.Floor(float64(p.X))), int(math.Floor(float64(p.Y)))
	s := stain{x: int16(x), y: int16(y), red: d.red, tick: tick}
	s.shape = b.splat(d).keep(x, y, func(x, y int) bool {
		return paintedAt(tiles, rl.Vector2{X: float32(x), Y: float32(y)})
	})
	if speed := math.Hypot(float64(d.vel.X), float64(d.vel.Y)); speed > 0 {
		l := float64(1 + b.rng.Intn(BloodStreakMax))
		s.tx = int8(math.Round(-float64(d.vel.X) / speed * l))
		s.ty = int8(math.Round(-float64(d.vel.Y) / speed * l))
	}
	s.run = b.runDown(x, y, func(x, y int) bool {
		p := rl.Vector2{X: float32(x) + 0.5, Y: float32(y) + 0.5}
		return paintedAt(tiles, p) && !pixelSolid(solids, p)
	})
	return s
}

// runDown is how far a stain at x, y runs: a random length, cut short where
// the pixel below is no longer wall it can run on.
func (b *bloodState) runDown(x, y int, onWall func(x, y int) bool) uint8 {
	n := b.rng.Intn(BloodRunMax + 1)
	for k := 1; k <= n; k++ {
		if !onWall(x, y+k) {
			return uint8(k - 1)
		}
	}
	return uint8(n)
}

// splat is the clump a droplet leaves where it lands: its own, splashed out.
func (b *bloodState) splat(d droplet) blot {
	return b.grow(d.shape, 1+b.rng.Intn(BloodSplatMax))
}

// hug is the stain a droplet leaves stopped by the level's collision at p: its
// clump covers only the collision and the pixels touching it, smeared along
// the surface rather than hanging in the air beside it.
func (b *bloodState) hug(d droplet, p rl.Vector2, solids []rl.Rectangle, tick uint32) stain {
	x, y := int(math.Floor(float64(p.X))), int(math.Floor(float64(p.Y)))
	solid := func(x, y int) bool {
		return pixelSolid(solids, rl.Vector2{X: float32(x) + 0.5, Y: float32(y) + 0.5})
	}
	shape := b.splat(d).keep(x, y, func(x, y int) bool {
		return solid(x, y) || solid(x+1, y) || solid(x-1, y) || solid(x, y+1) || solid(x, y-1)
	})
	// on the side of a wall it runs down along it; on a floor it has nowhere to go
	run := b.runDown(x, y, func(x, y int) bool {
		return !solid(x, y) && (solid(x+1, y) || solid(x-1, y))
	})
	return stain{x: int16(x), y: int16(y), red: d.red, run: run, tick: tick, shape: shape}
}

// fly moves every droplet on by dt under gravity. One that meets the level's
// collision sticks as a stain at the last pixel it was free in; one that has
// flown its reach and is over a painted pixel of a tile splashes onto it as a
// streak; both are stamped with tick. One that has flown too long, or fallen
// past the bottom of the level, is dropped. Each droplet moves in steps of at
// most a pixel, so none passes through a thin wall between two frames.
func (b *bloodState) fly(solids []rl.Rectangle, tiles []LevelObject, dt float32, tick uint32) {
	bottom := float32(math.Inf(-1))
	for _, r := range solids {
		bottom = max(bottom, r.Y+r.Height)
	}
	kept := b.drops[:0]
	for _, d := range b.drops {
		d.age += dt
		d.vel.Y += BloodGravity * dt
		travel := float32(math.Hypot(float64(d.vel.X), float64(d.vel.Y))) * dt
		n := max(1, int(math.Ceil(float64(travel))))
		step := rl.Vector2{X: d.vel.X * dt / float32(n), Y: d.vel.Y * dt / float32(n)}
		landed := false
		for range n {
			next := rl.Vector2{X: d.pos.X + step.X, Y: d.pos.Y + step.Y}
			if pixelSolid(solids, next) {
				b.addStain(b.hug(d, d.pos, solids, tick))
				landed = true
				break
			}
			d.pos = next
			if d.age >= d.reach && paintedAt(tiles, d.pos) {
				b.addStain(b.streak(d, d.pos, solids, tiles, tick))
				landed = true
				break
			}
		}
		if landed || d.age > BloodLife || d.pos.Y > bottom {
			continue
		}
		kept = append(kept, d)
	}
	b.drops = kept
}

// bled is how many droplets a fighter's change since the last look spills:
// none unless its structure dropped because a blow landed. Block-all is the one
// way to lose structure while guarding - a guarded block pays it back - and a
// clash costs it without a blow landing, so neither bleeds; a fighter thrown
// into a block at the wrong height is hit out of it, and does. Block-all is
// judged by the guard the fighter had before the drop: nothing lands on a
// fighter guarding everything, and one that is hit can have its guard moved in
// the same instant - a bot adapting to the blow - to block-all among others.
// More for a stun, most for a kill.
func bled(was seenFighter, a *Actor) int {
	if a.Structure >= was.structure || was.updown == 3 || a.State == StateClash {
		return 0
	}
	switch {
	case a.State == StateDeath && was.state != StateDeath:
		return BloodKill
	case a.State == StateAttacked && was.state != StateAttacked:
		return BloodStun
	}
	return BloodFlinch
}

// clashImpactsPassed is how many of the clash's impacts a fighter's animation
// has reached since the last look. Only the left fighter of the two is
// counted, so each impact is heard once.
func clashImpactsPassed(was seenFighter, a *Actor) int {
	if a.State != StateClash || a.Orientation != 1 {
		return 0
	}
	from := -1
	if was.state == StateClash {
		from = int(was.frame)
	}
	n := 0
	for _, f := range clashImpacts(a.Anim) {
		if int(f) > from && f <= a.AnimCurrentFrame {
			n++
		}
	}
	return n
}

// sprayDir is the way blood leaves a fighter: away from whoever hit it, or, with
// nobody to blame, backwards.
func sprayDir(a *Actor) float32 {
	if at := a.Attacker; at != nil {
		if at.Pos.X == a.Pos.X {
			return float32(at.Orientation)
		}
		return float32(facing(a.Pos.X - at.Pos.X))
	}
	return float32(-a.Orientation)
}

// watch sprays blood for every blow that has landed since the last look, and
// is where a blow stopped by a guard is heard.
func (b *bloodState) watch(w *World) {
	for _, a := range w.Actors {
		if was, ok := b.seen[a]; ok {
			if n := bled(was, a); n > 0 {
				r := a.collisionRect()
				b.spray(rl.Vector2{X: r.X + r.Width/2, Y: r.Y + r.Height/2}, sprayDir(a), n)
				punchHeard() // a blow that spills blood is a blow that is heard
			}
			if a.Blocks != was.blocks {
				blockHeard()
			}
			for range clashImpactsPassed(was, a) {
				punchHeard()
			}
		}
	}
	b.reseed(w)
}

// reseed makes what the blood has seen what the world holds now, spraying
// nothing: after a restore, a drop or rise in structure is time moving, not a
// blow.
func (b *bloodState) reseed(w *World) {
	clear(b.seen)
	for _, a := range w.Actors {
		b.seen[a] = seenFighter{a.Structure, a.State, a.Updown, a.Blocks, a.AnimCurrentFrame}
	}
}

// updateBlood runs the blood for one rendered frame. While a recorded frame is
// shown nothing sprays or flies and no droplet is left in the air; leaving a
// recorded frame for play deletes the stains after it, which belong to a future
// that is not going to happen now. While the match is otherwise held the
// droplets hold still.
func updateBlood(w *World, solids []rl.Rectangle, tiles []LevelObject, dt float32) {
	if w == nil {
		return
	}
	b := blood
	if TimeShowsHistory() {
		b.drops = b.drops[:0]
		b.reseed(w)
		b.history = true
		return
	}
	if b.history {
		b.dropStainsAfter(w.Tick)
		b.reseed(w)
		b.history = false
	}
	if matchHeld() {
		b.reseed(w)
		return
	}
	b.watch(w)
	b.fly(solids, tiles, dt, w.Tick)
}

// stainShown reports whether a stain is drawn this frame: while a recorded frame
// is shown, only the stains it had.
func stainShown(s *stain, w *World) bool {
	return !TimeShowsHistory() || s.tick <= w.Tick
}

// drawStains draws the stains in view. It is called inside the lit scene,
// between the level and the fighters.
func drawStains(w *World, view rl.Rectangle) {
	b := blood
	for i := range b.n {
		s := b.stainAt(i)
		x, y := float32(s.x), float32(s.y)
		const m = max(blotR, BloodStreakMax) // a clump or a tail reaches past its head
		if x < view.X-m || x >= view.X+view.Width+m || y < view.Y-m-BloodRunMax || y >= view.Y+view.Height+m || !stainShown(s, w) {
			continue
		}
		drawStain(s, ranDown(s, w.Tick))
	}
}

// drawBlot draws a clump centered on x, y: its middle in red, its rim - the
// pixels more than a step from the center - one red darker.
func drawBlot(bl blot, x, y int32, red uint8) {
	rim := bloodReds[max(int(red)-1, 0)]
	(bl | blotCenter).pixels(func(dx, dy int) {
		col := bloodReds[red]
		if iabs(dx)+iabs(dy) > 1 {
			col = rim
		}
		rl.DrawRectangle(x+int32(dx), y+int32(dy), 1, 1, col)
	})
}

// ranDown is how many pixels of its run a stain has run by tick: fast at
// first, slowing to a stop BloodRunTime seconds after it landed.
func ranDown(s *stain, tick uint32) int {
	if s.run == 0 || tick <= s.tick {
		return 0
	}
	t := min(1, float64(tick-s.tick)*FixedTimeStep/BloodRunTime)
	return int(math.Round(float64(s.run) * (1 - (1-t)*(1-t))))
}

// drawStain draws a stain's clump, and a streak's tail as its line of pixels
// back from the head, one red darker; then ran pixels of its run below the
// head, ending in a bead of its own red.
func drawStain(s *stain, ran int) {
	for k := 1; k <= ran; k++ {
		col := bloodReds[max(int(s.red)-1, 0)]
		if k == ran {
			col = bloodReds[s.red]
		}
		rl.DrawRectangle(int32(s.x), int32(s.y)+int32(k), 1, 1, col)
	}
	drawBlot(s.shape, int32(s.x), int32(s.y), s.red)
	n := max(iabs(int(s.tx)), iabs(int(s.ty)))
	tail := bloodReds[max(int(s.red)-1, 0)]
	for i := 1; i <= n; i++ {
		x := int32(s.x) + int32(math.Round(float64(s.tx)*float64(i)/float64(n)))
		y := int32(s.y) + int32(math.Round(float64(s.ty)*float64(i)/float64(n)))
		rl.DrawRectangle(x, y, 1, 1, tail)
	}
}

// drawDroplets draws the droplets in flight, over the fighters and still inside
// the lit scene.
func drawDroplets(view rl.Rectangle) {
	for _, d := range blood.drops {
		if !rl.CheckCollisionPointRec(d.pos, view) {
			continue
		}
		drawBlot(d.shape, int32(math.Floor(float64(d.pos.X))), int32(math.Floor(float64(d.pos.Y))), d.red)
	}
}

func iabs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
