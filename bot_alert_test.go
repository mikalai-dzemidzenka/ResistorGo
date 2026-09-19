package resistor

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// ambush puts a bot facing right with a fighter behind it, in reach and out of
// sight, ready to swing at its back.
func ambush(t *testing.T) (bot, attacker *Actor) {
	t.Helper()
	bot, attacker = botScene(t, row(-480, 48, 40), rl.Vector2{X: 0}, rl.Vector2{X: -26})
	bot.Orientation, bot.Updown = 1, 0
	attacker.Orientation, attacker.Updown = 1, 2 // facing the bot's back, swinging high
	if bot.canSee(attacker) {
		t.Fatal("the attacker is meant to be out of the bot's sight")
	}
	return bot, attacker
}

// Spec: "A blow from behind turns the bot round".
func TestABlowFromBehindTurnsTheBotRound(t *testing.T) {
	bot, attacker := ambush(t)
	bot.takeHit(attacker)
	if !bot.Bot.hasLast {
		t.Fatal("the bot learned nothing from being hit")
	}
	if bot.Bot.lastSeen != attacker.Pos {
		t.Errorf("the bot remembers %v, want where the blow came from, %v", bot.Bot.lastSeen, attacker.Pos)
	}
	for range 30 {
		bot.Input = botInput(bot)
		bot.Update(float32(FixedTimeStep))
		if bot.Orientation == -1 {
			return
		}
	}
	t.Error("the bot never turned toward the attack")
}

// Spec: "And then it fights".
func TestAndThenTheBotFightsBack(t *testing.T) {
	bot, attacker := ambush(t)
	bot.takeHit(attacker)
	for range 120 {
		bot.Input = botInput(bot)
		bot.Update(float32(FixedTimeStep))
		if bot.canSee(attacker) {
			return
		}
	}
	t.Errorf("the bot never came round to see its attacker: facing %d at x=%v", bot.Orientation, bot.Pos.X)
}

// Spec: "A blocked attack tells it the same".
func TestABlockedAttackAlertsTheBotToo(t *testing.T) {
	bot, attacker := ambush(t)
	bot.Updown = attacker.Updown // guarding where it is coming in
	bot.confirmBlock(attacker)
	if !bot.Bot.hasLast || bot.Bot.lastSeen != attacker.Pos {
		t.Errorf("after blocking, the bot remembers %+v, want where the attacker stood, %v", bot.Bot, attacker.Pos)
	}
}

// Spec: "A hit does not make the attacker visible".
func TestAHitDoesNotMakeTheAttackerVisible(t *testing.T) {
	bot, attacker := ambush(t)
	blowFrom := attacker.Pos
	bot.takeHit(attacker)
	attacker.Pos = rl.Vector2{X: 400} // struck and stepped away, still out of the cone
	for range 30 {
		bot.Input = botInput(bot)
		bot.Update(float32(FixedTimeStep))
	}
	if bot.Orientation != -1 {
		t.Errorf("the bot faces %d, want it looking where the blow came from", bot.Orientation)
	}
	if bot.Bot.lastSeen != blowFrom {
		t.Errorf("the bot is chasing %v, want it searching where the blow came from, %v", bot.Bot.lastSeen, blowFrom)
	}
	// and it gives that search up rather than hunting for ever
	for range SightSearchTicks + 60 {
		bot.Input = botInput(bot)
		bot.Update(float32(FixedTimeStep))
		if !bot.Bot.searching && !bot.Bot.hasLast {
			return
		}
	}
	t.Error("the bot searched for ever")
}

// Spec: "What it can see comes first".
func TestAVisibleFighterOutranksABlowFromBehind(t *testing.T) {
	bot, inView := botScene(t, row(-480, 48, 40), rl.Vector2{X: 0}, rl.Vector2{X: 150})
	bot.Orientation = 1
	behind := bot.W.Spawn(rl.Vector2{X: -26}, true)
	behind.Orientation, behind.Updown = 1, 2
	if !bot.canSee(inView) || bot.canSee(behind) {
		t.Fatal("the scene is not set up as one fighter in view and one behind")
	}
	bot.takeHit(behind)
	for range 30 {
		bot.Input = botInput(bot)
		bot.Update(float32(FixedTimeStep))
		if bot.Orientation != 1 {
			t.Fatalf("the bot turned away from the fighter it can see, to face %d", bot.Orientation)
		}
	}
	if bot.Bot.lastSeen != inView.Pos {
		t.Errorf("the bot is minding %v, want the fighter it can see at %v", bot.Bot.lastSeen, inView.Pos)
	}
}

// A human is never steered by any of it.
func TestAPlayerIsNotAlertedByBeingHit(t *testing.T) {
	player, attacker := botScene(t, row(-480, 48, 40), rl.Vector2{X: 0}, rl.Vector2{X: -26})
	player.IsPlayer = true
	attacker.Orientation, attacker.Updown = 1, 2
	player.Orientation = 1
	player.takeHit(attacker)
	if player.Bot != (botMind{}) {
		t.Errorf("a human-controlled fighter was given something to think about: %+v", player.Bot)
	}
}

// A fighter that has been broken stops being part of the fight, and that
// includes which way it is looking: the fall it plays out cannot be turned
// round by whatever its eyes would still have reported.
func TestADeadBotDoesNotTurn(t *testing.T) {
	bot, attacker := ambush(t)
	bot.takeHit(attacker) // it now remembers where the blow came from
	bot.enterDeathState(attacker)
	for range 30 {
		bot.Input = botInput(bot)
		bot.Update(float32(FixedTimeStep))
		if bot.Orientation != 1 {
			t.Fatalf("the dead bot turned to face %d", bot.Orientation)
		}
	}
}

// And it does not think either: a corpse holds no key down and its memory of
// where the fighter went stops being kept up.
func TestADeadBotDoesNotThink(t *testing.T) {
	bot, attacker := ambush(t)
	bot.takeHit(attacker)
	bot.enterDeathState(attacker)
	mind := bot.Bot
	for range 30 {
		if in := botInput(bot); in != (Input{}) {
			t.Fatalf("the dead bot asked for %+v, want no command at all", in)
		}
		bot.Update(float32(FixedTimeStep))
	}
	if bot.Bot != mind {
		t.Errorf("the dead bot went on looking: %+v, want %+v", bot.Bot, mind)
	}
}
