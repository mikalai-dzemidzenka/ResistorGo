package resistor

import (
	"bytes"
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// What a bot has in mind is the authority's business alone. Spec: "Sight and
// movement decisions belong to the authority" - nothing new on the wire.
func TestABotsMindNeverReachesTheWire(t *testing.T) {
	bot, target := botScene(t, row(-480, 48, 40), rl.Vector2{X: 0}, rl.Vector2{X: 200})
	bot.Input = botInput(bot)
	blank := encodeSnapshot(Snapshot{Tick: 7, Actors: []ActorRec{recOf(bot)}})

	// the same fighter, mid-search, at the same position and in the same state
	hide(target)
	for range 30 {
		saved := bot.Pos
		bot.Input = botInput(bot)
		bot.Pos = saved // hold it still: only the mind is allowed to differ
	}
	if !bot.Bot.searching {
		t.Fatal("the bot is not searching, so there is nothing to leave out of the snapshot")
	}
	searching := encodeSnapshot(Snapshot{Tick: 7, Actors: []ActorRec{recOf(bot)}})

	if !bytes.Equal(blank, searching) {
		t.Errorf("a searching bot serialises differently:\n blank     %v\n searching %v", blank, searching)
	}
}

// A client predicts its own fighter and nothing else: no bot decision is ever
// taken in a world that is not the authority.
func TestAClientRunsNoBotDecisions(t *testing.T) {
	floor()
	cl := NewWorld(false) // a client's predicted world
	bot := cl.Spawn(rl.Vector2{X: 200}, false)
	cl.Local = cl.Spawn(rl.Vector2{}, true)
	before := bot.Pos

	// The client's step is the one in net_client.go: it advances its own
	// fighter from its own input and leaves every other actor to the snapshots.
	for range 120 {
		cl.Tick++
		cl.Local.Update(float32(FixedTimeStep))
	}
	if bot.Pos != before {
		t.Errorf("a bot moved in a predicted world, from %v to %v", before, bot.Pos)
	}
	if bot.Bot.hasLast || bot.Bot.searching {
		t.Errorf("a bot in a predicted world formed an opinion: %+v", bot.Bot)
	}
}
