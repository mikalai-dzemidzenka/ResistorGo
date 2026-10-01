package resistor

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// A fighter that drops off the level dies the tick it is below it, and only
// a bot in a world where bots stay dead is left out of the removal.
func TestFallingOffTheLevelKills(t *testing.T) {
	for _, stay := range []bool{false, true} {
		floor()
		w := NewWorld(true)
		w.BotsStayDead = stay
		p := w.Spawn(rl.Vector2{X: floorHi + 200}, true) // past the end of the floor
		b := w.Spawn(rl.Vector2{X: floorLo - 200}, false)
		for i := 0; i < 600 && (p.State != StateDeath || b.State != StateDeath); i++ {
			steps(p, 1, Input{})
			steps(b, 1, Input{})
		}
		if p.State != StateDeath || b.State != StateDeath {
			t.Fatalf("stay=%v: fighters below the level in states %d and %d, want dead", stay, p.State, b.State)
		}
		if !p.Dying || b.Dying == stay {
			t.Errorf("stay=%v: booked for removal: player %v, bot %v", stay, p.Dying, b.Dying)
		}
		if p.DeathTicks != DeathFlickerTicks {
			t.Errorf("stay=%v: the player's death flicker was not skipped", stay)
		}
	}
}
