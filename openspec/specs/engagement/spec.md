# engagement Specification

## Purpose

Limits how many bots may fight one fighter at once and says where the rest go:
two engage, one from each side, and everyone else holds back in a line on its
own side, stepping in as places at the front are given up.

## Requirements

### Requirement: Two places at a fighter, one a side

Every living fighter SHALL have exactly two places at it, one on its left and
one on its right. A bot SHALL hold a place at a fighter before it closes on that
fighter, and each place SHALL be held by at most one bot. At most two bots SHALL
be engaged with any one fighter at a time.

#### Scenario: Two engage

- **WHEN** three or more bots are fighting one fighter
- **THEN** no more than two of them SHALL hold a place at it, and no bot
  without one SHALL swing at it

Being within reach is not the same as engaging: a fighter can be shoved into
the line, and the bot it lands next to has no say in that. What is bounded is
who may close and who may swing.

#### Scenario: One a side

- **WHEN** two bots are engaged with one fighter
- **THEN** one SHALL be on its left and the other on its right

#### Scenario: A lone bot engages at once

- **WHEN** one bot has a fighter in view and no other bot is fighting it
- **THEN** that bot SHALL take a place and close as it would with no line at all

#### Scenario: Each fighter has its own pair

- **WHEN** bots are fighting two different fighters
- **THEN** the places held at one SHALL NOT limit the places available at the
  other

### Requirement: Only a bot with a place fights

A bot SHALL throw a swing at a fighter only while holding a place at it. A bot
without a place SHALL NOT walk into the reach of a swing at its target of its
own accord.

#### Scenario: No place, no swing

- **WHEN** a bot has a fighter in view and holds no place at it
- **THEN** it SHALL NOT swing at that fighter

#### Scenario: No place, no closing

- **WHEN** a bot without a place is walking toward its target
- **THEN** it SHALL stop short of the reach of a swing at that target

#### Scenario: Taking a place lets it in

- **WHEN** a waiting bot takes a place that has been given up
- **THEN** it SHALL close on the target and fight it as the previous holder did

### Requirement: The crowd swings in turn

At most one attack SHALL be in the air at any one fighter at a time. A bot
SHALL NOT open an attack on a fighter while another bot in that fight is
attacking it, and SHALL NOT open one for a bounded beat after that attack ends.
A bot fighting a fighter alone SHALL be held back by none of this.

A fighter guards at one height, and two blows arriving together can be answered
only by guarding everything, which is not a fight. Spacing the two attackers
apart makes that rare; taking turns makes it impossible.

#### Scenario: One attack in the air

- **WHEN** two bots are engaged with one fighter and one of them is attacking it
- **THEN** the other SHALL NOT open an attack on it

#### Scenario: The beat after the blow

- **WHEN** a bot's attack on a fighter ends
- **THEN** no other bot in that fight SHALL open an attack on that fighter until
  the beat has run out

#### Scenario: A chain is one turn

- **WHEN** a bot carries its attack on into a further blow
- **THEN** that blow SHALL NOT wait for a turn of its own

#### Scenario: A lone attacker waits for nobody

- **WHEN** a bot is the only one fighting a fighter
- **THEN** it SHALL open attacks as it would with no crowd at all

#### Scenario: Waiting for a turn costs no more than the turn

- **WHEN** a bot is refused a turn
- **THEN** it SHALL open its attack as soon as the fighter comes free, rather
  than beginning its own gap between attacks over again

### Requirement: A bot crosses to the free side

A bot that wants a place SHALL take the one on the side of the target it is
already on when that place is free, and SHALL take the place on the far side
when its own side is taken and the far side is free. A bot SHALL NOT cross to
a place that is closed, and SHALL NOT cross at all while either of the
target's places is closed (see "A landed roll closes the place ahead of it").

#### Scenario: Its own side first

- **WHEN** a bot approaches a target and the place on its side is free
- **THEN** it SHALL take that place and SHALL NOT cross over

#### Scenario: The second attacker comes round

- **WHEN** a bot approaches a target from the side another bot already holds,
  and the far side is free and neither place is closed
- **THEN** it SHALL take the far place and SHALL end up on the opposite side of
  the target from the other bot

#### Scenario: Both taken

- **WHEN** a bot approaches a target and both places are held
- **THEN** it SHALL hold no place and SHALL wait

#### Scenario: A bot crossing over passes through

- **WHEN** a bot holding the place on a target's far side walks through that
  target to reach it
- **THEN** neither SHALL shove the other until the bot is across

#### Scenario: Overshooting takes the near place instead

- **WHEN** a bot holding a place ends up on the target's other side - a swing's
  lunge carries one through all the time - and the place on the side it now
  stands is free
- **THEN** it SHALL take that place rather than walk back round

#### Scenario: No crossing while a place is closed

- **WHEN** one of a target's places is closed and a bot holds the open place
  from the closed side
- **THEN** it SHALL give that place up and wait in the line on its own side
  rather than cross through the target

### Requirement: A bot never turns its back on a fight to hold its place

Walking into or along the line SHALL NOT cost a bot the fight it is waiting on.
A bot whose target has left the cone it looks down, but is alive and as near
and as unobstructed as sight otherwise asks, SHALL go on waiting in the line
rather than give the fight up and search for it.

#### Scenario: Stepping back is not losing it

- **WHEN** a waiting bot turns away from its target to walk to a spot further
  down the line
- **THEN** it SHALL keep waiting on that target, and SHALL NOT begin searching
  for it

#### Scenario: Out of range is still lost

- **WHEN** a waiting bot's target moves beyond sight range, or behind
  something solid
- **THEN** the bot SHALL give up its place and the fight

### Requirement: Bots with no place wait in a line

A bot that wants a place and has none SHALL walk to a spot on its own side of
the target, beyond the reach of a swing at that target, and SHALL hold there
facing the target. Waiting bots on one side SHALL stand at separate spots, each
at least a body further from the target than the one ahead of it, ordered so
that the nearest waiting bot stands at the front of the line.

#### Scenario: A line, not a pile

- **WHEN** two or more bots are waiting on the same side of a target
- **THEN** their bodies SHALL NOT overlap, and each SHALL stand further from the
  target than the one ahead of it

#### Scenario: Waiting is holding, not pressing

- **WHEN** a waiting bot has reached its spot in the line
- **THEN** it SHALL hold that spot and SHALL NOT walk further toward the target

#### Scenario: Facing the fight

- **WHEN** a bot waiting in line has reached its spot
- **THEN** it SHALL face the target

A bot still walking to its spot faces the way it is walking, which the engine
decides and not this capability.

#### Scenario: The line follows the fight

- **WHEN** the target moves
- **THEN** the waiting bots SHALL walk to keep their spots in the line relative
  to it, as far as the level lets them

#### Scenario: The line is not a wall

- **WHEN** a waiting bot cannot reach its spot - a hole, a wall or another
  fighter in the way
- **THEN** it SHALL come to rest where the level allows and SHALL NOT press into
  what stops it

### Requirement: A place is given up

A bot SHALL give up its place when it dies, when the fighter it held the place
at dies, when it loses its target, and when it takes a different target. Both a
death's releases SHALL happen as the fighter falls, not when its body leaves the
world: a body that stays on the floor SHALL hold no place and SHALL have none
held at it. A bot
SHALL also give up its place after holding it for a bounded time while another
bot is waiting on that side, and SHALL then go to the back of that line.

#### Scenario: Death frees the place

- **WHEN** a bot holding a place dies
- **THEN** that place SHALL become free, and a waiting bot SHALL be able to take
  it

#### Scenario: The fight ending frees the places

- **WHEN** the fighter the places belong to dies
- **THEN** no bot SHALL be left holding a place at it

#### Scenario: A body left lying holds nothing

- **WHEN** a bot dies in a match that leaves the bodies on the floor
- **THEN** the place it held SHALL come free for the bots waiting behind it, and
  no bot SHALL be left holding a place at the body

#### Scenario: Losing sight frees the place

- **WHEN** a bot holding a place no longer has that fighter in view and goes off
  to search for it
- **THEN** it SHALL give the place up

#### Scenario: Turning on someone else frees the place

- **WHEN** a bot holding a place at one fighter takes another fighter as its
  target
- **THEN** it SHALL give up the first place before taking one at the second

#### Scenario: The line advances

- **WHEN** a bot has held a place for the full hold time and another bot is
  waiting on that side
- **THEN** the holder SHALL give the place up, the waiting bot SHALL take it,
  and the former holder SHALL take a spot in the line behind it

#### Scenario: Nobody waiting, nobody moves

- **WHEN** a bot has held a place for the full hold time and no other bot is
  waiting on that side
- **THEN** it SHALL keep the place and go on fighting

### Requirement: A waiting bot is still a fighter

Waiting in line SHALL change nothing about a bot but where it walks and whether
it swings. A waiting bot SHALL be hit, shoved, stunned and killed as any fighter
is, SHALL go on guarding and drifting its guard height, and SHALL be turned
toward an attacker that strikes it as it would be with no line at all.

#### Scenario: Waiting is not invulnerable

- **WHEN** a swing reaches a bot waiting in line
- **THEN** it SHALL take the hit as any fighter does

#### Scenario: Struck from the line

- **WHEN** a bot waiting in line is hit by a fighter it cannot see
- **THEN** it SHALL react to that attack as a bot does to any attack from out of
  view

#### Scenario: The guard still moves

- **WHEN** a bot is waiting in line
- **THEN** its guard height SHALL go on drifting

### Requirement: Places are never held at an ally or the dead

A bot SHALL NOT hold a place at a fighter on its own team, and SHALL NOT hold
one at a dead fighter. Places SHALL be counted only among bots; a human fighter
SHALL take no place and SHALL be limited by none.

#### Scenario: No line at an ally

- **WHEN** a bot is standing near a fighter on its own team
- **THEN** it SHALL hold no place at it and SHALL form no line at it

#### Scenario: Humans are not queued

- **WHEN** a human fighter walks up to a fighter that already has both places
  held
- **THEN** that human SHALL be free to close and swing

### Requirement: The places are the authority's

Places SHALL be claimed, held and given up only in the authoritative
simulation. A predicting client SHALL claim no place and SHALL be shown where
the bots went by the snapshot alone. No part of a place SHALL be sent over the
wire.

#### Scenario: A client queues nobody

- **WHEN** a predicting client simulates a world with several bots and one
  fighter
- **THEN** no place SHALL be claimed on that client

#### Scenario: Nothing new on the wire

- **WHEN** a snapshot is encoded
- **THEN** it SHALL carry no place, side or line position, and the replicated
  actor record SHALL be the same size and shape as it is for a world in which
  no bot holds a place

### Requirement: A landed roll closes the place ahead of it

When a fighter lands a roll, the place on the side it rolled toward SHALL be
closed for 2 seconds. While a place is closed no bot SHALL hold it: a bot
holding it SHALL give it up and wait in the line on the side it stands, and no
bot SHALL take it. A bot SHALL only take the open place from the side that
place is on, so no bot crosses round the fighter while one of its places is
closed. The side that closes SHALL be the one the roll was travelling toward
when it landed, and SHALL NOT change for the rest of the 2 seconds, whichever
way the fighter turns. When the 2 seconds are up both places SHALL be open
again and the usual rules SHALL apply.

#### Scenario: Rolled out of a surround

- **WHEN** a fighter held at from both sides rolls out over one of the two bots
  and lands
- **THEN** for the next 2 seconds it SHALL be fought from the side it rolled
  away from only, and no bot SHALL come round to the side it rolled toward

#### Scenario: The holder of the closed place stands back

- **WHEN** a place closes while a bot holds it
- **THEN** that bot SHALL give it up and SHALL wait in the line rather than
  fight

#### Scenario: A bot ahead of the roll waits

- **WHEN** a place is closed and a bot approaches the fighter from the closed
  side
- **THEN** it SHALL take no place and SHALL wait in the line on that side, even
  when the open place is free

#### Scenario: Turning round does not move the closed side

- **WHEN** a fighter that has just landed a roll turns to face the other way
  within the 2 seconds
- **THEN** the closed place SHALL stay the one on the side it rolled toward

#### Scenario: The surround comes back after 2 seconds

- **WHEN** 2 seconds have passed since the roll landed
- **THEN** both places SHALL be open, and a bot SHALL again be able to cross to
  the far side when its own is taken

#### Scenario: A roll that does not land closes nothing

- **WHEN** a roll is interrupted by a stun or a death, or ends by catching a
  ledge
- **THEN** no place SHALL close
