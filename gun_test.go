package resistor

import (
	"fmt"
	"testing"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func TestGunAimFrames(t *testing.T) {
	from := rl.Vector2{X: 100, Y: 100}
	cases := []struct {
		name   string
		cursor rl.Vector2
		aim    uint8
		frame  uint8
	}{
		{"level", rl.Vector2{X: 150, Y: 100}, 128, 2},
		{"level, behind", rl.Vector2{X: 50, Y: 100}, 128, 2},
		{"highest", rl.Vector2{X: 110, Y: 0}, 0, 0},
		{"lowest", rl.Vector2{X: 110, Y: 200}, 255, 4},
		{"straight up clamps", rl.Vector2{X: 100, Y: 0}, 0, 0},
	}
	for _, c := range cases {
		aim := gunAim(from, c.cursor)
		if int(aim)-int(c.aim) > 1 || int(c.aim)-int(aim) > 1 {
			t.Errorf("%s: aim %d, want %d", c.name, aim, c.aim)
		}
		if f := aimFrame(aim); f != c.frame {
			t.Errorf("%s: frame %d, want %d", c.name, f, c.frame)
		}
	}
	if d := aimDeg(0); d != AimUpDeg {
		t.Errorf("aim 0 is %v degrees, want %v", d, AimUpDeg)
	}
	if d := aimDeg(255); d != AimDownDeg {
		t.Errorf("aim 255 is %v degrees, want %v", d, AimDownDeg)
	}
}

func TestMuzzleMirrors(t *testing.T) {
	a := &Actor{Pos: rl.Vector2{X: 10, Y: 20}, Orientation: 1, Aim: 128}
	right := a.muzzle()
	a.Orientation = -1
	left := a.muzzle()
	m := muzzleAt[2]
	if right != (rl.Vector2{X: 10 + m.X, Y: 20 + m.Y}) {
		t.Errorf("facing right the muzzle is at %v", right)
	}
	if left != (rl.Vector2{X: 10 + 47 - m.X, Y: 20 + m.Y}) {
		t.Errorf("facing left the muzzle is at %v, want x mirrored to 47-x", left)
	}
	if a.shotDir().X >= 0 {
		t.Errorf("facing left the shot goes %v", a.shotDir())
	}
}

// shooter is a fighter standing on the default floor at x, facing right,
// aimed level, in a fresh authoritative world.
func shooter(t *testing.T, x float32) (*World, *Actor) {
	t.Helper()
	w, a := walker(t, x)
	a.IsPlayer = true
	a.Team = NoTeam // the targets are bots; on their team every one would be an ally
	a.Aim = 128
	return w, a
}

// target stands a fighter on the floor in w with its left edge at x.
func target(w *World, x float32) *Actor {
	b := w.Spawn(rl.Vector2{X: x - 14}, false)
	b.OnGround = true
	return b
}

func TestShotStops(t *testing.T) {
	t.Run("nothing", func(t *testing.T) {
		_, a := shooter(t, 0)
		end, hit := a.shotEnd()
		if hit != nil || abs(end.X-(a.muzzle().X+ShotRange)) > 0.01 {
			t.Errorf("an empty line ends at %v hitting %v, want full range", end, hit)
		}
	})
	t.Run("wall before enemy", func(t *testing.T) {
		w, a := shooter(t, 0)
		addSolid(100, 0, 8, 48)
		defer floor()
		target(w, 140)
		end, hit := a.shotEnd()
		if hit != nil || abs(end.X-100) > 0.01 {
			t.Errorf("shot ends at %v hitting %v, want the wall face at 100", end, hit)
		}
	})
	t.Run("nearer enemy", func(t *testing.T) {
		w, a := shooter(t, 0)
		target(w, 160)
		near := target(w, 100)
		end, hit := a.shotEnd()
		if hit != near || abs(end.X-100) > 0.01 {
			t.Errorf("shot ends at %v hitting %v, want the nearer at 100", end, hit)
		}
	})
	t.Run("ally and dead passed", func(t *testing.T) {
		w, a := shooter(t, 0)
		a.Team = RedTeam
		ally := target(w, 80)
		ally.Team = RedTeam
		dead := target(w, 100)
		dead.State = StateDeath
		enemy := target(w, 140)
		if _, hit := a.shotEnd(); hit != enemy {
			t.Errorf("shot hit %v, want the enemy past the ally and the dead", hit)
		}
	})
	t.Run("through a drop-through platform", func(t *testing.T) {
		_, a := shooter(t, 0)
		a.Aim = 255 // lowest
		defer floor()
		// a platform across the line, just below the muzzle
		p := a.muzzle()
		pos := Vector2i{X: int32(p.X), Y: int32(p.Y) + 4}
		LE.objects[0][pos] = LevelObject{X: pos.X, Y: pos.Y, DropRects: []rl.Rectangle{{Width: 30, Height: 2}}}
		LE.buildIndex()
		end, _ := a.shotEnd()
		if end.Y <= p.Y+6 {
			t.Errorf("shot ends at %v, stopped by the drop-through platform at y %d", end, pos.Y)
		}
	})
}

func TestGunToggle(t *testing.T) {
	_, a := shooter(t, 0)
	steps(a, 1, Input{Gun: true})
	if a.Weapon != WeaponPistol || a.Anim != gunLegs {
		t.Fatalf("Q while standing: weapon %d anim %s, want the pistol drawn", a.Weapon, a.Anim)
	}
	steps(a, 1, Input{})
	steps(a, 1, Input{Gun: true})
	if a.Weapon != WeaponNone || a.Anim == gunLegs {
		t.Fatalf("Q again: weapon %d anim %s, want it holstered", a.Weapon, a.Anim)
	}

	_, a = shooter(t, 0)
	steps(a, 1, Input{Jump: true})
	for range 20 {
		steps(a, 1, Input{Gun: true})
	}
	for range 120 {
		steps(a, 1, Input{})
	}
	if a.State != StateIdle || a.Weapon != WeaponNone {
		t.Errorf("Q in the air: state %d weapon %d, want standing unarmed after landing", a.State, a.Weapon)
	}

	_, a = shooter(t, -480)
	steps(a, 90, Input{Left: true})
	if a.State != StateRun {
		t.Fatalf("held a move key and got state %d, want the run", a.State)
	}
	steps(a, 1, Input{Left: true, Gun: true})
	steps(a, 1, Input{Left: true})
	if a.State != StateIdle || a.Weapon != WeaponPistol || abs(a.Speed.X) > RunReach {
		t.Errorf("Q while running: state %d weapon %d speed %v, want walking armed", a.State, a.Weapon, a.Speed.X)
	}
}

func TestArmedWalksNeverRuns(t *testing.T) {
	_, a := shooter(t, -480)
	steps(a, 1, Input{Gun: true})
	for range 120 {
		steps(a, 1, Input{Left: true})
		if a.State == StateRun {
			t.Fatal("an armed fighter holding a move key entered the run")
		}
	}
	if a.Pos.X < -430 {
		t.Errorf("an armed fighter walked only to %v in 2 s", a.Pos.X)
	}
	steps(a, 1, Input{Jump: true})
	for range 120 {
		steps(a, 1, Input{})
	}
	if a.State != StateIdle || a.Weapon != WeaponPistol || a.Anim != gunLegs {
		t.Errorf("after a jump: state %d weapon %d anim %s, want standing with the gun legs", a.State, a.Weapon, a.Anim)
	}
}

func TestArmedDoesNotGuard(t *testing.T) {
	_, att, def := swing(t, 1) // guarding the height the swing comes at
	def.Weapon = WeaponPistol
	att.updateAttack()
	if def.Blocks != 0 || def.Structure >= MaxStructure {
		t.Errorf("a swing at an armed fighter guarding its height: blocks %d structure %v, want a hit", def.Blocks, def.Structure)
	}
}

// armed is a shooter with the pistol out and a fighter to shoot 60 px ahead.
func armed(t *testing.T) (*World, *Actor, *Actor) {
	t.Helper()
	w, a := shooter(t, 0)
	steps(a, 1, Input{Gun: true, Aim: 128})
	return w, a, target(w, 60)
}

func TestShotLandsAHit(t *testing.T) {
	_, a, b := armed(t)
	steps(a, 1, Input{Attack: true, Aim: 128})
	if b.Structure != MaxStructure-HitStructure || b.FlinchTicks == 0 || b.Attacker != a {
		t.Errorf("shot target: structure %v flinch %d, want a first hit's cost and a flinch", b.Structure, b.FlinchTicks)
	}
	if a.Energy != MaxEnergy || a.Shots != 1 {
		t.Errorf("shooter: energy %d shots %d, want energy untouched and one shot", a.Energy, a.Shots)
	}

	// a second click inside the cooldown does nothing, then or later
	steps(a, 5, Input{Aim: 128})
	steps(a, 1, Input{Attack: true, Aim: 128})
	steps(a, ShotCooldownTicks, Input{Aim: 128})
	if a.Shots != 1 {
		t.Errorf("a click within the cooldown fired: %d shots", a.Shots)
	}
	steps(a, 1, Input{Attack: true, Aim: 128})
	if a.Shots != 2 {
		t.Errorf("a click after the cooldown did not fire: %d shots", a.Shots)
	}
}

func TestShotIgnoresBlockAll(t *testing.T) {
	_, a, b := armed(t)
	b.Updown = 3
	steps(a, 1, Input{Attack: true, Aim: 128})
	if b.Structure != MaxStructure-HitStructure {
		t.Errorf("block-all took a shot for structure %v, want it hit", b.Structure)
	}
}

func TestShotKills(t *testing.T) {
	_, a, b := armed(t)
	b.Structure = HitStructure
	steps(a, 1, Input{Attack: true, Aim: 128})
	if b.State != StateDeath || b.Attacker != a {
		t.Errorf("a shot at the last structure: state %d attacker %v, want dead and credited", b.State, b.Attacker)
	}
}

func TestShotOnStunnedDoesNothing(t *testing.T) {
	_, a, b := armed(t)
	b.State = StateAttacked
	steps(a, 1, Input{Attack: true, Aim: 128})
	if b.Structure != MaxStructure {
		t.Errorf("a stunned fighter shot: structure %v, want it unchanged", b.Structure)
	}
}

func TestClientShotDecidesNothing(t *testing.T) {
	w, a, b := armed(t)
	w.Authority = false
	steps(a, 1, Input{Attack: true, Aim: 128})
	if a.Shots != 1 || b.Structure != MaxStructure {
		t.Errorf("predicted shot: shots %d target structure %v, want counted and nothing hit", a.Shots, b.Structure)
	}
}

func TestShotShownOnce(t *testing.T) {
	tracers, flashes = nil, nil
	defer func() { tracers, flashes, onShot = nil, nil, nil }()
	heard := 0
	onShot = func() { heard++ }
	w, a, _ := armed(t)
	steps(a, 1, Input{Attack: true, Aim: 128})
	showShots(w)
	showShots(w)
	if len(tracers) != 1 || len(flashes) != 1 || heard != 1 {
		t.Fatalf("one shot shown as %d tracers and %d flashes, heard %d times; want one of each", len(tracers), len(flashes), heard)
	}
	if n := len(withFlashes(nil)); n != 1 {
		t.Errorf("the lights carry %d flashes, want 1", n)
	}
	ageShots(FlashSec + 0.001)
	ageShots(0)
	if len(flashes) != 0 || len(tracers) != 1 {
		t.Errorf("after the flash's life: %d flashes %d tracers, want the flash gone and the tracer still there", len(flashes), len(tracers))
	}
	ageShots(TracerSec)
	ageShots(0)
	if len(tracers) != 0 {
		t.Errorf("after the tracer's life %d tracers remain", len(tracers))
	}
}

// A shot predicted ahead of the server is replayed on every correction until the
// server executes it; the count comes back the same and it is shown once.
func TestReplayShowsShotOnce(t *testing.T) {
	floor()
	Stage = StageLevel
	defer func() { Stage = StageMainMenu }()
	tracers, flashes = nil, nil
	defer func() { tracers, flashes = nil, nil }()

	sv := NewWorld(true)
	sa := sv.Spawn(rl.Vector2{}, true)
	c := &Client{W: NewWorld(false), ActorID: sa.ID}
	c.reconcile(recOf(sa), 0)
	local := c.W.Local
	dt := float32(FixedTimeStep)
	for range 20 {
		sa.Input = Input{}
		sa.Update(dt)
	}
	c.reconcile(recOf(sa), 0)

	cmds := []Input{{Gun: true, Aim: 128}, {Aim: 128}, {Attack: true, Aim: 128}, {Aim: 128}, {Aim: 128}}
	shown := 0
	for i, cmd := range cmds {
		c.Seq++
		cmd.Seq = c.Seq
		c.Ring[cmd.Seq%CmdRing] = cmd
		local.Input = cmd
		local.Update(dt)
		// the server is two commands behind, and a snapshot comes every tick
		if i >= 2 {
			done := c.Seq - 2
			sa.Input = c.Ring[done%CmdRing]
			sa.Update(dt)
			c.reconcile(recOf(sa), done)
		}
		n := len(tracers)
		showShots(c.W)
		shown += len(tracers) - n
	}
	if local.Shots != 1 || sa.Shots != 1 {
		t.Errorf("shots: client %d server %d, want 1 each", local.Shots, sa.Shots)
	}
	if shown != 1 {
		t.Errorf("the shot was shown %d times, want once", shown)
	}
}

// Q then LMB online: the server's fighter draws and fires, and the client's
// prediction agrees.
func TestNetGun(t *testing.T) {
	initNet()
	defer shutdownNet()
	Stage = StageLevel
	defer func() { Stage = StageMainMenu }()
	floor()

	cfg := DefaultServerConfig()
	cfg.Port = 27199
	cfg.MaxPlayers = 2
	s, err := NewServer(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { Srv = nil; s.Close() }()
	if err := startClient(fmt.Sprintf("127.0.0.1:%d", cfg.Port), false); err != nil {
		t.Fatal(err)
	}
	c := Cli
	pumpFor(1500*time.Millisecond, s, c)
	if !c.Welcomed || c.W.Local == nil {
		t.Fatal("client not in the match")
	}
	var srvA *Actor
	for _, p := range s.Peers {
		if p.Actor != nil {
			srvA = p.Actor
		}
	}
	if srvA == nil {
		t.Fatal("no fighter on the server")
	}

	pressedGun = true
	pumpFor(300*time.Millisecond, s, c)
	if srvA.Weapon != WeaponPistol || c.W.Local.Weapon != WeaponPistol {
		t.Fatalf("after Q: server weapon %d, client weapon %d, want the pistol on both", srvA.Weapon, c.W.Local.Weapon)
	}
	pressedAttack = true
	pumpFor(300*time.Millisecond, s, c)
	if srvA.Shots != 1 || c.W.Local.Shots != 1 {
		t.Errorf("after LMB: server shots %d, client shots %d, want 1 on both", srvA.Shots, c.W.Local.Shots)
	}
}

func TestShotSmokeAndDust(t *testing.T) {
	defer func() { tracers, flashes, puffs = nil, nil, nil }()
	count := func() (s, d int) {
		for _, p := range puffs {
			if p.kind == smoke {
				s++
			} else if p.kind == dust {
				d++
			}
		}
		return
	}

	// at a wall: smoke and dust
	tracers, flashes, puffs = nil, nil, nil
	w, a := shooter(t, 0)
	addSolid(100, 0, 8, 48)
	steps(a, 1, Input{Gun: true, Aim: 128})
	steps(a, 1, Input{Attack: true, Aim: 128})
	showShots(w)
	floor()
	if s, d := count(); s != SmokePuffs || d != DustPuffs {
		t.Errorf("a shot at a wall left %d smoke and %d dust, want %d and %d", s, d, SmokePuffs, DustPuffs)
	}
	smokeY := func() (y float32) {
		for _, p := range puffs {
			if p.kind == smoke {
				y += p.pos.Y / SmokePuffs
			}
		}
		return
	}
	y0 := smokeY()
	for range 48 {
		ageShots(1.0 / 60)
	}
	if y := smokeY(); y > y0-3 {
		t.Errorf("smoke went from y %v to %v in 0.8 s, want it risen", y0, y)
	}
	ageShots(smoke.life)
	ageShots(0)
	if len(puffs) != 0 {
		t.Errorf("%d puffs remain after their life", len(puffs))
	}

	// at a fighter: smoke only
	tracers, flashes, puffs = nil, nil, nil
	w, a, _ = armed(t)
	steps(a, 1, Input{Attack: true, Aim: 128})
	showShots(w)
	if s, d := count(); s != SmokePuffs || d != 0 {
		t.Errorf("a shot at a fighter left %d smoke and %d dust, want %d and none", s, d, SmokePuffs)
	}

	// into nothing: smoke only
	tracers, flashes, puffs = nil, nil, nil
	w, a = shooter(t, 0)
	steps(a, 1, Input{Gun: true, Aim: 128})
	steps(a, 1, Input{Attack: true, Aim: 128})
	showShots(w)
	if _, d := count(); d != 0 {
		t.Errorf("a shot into nothing left %d dust, want none", d)
	}
}

func TestShotDoesNotFreezeShooter(t *testing.T) {
	_, a, b := armed(t)
	a.Input = Input{Attack: true, Aim: 128}
	a.Update(float32(FixedTimeStep))
	if b.FreezeTicks == 0 || a.FreezeTicks != 0 {
		t.Errorf("after a landed shot: target frozen %d, shooter frozen %d; want only the target", b.FreezeTicks, a.FreezeTicks)
	}
}

func TestRecoilKick(t *testing.T) {
	_, a := shooter(t, 0)
	steps(a, 1, Input{Gun: true, Aim: 128})
	steps(a, 1, Input{Attack: true, Aim: 128})
	if f, _ := a.torsoFrame(); f != 1 {
		t.Errorf("right after a level shot the torso shows frame %d, want 1 (one up)", f)
	}
	steps(a, RecoilTicks, Input{Aim: 128})
	if f, back := a.torsoFrame(); f != 2 || back {
		t.Errorf("after the kick the torso shows frame %d back %t, want 2 at rest", f, back)
	}
	a.Aim = 0
	a.ShotCooldown = ShotCooldownTicks
	if f, back := a.torsoFrame(); f != 0 || !back {
		t.Errorf("kicked at the top aim: frame %d back %t, want 0 shifted back", f, back)
	}
}

func TestShotDebris(t *testing.T) {
	defer clearShots()
	kinds := func() (casings, sparks int) {
		for _, b := range debris {
			if b.spark {
				sparks++
			} else {
				casings++
			}
		}
		return
	}

	clearShots()
	w, a := shooter(t, 0)
	addSolid(100, 0, 8, 48)
	steps(a, 1, Input{Gun: true, Aim: 128})
	steps(a, 1, Input{Attack: true, Aim: 128})
	showShots(w)
	if c, s := kinds(); c != 1 || s != Sparks || len(holes) != 1 {
		t.Errorf("a wall shot left %d casings, %d sparks, %d holes; want 1, %d, 1", c, s, len(holes), Sparks)
	}
	if h := holes[0].at; !pixelSolid(LE.solids, h) {
		t.Errorf("the hole at %v is not in the wall", h)
	}
	// the casing lands on the floor and stays there
	for range 180 {
		ageShots(1.0 / 60)
	}
	floor()
	var casing *bit
	for i := range debris {
		if !debris[i].spark {
			casing = &debris[i]
		}
	}
	if casing == nil || !casing.rest {
		t.Fatalf("after 3 s the casing is %+v, want it lying still", casing)
	}
	if casing.pos.Y > floorY || casing.pos.Y < floorY-2 {
		t.Errorf("the casing rests at y %v, want it on the floor at %d", casing.pos.Y, floorY)
	}
	if _, s := kinds(); s != 0 {
		t.Errorf("%d sparks outlived their life", s)
	}
	ageShots(casingSec)
	ageShots(HoleSec)
	ageShots(0)
	if len(debris) != 0 || len(holes) != 0 {
		t.Errorf("after their lives: %d debris %d holes remain", len(debris), len(holes))
	}

	// a fighter shot: a casing, no sparks, no hole
	clearShots()
	w, a, _ = armed(t)
	steps(a, 1, Input{Attack: true, Aim: 128})
	showShots(w)
	if c, s := kinds(); c != 1 || s != 0 || len(holes) != 0 {
		t.Errorf("a fighter shot left %d casings, %d sparks, %d holes; want 1, 0, 0", c, s, len(holes))
	}

	clearShots()
	for i := range MaxHoles + 5 {
		addHole(rl.Vector2{X: float32(i)})
	}
	if len(holes) != MaxHoles || holes[0].at.X != 5 {
		t.Errorf("%d holes kept, oldest at %v; want %d with the first 5 dropped", len(holes), holes[0].at.X, MaxHoles)
	}
}

// A beam comes from every living fighter with the pistol drawn, and only while
// the setting is on.
func TestFlashlightBeams(t *testing.T) {
	defer func(on bool) { Cfg.Flashlight = on }(Cfg.Flashlight)
	armed, bare, dead := NewActor(ZeroVec2, true), NewActor(ZeroVec2, true), NewActor(ZeroVec2, true)
	armed.toggleGun()
	dead.toggleGun()
	dead.State = StateDeath
	actors := []*Actor{armed, bare, dead}
	Cfg.Flashlight = false
	if bs := flashlightBeams(actors); len(bs) != 0 {
		t.Errorf("%d beams with the setting off", len(bs))
	}
	Cfg.Flashlight = true
	bs := flashlightBeams(actors)
	if len(bs) != 1 || bs[0].dir != armed.shotDir() {
		t.Errorf("beams %v, want one along the armed fighter's aim", bs)
	}
	// on the climb the pistol is not drawn, and neither is its beam
	armed.enterClimbState()
	if bs := flashlightBeams(actors); len(bs) != 0 {
		t.Errorf("%d beams from a climbing fighter", len(bs))
	}
}

// A puff's colour starts at its own shade and darkens with age, never off its ramp.
func TestPuffShade(t *testing.T) {
	r := smoke.colors
	if puffShade(r, 0, 0) != r[0] || puffShade(r, 0.5, 1) != r[len(r)-1] {
		t.Error("the ramp's ends are not reached")
	}
	if puffShade(r, 0, 1) == puffShade(r, 0, 0) {
		t.Error("a puff does not darken as it ages")
	}
	if puffShade(r, 0.4, 0) == puffShade(r, 0, 0) {
		t.Error("puffs of different shades look alike")
	}
}
