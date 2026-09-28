## ADDED Requirements

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

## MODIFIED Requirements

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
