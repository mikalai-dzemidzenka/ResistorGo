package resistor

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// Every bot starts on the bot team and every human on none, and nobody without
// a team is anybody's ally - not even another human, so players still fight.
func TestTeamsByDefault(t *testing.T) {
	w := NewWorld(true)
	bot1 := w.Spawn(rl.Vector2{}, false)
	bot2 := w.Spawn(rl.Vector2{X: 100}, false)
	p1 := w.Spawn(rl.Vector2{X: 200}, true)
	p2 := w.Spawn(rl.Vector2{X: 300}, true)
	if bot1.Team != BotTeam || p1.Team != NoTeam {
		t.Fatalf("a bot starts on team %d and a human on %d, want %d and %d", bot1.Team, p1.Team, BotTeam, NoTeam)
	}
	if !bot1.allyOf(bot2) || !bot2.allyOf(bot1) {
		t.Error("two bots are not allies")
	}
	if bot1.allyOf(p1) || p1.allyOf(bot1) {
		t.Error("a bot and a human are allies")
	}
	if p1.allyOf(p2) {
		t.Error("two humans with no team are allies")
	}
}

// A bot never picks an ally to fight: with only a teammate in view it sees
// nobody, and with a teammate and a human it picks the human even when the
// teammate is nearer.
func TestABotDoesNotPickAnAlly(t *testing.T) {
	floor()
	w := NewWorld(true)
	bot := w.Spawn(rl.Vector2{}, false)
	bot.Orientation = 1
	mate := w.Spawn(rl.Vector2{X: 30}, false)
	if got := bot.seen(); got != nil {
		t.Errorf("with only a teammate in view the bot picked fighter %d", got.ID)
	}
	human := w.Spawn(rl.Vector2{X: 80}, true)
	if got := bot.seen(); got != human {
		t.Errorf("the bot picked %v, want the human over the nearer teammate", got)
	}
	_ = mate
}

// Bots on one team do not fight: left alone together for ten seconds, none of
// them is hit, stunned or dies.
func TestBotsOnOneTeamLeaveEachOtherAlone(t *testing.T) {
	floor()
	w := NewWorld(true)
	var bots []*Actor
	for _, x := range []float32{0, 40, 80} {
		b := w.Spawn(rl.Vector2{X: x}, false)
		b.Orientation = 1
		bots = append(bots, b)
	}
	for range 600 {
		w.Tick++
		for _, b := range bots {
			b.Input = botInput(b)
			b.FreezeTicks = 0
			b.Update(float32(FixedTimeStep))
		}
		for _, b := range bots {
			if b.Structure < MaxStructure || b.State == StateAttacked || b.State == StateDeath {
				t.Fatalf("a bot was hurt or stunned by its own team: structure %v, state %d", b.Structure, b.State)
			}
		}
	}
}

// A swing that reaches an ally passes through it: no hit, no block, and the
// fighter behind it is still struck.
func TestASwingPassesThroughAnAlly(t *testing.T) {
	floor()
	w := NewWorld(true)
	bot := w.Spawn(rl.Vector2{}, false)
	bot.Orientation, bot.Updown = 1, 1
	mate := w.Spawn(rl.Vector2{X: 20}, false)
	mate.Updown = 0 // a different height, so the swing would land rather than be blocked
	foe := w.Spawn(rl.Vector2{X: 24}, true)
	foe.Updown = 0
	bot.enterAttackState()
	for range 30 {
		bot.Update(float32(FixedTimeStep))
		bot.FreezeTicks = 0
	}
	if mate.Structure != MaxStructure || mate.State == StateBlock || mate.State == StateAttacked {
		t.Errorf("the swing touched the teammate: structure %v, state %d", mate.Structure, mate.State)
	}
	if foe.Structure == MaxStructure {
		t.Error("the swing did not reach the fighter behind the teammate")
	}
}

// Teammates standing in one another are left there: neither shoves, and
// neither is stunned or moved by the other.
func TestAlliesDoNotShoveEachOther(t *testing.T) {
	floor()
	w := NewWorld(true)
	a := w.Spawn(rl.Vector2{}, false)
	b := w.Spawn(rl.Vector2{X: -8}, false)
	a.Orientation, b.Orientation = 1, 1
	for _, f := range []*Actor{a, b} {
		f.checkCollision()
	}
	atX, bX := a.Pos.X, b.Pos.X
	for range pushTicks + 10 {
		run(w, 1)
		for _, f := range []*Actor{a, b} {
			if f.State == StatePush || f.State == StateAttacked {
				t.Fatalf("a teammate shoved or was shoved: state %d", f.State)
			}
		}
	}
	if a.Pos.X != atX || b.Pos.X != bX {
		t.Errorf("the teammates moved from %v and %v to %v and %v", atX, bX, a.Pos.X, b.Pos.X)
	}
}

// A bot comes back on the team it died on, whatever that was.
func TestABotRespawnsOnItsTeam(t *testing.T) {
	floor()
	w := NewWorld(true)
	b := w.Spawn(rl.Vector2{}, false)
	b.Team = 7
	w.Kill(b)
	for range RespawnTicks + 2 {
		w.Tick++
		w.updateRespawns()
	}
	if len(w.Actors) != 1 {
		t.Fatalf("%d fighters after the respawn, want 1", len(w.Actors))
	}
	if got := w.Actors[0].Team; got != 7 {
		t.Errorf("the bot came back on team %d, want 7", got)
	}
}
