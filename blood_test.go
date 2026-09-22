package resistor

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/token"
	"image"
	"image/color"
	"image/png"
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// flyFor runs the blood's droplets for the given seconds at 60 fps.
func flyFor(seconds float64, tick uint32) {
	for range int(seconds * 60) {
		blood.fly(LE.solids, nil, 1.0/60, tick)
	}
}

// Droplets sprayed over the floor land on it as stains, and none are left in
// the air once their life is over.
func TestBloodLandsOnTheFloor(t *testing.T) {
	floor()
	clearBlood()
	t.Cleanup(clearBlood)
	blood.spray(rl.Vector2{X: 0, Y: floorY - 20}, 1, 50)
	flyFor(BloodLife+1, 7)
	if len(blood.drops) != 0 {
		t.Errorf("%d droplets still in the air", len(blood.drops))
	}
	if blood.n == 0 {
		t.Fatal("no droplet landed")
	}
	for i := range blood.n {
		s := blood.stainAt(i)
		if int(s.y) != floorY-1 {
			t.Errorf("a stain at y=%d, want on top of the floor at %d", s.y, floorY-1)
		}
		if s.tick != 7 {
			t.Errorf("a stain stamped %d, want the tick it landed on", s.tick)
		}
	}
}

// Sprayed toward a wall close by, some of the blood stains the wall itself.
func TestBloodStainsAWall(t *testing.T) {
	walls()
	clearBlood()
	t.Cleanup(clearBlood)
	blood.spray(rl.Vector2{X: wallX + 6, Y: floorY - 24}, -1, 60)
	flyFor(BloodLife+1, 0)
	onWall := 0
	for i := range blood.n {
		if s := blood.stainAt(i); int(s.x) == wallX && int(s.y) < floorY-1 {
			onWall++
		}
	}
	if onWall == 0 {
		t.Error("none of the blood stained the wall")
	}
}

// Past the caps, the oldest droplets and stains give way to the newest.
func TestBloodIsCapped(t *testing.T) {
	floor()
	clearBlood()
	t.Cleanup(clearBlood)
	blood.spray(rl.Vector2{}, 1, BloodDroplets+100)
	if len(blood.drops) != BloodDroplets {
		t.Errorf("%d droplets in flight, want the cap %d", len(blood.drops), BloodDroplets)
	}
	for i := range BloodStains + 1000 {
		blood.addStain(stain{tick: uint32(i)})
	}
	if blood.n != BloodStains {
		t.Fatalf("%d stains, want the cap %d", blood.n, BloodStains)
	}
	for i := range blood.n {
		if want := uint32(1000 + i); blood.stainAt(i).tick != want {
			t.Fatalf("stain %d is from tick %d, want %d: the newest kept, in order", i, blood.stainAt(i).tick, want)
		}
	}
}

// Blood draws only from its own random source. The bots draw from the global
// one, so a single global draw here would make turning blood on change what a
// bot does.
func TestBloodHasItsOwnRandom(t *testing.T) {
	f, err := parser.ParseFile(token.NewFileSet(), "blood.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	ast.Inspect(f, func(n ast.Node) bool {
		sel, ok := n.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		if pkg, ok := sel.X.(*ast.Ident); ok && pkg.Name == "rand" && sel.Sel.Name != "New" && sel.Sel.Name != "NewSource" && sel.Sel.Name != "Rand" {
			t.Errorf("blood.go calls the global rand.%s", sel.Sel.Name)
		}
		return true
	})
}

// bleeding stands a fighter on the floor with a bot to its left facing it,
// the blood watching both, and returns the droplets the given blow sprays.
// prep, if given, sets the fighters up before the blood first looks at them.
func bleeding(t *testing.T, blow func(w *World, def, att *Actor), prep ...func(def *Actor)) []droplet {
	t.Helper()
	floor()
	clearBlood()
	t.Cleanup(clearBlood)
	w := NewWorld(true)
	def := w.Spawn(rl.Vector2{X: 0}, false)
	att := w.Spawn(rl.Vector2{X: -30}, false)
	def.Team, att.Team = NoTeam, NoTeam
	att.Orientation, def.Orientation = 1, -1
	for range 10 {
		run(w, 1)
	}
	for _, p := range prep {
		p(def)
	}
	blood.watch(w) // the first look only remembers
	blow(w, def, att)
	blood.watch(w)
	return blood.drops
}

// A flinch spills a little, away from the attacker; a stun more, a kill most.
func TestALandedBlowSpillsBlood(t *testing.T) {
	flinch := bleeding(t, func(w *World, def, att *Actor) { def.takeHit(att) })
	if len(flinch) != BloodFlinch {
		t.Errorf("a flinch sprayed %d droplets, want %d", len(flinch), BloodFlinch)
	}
	away := 0
	for _, d := range flinch {
		if d.vel.X > 0 {
			away++
		}
	}
	if away*2 <= len(flinch) {
		t.Errorf("%d of %d droplets flew away from the attacker, want most", away, len(flinch))
	}
	stun := bleeding(t, func(w *World, def, att *Actor) {
		def.HitCount = StunEveryHits - 1
		def.takeHit(att)
	})
	kill := bleeding(t, func(w *World, def, att *Actor) {
		def.Structure = 1
		def.takeHit(att)
	})
	if len(stun) != BloodStun || len(kill) != BloodKill {
		t.Errorf("a stun sprayed %d and a kill %d, want %d and %d", len(stun), len(kill), BloodStun, BloodKill)
	}
	// the hit moves the guard to block-all in the same instant, as a bot
	// adapting to a blow can: it still bled
	adapted := bleeding(t, func(w *World, def, att *Actor) {
		def.takeHit(att)
		def.Updown = 3
	})
	if len(adapted) != BloodFlinch {
		t.Errorf("a hit that left the guard at block-all sprayed %d droplets, want %d", len(adapted), BloodFlinch)
	}
}

// Blocks, block-all, clashes and shoves spill nothing, and neither does
// structure coming back.
func TestOnlyALandedBlowSpillsBlood(t *testing.T) {
	hurt := func(def *Actor) { def.Structure = MaxStructure - 1 }
	guardAll := func(def *Actor) { def.Updown = 3 }
	for name, c := range map[string]struct {
		blow func(w *World, def, att *Actor)
		prep []func(def *Actor)
	}{
		"a guarded block": {func(w *World, def, att *Actor) {
			def.Updown = att.Updown
			def.confirmBlock(att)
		}, []func(*Actor){hurt}},
		"block-all": {func(w *World, def, att *Actor) {
			def.confirmBlock(att)
		}, []func(*Actor){guardAll}},
		"a death from block-all": {func(w *World, def, att *Actor) {
			def.confirmBlock(att)
		}, []func(*Actor){guardAll, func(def *Actor) { def.Structure = BlockAllStructure / 2 }}},
		"a clash": {func(w *World, def, att *Actor) {
			startClash(att, def)
		}, nil},
		"a shove": {func(w *World, def, att *Actor) {
			def.enterShovedState(att)
		}, nil},
		"regeneration": {func(w *World, def, att *Actor) {
			def.restoreStructure(1)
		}, []func(*Actor){hurt}},
	} {
		if drops := bleeding(t, c.blow, c.prep...); len(drops) != 0 {
			t.Errorf("%s spilled %d droplets", name, len(drops))
		}
	}
}

// Blood is watched, sprayed and flown every frame of a fight, and the fight is
// the same as without it, tick for tick.
func TestBloodDoesNotChangeTheMatch(t *testing.T) {
	type seen struct {
		pos       rl.Vector2
		structure float32
		state     ActorState
		energy    int32
		frame     uint8
	}
	fight := func(withBlood bool) [][2]seen {
		floor()
		clearBlood()
		w := NewWorld(true)
		// players: a bot hit draws from the global random source to adapt its
		// guard, which would make two runs differ with or without blood
		a := w.Spawn(rl.Vector2{X: 0}, true)
		b := w.Spawn(rl.Vector2{X: 26}, true)
		a.Orientation, b.Orientation = 1, -1
		var got [][2]seen
		for tick := range 600 {
			a.Input = Input{Attack: tick%40 == 3, Updown: uint8(tick/100) % 3}
			b.Input = Input{Attack: tick%55 == 9, Updown: uint8(tick/70) % 3, FaceLeft: true}
			run(w, 1)
			if withBlood {
				blood.watch(w)
				blood.fly(LE.solids, nil, float32(FixedTimeStep), w.Tick)
			}
			got = append(got, [2]seen{
				{a.Pos, a.Structure, a.State, a.Energy, a.AnimCurrentFrame},
				{b.Pos, b.Structure, b.State, b.Energy, b.AnimCurrentFrame},
			})
		}
		return got
	}
	with := fight(true)
	if blood.n == 0 && len(blood.drops) == 0 {
		t.Fatal("no blood was spilled in the fight: nothing to compare")
	}
	without := fight(false)
	t.Cleanup(clearBlood)
	for i := range with {
		if with[i] != without[i] {
			t.Fatalf("tick %d: with blood %+v, without %+v", i, with[i], without[i])
		}
	}
}

// Rewinding hides the stains a frame did not have, going forward shows them
// again, and resuming from the frame deletes them; no restore reads as a blow.
func TestBloodFollowsTimeManipulation(t *testing.T) {
	pl := rewindMatch(t)
	clearBlood()
	t.Cleanup(clearBlood)
	bot := CL.Spawn(CL.spawnPos(), false)
	bot.Team, pl.Team = NoTeam, NoTeam
	ticks(10)
	updateBlood(CL, LE.solids, nil, 1.0/60)
	mark, early := rewindHead, CL.Tick
	blood.addStain(stain{tick: early})

	// a blow after the mark: structure down at the live edge. The bot takes it:
	// the player's guard is block-all with no mouse, which bleeds nothing
	bot.takeHit(pl)
	ticks(10)
	blood.addStain(stain{tick: CL.Tick})
	updateBlood(CL, LE.solids, nil, 1.0/60) // the blow sprays at the live edge
	if len(blood.drops) == 0 {
		t.Fatal("the blow at the live edge spilled nothing")
	}

	shown := func() (n int) {
		for i := range blood.n {
			if stainShown(blood.stainAt(i), CL) {
				n++
			}
		}
		return n
	}

	TimeMode = TimePaused
	rewindCursor = mark
	restoreFrame(CL)
	updateBlood(CL, LE.solids, nil, 1.0/60)
	if len(blood.drops) != 0 {
		t.Error("droplets stayed in the air over a recorded frame")
	}
	if n := shown(); n != 1 {
		t.Errorf("%d stains shown on the earlier frame, want only the one it had", n)
	}

	rewindCursor = rewindHead
	restoreFrame(CL)
	updateBlood(CL, LE.solids, nil, 1.0/60)
	if n := shown(); n != 2 || len(blood.drops) != 0 {
		t.Errorf("back at the newest frame: %d stains shown and %d droplets, want 2 and none: going forward sprayed or lost blood", n, len(blood.drops))
	}

	rewindCursor = mark
	restoreFrame(CL)
	rewindResumeHere()
	TimeMode = TimeOff
	updateBlood(CL, LE.solids, nil, 1.0/60)
	if blood.n != 1 {
		t.Errorf("%d stains after resuming before the blow, want only the earlier one", blood.n)
	}
	if len(blood.drops) != 0 {
		t.Error("resuming with structure back up read as a blow")
	}
}

// A fresh level has no blood.
func TestClearBloodCleansTheLevel(t *testing.T) {
	floor()
	blood.spray(rl.Vector2{}, 1, 20)
	blood.addStain(stain{})
	clearBlood()
	if blood.n != 0 || len(blood.drops) != 0 || len(blood.seen) != 0 {
		t.Errorf("after clearing: %d stains, %d droplets, %d fighters seen", blood.n, len(blood.drops), len(blood.seen))
	}
}

// A texture's mask is its painted pixels: a PNG with a transparent corner
// reads as painted everywhere but there.
func TestTextureMaskIsItsPaintedPixels(t *testing.T) {
	src := image.NewNRGBA(image.Rect(0, 0, 4, 3))
	for y := range 3 {
		for x := range 4 {
			if x < 2 && y < 2 {
				continue // the transparent corner
			}
			src.Set(x, y, color.NRGBA{200, 10, 10, 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, src); err != nil {
		t.Fatal(err)
	}
	data := buf.Bytes()
	img := rl.LoadImageFromMemory(".png", data, int32(len(data)))
	defer rl.UnloadImage(img)
	m := maskOf(img)
	for y := range 3 {
		for x := range 4 {
			if want := !(x < 2 && y < 2); m.at(x, y) != want {
				t.Errorf("pixel %d,%d painted %v, want %v", x, y, m.at(x, y), want)
			}
		}
	}
	if m.at(-1, 0) || m.at(4, 0) || m.at(0, 3) {
		t.Error("a pixel outside the texture reads as painted")
	}
}

// backdrop stands a 48x48 tile behind x from -24 to 24, over the floor, with
// the given mask, and returns it.
func backdrop(t *testing.T, painted bool) []LevelObject {
	t.Helper()
	m := &mask{w: 48, h: 48, bits: make([]uint64, 48*48/64)}
	if painted {
		for i := range m.bits {
			m.bits[i] = ^uint64(0)
		}
	}
	const name = "test/backdrop.png"
	texMasks[name] = m
	t.Cleanup(func() { delete(texMasks, name) })
	return []LevelObject{{Resource: name, X: -24, Y: 0}}
}

// splash sprays in front of the tiles, toward +X, and flies it all out.
func splash(tiles []LevelObject) {
	blood.spray(rl.Vector2{X: -10, Y: 24}, 1, 60)
	for range int((BloodLife + 1) * 60) {
		blood.fly(LE.solids, tiles, 1.0/60, 3)
	}
}

// streaks counts the stains with a tail.
func streaks() (n int) {
	for i := range blood.n {
		if s := blood.stainAt(i); s.tx != 0 || s.ty != 0 {
			n++
		}
	}
	return n
}

// In front of a painted tile, most of the blood splashes onto it as streaks
// pointing back the way they flew; the rest lands on the floor.
func TestBloodSplashesOntoTheWallBehind(t *testing.T) {
	floor()
	clearBlood()
	t.Cleanup(clearBlood)
	tiles := backdrop(t, true)
	splash(tiles)
	n := streaks()
	if n*2 <= 60 {
		t.Fatalf("%d of 60 droplets streaked the wall behind, want most", n)
	}
	back := 0
	for i := range blood.n {
		s := blood.stainAt(i)
		if s.tx == 0 && s.ty == 0 {
			continue
		}
		if s.x < -24 || s.x >= 24 || s.y < 0 || s.y >= 48 {
			t.Errorf("a streak at %d,%d, off the tile", s.x, s.y)
		}
		if l := max(iabs(int(s.tx)), iabs(int(s.ty))); l < 1 || l > BloodStreakMax {
			t.Errorf("a streak's tail runs %d px, want 1 to %d", l, BloodStreakMax)
		}
		if s.tx <= 0 {
			back++
		}
	}
	if back*2 <= n {
		t.Errorf("%d of %d streaks trail back toward the blow, want most", back, n)
	}
	if blood.n > BloodStains || len(blood.drops) != 0 {
		t.Errorf("%d stains and %d droplets left in the air", blood.n, len(blood.drops))
	}
}

// Over empty background, and over a tile's transparent pixels, nothing sticks
// in the air: the blood falls and stains the floor.
func TestBloodDoesNotHangInTheVoid(t *testing.T) {
	for name, tiles := range map[string]func(t *testing.T) []LevelObject{
		"empty background":   func(t *testing.T) []LevelObject { return nil },
		"a transparent tile": func(t *testing.T) []LevelObject { return backdrop(t, false) },
	} {
		floor()
		clearBlood()
		splash(tiles(t))
		if n := streaks(); n != 0 {
			t.Errorf("%s: %d droplets stuck in the air", name, n)
		}
		for i := range blood.n {
			if s := blood.stainAt(i); int(s.y) != floorY-1 {
				t.Errorf("%s: a stain at y=%d, want on the floor at %d", name, s.y, floorY-1)
			}
		}
		if blood.n == 0 {
			t.Errorf("%s: no blood reached the floor", name)
		}
	}
	clearBlood()
}
