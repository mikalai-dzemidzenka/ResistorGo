## Why

A bot never learns anything from being hit. Its guard height moves on a timer and nothing else: `botInput` re-rolls `Updown` on about one tick in sixty-one (`actor.go:800`), uniformly at random, whatever just happened to it. So the same attack works forever. A player who finds a height the bot is not guarding can stand there and repeat it, and the only thing that ever moves the guard off that height is the coin the bot flips against the clock.

That makes the one opponent available offline a poor teacher for the game's central read. Combat is built on three heights and a block-all, attack chains and guard changes; the fighter you practise against should occasionally answer a hit rather than ignore it.

## What Changes

- **A bot that takes a landed hit adapts its guard a third of the time.** On each hit that lands, the bot has a one-in-three chance of moving `Updown` to the height the hit came in at — the attacker's `Updown`. The other two thirds it keeps whatever guard it had.
- **It applies to a flinch and to a stun, not to a death.** The roll sits in `takeHit` past the point where a broken structure returns, so a fighter that dies to the hit does not adapt on its way out.
- **Bots only, and explicitly so.** A human's guard follows the mouse and `updateUpdown` rewrites it from the command every tick, so a player would be unaffected in practice — but the roll is gated on `!a.IsPlayer` rather than left to the cursor to undo, because an ungated roll flips a human's `Updown` for one server tick, and another attacker's hit test resolved on that tick reads `guarding := a.Updown == actor.Updown` against the wrong height.
- **The bot's existing random wander is left alone.** The one-in-sixty-one re-roll stays as it is; adaptation is added beside it, not in place of it.
- Not breaking. No wire change, no protocol version, no new replicated state: `Updown` and `PrevUpdown` are already in the snapshot, and hit resolution already runs only on the authority.

### Why this needs no prediction work

Hits are resolved inside `if a.W.Authority && !a.SwingHit` in `updateAttack` (`actor.go:1174`), so `takeHit` runs on the server and in an offline game, never in a client's predicted world. A client learns the new guard the way it learns every other consequence of a hit — from the next snapshot, where `Updown` already rides. Nothing random ever runs on both sides, so there is nothing for the two to disagree about.

## Capabilities

### New Capabilities

- `bots`: what the computer-controlled fighter does. Bot behaviour is currently unspecified anywhere — the guard wander, the approach distance and the swing odds all live only in `botInput` — so this change opens the capability with the one requirement it adds, rather than filing bot AI under `structure`, which is the rules of combat and applies to every fighter equally.

### Modified Capabilities

None. *A landed hit costs structure and flinches* keeps saying exactly what it says; adapting a guard is something one kind of fighter does in response, not a change to what the hit costs or does.

## Impact

- **Code**
  - `actor.go`: a `GuardAdaptOdds` constant beside the other combat tunables, and four lines in `takeHit` between the death return and the stun count.
- **Tests**: that a bot's guard can move to the attacker's height and that it does not always do so; that a player's guard is never touched; that a bot killed by the hit does not adapt.
- **Not affected**: the wire format, `ProtocolVersion`, prediction and reconciliation, the existing guard wander, every human-controlled fighter, and blocked attacks — `confirmBlock` is a different outcome and is out of scope here.

## Non-goals

- **Making the bot good.** This is one reaction, not an AI pass. The bot still approaches on a fixed distance, still swings on a per-tick coin flip, and still wanders its guard on a timer.
- **Adapting on a block.** A bot that guessed right could reinforce the guard that worked, which would make it stickier and harder to read. That is a separate tuning question, and doing both at once would make the effect of either impossible to judge by eye.
- **Giving a human the same treatment.** Taking the guard away from the cursor needs a real override with a lock length of its own, and it takes control off the player mid-fight. If it is ever wanted it is its own change.
- **Tuning the one-in-three.** The number is a starting point to be judged in play, not a derived value.
