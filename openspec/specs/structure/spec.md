# structure Specification

## Purpose

Makes a fight accumulate: hits, blocks and time all move one pool that decides whether a fighter merely flinches, is stunned, or goes down.

## Requirements

### Requirement: Structure pool
Every fighter SHALL have a structure pool with a capacity of 12 hit-units, full when the fighter enters the world. Structure SHALL never exceed that capacity or fall below zero. It is the fighter's only health model: nothing else tracks how close a fighter is to defeat.

#### Scenario: Full at spawn
- **WHEN** a fighter spawns, whether at the start of a match or on respawn
- **THEN** its structure is 12

#### Scenario: Restoring cannot overfill
- **WHEN** a fighter at full structure blocks an attack
- **THEN** its structure is still 12

### Requirement: A landed hit costs structure and flinches
A hit that lands SHALL cost the defender 1 structure and SHALL produce a flinch: the hit flash, the hit-stop freeze that both fighters already share, a small knockback and the hit animation. A flinch SHALL NOT lock the defender out of acting — it can move, attack, block and jump during one — and SHALL NOT interrupt an action the defender had already started.

The knockback SHALL move the fighter, not merely be given to it: after a flinch the fighter SHALL be somewhere it was not when the hit landed. It SHALL remain clearly smaller than the throw a stun gives, so the two read as different events.

#### Scenario: First hit does not stun
- **WHEN** a fighter at full structure takes one hit
- **THEN** its structure is 11, it flashes and is knocked back slightly, and it can act on the next tick

#### Scenario: The small knockback actually moves the fighter
- **WHEN** a fighter is flinched and then left alone until the knockback has run out
- **THEN** its position differs from where it was when the hit landed

#### Scenario: A flinch stays smaller than a stun
- **WHEN** one fighter is flinched and another stunned by attackers facing the same way, and both are left alone until their knockbacks have run out
- **THEN** the flinched one has travelled a visibly shorter distance than the stunned one

#### Scenario: Flinch does not interrupt a swing
- **WHEN** a fighter is hit while its own attack animation is playing
- **THEN** that attack continues and can still land

### Requirement: Every third landed hit stuns
The third hit to land on a fighter SHALL stun it: the defender enters the hit-stun state, is locked out until the stun animation finishes, and is thrown back noticeably further than a flinch throws it. The count SHALL be per fighter and SHALL reset when the stun lands, when the fighter dies, and when its structure begins regenerating. A block SHALL NOT advance it.

#### Scenario: Third hit of a run
- **WHEN** a fighter takes three hits without its structure regenerating in between
- **THEN** the first two flinch it and the third stuns it, locking it out and throwing it back further

#### Scenario: The count restarts after a stun
- **WHEN** a fighter has just been stunned and takes another hit
- **THEN** that hit flinches it, and the stun comes two hits later

#### Scenario: Blocks do not count
- **WHEN** a fighter takes a hit, blocks two attacks, then takes another hit
- **THEN** neither hit stuns it

### Requirement: One landing per swing
A single swing SHALL land on a given fighter at most once, whether it hits or is blocked. A swing whose attack box stays over a defender for several ticks SHALL NOT cost structure more than once.

#### Scenario: Overlapping boxes do not drain structure
- **WHEN** an attack box overlaps a defender for the whole of a swing
- **THEN** the defender loses structure once for that swing

#### Scenario: The next swing lands again
- **WHEN** an attacker chains a second swing into the same defender
- **THEN** that second swing costs the defender structure of its own

### Requirement: Blocking and structure
A block that holds until the swing's hit frame SHALL restore 0.5 structure to the defender, except when the defender is guarding with block-all (guard height 3), which SHALL instead cost 0.4 structure — less than the 1 a landed hit costs. A block SHALL NOT flinch or stun the defender.

#### Scenario: Matching guard restores
- **WHEN** a fighter at 8 structure blocks an attack with a guard at the attacker's height
- **THEN** its structure is 8.5 and it is not flinched

#### Scenario: Block-all costs, but less than a hit
- **WHEN** a fighter at 8 structure blocks an attack while guarding with block-all
- **THEN** its structure is 7.6, which is more structure than the 7 it would have had if the attack had landed

#### Scenario: A guard that drops mid-swing is a hit
- **WHEN** a fighter starts blocking but changes guard height before the swing's hit frame, so that it no longer matches and is not block-all
- **THEN** the attack lands as a hit, with a hit's structure cost, instead of restoring any

### Requirement: Structure regenerates after a pause
When a fighter has not lost structure for 1.5 seconds, its structure SHALL refill at 1 per second up to capacity. Losing structure SHALL restart that wait.

#### Scenario: Recovering between exchanges
- **WHEN** a fighter that has lost 4 structure is left alone for 5.5 seconds
- **THEN** its structure has refilled by 4 and it is back at full

#### Scenario: Half recovered
- **WHEN** a fighter that has lost 4 structure is left alone for 3.5 seconds, which once refilled all of it
- **THEN** its structure has refilled by 2, and it is still 2 short of full

#### Scenario: Pressure prevents recovery
- **WHEN** a fighter is hit once per second
- **THEN** its structure never regenerates, and it is stunned on every third hit

### Requirement: Breaking structure kills
Structure reaching zero SHALL put the fighter into a Death state, whether the last cost came from a hit or from block-all. In the Death state the fighter SHALL play the death animation once, SHALL NOT act, SHALL NOT be hit or blocked by anyone, and SHALL NOT take further structure changes. When the animation finishes the fighter SHALL flicker for about 0.6 seconds and then be removed from the world, except a bot in a single-player match, which SHALL stay lying where it fell, visible and without flickering, for the rest of the match.

#### Scenario: Twelve hits
- **WHEN** a fighter takes 12 hits with no regeneration in between
- **THEN** it is stunned on the third, sixth and ninth, and the twelfth breaks its structure and puts it into the Death state

#### Scenario: Broken by block-all
- **WHEN** a fighter guarding with block-all takes the attack that drops its structure to zero
- **THEN** it enters the Death state, the same as if the attack had landed

#### Scenario: A dying fighter is out of the fight
- **WHEN** an attacker swings at a fighter in the Death state
- **THEN** nothing is hit or blocked and the dying fighter's animation is undisturbed

#### Scenario: Flicker then gone
- **WHEN** the death animation finishes
- **THEN** the fighter flickers in place for about 0.6 seconds and is then no longer in the world

#### Scenario: Single-player bodies stay
- **WHEN** a bot's death animation finishes in a single-player match
- **THEN** the body stays where it fell, does not flicker and is never removed

### Requirement: Respawning
A fighter removed by death SHALL return 2 seconds later at a spawn position not occupied by another living fighter's collision box, with full structure and none of its old state. A player whose fighter died SHALL be given the new fighter and control it; a bot SHALL return as a bot. A bot that dies in a single-player match is never removed, and so never returns.

#### Scenario: Bot returns
- **WHEN** a bot in a networked match dies
- **THEN** 2 seconds after it disappears a bot is fighting again, with full structure

#### Scenario: Player returns
- **WHEN** a player's fighter dies in a networked match
- **THEN** the player has no fighter while it is dead, and 2 seconds after it disappears they control a new fighter at a free spawn position

#### Scenario: Nothing carries over
- **WHEN** a fighter that died while stunned and mid-attack respawns
- **THEN** it is idle, at full structure, with its hit count back to zero

### Requirement: What a dying fighter's own client shows
A client SHALL show its own fighter dying exactly as it shows anyone else's: while the authority is still reporting that fighter, the client SHALL play the fall and the flicker from what it is told. When the authority stops reporting it, the client SHALL take it out of its world as it does any other fighter that has gone, and the player SHALL have no fighter until it is given one.

A fighter's flicker SHALL be bounded whether or not the fighter is removed on time. Past the end of its flicker the fighter SHALL stay hidden rather than keep blinking, so a client waiting to be told something never presents a strobing sprite.

#### Scenario: Death seen from the inside
- **WHEN** a player's own fighter dies in a networked match
- **THEN** that player sees it fall and flicker, and then sees it gone

#### Scenario: Nothing is left blinking
- **WHEN** a client holds a dead fighter for longer than its flicker lasts, whatever the reason
- **THEN** the fighter stays hidden from that point instead of blinking on and off

#### Scenario: The gap between
- **WHEN** a player's fighter has gone and the respawn has not arrived
- **THEN** the player has no fighter on screen and no fighter under its control, and the view stays where it was

#### Scenario: Single player hands over too
- **WHEN** a player's fighter dies and respawns in an offline match
- **THEN** the new fighter is the player's own: it takes the player's input, and the effects that follow the player's fighter follow it

### Requirement: Restarting a single-player level
In a single-player match, outside edit mode, R SHALL restart the level: the match SHALL start over as it started from the menu, with the player and every bot alive at full structure on their spawn positions, and the recorded history for time manipulation SHALL be dropped. R SHALL do nothing in a networked match.

#### Scenario: Bodies get back up
- **WHEN** the player presses R after beating every bot
- **THEN** the level starts over with the player and all the bots alive

### Requirement: Structure is decided by the authority
Structure changes, stuns, structure breaks, deaths and respawns SHALL be decided only by the authoritative simulation, as hits and blocks already are. A client's predicted world SHALL display the structure, Death state and removals it is told about, and SHALL NOT invent them.

#### Scenario: A predicting client does not kill anyone
- **WHEN** a client's predicted world runs a swing that overlaps another fighter
- **THEN** no structure is deducted locally and no fighter dies until the server's snapshot says so

#### Scenario: Server corrects the client
- **WHEN** a snapshot reports a different structure for a fighter than the client is showing
- **THEN** the client shows the server's value
