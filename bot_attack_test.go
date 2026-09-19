package resistor

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// fight runs the bot against the fighter it is facing for n ticks and reports
// how many swings it opened and how many blows of them reached the other
// fighter. The target guards with block-all, so every blow that arrives is
// answered the same way and nothing about the count depends on which height the
// bot happens to be swinging at.
//
// The hit test runs inside the swinging fighter's own update, so blows resolve
// on the target without it acting; it is still updated, because the lunge of
// every swing walks the bot into it and the bot shoves it off, and a target
// never updated would never come out of the stun that leaves it in. Guarding
// everything, it never shoves back.
func fight(bot, target *Actor, n int) (swings, landed int) {
	target.Updown, target.Input.Updown = 3, 3
	was, structure := bot.State, target.Structure
	for range n {
		bot.Input = botInput(bot)
		bot.Update(float32(FixedTimeStep))
		target.FreezeTicks = 0
		target.Update(float32(FixedTimeStep))
		if bot.State == StateAttack && was != StateAttack {
			swings++
		}
		if bot.AttackCount > 0 { // a chained blow, counted where it is thrown
			swings += int(bot.AttackCount)
			bot.AttackCount = 0
		}
		if target.Structure < structure {
			landed++
			structure = target.Structure
			target.Structure = MaxStructure // it is here to be swung at, not to die
			structure = MaxStructure
		}
		was = bot.State
	}
	return swings, landed
}

// A bot with a fighter in reach presses the attack instead of waiting on a coin
// flip to let it, and the blows land: the reach it stops at is the reach of its
// arm.
//
// The floor is what the model it replaced could manage at its very best - a
// swing a second, none of them chained, and that from a distance it could not
// actually cross - so a bot that fell back to waiting on the dice would fail
// this. The pool is what holds the rate down over a window this long: swinging
// stops the refill, so the ten seconds are bursts with gaps, not one long chain.
func TestBotInReachKeepsAttacking(t *testing.T) {
	bot, target := botScene(t, nil, rl.Vector2{}, rl.Vector2{X: 20})
	swings, landed := fight(bot, target, 600)
	t.Logf("%d swings, %d of them landed, in ten seconds", swings, landed)
	if swings < 12 {
		t.Errorf("%d swings in ten seconds, want at least 12", swings)
	}
	if landed < 10 {
		t.Errorf("%d of them reached the target, want at least 10", landed)
	}
}

// The distance a bot stops at is the distance its swing crosses. Measured
// against the hit test itself: wherever the bot believes it is in range, a
// swing thrown there reaches the other fighter's collision box.
//
// Swept both ways round. The reach used to be three pixels longer facing right
// than facing left, so a sweep on one side alone could not have seen it.
func TestBotStopsWhereItsSwingLands(t *testing.T) {
	var reachRight, reachLeft int
	for dx := float32(-80); dx <= 80; dx++ {
		bot, target := botScene(t, nil, rl.Vector2{}, rl.Vector2{X: dx})
		// the swing inSwingRange measures is the one thrown at the target, so
		// the bot is turned to it rather than left facing right
		bot.Orientation = facing(dx)
		if !bot.inSwingRange(target) {
			continue
		}
		if dx > 0 {
			reachRight++
		} else if dx < 0 {
			reachLeft++
		}
		box, ok := bot.attackBoxAt(bot.Updown, bot.Pos)
		if !ok {
			t.Fatalf("dx=%v: the bot has no attack box at guard height %d", dx, bot.Updown)
		}
		box.X += float32(bot.Orientation) * AttackDash // the dash the swing takes
		if !aabb(box, target.collisionRect()) {
			t.Errorf("dx=%v: the bot would swing here, and the swing does not reach", dx)
		}
	}
	if reachRight != reachLeft {
		t.Errorf("the bot reaches %d pixels ahead of it and %d behind - the two should mirror", reachRight, reachLeft)
	}

	// and the distance it used to stop at is not one of them
	bot, target := botScene(t, nil, rl.Vector2{}, rl.Vector2{X: 50})
	if bot.inSwingRange(target) {
		t.Error("the bot still thinks it can reach a fighter 50 pixels away")
	}
}

// Out of reach it walks rather than swinging at the air.
func TestBotOutOfReachDoesNotSwing(t *testing.T) {
	bot, target := botScene(t, nil, rl.Vector2{}, rl.Vector2{X: 120})
	for range 30 {
		if bot.inSwingRange(target) {
			break
		}
		bot.Input = botInput(bot)
		if bot.Input.Attack {
			t.Fatalf("the bot pressed the attack from %v pixels away", target.Pos.X-bot.Pos.X)
		}
		bot.Update(float32(FixedTimeStep))
	}
}

// A bot presses on into the chain, so a swing is followed by the next blow
// rather than by the recovery every time.
func TestBotChainsItsSwings(t *testing.T) {
	bot, target := botScene(t, nil, rl.Vector2{}, rl.Vector2{X: 20})
	target.Updown = 3
	chained := false
	for range 600 {
		bot.Input = botInput(bot)
		bot.Update(float32(FixedTimeStep))
		if bot.AttackCount > 0 {
			chained = true
			break
		}
		target.Structure = MaxStructure
	}
	if !chained {
		t.Error("the bot never followed a swing with the next blow of the chain")
	}
}

// It keeps a reserve. Below what it takes to open an attack it holds off, and
// with the pool back above it presses again - so the pressure comes in bursts
// and a bot is never left swinging at an empty pool.
func TestBotKeepsAnEnergyReserve(t *testing.T) {
	bot, target := botScene(t, nil, rl.Vector2{}, rl.Vector2{X: 20})
	bot.checkCollision() // stand it on the floor: a bot in the air throws nothing
	if !bot.OnGround || !bot.inSwingRange(target) {
		t.Fatalf("the scene is meant to put the target in reach of a standing bot: ground=%t reach=%t", bot.OnGround, bot.inSwingRange(target))
	}
	bot.Energy = BotOpenSwings*EnergySwing - 1
	for range 60 {
		bot.Input = botInput(bot)
		if bot.Input.Attack {
			t.Fatalf("the bot opened an attack with %d energy, its reserve being %d", bot.Energy, BotOpenSwings*EnergySwing)
		}
		bot.Bot.swingWait = 0 // the wait is not what is being measured here
	}
	bot.Energy = MaxEnergy
	pressed := false
	for range 60 {
		bot.Input = botInput(bot)
		if bot.Input.Attack {
			pressed = true
			break
		}
	}
	if !pressed {
		t.Error("with a full pool the bot still would not open an attack")
	}
}

// The guard drifts while the bot is not swinging, and holds still while it is:
// changing height mid-chain throws a transition animation that costs more than
// the change is worth.
func TestBotGuardHoldsStillMidSwing(t *testing.T) {
	bot, _ := botScene(t, nil, rl.Vector2{}, rl.Vector2{X: 20})
	bot.State, bot.Updown = StateAttack, 1
	for range 2000 {
		if in := botInput(bot); in.Updown != bot.Updown {
			t.Fatalf("the bot moved its guard to %d mid-swing", in.Updown)
		}
	}
	bot.State = StateIdle
	drifted := false
	for range 2000 {
		if in := botInput(bot); in.Updown != bot.Updown {
			drifted = true
			break
		}
	}
	if !drifted {
		t.Error("the guard never drifted while the bot was standing")
	}
}
