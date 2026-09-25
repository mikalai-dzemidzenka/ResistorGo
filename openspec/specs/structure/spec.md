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
A hit that lands SHALL cost the defender structure and SHALL produce a flinch: the hit flash, the hit-stop freeze that both fighters already share, and a small knockback. A flinch SHALL NOT play an animation of its own: the defender SHALL keep whatever animation it was already playing, and SHALL be shown to have been hit by the flash alone. A flinch SHALL NOT lock the defender out of acting — it can move, attack, block and jump during one — and SHALL NOT interrupt an action the defender had already started.

The flash SHALL hold for about 0.2 seconds and then stop on its own. It SHALL also stop early with whatever replaces the flinch: a fighter that swings, blocks, is stunned or dies while flashing SHALL stop flashing at that moment, so no fighter is ever left lit.

The knockback SHALL move the fighter, not merely be given to it: after a flinch the fighter SHALL be somewhere it was not when the hit landed. It SHALL remain clearly smaller than the throw a stun gives, so the two read as different events.

What a hit costs SHALL depend on how many hits have already landed on that fighter in the same run. The first hit of a run SHALL cost 1 structure; each hit after it SHALL cost half a point more than the one before — 1, 1.5, 2, 2.5, 3 — and the cost SHALL NOT climb past 3 however long the run goes on. The run SHALL be per fighter: it counts what has landed on that fighter, whoever landed it.

The hit-stop freeze SHALL hold input, never discard it. An attack or a jump pressed while a fighter is frozen SHALL be held for that fighter and offered to its first tick after the freeze ends, alongside anything pressed on that tick, and SHALL then take effect only as far as the fighter's state at that moment allows. This holds for both fighters the freeze covers: the defender's counter and the attacker's next swing are equally kept. A freeze is therefore unlike a paused mode, which is not holding input at all — while a paused mode owns the mouse and keyboard the press belongs to that mode, and the simulation's latch is cleared because it was never the intended consumer.

Input held across a freeze SHALL NOT outlive it: whatever the first tick after the freeze does with the press, held or acted on, the fighter SHALL NOT be holding it afterwards. A held button SHALL NOT produce more than the one action it was pressed for, however many frozen ticks it spans. A fighter that dies SHALL discard anything it was holding, so nothing crosses into the fighter that replaces it.

#### Scenario: First hit does not stun
- **WHEN** a fighter at full structure takes one hit
- **THEN** its structure is 11, it flashes and is knocked back slightly, and it can act on its first tick after the freeze

#### Scenario: The small knockback actually moves the fighter
- **WHEN** a fighter is flinched and then left alone until the knockback has run out
- **THEN** its position differs from where it was when the hit landed

#### Scenario: A flinch stays smaller than a stun
- **WHEN** one fighter is flinched and another stunned by attackers facing the same way, and both are left alone until their knockbacks have run out
- **THEN** the flinched one has travelled a visibly shorter distance than the stunned one

#### Scenario: A run costs more as it goes
- **WHEN** a fighter at full structure takes five hits in a row with nothing to break the run
- **THEN** they cost 1, 1.5, 2, 2.5 and 3, leaving it at 2 structure

#### Scenario: The cost stops climbing
- **WHEN** a run reaches a sixth and a seventh hit
- **THEN** each costs 3, the same as the fifth

#### Scenario: Six hits end a fight
- **WHEN** a fighter at full structure takes six hits in a row
- **THEN** its structure reaches zero and it dies

#### Scenario: The stun does not relieve it
- **WHEN** a fighter takes four hits in a row, the third of which stuns it
- **THEN** the fourth costs 2.5, as the fourth hit of a run does, rather than starting the run again

#### Scenario: Time ends a run
- **WHEN** a fighter is left alone long enough for its structure to start refilling, and is then hit again
- **THEN** that hit costs 1, as the first hit of a run does

#### Scenario: Flinch does not interrupt a swing
- **WHEN** a fighter is hit while its own attack animation is playing
- **THEN** that attack continues and can still land

#### Scenario: A flinch leaves the animation alone
- **WHEN** a running fighter takes a hit that does not stun it
- **THEN** it goes on running, with the running animation uninterrupted, and flashes

#### Scenario: The flash runs out
- **WHEN** a fighter has been flinched and half a second passes with nothing else happening to it
- **THEN** it is no longer flashing

#### Scenario: Acting ends the flash
- **WHEN** a flinched fighter swings while it is still flashing
- **THEN** the flash stops as the swing starts

#### Scenario: A counter pressed during the freeze still comes out
- **WHEN** a fighter is hit and the player presses attack during the hit-stop freeze
- **THEN** the fighter attacks on its first tick after the freeze, and pressing attack a moment later has the same result

#### Scenario: The attacker's next swing is kept too
- **WHEN** a swing lands and the player presses attack during the freeze it causes
- **THEN** that press is queued as the chain's next swing, exactly as a press later in the same swing would be

#### Scenario: A jump pressed during the freeze is kept
- **WHEN** a fighter is hit and the player presses jump during the hit-stop freeze
- **THEN** the fighter leaves the ground on its first tick after the freeze, as it would have done had the press landed on that tick

#### Scenario: One press is one action
- **WHEN** the player presses attack once during a freeze that lasts several ticks
- **THEN** exactly one attack results

#### Scenario: A stun still locks the defender out
- **WHEN** a fighter is stunned and the player presses attack during the hit-stop freeze
- **THEN** the fighter does not attack, and the press does not survive the stun to fire when it ends

#### Scenario: A press cannot be afforded away while it waits
- **WHEN** a fighter with exactly enough energy for one swing is hit and the player presses attack during the freeze
- **THEN** the swing comes out on the first tick after the freeze and is paid for, because nothing spent or refilled the pool while the fighter was frozen

#### Scenario: Death discards a held press
- **WHEN** the hit that breaks a fighter's structure is the one during whose freeze the player presses attack
- **THEN** the fighter dies holding nothing, and the press does not act for it or for the fighter the player is given next

### Requirement: Every third landed hit stuns
The third hit to land on a fighter SHALL stun it: the defender enters the hit-stun state, is locked out until the stun animation finishes, and is thrown back noticeably further than a flinch throws it. The stun SHALL be the only hit that plays the hit animation, so the animation tells the player their fighter has lost control rather than merely been hit; the flash SHALL cover the opening of that animation, not the whole of it. The count SHALL be per fighter and SHALL reset when the stun lands, when the fighter dies, and when its structure begins regenerating. A block SHALL NOT advance it.

The throw-back SHALL move the fighter, not merely be given to it: a stunned fighter SHALL end up a real distance from where the hit landed, far enough that it reads as being thrown rather than as a jitter. Being locked out SHALL NOT cap how far the hit carries it — the lockout governs what the fighter may do, never how far the attack throws it — and the throw SHALL go the way the attacker was facing.

#### Scenario: Third hit of a run
- **WHEN** a fighter takes three hits without its structure regenerating in between
- **THEN** the first two flinch it and the third stuns it, locking it out and throwing it back further

#### Scenario: Only the stun has an animation
- **WHEN** a fighter takes three hits in a row
- **THEN** neither of the first two changes the animation it was playing, and the third plays the hit animation

#### Scenario: The throw carries a real distance
- **WHEN** a fighter is stunned and then left alone until the throw has run out
- **THEN** it has been carried at least ten whole pixels from where it was hit

#### Scenario: The lockout does not cap the throw
- **WHEN** a fighter is stunned
- **THEN** the distance it travels is set by the knockback the hit gave it and the friction that slows it, and by nothing belonging to the state it was put into

#### Scenario: The throw follows the attacker's facing
- **WHEN** a fighter is stunned by an attacker facing left, and another by an attacker facing right
- **THEN** each is carried the way its own attacker was facing

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
A block that holds until the swing's hit frame SHALL restore 0.5 structure to the defender, except when the defender is guarding with block-all (guard height 3), which SHALL instead cost 0.4 structure — less than a landed hit costs. A block SHALL NOT flinch or stun the defender.

A successful block SHALL end the run of hits against the defender, whichever guard stopped the attack: a guard at the attacker's height or block-all. The next hit to land on that fighter SHALL cost 1, as the first hit of a run does. Stopping an attack is therefore worth more than the structure it restores — it is how a defender gets out from under a run.

#### Scenario: Matching guard restores
- **WHEN** a fighter at 8 structure blocks an attack with a guard at the attacker's height
- **THEN** its structure is 8.5 and it is not flinched

#### Scenario: Block-all costs, but less than a hit
- **WHEN** a fighter at 8 structure blocks an attack while guarding with block-all
- **THEN** its structure is 7.6, which is more structure than the 7 it would have had if the attack had landed

#### Scenario: A block breaks the run
- **WHEN** a fighter that has taken three hits in a row blocks the next attack at the attacker's height, and is then hit again
- **THEN** that hit costs 1 rather than the 2.5 a fourth hit of the run would have cost

#### Scenario: Block-all breaks it too
- **WHEN** a fighter that has taken three hits in a row stops the next attack with block-all, and is then hit again
- **THEN** that hit costs 1

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
- **THEN** its structure never regenerates: it is stunned on the third hit and broken on the sixth

### Requirement: Breaking structure kills
Structure reaching zero SHALL put the fighter into a Death state, whether the last cost came from a hit or from block-all. In the Death state the fighter SHALL play the death animation once, SHALL NOT act, SHALL NOT be hit or blocked by anyone, and SHALL NOT take further structure changes. When the animation finishes the fighter SHALL flicker for about 0.6 seconds and then be removed from the world, except a bot in a single-player match, which SHALL stay lying where it fell, visible and without flickering, for the rest of the match.

#### Scenario: A clean run breaks it
- **WHEN** a fighter at full structure takes hits with no block and no regeneration in between
- **THEN** it is stunned on the third, and the sixth breaks its structure and puts it into the Death state

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
