## 1. The two places on a fighter

- [x] 1.1 Add `Engaged [2]*Actor` to `Actor` in `actor.go` - index 0 the left
      place, index 1 the right - with a comment saying it is server-only bot
      bookkeeping and is neither replicated nor rewound. Verify `go build ./...`
      passes and `go test ./...` is still green (no behaviour change yet).
- [x] 1.2 Add `engageSide(t *Actor) int8` - which side of `t` this fighter is
      on, falling back to `Orientation` when the two share a pixel column, the
      way `pushDir` does. Verify with a unit test covering left, right and the
      same-column tie-break.
- [x] 1.3 Add `claimPlace(t *Actor) bool` and `releasePlace()` on `Actor`:
      claim takes the free place on its own side, else the free place on the far
      side, else returns false; release nils whatever the bot holds and clears
      the record of it. Both no-ops on a non-authoritative world, on an ally and
      on a dead target. Verify with unit tests: a lone bot claims, the second bot
      from the same side gets the far place, the third gets nothing, and release
      frees the place for the next claimer.
- [x] 1.4 Record the claim in `botMind` (`target *Actor`, `side int8`) and
      release it from `forget()`. Verify a test that makes a bot lose its search
      leaves both places at the old target free.

## 2. Keeping the places clean

- [x] 2.1 In `World.Remove`, nil the departing actor out of every other actor's
      `Engaged` pair and out of every bot's recorded target, in the same loop
      that already clears `Attacker`. Verify with a test that kills a slot
      holder, runs the reap, and checks the place is free and claimable.
- [x] 2.2 Verify a test that kills the fighter the places belong to leaves no
      bot recording a place at it, and that the bots go back to searching.

## 3. The line

- [x] 3.1 Add the line constants to `bot.go` beside the other feel numbers:
      `lineFront` derived from the attack box width, `AttackDash` and the
      collision box, and `lineSpacing` derived from `CollisionBox.Width`. Verify
      by a test asserting `lineFront` is beyond `inSwingRange` for two fighters
      that far apart, and `lineSpacing` is wider than a collision box.
- [x] 3.2 Add `lineRank(t *Actor, side int8) int` - how many bots on that side
      of `t` are waiting and nearer to it than this one - walking `w.Actors` and
      reading recorded targets only, with no sight calls. Verify with a unit test
      placing three waiting bots on one side and checking the ranks are 0, 1, 2
      in order of distance.
- [x] 3.3 Add `lineSpot(t *Actor, side int8, rank int) rl.Vector2`. Verify with a
      unit test that consecutive ranks are `lineSpacing` apart and that rank 0 is
      `lineFront` from the target on the given side.

## 4. Wiring it into the bot

- [x] 4.1 In `botInput`, after `seen()` returns a target: release any place held
      at a different fighter, then try to claim. A holder walks to `engageSpot`
      on its own side and swings only once `atPlace` - gating the claim alone
      left a far-side holder fighting from whichever side it arrived on. Verify
      every existing bot test still passes (`go test ./... -run Bot`).
- [x] 4.2 Add the waiting path for a bot that holds no place: walk to
      `lineSpot` with `walkStep(in, botStep(a, spot))`, face the target on
      arrival, and never press attack. Verify with a test that three bots on one
      player leave exactly two holding places, one a side, none of the others
      ever swinging, and the third standing beyond `lineFront`.
- [x] 4.3 Release the place when `seen()` returns nil, before the search
      begins. Verify with a test that walks the target out of sight and checks
      both places come free.
- [x] 4.4 Guard every claim and release with `a.W != nil && a.W.Authority`.
      Verify with a test that runs a predicting client world (`NewWorld(false)`)
      with several bots and asserts no `Engaged` entry is ever written.

## 5. Rotation

- [x] 5.1 Add `EngageHoldTicks` and `EngageYieldTicks` to `bot.go`'s feel
      constants, and a hold counter to `botMind`. Verify they are documented in
      the same voice as the neighbouring constants.
- [x] 5.2 Release the place at `EngageHoldTicks` when another bot is waiting on
      that side, arm the yield cool-down, and refuse claims while it runs.
      Verify with a test that runs two bots on one side past the hold time and
      checks the place changes hands and the former holder ends up behind the
      new one.
- [x] 5.3 Verify with a test that a holder with nobody waiting on its side keeps
      the place past `EngageHoldTicks` and goes on swinging.

## 6. What the building turned up

- [x] 6.0a Let a bot crossing to the far place, and the fighter it crosses, pass
      through one another: `Actor.crossing`, honoured by `crowding` and the push
      sweep. Without it the shove threw the target clear every time the crosser
      arrived. Verify with the three-bot test, which only puts the two engaged
      bots on opposite sides once this holds.
- [x] 6.0b Let a holder that has ended up on the target's far side take the near
      place when it is free, instead of walking back round. Verify
      `TestABotFollowsAFighterTheLengthOfTheCourse` passes - it fails without
      this, the bot turning away from what it was chasing and going blind.
- [x] 6.0c Add `stillInTheFight` and keep a waiting bot's fight while its target
      is alive and passes every sight test but the cone. The engine faces a
      fighter the way it walks, so stepping down the line turns a bot's back on
      the fight. Verify the five-bot line test holds three bots waiting instead
      of losing them to a turn-search-turn cycle.

## 7. Whole-behaviour checks

- [x] 7.1 Add `engagement_test.go` covering the spec's scenarios end to end: two
      places one a side, the second attacker crossing to the far side, nobody
      swinging from the line, the line spaced and ordered, waiting bots still
      taking hits and still drifting their guard, and no places held at allies.
      Verify `go test ./...` passes, repeatedly - bot behaviour draws on the
      global rand, so test order changes what the bots do.
- [x] 7.2 Verify the snapshot is untouched: `actorRecSize` is still 46 and the
      existing net tests pass unchanged.
- [x] 7.3 Build with `.\build.ps1`, never a plain `go build`.
- [ ] 7.4 Play a single-player match and confirm by eye that three bots give two
      attackers and a waiting line, that the line follows the player, and that
      the line advances when a holder's turn is up. Needs a window and a pair of
      eyes; the tests cover the geometry but not how it reads in play.
