## Why

Every bot that can see a fighter walks all the way in and swings. Three of them
on one player is three swings arriving at once from wherever each happened to
approach, and no amount of blocking answers that - the player is simply buried.
It also looks wrong: the bots pile into the same pixel column, and now that
teammates no longer shove one another apart they stay piled.

Beat-em-ups have solved this the same way for thirty years. Only a couple of
enemies engage; the rest hold back in a loose line, in view, visibly waiting
their turn, and step in when a slot opens. The fight stays readable, the player
can work one side at a time, and the crowd still reads as a crowd.

## What Changes

- Every living fighter owns two engagement slots, one on its left and one on its
  right. A bot must hold a slot on the target's side before it may close to
  swinging range or throw a swing at it. At most two bots fight a given fighter
  at once, one from each side.
- The slots belong to whoever is being fought, not to the human: a bot claims a
  slot on the target `seen()` picked. Each human in a multiplayer game has its
  own pair, and bots fighting bots of another team queue the same way.
- A bot with no slot takes the free slot on the far side if there is one -
  which is how the second attacker ends up behind the player rather than beside
  the first - and otherwise waits in line. The line stands on the bot's own side
  of the target, spaced a body apart, each waiting bot facing the target and
  holding its ground rather than pressing in.
- A slot holder gives it up when it dies, loses its target, or changes target,
  and when it has held the slot long enough with somebody waiting behind it: the
  holder goes to the back of its line and the next bot steps in. Without that
  rotation the line never moves and the bots at the back are scenery.
- Waiting bots still guard, still drift their guard height, and are still fought
  as anything else is: they are not invulnerable, not frozen, and a player who
  walks down the line is in reach of every bot in it.
- Nothing about this is replicated. Bot decisions are the authority's, so the
  slots live on the server and a client is only ever shown where the bots went.

Assumption recorded rather than asked: the rotation above. The request asks for
a line and a limit of two, not for the line to advance, but a line that never
advances is a queue of statues - so a holder that has had its turn steps back.
The hold time is one tuning constant and can be set high enough to disable it.

## Capabilities

### New Capabilities
- `engagement`: how many fighters may close on one fighter at once, which side
  each comes from, where the ones that may not go instead, and when a place at
  the front is given up.

### Modified Capabilities

<!-- none: no existing capability under openspec/specs/ covers who may approach
     whom. Bot sight and pathing are specified by the unarchived
     `bot-sight-and-pathing` change, whose requirements this does not alter -
     what a bot can see is unchanged, only what it does about it. -->

## Impact

`bot.go` carries it: the claim, the line spot, and the branch in `botInput`
between a slot holder (walk to its own side of the fighter and swing from
there) and a waiting bot (walk to its spot in the line and hold). `actor.go`
gains the slot pair on `Actor` and the release; `world.go`'s `Remove` clears a
dead fighter out of any slot it held or owned, the way it already clears
`Attacker`.

The shove gains a second exemption beside the one for teammates: a bot walking
round to the far side of a fighter, and that fighter, pass through one another
until the crossing is done. Without it the shove threw the target clear every
time the crosser reached it and the far place could never be taken. That
qualifies the `shove` capability, which has no spec under `openspec/specs/`
yet - it is still the unarchived `shove-crowded-fighters` change - so the
requirement is stated here, in `engagement`, where the crossing it describes
lives.

Bot tests that stand one bot against one target are unaffected - a lone bot
claims the free slot on its side and behaves exactly as it does now. Tests that
stand several bots on one target will see only two of them arrive.

Multiplayer: server-only, no new packet fields, no change to the snapshot.
Offline single-player runs the same code through `game.go`.

Rests on the working-tree change that stops teammates shoving one another: a
line of bots standing a body apart is only stable because none of them throws
the one in front of it off.
