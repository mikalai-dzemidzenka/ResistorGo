package resistor

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// offlineMatch installs a one-fighter offline world and restores every global
// it borrows. stepOffline drives the real loop, sampleInput included: it reads
// raylib's input state, which is all zeroes without a window and safe to call.
func offlineMatch(t *testing.T) *Actor {
	t.Helper()
	floor()
	cl, acc, pt, cam := CL, Accumulator, PhysicsTime, Cam
	t.Cleanup(func() { CL, Accumulator, PhysicsTime, Cam = cl, acc, pt, cam })

	CL = NewWorld(true)
	pl := CL.Spawn(rl.Vector2{}, true)
	CL.Local = pl
	Accumulator = 0
	return pl
}

// ticks runs exactly n fixed steps of the offline loop.
func ticks(n int) {
	for range n {
		Accumulator += FixedTimeStep * 1.0001
		stepOffline(0)
	}
}

// A death nils CL.Local through World.Remove and the respawn spawns a fighter
// nothing points at, so the camera stayed with the corpse while the player -
// whose input routes on IsPlayer, not on Local - walked away controlling
// something off screen.
func TestOfflineRespawnHandsOverTheFighter(t *testing.T) {
	pl := offlineMatch(t)

	CL.Kill(pl)
	ticks(RespawnTicks + 10)

	if CL.ByID(pl.ID) != nil {
		t.Fatalf("the dead fighter is still in the world")
	}
	if CL.Local == nil {
		t.Fatalf("no local fighter after the respawn: the camera has nothing to follow")
	}
	if CL.Local == pl {
		t.Fatalf("the local fighter is still the corpse")
	}
	if !CL.Local.IsPlayer {
		t.Errorf("the adopted fighter is not the player's")
	}
	if CL.ByID(CL.Local.ID) != CL.Local {
		t.Errorf("the local fighter is not in the world")
	}
}

// The camera follows w.Local and nothing else: game.go takes the follow branch
// only when it is non-nil, and otherwise leaves the target where it was.
func TestOfflineRespawnGivesTheCameraATarget(t *testing.T) {
	pl := offlineMatch(t)
	pl.Pos = rl.Vector2{X: 400}

	CL.Kill(pl)
	ticks(RespawnTicks + 10)

	local := CL.Local
	if local == nil {
		t.Fatalf("no local fighter to follow")
	}
	Cam.SetTarget(rl.Vector2{X: local.Pos.X + 24, Y: local.Pos.Y + 24})
	if Cam.Target.X != local.Pos.X+24 || Cam.Target.Y != local.Pos.Y+24 {
		t.Errorf("the camera target is %v, want it on the respawned fighter at %v", Cam.Target, local.Pos)
	}
}

// presentEffects keys the attack zoom and the hit shake on w.Local, so they
// went quiet after a respawn along with the camera.
func TestOfflineRespawnRestoresLocalEffects(t *testing.T) {
	pl := offlineMatch(t)

	CL.Kill(pl)
	ticks(RespawnTicks + 10)
	if CL.Local == nil {
		t.Fatalf("no local fighter after the respawn")
	}

	Cam.ZoomSpeed = 0
	CL.Local.Events |= EvAttackStart
	presentEffects(CL)
	if Cam.ZoomSpeed != AttackZoomSpeed {
		t.Errorf("the respawned fighter's attack raised the zoom by %v, want %v", Cam.ZoomSpeed, float32(AttackZoomSpeed))
	}
}

// A new match opens with the camera on the fighter, not where the last match
// left it.
func TestNewMatchSnapsTheCamera(t *testing.T) {
	pl := offlineMatch(t)
	Cam.Target = rl.Vector2{X: 5000, Y: 5000}
	Cam.Camera2D.Target = Cam.Target
	snapCamera()
	if want := camFocus(pl); Cam.Target != want || Cam.Camera2D.Target != want {
		t.Errorf("camera target %v, view %v, want both on the fighter at %v", Cam.Target, Cam.Camera2D.Target, want)
	}
}
