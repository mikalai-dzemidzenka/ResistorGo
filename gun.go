package resistor

import (
	"math"
	"math/rand"
	"slices"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	WeaponNone   uint8 = 0
	WeaponPistol uint8 = 1
	weaponCount        = 2 // one past the last weapon; a replicated one is checked against it

	ShotRange         = 240 // px a shot carries when nothing stops it
	ShotCooldownTicks = 15  // a quarter of a second between shots
	FlashSec          = 0.05
	FlashRadius       = 80
	TracerSec         = 0.06
)

// The pistol sheet's five directions, up to down: the angle each is drawn at, in
// degrees below horizontal, and where its muzzle is in the 48x48 frame when the
// fighter faces right. Measured from pistol-top.png; to be tuned by eye in play.
var (
	aimFrameDeg = [5]float32{-55, -20, 0, 30, 55}
	muzzleAt    = [5]rl.Vector2{{X: 38, Y: 5}, {X: 41, Y: 11}, {X: 40, Y: 15}, {X: 38, Y: 23}, {X: 37, Y: 27}}
)

// gunShoulder is the point aim is measured from, facing right, in the frame.
var gunShoulder = rl.Vector2{X: 26, Y: 13}

// AimUpDeg and AimDownDeg are the ends of the aim: the sheet's first and last
// directions, so a shot never leaves at an angle the arm is not drawn at.
var (
	AimUpDeg   = aimFrameDeg[0]
	AimDownDeg = aimFrameDeg[len(aimFrameDeg)-1]
)

// framePoint turns a point in the frame drawn facing right into the world,
// mirrored when the fighter faces left: the sheet is flipped by a negative
// source width, so column x lands on column 47-x.
func framePoint(pos rl.Vector2, o int8, p rl.Vector2) rl.Vector2 {
	if o < 0 {
		p.X = 47 - p.X
	}
	return rl.Vector2{X: pos.X + p.X, Y: pos.Y + p.Y}
}

// gunAim turns the cursor's direction from the shoulder into an aim: 0 at
// AimUpDeg or higher, 255 at AimDownDeg or lower. It is measured toward the
// side the cursor is on, which is the side the fighter faces.
func gunAim(from, cursor rl.Vector2) uint8 {
	deg := float32(math.Atan2(float64(cursor.Y-from.Y), math.Abs(float64(cursor.X-from.X))) * 180 / math.Pi)
	return uint8(math.Round(float64(clamp((deg-AimUpDeg)/(AimDownDeg-AimUpDeg), 0, 1) * 255)))
}

// aimDeg is the angle an aim stands for, in degrees below horizontal.
func aimDeg(aim uint8) float32 {
	return lerp(AimUpDeg, AimDownDeg, float32(aim)/255)
}

// aimFrame is the sheet's direction nearest the aim.
func aimFrame(aim uint8) uint8 {
	deg := aimDeg(aim)
	best := 0
	for i, d := range aimFrameDeg {
		if abs(d-deg) < abs(aimFrameDeg[best]-deg) {
			best = i
		}
	}
	return uint8(best)
}

// muzzle is where a shot leaves from: the drawn frame's muzzle.
func (a *Actor) muzzle() rl.Vector2 {
	return framePoint(a.Pos, a.Orientation, muzzleAt[aimFrame(a.Aim)])
}

// shotDir is the unit direction a shot travels along: the exact aim, not the
// frame it is drawn with, turned the way the fighter faces.
func (a *Actor) shotDir() rl.Vector2 {
	rad := float64(aimDeg(a.Aim)) * math.Pi / 180
	return rl.Vector2{X: float32(math.Cos(rad)) * float32(a.Orientation), Y: float32(math.Sin(rad))}
}

// segEnter is how far along the segment from p by d it first enters r, as a
// fraction of d, and whether it enters at all. A segment starting inside r
// enters at 0.
func segEnter(p, d rl.Vector2, r rl.Rectangle) (float32, bool) {
	t0, t1 := float32(0), float32(1)
	if !clipSlab(p.X, d.X, r.X, r.X+r.Width, &t0, &t1) {
		return 0, false
	}
	if !clipSlab(p.Y, d.Y, r.Y, r.Y+r.Height, &t0, &t1) {
		return 0, false
	}
	return t0, true
}

// shotEnd is where a shot fired now stops, and the fighter it stops at, if any:
// the nearer of the first solid it crosses - drop-through platforms let it by -
// and the first living fighter that is not the shooter's ally. It reads the
// world as it stands, so the server's rewind is what makes it lag-compensated,
// and every machine can call it again to draw the tracer.
func (a *Actor) shotEnd() (rl.Vector2, *Actor) {
	p := a.muzzle()
	dir := a.shotDir()
	d := rl.Vector2{X: dir.X * ShotRange, Y: dir.Y * ShotRange}
	best := float32(1)
	var hit *Actor
	for i, c := range LE.solids {
		if dropThrough(i) {
			continue
		}
		if t, ok := segEnter(p, d, c); ok && t < best {
			best = t
		}
	}
	if a.W != nil {
		for _, b := range a.W.Actors {
			if b == a || b.State == StateDeath || a.allyOf(b) {
				continue
			}
			if t, ok := segEnter(p, d, b.collisionRect()); ok && t < best {
				best, hit = t, b
			}
		}
	}
	return rl.Vector2{X: p.X + d.X*best, Y: p.Y + d.Y*best}, hit
}

const (
	gunLegs   = "resources/threefour/gun-bot-idle.png"
	pistolTop = "resources/threefour/pistol-top.png"
)

// toggleGun draws the pistol, or holsters it. Drawn from block-all, the guard
// drops to mid: block-all roots a fighter, and an armed one has no guard.
func (a *Actor) toggleGun() {
	if a.Weapon != WeaponNone {
		a.Weapon = WeaponNone
		a.SetAnimFrame("resources/threefour/updown.png", a.Updown)
		return
	}
	a.Weapon = WeaponPistol
	if a.Updown == 3 {
		a.Updown = 1
	}
	a.SetAnimFrame(gunLegs, 0)
}

// updateArmed is standing and walking with a gun: the legs hold their idle
// frame (until there are walking ones), the aim follows the command, and LMB
// fires rather than swings. It reports whether the fighter left the stand.
func (a *Actor) updateArmed() bool {
	a.Aim = a.Input.Aim
	if a.Anim != gunLegs {
		a.SetAnimFrame(gunLegs, 0)
	}
	if a.Input.Roll && a.OnGround && a.enterDuckState() {
		return true
	}
	if a.Input.Attack {
		a.fire()
	}
	if a.Input.Jump {
		a.enterJumpStartState()
		return true
	}
	if a.Input.Down && a.OnGround && a.onDropThrough() {
		a.enterClimbDownState()
		return true
	}
	return false
}

// fire shoots along the aim, if the gun is ready. Every machine counts the shot,
// so each shows it; only the authority decides what it hits. A stunned or
// clashing fighter stops the shot and takes nothing from it, as it would a
// swing; a blocking one is taken out of the block, as a swing that got past
// it does.
func (a *Actor) fire() {
	if a.ShotCooldown > 0 {
		return
	}
	a.ShotCooldown = ShotCooldownTicks
	a.Shots++
	if a.W == nil || !a.W.Authority {
		return
	}
	_, b := a.shotEnd()
	if b == nil || b.State == StateAttacked || b.State == StateClash {
		return
	}
	if b.State == StateBlock {
		b.exitBlockState()
	}
	b.takeHit(a)
	a.FreezeTicks = 0 // hit-stop jars the fist that landed; a gun's hand is not frozen
}

// A shot as shown: a tracer from the muzzle to where it stopped, and the flash
// at the muzzle. Local presentation, aged by wall-clock time like the blood;
// not simulation state, so time manipulation leaves them be.
type tracer struct {
	from, to rl.Vector2
	left     float32
}

type flash struct {
	at, dir rl.Vector2
	left    float32
}

var (
	tracers []tracer
	flashes []flash
)

var (
	tracerColor = rl.Color{R: 255, G: 240, B: 180, A: 255}
	flashColor  = [3]uint8{255, 200, 120}
)

// showShots shows every shot fired since the last frame: a fighter whose shot
// count has moved on since it was last seen. The count is replicated and a
// client's replay arrives at the same one, so each shot is shown once on every
// machine, whoever fired it and however many times its tick was simulated.
func showShots(w *World) {
	for _, a := range w.Actors {
		if a.Shots == a.SeenShots {
			continue
		}
		a.SeenShots = a.Shots
		end, hit := a.shotEnd()
		from := a.muzzle()
		// stopped short with no fighter to stop it: a wall
		wall := hit == nil && rl.Vector2Distance(from, end) < ShotRange-0.5
		// drawn from the sprite, not the simulated position it lags or leads
		dp := a.drawPos()
		off := rl.Vector2{X: dp.X - a.Pos.X, Y: dp.Y - a.Pos.Y}
		from = rl.Vector2{X: from.X + off.X, Y: from.Y + off.Y}
		end = rl.Vector2{X: end.X + off.X, Y: end.Y + off.Y}
		tracers = append(tracers, tracer{from: from, to: end, left: TracerSec})
		dir := a.shotDir()
		flashes = append(flashes, flash{at: from, dir: dir, left: FlashSec})
		shotHeard()
		debris = append(debris, newCasing(from, dir, a.Orientation))
		if a == w.Local {
			camKick = rl.Vector2{X: -dir.X * CamKickPx, Y: -dir.Y * CamKickPx}
		}
		for range SmokePuffs {
			puffs = append(puffs, newPuff(from, dir, smoke))
		}
		if wall {
			back := rl.Vector2{X: -dir.X, Y: -dir.Y}
			for range DustPuffs {
				puffs = append(puffs, newPuff(end, back, dust))
			}
			out := rl.Vector2{X: end.X - dir.X*0.5, Y: end.Y - dir.Y*0.5}
			for range Sparks {
				debris = append(debris, newSpark(out, back))
			}
			addHole(rl.Vector2{X: end.X + dir.X*0.5, Y: end.Y + dir.Y*0.5})
		}
	}
}

// ageShots runs the shown shots down and lets the spent ones go.
func ageShots(dt float32) {
	tracers = slices.DeleteFunc(tracers, func(t tracer) bool { return t.left <= 0 })
	flashes = slices.DeleteFunc(flashes, func(f flash) bool { return f.left <= 0 })
	puffs = slices.DeleteFunc(puffs, func(p puff) bool { return p.age >= p.kind.life })
	for i := range puffs {
		puffs[i].step(dt)
	}
	debris = slices.DeleteFunc(debris, func(b bit) bool { return b.age >= b.life })
	for i := range debris {
		debris[i].step(LE.solids, dt)
	}
	holes = slices.DeleteFunc(holes, func(h hole) bool { return h.age >= HoleSec })
	for i := range holes {
		holes[i].age += dt
	}
	keep := float32(math.Exp(-CamKickEase * float64(dt)))
	camKick = rl.Vector2{X: camKick.X * keep, Y: camKick.Y * keep}
	for i := range tracers {
		tracers[i].left -= dt
	}
	for i := range flashes {
		flashes[i].left -= dt
	}
}

// withFlashes is the level's lights with a light at every muzzle flash. A new
// slice: the level's own is what the editor saves.
func withFlashes(lights []LevelLight) []LevelLight {
	if len(flashes) == 0 {
		return lights
	}
	all := slices.Clone(lights)
	for _, f := range flashes {
		all = append(all, LevelLight{
			X: int32(f.at.X), Y: int32(f.at.Y),
			Radius: FlashRadius, Color: flashColor, Intensity: 1,
		})
	}
	return all
}

// drawShotGlow draws what a shot lights up by itself - the tracer, the fire at
// the muzzle and the sparks - over the lit scene, so the dark cannot dim it.
func drawShotGlow() {
	for _, t := range tracers {
		rl.DrawLineV(t.from, t.to, tracerColor)
	}
	for _, f := range flashes {
		drawMuzzleFire(f.at, f.dir)
	}
	for _, b := range debris {
		if b.spark {
			rl.DrawPixel(int32(math.Floor(float64(b.pos.X))), int32(math.Floor(float64(b.pos.Y))), sparkColor)
		}
	}
}

// drawMuzzleFire is a small star of fire at the muzzle: a tongue along the
// shot, white at the root and yellow at the tip, and two short prongs either
// side of it.
func drawMuzzleFire(at, dir rl.Vector2) {
	n := rl.Vector2{X: -dir.Y, Y: dir.X}
	px := func(along, side float32, c rl.Color) {
		p := rl.Vector2{X: at.X + dir.X*along + n.X*side, Y: at.Y + dir.Y*along + n.Y*side}
		rl.DrawPixel(int32(math.Floor(float64(p.X))), int32(math.Floor(float64(p.Y))), c)
	}
	for i := float32(0); i <= 4; i++ {
		c := fireWhite
		if i > 2 {
			c = fireYellow
		}
		px(i, 0, c)
	}
	for _, side := range []float32{-1, 1} {
		px(1, side, fireYellow)
		px(2, side*2, fireOrange)
	}
}

const (
	SmokePuffs = 5
	DustPuffs  = 3
)

// puffKind is what smoke and dust differ in. A feel table, to be tuned by eye.
type puffKind struct {
	life       float32    // seconds
	speed      float32    // px/s it leaves at, along its direction
	spread     float32    // px/s of random scatter on top
	drag       float32    // per second: how fast it loses its speed
	lift       float32    // px/s² it rises with
	grow, size float32    // radius it starts at, and ends at
	colors     []rl.Color // lightest first
}

var (
	smoke = &puffKind{life: 0.9, speed: 90, spread: 24, drag: 4, lift: 80, grow: 0.5, size: 3.5, colors: []rl.Color{
		{R: 199, G: 207, B: 221, A: 255}, {R: 180, G: 180, B: 180, A: 255}, {R: 156, G: 156, B: 156, A: 255},
		{R: 133, G: 133, B: 133, A: 255}, {R: 113, G: 113, B: 113, A: 255}, {R: 93, G: 93, B: 93, A: 255},
	}}
	dust = &puffKind{life: 0.5, speed: 25, spread: 10, drag: 5, lift: 20, grow: 0.5, size: 2, colors: []rl.Color{{R: 150, G: 130, B: 105, A: 255}}}
)

// A puff of smoke or dust: local presentation, like the tracer.
type puff struct {
	pos, vel rl.Vector2
	age      float32
	shade    float32 // where on its kind's ramp it starts, 0 to 0.5
	kind     *puffKind
}

var puffs []puff

func newPuff(at, dir rl.Vector2, k *puffKind) puff {
	s := k.speed * (0.5 + rand.Float32()*0.7)
	return puff{
		pos: at,
		vel: rl.Vector2{
			X: dir.X*s + (rand.Float32()*2-1)*k.spread,
			Y: dir.Y*s + (rand.Float32()*2-1)*k.spread,
		},
		shade: rand.Float32() * 0.5,
		kind:  k,
	}
}

// step slows the puff, lifts it and ages it.
func (p *puff) step(dt float32) {
	keep := float32(math.Exp(float64(-p.kind.drag * dt)))
	p.vel.X *= keep
	p.vel.Y = p.vel.Y*keep - p.kind.lift*dt
	p.pos.X += p.vel.X * dt
	p.pos.Y += p.vel.Y * dt
	p.age += dt
}

// drawPuffs draws every puff as a pixel disc that loses pixels as it ages -
// whole, then a checkerboard, then one pixel in four - rather than fading
// through alpha, which the palette pass would turn into bands.
func drawPuffs(view rl.Rectangle) {
	for _, p := range puffs {
		if !rl.CheckCollisionPointRec(p.pos, view) {
			continue
		}
		f := p.age / p.kind.life
		r := lerp(p.kind.grow, p.kind.size, float32(math.Sqrt(float64(f))))
		cx, cy := int32(math.Floor(float64(p.pos.X))), int32(math.Floor(float64(p.pos.Y)))
		ri := int32(math.Ceil(float64(r)))
		for dy := -ri; dy <= ri; dy++ {
			for dx := -ri; dx <= ri; dx++ {
				if float32(dx*dx+dy*dy) > r*r || !puffPixel(f, cx+dx, cy+dy) {
					continue
				}
				rl.DrawPixel(cx+dx, cy+dy, puffShade(p.kind.colors, p.shade, f))
			}
		}
	}
}

// puffShade is a puff's one flat colour from a lightest-first ramp: each puff
// starts at its own shade, so a shot's puffs differ, and steps darker over the
// half of the ramp below it as it ages (f from 0 to 1).
func puffShade(ramp []rl.Color, shade, f float32) rl.Color {
	return ramp[min(int((shade+0.5*f)*float32(len(ramp))), len(ramp)-1)]
}

// puffPixel is whether a puff this far through its life keeps the world pixel
// x, y. The pattern is fixed to the world, so a drifting puff thins out rather
// than shimmering.
func puffPixel(f float32, x, y int32) bool {
	switch {
	case f < 0.45:
		return true
	case f < 0.75:
		return (x+y)&1 == 0
	default:
		return x&1 == 0 && y&1 == 0
	}
}

const (
	RecoilTicks = 6  // how long the torso shows the kick after a shot
	CamKickPx   = 12 // how far the local shooter's view jolts, in screen pixels
	CamKickEase = 30 // per second: how fast it settles back
	Sparks      = 4
	HoleSec     = 20
	MaxHoles    = 64
	bitGravity  = 500
	bitMaxFall  = 180 // px/s: a casing falls at most 3 px a tick, so no floor is thin enough to fall through
	casingSec   = 3
)

var (
	fireWhite   = rl.Color{R: 255, G: 255, B: 230, A: 255}
	fireYellow  = rl.Color{R: 255, G: 220, B: 90, A: 255}
	fireOrange  = rl.Color{R: 255, G: 150, B: 50, A: 255}
	sparkColor  = rl.Color{R: 255, G: 235, B: 150, A: 255}
	casingColor = rl.Color{R: 205, G: 165, B: 70, A: 255}
	holeColor   = rl.Color{R: 25, G: 20, B: 20, A: 255}
)

// camKick is the local shooter's view jolt, added to the camera's offset and
// eased back to nothing.
var camKick rl.Vector2

// torsoFrame is the pistol frame a fighter is drawn with: the aim's, kicked one
// direction up for the first ticks after a shot. It is worked out from the
// replicated cooldown, so every machine shows the same kick. At the top
// direction there is none higher, and the kick is a pixel back instead.
func (a *Actor) torsoFrame() (frame uint8, back bool) {
	frame = aimFrame(a.Aim)
	if a.ShotCooldown <= ShotCooldownTicks-RecoilTicks {
		return frame, false
	}
	if frame == 0 {
		return 0, true
	}
	return frame - 1, false
}

// A bit is a casing or a spark: a point that flies, falls and bounces off the
// level. Local presentation, like the puffs.
type bit struct {
	pos, vel  rl.Vector2
	age, life float32
	spark     bool
	rest      bool
}

var debris []bit

// newCasing throws a casing out of the pistol, back over the shooter's shoulder.
func newCasing(muzzle, dir rl.Vector2, o int8) bit {
	return bit{
		pos:  rl.Vector2{X: muzzle.X - dir.X*6, Y: muzzle.Y - dir.Y*6},
		vel:  rl.Vector2{X: -float32(o) * (25 + rand.Float32()*20), Y: -(60 + rand.Float32()*30)},
		life: casingSec,
	}
}

// newSpark throws a spark back out of a wall, scattered about the way it came.
func newSpark(at, back rl.Vector2) bit {
	s := 60 + rand.Float32()*80
	return bit{
		pos:   at,
		vel:   rl.Vector2{X: back.X*s + (rand.Float32()*2-1)*60, Y: back.Y*s + (rand.Float32()*2-1)*60},
		life:  0.08 + rand.Float32()*0.07,
		spark: true,
	}
}

// step moves the bit a frame: gravity, then each axis on its own against the
// level, so a casing slides along a floor and bounces off a wall. A spark that
// touches the level goes out; a casing loses most of its speed, and one landing
// slowly enough stays where it is.
func (b *bit) step(solids []rl.Rectangle, dt float32) {
	b.age += dt
	if b.rest {
		return
	}
	b.vel.Y = min(b.vel.Y+bitGravity*dt, bitMaxFall)
	nx := b.pos.X + b.vel.X*dt
	if pixelSolid(solids, rl.Vector2{X: nx, Y: b.pos.Y}) {
		if b.spark {
			b.age = b.life
			return
		}
		b.vel.X = -b.vel.X * 0.4
		nx = b.pos.X
	}
	ny := b.pos.Y + b.vel.Y*dt
	if pixelSolid(solids, rl.Vector2{X: nx, Y: ny}) {
		if b.spark {
			b.age = b.life
			return
		}
		if b.vel.Y > 0 && b.vel.Y < 40 {
			b.rest = true
		}
		b.vel.Y = -b.vel.Y * 0.35
		b.vel.X *= 0.6
		ny = b.pos.Y
	}
	b.pos = rl.Vector2{X: nx, Y: ny}
}

// drawCasings draws every casing: 2x1 lying still, tumbling end over end while
// it flies.
func drawCasings(view rl.Rectangle) {
	for _, b := range debris {
		if b.spark || !rl.CheckCollisionPointRec(b.pos, view) {
			continue
		}
		x, y := int32(math.Floor(float64(b.pos.X))), int32(math.Floor(float64(b.pos.Y)))
		if !b.rest && int(b.age/0.05)%2 == 1 {
			rl.DrawRectangle(x, y-1, 1, 2, casingColor)
		} else {
			rl.DrawRectangle(x, y, 2, 1, casingColor)
		}
	}
}

// A hole is where a shot went into a wall: a point just inside it.
type hole struct {
	at  rl.Vector2
	age float32
}

var holes []hole

// addHole keeps a hole, dropping the oldest once there are MaxHoles.
func addHole(at rl.Vector2) {
	if len(holes) >= MaxHoles {
		holes = holes[1:]
	}
	holes = append(holes, hole{at: at})
}

func drawHoles(view rl.Rectangle) {
	for _, h := range holes {
		if rl.CheckCollisionPointRec(h.at, view) {
			rl.DrawPixel(int32(math.Floor(float64(h.at.X))), int32(math.Floor(float64(h.at.Y))), holeColor)
		}
	}
}

// clearShots forgets every shot shown: a match starting over starts clean.
func clearShots() {
	tracers, flashes, puffs, debris, holes = nil, nil, nil, nil, nil
	camKick = rl.Vector2{}
}
