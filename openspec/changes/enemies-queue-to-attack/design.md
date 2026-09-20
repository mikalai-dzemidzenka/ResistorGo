## Context

See proposal.md - Why.

What the bot already is, and what this has to fit into:

- `botInput(a)` is one function producing one bot's command for one tick. It
  calls `seen()` for the nearest visible non-ally, `inSwingRange(t)` to decide
  whether to stop and swing, and `walk`/`botStep` for everything else. There is
  no world-level bot pass and no director.
- Bots are stepped one at a time, command and update together - `net_server.go`
  in its `s.Bots` loop, `game.go` in the offline actor loop. A decision an
  earlier bot makes this tick is already visible to a later one.
- Bot thinking is the authority's alone. A predicting client never calls
  `botInput`; offline play is an authoritative world (`CL = NewWorld(true)`).
- `botMind` is the per-bot memory carried between ticks, reset by `forget()`.
- `World.Remove` already clears dangling `*Actor` pointers - it walks the actors
  and nils any `Attacker` pointing at the one going away.
- `ActorRec` is a fixed 46-byte record of explicit fields, and
  `ActorRewindFrame` holds five fields of position and animation. Neither copies
  the `Actor` struct wholesale, so a new pointer field on `Actor` travels
  nowhere on its own.

## Goals / Non-Goals

**Goals:**

- One place to ask "may this bot close on that fighter", cheap enough to ask
  every tick for every bot, and stable enough not to flap between ticks.
- A bot already holding a place fights exactly as it does today, so the existing
  bot tests keep passing unchanged.
- Nothing added to the wire or to the rewind buffer.

**Non-Goals:**

- Pathing round the target to reach the far side. A bot crossing over is given
  a goal on the far side and walks at it through `botStep` like any other goal;
  if the level is in the way it stops where the level stops it, which `botStep`
  already answers. There is no route-finding.
- Any notion of a "director" tuning difficulty, deciding who attacks when, or
  coordinating the line beyond the two rules here.
- Slots for human fighters. Humans are not gated by anything.
- Formation on more than one axis. The line is along X on one side; bots on
  ledges above or below are not arranged.

## Decisions

### The two places live on the fighter they are at

`Actor` gains `Engaged [2]*Actor` - index 0 the left place, index 1 the right -
holding the bot at each, or nil. A bot claims by writing itself in; it releases
by writing nil. `botMind` records what it holds (`target *Actor`, `side int8`)
so that release is a two-line operation and so that the line can be counted
without re-running sight for every bot.

Alternatives considered:

- **Rank recomputed each tick from scratch**: for each bot, scan every other bot
  and see which ones are fighting the same fighter, then take the nearest two.
  Rejected twice over. It needs each other bot's target, which means calling
  `seen()` for bots other than the one being stepped - and `seen()` sweeps the
  level through `lineBlocked` for every candidate, so the cost goes quadratic in
  bots times solids. And a rank is not stable: two bots within a pixel of the
  same distance swap places every tick, and the one that loses the rank walks
  out of reach and back in, for ever. A held place does not flap - the holder
  keeps it until something named in the spec takes it away.
- **A `map[*Actor][2]*Actor` on `World`**: the same behaviour with worse
  company. The project keeps fighter state on the fighter, and `World.Remove` is
  already the place where pointers between actors are cleaned up.

### Claim on one's own side, cross only when that is taken

The side is `facing(a.Pos.X - t.Pos.X)`: which side of the target the bot is
already on. Claim order is own side, then far side, then nothing. Taking the far
side is what puts the second attacker behind the player, which is the shape the
request asks for; preferring one's own side first keeps a lone bot from walking
a lap round its target for no reason.

A bot on exactly the target's column has no side; it takes the one its
orientation points at, the same tie-break `pushDir` already uses for a shove
thrown at a fighter on the same pixel column.

### A holder walks to its own side, and passes through to get there

Found in the building, not in the planning: gating the claim is not enough. A
holder still walked at `t.Pos` and stopped the moment `inSwingRange` said so, so
a bot that took the far place fought from whichever side it happened to arrive
on and "one a side" was never true on the ground. Three things came out of it:

- A holder's goal is `engageSpot` - `engageStand` from the target, on its own
  side - and the swing is gated on `atPlace`, so a bot on the wrong side walks
  through instead of stopping. For a bot on its own side nothing changes: the
  swing check fires long before the goal is reached, which is why every existing
  bot test still passes.
- A crossing bot and the fighter it is crossing pass through one another
  (`Actor.crossing`, honoured by `crowding` and the push sweep). Without it the
  shove threw the target further off every time the crosser reached it, and
  since the crosser walks *at* that target, it spent the match chasing the
  fighter it was trying to get behind. This is a second exemption in the shove,
  after the one for teammates, and it is the reason the far place is reachable
  at all.
- A holder that finds itself on the far side with the near place free takes the
  near one instead of walking back. `TestABotFollowsAFighterTheLengthOfTheCourse`
  caught this: the lunge carries a bot through its target constantly, and one
  that walked back to its claimed side turned round, lost sight of what it was
  chasing and never got it back.

### A bot stepping into line does not forget the fight

Also found in the building. `updateHorSpeed` faces a fighter the way it walks -
there is no walking backwards in this engine - so a bot sent to a spot further
down the line turns its back on the fight, leaves the sight cone, and with
nothing else said forgets the fight and sets off to search the ground it is
standing on. Turn, see, step back, turn again, for the rest of the match.

`stillInTheFight` is the answer: a waiting bot keeps its fight while the target
is alive and passes every sight test except the cone. Losing it needs the
target to actually go - out of range, behind something solid, or down.

Alternative considered: have the waiting bot walk backwards while still facing
the fight, the way a mook in a beat-em-up backs off. Rejected - `updateHorSpeed`
turns a fighter on the movement key itself, so it would mean a second way to
move, and turning already costs two thirds of the speed by design.

### Waiting is a goal, not a new state

No new `State`, no new animation, nothing replicated. A waiting bot is handed a
different goal - its spot in the line - and walks to it with the same
`walkStep(in, botStep(a, spot))` every other bot walk goes through. Arrival,
braking, holes, walls and giving up are all already `botStep`'s answers, which
is why "the line is not a wall" in the spec costs nothing to satisfy.

The spot is `t.Pos.X + side*(lineFront + rank*lineSpacing)`:

- `lineFront` - the front of the line, just past the reach of a swing thrown by
  or at the target, derived from the attack box width (24), the dash that comes
  with a swing (`AttackDash`) and the collision box, not typed in as a number,
  so retuning the arm carries the line with it.
- `lineSpacing` - a body and a gap, derived from `CollisionBox.Width` (14). Wide
  enough that two waiting bots do not overlap, which matters now that teammates
  no longer shove one another apart.
- `rank` - how many bots on this side of this target are waiting and nearer to
  it than this one. Counted by walking `w.Actors` and reading each bot's
  recorded `Bot.target` and position: no sight calls, no allocation.

`botArrived` already gives the spot an 8-pixel dead zone, so a bot at its spot
stands rather than shuffling.

### Rotation is a hold counter plus a cool-down

`botMind` counts ticks held. At `EngageHoldTicks`, if any bot is waiting on that
side, the holder releases. It cannot simply re-claim on its next tick - it is
usually the nearest bot and would take the place straight back - so releasing
also arms a short cool-down (`EngageYieldTicks`) during which the bot claims
nothing and walks to a line spot. That is long enough for the waiting bot, which
is stepped in the same loop, to take the place.

Alternative considered: give the place to the nearest waiting bot directly, at
the moment of release. Rejected - it makes one bot write another bot's mind, and
the cool-down gets the same result with a counter.

Set `EngageHoldTicks` beyond any plausible fight and the rotation is off; the
rest of the design does not depend on it.

### Release is centralised, and `World.Remove` owns the dangling half

Releases inside `botInput` cover losing the target, changing target and
yielding. The two that happen outside it - the holder dying, and the fighter the
places belong to dying - are handled where `Attacker` already is, in
`World.Remove`: nil the leaving actor out of every `Engaged` pair, and nil every
bot's `Bot.target` that points at it. A bot left holding a place at a corpse
would gate a slot that nobody can ever free.

`forget()` gains the release too, so the existing "gave up the search" path
cannot leak a place.

### Claims are guarded by `W.Authority`

`botInput` is only called on authoritative worlds today, but the claim writes
to another actor, so the guard is stated rather than assumed - the same guard
the shove's displacement uses.

## Risks / Trade-offs

- **The far side is unreachable** (a pit, a wall, the target against the edge of
  the level) → the bot claims the far place, walks at it, and `botStep` stops it
  where the level does. The place is held by a bot that cannot arrive, so one of
  the two attack slots is wasted. Mitigated by the rotation: the hold counter
  runs whether or not the holder ever got there, so the place comes free and
  another bot gets a turn. Not solved in general - that wants pathing, which is
  a non-goal.
- **Two bots arriving on the same tick** → they are stepped in sequence, so the
  first to be stepped claims and the second sees the place taken. Stable, but
  order-dependent: the winner is the one earlier in `s.Bots`/`w.Actors`. That
  order is fixed and identical on server and offline play, and nothing is
  predicted on a client, so it cannot desync.
- **A line spot inside a wall or over a hole** → `botStep` brakes short and the
  line bunches up there. Ugly, not broken, and the same thing the bot already
  does with any goal it cannot reach.
- **Existing bot tests** → any test that stands several bots on one target and
  expects all of them to swing will now see two. `TestBotsOnOneTeamLeaveEachOtherAlone`
  is unaffected (allies take no places). Single-bot tests are unaffected by
  construction.
- **A waiting bot is an easy target** → deliberate, and the reason the rotation
  exists. It is also how the player gets to pick the crowd apart, which is the
  point of the line.
- **A crossing bot overlaps the fighter it is crossing** → for the few ticks it
  takes to walk through. Two sprites briefly on the same column, in a game with
  no depth to walk round in. The shove parts them the tick the crossing ends,
  which is what keeps it to a few ticks.
- **A waiting bot walking to a spot further out has its back to the fight** →
  it can be hit from behind while it does, which `alerted` already answers as it
  does for any other blow from out of view. The alternative was a bot that could
  never join the line at all.

## Migration Plan

Not applicable: no persisted data, no wire format change, no version gate.
Server and client of different builds interoperate because nothing here is sent.
Rollback is reverting the change.
